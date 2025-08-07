package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"zatrano/models"
	"zatrano/pkg/filemanager"
	"zatrano/pkg/flashmessages"
	"zatrano/pkg/queryparams"
	"zatrano/pkg/renderer"
	"zatrano/requests"
	"zatrano/services"

	"github.com/gofiber/fiber/v2"
)

type PanelCardHandler struct {
	cardService        services.ICardService
	bankService        services.IBankService
	socialMediaService services.ISocialMediaService
}

func NewPanelCardHandler() *PanelCardHandler {
	return &PanelCardHandler{
		cardService:        services.NewCardService(),
		bankService:        services.NewBankService(),
		socialMediaService: services.NewSocialMediaService(),
	}
}

func (h *PanelCardHandler) ListCards(c *fiber.Ctx) error {
	var params queryparams.ListParams
	if err := c.QueryParser(&params); err != nil {
		params = queryparams.ListParams{}
	}
	params.ApplyDefaults()
	params.OrderBy = "asc"
	params.SortBy = "name"

	paginatedResult, err := h.cardService.GetAllCards(params)

	renderData := fiber.Map{
		"Title":  "Kartvizitler",
		"Result": paginatedResult,
		"Params": params,
	}

	if err != nil {
		renderData[renderer.FlashErrorKeyView] = "Kartvizitler getirilirken bir hata oluştu."
		renderData["Result"] = &queryparams.PaginatedResult{
			Data: []models.Card{},
			Meta: queryparams.PaginationMeta{
				CurrentPage: params.Page,
				PerPage:     params.PerPage,
			},
		}
	}

	return renderer.Render(c, "panel/cards/list", "layouts/panel", renderData, http.StatusOK)
}

func (h *PanelCardHandler) ShowCreateCard(c *fiber.Ctx) error {
	banksResult, _ := h.bankService.GetAllBanks(queryparams.ListParams{PerPage: 1000})
	socialMediasResult, _ := h.socialMediaService.GetAllSocialMedias(queryparams.ListParams{PerPage: 1000})

	return renderer.Render(c, "panel/cards/create", "layouts/panel", fiber.Map{
		"Title":        "Yeni Kartvizit Ekle",
		"Banks":        banksResult.Data,
		"SocialMedias": socialMediasResult.Data,
	})
}

func (h *PanelCardHandler) CreateCard(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidateCardRequest(c)
	if err != nil {
		return h.renderCardFormError(c, "panel/cards/create", "Yeni Kartvizit Ekle", req, err.Error())
	}

	card := &models.Card{
		Name:       req.Name,
		Slug:       req.Slug,
		Title:      req.Title,
		Telephone:  req.Telephone,
		Email:      req.Email,
		Location:   req.Location,
		WebsiteUrl: req.WebsiteUrl,
		StoreUrl:   req.StoreUrl,
		IsActive:   req.IsActive == "true",
		IsFree:     req.IsFree == "true",
	}

	newFileName, err := filemanager.UploadFile(c, "photo", "cards")
	if err != nil && err != filemanager.ErrFileNotProvided {
		return h.renderCardFormError(c, "panel/cards/create", "Yeni Kartvizit Ekle", req, "Fotoğraf yüklenemedi: "+err.Error())
	}
	card.Photo = newFileName

	for _, cb := range req.CardBanks {
		card.CardBanks = append(card.CardBanks, models.CardBank{BankID: cb.BankID, IBAN: cb.IBAN})
	}
	for _, cs := range req.CardSocialMedia {
		card.CardSocialMedia = append(card.CardSocialMedia, models.CardSocialMedia{SocialMediaID: cs.SocialMediaID, URL: cs.URL})
	}

	if err := h.cardService.CreateCardWithRelations(c.UserContext(), card); err != nil {
		filemanager.DeleteFile("cards", newFileName)
		return h.renderCardFormError(c, "panel/cards/create", "Yeni Kartvizit Ekle", req, "Kartvizit oluşturulamadı: "+err.Error())
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Kartvizit başarıyla oluşturuldu.")
	return c.Redirect("/panel/cards", http.StatusFound)
}

func (h *PanelCardHandler) ShowUpdateCard(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Geçersiz Kartvizit ID'si.")
		return c.Redirect("/panel/cards", http.StatusSeeOther)
	}

	card, err := h.cardService.GetCardByID(uint(id))
	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Kartvizit bulunamadı.")
		return c.Redirect("/panel/cards", http.StatusSeeOther)
	}

	banksResult, _ := h.bankService.GetAllBanks(queryparams.ListParams{PerPage: 1000})
	socialMediasResult, _ := h.socialMediaService.GetAllSocialMedias(queryparams.ListParams{PerPage: 1000})

	return renderer.Render(c, "panel/cards/update", "layouts/panel", fiber.Map{
		"Title":        "Kartvizit Düzenle",
		"Card":         card,
		"Banks":        banksResult.Data,
		"SocialMedias": socialMediasResult.Data,
	})
}

