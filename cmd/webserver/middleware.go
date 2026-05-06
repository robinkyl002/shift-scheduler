package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
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

func validateSchedule(schedule ScheduleSubmission) ValidationResult {
	// calculateDailyMinutes()
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
		// shiftBlocks := buildShiftBlocks(day, daySlots)

		// for _, block := range shiftBlocks {
		// 	if block.SlotCount*10 < 180 {
		// 		startTime := formTimeFromBlock(block, true)
		// 		endTime := formTimeFromBlock(block, false)
		// 		result.Errors = append(result.Errors, "Shift from "+startTime+" to "+endTime+" is too short")
		// 	}
		// }

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
	return (t.Hour-8)*60 + (t.Slice * 10)
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
	shiftBlocks := []ShiftBlock{}

	currShiftBlock := ShiftBlock{
		Day:         day,
		StartMinute: -1,
		EndMinute:   -1,
		SlotCount:   0,
	}
	for _, slot := range slots {
		log.Print(currShiftBlock)
		log.Print(slot)
	}

	return shiftBlocks
}

func calculateDayMinutes(slots []TimeSlot) int {
	return len(slots) * 10
}

func calculateWeeklyMinutes(slots []TimeSlot) int {
	return len(slots) * 10
}

func formTimeFromBlock(block ShiftBlock, start bool) string {
	var hour int
	var minute int
	if start {
		hour = (block.StartMinute / 10) + 8
		minute = (block.StartMinute % 10)
	} else {
		hour = (block.EndMinute / 10) + 8
		minute = (block.EndMinute % 10)
	}

	return fmt.Sprintf("%d:%d", hour, minute)
}

func weeklyMinutesToHours(totalMinutes int) string {
	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	return fmt.Sprintf("%dh %dm", hours, minutes)
}
