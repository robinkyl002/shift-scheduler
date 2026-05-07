package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

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

func validateSchedule(schedule ScheduleSubmission) ValidationResult {
	result := ValidationResult{
		Errors:        []string{},
		DailyMinutes:  map[string]int{},
		WeeklyMinutes: 0,
	}

	slots := schedule.Slots

	sort.Slice(slots, func(i, j int) bool {
		if slots[i].Day != slots[j].Day {
			return dayOrder[slots[i].Day] < dayOrder[slots[j].Day]
		}
		if slots[i].Hour != slots[j].Hour {
			return slots[i].Hour < slots[j].Hour
		}
		return slots[i].Slice < slots[j].Slice
	})

	slotsByDay := groupSlotsByDay(slots)

	for day, daySlots := range slotsByDay {
		shiftBlocks := buildShiftBlocks(day, daySlots)

		for _, block := range shiftBlocks {
			if block.SlotCount*10 < 180 {
				startTime := formTimeFromBlock(block, true)
				endTime := formTimeFromBlock(block, false)
				result.Errors = append(result.Errors, "Shift from "+startTime+" to "+endTime+" is too short")
			}
		}

		if calculateDayMinutes(daySlots) > 540 {
			result.Errors = append(result.Errors, "You may not work more than 9 hours per day. You currently have a shift longer than this on "+day)
		}
	}

	if weeklyMinutes := calculateWeeklyMinutes(slots); weeklyMinutes < 1200 || weeklyMinutes > 2400 {
		result.Errors = append(result.Errors,
			"You must be scheduled to work between 20 and 40 hours. You are trying to submit a schedule to work "+weeklyMinutesToHours(weeklyMinutes))
	}

	return result
}

func (t TimeSlot) StartMinute() int {
	return (t.Hour-scheduleStartHour)*60 + (t.Slice * minutesPerSlice)
}

func groupSlotsByDay(slots []TimeSlot) map[string][]TimeSlot {
	groups := make(map[string][]TimeSlot)

	for _, slot := range slots {
		if _, exists := groups[slot.Day]; !exists {
			groups[slot.Day] = []TimeSlot{slot}
		} else {
			groups[slot.Day] = append(groups[slot.Day], slot)
		}
	}
	return groups
}

func buildShiftBlocks(day string, slots []TimeSlot) []ShiftBlock {
	if len(slots) == 0 {
		return []ShiftBlock{}
	}

	shiftBlocks := []ShiftBlock{}

	currShiftBlock := ShiftBlock{
		Day:         day,
		StartMinute: slots[0].StartMinute(),
		EndMinute:   slots[0].StartMinute() + minutesPerSlice,
		SlotCount:   1,
	}
	for _, slot := range slots[1:] {

		if slot.StartMinute() == currShiftBlock.EndMinute {
			currShiftBlock.EndMinute += minutesPerSlice
			currShiftBlock.SlotCount++
		} else {
			shiftBlocks = append(shiftBlocks, currShiftBlock)
			currShiftBlock = ShiftBlock{
				Day:         day,
				StartMinute: slot.StartMinute(),
				EndMinute:   slot.StartMinute() + minutesPerSlice,
				SlotCount:   1,
			}
		}
	}

	shiftBlocks = append(shiftBlocks, currShiftBlock)

	return shiftBlocks
}

func calculateDayMinutes(slots []TimeSlot) int {
	return len(slots) * minutesPerSlice
}

func calculateWeeklyMinutes(slots []TimeSlot) int {
	return len(slots) * minutesPerSlice
}

func formTimeFromBlock(block ShiftBlock, start bool) string {
	var hour int
	var minute int
	if start {
		hour = (block.StartMinute / 60) + scheduleStartHour

		if hour > 12 {
			hour -= 12
		}
		minute = (block.StartMinute % 60)
	} else {
		hour = (block.EndMinute / 60) + scheduleStartHour
		minute = (block.EndMinute % 60)
	}

	if hour > 12 {
		hour -= 12
	}

	if minute == 0 {
		return fmt.Sprintf("%d:00", hour)
	} else {
		return fmt.Sprintf("%d:%d", hour, minute)
	}

}

func weeklyMinutesToHours(totalMinutes int) string {
	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	return fmt.Sprintf("%dh %dm", hours, minutes)
}
