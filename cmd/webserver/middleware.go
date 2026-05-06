package main

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type SchedulePageData struct {
	Days          []string
	Hours         []int
	SlicesPerHour int
}

func buildSchedulePageData() SchedulePageData {
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
	hours := []int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17}
	slicesPerHour := 6
	return SchedulePageData{
		Days:          days,
		Hours:         hours,
		SlicesPerHour: slicesPerHour,
	}
}

func parseSelectedSlots(raw string) ([]TimeSlot, error) {
	var slotKeys []string
	err := json.Unmarshal([]byte(raw), &slotKeys)

	if err != nil {
		return nil, err
	}

	var slots []TimeSlot
	for _, key := range slotKeys {
		parts := strings.Split(key, "|")

		if len(parts) != 3 {
			return nil, errors.New("Invalid slot format")
		}

		day := parts[0]

		hour, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, errors.New("Invalid hour value")
		}

		slice, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, errors.New("Invalid slice value")
		}

		slot := TimeSlot{
			Day:   day,
			Hour:  hour,
			Slice: slice,
		}

		slots = append(slots, slot)
	}

	return slots, nil

}
