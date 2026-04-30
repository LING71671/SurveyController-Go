package runner

import "sync"

type WorkerStatus string

const (
	WorkerStatusIdle    WorkerStatus = "idle"
	WorkerStatusRunning WorkerStatus = "running"
	WorkerStatusStopped WorkerStatus = "stopped"
)

type StateOptions struct {
	Target           int
	FailureThreshold int
}

type RunState struct {
	mu        sync.Mutex
	target    int
	threshold int
	success   int
	failure   int
	workers   map[int]WorkerProgress
}

type WorkerProgress struct {
	ID        int
	Status    WorkerStatus
	Successes int
	Failures  int
	Message   string
}

type StateSnapshot struct {
	Target           int
	FailureThreshold int
	Successes        int
	Failures         int
	Workers          map[int]WorkerProgress
}

func NewRunState(options StateOptions) *RunState {
	return &RunState{
		target:    options.Target,
		threshold: options.FailureThreshold,
		workers:   map[int]WorkerProgress{},
	}
}

func (s *RunState) RecordSuccess(workerID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.success++
	progress := s.worker(workerID)
	progress.Successes++
	progress.Status = WorkerStatusRunning
	s.workers[workerID] = progress
}

func (s *RunState) RecordFailure(workerID int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failure++
	progress := s.worker(workerID)
	progress.Failures++
	progress.Status = WorkerStatusRunning
	progress.Message = message
	s.workers[workerID] = progress
}

func (s *RunState) SetWorkerStatus(workerID int, status WorkerStatus, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	progress := s.worker(workerID)
	progress.Status = status
	progress.Message = message
	s.workers[workerID] = progress
}

func (s *RunState) Snapshot() StateSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	workers := make(map[int]WorkerProgress, len(s.workers))
	for id, progress := range s.workers {
		workers[id] = progress
	}
	return StateSnapshot{
		Target:           s.target,
		FailureThreshold: s.threshold,
		Successes:        s.success,
		Failures:         s.failure,
		Workers:          workers,
	}
}

func (s *RunState) ShouldStop() bool {
	snapshot := s.Snapshot()
	return snapshot.TargetReached() || snapshot.FailureThresholdReached()
}

func (s StateSnapshot) TargetReached() bool {
	return s.Target > 0 && s.Successes >= s.Target
}

func (s StateSnapshot) FailureThresholdReached() bool {
	return s.FailureThreshold > 0 && s.Failures >= s.FailureThreshold
}

func (s *RunState) worker(workerID int) WorkerProgress {
	progress, ok := s.workers[workerID]
	if !ok {
		progress = WorkerProgress{
			ID:     workerID,
			Status: WorkerStatusIdle,
		}
	}
	return progress
}
