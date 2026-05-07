package main

import (
	"strconv"
)

func buildWeekViewData(slots []TimeSlot, readOnly bool) WeekViewData {
	selected := buildSelectedSlotSet(slots)
	rows := buildWeekViewRows(selected)

	return WeekViewData{
		Hours:              append([]int(nil), weekHours...),
		Rows:               rows,
		WeeklyTotalMinutes: calculateWeeklyMinutes(slots),
		ReadOnly:           readOnly,
	}
}

func buildWeekViewRows(selected map[string]struct{}) []WeekViewRow {
	rows := make([]WeekViewRow, 0, len(weekDays))

	for _, day := range weekDays {
		cells, totalMinutes := buildWeekViewCells(day, selected)
		rows = append(rows, WeekViewRow{
			Day:          day,
			Cells:        cells,
			TotalMinutes: totalMinutes,
		})
	}

	return rows
}

func buildWeekViewCells(day string, selected map[string]struct{}) ([]WeekViewCell, int) {
	cells := make([]WeekViewCell, 0, len(weekHours))
	totalMinutes := 0

	for _, hour := range weekHours {
		slices, selectedCount := buildWeekViewSlices(day, hour, selected)
		totalMinutes += selectedCount * minutesPerSlice

		cells = append(cells, WeekViewCell{
			Hour:   hour,
			Slices: slices,
		})
	}

	return cells, totalMinutes
}

func buildWeekViewSlices(day string, hour int, selected map[string]struct{}) ([]WeekViewSlice, int) {
	slices := make([]WeekViewSlice, 0, slicesPerHour)
	selectedCount := 0

	for i := 0; i < slicesPerHour; i++ {
		_, isSelected := selected[slotKey(day, hour, i)]
		if isSelected {
			selectedCount++
		}

		slices = append(slices, WeekViewSlice{
			Index:    i,
			Selected: isSelected,
		})
	}

	return slices, selectedCount
}

func buildSelectedSlotSet(slots []TimeSlot) map[string]struct{} {
	selected := make(map[string]struct{}, len(slots))

	for _, slot := range slots {
		selected[slotKey(slot.Day, slot.Hour, slot.Slice)] = struct{}{}
	}

	return selected
}

func slotKey(day string, hour int, slice int) string {
	return day + "|" + strconv.Itoa(hour) + "|" + strconv.Itoa(slice)
}
