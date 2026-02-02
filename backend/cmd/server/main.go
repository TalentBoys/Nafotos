package main

import (
	"awesome-sharing/internal/api"
	"awesome-sharing/internal/config"
	"awesome-sharing/internal/database"
	"awesome-sharing/internal/initialization"
	"awesome-sharing/internal/services"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Println("╔════════════════════════════════════════════════════════╗")
	log.Println("║          AwesomeSharing Server v2.0                   ║")
	log.Println("╚════════════════════════════════════════════════════════╝")
	log.Printf("Port: %s", cfg.Port)
	log.Printf("Config directory: %s", cfg.ConfigDir)
	log.Printf("Upload directory: %s", cfg.UploadDir)
	log.Printf("Database path: %s", cfg.DBPath)
	log.Println("")

	// Initialize database
	db, err := database.Initialize(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("✓ Database initialized successfully")

	// Initialize all services first (before any data operations)
	log.Println("\nInitializing services...")
	authService := services.NewAuthService(db.DB)
	settingsService := services.NewSettingsService(db.DB)
	folderService := services.NewFolderService(db.DB)
	permissionGroupService := services.NewPermissionGroupService(db.DB)
	albumService := services.NewAlbumService(db.DB)
	shareService := services.NewShareService(db.DB)
	domainConfigService := services.NewDomainConfigService(db)
	scanner := services.NewFileScanner(db, folderService, cfg.ThumbsDir)
	thumbService := services.NewThumbnailService(cfg.ThumbsDir)
	validatorService := services.NewFileValidatorService(db.DB, folderService)
	log.Println("✓ All services initialized")

	// Initialize default data (admin user, migrate mount points)
	log.Println("\nInitializing default data...")
	if err := initialization.InitializeDefaultData(db.DB); err != nil {
		log.Printf("Warning: Failed to initialize default data: %v", err)
	}

	// Initialize default mount points (legacy support)
	initializeMountPoints(db, cfg)

	// Wait a moment to ensure all initialization is complete
	time.Sleep(500 * time.Millisecond)

	// Start periodic scanning in the background (delay first scan)
	go func() {
		// Wait 5 seconds before first scan to avoid conflicts
		time.Sleep(5 * time.Second)
		log.Println("Starting initial folder scan...")
		scanner.ScanAllFolders()
		log.Println("✓ Initial scan complete")

		// Now start periodic scanning
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			scanner.ScanAllFolders()
		}
	}()
	log.Println("✓ Background file scanner scheduled (first scan in 5 seconds)")

	// Start periodic file validation and cleanup in background
	// Can be disabled with DISABLE_FILE_VALIDATION=true
	// Run AFTER the initial scan to avoid database lock conflicts
	if os.Getenv("DISABLE_FILE_VALIDATION") != "true" {
		go func() {
			// Wait 30 seconds to let initial scan complete
			time.Sleep(30 * time.Second)
			log.Println("Running initial file validation and cleanup...")
			if count, err := validatorService.CleanupAllInvalidFiles(); err == nil {
				if count > 0 {
					log.Printf("✓ Initial cleanup: removed %d missing files", count)
				} else {
					log.Println("✓ Initial cleanup: no invalid files found")
				}
			} else {
				log.Printf("✗ Initial cleanup failed: %v", err)
			}

			// Run cleanup every 6 hours
			ticker := time.NewTicker(6 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				if count, err := validatorService.CleanupAllInvalidFiles(); err == nil && count > 0 {
					log.Printf("✓ Periodic cleanup: removed %d missing files", count)
				}
			}
		}()
		log.Println("✓ Background file validator scheduled (first cleanup in 30 seconds, after initial scan)")
	} else {
		log.Println("⚠ File validation disabled by DISABLE_FILE_VALIDATION env var")
	}

	// Start periodic session cleanup
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			initialization.CleanupExpiredSessions(db.DB)
		}
	}()
	log.Println("✓ Session cleanup task started (1-hour interval)")

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "AwesomeSharing v2.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup all handlers
	handler := api.NewHandler(db, scanner, thumbService, validatorService, folderService, permissionGroupService)
	authHandler := api.NewAuthHandler(authService, settingsService)
	userHandler := api.NewUserHandler(authService)
	folderHandler := api.NewFolderHandler(folderService, scanner)
	permissionGroupHandler := api.NewPermissionGroupHandler(permissionGroupService)
	albumHandler := api.NewAlbumHandler(albumService)
	shareHandler := api.NewShareHandler(shareService, settingsService, domainConfigService, db, validatorService)
	settingsHandler := api.NewSettingsHandler(settingsService)
	domainConfigHandler := api.NewDomainConfigHandlers(domainConfigService)
	uploadHandler := api.NewUploadHandler(folderService, scanner)

	// Setup routes (v2 with authentication)
	api.SetupRoutesV2(
		app,
		db.DB,
		handler,
		authHandler,
		userHandler,
		folderHandler,
		permissionGroupHandler,
		albumHandler,
		shareHandler,
		settingsHandler,
		domainConfigHandler,
		uploadHandler,
		authService,
		cfg.AllowedOrigin,
	)

	log.Println("\n✓ API routes configured")
	log.Println("")
	log.Println("╔════════════════════════════════════════════════════════╗")
	log.Printf("║  🚀 Server READY at http://localhost:%s            ║", cfg.Port)
	log.Println("╚════════════════════════════════════════════════════════╝")
	log.Println("")
	log.Println("📋 API Endpoints:")
	log.Println("   Auth:            /api/auth/login, /api/auth/register")
	log.Println("   Users:           /api/users (admin)")
	log.Println("   Folders:         /api/folders")
	log.Println("   Permission Groups: /api/permission-groups")
	log.Println("   Albums:          /api/albums-v2")
	log.Println("   Shares:          /api/shares")
	log.Println("   Settings:        /api/settings (admin)")
	log.Println("   Public:          /api/s/:id (share access)")
	log.Println("")
	log.Println("✅ SERVER IS NOW ACCEPTING CONNECTIONS")
	log.Println("   Default login: admin / admin")
	log.Println("")

	// Start server
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initializeMountPoints(db *database.DB, cfg *config.Config) {
	mountPoints := []struct {
		Path string
		Name string
	}{
		{cfg.ConfigDir, "Config"},
		{cfg.UploadDir, "Upload"},
	}

	for _, mp := range mountPoints {
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM mount_points WHERE path = ?", mp.Path).Scan(&exists)
		if err != nil {
			log.Printf("Error checking mount point: %v", err)
			continue
		}

		if exists == 0 {
			_, err := db.Exec("INSERT INTO mount_points (path, name, enabled) VALUES (?, ?, 1)", mp.Path, mp.Name)
			if err != nil {
				log.Printf("Error creating mount point: %v", err)
			} else {
				log.Printf("Created mount point: %s (%s)", mp.Name, mp.Path)
			}
		}
	}
}
