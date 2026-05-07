package main

import (
	"encoding/json"
	"os"
	"sync"
)

var scheduleMu sync.Mutex

func loadSchedules() (ScheduleFile, error) {
	scheduleFile, err := os.ReadFile("schedules.json")
	if err != nil {
		// http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return ScheduleFile{}, err
	}

	var schedules ScheduleFile
	err = json.Unmarshal(scheduleFile, &schedules)
	if err != nil {
		// http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return ScheduleFile{}, err
	}

	return schedules, nil
}

func saveSchedules(schedules ScheduleFile) error {
	updated, err := json.MarshalIndent(schedules, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("schedules.json", updated, 0644)
}

func findScheduleIndex(schedules []ScheduleSubmission, username string) int {

	for i, schedule := range schedules {
		if schedule.Username == username {
			return i
		}
	}
	return -1
}

func upsertSchedule(schedules *ScheduleFile, submission ScheduleSubmission) {
	found := findScheduleIndex(schedules.Schedules, submission.Username)

	if found >= 0 {
		schedules.Schedules[found] = submission
	} else {
		schedules.Schedules = append(schedules.Schedules, submission)
	}
}
