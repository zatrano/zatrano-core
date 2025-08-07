package requests

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type CardRequest struct {
	Name            string                   `form:"name" validate:"required"`
	Slug            string                   `form:"slug" validate:"required"`
	Title           string                   `form:"title" validate:"-"`
	Photo           string                   `form:"photo" validate:"-"`
	Telephone       string                   `form:"telephone" validate:"-"`
	Email           string                   `form:"email" validate:"-"`
	Location        string                   `form:"location" validate:"-"`
	WebsiteUrl      string                   `form:"website_url" validate:"-"`
	StoreUrl        string                   `form:"store_url" validate:"-"`
	IsActive        string                   `form:"is_active" validate:"required,oneof=true false"`
	IsFree          string                   `form:"is_free" validate:"required,oneof=true false"`
	CardBanks       []CardBankRequest        `form:"card_banks" validate:"dive"`
	CardSocialMedia []CardSocialMediaRequest `form:"card_social_media" validate:"dive"`
}

type CardBankRequest struct {
	ID     uint   `validate:"-"`
	BankID uint   `form:"bank_id" validate:"-"`
	IBAN   string `form:"iban" validate:"-"`
}

type CardSocialMediaRequest struct {
	ID            uint   `validate:"-"`
	SocialMediaID uint   `form:"social_media_id" validate:"-"`
	URL           string `form:"url" validate:"-"`
}

func ParseAndValidateCardRequest(c *fiber.Ctx) (CardRequest, error) {
	var req CardRequest

	if err := c.BodyParser(&req); err != nil {
		return req, errors.New("geçersiz istek formatı")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		field := validationErrors[0].Field()
		tag := validationErrors[0].Tag()
		errorMessages := map[string]string{
			"Name_required":     "Kart adı zorunludur.",
			"Slug_required":     "Slug zorunludur.",
			"IsActive_required": "Durum (Aktif/Pasif) seçilmelidir.",
			"IsActive_oneof":    "Durum için geçersiz bir değer seçildi.",
			"IsFree_required":   "Ücretsiz/Premium seçilmelidir.",
			"IsFree_oneof":      "Ücretsiz/Premium için geçersiz bir değer seçildi.",
		}
		if msg, ok := errorMessages[field+"_"+tag]; ok {
			return req, errors.New(msg)
		}
		return req, errors.New("lütfen formdaki hataları düzeltin")
	}
	return req, nil
}
