package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

func Struct(s any) error {
	return v.Struct(s)
}

// FormatError — validator xatosini foydalanuvchi uchun tushunarli matnga aylantiradi.
func FormatError(err error) string {
	var ves validator.ValidationErrors
	if errs, ok := err.(validator.ValidationErrors); ok {
		ves = errs
	} else {
		return err.Error()
	}

	msgs := make([]string, 0, len(ves))
	for _, e := range ves {
		msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field(), messageFor(e)))
	}
	return strings.Join(msgs, "; ")
}

func messageFor(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "majburiy"
	case "email":
		return "noto'g'ri email"
	case "e164":
		return "noto'g'ri telefon (E.164: +998901234567)"
	case "min":
		return fmt.Sprintf("kamida %s belgi", e.Param())
	case "max":
		return fmt.Sprintf("ko'pi bilan %s belgi", e.Param())
	case "len":
		return fmt.Sprintf("aniq %s belgi", e.Param())
	case "oneof":
		return fmt.Sprintf("biri bo'lishi kerak: %s", e.Param())
	default:
		return e.Tag()
	}
}
