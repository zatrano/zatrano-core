package handlers

import (
	"net/http"
	"strings"

	"zatrano/models"
	"zatrano/pkg/flashmessages"
	"zatrano/pkg/queryparams"
	"zatrano/pkg/renderer"
	"zatrano/requests"
	"zatrano/services"

	"github.com/gofiber/fiber/v2"
)

type DashboardSocialMediaHandler struct {
	socialMediaService services.ISocialMediaService
}

func NewDashboardSocialMediaHandler() *DashboardSocialMediaHandler {
	return &DashboardSocialMediaHandler{
		socialMediaService: services.NewSocialMediaService(),
	}
}

func (h *DashboardSocialMediaHandler) ListSocialMedias(c *fiber.Ctx) error {
	var params queryparams.ListParams

	if err := c.QueryParser(&params); err != nil {
		params = queryparams.ListParams{}
	}

	params.ApplyDefaults()
	params.OrderBy = "asc"
	params.SortBy = "name"

	paginatedResult, err := h.socialMediaService.GetAllSocialMedias(params)

	renderData := fiber.Map{
		"Title":  "Sosyal Medya",
		"Result": paginatedResult,
		"Params": params,
	}

	if err != nil {
		renderData[renderer.FlashErrorKeyView] = "Sosyal medya kayıtları getirilirken bir hata oluştu."
		renderData["Result"] = &queryparams.PaginatedResult{
			Data: []models.SocialMedia{},
			Meta: queryparams.PaginationMeta{CurrentPage: params.Page, PerPage: params.PerPage},
		}
	}

	return renderer.Render(c, "dashboard/social-media/list", "layouts/dashboard", renderData, http.StatusOK)
}

func (h *DashboardSocialMediaHandler) ShowCreateSocialMedia(c *fiber.Ctx) error {
	return renderer.Render(c, "dashboard/social-media/create", "layouts/dashboard", fiber.Map{
		"Title": "Yeni Sosyal Medya Ekle",
	})
}

func (h *DashboardSocialMediaHandler) CreateSocialMedia(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidateSocialMediaRequest(c)

	if err != nil {
		return renderSocialMediaFormError(c, "dashboard/social-media/create", "Yeni Sosyal Medya Ekle", req, err.Error())
	}

	socialMedia := &models.SocialMedia{
		Name:     req.Name,
		Icon:     req.Icon,
		IsActive: req.IsActive == "true",
	}

	if err := h.socialMediaService.CreateSocialMedia(c.UserContext(), socialMedia); err != nil {
		return renderSocialMediaFormError(c, "dashboard/social-media/create", "Yeni Sosyal Medya Ekle", req, "Kayıt oluşturulamadı: "+err.Error())
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Kayıt başarıyla oluşturuldu.")
	return c.Redirect("/dashboard/social-media", fiber.StatusFound)
}

func (h *DashboardSocialMediaHandler) ShowUpdateSocialMedia(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz ID")
	}

	socialMedia, err := h.socialMediaService.GetSocialMediaByID(uint(id))

	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Kayıt bulunamadı.")
		return c.Redirect("/dashboard/social-media", fiber.StatusSeeOther)
	}

	return renderer.Render(c, "dashboard/social-media/update", "layouts/dashboard", fiber.Map{
		"Title":       "Sosyal Medya Düzenle",
		"SocialMedia": socialMedia,
	})
}

func (h *DashboardSocialMediaHandler) UpdateSocialMedia(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz ID")
	}

	req, err := requests.ParseAndValidateSocialMediaRequest(c)

	if err != nil {
		existingSocialMedia, dbErr := h.socialMediaService.GetSocialMediaByID(uint(id))

		if dbErr != nil {
			_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Güncellenecek sosyal medya bulunamadı.")
			return c.Redirect("/dashboard/social-media", fiber.StatusSeeOther)
		}

		existingSocialMedia.Name = req.Name
		existingSocialMedia.Icon = req.Icon
		existingSocialMedia.IsActive = req.IsActive == "true"

		return renderSocialMediaFormError(c, "dashboard/social-media/update", "Sosyal Medya Düzenle", req, err.Error(), existingSocialMedia)
	}

	socialMedia := &models.SocialMedia{
		Name:     req.Name,
		Icon:     req.Icon,
		IsActive: req.IsActive == "true",
	}

	userID, _ := c.Locals("userID").(uint)

	if err := h.socialMediaService.UpdateSocialMedia(c.UserContext(), uint(id), socialMedia, userID); err != nil {
		return renderSocialMediaFormError(c, "dashboard/social-media/update", "Sosyal Medya Düzenle", req, "Sosyal medya güncellenemedi: "+err.Error())
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Sosyal medya başarıyla güncellendi.")
	return c.Redirect("/dashboard/social-media", fiber.StatusFound)
}

func (h *DashboardSocialMediaHandler) DeleteSocialMedia(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz ID")
	}

	if err := h.socialMediaService.DeleteSocialMedia(c.UserContext(), uint(id)); err != nil {
		errMsg := "Kayıt silinemedi: " + err.Error()

		if strings.Contains(c.Get("Accept"), "application/json") {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errMsg})
		}

		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, errMsg)
		return c.Redirect("/dashboard/social-media", fiber.StatusSeeOther)
	}

	if strings.Contains(c.Get("Accept"), "application/json") {
		return c.JSON(fiber.Map{"message": "Kayıt başarıyla silindi."})
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Kayıt başarıyla silindi.")
	return c.Redirect("/dashboard/social-media", fiber.StatusFound)
}

func renderSocialMediaFormError(c *fiber.Ctx, template, title string, req any, message string, fallback ...*models.SocialMedia) error {
	form, ok := req.(requests.SocialMediaRequest)
	if !ok {
		return c.Status(http.StatusInternalServerError).SendString("Sunucu Hatası")
	}

	socialMedia := &models.SocialMedia{
		Name:     form.Name,
		Icon:     form.Icon,
		IsActive: form.IsActive == "true",
	}

	if len(fallback) > 0 && fallback[0] != nil {
		socialMedia = fallback[0]
	}

	return renderer.Render(c, template, "layouts/dashboard", fiber.Map{
		"Title":                    title,
		renderer.FlashErrorKeyView: message,
		"SocialMedia":              socialMedia,
	}, http.StatusBadRequest)
}
