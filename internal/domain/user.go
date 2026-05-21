package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser   UserRole = "user"
	RoleMaster UserRole = "master"
	RoleSeller UserRole = "seller"
	RoleAdmin  UserRole = "admin"
)

// UserBrief — followers/following ro'yxati uchun qisqa ma'lumot.
type UserBrief struct {
	ID         uuid.UUID `json:"id"`
	Username   *string   `json:"username,omitempty"`
	FullName   string    `json:"full_name"`
	AvatarURL  *string   `json:"avatar_url,omitempty"`
	IsVerified bool      `json:"is_verified"`
	IsBusiness bool      `json:"is_business"`
	Following  bool      `json:"following"`
}

type User struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Email           *string    `json:"email,omitempty" db:"email"`
	Phone           string     `json:"phone" db:"phone"`
	Username        *string    `json:"username,omitempty" db:"username"`
	FullName        string     `json:"full_name" db:"full_name"`
	AvatarURL       *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Bio             *string    `json:"bio,omitempty" db:"bio"`
	Role            UserRole   `json:"role" db:"role"`
	IsBusiness      bool       `json:"is_business" db:"is_business"`
	IsVerified      bool       `json:"is_verified" db:"is_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty" db:"phone_verified_at"`
	Language        string     `json:"language" db:"language"`
	CountryCode     *string    `json:"country_code,omitempty" db:"country_code"`
	LastLat         *float64   `json:"-" db:"last_lat"`
	LastLng         *float64   `json:"-" db:"last_lng"`
	LastActiveAt    *time.Time `json:"last_active_at,omitempty" db:"last_active_at"`
	PostsCount      int        `json:"posts_count" db:"posts_count"`
	FollowersCount  int        `json:"followers_count" db:"followers_count"`
	FollowingCount  int        `json:"following_count" db:"following_count"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}
