package requests

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type SocialMediaRequest struct {
	Name     string `form:"name" validate:"required,min=2"`
	Icon     string `form:"icon" validate:"required"`
	IsActive string `form:"is_active" validate:"required,oneof=true false"`
}

func ParseAndValidateSocialMediaRequest(c *fiber.Ctx) (SocialMediaRequest, error) {
	var req SocialMediaRequest

	if err := c.BodyParser(&req); err != nil {
		return req, errors.New("geçersiz istek formatı")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		field := validationErrors[0].Field()
		tag := validationErrors[0].Tag()
		errorMessages := map[string]string{
			"Name_required":     "Sosyal medya adı zorunludur.",
			"Name_min":          "Sosyal medya adı en az 2 karakter olmalıdır.",
			"Icon_required":     "İkon zorunludur.",
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
