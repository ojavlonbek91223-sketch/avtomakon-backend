package domain

import (
	"time"

	"github.com/google/uuid"
)

type ConversationType string

const (
	ConversationDirect ConversationType = "direct"
	ConversationGroup  ConversationType = "group"
)

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeFile     MessageType = "file"
	MessageTypeLocation MessageType = "location"
	MessageTypeProduct  MessageType = "product"
	MessageTypeSystem   MessageType = "system"
)

type Conversation struct {
	ID             uuid.UUID         `json:"id"`
	Type           ConversationType  `json:"type"`
	OtherUser      *ChatPartner      `json:"other_user,omitempty"`
	LastMessage    *Message          `json:"last_message,omitempty"`
	UnreadCount    int               `json:"unread_count"`
	CreatedAt      time.Time         `json:"created_at"`
	LastMessageAt  *time.Time        `json:"last_message_at,omitempty"`
}

type ChatPartner struct {
	ID         uuid.UUID  `json:"id"`
	FullName   string     `json:"full_name"`
	Username   *string    `json:"username,omitempty"`
	AvatarURL  *string    `json:"avatar_url,omitempty"`
	IsVerified bool       `json:"is_verified"`
	IsBusiness bool       `json:"is_business"`
	IsOnline   bool       `json:"is_online"`
}

type Message struct {
	ID              uuid.UUID   `json:"id"`
	ConversationID  uuid.UUID   `json:"conversation_id"`
	SenderID        uuid.UUID   `json:"sender_id"`
	Sender          *ChatPartner `json:"sender,omitempty"`
	Type            MessageType `json:"type"`
	Text            *string     `json:"text,omitempty"`
	MediaURL        *string     `json:"media_url,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
}

type SendMessageInput struct {
	Type     MessageType `json:"type" validate:"required,oneof=text image file location product"`
	Text     string      `json:"text" validate:"omitempty,max=4000"`
	MediaURL string      `json:"media_url" validate:"omitempty,url"`
}

type StartConversationInput struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}
