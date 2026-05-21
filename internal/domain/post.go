package domain

import (
	"time"

	"github.com/google/uuid"
)

type PostMediaType string

const (
	MediaTypeImage    PostMediaType = "image"
	MediaTypeVideo    PostMediaType = "video"
	MediaTypeCarousel PostMediaType = "carousel"
)

type PostVisibility string

const (
	VisibilityPublic    PostVisibility = "public"
	VisibilityFollowers PostVisibility = "followers"
	VisibilityPrivate   PostVisibility = "private"
)

type Post struct {
	ID            uuid.UUID      `json:"id"`
	Author        *PostAuthor    `json:"author"`
	Caption       *string        `json:"caption,omitempty"`
	MediaType     PostMediaType  `json:"media_type"`
	CoverURL      *string        `json:"cover_url,omitempty"`
	Media         []PostMedia    `json:"media"`
	Hashtags      []string       `json:"hashtags"`
	LocationName  *string        `json:"location_name,omitempty"`
	Stats         PostStats      `json:"stats"`
	Viewer        *PostViewer    `json:"viewer,omitempty"`
	Visibility    PostVisibility `json:"visibility"`
	CreatedAt     time.Time      `json:"created_at"`
}

type PostAuthor struct {
	ID         uuid.UUID `json:"id"`
	Username   *string   `json:"username,omitempty"`
	FullName   string    `json:"full_name"`
	AvatarURL  *string   `json:"avatar_url,omitempty"`
	IsVerified bool      `json:"is_verified"`
	IsBusiness bool      `json:"is_business"`
}

type PostMedia struct {
	URL             string  `json:"url"`
	ThumbnailURL    *string `json:"thumbnail_url,omitempty"`
	Type            string  `json:"type"`
	DurationSeconds *int    `json:"duration_seconds,omitempty"`
	Width           *int    `json:"width,omitempty"`
	Height          *int    `json:"height,omitempty"`
	OrderIndex      int     `json:"-"`
}

type PostStats struct {
	Reactions  int `json:"reactions"` // 4 ta turdagi reaksiyaning umumiy soni
	ThumbsUp   int `json:"thumbs_up"`
	OK         int `json:"ok"`
	Handshake  int `json:"handshake"`
	ThumbsDown int `json:"thumbs_down"`
	Comments   int `json:"comments"`
	Saves      int `json:"saves"`
	Shares     int `json:"shares"`
}

type PostViewer struct {
	Reaction        *ReactionType `json:"reaction,omitempty"`
	Saved           bool          `json:"saved"`
	FollowingAuthor bool          `json:"following_author"`
}

type ReactionType string

const (
	ReactionThumbsUp   ReactionType = "thumbs_up"
	ReactionOK         ReactionType = "ok"
	ReactionHandshake  ReactionType = "handshake"
	ReactionThumbsDown ReactionType = "thumbs_down"
)

type SetReactionInput struct {
	Reaction ReactionType `json:"reaction" validate:"required,oneof=thumbs_up ok handshake thumbs_down"`
}

type ReactionResult struct {
	Reaction       *ReactionType `json:"reaction,omitempty"`
	ReactionsCount int           `json:"reactions_count"`
	ThumbsUp       int           `json:"thumbs_up"`
	OK             int           `json:"ok"`
	Handshake      int           `json:"handshake"`
	ThumbsDown     int           `json:"thumbs_down"`
}

// CreatePostInput
type CreatePostInput struct {
	Caption    string             `json:"caption" validate:"omitempty,max=2000"`
	MediaType  PostMediaType      `json:"media_type" validate:"required,oneof=image video carousel"`
	Media      []CreateMediaInput `json:"media" validate:"required,min=1,max=10,dive"`
	Hashtags   []string           `json:"hashtags" validate:"omitempty,max=20,dive,min=1,max=50"`
	Location   *PostLocationInput `json:"location,omitempty"`
	Visibility PostVisibility     `json:"visibility" validate:"omitempty,oneof=public followers private"`
}

type CreateMediaInput struct {
	URL          string  `json:"url" validate:"required,url"`
	ThumbnailURL *string `json:"thumbnail_url,omitempty"`
	Type         string  `json:"type" validate:"required,oneof=image video"`
	Order        int     `json:"order"`
	Width        *int    `json:"width,omitempty"`
	Height       *int    `json:"height,omitempty"`
}

type PostLocationInput struct {
	Name string  `json:"name" validate:"required,max=200"`
	Lat  float64 `json:"lat" validate:"required,latitude"`
	Lng  float64 `json:"lng" validate:"required,longitude"`
}

// FeedKind — qaysi feed turi.
type FeedKind string

const (
	FeedForYou    FeedKind = "for_you"
	FeedFollowing FeedKind = "following"
	FeedTrending  FeedKind = "trending"
)

// FeedParams — feed so'rovi parametrlari.
type FeedParams struct {
	Kind   FeedKind
	Cursor *time.Time
	Limit  int
}

// FeedResult — cursor pagination natijasi.
type FeedResult struct {
	Posts      []*Post    `json:"data"`
	NextCursor *time.Time `json:"-"`
	HasMore    bool       `json:"-"`
}

// Xato turlari
var (
	ErrPostNotFound = postError("post topilmadi")
	ErrNotPostOwner = postError("siz post egasi emassiz")
)

type postError string

func (e postError) Error() string { return string(e) }
