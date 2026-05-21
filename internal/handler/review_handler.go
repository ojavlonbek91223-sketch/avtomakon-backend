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

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

func (h *ReviewHandler) ListBusinessReviews(c *fiber.Ctx) error {
	return h.listByTarget(c, domain.ReviewTargetBusiness)
}

func (h *ReviewHandler) ListProductReviews(c *fiber.Ctx) error {
	return h.listByTarget(c, domain.ReviewTargetProduct)
}

func (h *ReviewHandler) listByTarget(c *fiber.Ctx, targetType domain.ReviewTargetType) error {
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID noto'g'ri")
	}
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	list, err := h.svc.ListByTarget(c.Context(), targetType, targetID, limit, offset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *ReviewHandler) CreateBusinessReview(c *fiber.Ctx) error {
	return h.create(c, domain.ReviewTargetBusiness)
}

func (h *ReviewHandler) CreateProductReview(c *fiber.Ctx) error {
	return h.create(c, domain.ReviewTargetProduct)
}

func (h *ReviewHandler) create(c *fiber.Ctx, targetType domain.ReviewTargetType) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID noto'g'ri")
	}

	var in domain.CreateReviewInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	rev, err := h.svc.Create(c.Context(), userID, targetID, targetType, in)
	if err != nil {
		if errors.Is(err, domain.ErrReviewExists) {
			return fiber.NewError(fiber.StatusConflict, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": rev})
}

func (h *ReviewHandler) MyReviews(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	limit := c.QueryInt("limit", 30)
	offset := c.QueryInt("offset", 0)
	list, err := h.svc.MyReviews(c.Context(), userID, limit, offset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *ReviewHandler) Reply(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	reviewID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID noto'g'ri")
	}
	var in domain.ReplyToReviewInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}
	if err := h.svc.Reply(c.Context(), reviewID, userID, in.Text); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ReviewHandler) Delete(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	reviewID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID noto'g'ri")
	}
	if err := h.svc.Delete(c.Context(), reviewID, userID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
