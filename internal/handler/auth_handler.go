package handler

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/middleware"
	"github.com/avtomakon/backend/internal/pkg/validator"
	"github.com/avtomakon/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) SendOTP(c *fiber.Ctx) error {
	var in domain.SendOTPInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	result, err := h.auth.SendOTP(c.Context(), in)
	if err != nil {
		return mapAuthError(err)
	}

	return c.JSON(fiber.Map{"data": result})
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var in domain.RegisterInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	result, err := h.auth.Register(c.Context(), in, c.IP())
	if err != nil {
		return mapAuthError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": result})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var in domain.LoginInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	result, err := h.auth.Login(c.Context(), in, c.IP())
	if err != nil {
		return mapAuthError(err)
	}

	return c.JSON(fiber.Map{"data": result})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var in domain.RefreshInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}
	if err := validator.Struct(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validator.FormatError(err))
	}

	tokens, err := h.auth.Refresh(c.Context(), in.RefreshToken, c.IP())
	if err != nil {
		return mapAuthError(err)
	}

	return c.JSON(fiber.Map{"data": fiber.Map{"tokens": tokens}})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var in domain.RefreshInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "noto'g'ri JSON")
	}

	if err := h.auth.Logout(c.Context(), in.RefreshToken); err != nil {
		return mapAuthError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "auth kerak")
	}
	if err := h.auth.LogoutAll(c.Context(), userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "chiqishda xato")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// mapAuthError — domain xatosini HTTP xatoga aylantiradi.
func mapAuthError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrPhoneAlreadyExists),
		errors.Is(err, domain.ErrUsernameAlreadyExists),
		errors.Is(err, domain.ErrEmailAlreadyExists):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrOTPInvalid),
		errors.Is(err, domain.ErrOTPExpired):
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, domain.ErrOTPTooManyAttempts):
		return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
	case errors.Is(err, domain.ErrRefreshTokenInvalid):
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}
	// Wrapped errorlar uchun (masalan, "ErrOTPTooManyAttempts (kuting...)" )
	if strings.Contains(err.Error(), domain.ErrOTPTooManyAttempts.Error()) {
		return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
	}
	return fiber.NewError(fiber.StatusInternalServerError, "ichki xato")
}
