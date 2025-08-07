package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"zatrano/models"
	"zatrano/pkg/filemanager"
	"zatrano/pkg/flashmessages"
	"zatrano/pkg/queryparams"
	"zatrano/pkg/renderer"
	"zatrano/requests"
	"zatrano/services"

	"github.com/gofiber/fiber/v2"
)

type DashboardInvitationHandler struct {
	invitationService services.IInvitationService
	categoryService   services.IInvitationCategoryService
}

func NewDashboardInvitationHandler() *DashboardInvitationHandler {
	return &DashboardInvitationHandler{
		invitationService: services.NewInvitationService(),
		categoryService:   services.NewInvitationCategoryService(),
	}
}

func (h *DashboardInvitationHandler) ListInvitations(c *fiber.Ctx) error {
	var params queryparams.ListParams
	if err := c.QueryParser(&params); err != nil {
		params = queryparams.ListParams{}
	}
	params.ApplyDefaults()
	params.OrderBy = "asc"
	params.SortBy = "name"

	paginatedResult, err := h.invitationService.GetAllInvitations(params)

	renderData := fiber.Map{
		"Title":  "Davetiyeler",
		"Result": paginatedResult,
		"Params": params,
	}

	if err != nil {
		renderData[renderer.FlashErrorKeyView] = "Davetiyeler getirilirken bir hata oluştu."
		renderData["Result"] = &queryparams.PaginatedResult{
			Data: []models.Invitation{},
			Meta: queryparams.PaginationMeta{
				CurrentPage: params.Page,
				PerPage:     params.PerPage,
			},
		}
	}

	return renderer.Render(c, "dashboard/invitations/list", "layouts/dashboard", renderData, http.StatusOK)
}

func (h *DashboardInvitationHandler) ShowCreateInvitation(c *fiber.Ctx) error {
	categories, _ := h.categoryService.GetAllCategories(queryparams.DefaultListParams())

	return renderer.Render(c, "dashboard/invitations/create", "layouts/dashboard", fiber.Map{
		"Title":      "Yeni Davetiye Ekle",
		"Categories": categories,
	})
}

func (h *DashboardInvitationHandler) CreateInvitation(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidateInvitationRequest(c)
	if err != nil {
		return h.renderInvitationFormError(c, "dashboard/invitations/create", "Yeni Davetiye Ekle", req, err.Error())
	}

	newFileName, err := filemanager.UploadFile(c, "image", "invitations")
	if err != nil && err != filemanager.ErrFileNotProvided {
		return h.renderInvitationFormError(c, "dashboard/invitations/create", "Yeni Davetiye Ekle", req, "Resim yüklenemedi: "+err.Error())
	}

	dateValue, _ := time.Parse("2006-01-02", req.Date)
	invitation := &models.Invitation{
		CategoryID:    req.CategoryID,
		Image:         newFileName,
		Venue:         req.Venue,
		Address:       req.Address,
		Location:      req.Location,
		Telephone:     req.Telephone,
		Date:          dateValue,
		Time:          req.Time,
		IsConfirmed:   req.IsConfirmed == "true",
		IsParticipant: req.IsParticipant == "true",
		IsFree:        req.IsFree == "true",
		InvitationDetail: &models.InvitationDetail{
			Title:              req.Detail.Title,
			BrideName:          req.Detail.BrideName,
			BrideSurname:       req.Detail.BrideSurname,
			BrideMotherName:    req.Detail.BrideMotherName,
			BrideMotherSurname: req.Detail.BrideMotherSurname,
			BrideFatherName:    req.Detail.BrideFatherName,
			BrideFatherSurname: req.Detail.BrideFatherSurname,
			GroomName:          req.Detail.GroomName,
			GroomSurname:       req.Detail.GroomSurname,
			GroomMotherName:    req.Detail.GroomMotherName,
			GroomMotherSurname: req.Detail.GroomMotherSurname,
			GroomFatherName:    req.Detail.GroomFatherName,
			GroomFatherSurname: req.Detail.GroomFatherSurname,
			Person:             req.Detail.Person,
			MotherName:         req.Detail.MotherName,
			MotherSurname:      req.Detail.MotherSurname,
			FatherName:         req.Detail.FatherName,
			FatherSurname:      req.Detail.FatherSurname,
			IsMotherLive:       req.Detail.IsMotherLive == "true",
			IsFatherLive:       req.Detail.IsFatherLive == "true",
			IsBrideMotherLive:  req.Detail.IsBrideMotherLive == "true",
			IsBrideFatherLive:  req.Detail.IsBrideFatherLive == "true",
			IsGroomMotherLive:  req.Detail.IsGroomMotherLive == "true",
			IsGroomFatherLive:  req.Detail.IsGroomFatherLive == "true",
		},
	}

	if err := h.invitationService.CreateInvitationWithRelations(c.UserContext(), invitation); err != nil {
		filemanager.DeleteFile("invitations", newFileName)
		return h.renderInvitationFormError(c, "dashboard/invitations/create", "Yeni Davetiye Ekle", req, "Davetiye oluşturulamadı: "+err.Error())
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Davetiye başarıyla oluşturuldu.")
	return c.Redirect("/dashboard/invitations", http.StatusFound)
}

func (h *DashboardInvitationHandler) ShowUpdateInvitation(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Geçersiz davetiye ID'si.")
		return c.Redirect("/dashboard/invitations", http.StatusSeeOther)
	}

	invitation, err := h.invitationService.GetInvitationByID(uint(id))
	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Davetiye bulunamadı.")
		return c.Redirect("/dashboard/invitations", http.StatusSeeOther)
	}

	categories, _ := h.categoryService.GetAllCategories(queryparams.DefaultListParams())

	return renderer.Render(c, "dashboard/invitations/update", "layouts/dashboard", fiber.Map{
		"Title":      "Davetiye Düzenle",
		"Invitation": invitation,
		"Categories": categories,
	})
}

