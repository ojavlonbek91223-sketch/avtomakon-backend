package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/service"
)

type MarketHandler struct {
	svc *service.MarketService
}

func NewMarketHandler(svc *service.MarketService) *MarketHandler {
	return &MarketHandler{svc: svc}
}

func (h *MarketHandler) Categories(c *fiber.Ctx) error {
	list, err := h.svc.Categories(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *MarketHandler) Products(c *fiber.Ctx) error {
	params := domain.ProductsParams{
		Search: c.Query("search"),
		Sort:   c.Query("sort", "popular"),
		Limit:  c.QueryInt("limit", 20),
		Offset: c.QueryInt("offset", 0),
	}

	if raw := c.Query("category"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "category ID noto'g'ri")
		}
		params.CategoryID = &id
	}

	list, err := h.svc.ListProducts(c.Context(), params)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"data": list,
		"meta": fiber.Map{"count": len(list)},
	})
}

func (h *MarketHandler) GetProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID noto'g'ri")
	}
	p, err := h.svc.GetProduct(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(fiber.Map{"data": p})
}

func (h *MarketHandler) Featured(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	list, err := h.svc.Featured(c.Context(), limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *MarketHandler) Promotions(c *fiber.Ctx) error {
	list, err := h.svc.Promotions(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": list})
}
