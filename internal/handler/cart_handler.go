package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/middleware"
	"github.com/avtomakon/backend/internal/pkg/validator"
	"github.com/avtomakon/backend/internal/service"
)

type CartHandler struct {
	svc *service.CartService
}

func NewCartHandler(svc *service.CartService) *CartHandler {
	return &CartHandler{svc: svc}
}

func (h *CartHandler) Get(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	cart, err := h.svc.Get(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": cart})
}

func (h *CartHandler) Add(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	var in domain.AddToCartInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}
	if err := h.svc.Add(c.Context(), userID, in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CartHandler) Update(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	itemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "item ID noto'g'ri")
	}
	var in domain.UpdateCartItemInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}
	if err := h.svc.Update(c.Context(), userID, itemID, in.Quantity); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CartHandler) Remove(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	itemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "item ID noto'g'ri")
	}
	if err := h.svc.Remove(c.Context(), userID, itemID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
