package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type PostRepository struct {
	pool *pgxpool.Pool
}

func NewPostRepository(pool *pgxpool.Pool) *PostRepository {
	return &PostRepository{pool: pool}
}

type CreatePostDBInput struct {
	AuthorID     uuid.UUID
	Caption      *string
	MediaType    domain.PostMediaType
	CoverURL     *string
	LocationName *string
	LocationLat  *float64
	LocationLng  *float64
	Visibility   domain.PostVisibility
	Media        []domain.CreateMediaInput
	Hashtags     []string
}

// Create — tranzaksiyada post, media va hashtag'larni yaratadi.
func (r *PostRepository) Create(ctx context.Context, in CreatePostDBInput) (uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	var postID uuid.UUID
	const postQuery = `
		INSERT INTO posts (author_id, caption, media_type, cover_url,
		                   location_name, location, visibility)
		VALUES ($1, $2, $3, $4, $5,
		        CASE WHEN $6::float8 IS NOT NULL AND $7::float8 IS NOT NULL
		             THEN ST_SetSRID(ST_MakePoint($7, $6), 4326)::geography
		             ELSE NULL END,
		        COALESCE($8, 'public')::post_visibility)
		RETURNING id
	`
	err = tx.QueryRow(ctx, postQuery,
		in.AuthorID, in.Caption, in.MediaType, in.CoverURL,
		in.LocationName, in.LocationLat, in.LocationLng, in.Visibility,
	).Scan(&postID)
	if err != nil {
		return uuid.Nil, err
	}

	// Media
	for _, m := range in.Media {
		const mq = `
			INSERT INTO post_media (post_id, url, thumbnail_url, type, width, height, order_index)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err := tx.Exec(ctx, mq, postID, m.URL, m.ThumbnailURL, m.Type, m.Width, m.Height, m.Order)
		if err != nil {
			return uuid.Nil, err
		}
	}

	// Hashtags
	for _, tag := range in.Hashtags {
		var tagID uuid.UUID
		const upsert = `
			INSERT INTO hashtags (name) VALUES (LOWER($1))
			ON CONFLICT (name) DO UPDATE SET posts_count = hashtags.posts_count + 1
			RETURNING id
		`
		if err := tx.QueryRow(ctx, upsert, tag).Scan(&tagID); err != nil {
			return uuid.Nil, err
		}
		const link = `INSERT INTO post_hashtags (post_id, hashtag_id) VALUES ($1, $2)`
		if _, err := tx.Exec(ctx, link, postID, tagID); err != nil {
			return uuid.Nil, err
		}
	}

	// Author counter
	_, err = tx.Exec(ctx, `UPDATE users SET posts_count = posts_count + 1 WHERE id = $1`, in.AuthorID)
	if err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return postID, nil
}

// ListFeed — public postlar cursor pagination bilan.
// viewerID ixtiyoriy — agar berilsa, viewer (liked/saved/following) ham yuklanadi.
func (r *PostRepository) ListFeed(ctx context.Context, kind domain.FeedKind, viewerID *uuid.UUID, cursor *time.Time, limit int) ([]*domain.Post, error) {
	args := []any{}
	where := "p.deleted_at IS NULL AND p.is_published = TRUE"

	switch kind {
	case domain.FeedFollowing:
		if viewerID == nil {
			return nil, errors.New("following feed uchun auth kerak")
		}
		where += " AND (p.visibility = 'public' OR p.visibility = 'followers') AND p.author_id IN (SELECT following_id FROM follows WHERE follower_id = $1)"
		args = append(args, *viewerID)
	default:
		where += " AND p.visibility = 'public'"
	}

	if cursor != nil {
		args = append(args, *cursor)
		where += " AND p.created_at < $" + itoa(len(args))
	}

	args = append(args, limit+1)
	limitParam := "$" + itoa(len(args))

	query := `
		SELECT p.id, p.author_id, p.caption, p.media_type, p.cover_url,
		       p.location_name, p.visibility,
		       p.thumbs_up_count + p.ok_count + p.handshake_count + p.thumbs_down_count AS reactions_total,
		       p.thumbs_up_count, p.ok_count, p.handshake_count, p.thumbs_down_count,
		       p.comments_count, p.saves_count, p.shares_count,
		       p.created_at,
		       u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business
		FROM posts p
		JOIN users u ON u.id = p.author_id AND u.deleted_at IS NULL
		WHERE ` + where + `
		ORDER BY p.created_at DESC
		LIMIT ` + limitParam

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*domain.Post
	var postIDs []uuid.UUID
	for rows.Next() {
		var p domain.Post
		var author domain.PostAuthor
		var authorID uuid.UUID
		err := rows.Scan(
			&p.ID, &authorID, &p.Caption, &p.MediaType, &p.CoverURL,
			&p.LocationName, &p.Visibility,
			&p.Stats.Reactions, &p.Stats.ThumbsUp, &p.Stats.OK,
			&p.Stats.Handshake, &p.Stats.ThumbsDown,
			&p.Stats.Comments, &p.Stats.Saves, &p.Stats.Shares,
			&p.CreatedAt,
			&author.Username, &author.FullName, &author.AvatarURL,
			&author.IsVerified, &author.IsBusiness,
		)
		if err != nil {
			return nil, err
		}
		author.ID = authorID
		p.Author = &author
		p.Hashtags = []string{}
		p.Media = []domain.PostMedia{}
		posts = append(posts, &p)
		postIDs = append(postIDs, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return posts, nil
	}

	if err := r.loadMediaForPosts(ctx, posts, postIDs); err != nil {
		return nil, err
	}
	if err := r.loadHashtagsForPosts(ctx, posts, postIDs); err != nil {
		return nil, err
	}
	if viewerID != nil {
		if err := r.loadViewerForPosts(ctx, posts, postIDs, *viewerID); err != nil {
			return nil, err
		}
	}

	return posts, nil
}

// ListVideos — faqat video media_type'li public postlar (videolar ekrani uchun).
func (r *PostRepository) ListVideos(ctx context.Context, viewerID *uuid.UUID, cursor *time.Time, limit int) ([]*domain.Post, error) {
	args := []any{}
	where := "p.deleted_at IS NULL AND p.is_published = TRUE AND p.visibility = 'public' AND p.media_type = 'video'"
	if cursor != nil {
		args = append(args, *cursor)
		where += " AND p.created_at < $" + itoa(len(args))
	}
	args = append(args, limit+1)
	limitParam := "$" + itoa(len(args))

	query := `
		SELECT p.id, p.author_id, p.caption, p.media_type, p.cover_url,
		       p.location_name, p.visibility,
		       p.thumbs_up_count + p.ok_count + p.handshake_count + p.thumbs_down_count AS reactions_total,
		       p.thumbs_up_count, p.ok_count, p.handshake_count, p.thumbs_down_count,
		       p.comments_count, p.saves_count, p.shares_count,
		       p.created_at,
		       u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business
		FROM posts p
		JOIN users u ON u.id = p.author_id AND u.deleted_at IS NULL
		WHERE ` + where + `
		ORDER BY p.created_at DESC
		LIMIT ` + limitParam

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*domain.Post
	var postIDs []uuid.UUID
	for rows.Next() {
		var p domain.Post
		var author domain.PostAuthor
		var authorID uuid.UUID
		err := rows.Scan(
			&p.ID, &authorID, &p.Caption, &p.MediaType, &p.CoverURL,
			&p.LocationName, &p.Visibility,
			&p.Stats.Reactions, &p.Stats.ThumbsUp, &p.Stats.OK,
			&p.Stats.Handshake, &p.Stats.ThumbsDown,
			&p.Stats.Comments, &p.Stats.Saves, &p.Stats.Shares,
			&p.CreatedAt,
			&author.Username, &author.FullName, &author.AvatarURL,
			&author.IsVerified, &author.IsBusiness,
		)
		if err != nil {
			return nil, err
		}
		author.ID = authorID
		p.Author = &author
		p.Hashtags = []string{}
		p.Media = []domain.PostMedia{}
		posts = append(posts, &p)
		postIDs = append(postIDs, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return posts, nil
	}

	if err := r.loadMediaForPosts(ctx, posts, postIDs); err != nil {
		return nil, err
	}
	if err := r.loadHashtagsForPosts(ctx, posts, postIDs); err != nil {
		return nil, err
	}
	if viewerID != nil {
		if err := r.loadViewerForPosts(ctx, posts, postIDs, *viewerID); err != nil {
			return nil, err
		}
	}
	return posts, nil
}

// ListByUser — bitta foydalanuvchining postlari (eng yangi birinchi).
// viewerID == userID bo'lsa barcha postlar; aks holda faqat public.
func (r *PostRepository) ListByUser(ctx context.Context, userID uuid.UUID, viewerID *uuid.UUID, cursor *time.Time, limit int) ([]*domain.Post, error) {
	args := []any{userID}
	where := "p.deleted_at IS NULL AND p.is_published = TRUE AND p.author_id = $1"
	if viewerID == nil || *viewerID != userID {
		where += " AND p.visibility = 'public'"
	}
	if cursor != nil {
		args = append(args, *cursor)
		where += " AND p.created_at < $" + itoa(len(args))
	}
	args = append(args, limit+1)
	limitParam := "$" + itoa(len(args))

	query := `
		SELECT p.id, p.author_id, p.caption, p.media_type, p.cover_url,
		       p.location_name, p.visibility,
		       p.thumbs_up_count + p.ok_count + p.handshake_count + p.thumbs_down_count AS reactions_total,
		       p.thumbs_up_count, p.ok_count, p.handshake_count, p.thumbs_down_count,
		       p.comments_count, p.saves_count, p.shares_count,
		       p.created_at,
		       u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business
		FROM posts p
		JOIN users u ON u.id = p.author_id AND u.deleted_at IS NULL
		WHERE ` + where + `
		ORDER BY p.created_at DESC
		LIMIT ` + limitParam

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*domain.Post
	var postIDs []uuid.UUID
	for rows.Next() {
		var p domain.Post
		var author domain.PostAuthor
		var authorID uuid.UUID
		err := rows.Scan(
			&p.ID, &authorID, &p.Caption, &p.MediaType, &p.CoverURL,
			&p.LocationName, &p.Visibility,
			&p.Stats.Reactions, &p.Stats.ThumbsUp, &p.Stats.OK,
			&p.Stats.Handshake, &p.Stats.ThumbsDown,
			&p.Stats.Comments, &p.Stats.Saves, &p.Stats.Shares,
			&p.CreatedAt,
			&author.Username, &author.FullName, &author.AvatarURL,
			&author.IsVerified, &author.IsBusiness,
		)
		if err != nil {
			return nil, err
		}
		author.ID = authorID
		p.Author = &author
		p.Hashtags = []string{}
		p.Media = []domain.PostMedia{}
		posts = append(posts, &p)
		postIDs = append(postIDs, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return posts, nil
	}

	if err := r.loadMediaForPosts(ctx, posts, postIDs); err != nil {
		return nil, err
	}
	if err := r.loadHashtagsForPosts(ctx, posts, postIDs); err != nil {
		return nil, err
	}
	if viewerID != nil {
		if err := r.loadViewerForPosts(ctx, posts, postIDs, *viewerID); err != nil {
			return nil, err
		}
	}
	return posts, nil
}

// ListSaved — joriy foydalanuvchi saqlagan postlari.
func (r *PostRepository) ListSaved(ctx context.Context, userID uuid.UUID, cursor *time.Time, limit int) ([]*domain.Post, error) {
	args := []any{userID}
	where := "p.deleted_at IS NULL AND p.is_published = TRUE AND EXISTS (SELECT 1 FROM post_saves ps WHERE ps.user_id = $1 AND ps.post_id = p.id)"
	if cursor != nil {
		args = append(args, *cursor)
		where += " AND p.created_at < $" + itoa(len(args))
	}
	args = append(args, limit+1)
	limitParam := "$" + itoa(len(args))

	query := `
		SELECT p.id, p.author_id, p.caption, p.media_type, p.cover_url,
		       p.location_name, p.visibility,
		       p.thumbs_up_count + p.ok_count + p.handshake_count + p.thumbs_down_count AS reactions_total,
		       p.thumbs_up_count, p.ok_count, p.handshake_count, p.thumbs_down_count,
		       p.comments_count, p.saves_count, p.shares_count,
		       p.created_at,
		       u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business
		FROM posts p
		JOIN users u ON u.id = p.author_id AND u.deleted_at IS NULL
		WHERE ` + where + `
		ORDER BY p.created_at DESC
		LIMIT ` + limitParam

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*domain.Post
	var postIDs []uuid.UUID
	for rows.Next() {
		var p domain.Post
		var author domain.PostAuthor
		var authorID uuid.UUID
		err := rows.Scan(
			&p.ID, &authorID, &p.Caption, &p.MediaType, &p.CoverURL,
			&p.LocationName, &p.Visibility,
			&p.Stats.Reactions, &p.Stats.ThumbsUp, &p.Stats.OK,
			&p.Stats.Handshake, &p.Stats.ThumbsDown,
			&p.Stats.Comments, &p.Stats.Saves, &p.Stats.Shares,
			&p.CreatedAt,
			&author.Username, &author.FullName, &author.AvatarURL,
			&author.IsVerified, &author.IsBusiness,
		)
		if err != nil {
			return nil, err
		}
		author.ID = authorID
		p.Author = &author
		p.Hashtags = []string{}
		p.Media = []domain.PostMedia{}
		posts = append(posts, &p)
		postIDs = append(postIDs, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return posts, nil
	}

	if err := r.loadMediaForPosts(ctx, posts, postIDs); err != nil {
		return nil, err
	}
	if err := r.loadHashtagsForPosts(ctx, posts, postIDs); err != nil {
		return nil, err
	}
	if err := r.loadViewerForPosts(ctx, posts, postIDs, userID); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) loadMediaForPosts(ctx context.Context, posts []*domain.Post, ids []uuid.UUID) error {
	const query = `
		SELECT post_id, url, thumbnail_url, type, duration_seconds, width, height, order_index
		FROM post_media WHERE post_id = ANY($1) ORDER BY post_id, order_index
	`
	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	byPost := make(map[uuid.UUID][]domain.PostMedia)
	for rows.Next() {
		var postID uuid.UUID
		var m domain.PostMedia
		err := rows.Scan(&postID, &m.URL, &m.ThumbnailURL, &m.Type,
			&m.DurationSeconds, &m.Width, &m.Height, &m.OrderIndex)
		if err != nil {
			return err
		}
		byPost[postID] = append(byPost[postID], m)
	}

	for _, p := range posts {
		if media, ok := byPost[p.ID]; ok {
			p.Media = media
		}
	}
	return nil
}

func (r *PostRepository) loadHashtagsForPosts(ctx context.Context, posts []*domain.Post, ids []uuid.UUID) error {
	const query = `
		SELECT ph.post_id, h.name
		FROM post_hashtags ph
		JOIN hashtags h ON h.id = ph.hashtag_id
		WHERE ph.post_id = ANY($1)
	`
	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	byPost := make(map[uuid.UUID][]string)
	for rows.Next() {
		var postID uuid.UUID
		var name string
		if err := rows.Scan(&postID, &name); err != nil {
			return err
		}
		byPost[postID] = append(byPost[postID], name)
	}

	for _, p := range posts {
		if tags, ok := byPost[p.ID]; ok {
			p.Hashtags = tags
		}
	}
	return nil
}

func (r *PostRepository) loadViewerForPosts(ctx context.Context, posts []*domain.Post, ids []uuid.UUID, viewerID uuid.UUID) error {
	const query = `
		SELECT p.id,
		       (SELECT reaction FROM post_reactions WHERE user_id = $1 AND post_id = p.id) AS reaction,
		       EXISTS(SELECT 1 FROM post_saves WHERE user_id = $1 AND post_id = p.id) AS saved,
		       EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = p.author_id) AS following
		FROM posts p
		WHERE p.id = ANY($2)
	`
	rows, err := r.pool.Query(ctx, query, viewerID, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	byID := make(map[uuid.UUID]*domain.PostViewer)
	for rows.Next() {
		var id uuid.UUID
		v := &domain.PostViewer{}
		var reactionStr *string
		if err := rows.Scan(&id, &reactionStr, &v.Saved, &v.FollowingAuthor); err != nil {
			return err
		}
		if reactionStr != nil {
			rt := domain.ReactionType(*reactionStr)
			v.Reaction = &rt
		}
		byID[id] = v
	}

	for _, p := range posts {
		if v, ok := byID[p.ID]; ok {
			p.Viewer = v
		}
	}
	return nil
}

func (r *PostRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	posts, err := r.ListFeed(ctx, domain.FeedForYou, nil, nil, 1)
	if err != nil {
		return nil, err
	}
	for _, p := range posts {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrPostNotFound
}

// PostAuthor — post muallifining ID'sini qaytaradi (bildirishnoma uchun).
func (r *PostRepository) PostAuthor(ctx context.Context, postID uuid.UUID) (uuid.UUID, error) {
	var author uuid.UUID
	err := r.pool.QueryRow(ctx,
		`SELECT author_id FROM posts WHERE id = $1 AND deleted_at IS NULL`,
		postID).Scan(&author)
	return author, err
}

func (r *PostRepository) Delete(ctx context.Context, postID, authorID uuid.UUID) error {
	const query = `
		UPDATE posts SET deleted_at = NOW()
		WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL
		RETURNING id
	`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, postID, authorID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrPostNotFound
	}
	return err
}

// SetReaction — foydalanuvchining reaksiyasini o'rnatadi (yoki o'zgartiradi).
// Bir foydalanuvchi bir postga faqat bitta reaksiya beradi.
func (r *PostRepository) SetReaction(ctx context.Context, postID, userID uuid.UUID, reaction domain.ReactionType) (*domain.ReactionResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Eski reaksiyani topish (counter'ni to'g'rilash uchun)
	var oldReaction *string
	err = tx.QueryRow(ctx,
		`SELECT reaction::text FROM post_reactions WHERE user_id = $1 AND post_id = $2`,
		userID, postID).Scan(&oldReaction)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	// Yangi reaksiyani yozish (upsert)
	_, err = tx.Exec(ctx, `
		INSERT INTO post_reactions (user_id, post_id, reaction) VALUES ($1, $2, $3)
		ON CONFLICT (user_id, post_id) DO UPDATE SET reaction = EXCLUDED.reaction
	`, userID, postID, reaction)
	if err != nil {
		return nil, err
	}

	// Counter'larni yangilash
	if oldReaction == nil {
		// Yangi reaksiya — +1
		_, err = tx.Exec(ctx, reactionIncSQL(string(reaction)), postID)
	} else if *oldReaction != string(reaction) {
		// O'zgartirish — eski -1, yangi +1
		_, err = tx.Exec(ctx, reactionDecSQL(*oldReaction), postID)
		if err == nil {
			_, err = tx.Exec(ctx, reactionIncSQL(string(reaction)), postID)
		}
	}
	if err != nil {
		return nil, err
	}

	result, err := r.fetchReactionCounts(ctx, tx, postID)
	if err != nil {
		return nil, err
	}
	rt := reaction
	result.Reaction = &rt

	return result, tx.Commit(ctx)
}

// RemoveReaction — foydalanuvchining reaksiyasini olib tashlaydi.
func (r *PostRepository) RemoveReaction(ctx context.Context, postID, userID uuid.UUID) (*domain.ReactionResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var oldReaction string
	err = tx.QueryRow(ctx, `
		DELETE FROM post_reactions WHERE user_id = $1 AND post_id = $2
		RETURNING reaction::text
	`, userID, postID).Scan(&oldReaction)
	if err == nil {
		_, err = tx.Exec(ctx, reactionDecSQL(oldReaction), postID)
		if err != nil {
			return nil, err
		}
	} else if err != pgx.ErrNoRows {
		return nil, err
	}

	result, err := r.fetchReactionCounts(ctx, tx, postID)
	if err != nil {
		return nil, err
	}
	return result, tx.Commit(ctx)
}

func (r *PostRepository) fetchReactionCounts(ctx context.Context, tx pgx.Tx, postID uuid.UUID) (*domain.ReactionResult, error) {
	res := &domain.ReactionResult{}
	err := tx.QueryRow(ctx, `
		SELECT thumbs_up_count, ok_count, handshake_count, thumbs_down_count
		FROM posts WHERE id = $1
	`, postID).Scan(&res.ThumbsUp, &res.OK, &res.Handshake, &res.ThumbsDown)
	if err != nil {
		return nil, err
	}
	res.ReactionsCount = res.ThumbsUp + res.OK + res.Handshake + res.ThumbsDown
	return res, nil
}

func reactionIncSQL(r string) string {
	col := reactionColumn(r)
	return `UPDATE posts SET ` + col + ` = ` + col + ` + 1 WHERE id = $1`
}

func reactionDecSQL(r string) string {
	col := reactionColumn(r)
	return `UPDATE posts SET ` + col + ` = GREATEST(` + col + ` - 1, 0) WHERE id = $1`
}

func reactionColumn(r string) string {
	switch r {
	case "ok":
		return "ok_count"
	case "handshake":
		return "handshake_count"
	case "thumbs_down":
		return "thumbs_down_count"
	default:
		return "thumbs_up_count"
	}
}

func (r *PostRepository) Save(ctx context.Context, postID, userID uuid.UUID) error {
	const q = `
		INSERT INTO post_saves (user_id, post_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	cmd, err := r.pool.Exec(ctx, q, userID, postID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() > 0 {
		_, err = r.pool.Exec(ctx, `UPDATE posts SET saves_count = saves_count + 1 WHERE id = $1`, postID)
	}
	return err
}

func (r *PostRepository) Unsave(ctx context.Context, postID, userID uuid.UUID) error {
	const q = `DELETE FROM post_saves WHERE user_id = $1 AND post_id = $2`
	cmd, err := r.pool.Exec(ctx, q, userID, postID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() > 0 {
		_, err = r.pool.Exec(ctx, `UPDATE posts SET saves_count = GREATEST(saves_count - 1, 0) WHERE id = $1`, postID)
	}
	return err
}

// itoa — kichik helper (strconv'sis).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
