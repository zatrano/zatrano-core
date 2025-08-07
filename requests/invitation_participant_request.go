package requests

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type InvitationParticipantRequest struct {
	Title       string `form:"title" validate:"required,min=2"`
	PhoneNumber string `form:"phone_number" validate:"required,min=10"`
	GuestCount  string `form:"guest_count" validate:"required"`
}

func ParseAndValidateInvitationParticipantRequest(c *fiber.Ctx) (InvitationParticipantRequest, error) {
	var req InvitationParticipantRequest

	if err := c.BodyParser(&req); err != nil {
		return req, errors.New("geçersiz istek formatı")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		field := validationErrors[0].Field()
		tag := validationErrors[0].Tag()
		errorMessages := map[string]string{
			"Title_required":       "Ad Soyad zorunludur.",
			"Title_min":            "Ad Soyad en az 2 karakter olmalıdır.",
			"PhoneNumber_required": "Telefon numarası zorunludur.",
			"PhoneNumber_min":      "Telefon numarası en az 10 karakter olmalıdır.",
			"GuestCount_required":  "Kişi sayısı zorunludur.",
		}
		if msg, ok := errorMessages[field+"_"+tag]; ok {
			return req, errors.New(msg)
		}
		return req, errors.New("lütfen formdaki hataları düzeltin")
	}
	return req, nil
}
