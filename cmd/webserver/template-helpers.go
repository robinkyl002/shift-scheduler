package main

import "fmt"

func formatHour(hour int) string {
	if hour < 12 {
		return fmt.Sprintf("%dam", hour)
	} else if hour == 12 {
		return "12pm"
	}
	return fmt.Sprintf("%dpm", hour-12)
}

func formatMinutes(totalMinutes int) string {
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