func (h *DashboardInvitationHandler) UpdateInvitation(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Geçersiz davetiye ID'si.")
		return c.Redirect("/dashboard/invitations", http.StatusSeeOther)
	}

	req, err := requests.ParseAndValidateInvitationRequest(c)
	if err != nil {
		existingInvitation, dbErr := h.invitationService.GetInvitationByID(uint(id))
		if dbErr != nil {
			_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Güncellenecek davetiye bulunamadı.")
			return c.Redirect("/dashboard/invitations", http.StatusSeeOther)
		}
		return h.renderInvitationFormError(c, "dashboard/invitations/update", "Davetiye Düzenle", req, err.Error(), existingInvitation)
	}

	existingInvitation, err := h.invitationService.GetInvitationByID(uint(id))
	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Güncellenecek davetiye bulunamadı.")
		return c.Redirect("/dashboard/invitations", http.StatusSeeOther)
	}

	newFileName, err := filemanager.UploadFile(c, "image", "invitations")
	var oldPhotoToDelete string
	if newFileName != "" {
		oldPhotoToDelete = existingInvitation.Image
		existingInvitation.Image = newFileName
	}

	dateValue, _ := time.Parse("2006-01-02", req.Date)
	existingInvitation.CategoryID = req.CategoryID
	existingInvitation.Venue = req.Venue
	existingInvitation.Address = req.Address
	existingInvitation.Location = req.Location
	existingInvitation.Telephone = req.Telephone
	existingInvitation.Date = dateValue
	existingInvitation.Time = req.Time
	existingInvitation.IsConfirmed = req.IsConfirmed == "true"
	existingInvitation.IsParticipant = req.IsParticipant == "true"
	existingInvitation.IsFree = req.IsFree == "true"

	if existingInvitation.InvitationDetail != nil {
		existingInvitation.InvitationDetail.Title = req.Detail.Title
		existingInvitation.InvitationDetail.BrideName = req.Detail.BrideName
		existingInvitation.InvitationDetail.BrideSurname = req.Detail.BrideSurname
		existingInvitation.InvitationDetail.BrideMotherName = req.Detail.BrideMotherName
		existingInvitation.InvitationDetail.BrideMotherSurname = req.Detail.BrideMotherSurname
		existingInvitation.InvitationDetail.BrideFatherName = req.Detail.BrideFatherName
		existingInvitation.InvitationDetail.BrideFatherSurname = req.Detail.BrideFatherSurname
		existingInvitation.InvitationDetail.GroomName = req.Detail.GroomName
		existingInvitation.InvitationDetail.GroomSurname = req.Detail.GroomSurname
		existingInvitation.InvitationDetail.GroomMotherName = req.Detail.GroomMotherName
		existingInvitation.InvitationDetail.GroomMotherSurname = req.Detail.GroomMotherSurname
		existingInvitation.InvitationDetail.GroomFatherName = req.Detail.GroomFatherName
		existingInvitation.InvitationDetail.GroomFatherSurname = req.Detail.GroomFatherSurname
		existingInvitation.InvitationDetail.Person = req.Detail.Person
		existingInvitation.InvitationDetail.MotherName = req.Detail.MotherName
		existingInvitation.InvitationDetail.MotherSurname = req.Detail.MotherSurname
		existingInvitation.InvitationDetail.FatherName = req.Detail.FatherName
		existingInvitation.InvitationDetail.FatherSurname = req.Detail.FatherSurname
		existingInvitation.InvitationDetail.IsMotherLive = req.Detail.IsMotherLive == "true"
		existingInvitation.InvitationDetail.IsFatherLive = req.Detail.IsFatherLive == "true"
		existingInvitation.InvitationDetail.IsBrideMotherLive = req.Detail.IsBrideMotherLive == "true"
		existingInvitation.InvitationDetail.IsBrideFatherLive = req.Detail.IsBrideFatherLive == "true"
		existingInvitation.InvitationDetail.IsGroomMotherLive = req.Detail.IsGroomMotherLive == "true"
		existingInvitation.InvitationDetail.IsGroomFatherLive = req.Detail.IsGroomFatherLive == "true"
	}

	if err := h.invitationService.UpdateInvitationWithRelations(c.UserContext(), existingInvitation); err != nil {
		if newFileName != "" {
			filemanager.DeleteFile("invitations", newFileName)
		}
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Davetiye güncellenemedi: "+err.Error())
		return c.Redirect("/dashboard/invitations/update/"+strconv.Itoa(id), http.StatusSeeOther)
	}

	if oldPhotoToDelete != "" {
		filemanager.DeleteFile("invitations", oldPhotoToDelete)
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Davetiye başarıyla güncellendi.")
	return c.Redirect("/dashboard/invitations", http.StatusFound)
}

