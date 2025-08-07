package routes

import (
	"zatrano/configs/limiterconfig"
	"zatrano/middlewares"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func SetupRoutes(app *fiber.App) {
	app.Use(limiter.New(limiterconfig.GetLimiterConfig()))

	app.Use(middlewares.SessionMiddleware())

	app.Use(middlewares.ZapLogger())

	registerWebsiteRoutes(app)
	registerAuthRoutes(app)
	registerDashboardRoutes(app)
	registerPanelRoutes(app)
}
