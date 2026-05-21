package handler

import (
	"strings"

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

	obj, info, err := h.minio.Get(c.Context(), objectName)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "fayl topilmadi")
	}

	contentType := info.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderCacheControl, "public, max-age=31536000, immutable")

	// SendStream obyekt io.Closer bo'lgani uchun o'qib bo'lgach uni yopadi.
	return c.SendStream(obj, int(info.Size))
}
