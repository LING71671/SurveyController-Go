package answer

import (
	"fmt"
	"math"
)

func Mean(values []float64) (float64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("values are required")
	}
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values)), nil
}

func Variance(values []float64) (float64, error) {
	if len(values) < 2 {
		return 0, fmt.Errorf("at least two values are required")
	}
	mean, err := Mean(values)
	if err != nil {
		return 0, err
	}
	sum := 0.0
	for _, value := range values {
		delta := value - mean
		sum += delta * delta
	}
	return sum / float64(len(values)-1), nil
}

func CronbachAlpha(rows [][]float64) (float64, error) {
	if len(rows) < 2 {
		return 0, fmt.Errorf("at least two rows are required")
	}
	itemCount := len(rows[0])
	if itemCount < 2 {
		return 0, fmt.Errorf("at least two items are required")
	}

	itemValues := make([][]float64, itemCount)
	totals := make([]float64, len(rows))
	for rowIndex, row := range rows {
		if len(row) != itemCount {
			return 0, fmt.Errorf("row %d has %d items, want %d", rowIndex, len(row), itemCount)
		}
		for itemIndex, value := range row {
			itemValues[itemIndex] = append(itemValues[itemIndex], value)
			totals[rowIndex] += value
		}
	}

	itemVarianceSum := 0.0
	for _, values := range itemValues {
		variance, err := Variance(values)
		if err != nil {
			return 0, err
		}
		itemVarianceSum += variance
	}
	totalVariance, err := Variance(totals)
	if err != nil {
		return 0, err
	}
	if totalVariance == 0 {
		return 0, fmt.Errorf("total variance must be greater than 0")
	}

	k := float64(itemCount)
	return (k / (k - 1)) * (1 - itemVarianceSum/totalVariance), nil
}

func DimensionMean(values []float64) (float64, error) {
	return Mean(values)
}

func ConsistentDimension(values []float64, target float64, tolerance float64) (bool, error) {
	if tolerance < 0 {
		return false, fmt.Errorf("tolerance must not be negative")
	}
	mean, err := DimensionMean(values)
	if err != nil {
		return false, err
	}
	return math.Abs(mean-target) <= tolerance, nil
}
