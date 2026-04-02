package domain

// Format waktu yang diharapkan dari JSON: "2026-04-01 15:30:00"
type ScheduleReq struct {
	To      string `json:"to" binding:"required"`
	IsGroup bool   `json:"is_group" default:"false"`
	Message string `json:"message" binding:"required"`
	RunAt   string `json:"run_at" binding:"required"`
}

type SchedulerUsecase interface {
	Start()
	Stop()
	ScheduleMessage(req ScheduleReq) error
	ScheduleMedia(req ScheduleMediaReq) error // Kontrak baru
}

type ScheduleMediaReq struct {
	To        string
	IsGroup   bool
	FileBytes []byte
	FileName  string
	MimeType  string
	Caption   string
	MediaType string // "document", "image", "video"
	RunAt     string // Format: "YYYY-MM-DD HH:MM:SS"
}