func (h *DashboardInvitationHandler) DeleteInvitation(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz ID")
	}

	invitation, err := h.invitationService.GetInvitationByID(uint(id))
	if err == nil && invitation.Image != "" {
		filemanager.DeleteFile("invitations", invitation.Image)
	}

	if err := h.invitationService.DeleteInvitationWithRelations(c.UserContext(), uint(id)); err != nil {
		errMsg := "Davetiye silinemedi: " + err.Error()
		if strings.Contains(c.Get("Accept"), "application/json") {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errMsg})
		}
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, errMsg)
		return c.Redirect("/dashboard/invitations", fiber.StatusSeeOther)
	}

	if strings.Contains(c.Get("Accept"), "application/json") {
		return c.JSON(fiber.Map{"message": "Davetiye başarıyla silindi."})
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Davetiye başarıyla silindi.")
	return c.Redirect("/dashboard/invitations", http.StatusFound)
}

func (h *DashboardInvitationHandler) ShowInvitation(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Geçersiz davetiye anahtarı.")
		return c.Redirect("/dashboard/invitations", http.StatusSeeOther)
	}

	invitation, err := h.invitationService.GetInvitationByKey(c.UserContext(), key)
	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Davetiye bulunamadı.")
		return c.Redirect("/dashboard/invitations", http.StatusSeeOther)
	}

	return renderer.Render(c, "dashboard/invitations/show", "layouts/dashboard", fiber.Map{
		"Title":      "Davetiye Detayları",
		"Invitation": invitation,
	})
}

func (h *DashboardInvitationHandler) renderInvitationFormError(c *fiber.Ctx, template, title string, req any, message string, fallback ...*models.Invitation) error {
	form, ok := req.(requests.InvitationRequest)
	if !ok {
		return c.Status(http.StatusInternalServerError).SendString("Sunucu Hatası")
	}

	invitation := &models.Invitation{
		CategoryID:    form.CategoryID,
		Venue:         form.Venue,
		Address:       form.Address,
		Location:      form.Location,
		Telephone:     form.Telephone,
		Time:          form.Time,
		IsConfirmed:   form.IsConfirmed == "true",
		IsParticipant: form.IsParticipant == "true",
		IsFree:        form.IsFree == "true",
		InvitationDetail: &models.InvitationDetail{
			Title:              form.Detail.Title,
			BrideName:          form.Detail.BrideName,
			BrideSurname:       form.Detail.BrideSurname,
			BrideMotherName:    form.Detail.BrideMotherName,
			BrideMotherSurname: form.Detail.BrideMotherSurname,
			BrideFatherName:    form.Detail.BrideFatherName,
			BrideFatherSurname: form.Detail.BrideFatherSurname,
			GroomName:          form.Detail.GroomName,
			GroomSurname:       form.Detail.GroomSurname,
			GroomMotherName:    form.Detail.GroomMotherName,
			GroomMotherSurname: form.Detail.GroomMotherSurname,
			GroomFatherName:    form.Detail.GroomFatherName,
			GroomFatherSurname: form.Detail.GroomFatherSurname,
			Person:             form.Detail.Person,
			MotherName:         form.Detail.MotherName,
			MotherSurname:      form.Detail.MotherSurname,
			FatherName:         form.Detail.FatherName,
			FatherSurname:      form.Detail.FatherSurname,
			IsMotherLive:       form.Detail.IsMotherLive == "true",
			IsFatherLive:       form.Detail.IsFatherLive == "true",
			IsBrideMotherLive:  form.Detail.IsBrideMotherLive == "true",
			IsBrideFatherLive:  form.Detail.IsBrideFatherLive == "true",
			IsGroomMotherLive:  form.Detail.IsGroomMotherLive == "true",
			IsGroomFatherLive:  form.Detail.IsGroomFatherLive == "true",
		},
	}

	if len(fallback) > 0 && fallback[0] != nil {
		invitation = fallback[0]
	}

	categories, _ := h.categoryService.GetAllCategories(queryparams.DefaultListParams())

	return renderer.Render(c, template, "layouts/dashboard", fiber.Map{
		"Title":                    title,
		renderer.FlashErrorKeyView: message,
		"Invitation":               invitation,
		"Categories":               categories,
	}, http.StatusBadRequest)
}
