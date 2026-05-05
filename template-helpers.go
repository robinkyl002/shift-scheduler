package main

import "fmt"

func formatHour(hour int) string {
	if hour < 12 {
		return fmt.Sprintf("%dam", hour)
	}
	return fmt.Sprintf("%dpm", hour-12)
}

func sliceIndices(n int) []int {
	values := make([]int, n)
	for i := range values {
		values[i] = i
	}
	return values
}
