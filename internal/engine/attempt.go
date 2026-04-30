package engine

import (
	"context"
	"errors"
	"fmt"
)

type AttemptResource interface {
	ResourceName() string
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type ExecutionAttempt struct {
	resources  []AttemptResource
	finalized  bool
	committed  bool
	rolledBack bool
}

type AttemptSnapshot struct {
	ResourceCount int
	Finalized     bool
	Committed     bool
	RolledBack    bool
}

func NewExecutionAttempt(capacity int) *ExecutionAttempt {
	if capacity < 0 {
		capacity = 0
	}
	return &ExecutionAttempt{
		resources: make([]AttemptResource, 0, capacity),
	}
}

func (a *ExecutionAttempt) AddResource(resource AttemptResource) error {
	if a == nil {
		return fmt.Errorf("execution attempt is nil")
	}
	if a.finalized {
		return fmt.Errorf("execution attempt is already finalized")
	}
	if resource == nil {
		return fmt.Errorf("attempt resource is nil")
	}
	a.resources = append(a.resources, resource)
	return nil
}

func (a *ExecutionAttempt) Commit(ctx context.Context) error {
	if err := a.ensureOpen(); err != nil {
		return err
	}

	committed := 0
	for _, resource := range a.resources {
		if err := resource.Commit(ctx); err != nil {
			rollbackErr := a.rollbackCommitted(ctx, committed)
			a.finalized = true
			a.rolledBack = true
			return errors.Join(wrapResourceError("commit", resource, err), rollbackErr)
		}
		committed++
	}
	a.finalized = true
	a.committed = true
	return nil
}

func (a *ExecutionAttempt) Rollback(ctx context.Context) error {
	if err := a.ensureOpen(); err != nil {
		return err
	}

	a.finalized = true
	a.rolledBack = true
	return a.rollbackCommitted(ctx, len(a.resources))
}

func (a *ExecutionAttempt) Finalize(ctx context.Context, result SubmissionResult) error {
	if result.Success {
		return a.Commit(ctx)
	}
	return a.Rollback(ctx)
}

func (a *ExecutionAttempt) Snapshot() AttemptSnapshot {
	if a == nil {
		return AttemptSnapshot{}
	}
	return AttemptSnapshot{
		ResourceCount: len(a.resources),
		Finalized:     a.finalized,
		Committed:     a.committed,
		RolledBack:    a.rolledBack,
	}
}

func (a *ExecutionAttempt) ensureOpen() error {
	if a == nil {
		return fmt.Errorf("execution attempt is nil")
	}
	if a.finalized {
		return fmt.Errorf("execution attempt is already finalized")
	}
	return nil
}

func (a *ExecutionAttempt) rollbackCommitted(ctx context.Context, count int) error {
	var joined error
	for i := count - 1; i >= 0; i-- {
		resource := a.resources[i]
		if err := resource.Rollback(ctx); err != nil {
			joined = errors.Join(joined, wrapResourceError("rollback", resource, err))
		}
	}
	return joined
}

func wrapResourceError(action string, resource AttemptResource, err error) error {
	name := resource.ResourceName()
	if name == "" {
		name = "unnamed"
	}
	return fmt.Errorf("%s %s: %w", action, name, err)
}
