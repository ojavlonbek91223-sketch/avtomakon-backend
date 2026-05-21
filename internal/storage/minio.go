package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client *minio.Client
	bucket string
	public string // public URL prefix (CDN bo'lsa, shu ishlatiladi)
}

func NewMinIOClient(endpoint, accessKey, secretKey, bucket string, useSSL bool, publicURL string) (*MinIOClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	// Bucket mavjudligini tekshirish va yaratish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("bucket create: %w", err)
		}
		// Public read policy
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, bucket)
		_ = client.SetBucketPolicy(ctx, bucket, policy)
	}

	// Public URL prefiksi: backend proxy (tunnel orqali ishlaydi) yoki
	// to'g'ridan-to'g'ri MinIO endpoint (faqat lokal tarmoq).
	public := strings.TrimRight(publicURL, "/") + "/files"
	if publicURL == "" {
		scheme := "http"
		if useSSL {
			scheme = "https"
		}
		public = fmt.Sprintf("%s://%s/%s", scheme, endpoint, bucket)
	}

	return &MinIOClient{
		client: client,
		bucket: bucket,
		public: public,
	}, nil
}

// Get — obyektni MinIO'dan o'qiydi (backend proxy uchun).
// Qaytarilgan obyektni chaqiruvchi yopishi shart.
func (m *MinIOClient) Get(ctx context.Context, objectName string) (*minio.Object, minio.ObjectInfo, error) {
	obj, err := m.client.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, minio.ObjectInfo{}, err
	}
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, minio.ObjectInfo{}, err
	}
	return obj, info, nil
}

type UploadInput struct {
	Reader      io.Reader
	Size        int64
	MimeType    string
	OriginalName string
	Purpose     string
	OwnerID     uuid.UUID
}

type UploadResult struct {
	URL      string
	FileName string
}

// Upload — faylni MinIO ga yuklaydi va public URL qaytaradi.
func (m *MinIOClient) Upload(ctx context.Context, in UploadInput) (*UploadResult, error) {
	ext := extensionFromMime(in.MimeType, in.OriginalName)
	name := fmt.Sprintf("%s/%s/%s%s",
		in.Purpose,
		in.OwnerID.String(),
		uuid.NewString(),
		ext,
	)

	_, err := m.client.PutObject(ctx, m.bucket, name, in.Reader, in.Size, minio.PutObjectOptions{
		ContentType: in.MimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("upload: %w", err)
	}

	return &UploadResult{
		URL:      m.public + "/" + name,
		FileName: name,
	}, nil
}

func extensionFromMime(mime, originalName string) string {
	switch strings.ToLower(mime) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	case "application/pdf":
		return ".pdf"
	}
	// Fallback — original fayl nomidan
	if i := strings.LastIndex(originalName, "."); i >= 0 {
		return originalName[i:]
	}
	return ""
}
