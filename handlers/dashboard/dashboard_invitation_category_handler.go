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

type DashboardInvitationCategoryHandler struct {
	categoryService services.IInvitationCategoryService
}

func NewDashboardInvitationCategoryHandler() *DashboardInvitationCategoryHandler {
	return &DashboardInvitationCategoryHandler{
		categoryService: services.NewInvitationCategoryService(),
	}
}

func (h *DashboardInvitationCategoryHandler) ListCategories(c *fiber.Ctx) error {
	var params queryparams.ListParams
	if err := c.QueryParser(&params); err != nil {
		params = queryparams.ListParams{}
	}
	params.ApplyDefaults()
	params.OrderBy = "asc"
	params.SortBy = "name"

	paginatedResult, err := h.categoryService.GetAllCategories(params)

	renderData := fiber.Map{
		"Title":  "Davet Kategorileri",
		"Result": paginatedResult,
		"Params": params,
	}

	if err != nil {
		renderData[renderer.FlashErrorKeyView] = "Kategoriler getirilirken bir hata oluştu."
		renderData["Result"] = &queryparams.PaginatedResult{
			Data: []models.InvitationCategory{},
			Meta: queryparams.PaginationMeta{
				CurrentPage: params.Page,
				PerPage:     params.PerPage,
			},
		}
	}

	return renderer.Render(c, "dashboard/invitation-categories/list", "layouts/dashboard", renderData, http.StatusOK)
}

func (h *DashboardInvitationCategoryHandler) ShowCreateCategory(c *fiber.Ctx) error {
	return renderer.Render(c, "dashboard/invitation-categories/create", "layouts/dashboard", fiber.Map{
		"Title": "Yeni Kategori Ekle",
	})
}

func (h *DashboardInvitationCategoryHandler) CreateCategory(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidateInvitationCategoryRequest(c)
	if err != nil {
		return renderCategoryFormError(c, "dashboard/invitation-categories/create", "Yeni Kategori Ekle", req, err.Error())
	}

	category := &models.InvitationCategory{
		Name:     req.Name,
		Icon:     req.Icon,
		Template: req.Template,
		IsActive: req.IsActive == "true",
	}

	if err := h.categoryService.CreateCategory(c.UserContext(), category); err != nil {
		return renderCategoryFormError(c, "dashboard/invitation-categories/create", "Yeni Kategori Ekle", req, "Kategori oluşturulamadı: "+err.Error())
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Kategori başarıyla oluşturuldu.")
	return c.Redirect("/dashboard/invitation-categories", fiber.StatusFound)
}

func (h *DashboardInvitationCategoryHandler) ShowUpdateCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz ID")
	}

	category, err := h.categoryService.GetCategoryByID(uint(id))
	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Kategori bulunamadı.")
		return c.Redirect("/dashboard/invitation-categories", fiber.StatusSeeOther)
	}

	return renderer.Render(c, "dashboard/invitation-categories/update", "layouts/dashboard", fiber.Map{
		"Title":              "Kategori Düzenle",
		"InvitationCategory": category,
	})
}

func (h *DashboardInvitationCategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz ID")
	}

	req, err := requests.ParseAndValidateInvitationCategoryRequest(c)
	if err != nil {
		existingCategory, dbErr := h.categoryService.GetCategoryByID(uint(id))
		if dbErr != nil {
			_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Güncellenecek kategori bulunamadı.")
			return c.Redirect("/dashboard/invitation-categories", fiber.StatusSeeOther)
		}

		existingCategory.Name = req.Name
		existingCategory.Icon = req.Icon
		existingCategory.Template = req.Template
		existingCategory.IsActive = req.IsActive == "true"

		return renderCategoryFormError(c, "dashboard/invitation-categories/update", "Kategori Düzenle", req, err.Error(), existingCategory)
	}

	category := &models.InvitationCategory{
		Name:     req.Name,
		Icon:     req.Icon,
		Template: req.Template,
		IsActive: req.IsActive == "true",
	}

	userID, _ := c.Locals("userID").(uint)
	if err := h.categoryService.UpdateCategory(c.UserContext(), uint(id), category, userID); err != nil {
		return renderCategoryFormError(c, "dashboard/invitation-categories/update", "Kategori Düzenle", req, "Kategori güncellenemedi: "+err.Error())
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Kategori başarıyla güncellendi.")
	return c.Redirect("/dashboard/invitation-categories", fiber.StatusFound)
}

func (h *DashboardInvitationCategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz ID")
	}

	if err := h.categoryService.DeleteCategory(c.UserContext(), uint(id)); err != nil {
		errMsg := "Kategori silinemedi: " + err.Error()

		if strings.Contains(c.Get("Accept"), "application/json") {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errMsg})
		}

		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, errMsg)
		return c.Redirect("/dashboard/invitation-categories", fiber.StatusSeeOther)
	}

	if strings.Contains(c.Get("Accept"), "application/json") {
		return c.JSON(fiber.Map{"message": "Kategori başarıyla silindi."})
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Kategori başarıyla silindi.")
	return c.Redirect("/dashboard/invitation-categories", fiber.StatusFound)
}

func renderCategoryFormError(c *fiber.Ctx, template, title string, req any, message string, fallback ...*models.InvitationCategory) error {
	form, ok := req.(requests.InvitationCategoryRequest)
	if !ok {
		return c.Status(http.StatusInternalServerError).SendString("Sunucu Hatası")
	}

	category := &models.InvitationCategory{
		Name:     form.Name,
		Icon:     form.Icon,
		Template: form.Template,
		IsActive: form.IsActive == "true",
	}

	if len(fallback) > 0 && fallback[0] != nil {
		category = fallback[0]
	}

	return renderer.Render(c, template, "layouts/dashboard", fiber.Map{
		"Title":                    title,
		renderer.FlashErrorKeyView: message,
		"InvitationCategory":       category,
	}, http.StatusBadRequest)
}
