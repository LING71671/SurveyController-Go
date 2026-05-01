package runner

import "math"

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
	StopRequested     bool    `json:"stop_requested"`
	StopReason        string  `json:"stop_reason,omitempty"`
	StopFailureReason string  `json:"stop_failure_reason,omitempty"`
	WorkerCount       int     `json:"worker_count"`
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
