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

type DashboardUserHandler struct {
	userService services.IUserService
}

func NewDashboardUserHandler() *DashboardUserHandler {
	return &DashboardUserHandler{userService: services.NewUserService()}
}

func (h *DashboardUserHandler) ListUsers(c *fiber.Ctx) error {
	var params queryparams.ListParams

	if err := c.QueryParser(&params); err != nil {
		params = queryparams.ListParams{}
	}

	params.ApplyDefaults()
	params.OrderBy = "asc"
	params.SortBy = "name"

	paginatedResult, err := h.userService.GetAllUsers(params)

	renderData := fiber.Map{
		"Title":  "Kullanıcılar",
		"Result": paginatedResult,
		"Params": params,
	}
	if err != nil {
		renderData[renderer.FlashErrorKeyView] = "Kullanıcılar getirilirken bir hata oluştu."
		renderData["Result"] = &queryparams.PaginatedResult{
			Data: []models.User{},
			Meta: queryparams.PaginationMeta{CurrentPage: params.Page, PerPage: params.PerPage},
		}
	}
	return renderer.Render(c, "dashboard/users/list", "layouts/dashboard", renderData, http.StatusOK)
}

func (h *DashboardUserHandler) ShowCreateUser(c *fiber.Ctx) error {
	return renderer.Render(c, "dashboard/users/create", "layouts/dashboard", fiber.Map{
		"Title": "Yeni Kullanıcı Ekle",
	})
}

func (h *DashboardUserHandler) CreateUser(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidateUserRequest(c)

	if err != nil {
		return renderUserFormError(c, "dashboard/users/create", "Yeni Kullanıcı Ekle", req, err.Error())
	}

	if req.Type != string(models.Dashboard) && req.Type != string(models.Panel) {
		return renderUserFormError(c, "dashboard/users/create", "Yeni Kullanıcı Ekle", req, "Geçersiz kullanıcı tipi seçildi.")
	}

	user := &models.User{
		Name:              req.Name,
		Email:             req.Email,
		Password:          req.Password,
		Status:            req.Status == "true",
		Type:              models.UserType(req.Type),
		ResetToken:        req.ResetToken,
		EmailVerified:     req.EmailVerified == "true",
		VerificationToken: req.VerificationToken,
		Provider:          req.Provider,
		ProviderID:        req.ProviderID,
	}

	if err := h.userService.CreateUser(c.UserContext(), user); err != nil {
		return renderUserFormError(c, "dashboard/users/create", "Yeni Kullanıcı Ekle", req, "Kullanıcı oluşturulamadı: "+err.Error())
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Kullanıcı başarıyla oluşturuldu.")
	return c.Redirect("/dashboard/users", fiber.StatusFound)
}

func (h *DashboardUserHandler) ShowUpdateUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Geçersiz kullanıcı ID")
	}

	user, err := h.userService.GetUserByID(uint(id))

	if err != nil {
		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, "Kullanıcı bulunamadı.")
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}

	return renderer.Render(c, "dashboard/users/update", "layouts/dashboard", fiber.Map{
		"Title": "Kullanıcı Düzenle",
		"User":  user,
	})
}

func (h *DashboardUserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz kullanıcı ID")
	}

	req, err := requests.ParseAndValidateUserRequest(c)

	if err != nil {
		return renderUserFormError(c, "dashboard/users/update", "Kullanıcı Düzenle", req, err.Error())
	}

	if req.Type != string(models.Dashboard) && req.Type != string(models.Panel) {
		return renderUserFormError(c, "dashboard/users/update", "Kullanıcı Düzenle", req, "Geçersiz kullanıcı tipi seçildi.")
	}

	user := &models.User{
		Name:              req.Name,
		Email:             req.Email,
		Status:            req.Status == "true",
		Type:              models.UserType(req.Type),
		ResetToken:        req.ResetToken,
		EmailVerified:     req.EmailVerified == "true",
		VerificationToken: req.VerificationToken,
		Provider:          req.Provider,
		ProviderID:        req.ProviderID,
	}

	if req.Password != "" {
		user.Password = req.Password
	}

	userID, _ := c.Locals("userID").(uint)

	if err := h.userService.UpdateUser(c.UserContext(), uint(id), user, userID); err != nil {
		return renderUserFormError(c, "dashboard/users/update", "Kullanıcı Düzenle", req, "Kullanıcı güncellenemedi: "+err.Error())
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Kullanıcı başarıyla güncellendi.")
	return c.Redirect("/dashboard/users", fiber.StatusFound)
}

func (h *DashboardUserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Geçersiz kullanıcı ID")
	}

	if err := h.userService.DeleteUser(c.UserContext(), uint(id)); err != nil {
		errMsg := "Kullanıcı silinemedi: " + err.Error()

		if strings.Contains(c.Get("Accept"), "application/json") {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errMsg})
		}

		_ = flashmessages.SetFlashMessage(c, flashmessages.FlashErrorKey, errMsg)
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}

	if strings.Contains(c.Get("Accept"), "application/json") {
		return c.JSON(fiber.Map{"message": "Kullanıcı başarıyla silindi."})
	}

	_ = flashmessages.SetFlashMessage(c, flashmessages.FlashSuccessKey, "Kullanıcı başarıyla silindi.")
	return c.Redirect("/dashboard/users", fiber.StatusFound)
}

func renderUserFormError(c *fiber.Ctx, template, title string, req any, message string) error {
	form, ok := req.(requests.UserRequest)
	if !ok {
		return c.Status(http.StatusInternalServerError).SendString("Sunucu Hatası")
	}

	user := &models.User{
		Name:              form.Name,
		Email:             form.Email,
		Status:            form.Status == "true",
		Type:              models.UserType(form.Type),
		ResetToken:        form.ResetToken,
		EmailVerified:     form.EmailVerified == "true",
		VerificationToken: form.VerificationToken,
		Provider:          form.Provider,
		ProviderID:        form.ProviderID,
	}

	if form.Password != "" {
		user.Password = form.Password
	}

	return renderer.Render(c, template, "layouts/dashboard", fiber.Map{
		"Title":                    title,
		renderer.FlashErrorKeyView: message,
		"User":                     user,
	}, http.StatusBadRequest)
}
