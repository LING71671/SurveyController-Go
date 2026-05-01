package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/LING71671/SurveyController-go/internal/engine"
	"github.com/LING71671/SurveyController-go/internal/logging"
)

type Task func(ctx context.Context, workerID int) error
type SubmissionTask func(ctx context.Context, workerID int) (engine.SubmissionResult, error)
type SubmissionTaskGenerator func(index int) (SubmissionTask, error)

const DefaultMaxWorkerConcurrency = engine.LightWorkerConcurrencyBaseline

type PoolOptions struct {
	Concurrency      int
	Target           int
	FailureThreshold int
	Events           chan<- logging.RunEvent
}

type WorkerPool struct {
	options PoolOptions
	state   *RunState
}

func NewWorkerPool(options PoolOptions) (*WorkerPool, error) {
	if options.Concurrency <= 0 {
		return nil, fmt.Errorf("concurrency must be greater than 0")
	}
	if options.Concurrency > DefaultMaxWorkerConcurrency {
		return nil, fmt.Errorf("concurrency must not exceed %d", DefaultMaxWorkerConcurrency)
	}
	if options.Target < 0 {
		return nil, fmt.Errorf("target must not be negative")
	}
	return &WorkerPool{
		options: options,
		state: NewRunState(StateOptions{
			Target:           options.Target,
			FailureThreshold: options.FailureThreshold,
		}),
	}, nil
}

func (p *WorkerPool) Run(ctx context.Context, tasks []Task) StateSnapshot {
	p.emit(logging.NewEvent(logging.EventRunStarted, "run started"))
	workerCount := p.workerCount(len(tasks))
	taskCh := make(chan Task, workerCount)
	var wg sync.WaitGroup

	for workerID := 1; workerID <= workerCount; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p.worker(ctx, id, taskCh)
		}(workerID)
	}

enqueue:
	for _, task := range tasks {
		if p.state.ShouldStop() {
			break
		}
		select {
		case <-ctx.Done():
			break enqueue
		case taskCh <- task:
		}
	}
	close(taskCh)
	wg.Wait()

	snapshot := p.state.Snapshot()
	event := logging.NewEvent(logging.EventRunFinished, "run finished")
	event.Fields = map[string]any{
		"successes": snapshot.Successes,
		"failures":  snapshot.Failures,
	}
	p.emit(event)
	return snapshot
}

func (p *WorkerPool) RunSubmissions(ctx context.Context, tasks []SubmissionTask) StateSnapshot {
	snapshot, _ := p.RunGeneratedSubmissions(ctx, len(tasks), func(index int) (SubmissionTask, error) {
		return tasks[index], nil
	})
	return snapshot
}

func (p *WorkerPool) RunGeneratedSubmissions(ctx context.Context, taskCount int, next SubmissionTaskGenerator) (StateSnapshot, error) {
	if taskCount < 0 {
		return p.state.Snapshot(), fmt.Errorf("task count must not be negative")
	}
	if taskCount > 0 && next == nil {
		return p.state.Snapshot(), fmt.Errorf("submission task generator is required")
	}

	p.emit(logging.NewEvent(logging.EventRunStarted, "run started"))
	workerCount := p.workerCount(taskCount)
	taskCh := make(chan SubmissionTask, workerCount)
	stopCh := make(chan struct{})
	var stopOnce sync.Once
	signalStop := func() {
		stopOnce.Do(func() {
			close(stopCh)
		})
	}
	var wg sync.WaitGroup

	for workerID := 1; workerID <= workerCount; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p.submissionWorker(ctx, id, taskCh, signalStop)
		}(workerID)
	}

	var generateErr error
enqueue:
	for index := 0; index < taskCount; index++ {
		if p.state.ShouldStop() {
			break
		}
		task, err := next(index)
		if err != nil {
			generateErr = err
			break
		}
		select {
		case <-ctx.Done():
			break enqueue
		case <-stopCh:
			break enqueue
		case taskCh <- task:
		}
	}
	close(taskCh)
	wg.Wait()

	return p.finishRun(), generateErr
}

func (p *WorkerPool) worker(ctx context.Context, workerID int, tasks <-chan Task) {
	p.state.SetWorkerStatus(workerID, WorkerStatusRunning, "worker started")
	event := logging.NewEvent(logging.EventWorkerStarted, "worker started")
	event.WorkerID = workerID
	p.emit(event)
	defer p.state.SetWorkerStatus(workerID, WorkerStatusStopped, "worker stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-tasks:
			if !ok || p.state.ShouldStop() {
				return
			}
			if err := task(ctx, workerID); err != nil {
				p.state.RecordFailureWithCode(workerID, err.Error(), errorCode(err))
				if p.eventsEnabled() {
					event := logging.NewEvent(logging.EventSubmissionFailure, err.Error())
					event.Level = logging.LevelError
					event.WorkerID = workerID
					addErrorFields(&event, err)
					p.emit(event)
				}
				continue
			}
			p.state.RecordSuccess(workerID)
			if p.eventsEnabled() {
				event := logging.NewEvent(logging.EventSubmissionSuccess, "submission succeeded")
				event.WorkerID = workerID
				p.emit(event)
			}
		}
	}
}

func (p *WorkerPool) submissionWorker(ctx context.Context, workerID int, tasks <-chan SubmissionTask, signalStop func()) {
	p.startWorker(workerID)
	defer p.state.SetWorkerStatus(workerID, WorkerStatusStopped, "worker stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-tasks:
			if !ok || p.state.ShouldStop() {
				if p.state.ShouldStop() && signalStop != nil {
					signalStop()
				}
				return
			}
			result, err := task(ctx, workerID)
			if err != nil {
				p.state.RecordFailureWithCode(workerID, err.Error(), errorCode(err))
				if p.eventsEnabled() {
					event := logging.NewEvent(logging.EventSubmissionFailure, err.Error())
					event.Level = logging.LevelError
					event.WorkerID = workerID
					addErrorFields(&event, err)
					p.emit(event)
				}
				if p.state.ShouldStop() && signalStop != nil {
					signalStop()
					return
				}
				continue
			}
			p.state.RecordSubmissionResult(workerID, result)
			if p.eventsEnabled() {
				p.emit(EventForSubmissionResult(workerID, result))
			}
			if p.state.ShouldStop() && signalStop != nil {
				signalStop()
				return
			}
		}
	}
}

func (p *WorkerPool) startWorker(workerID int) {
	p.state.SetWorkerStatus(workerID, WorkerStatusRunning, "worker started")
	event := logging.NewEvent(logging.EventWorkerStarted, "worker started")
	event.WorkerID = workerID
	p.emit(event)
}

func (p *WorkerPool) finishRun() StateSnapshot {
	snapshot := p.state.Snapshot()
	event := logging.NewEvent(logging.EventRunFinished, "run finished")
	event.Fields = map[string]any{
		"successes":      snapshot.Successes,
		"failures":       snapshot.Failures,
		"stop_requested": snapshot.StopRequested,
	}
	p.emit(event)
	return snapshot
}

func (p *WorkerPool) emit(event logging.RunEvent) {
	if p.options.Events == nil {
		return
	}
	select {
	case p.options.Events <- event:
	default:
	}
}

func (p *WorkerPool) eventsEnabled() bool {
	return p.options.Events != nil
}

func (p *WorkerPool) workerCount(taskCount int) int {
	if taskCount < p.options.Concurrency {
		return taskCount
	}
	return p.options.Concurrency
}
