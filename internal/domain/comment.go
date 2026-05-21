package domain

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID         uuid.UUID    `json:"id"`
	PostID     uuid.UUID    `json:"post_id"`
	Author     *PostAuthor  `json:"user"`
	ParentID   *uuid.UUID   `json:"parent_id,omitempty"`
	Text       string       `json:"text"`
	LikesCount int          `json:"likes_count"`
	RepliesCount int        `json:"replies_count"`
	ViewerLiked bool        `json:"viewer_liked"`
	CreatedAt  time.Time    `json:"created_at"`
}

type CreateCommentInput struct {
	Text     string  `json:"text" validate:"required,min=1,max=1000"`
	ParentID *string `json:"parent_id" validate:"omitempty,uuid"`
}
