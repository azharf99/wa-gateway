package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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
		api.POST("/media", handler.ScheduleMedia) // Tambahkan rute ini
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

func (h *SchedulerHandler) ScheduleMedia(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{Status: "error", Message: "File tidak ditemukan"})
		return
	}
	defer file.Close()

	to := c.PostForm("to")
	caption := c.PostForm("caption")
	mediaType := c.PostForm("media_type")
	runAt := c.PostForm("run_at")
	isGroupStr := c.PostForm("is_group")

	isGroup, _ := strconv.ParseBool(isGroupStr)

	if mediaType == "" {
		mediaType = "document"
	}
	if runAt == "" {
		c.JSON(http.StatusBadRequest, domain.Response{Status: "error", Message: "Waktu penjadwalan (run_at) wajib diisi"})
		return
	}

	mimeType := header.Header.Get("Content-Type")

	// Validasi Defensive Programming
	if mediaType == "image" && !strings.HasPrefix(mimeType, "image/") {
		c.JSON(http.StatusBadRequest, domain.Response{
			Status:  "error",
			Message: fmt.Sprintf("Validasi Gagal: Tipe diset 'image', tapi file berupa '%s'", mimeType),
		})
		return
	}
	if mediaType == "video" && !strings.HasPrefix(mimeType, "video/") {
		c.JSON(http.StatusBadRequest, domain.Response{
			Status:  "error",
			Message: fmt.Sprintf("Validasi Gagal: Tipe diset 'video', tapi file berupa '%s'", mimeType),
		})
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{Status: "error", Message: "Gagal membaca file"})
		return
	}

	req := domain.ScheduleMediaReq{
		To:        to,
		IsGroup:   isGroup,
		FileBytes: fileBytes,
		FileName:  header.Filename,
		MimeType:  mimeType,
		Caption:   caption,
		MediaType: mediaType,
		RunAt:     runAt,
	}

	if err := h.uc.ScheduleMedia(req); err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{
			Status:  "error",
			Message: "Gagal menjadwalkan media",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Media berhasil dijadwalkan",
		Data:    map[string]string{"run_at": req.RunAt, "file": req.FileName},
	})
}
