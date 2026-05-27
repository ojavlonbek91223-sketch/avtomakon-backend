package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/middleware"
	"github.com/avtomakon/backend/internal/pkg/validator"
	"github.com/avtomakon/backend/internal/service"
)

type PostHandler struct {
	svc *service.PostService
}

func NewPostHandler(svc *service.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

// ListFeed — GET /posts?feed=for_you|following&limit=10&cursor=xxx
func (h *PostHandler) ListFeed(c *fiber.Ctx) error {
	kind := domain.FeedKind(c.Query("feed", string(domain.FeedForYou)))
	limit := c.QueryInt("limit", 10)

	var cursor *time.Time
	if raw := c.Query("cursor"); raw != "" {
		t, err := decodeCursor(raw)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "cursor noto'g'ri")
		}
		cursor = &t
	}

	var viewerID *uuid.UUID
	if id, ok := middleware.GetUserID(c); ok {
		viewerID = &id
	}

	result, err := h.svc.ListFeed(c.Context(), kind, viewerID, cursor, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	meta := fiber.Map{"has_more": result.HasMore}
	if result.NextCursor != nil {
		meta["next_cursor"] = encodeCursor(*result.NextCursor)
	}

	return c.JSON(fiber.Map{
		"data": result.Posts,
		"meta": meta,
	})
}

// Create — POST /posts
func (h *PostHandler) Create(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}

	var in domain.CreatePostInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	id, err := h.svc.Create(c.Context(), userID, in)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{"id": id},
	})
}

// ListVideos — GET /posts/videos — faqat video media_type'li postlar
func (h *PostHandler) ListVideos(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 21)

	var cursor *time.Time
	if raw := c.Query("cursor"); raw != "" {
		t, err := decodeCursor(raw)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "cursor noto'g'ri")
		}
		cursor = &t
	}

	var viewerID *uuid.UUID
	if vid, ok := middleware.GetUserID(c); ok {
		viewerID = &vid
	}

	result, err := h.svc.ListVideos(c.Context(), viewerID, cursor, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	meta := fiber.Map{"has_more": result.HasMore}
	if result.NextCursor != nil {
		meta["next_cursor"] = encodeCursor(*result.NextCursor)
	}

	return c.JSON(fiber.Map{"data": result.Posts, "meta": meta})
}

// SavedPosts — GET /posts/saved — joriy foydalanuvchi saqlagan postlari
func (h *PostHandler) SavedPosts(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}

	limit := c.QueryInt("limit", 21)

	var cursor *time.Time
	if raw := c.Query("cursor"); raw != "" {
		t, err := decodeCursor(raw)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "cursor noto'g'ri")
		}
		cursor = &t
	}

	result, err := h.svc.ListSaved(c.Context(), userID, cursor, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	meta := fiber.Map{"has_more": result.HasMore}
	if result.NextCursor != nil {
		meta["next_cursor"] = encodeCursor(*result.NextCursor)
	}

	return c.JSON(fiber.Map{"data": result.Posts, "meta": meta})
}

// UserPosts — GET /users/:id/posts — foydalanuvchining postlari (profil grid'i uchun)
func (h *PostHandler) UserPosts(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri ID")
	}

	limit := c.QueryInt("limit", 21)

	var cursor *time.Time
	if raw := c.Query("cursor"); raw != "" {
		t, err := decodeCursor(raw)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "cursor noto'g'ri")
		}
		cursor = &t
	}

	var viewerID *uuid.UUID
	if vid, ok := middleware.GetUserID(c); ok {
		viewerID = &vid
	}

	result, err := h.svc.ListByUser(c.Context(), id, viewerID, cursor, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	meta := fiber.Map{"has_more": result.HasMore}
	if result.NextCursor != nil {
		meta["next_cursor"] = encodeCursor(*result.NextCursor)
	}

	return c.JSON(fiber.Map{"data": result.Posts, "meta": meta})
}

// Delete — DELETE /posts/:id
func (h *PostHandler) Delete(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}

	postID, err := parsePostID(c)
	if err != nil {
		return err
	}

	if err := h.svc.Delete(c.Context(), postID, userID); err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// SetReaction — POST /posts/:id/react  body: {"reaction": "thumbs_up"|"ok"|"handshake"|"thumbs_down"}
func (h *PostHandler) SetReaction(c *fiber.Ctx) error {
	userID, postID, err := getUserAndPostID(c)
	if err != nil {
		return err
	}
	var in domain.SetReactionInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	res, err := h.svc.SetReaction(c.Context(), postID, userID, in.Reaction)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": res})
}

// RemoveReaction — DELETE /posts/:id/react
func (h *PostHandler) RemoveReaction(c *fiber.Ctx) error {
	userID, postID, err := getUserAndPostID(c)
	if err != nil {
		return err
	}
	res, err := h.svc.RemoveReaction(c.Context(), postID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": res})
}

func (h *PostHandler) Save(c *fiber.Ctx) error {
	userID, postID, err := getUserAndPostID(c)
	if err != nil {
		return err
	}
	if err := h.svc.Save(c.Context(), postID, userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"saved": true}})
}

func (h *PostHandler) Unsave(c *fiber.Ctx) error {
	userID, postID, err := getUserAndPostID(c)
	if err != nil {
		return err
	}
	if err := h.svc.Unsave(c.Context(), postID, userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"saved": false}})
}

// Yordamchilar

func getUserAndPostID(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	postID, err := parsePostID(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return userID, postID, nil
}

func parsePostID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "post ID noto'g'ri")
	}
	return id, nil
}

// cursor — bazaviy timestamp'ni base64 qiladi (foydalanuvchi ko'rmaydi)
type cursorPayload struct {
	T time.Time `json:"t"`
}

func encodeCursor(t time.Time) string {
	b, _ := json.Marshal(cursorPayload{T: t})
	return base64.URLEncoding.EncodeToString(b)
}

func decodeCursor(s string) (time.Time, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, err
	}
	var p cursorPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return time.Time{}, err
	}
	return p.T, nil
}
