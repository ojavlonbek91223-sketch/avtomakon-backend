package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/middleware"
	"github.com/avtomakon/backend/internal/pkg/validator"
	"github.com/avtomakon/backend/internal/service"
)

type CommentHandler struct {
	svc *service.CommentService
}

func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

func (h *CommentHandler) List(c *fiber.Ctx) error {
	postID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "post ID noto'g'ri")
	}

	var viewerID *uuid.UUID
	if v, ok := middleware.GetUserID(c); ok {
		viewerID = &v
	}

	limit := c.QueryInt("limit", 20)
	list, err := h.svc.List(c.Context(), postID, viewerID, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *CommentHandler) Create(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	postID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "post ID noto'g'ri")
	}

	var in domain.CreateCommentInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	comment, err := h.svc.Create(c.Context(), postID, userID, in)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": comment})
}

func (h *CommentHandler) Delete(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	commentID, err := uuid.Parse(c.Params("comment_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "comment ID noto'g'ri")
	}
	if err := h.svc.Delete(c.Context(), commentID, userID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CommentHandler) Like(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	commentID, err := uuid.Parse(c.Params("comment_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "comment ID noto'g'ri")
	}
	liked, count, err := h.svc.ToggleLike(c.Context(), commentID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{"liked": liked, "likes_count": count},
	})
}
