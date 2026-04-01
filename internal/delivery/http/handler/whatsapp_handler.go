package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/gin-gonic/gin"
)

type WhatsAppHandler struct {
	uc domain.WhatsAppUsecase
}

// Constructor Handler meng-inject Usecase
func NewWhatsAppHandler(r *gin.RouterGroup, uc domain.WhatsAppUsecase) {
	handler := &WhatsAppHandler{
		uc: uc,
	}

	api := r.Group("/api/v1/whatsapp")
	{
		api.GET("/status", handler.Status)
		api.POST("/send", handler.SendMessage)
		api.POST("/broadcast", handler.Broadcast)
		api.POST("/media", handler.SendMedia)
		api.GET("/groups", handler.GetGroups)
	}
}

func (h *WhatsAppHandler) Status(c *gin.Context) {
	status := h.uc.CheckStatus()
	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "WhatsApp connection status",
		Data:    map[string]string{"state": status},
	})
}

func (h *WhatsAppHandler) SendMessage(c *gin.Context) {
	var req domain.SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{
			Status:  "error",
			Message: "Invalid request payload",
			Data:    err.Error(),
		})
		return
	}

	msgID, err := h.uc.SendMessage(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{
			Status:  "error",
			Message: "Failed to send message",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Message sent successfully",
		Data:    map[string]string{"message_id": msgID},
	})
}

func (h *WhatsAppHandler) Broadcast(c *gin.Context) {
	var req domain.BroadcastReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{
			Status:  "error",
			Message: "Format payload array tidak valid",
			Data:    err.Error(),
		})
		return
	}

	// Panggil usecase. Karena di dalamnya ada goroutine, ini akan langsung return.
	h.uc.BroadcastMessages(req)

	c.JSON(http.StatusAccepted, domain.Response{
		Status:  "success",
		Message: "Broadcast diterima dan sedang diproses di background",
		Data: map[string]interface{}{
			"total_recipients": len(req.Recipients),
			"estimated_time":   "Tergantung jumlah antrean (estimasi ~10 detik per pesan)",
		},
	})
}

func (h *WhatsAppHandler) SendMedia(c *gin.Context) {
	// 1. Ambil file dari request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{
			Status:  "error",
			Message: "File tidak ditemukan dalam request",
			Data:    err.Error(),
		})
		return
	}
	defer file.Close()

	// 2. Baca file ke dalam bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{
			Status:  "error",
			Message: "Gagal membaca file",
			Data:    err.Error(),
		})
		return
	}

	// 3. Ambil parameter form lainnya
	to := c.PostForm("to")
	caption := c.PostForm("caption")
	mediaType := c.PostForm("media_type") // "document", "image", "video"
	isGroupStr := c.PostForm("is_group")

	isGroup, _ := strconv.ParseBool(isGroupStr)

	if mediaType == "" {
		mediaType = "document" // Default
	}

	// 4. Susun Request Payload
	req := domain.SendMediaReq{
		To:        to,
		IsGroup:   isGroup,
		FileBytes: fileBytes,
		FileName:  header.Filename,
		MimeType:  header.Header.Get("Content-Type"),
		Caption:   caption,
		MediaType: mediaType,
	}

	// 5. Eksekusi Usecase
	msgID, err := h.uc.SendMedia(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{
			Status:  "error",
			Message: "Gagal mengirim media",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Media berhasil dikirim",
		Data:    map[string]string{"message_id": msgID},
	})
}

func (h *WhatsAppHandler) GetGroups(c *gin.Context) {
	groups, err := h.uc.GetJoinedGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{
			Status:  "error",
			Message: "Gagal mengambil daftar grup",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Berhasil mengambil daftar grup",
		Data:    groups,
	})
}
