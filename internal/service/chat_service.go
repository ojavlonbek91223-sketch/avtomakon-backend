package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
	ws "github.com/avtomakon/backend/internal/websocket"
)

type ChatService struct {
	repo *postgres.ChatRepository
	hub  *ws.Hub
}

func NewChatService(repo *postgres.ChatRepository, hub *ws.Hub) *ChatService {
	return &ChatService{repo: repo, hub: hub}
}

func (s *ChatService) ListConversations(ctx context.Context, userID uuid.UUID, filter string) ([]*domain.Conversation, error) {
	convs, err := s.repo.ListConversations(ctx, userID, filter)
	if err != nil {
		return nil, err
	}
	if convs == nil {
		convs = []*domain.Conversation{}
	}
	return convs, nil
}

func (s *ChatService) StartConversation(ctx context.Context, userA uuid.UUID, otherID string) (uuid.UUID, error) {
	userB, err := uuid.Parse(otherID)
	if err != nil {
		return uuid.Nil, errors.New("noto'g'ri user ID")
	}
	return s.repo.FindOrCreateDirect(ctx, userA, userB)
}

func (s *ChatService) Messages(ctx context.Context, conversationID, userID uuid.UUID, limit int) ([]*domain.Message, error) {
	ok, err := s.repo.IsMember(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("siz bu suhbatga kira olmaysiz")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	msgs, err := s.repo.Messages(ctx, conversationID, limit)
	if err != nil {
		return nil, err
	}
	if msgs == nil {
		msgs = []*domain.Message{}
	}
	return msgs, nil
}

func (s *ChatService) SendMessage(ctx context.Context, conversationID, senderID uuid.UUID, in domain.SendMessageInput) (*domain.Message, error) {
	ok, err := s.repo.IsMember(ctx, conversationID, senderID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("siz bu suhbat a'zosi emassiz")
	}
	in.Text = strings.TrimSpace(in.Text)
	if in.Type == domain.MessageTypeText && in.Text == "" {
		return nil, errors.New("xabar matni bo'sh")
	}

	msg, err := s.repo.SendMessage(ctx, conversationID, senderID, in)
	if err != nil {
		return nil, err
	}

	// Boshqa a'zoga real-time yuborish
	other, err := s.repo.OtherMember(ctx, conversationID, senderID)
	if err == nil && other != uuid.Nil {
		s.hub.SendToUser(other, ws.Event{
			Event: "message.new",
			Data: map[string]any{
				"conversation_id": conversationID,
				"message":         msg,
			},
		})
	}

	return msg, nil
}

func (s *ChatService) MarkRead(ctx context.Context, conversationID, userID, messageID uuid.UUID) error {
	return s.repo.MarkRead(ctx, conversationID, userID, messageID)
}
