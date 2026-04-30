package answer

import (
	"math"
	"strings"
	"testing"
)

func TestMeanAndVariance(t *testing.T) {
	mean, err := Mean([]float64{1, 2, 3})
	if err != nil {
		t.Fatalf("Mean() returned error: %v", err)
	}
	if mean != 2 {
		t.Fatalf("Mean() = %f, want 2", mean)
	}

	variance, err := Variance([]float64{1, 2, 3})
	if err != nil {
		t.Fatalf("Variance() returned error: %v", err)
	}
	if variance != 1 {
		t.Fatalf("Variance() = %f, want 1", variance)
	}
}

func TestCronbachAlpha(t *testing.T) {
	alpha, err := CronbachAlpha([][]float64{
		{1, 2, 3},
		{2, 3, 4},
		{3, 4, 5},
		{4, 5, 6},
	})
	if err != nil {
		t.Fatalf("CronbachAlpha() returned error: %v", err)
	}
	if math.Abs(alpha-1) > 0.0001 {
		t.Fatalf("CronbachAlpha() = %f, want about 1", alpha)
	}
}

func TestCronbachAlphaRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		rows [][]float64
		want string
	}{
		{name: "rows", rows: [][]float64{{1, 2}}, want: "rows"},
		{name: "items", rows: [][]float64{{1}, {2}}, want: "items"},
		{name: "ragged", rows: [][]float64{{1, 2}, {1}}, want: "row 1"},
		{name: "variance", rows: [][]float64{{1, 1}, {1, 1}}, want: "total variance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CronbachAlpha(tt.rows)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("CronbachAlpha() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestConsistentDimension(t *testing.T) {
	ok, err := ConsistentDimension([]float64{4, 5, 5}, 4.5, 0.2)
	if err != nil {
		t.Fatalf("ConsistentDimension() returned error: %v", err)
	}
	if !ok {
		t.Fatalf("ConsistentDimension() = false, want true")
	}

	ok, err = ConsistentDimension([]float64{1, 2, 3}, 5, 0.5)
	if err != nil {
		t.Fatalf("ConsistentDimension() returned error: %v", err)
	}
	if ok {
		t.Fatalf("ConsistentDimension() = true, want false")
	}
}

func TestConsistentDimensionRejectsNegativeTolerance(t *testing.T) {
	_, err := ConsistentDimension([]float64{1, 2}, 1, -1)
	if err == nil || !strings.Contains(err.Error(), "tolerance") {
		t.Fatalf("ConsistentDimension() error = %v, want tolerance error", err)
	}
}
