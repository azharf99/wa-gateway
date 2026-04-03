package handler

import (
	"strconv"

	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/gin-gonic/gin"
)

type ReminderHandler struct {
	uc domain.ReminderUsecase
}

func NewReminderHandler(r *gin.RouterGroup, uc domain.ReminderUsecase) {
	handler := &ReminderHandler{uc: uc}

	group := r.Group("/api/reminders")
	{
		group.GET("/list", handler.List)
		group.POST("/create", handler.Create)
		group.PUT("/update", handler.Update)
		group.DELETE("/delete", handler.Delete)
	}
}

func (h *ReminderHandler) List(c *gin.Context) {
	data, err := h.uc.ListReminders(c.Request.Context())
	if err != nil {
		c.JSON(500, domain.Response{Status: "error", Message: err.Error()})
		return
	}
	c.JSON(200, domain.Response{Status: "success", Data: data})
}

func (h *ReminderHandler) Create(c *gin.Context) {
	var req domain.Reminder
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, domain.Response{Status: "error", Message: "Invalid payload"})
		return
	}

	if err := h.uc.AddReminder(c.Request.Context(), req); err != nil {
		c.JSON(500, domain.Response{Status: "error", Message: err.Error()})
		return
	}
	c.JSON(200, domain.Response{Status: "success", Message: "Reminder berhasil dibuat"})
}

func (h *ReminderHandler) Update(c *gin.Context) {
	var req domain.Reminder
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, domain.Response{Status: "error", Message: "Invalid payload"})
		return
	}
	if err := h.uc.UpdateReminder(c.Request.Context(), req); err != nil {
		c.JSON(500, domain.Response{Status: "error", Message: err.Error()})
		return
	}
	c.JSON(200, domain.Response{Status: "success", Message: "Reminder berhasil diperbarui"})
}

func (h *ReminderHandler) Delete(c *gin.Context) {
	idStr := c.Query("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	if err := h.uc.DeleteReminder(c.Request.Context(), id); err != nil {
		c.JSON(500, domain.Response{Status: "error", Message: err.Error()})
		return
	}
	c.JSON(200, domain.Response{Status: "success", Message: "Reminder dihapus"})
}
