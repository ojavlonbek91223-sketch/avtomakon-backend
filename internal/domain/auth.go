package domain

import (
	"errors"
	"time"
)

// Xato turlari
var (
	ErrUserNotFound          = errors.New("foydalanuvchi topilmadi")
	ErrPhoneAlreadyExists    = errors.New("bu telefon raqami band")
	ErrUsernameAlreadyExists = errors.New("bu username band")
	ErrEmailAlreadyExists    = errors.New("bu email band")
	ErrInvalidCredentials    = errors.New("username yoki parol noto'g'ri")
	ErrOTPInvalid            = errors.New("OTP kod noto'g'ri")
	ErrOTPExpired            = errors.New("OTP muddati o'tdi")
	ErrOTPTooManyAttempts    = errors.New("juda ko'p urinish, qayta urinib ko'ring")
	ErrRefreshTokenInvalid   = errors.New("refresh token yaroqsiz")
	ErrPhoneNotVerified      = errors.New("avval telefon raqamingizni tasdiqlang")
)

// OTPPurpose — OTP kod nima maqsadda yuborilgan.
type OTPPurpose string

const (
	OTPPurposeSignup        OTPPurpose = "signup"
	OTPPurposeLogin         OTPPurpose = "login"
	OTPPurposeResetPassword OTPPurpose = "reset_password"
	OTPPurposeVerifyPhone   OTPPurpose = "verify_phone"
)

// SendOTPInput
type SendOTPInput struct {
	Phone   string     `json:"phone" validate:"required,e164"`
	Purpose OTPPurpose `json:"purpose" validate:"required,oneof=signup login reset_password verify_phone"`
}

type SendOTPResult struct {
	Message    string `json:"message"`
	ExpiresIn  int    `json:"expires_in"`
	RetryAfter int    `json:"retry_after"`
	DevCode    string `json:"dev_code,omitempty"`
}

// RegisterInput — bir bosqichli ro'yxat (OTP siz, dev rejimi).
type RegisterInput struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Password string `json:"password" validate:"required,min=6,max=72"`
	Phone    string `json:"phone" validate:"required,e164"`
	FullName string `json:"full_name" validate:"omitempty,min=2,max=100"`
	Language string `json:"language" validate:"omitempty,oneof=uz ru en"`
}

// LoginInput — username yoki telefon orqali kirish.
type LoginInput struct {
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required"`
}

// RefreshInput
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResult — register va login uchun.
type AuthResult struct {
	User   *User      `json:"user"`
	Tokens *TokenPair `json:"tokens"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// RefreshToken — DB modeli.
type RefreshToken struct {
	ID         string
	UserID     string
	TokenHash  string
	DeviceInfo map[string]any
	IPAddress  *string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}
