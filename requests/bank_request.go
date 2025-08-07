package requests

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type BankRequest struct {
	Name     string `form:"name" validate:"required,min=2"`
	IsActive string `form:"is_active" validate:"required,oneof=true false"`
}

func ParseAndValidateBankRequest(c *fiber.Ctx) (BankRequest, error) {
	var req BankRequest

	if err := c.BodyParser(&req); err != nil {
		return req, errors.New("geçersiz istek formatı")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		field := validationErrors[0].Field()
		tag := validationErrors[0].Tag()

		errorMessages := map[string]string{
			"Name_required":     "Banka adı zorunludur.",
			"Name_min":          "Banka adı en az 2 karakter olmalıdır.",
			"IsActive_required": "Durum (Aktif/Pasif) seçilmelidir.",
			"IsActive_oneof":    "Durum için geçersiz bir değer seçildi.",
		}

		if msg, ok := errorMessages[field+"_"+tag]; ok {
			return req, errors.New(msg)
		}

		return req, errors.New("lütfen formdaki hataları düzeltin")
	}

	return req, nil
}
