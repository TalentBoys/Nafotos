package api

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"awesome-sharing/internal/database"
	"awesome-sharing/internal/middleware"
	"awesome-sharing/internal/models"
	"awesome-sharing/internal/services"
)

type ShareHandler struct {
	shareService        *services.ShareService
	settingsService     *services.SettingsService
	domainConfigService *services.DomainConfigService
	db                  *database.DB
	validator           *services.FileValidatorService
}

func NewShareHandler(shareService *services.ShareService, settingsService *services.SettingsService, domainConfigService *services.DomainConfigService, db *database.DB, validator *services.FileValidatorService) *ShareHandler {
	return &ShareHandler{
		shareService:        shareService,
		settingsService:     settingsService,
		domainConfigService: domainConfigService,
		db:                  db,
		validator:           validator,
	}
}

// ListShares returns all shares for the current user
// GET /api/shares
func (h *ShareHandler) ListShares(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	shares, err := h.shareService.ListSharesByOwner(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch shares",
		})
	}

	return c.JSON(fiber.Map{
		"shares": shares,
		"total":  len(shares),
	})
}

// GetShare returns a specific share
// GET /api/shares/:id
func (h *ShareHandler) GetShare(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id := c.Params("id")
	share, err := h.shareService.GetShare(id)
	if err != nil {
		if err == services.ErrShareNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Share not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch share",
		})
	}

	// Check ownership
	if share.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(fiber.Map{
		"share": share,
	})
}

// CreateShare creates a new share
// POST /api/shares
func (h *ShareHandler) CreateShare(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req struct {
		ShareType    string     `json:"share_type"`   // 'file' or 'album'
		ResourceID   int64      `json:"resource_id"`
		AccessType   string     `json:"access_type"`  // 'public' or 'private'
		Password     string     `json:"password"`
		RequiresAuth bool       `json:"requires_auth"`
		ExpiresIn    *int       `json:"expires_in"`   // Hours
		MaxViews     *int       `json:"max_views"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate
	if req.ShareType != "file" && req.ShareType != "album" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Share type must be 'file' or 'album'",
		})
	}

	if req.ResourceID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Resource ID is required",
		})
	}

	if req.AccessType == "" {
		req.AccessType = "public"
	}

	if req.AccessType != "public" && req.AccessType != "private" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Access type must be 'public' or 'private'",
		})
	}

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour)
		expiresAt = &expiry
	}

	share, err := h.shareService.CreateShare(
		req.ShareType,
		req.ResourceID,
		user.ID,
		req.AccessType,
		req.Password,
		req.RequiresAuth,
		expiresAt,
		req.MaxViews,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create share",
		})
	}

	// Get domain for full URL
	baseURL, err := h.domainConfigService.GetFullURL()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Domain not configured. Please configure the domain in settings first.",
		})
	}

	fullURL := baseURL + "/s/" + share.ID

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"share": share,
		"url":   fullURL,
	})
}

// UpdateShare updates a share
// PUT /api/shares/:id
func (h *ShareHandler) UpdateShare(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id := c.Params("id")

	// Check ownership
	share, err := h.shareService.GetShare(id)
	if err != nil {
		if err == services.ErrShareNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Share not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch share",
		})
	}

	if share.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	var req struct {
		Enabled      *bool   `json:"enabled"`
		MaxViews     *int    `json:"max_views"`
		Password     *string `json:"password"`
		RequiresAuth *bool   `json:"requires_auth"`
		ExpiresIn    *int    `json:"expires_in"` // Hours from now, null to remove expiration
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	updates := make(map[string]interface{})
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.MaxViews != nil {
		updates["max_views"] = *req.MaxViews
	}
	if req.Password != nil {
		updates["password"] = *req.Password
	}
	if req.RequiresAuth != nil {
		updates["requires_auth"] = *req.RequiresAuth
	}
	if req.ExpiresIn != nil {
		if *req.ExpiresIn > 0 {
			expiry := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour)
			updates["expires_at"] = &expiry
		} else {
			updates["expires_at"] = nil
		}
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No fields to update",
		})
	}

	err = h.shareService.UpdateShare(id, updates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update share",
		})
	}

	updatedShare, err := h.shareService.GetShare(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch updated share",
		})
	}

	return c.JSON(fiber.Map{
		"share": updatedShare,
	})
}

// DeleteShare deletes a share
// DELETE /api/shares/:id
func (h *ShareHandler) DeleteShare(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id := c.Params("id")

	// Check ownership
	share, err := h.shareService.GetShare(id)
	if err != nil {
		if err == services.ErrShareNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Share not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch share",
		})
	}

	if share.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	err = h.shareService.DeleteShare(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete share",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Share deleted successfully",
	})
}

// ExtendShare extends the expiration of a share
// POST /api/shares/:id/extend
func (h *ShareHandler) ExtendShare(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id := c.Params("id")

	// Check ownership
	share, err := h.shareService.GetShare(id)
	if err != nil {
		if err == services.ErrShareNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Share not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch share",
		})
	}

	if share.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	var req struct {
		Hours int `json:"hours"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Hours <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Hours must be positive",
		})
	}

	duration := time.Duration(req.Hours) * time.Hour
	err = h.shareService.ExtendShare(id, duration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to extend share",
		})
	}

	updatedShare, err := h.shareService.GetShare(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch updated share",
		})
	}

	return c.JSON(fiber.Map{
		"share": updatedShare,
	})
}

// DeleteExpiredShares deletes all expired shares
// DELETE /api/shares/expired
func (h *ShareHandler) DeleteExpiredShares(c *fiber.Ctx) error {
	count, err := h.shareService.DeleteExpiredShares()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete expired shares",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Expired shares deleted successfully",
		"deleted": count,
	})
}

