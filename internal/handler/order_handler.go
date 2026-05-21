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

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) Create(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	var in domain.CreateOrderInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}
	order, err := h.svc.Create(c.Context(), userID, in)
	if err != nil {
		if errors.Is(err, domain.ErrCartEmpty) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": order})
}

func (h *OrderHandler) List(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	status := c.Query("status")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	list, err := h.svc.List(c.Context(), userID, status, limit, offset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *OrderHandler) Cancel(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "order ID noto'g'ri")
	}
	if err := h.svc.Cancel(c.Context(), orderID, userID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
