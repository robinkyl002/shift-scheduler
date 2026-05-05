package main

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