// GetShareAccessLog returns access log for a share
// GET /api/shares/:id/access-log
func (h *ShareHandler) GetShareAccessLog(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id := c.Params("id")

	// Check ownership
	share, err := h.shareService.GetShare(id)
	if err != nil {
		if err == services.ErrShareNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Share not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch share",
		})
	}

	if share.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	logs, err := h.shareService.GetAccessLog(id, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch access log",
		})
	}

	return c.JSON(fiber.Map{
		"logs":  logs,
		"total": len(logs),
	})
}

// AccessShare - Public endpoint for accessing a share
// GET /api/s/:id
func (h *ShareHandler) AccessShare(c *fiber.Ctx) error {
	id := c.Params("id")
	password := c.Query("password", "")

	// Get user if authenticated (optional)
	var userID *int64
	user := middleware.GetUser(c)
	if user != nil {
		userID = &user.ID
	}

	// Validate access
	share, err := h.shareService.ValidateShareAccess(id, password, userID)
	if err != nil {
		if err == services.ErrShareNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Share not found",
			})
		}
		if err == services.ErrShareExpired {
			return c.Status(fiber.StatusGone).JSON(fiber.Map{
				"error": "This share has expired",
			})
		}
		if err == services.ErrShareDisabled {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "This share has been disabled",
			})
		}
		if err == services.ErrMaxViewsReached {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Maximum views reached for this share",
			})
		}
		if err == services.ErrInvalidPassword {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid password",
				"requires_password": true,
			})
		}
		if err == services.ErrAccessDenied {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":         "Access denied. Please login to access this share.",
				"requires_auth": true,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to access share",
		})
	}

	// Log access
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	err = h.shareService.LogAccess(id, userID, ipAddress, userAgent)
	if err != nil {
		// Log error but don't fail the request
		// log.Printf("Failed to log share access: %v", err)
	}

	// Refresh share to get updated view_count (after LogAccess incremented it)
	share, err = h.shareService.GetShare(id)
	if err != nil {
		// If refresh fails, continue with old data (non-critical)
		// The view_count will be off by 1, but share is still accessible
	}

	// Generate access token for accessing the shared resource
	accessToken, err := h.shareService.GenerateAccessToken(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate access token",
		})
	}

	return c.JSON(fiber.Map{
		"share":        share,
		"access_token": accessToken,
	})
}

// GrantSharePermission grants a user access to a private share
// POST /api/shares/:id/permissions
func (h *ShareHandler) GrantSharePermission(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id := c.Params("id")

	// Check ownership
	share, err := h.shareService.GetShare(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch share",
		})
	}

	if share.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	var req struct {
		UserID int64 `json:"user_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	err = h.shareService.GrantSharePermission(id, req.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to grant permission",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Permission granted successfully",
	})
}

// RevokeSharePermission revokes a user's access to a private share
// DELETE /api/shares/:id/permissions/:userId
func (h *ShareHandler) RevokeSharePermission(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id := c.Params("id")
	userId, err := strconv.ParseInt(c.Params("userId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Check ownership
	share, err := h.shareService.GetShare(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch share",
		})
	}

	if share.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	err = h.shareService.RevokeSharePermission(id, userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke permission",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Permission revoked successfully",
	})
}

// GetPublicFile - Public endpoint for accessing a file via share token
// GET /api/public/files/:id
func (h *ShareHandler) GetPublicFile(c *fiber.Ctx) error {
	fileIDStr := c.Params("id")
	token := c.Query("token", "")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Access token required",
		})
	}

	// Validate the access token
	_, resourceID, err := h.shareService.ValidateAccessToken(token)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Invalid or expired access token",
		})
	}

	// Parse file ID
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}

	// Verify the file ID matches the shared resource
	if fileID != resourceID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "File does not match shared resource",
		})
	}

	// Get the file
	var file models.File
	err = h.db.QueryRow(`
		SELECT id, filename, file_type, size, width, height, taken_at, created_at, updated_at
		FROM files WHERE id = ?
	`, fileID).Scan(&file.ID, &file.Filename, &file.FileType, &file.Size, &file.Width, &file.Height,
		&file.TakenAt, &file.CreatedAt, &file.UpdatedAt)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	// Validate file and get absolute path
	files := h.validator.ValidateFiles([]models.File{file})
	if len(files) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found or deleted",
		})
	}

	return c.JSON(files[0])
}

// DownloadPublicFile - Public endpoint for downloading a file via share token
// GET /api/public/files/:id/download
func (h *ShareHandler) DownloadPublicFile(c *fiber.Ctx) error {
	fileIDStr := c.Params("id")
	token := c.Query("token", "")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Access token required",
		})
	}

	// Validate the access token
	_, resourceID, err := h.shareService.ValidateAccessToken(token)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Invalid or expired access token",
		})
	}

	// Parse file ID
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}

	// Verify the file ID matches the shared resource
	if fileID != resourceID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "File does not match shared resource",
		})
	}

	// Get the file
	var file models.File
	err = h.db.QueryRow(`
		SELECT id, filename, file_type, size, width, height, taken_at, created_at, updated_at
		FROM files WHERE id = ?
	`, fileID).Scan(&file.ID, &file.Filename, &file.FileType, &file.Size, &file.Width, &file.Height,
		&file.TakenAt, &file.CreatedAt, &file.UpdatedAt)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	// Validate file and get absolute path
	files := h.validator.ValidateFiles([]models.File{file})
	if len(files) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found or deleted",
		})
	}

	// Set Content-Disposition header to force download
	c.Set("Content-Disposition", "attachment; filename=\""+files[0].Filename+"\"")

	// Send file
	return c.SendFile(files[0].AbsolutePath)
}
