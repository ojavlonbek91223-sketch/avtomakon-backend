package handler

import (
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/avtomakon/backend/internal/middleware"
	"github.com/avtomakon/backend/internal/service"
	"github.com/avtomakon/backend/internal/storage"
)

// detectMime — fayl turini MAZMUNIDAN (magic bytes) aniqlaydi. Telefon
// HEIC/turli formatlarni yuborganda, mijoz yuborgan Content-Type ishonchsiz.
func detectMime(data []byte, fallback string) string {
	if len(data) >= 12 && string(data[4:8]) == "ftyp" {
		switch string(data[8:12]) {
		case "heic", "heix", "hevc", "hevx", "mif1", "msf1", "heim", "heis":
			return "image/heic"
		case "avif":
			return "image/avif"
		case "qt  ":
			return "video/quicktime"
		default:
			return "video/mp4" // isom, mp41, mp42, M4V, 3gp, dash...
		}
	}
	ct := http.DetectContentType(data) // jpeg, png, gif, webp, pdf...
	if ct != "application/octet-stream" {
		// "image/jpeg; charset=..." bo'lishi mumkin — faqat turini olamiz
		if i := indexByte(ct, ';'); i >= 0 {
			ct = ct[:i]
		}
		return ct
	}
	if fallback == "" {
		return "application/octet-stream"
	}
	return fallback
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

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

	// Fayl turini MAZMUNIDAN aniqlaymiz (telefon HEIC/turli formatlar uchun)
	head := make([]byte, 512)
	hn, _ := file.Read(head)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "fayl o'qishda xato")
	}
	mime := detectMime(head[:hn], fileHeader.Header.Get("Content-Type"))

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
