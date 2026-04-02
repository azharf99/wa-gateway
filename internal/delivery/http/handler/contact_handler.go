package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/gin-gonic/gin"
)

type ContactHandler struct {
	uc domain.ContactUsecase
}

func NewContactHandler(r *gin.RouterGroup, uc domain.ContactUsecase) {
	handler := &ContactHandler{
		uc: uc,
	}

	api := r.Group("/api/v1/contacts")
	{
		api.GET("/list", handler.ListContacts)
		api.POST("/create", handler.AddContact)
		api.PUT("/update/:id", handler.UpdateContact)
		api.DELETE("/delete/:id", handler.RemoveContact)
		api.POST("/import", handler.ImportCSV)
	}
}

func (h *ContactHandler) ListContacts(c *gin.Context) {
	contacts, err := h.uc.ListContacts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{
			Status:  "error",
			Message: "Gagal mengambil daftar kontak",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Daftar kontak berhasil diambil",
		Data:    contacts,
	})
}

func (h *ContactHandler) AddContact(c *gin.Context) {
	var req domain.Contact
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{Status: "error", Message: "Data tidak valid"})
		return
	}
	err := h.uc.AddContact(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{
			Status:  "error",
			Message: "Gagal menambahkan kontak",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.Response{
		Status:  "success",
		Message: "Kontak berhasil ditambahkan",
		Data:    req,
	})
}

func (h *ContactHandler) UpdateContact(c *gin.Context) {
	var req domain.Contact
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{Status: "error", Message: "Data tidak valid"})
		return
	}
	err := h.uc.UpdateContact(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{
			Status:  "error",
			Message: "Gagal memperbarui kontak",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Kontak berhasil diperbarui",
		Data:    req,
	})
}

func (h *ContactHandler) RemoveContact(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{Status: "error", Message: "Parameter ID tidak valid"})
		return
	}
	err = h.uc.RemoveContact(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{
			Status:  "error",
			Message: "Gagal menghapus kontak",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Kontak berhasil dihapus",
	})
}

func (h *ContactHandler) ImportCSV(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{Status: "error", Message: "File CSV tidak ditemukan"})
		return
	}
	defer file.Close()

	fileBytes, _ := io.ReadAll(file)
	err = h.uc.ImportFromCSV(c.Request.Context(), fileBytes)

	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{
			Status:  "error",
			Message: "Gagal mengimpor data",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Ribuan kontak berhasil diimpor ke sistem!",
	})
}
