package domain

// Format waktu yang diharapkan dari JSON: "2026-04-01 15:30:00"
type ScheduleReq struct {
	To      string `json:"to" binding:"required"`
	Message string `json:"message" binding:"required"`
	RunAt   string `json:"run_at" binding:"required"` 
}

type SchedulerUsecase interface {
	Start()
	Stop()
	ScheduleMessage(req ScheduleReq) error
}