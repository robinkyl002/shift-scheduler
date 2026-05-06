package main

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UserFile struct {
	Users []User `json:"users"`
}

type TemplateData struct {
	IsAuthenticated bool
	Username        string
	Role            string
	Schedule        SchedulePageData
}

type ScheduleStatus string

const (
	StatusDraft    ScheduleStatus = "draft"
	StatusPending  ScheduleStatus = "pending"
	StatusApproved ScheduleStatus = "approved"
	StatusRejected ScheduleStatus = "rejected"
)

type TimeSlot struct {
	Day   string `json:"day"`
	Hour  int    `json:"hour"`
	Slice int    `json:"slice"`
}

type ScheduleSubmission struct {
	Username         string         `json:"username"`
	Slots            []TimeSlot     `json:"slots"`
	Status           ScheduleStatus `json:"status"`
	RejectionComment string         `json:"rejection_comment,omitempty"`
	SubmittedAt      string         `json:"submitted_at"`
	UpdatedAt        string         `json:"updated_at"`
}

type ScheduleFile struct {
	Schedules []ScheduleSubmission `json:"schedules"`
}
