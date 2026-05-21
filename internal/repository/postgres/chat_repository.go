package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type ChatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool}
}

// ListConversations — foydalanuvchining barcha suhbatlari, oxirgi xabari va o'qilmaganlar soni bilan.
func (r *ChatRepository) ListConversations(ctx context.Context, userID uuid.UUID, filter string) ([]*domain.Conversation, error) {
	// Filter hozircha barchasini qaytaradi (kelajakda tag/kategoriya bo'lishi mumkin).
	const query = `
		SELECT c.id, c.type, c.created_at, c.last_message_at,
		       other.id, other.full_name, other.username, other.avatar_url,
		       other.is_verified, other.is_business,
		       CASE WHEN other.last_active_at > NOW() - INTERVAL '2 minutes' THEN TRUE ELSE FALSE END,
		       lm.id, lm.sender_id, lm.type, lm.text, lm.media_url, lm.created_at,
		       COALESCE((
		         SELECT COUNT(*) FROM messages m2
		         JOIN conversation_members cm2 ON cm2.conversation_id = m2.conversation_id AND cm2.user_id = $1
		         WHERE m2.conversation_id = c.id
		           AND m2.sender_id <> $1
		           AND m2.deleted_at IS NULL
		           AND (cm2.last_read_message_id IS NULL OR m2.created_at > (
		             SELECT created_at FROM messages WHERE id = cm2.last_read_message_id
		           ))
		       ), 0)
		FROM conversations c
		JOIN conversation_members me ON me.conversation_id = c.id AND me.user_id = $1
		JOIN conversation_members om ON om.conversation_id = c.id AND om.user_id <> $1
		JOIN users other ON other.id = om.user_id
		LEFT JOIN messages lm ON lm.id = c.last_message_id
		WHERE other.deleted_at IS NULL
		ORDER BY c.last_message_at DESC NULLS LAST
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []*domain.Conversation
	for rows.Next() {
		conv := &domain.Conversation{OtherUser: &domain.ChatPartner{}}

		var lmID *uuid.UUID
		var lmSenderID *uuid.UUID
		var lmType *domain.MessageType
		var lmText, lmMediaURL *string
		var lmCreatedAt *struct{}

		err := rows.Scan(
			&conv.ID, &conv.Type, &conv.CreatedAt, &conv.LastMessageAt,
			&conv.OtherUser.ID, &conv.OtherUser.FullName, &conv.OtherUser.Username,
			&conv.OtherUser.AvatarURL, &conv.OtherUser.IsVerified,
			&conv.OtherUser.IsBusiness, &conv.OtherUser.IsOnline,
			&lmID, &lmSenderID, &lmType, &lmText, &lmMediaURL, &lmCreatedAt,
			&conv.UnreadCount,
		)
		_ = lmCreatedAt
		if err != nil {
			return nil, err
		}

		if lmID != nil && conv.LastMessageAt != nil {
			conv.LastMessage = &domain.Message{
				ID:             *lmID,
				ConversationID: conv.ID,
				SenderID:       *lmSenderID,
				Type:           *lmType,
				Text:           lmText,
				MediaURL:       lmMediaURL,
				CreatedAt:      *conv.LastMessageAt,
			}
		}

		convs = append(convs, conv)
	}

	return convs, rows.Err()
}

// FindOrCreateDirect — ikki foydalanuvchi o'rtasidagi suhbatni topadi yoki yaratadi.
func (r *ChatRepository) FindOrCreateDirect(ctx context.Context, userA, userB uuid.UUID) (uuid.UUID, error) {
	if userA == userB {
		return uuid.Nil, errors.New("o'zingiz bilan suhbat ocha olmaysiz")
	}

	// Mavjudligini tekshirish
	const findQuery = `
		SELECT c.id FROM conversations c
		WHERE c.type = 'direct'
		  AND EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id = c.id AND user_id = $1)
		  AND EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id = c.id AND user_id = $2)
		LIMIT 1
	`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, findQuery, userA, userB).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}

	// Yangi yaratish
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO conversations (type) VALUES ('direct') RETURNING id`).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO conversation_members (conversation_id, user_id)
		 VALUES ($1, $2), ($1, $3)`,
		id, userA, userB)
	if err != nil {
		return uuid.Nil, err
	}

	return id, tx.Commit(ctx)
}

// IsMember — foydalanuvchi shu suhbat a'zosimi?
func (r *ChatRepository) IsMember(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM conversation_members
		 WHERE conversation_id = $1 AND user_id = $2)`,
		conversationID, userID).Scan(&exists)
	return exists, err
}

func (r *ChatRepository) Messages(ctx context.Context, conversationID uuid.UUID, limit int) ([]*domain.Message, error) {
	const query = `
		SELECT id, conversation_id, sender_id, type, text, media_url, created_at
		FROM messages
		WHERE conversation_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*domain.Message
	for rows.Next() {
		m := &domain.Message{}
		err := rows.Scan(
			&m.ID, &m.ConversationID, &m.SenderID, &m.Type,
			&m.Text, &m.MediaURL, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}

	// Eng eski boshida bo'lishi uchun reverse
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, rows.Err()
}

func (r *ChatRepository) SendMessage(ctx context.Context, conversationID, senderID uuid.UUID, in domain.SendMessageInput) (*domain.Message, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	m := &domain.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Type:           in.Type,
	}
	if in.Text != "" {
		m.Text = &in.Text
	}
	if in.MediaURL != "" {
		m.MediaURL = &in.MediaURL
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, sender_id, type, text, media_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, conversationID, senderID, in.Type, m.Text, m.MediaURL).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE conversations SET last_message_id = $1, last_message_at = NOW()
		WHERE id = $2
	`, m.ID, conversationID)
	if err != nil {
		return nil, err
	}

	return m, tx.Commit(ctx)
}

func (r *ChatRepository) MarkRead(ctx context.Context, conversationID, userID, messageID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE conversation_members SET last_read_message_id = $1
		WHERE conversation_id = $2 AND user_id = $3
	`, messageID, conversationID, userID)
	return err
}

// OtherMember — direct suhbatdagi boshqa a'zo (real-time bildirishnoma uchun).
func (r *ChatRepository) OtherMember(ctx context.Context, conversationID, userID uuid.UUID) (uuid.UUID, error) {
	var other uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT user_id FROM conversation_members
		WHERE conversation_id = $1 AND user_id <> $2 LIMIT 1
	`, conversationID, userID).Scan(&other)
	return other, err
}
