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

func NewMinIOClient(endpoint, accessKey, secretKey, bucket string, useSSL bool, region, publicURL string) (*MinIOClient, error) {
	// Cloudflare R2 doim "auto" region talab qiladi — aks holda minio-go
	// GetBucketLocation/HeadBucket'da "Access Denied" oladi.
	if strings.Contains(endpoint, "r2.cloudflarestorage.com") {
		region = "auto"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	// Bucket tekshiruvi — XATO BO'LSA HAM davom etamiz. Ba'zi provayderlar/tokenlar
	// bucket-darajali amalni (HeadBucket) cheklaydi, lekin obyekt yuklash ishlaydi.
	// Bucket allaqachon yaratilgan deb hisoblaymiz.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if exists, berr := client.BucketExists(ctx, bucket); berr == nil && !exists {
		if merr := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); merr == nil {
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

// PresignedGet — obyekt uchun vaqtinchalik imzolangan URL qaytaradi.
// Telefon shu URL orqali to'g'ridan-to'g'ri R2'dan oladi (backend trafigi tejaladi,
// R2 HEAD/Stat cheklovi chetlab o'tiladi).
func (m *MinIOClient) PresignedGet(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	u, err := m.client.PresignedGetObject(ctx, m.bucket, objectName, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
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
