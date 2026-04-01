package handler

import (
	"net/http"

	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/gin-gonic/gin"
)

type SchedulerHandler struct {
	uc domain.SchedulerUsecase
}

func NewSchedulerHandler(r *gin.RouterGroup, uc domain.SchedulerUsecase) {
	handler := &SchedulerHandler{
		uc: uc,
	}

	api := r.Group("/api/v1/schedule")
	{
		api.POST("/message", handler.ScheduleMessage)
	}
}

func (h *SchedulerHandler) ScheduleMessage(c *gin.Context) {
	var req domain.ScheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{
			Status:  "error",
			Message: "Payload tidak valid",
			Data:    err.Error(),
		})
		return
	}

	if err := h.uc.ScheduleMessage(req); err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{
			Status:  "error",
			Message: "Gagal menjadwalkan pesan",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Pesan berhasil dijadwalkan",
		Data:    map[string]string{"run_at": req.RunAt},
	})
}
