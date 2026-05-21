package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/pkg/hash"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

const (
	otpLength          = 4
	otpTTL             = 5 * time.Minute
	otpResendCooldown  = 60 * time.Second
	otpMaxAttempts     = 5
)

type OTPService struct {
	repo    *postgres.OTPRepository
	sms     SMSProvider
	isDev   bool
}

func NewOTPService(repo *postgres.OTPRepository, sms SMSProvider, isDev bool) *OTPService {
	return &OTPService{repo: repo, sms: sms, isDev: isDev}
}

// Send — OTP yaratadi, hash qiladi, DB ga yozadi, SMS yuboradi.
// Dev rejimida kodning o'zini qaytaradi (test qilish uchun).
func (s *OTPService) Send(ctx context.Context, phone string, purpose domain.OTPPurpose) (devCode string, err error) {
	// Rate limit: oxirgi OTP'dan keyin >= 60 soniya o'tgan bo'lsin.
	last, err := s.repo.LastSentAt(ctx, phone, purpose)
	if err != nil {
		return "", err
	}
	if last != nil && time.Since(*last) < otpResendCooldown {
		remaining := otpResendCooldown - time.Since(*last)
		return "", fmt.Errorf("%w (kuting: %.0fs)", domain.ErrOTPTooManyAttempts, remaining.Seconds())
	}

	code, err := generateNumericCode(otpLength)
	if err != nil {
		return "", err
	}

	codeHash, err := hash.HashPassword(code)
	if err != nil {
		return "", err
	}

	err = s.repo.Create(ctx, postgres.CreateOTPInput{
		Identifier: phone,
		CodeHash:   codeHash,
		Purpose:    purpose,
		ExpiresAt:  time.Now().Add(otpTTL),
	})
	if err != nil {
		return "", err
	}

	message := fmt.Sprintf("AvtoMakon tasdiqlash kodi: %s. Hech kimga aytmang.", code)
	if err := s.sms.Send(ctx, phone, message); err != nil {
		return "", err
	}

	if s.isDev {
		return code, nil
	}
	return "", nil
}

// Verify — telefon va kodni tekshiradi. Muvaffaqiyatli bo'lsa OTP'ni used qilib qo'yadi.
func (s *OTPService) Verify(ctx context.Context, phone, code string, purpose domain.OTPPurpose) error {
	id, codeHash, attempts, expiresAt, err := s.repo.FindLatestActive(ctx, phone, purpose)
	if err != nil {
		return err
	}

	if time.Now().After(expiresAt) {
		return domain.ErrOTPExpired
	}

	if attempts >= otpMaxAttempts {
		return domain.ErrOTPTooManyAttempts
	}

	ok, err := hash.VerifyPassword(code, codeHash)
	if err != nil {
		return err
	}

	if !ok {
		_ = s.repo.IncrementAttempts(ctx, id)
		return domain.ErrOTPInvalid
	}

	return s.repo.MarkUsed(ctx, id)
}

func generateNumericCode(length int) (string, error) {
	max := big.NewInt(10)
	out := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = byte('0' + n.Int64())
	}
	return string(out), nil
}
