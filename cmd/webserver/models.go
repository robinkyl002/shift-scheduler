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

	CurrentSubmission     *ScheduleSubmission
	CurrentScheduleStatus ScheduleStatus
	ValidationErrors      []string

	PendingSchedules        []PendingScheduleSummary
	SelectedPendingSchedule *ScheduleSubmission

	WeekView WeekViewData

	AdminReviewErrors  []string
	AdminReviewComment string
	ShowRejectFields   bool
	AdminReviewStatus  string
}

type PendingScheduleSummary struct {
	Username    string
	SubmittedAt string
	Status      ScheduleStatus
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

type ShiftBlock struct {
	Day         string
	StartMinute int
	EndMinute   int
	SlotCount   int
}

var dayOrder = map[string]int{
	"Mon": 0,
	"Tue": 1,
	"Wed": 2,
	"Thu": 3,
	"Fri": 4,
}

type ValidationResult struct {
	Errors        []string
	DailyMinutes  map[string]int
	WeeklyMinutes int
}

type WeekViewData struct {
	Hours              []int
	Rows               []WeekViewRow
	WeeklyTotalMinutes int
	ReadOnly           bool
}

type WeekViewRow struct {
	Day          string
	Cells        []WeekViewCell
	TotalMinutes int
}

type WeekViewCell struct {
	Hour   int
	Slices []WeekViewSlice
}

type WeekViewSlice struct {
	Index    int
	Selected bool
}
