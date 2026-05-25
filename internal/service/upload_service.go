package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/storage"
)

type UploadService struct {
	pool  *pgxpool.Pool
	minio *storage.MinIOClient
}

func NewUploadService(pool *pgxpool.Pool, minio *storage.MinIOClient) *UploadService {
	return &UploadService{pool: pool, minio: minio}
}

// Cheklovlar
const (
	maxImageSize    = 10 * 1024 * 1024  // 10 MB
	maxVideoSize    = 100 * 1024 * 1024 // 100 MB
	maxDocumentSize = 20 * 1024 * 1024  // 20 MB
)

var allowedMimes = map[string]bool{
	"image/jpeg":       true,
	"image/jpg":        true,
	"image/png":        true,
	"image/webp":       true,
	"image/gif":        true,
	"image/heic":       true, // iPhone/Samsung kameralar
	"image/heif":       true,
	"image/avif":       true,
	"video/mp4":        true,
	"video/quicktime":  true,
	"video/3gpp":       true,
	"video/webm":       true,
	"video/x-matroska": true,
	"application/pdf":  true,
}

var allowedPurposes = map[string]bool{
	"avatar":          true,
	"post_media":      true,
	"document":        true,
	"business_photo":  true,
	"message_media":   true,
	"product_image":   true,
}

type UploadResult struct {
	ID       uuid.UUID `json:"id"`
	URL      string    `json:"url"`
	MimeType string    `json:"mime_type"`
	Size     int64     `json:"size_bytes"`
}

func (s *UploadService) Upload(ctx context.Context, in storage.UploadInput) (*UploadResult, error) {
	// Validatsiya
	if !allowedPurposes[in.Purpose] {
		return nil, fmt.Errorf("noto'g'ri purpose: %s", in.Purpose)
	}
	if !allowedMimes[in.MimeType] {
		return nil, fmt.Errorf("noto'g'ri fayl turi: %s", in.MimeType)
	}

	// Hajm tekshiruvi
	switch {
	case isImage(in.MimeType) && in.Size > maxImageSize:
		return nil, errors.New("rasm 10 MB dan oshmasin")
	case isVideo(in.MimeType) && in.Size > maxVideoSize:
		return nil, errors.New("video 100 MB dan oshmasin")
	case isDocument(in.MimeType) && in.Size > maxDocumentSize:
		return nil, errors.New("hujjat 20 MB dan oshmasin")
	}

	// MinIO ga yuklash
	uploaded, err := s.minio.Upload(ctx, in)
	if err != nil {
		return nil, err
	}

	// DB ga yozish
	var id uuid.UUID
	err = s.pool.QueryRow(ctx, `
		INSERT INTO uploaded_files (owner_id, purpose, filename, mime_type, size_bytes, url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, in.OwnerID, in.Purpose, uploaded.FileName, in.MimeType, in.Size, uploaded.URL).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		ID:       id,
		URL:      uploaded.URL,
		MimeType: in.MimeType,
		Size:     in.Size,
	}, nil
}

func isImage(mime string) bool {
	return mime == "image/jpeg" || mime == "image/jpg" || mime == "image/png" ||
		mime == "image/webp" || mime == "image/gif" || mime == "image/heic" ||
		mime == "image/heif" || mime == "image/avif"
}

func isVideo(mime string) bool {
	return mime == "video/mp4" || mime == "video/quicktime" ||
		mime == "video/3gpp" || mime == "video/webm" || mime == "video/x-matroska"
}

func isDocument(mime string) bool {
	return mime == "application/pdf"
}
