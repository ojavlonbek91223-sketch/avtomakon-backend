package handler

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/avtomakon/backend/internal/storage"
)

type FileHandler struct {
	minio *storage.MinIOClient
}

func NewFileHandler(minio *storage.MinIOClient) *FileHandler {
	return &FileHandler{minio: minio}
}

// Serve — GET /files/* — MinIO obyektini backend (tunnel) orqali ko'rsatadi.
// Bu telefon localhost:9000 ga ulana olmagani uchun zarur.
func (h *FileHandler) Serve(c *fiber.Ctx) error {
	if h.minio == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "fayl xizmati mavjud emas")
	}

	objectName := strings.TrimPrefix(c.Params("*"), "/")
	if objectName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "fayl nomi kerak")
	}

	// Imzolangan vaqtinchalik R2 havolasini yaratamiz va telefonni unga yo'naltiramiz.
	url, err := h.minio.PresignedGet(context.Background(), objectName, 24*time.Hour)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "fayl topilmadi")
	}
	return c.Redirect(url, fiber.StatusFound)
}
