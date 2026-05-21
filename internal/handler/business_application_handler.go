package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/middleware"
	"github.com/avtomakon/backend/internal/pkg/validator"
	"github.com/avtomakon/backend/internal/service"
)

type BusinessApplicationHandler struct {
	svc *service.BusinessApplicationService
}

func NewBusinessApplicationHandler(svc *service.BusinessApplicationService) *BusinessApplicationHandler {
	return &BusinessApplicationHandler{svc: svc}
}

// Apply — POST /business-applications
func (h *BusinessApplicationHandler) Apply(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}

	var in domain.ApplyBusinessInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	app, err := h.svc.Apply(c.Context(), userID, in)
	if err != nil {
		if errors.Is(err, domain.ErrApplicationAlreadyPending) {
			return fiber.NewError(fiber.StatusConflict, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"application": app,
			"message":     "Arizangiz qabul qilindi. 1-3 ish kuni ichida ko'rib chiqamiz.",
		},
	})
}

// Mine — GET /business-applications/mine
func (h *BusinessApplicationHandler) Mine(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}

	apps, err := h.svc.Mine(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"data": apps})
}
