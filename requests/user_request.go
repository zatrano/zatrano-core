package requests

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type UserRequest struct {
	Name              string `form:"name" validate:"required,min=2"`
	Email             string `form:"email" validate:"required,email"`
	Password          string `form:"password"`
	Status            string `form:"status" validate:"required,oneof=true false"`
	Type              string `form:"type" validate:"required,oneof=dashboard panel"`
	ResetToken        string `form:"reset_token"`
	EmailVerified     string `form:"email_verified" validate:"required,oneof=true false"`
	VerificationToken string `form:"verification_token"`
	Provider          string `form:"provider"`
	ProviderID        string `form:"provider_id"`
}

func ParseAndValidateUserRequest(c *fiber.Ctx) (UserRequest, error) {
	var req UserRequest

	if err := c.BodyParser(&req); err != nil {
		return req, errors.New("geçersiz istek formatı")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		field := validationErrors[0].Field()
		tag := validationErrors[0].Tag()
		errorMessages := map[string]string{
			"Name_required":          "Kullanıcı adı zorunludur.",
			"Name_min":               "Kullanıcı adı en az 2 karakter olmalıdır.",
			"Email_required":         "E-posta adresi zorunludur.",
			"Email_email":            "Geçerli bir e-posta adresi giriniz.",
			"Status_required":        "Durum seçilmelidir.",
			"Status_oneof":           "Durum için geçersiz bir değer seçildi.",
			"Type_required":          "Kullanıcı tipi seçilmelidir.",
			"Type_oneof":             "Kullanıcı tipi geçersiz.",
			"EmailVerified_required": "E-posta doğrulama durumu seçilmelidir.",
			"EmailVerified_oneof":    "E-posta doğrulama durumu için geçersiz bir değer seçildi.",
		}
		if msg, ok := errorMessages[field+"_"+tag]; ok {
			return req, errors.New(msg)
		}
		return req, errors.New("lütfen formdaki hataları düzeltin")
	}
	return req, nil
}
