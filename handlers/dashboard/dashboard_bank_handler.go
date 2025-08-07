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

type DashboardBankHandler struct {
	bankService services.IBankService
}

func NewDashboardBankHandler() *DashboardBankHandler {
	return &DashboardBankHandler{
		bankService: services.NewBankService(),
	}
}

func (h *DashboardBankHandler) ListBanks(c *fiber.Ctx) error {
	var params queryparams.ListParams

	if err := c.QueryParser(&params); err != nil {
		params = queryparams.ListParams{}
	}

	params.ApplyDefaults()
	params.OrderBy = "asc"
	params.SortBy = "name"

	paginatedResult, err := h.bankService.GetAllBanks(params)

	renderData := fiber.Map{
		"Title":  "Bankalar",
		"Result": paginatedResult,
		"Params": params,
	}

	if err != nil {
		renderData[renderer.FlashErrorKeyView] = "Bankalar getirilirken bir hata oluştu."
		renderData["Result"] = &queryparams.PaginatedResult{
			Data: []models.Bank{},
			Meta: queryparams.PaginationMeta{
				CurrentPage: params.Page,
				PerPage:     params.PerPage,
			},
		}
	}

	return renderer.Render(c, "dashboard/banks/list", "layouts/dashboard", renderData, http.StatusOK)
}

func (h *DashboardBankHandler) ShowCreateBank(c *fiber.Ctx) error {
	return renderer.Render(c, "dashboard/banks/create", "layouts/dashboard", fiber.Map{
		"Title": "Yeni Banka Ekle",
	})
}

func (h *DashboardBankHandler) CreateBank(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidateBankRequest(c)

	if err != nil {
		return renderBankFormError(c, "dashboard/banks/create", "Yeni Banka Ekle", req, err.Error())
	}

	bank := &models.Bank{
		Name:     req.Name,
		IsActive: req.IsActive == "true",
	}

	if err := h.bankService.CreateBank(c.UserContext(), bank); err != nil {
		return renderBankFormError(c, "dashboard/banks/create", "Yeni Banka Ekle", req, "Banka oluşturulamadı: "+err.Error())
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Banka başarıyla oluşturuldu.")
	return c.Redirect("/dashboard/banks", fiber.StatusFound)
}

func (h *DashboardBankHandler) ShowUpdateBank(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz ID")
	}

	bank, err := h.bankService.GetBankByID(uint(id))

	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Banka bulunamadı.")
		return c.Redirect("/dashboard/banks", fiber.StatusSeeOther)
	}

	return renderer.Render(c, "dashboard/banks/update", "layouts/dashboard", fiber.Map{
		"Title": "Banka Düzenle",
		"Bank":  bank,
	})
}

func (h *DashboardBankHandler) UpdateBank(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz ID")
	}

	req, err := requests.ParseAndValidateBankRequest(c)

	if err != nil {
		existingBank, dbErr := h.bankService.GetBankByID(uint(id))

		if dbErr != nil {
			_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Güncellenecek banka bulunamadı.")
			return c.Redirect("/dashboard/banks", fiber.StatusSeeOther)
		}

		existingBank.Name = req.Name
		existingBank.IsActive = req.IsActive == "true"

		return renderBankFormError(c, "dashboard/banks/update", "Banka Düzenle", req, err.Error(), existingBank)
	}

	bank := &models.Bank{
		Name:     req.Name,
		IsActive: req.IsActive == "true",
	}

	userID, _ := c.Locals("userID").(uint)

	if err := h.bankService.UpdateBank(c.UserContext(), uint(id), bank, userID); err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Banka güncellenemedi: "+err.Error())

		return c.Redirect("/dashboard/banks/update/"+c.Params("id"), fiber.StatusSeeOther)
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Banka başarıyla güncellendi.")
	return c.Redirect("/dashboard/banks", fiber.StatusFound)
}

func (h *DashboardBankHandler) DeleteBank(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz ID")
	}

	if err := h.bankService.DeleteBank(c.UserContext(), uint(id)); err != nil {
		errMsg := "Banka silinemedi: " + err.Error()

		if strings.Contains(c.Get("Accept"), "application/json") {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errMsg})
		}

		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, errMsg)
		return c.Redirect("/dashboard/banks", fiber.StatusSeeOther)
	}

	if strings.Contains(c.Get("Accept"), "application/json") {
		return c.JSON(fiber.Map{"message": "Banka başarıyla silindi."})
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Banka başarıyla silindi.")
	return c.Redirect("/dashboard/banks", fiber.StatusFound)
}

func renderBankFormError(c *fiber.Ctx, template, title string, req any, message string, fallback ...*models.Bank) error {
	form, ok := req.(requests.BankRequest)

	if !ok {
		return c.Status(http.StatusInternalServerError).SendString("Sunucu Hatası")
	}

	bank := &models.Bank{
		Name:     form.Name,
		IsActive: form.IsActive == "true",
	}

	if len(fallback) > 0 && fallback[0] != nil {
		bank = fallback[0]
	}

	return renderer.Render(c, template, "layouts/dashboard", fiber.Map{
		"Title":                    title,
		renderer.FlashErrorKeyView: message,
		"Bank":                     bank,
	}, http.StatusBadRequest)
}
