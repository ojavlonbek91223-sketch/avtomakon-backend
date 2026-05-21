package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/middleware"
	"github.com/avtomakon/backend/internal/pkg/validator"
	"github.com/avtomakon/backend/internal/service"
)

type ChatHandler struct {
	svc *service.ChatService
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{svc: svc}
}

func (h *ChatHandler) List(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	filter := c.Query("filter", "all")
	list, err := h.svc.ListConversations(c.Context(), userID, filter)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *ChatHandler) Start(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}

	var in domain.StartConversationInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	id, err := h.svc.StartConversation(c.Context(), userID, in.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{"id": id},
	})
}

func (h *ChatHandler) Messages(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	convID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID noto'g'ri")
	}
	limit := c.QueryInt("limit", 50)
	msgs, err := h.svc.Messages(c.Context(), convID, userID, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	}
	return c.JSON(fiber.Map{"data": msgs})
}

func (h *ChatHandler) Send(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	convID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID noto'g'ri")
	}

	var in domain.SendMessageInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	msg, err := h.svc.SendMessage(c.Context(), convID, userID, in)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": msg})
}

func (h *ChatHandler) MarkRead(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	convID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID noto'g'ri")
	}

	var body struct {
		MessageID string `json:"message_id" validate:"required,uuid"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	msgID, err := uuid.Parse(body.MessageID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "message_id noto'g'ri")
	}

	if err := h.svc.MarkRead(c.Context(), convID, userID, msgID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
