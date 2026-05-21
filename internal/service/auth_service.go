package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/pkg/hash"
	"github.com/avtomakon/backend/internal/pkg/jwt"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

type AuthService struct {
	users    *postgres.UserRepository
	refresh  *postgres.RefreshTokenRepository
	otp      *OTPService
	jwt      *jwt.Manager
	refreshTTL time.Duration
}

func NewAuthService(
	users *postgres.UserRepository,
	refresh *postgres.RefreshTokenRepository,
	otp *OTPService,
	jwt *jwt.Manager,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users: users, refresh: refresh, otp: otp,
		jwt: jwt, refreshTTL: refreshTTL,
	}
}

func (s *AuthService) SendOTP(ctx context.Context, in domain.SendOTPInput) (*domain.SendOTPResult, error) {
	devCode, err := s.otp.Send(ctx, in.Phone, in.Purpose)
	if err != nil {
		return nil, err
	}

	return &domain.SendOTPResult{
		Message:    "Kod yuborildi",
		ExpiresIn:  int(otpTTL.Seconds()),
		RetryAfter: int(otpResendCooldown.Seconds()),
		DevCode:    devCode,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, in domain.RegisterInput, ip string) (*domain.AuthResult, error) {
	// OTP olib tashlandi — bir bosqichli ro'yxat (username + parol + telefon)

	// 1. Parol hash
	passwordHash, err := hash.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	// 2. User yaratish (telefon avtomatik tasdiqlangan, dev rejimi)
	now := time.Now()
	user, err := s.users.Create(ctx, postgres.CreateInput{
		Username:        in.Username,
		Phone:           in.Phone,
		PasswordHash:    passwordHash,
		FullName:        in.FullName,
		Language:        in.Language,
		PhoneVerifiedAt: &now,
	})
	if err != nil {
		return nil, err
	}

	// 3. Tokenlar
	tokens, err := s.issueTokens(ctx, user, ip)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResult{User: user, Tokens: tokens}, nil
}

func (s *AuthService) Login(ctx context.Context, in domain.LoginInput, ip string) (*domain.AuthResult, error) {
	user, passwordHash, err := s.users.FindByUsernameOrPhone(ctx, in.Username)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	ok, err := hash.VerifyPassword(in.Password, passwordHash)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrInvalidCredentials
	}

	tokens, err := s.issueTokens(ctx, user, ip)
	if err != nil {
		return nil, err
	}

	_ = s.users.UpdateLastActive(ctx, user.ID)
	return &domain.AuthResult{User: user, Tokens: tokens}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken, ip string) (*domain.TokenPair, error) {
	tokenHash := hashToken(refreshToken)

	rt, err := s.refresh.FindActive(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(rt.UserID)
	if err != nil {
		return nil, domain.ErrRefreshTokenInvalid
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Rotatsiya: eski refresh tokenni revoke qilamiz, yangisini beramiz.
	_ = s.refresh.Revoke(ctx, tokenHash)

	return s.issueTokens(ctx, user, ip)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.refresh.Revoke(ctx, hashToken(refreshToken))
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.refresh.RevokeAllForUser(ctx, userID)
}

// issueTokens — access + refresh token yaratadi va refresh'ni DB ga yozadi.
func (s *AuthService) issueTokens(ctx context.Context, user *domain.User, ip string) (*domain.TokenPair, error) {
	access, err := s.jwt.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	refresh, err := generateOpaqueToken()
	if err != nil {
		return nil, err
	}

	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}

	_, err = s.refresh.Create(ctx, postgres.CreateRefreshInput{
		UserID:    user.ID,
		TokenHash: hashToken(refresh),
		IPAddress: ipPtr,
		ExpiresAt: time.Now().Add(s.refreshTTL),
	})
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    900, // 15 minut
		TokenType:    "Bearer",
	}, nil
}

// generateOpaqueToken — 32 baytli random refresh token.
func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "rt_" + base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
