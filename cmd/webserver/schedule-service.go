package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

var scheduleMu sync.Mutex

func loadSchedules() (ScheduleFile, error) {
	scheduleFile, err := os.ReadFile("schedules.json")
	if err != nil {
		return ScheduleFile{}, err
	}

	var schedules ScheduleFile
	err = json.Unmarshal(scheduleFile, &schedules)
	if err != nil {
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

func findScheduleByUsername(schedules []ScheduleSubmission, username string) (*ScheduleSubmission, bool) {
	for i := range schedules {
		if schedules[i].Username == username {
			return &schedules[i], true
		}
	}

	return nil, false
}

func upsertScheduleAtIndex(schedules *ScheduleFile, index int, submission ScheduleSubmission) {
	// found := findScheduleIndex(schedules.Schedules, submission.Username)

	if index >= 0 {
		schedules.Schedules[index] = submission
	} else {
		schedules.Schedules = append(schedules.Schedules, submission)
	}
}

func buildScheduleSubmission(
	username string,
	slots []TimeSlot,
	existing *ScheduleSubmission,
	now time.Time,
) ScheduleSubmission {
	submission := ScheduleSubmission{
		Username:         username,
		Slots:            slots,
		Status:           StatusPending,
		RejectionComment: "",
		UpdatedAt:        now.String(),
	}

	if existing != nil {
		submission.SubmittedAt = existing.SubmittedAt
	} else {
		submission.SubmittedAt = now.String()
	}

	return submission
}
