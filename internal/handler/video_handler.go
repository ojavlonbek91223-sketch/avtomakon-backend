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

type VideoHandler struct {
	svc *service.VideoService
}

func NewVideoHandler(svc *service.VideoService) *VideoHandler {
	return &VideoHandler{svc: svc}
}

func (h *VideoHandler) List(c *fiber.Ctx) error {
	search := c.Query("search")
	category := c.Query("category")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	list, err := h.svc.List(c.Context(), search, category, limit, offset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *VideoHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID noto'g'ri")
	}
	v, err := h.svc.Get(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrVideoNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": v})
}

func (h *VideoHandler) Create(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	var in domain.CreateVideoInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}
	id, err := h.svc.Create(c.Context(), userID, in)
	if err != nil {
		if errors.Is(err, domain.ErrOnlyBusinessVideo) {
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{"id": id},
	})
}
