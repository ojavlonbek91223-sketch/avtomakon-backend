package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/service"
)

type BusinessHandler struct {
	svc *service.BusinessService
}

func NewBusinessHandler(svc *service.BusinessService) *BusinessHandler {
	return &BusinessHandler{svc: svc}
}

// Nearby — GET /businesses?lat=41.3&lng=69.27&radius=5000&type=master
func (h *BusinessHandler) Nearby(c *fiber.Ctx) error {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "lat majburiy")
	}
	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "lng majburiy")
	}

	radius := c.QueryInt("radius", 10000)
	bizType := c.Query("type", "")
	limit := c.QueryInt("limit", 50)

	if bizType != "" && bizType != "master" && bizType != "seller" {
		return fiber.NewError(fiber.StatusBadRequest, "type: master yoki seller bo'lishi kerak")
	}

	businesses, err := h.svc.Nearby(c.Context(), domain.NearbyParams{
		Lat: lat, Lng: lng, RadiusM: radius, Type: bizType, Limit: limit,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if businesses == nil {
		businesses = []*domain.Business{}
	}

	return c.JSON(fiber.Map{
		"data": businesses,
		"meta": fiber.Map{"total": len(businesses)},
	})
}

func (h *BusinessHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri ID")
	}

	biz, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrBusinessNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"data": biz})
}
