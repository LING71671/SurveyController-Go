package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/LING71671/SurveyController-go/internal/logging"
)

type Task func(ctx context.Context, workerID int) error

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
	taskCh := make(chan Task)
	var wg sync.WaitGroup

	for workerID := 1; workerID <= p.options.Concurrency; workerID++ {
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
	p.emit(logging.RunEvent{
		Type:    logging.EventRunFinished,
		Level:   logging.LevelInfo,
		Message: "run finished",
		Fields: map[string]any{
			"successes": snapshot.Successes,
			"failures":  snapshot.Failures,
		},
	})
	return snapshot
}

func (p *WorkerPool) worker(ctx context.Context, workerID int, tasks <-chan Task) {
	p.state.SetWorkerStatus(workerID, WorkerStatusRunning, "worker started")
	p.emit(logging.RunEvent{
		Type:     logging.EventWorkerStarted,
		Level:    logging.LevelInfo,
		WorkerID: workerID,
		Message:  "worker started",
	})
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
				p.state.RecordFailure(workerID, err.Error())
				p.emit(logging.RunEvent{
					Type:     logging.EventSubmissionFailure,
					Level:    logging.LevelError,
					WorkerID: workerID,
					Message:  err.Error(),
				})
				continue
			}
			p.state.RecordSuccess(workerID)
			p.emit(logging.RunEvent{
				Type:     logging.EventSubmissionSuccess,
				Level:    logging.LevelInfo,
				WorkerID: workerID,
				Message:  "submission succeeded",
			})
		}
	}
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
