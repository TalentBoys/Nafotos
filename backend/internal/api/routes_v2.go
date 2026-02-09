package api

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"awesome-sharing/internal/middleware"
	"awesome-sharing/internal/services"
)

// SetupRoutesV2 sets up all API routes including new authentication and features
func SetupRoutesV2(
	app *fiber.App,
	db *sql.DB,
	handler *Handler,
	authHandler *AuthHandler,
	userHandler *UserHandler,
	folderHandler *FolderHandler,
	permissionGroupHandler *PermissionGroupHandler,
	albumHandler *AlbumHandler,
	shareHandler *ShareHandler,
	settingsHandler *SettingsHandler,
	domainConfigHandler *DomainConfigHandlers,
	uploadHandler *UploadHandler,
	authService *services.AuthService,
	allowedOrigin string,
) {
	// Middleware
	app.Use(logger.New())

	// CORS configuration
	corsConfig := cors.Config{
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		ExposeHeaders:    "Set-Cookie",
	}

	// Handle CORS based on allowed origin
	if allowedOrigin == "*" {
		// Wildcard origin - cannot use credentials
		corsConfig.AllowOrigins = "*"
		corsConfig.AllowCredentials = false
	} else {
		// Specific origin - can use credentials
		corsConfig.AllowOrigins = allowedOrigin
		corsConfig.AllowCredentials = true
	}

	app.Use(cors.New(corsConfig))

	// API routes
	api := app.Group("/api")

	// Public routes (no authentication required)
	public := api.Group("")
	{
		// Health check
		public.Get("/health", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"status": "ok"})
		})

		// Public settings
		public.Get("/settings/public", settingsHandler.GetPublicSettings)

		// Public share access (with optional auth to support requires_auth)
		public.Get("/s/:id", middleware.OptionalAuthMiddleware(authService), shareHandler.AccessShare)

		// Public file access (requires valid share token)
		public.Get("/public/files/:id", shareHandler.GetPublicFile)
		public.Get("/public/files/:id/download", shareHandler.DownloadPublicFile)
	}

	// Auth routes (some require auth, some don't)
	auth := api.Group("/auth")
	{
		auth.Post("/login", authHandler.Login)
		auth.Post("/register", middleware.OptionalAuthMiddleware(authService), authHandler.Register)
		auth.Post("/logout", middleware.AuthMiddleware(authService), authHandler.Logout)
		auth.Get("/me", middleware.AuthMiddleware(authService), authHandler.Me)
		auth.Post("/change-password", middleware.AuthMiddleware(authService), authHandler.ChangePassword)
	}

	// Protected routes (require authentication)
	protected := api.Group("", middleware.AuthMiddleware(authService))
	{
		// Legacy file routes (keep for backwards compatibility)
		protected.Get("/files", handler.GetFiles)
		protected.Get("/files/:id", handler.GetFileByID)
		protected.Get("/files/:id/thumbnail", handler.GetFileThumbnail)
		protected.Get("/files/:id/download", handler.DownloadFile)
		protected.Get("/timeline", handler.GetTimeline)
		protected.Get("/timeline/years", handler.GetTimelineYears)
		protected.Get("/search", handler.SearchFiles)
		protected.Get("/mount-points", handler.GetMountPoints)
		protected.Post("/scan", handler.TriggerScan)
		protected.Post("/cleanup", handler.CleanupDeletedFiles)
		protected.Get("/tags", handler.GetTags)
		protected.Post("/tags", handler.CreateTag)

		// Legacy album routes (keep for compatibility)
		protected.Get("/albums", handler.GetAlbums)
		protected.Post("/albums", handler.CreateAlbum)

		// User management (admin only)
		users := protected.Group("/users", middleware.AdminOrOwnerMiddleware())
		{
			users.Get("", userHandler.ListUsers)
			users.Get("/search", userHandler.SearchUsers)
			users.Get("/stats", userHandler.GetUserStats)
			users.Post("", userHandler.CreateUser)
			users.Post("/export", userHandler.ExportUsers)
			users.Post("/bulk/enable-disable", userHandler.BulkEnableDisable)
			users.Post("/bulk/delete", userHandler.BulkDelete)
			users.Get("/:id", userHandler.GetUser)
			users.Put("/:id", userHandler.UpdateUser)
			users.Delete("/:id", userHandler.DeleteUser)
			users.Put("/:id/toggle", userHandler.ToggleUser)
			users.Post("/:id/reset-password", userHandler.ResetPassword)
			users.Get("/:id/activity-logs", userHandler.GetUserActivityLogs)
		}

		// Folders (replaces libraries)
		folders := protected.Group("/folders")
		{
			folders.Get("", folderHandler.ListFolders)
			folders.Post("", middleware.AdminOnlyMiddleware(), folderHandler.CreateFolder)
			folders.Post("/browse", middleware.AdminOnlyMiddleware(), folderHandler.BrowseDirectoryTree)
			folders.Get("/:id", folderHandler.GetFolder)
			folders.Put("/:id", middleware.AdminOnlyMiddleware(), folderHandler.UpdateFolder)
			folders.Delete("/:id", middleware.AdminOnlyMiddleware(), folderHandler.DeleteFolder)

			// Folder operations
			folders.Put("/:id/toggle", middleware.AdminOnlyMiddleware(), folderHandler.ToggleFolder)
			folders.Post("/:id/scan", middleware.AdminOnlyMiddleware(), folderHandler.ScanFolder)

			// Folder files
			folders.Get("/:id/files", folderHandler.ListFilesInFolder)
		}

		// Permission Groups (for managing folder access)
		permissionGroups := protected.Group("/permission-groups")
		{
			permissionGroups.Get("", permissionGroupHandler.ListPermissionGroups)
			permissionGroups.Post("", middleware.AdminOnlyMiddleware(), permissionGroupHandler.CreatePermissionGroup)
			permissionGroups.Get("/:id", permissionGroupHandler.GetPermissionGroup)
			permissionGroups.Put("/:id", middleware.AdminOnlyMiddleware(), permissionGroupHandler.UpdatePermissionGroup)
			permissionGroups.Delete("/:id", middleware.AdminOnlyMiddleware(), permissionGroupHandler.DeletePermissionGroup)

			// Folder management in permission groups
			permissionGroups.Get("/:id/folders", permissionGroupHandler.ListFoldersInGroup)
			permissionGroups.Post("/:id/folders", middleware.AdminOnlyMiddleware(), permissionGroupHandler.AddFolderToGroup)
			permissionGroups.Delete("/:id/folders/:folderId", middleware.AdminOnlyMiddleware(), permissionGroupHandler.RemoveFolderFromGroup)

			// Permission management
			permissionGroups.Get("/:id/permissions", permissionGroupHandler.ListPermissions)
			permissionGroups.Post("/:id/permissions", middleware.AdminOnlyMiddleware(), permissionGroupHandler.GrantPermission)
			permissionGroups.Delete("/:id/permissions/:userId", middleware.AdminOnlyMiddleware(), permissionGroupHandler.RevokePermission)
		}

		// Enhanced albums (v2)
		albums := protected.Group("/albums-v2")
		{
			albums.Get("", albumHandler.ListAlbums)
			albums.Post("", albumHandler.CreateAlbum)
			albums.Get("/:id", albumHandler.GetAlbum)
			albums.Put("/:id", albumHandler.UpdateAlbum)
			albums.Delete("/:id", albumHandler.DeleteAlbum)

			// Album items (dynamic query from file_folder_mappings)
			albums.Get("/:id/items", albumHandler.ListAlbumItems)

			// Album folders (folder-based configuration)
			albums.Get("/:id/folders", albumHandler.ListAlbumFolders)
			albums.Post("/:id/folders", albumHandler.AddAlbumFolders)
			albums.Delete("/:id/folders/:folderId", albumHandler.RemoveAlbumFolder)
		}

		// Shares
		shares := protected.Group("/shares")
		{
			shares.Get("", shareHandler.ListShares)
			shares.Post("", shareHandler.CreateShare)
			shares.Get("/:id", shareHandler.GetShare)
			shares.Put("/:id", shareHandler.UpdateShare)
			shares.Delete("/:id", shareHandler.DeleteShare)

			// Share operations
			shares.Post("/:id/extend", shareHandler.ExtendShare)
			shares.Get("/:id/access-log", shareHandler.GetShareAccessLog)

			// Share permissions (for private shares)
			shares.Post("/:id/permissions", shareHandler.GrantSharePermission)
			shares.Delete("/:id/permissions/:userId", shareHandler.RevokeSharePermission)

			// Bulk operations
			shares.Delete("/expired", shareHandler.DeleteExpiredShares)
		}

		// Upload
		upload := protected.Group("/upload")
		{
			upload.Post("", uploadHandler.UploadFiles)
			upload.Post("/browse", uploadHandler.BrowseUploadTarget)
			upload.Post("/create-directory", uploadHandler.CreateDirectory)
		}

		// System settings (admin only)
		settings := protected.Group("/settings", middleware.AdminOnlyMiddleware())
		{
			settings.Get("", settingsHandler.GetSettings)
			settings.Put("", settingsHandler.UpdateSettings)
			settings.Get("/domain", settingsHandler.GetDomain)
			settings.Put("/domain", settingsHandler.UpdateDomain)
		}

		// Domain configuration (admin only)
		domainConfig := protected.Group("/domain-config", middleware.AdminOnlyMiddleware())
		{
			domainConfig.Get("", domainConfigHandler.GetDomainConfig)
			domainConfig.Post("", domainConfigHandler.SaveDomainConfig)
		}
	}
}
