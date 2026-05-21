package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/middleware"
	"github.com/avtomakon/backend/internal/repository/postgres"
	"github.com/avtomakon/backend/internal/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Me — GET /users/me
func (h *UserHandler) Me(c *fiber.Ctx) error {
	id, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	user, err := h.svc.Me(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": user})
}

type updateProfileRequest struct {
	FullName  *string `json:"full_name" validate:"omitempty,min=2,max=100"`
	Username  *string `json:"username" validate:"omitempty,min=3,max=50,alphanum"`
	Bio       *string `json:"bio" validate:"omitempty,max=200"`
	AvatarURL *string `json:"avatar_url" validate:"omitempty,url"`
	Language  *string `json:"language" validate:"omitempty,oneof=uz ru en"`
}

// UpdateMe — PATCH /users/me
func (h *UserHandler) UpdateMe(c *fiber.Ctx) error {
	id, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	var req updateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}

	user, err := h.svc.UpdateProfile(c.Context(), id, postgres.UpdateProfileInput{
		FullName:  req.FullName,
		Username:  req.Username,
		Bio:       req.Bio,
		AvatarURL: req.AvatarURL,
		Language:  req.Language,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"data": user})
}

// Get — GET /users/:id (public + optional viewer)
func (h *UserHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri ID")
	}

	var viewerID *uuid.UUID
	if v, ok := middleware.GetUserID(c); ok {
		viewerID = &v
	}

	user, following, err := h.svc.GetPublic(c.Context(), id, viewerID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"user":      user,
			"following": following,
		},
	})
}

// Follow — POST /users/:id/follow
func (h *UserHandler) Follow(c *fiber.Ctx) error {
	viewerID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri ID")
	}
	count, err := h.svc.Follow(c.Context(), viewerID, targetID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{"following": true, "followers_count": count},
	})
}

// Followers — GET /users/:id/followers
func (h *UserHandler) Followers(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri ID")
	}
	var viewerID *uuid.UUID
	if v, ok := middleware.GetUserID(c); ok {
		viewerID = &v
	}
	list, err := h.svc.Followers(c.Context(), id, viewerID, c.QueryInt("limit", 30), c.QueryInt("offset", 0))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": list})
}

// Following — GET /users/:id/following
func (h *UserHandler) Following(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri ID")
	}
	var viewerID *uuid.UUID
	if v, ok := middleware.GetUserID(c); ok {
		viewerID = &v
	}
	list, err := h.svc.Following(c.Context(), id, viewerID, c.QueryInt("limit", 30), c.QueryInt("offset", 0))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": list})
}

// Unfollow — DELETE /users/:id/follow
func (h *UserHandler) Unfollow(c *fiber.Ctx) error {
	viewerID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri ID")
	}
	count, err := h.svc.Unfollow(c.Context(), viewerID, targetID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{"following": false, "followers_count": count},
	})
}
