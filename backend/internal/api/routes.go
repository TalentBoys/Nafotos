package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App, handler *Handler, allowedOrigin string) {
	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigin,
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// API routes
	api := app.Group("/api")

	// Files
	api.Get("/files", handler.GetFiles)
	api.Get("/files/:id", handler.GetFileByID)
	api.Get("/files/:id/thumbnail", handler.GetFileThumbnail)
	api.Get("/files/:id/download", handler.DownloadFile)

	// Timeline
	api.Get("/timeline", handler.GetTimeline)

	// Search
	api.Get("/search", handler.SearchFiles)

	// Mount points
	api.Get("/mount-points", handler.GetMountPoints)

	// Scan
	api.Post("/scan", handler.TriggerScan)

	// Cleanup
	api.Post("/cleanup", handler.CleanupDeletedFiles)

	// Tags
	api.Get("/tags", handler.GetTags)
	api.Post("/tags", handler.CreateTag)

	// Albums
	api.Get("/albums", handler.GetAlbums)
	api.Post("/albums", handler.CreateAlbum)

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
