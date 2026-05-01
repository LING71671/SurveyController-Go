package runner

import (
	"math"
	"time"
)

type RunPlanReport struct {
	Provider          string  `json:"provider"`
	URL               string  `json:"url"`
	Mode              string  `json:"mode"`
	Target            int     `json:"target"`
	Concurrency       int     `json:"concurrency"`
	Successes         int     `json:"successes"`
	Failures          int     `json:"failures"`
	Completed         int     `json:"completed"`
	CompletionRate    float64 `json:"completion_rate"`
	SuccessRate       float64 `json:"success_rate"`
	DurationMS        int64   `json:"duration_ms,omitempty"`
	ThroughputPerSec  float64 `json:"throughput_per_second,omitempty"`
	Goroutines        int     `json:"goroutines,omitempty"`
	HeapAllocBytes    uint64  `json:"heap_alloc_bytes,omitempty"`
	HeapAllocDelta    int64   `json:"heap_alloc_delta_bytes,omitempty"`
	TotalAllocDelta   uint64  `json:"total_alloc_delta_bytes,omitempty"`
	StopRequested     bool    `json:"stop_requested"`
	FailureThreshold  bool    `json:"failure_threshold_reached"`
	StopReason        string  `json:"stop_reason,omitempty"`
	StopFailureReason string  `json:"stop_failure_reason,omitempty"`
	WorkerCount       int     `json:"worker_count"`
}

type RunResourceMetrics struct {
	Goroutines      int
	HeapAllocBytes  uint64
	HeapAllocDelta  int64
	TotalAllocDelta uint64
}

func NewTimedRunPlanReport(plan Plan, snapshot StateSnapshot, elapsed time.Duration) RunPlanReport {
	report := NewRunPlanReport(plan, snapshot)
	if elapsed <= 0 {
		return report
	}
	report.DurationMS = elapsed.Milliseconds()
	report.ThroughputPerSec = ratioPerSecond(report.Completed, elapsed)
	return report
}

func (r RunPlanReport) WithResourceMetrics(metrics RunResourceMetrics) RunPlanReport {
	r.Goroutines = metrics.Goroutines
	r.HeapAllocBytes = metrics.HeapAllocBytes
	r.HeapAllocDelta = metrics.HeapAllocDelta
	r.TotalAllocDelta = metrics.TotalAllocDelta
	return r
}

func NewRunPlanReport(plan Plan, snapshot StateSnapshot) RunPlanReport {
	completed := snapshot.Successes + snapshot.Failures
	return RunPlanReport{
		Provider:          plan.Provider,
		URL:               plan.URL,
		Mode:              plan.Mode.String(),
		Target:            plan.Target,
		Concurrency:       plan.Concurrency,
		Successes:         snapshot.Successes,
		Failures:          snapshot.Failures,
		Completed:         completed,
		CompletionRate:    ratio(completed, plan.Target),
		SuccessRate:       ratio(snapshot.Successes, completed),
		StopRequested:     snapshot.StopRequested,
		FailureThreshold:  snapshot.FailureThresholdReached(),
		StopReason:        snapshot.StopReason,
		StopFailureReason: snapshot.StopFailureReason,
		WorkerCount:       len(snapshot.Workers),
	}
}

func (r RunPlanReport) TargetReached() bool {
	return r.Target > 0 && r.Successes >= r.Target
}

func (r RunPlanReport) HasFailures() bool {
	return r.Failures > 0
}

func ratio(value int, total int) float64 {
	if total <= 0 {
		return 0
	}
	return roundRatio(float64(value) / float64(total))
}

func roundRatio(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func ratioPerSecond(value int, elapsed time.Duration) float64 {
	if value <= 0 || elapsed <= 0 {
		return 0
	}
	return roundRatio(float64(value) / elapsed.Seconds())
}
