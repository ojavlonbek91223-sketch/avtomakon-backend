package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/middleware"
	"github.com/avtomakon/backend/internal/pkg/validator"
	"github.com/avtomakon/backend/internal/service"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) List(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}

	unreadOnly := c.QueryBool("unread_only", false)
	limit := c.QueryInt("limit", 30)

	list, err := h.svc.List(c.Context(), userID, unreadOnly, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	count, _ := h.svc.UnreadCount(c.Context(), userID)

	return c.JSON(fiber.Map{
		"data": list,
		"meta": fiber.Map{"unread_count": count},
	})
}

func (h *NotificationHandler) UnreadCount(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	count, err := h.svc.UnreadCount(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{"unread_count": count},
	})
}

func (h *NotificationHandler) MarkAllRead(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	if err := h.svc.MarkAllRead(c.Context(), userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *NotificationHandler) RegisterPushToken(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}

	var in domain.PushToken
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	if err := h.svc.SavePushToken(c.Context(), userID, in.Token, in.Platform); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
