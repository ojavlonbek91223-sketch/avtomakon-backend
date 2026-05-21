package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/avtomakon/backend/internal/middleware"
	"github.com/avtomakon/backend/internal/service"
	"github.com/avtomakon/backend/internal/storage"
)

type UploadHandler struct {
	svc *service.UploadService
}

func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

// Upload — POST /uploads (multipart/form-data)
//   - "file" — fayl
//   - "purpose" — avatar | post_media | document | business_photo | message_media | product_image
func (h *UploadHandler) Upload(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}

	purpose := c.FormValue("purpose")
	if purpose == "" {
		return fiber.NewError(fiber.StatusBadRequest, "purpose majburiy")
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "fayl kerak")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer file.Close()

	mime := fileHeader.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	result, err := h.svc.Upload(c.Context(), storage.UploadInput{
		Reader:       file,
		Size:         fileHeader.Size,
		MimeType:     mime,
		OriginalName: fileHeader.Filename,
		Purpose:      purpose,
		OwnerID:      userID,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": result})
}