func (h *PanelCardHandler) UpdateCard(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Geçersiz Kartvizit ID'si.")
		return c.Redirect("/panel/cards", http.StatusSeeOther)
	}

	req, err := requests.ParseAndValidateCardRequest(c)
	if err != nil {
		existingCard, dbErr := h.cardService.GetCardByID(uint(id))
		if dbErr != nil {
			_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Güncellenecek Kartvizit bulunamadı.")
			return c.Redirect("/panel/cards", http.StatusSeeOther)
		}
		return h.renderCardFormError(c, "panel/cards/update", "Kartvizit Düzenle", req, err.Error(), existingCard)
	}

	existingCard, err := h.cardService.GetCardByID(uint(id))
	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Güncellenecek Kartvizit bulunamadı.")
		return c.Redirect("/panel/cards", http.StatusSeeOther)
	}

	if req.Slug != existingCard.Slug {
		isAvailable, err := h.cardService.IsSlugAvailable(req.Slug, uint(id))
		if err != nil || !isAvailable {
			_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Bu kullanıcı adı zaten alınmış.")
			return c.Redirect("/panel/cards/update/"+strconv.Itoa(id), http.StatusSeeOther)
		}
	}

	newFileName, err := filemanager.UploadFile(c, "photo", "cards")
	var oldPhotoToDelete string
	if newFileName != "" {
		oldPhotoToDelete = existingCard.Photo
		existingCard.Photo = newFileName
	}

	existingCard.Name = req.Name
	existingCard.Slug = req.Slug
	existingCard.Title = req.Title
	existingCard.Telephone = req.Telephone
	existingCard.Email = req.Email
	existingCard.Location = req.Location
	existingCard.WebsiteUrl = req.WebsiteUrl
	existingCard.StoreUrl = req.StoreUrl
	existingCard.IsActive = req.IsActive == "true"
	existingCard.IsFree = req.IsFree == "true"

	existingCard.CardBanks = []models.CardBank{}
	for _, cb := range req.CardBanks {
		existingCard.CardBanks = append(existingCard.CardBanks, models.CardBank{
			BankID: cb.BankID,
			IBAN:   cb.IBAN,
		})
	}

	existingCard.CardSocialMedia = []models.CardSocialMedia{}
	for _, cs := range req.CardSocialMedia {
		existingCard.CardSocialMedia = append(existingCard.CardSocialMedia, models.CardSocialMedia{
			SocialMediaID: cs.SocialMediaID,
			URL:           cs.URL,
		})
	}

	if err := h.cardService.UpdateCardWithRelations(c.UserContext(), existingCard); err != nil {
		if newFileName != "" {
			filemanager.DeleteFile("cards", newFileName)
		}
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Kartvizit güncellenemedi: "+err.Error())
		return c.Redirect("/panel/cards/update/"+strconv.Itoa(id), http.StatusSeeOther)
	}

	if oldPhotoToDelete != "" {
		filemanager.DeleteFile("cards", oldPhotoToDelete)
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Kartvizit başarıyla güncellendi.")
	return c.Redirect("/panel/cards", http.StatusFound)
}

func (h *PanelCardHandler) DeleteCard(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz ID")
	}

	card, err := h.cardService.GetCardByID(uint(id))
	if err == nil && card.Photo != "" {
		filemanager.DeleteFile("cards", card.Photo)
	}

	if err := h.cardService.DeleteCardWithRelations(c.UserContext(), uint(id)); err != nil {
		errMsg := "Kartvizit silinemedi: " + err.Error()
		if strings.Contains(c.Get("Accept"), "application/json") {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errMsg})
		}
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, errMsg)
		return c.Redirect("/panel/cards", fiber.StatusSeeOther)
	}

	if strings.Contains(c.Get("Accept"), "application/json") {
		return c.JSON(fiber.Map{"message": "Kartvizit başarıyla silindi."})
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Kartvizit başarıyla silindi.")
	return c.Redirect("/panel/cards", http.StatusFound)
}

func (h *PanelCardHandler) SlugCheck(c *fiber.Ctx) error {
	slug := c.Query("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"is_available": false,
			"message":      "Slug parametresi eksik.",
		})
	}

	excludeIDStr := c.Query("exclude_id")
	var excludeID uint = 0
	if excludeIDStr != "" {
		id, err := strconv.ParseUint(excludeIDStr, 10, 32)
		if err == nil {
			excludeID = uint(id)
		}
	}

	isAvailable, err := h.cardService.IsSlugAvailable(slug, excludeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"is_available": false,
			"message":      "Sunucu hatası.",
		})
	}

	return c.JSON(fiber.Map{
		"is_available": isAvailable,
	})
}

func (h *PanelCardHandler) renderCardFormError(c *fiber.Ctx, template, title string, req any, message string, fallback ...*models.Card) error {
	form, ok := req.(requests.CardRequest)
	if !ok {
		return c.Status(http.StatusInternalServerError).SendString("Sunucu Hatası")
	}

	card := &models.Card{
		Name:       form.Name,
		Slug:       form.Slug,
		Title:      form.Title,
		Telephone:  form.Telephone,
		Email:      form.Email,
		Location:   form.Location,
		WebsiteUrl: form.WebsiteUrl,
		StoreUrl:   form.StoreUrl,
		IsActive:   form.IsActive == "true",
		IsFree:     form.IsFree == "true",
	}

	if len(fallback) > 0 && fallback[0] != nil {
		card = fallback[0]
	}

	banksResult, _ := h.bankService.GetAllBanks(queryparams.ListParams{PerPage: 1000})
	socialMediasResult, _ := h.socialMediaService.GetAllSocialMedias(queryparams.ListParams{PerPage: 1000})

	return renderer.Render(c, template, "layouts/panel", fiber.Map{
		"Title":                    title,
		renderer.FlashErrorKeyView: message,
		"Card":                     card,
		"Banks":                    banksResult.Data,
		"SocialMedias":             socialMediasResult.Data,
	}, http.StatusBadRequest)
}
