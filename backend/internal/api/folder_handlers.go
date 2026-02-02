package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"awesome-sharing/internal/middleware"
	"awesome-sharing/internal/services"
)

type FolderHandler struct {
	folderService  *services.FolderService
	scannerService *services.FileScanner
}

func NewFolderHandler(folderService *services.FolderService, scannerService *services.FileScanner) *FolderHandler {
	return &FolderHandler{
		folderService:  folderService,
		scannerService: scannerService,
	}
}

// CreateFolder creates a new folder
// POST /api/folders
func (h *FolderHandler) CreateFolder(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can create folders
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	var req struct {
		Name         string `json:"name"`
		AbsolutePath string `json:"absolute_path"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" || req.AbsolutePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and absolute path are required",
		})
	}

	folder, err := h.folderService.CreateFolder(req.Name, req.AbsolutePath, user.ID)
	if err != nil {
		if err == services.ErrFolderPathConflict {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Folder path conflicts with existing folder (parent/child relationship)",
			})
		}
		if err == services.ErrFolderPathNotAbsolute {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Folder path must be absolute",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create folder",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"folder": folder,
	})
}

// ListFolders lists all folders accessible to the user
// GET /api/folders
func (h *FolderHandler) ListFolders(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	isAdmin := user.Role == "admin" || user.Role == "server_owner"
	folders, err := h.folderService.ListFolders(user.ID, isAdmin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list folders",
		})
	}

	return c.JSON(fiber.Map{
		"folders": folders,
		"total":   len(folders),
	})
}

// GetFolder retrieves a specific folder
// GET /api/folders/:id
func (h *FolderHandler) GetFolder(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid folder ID",
		})
	}

	folder, err := h.folderService.GetFolder(id)
	if err != nil {
		if err == services.ErrFolderNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Folder not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get folder",
		})
	}

	return c.JSON(fiber.Map{
		"folder": folder,
	})
}

// UpdateFolder updates a folder
// PUT /api/folders/:id
func (h *FolderHandler) UpdateFolder(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can update folders
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid folder ID",
		})
	}

	var req struct {
		Name         string `json:"name"`
		AbsolutePath string `json:"absolute_path"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	err = h.folderService.UpdateFolder(id, req.Name, req.AbsolutePath)
	if err != nil {
		if err == services.ErrFolderPathConflict {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Folder path conflicts with existing folder",
			})
		}
		if err == services.ErrFolderPathNotAbsolute {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Folder path must be absolute",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update folder",
		})
	}

	updatedFolder, err := h.folderService.GetFolder(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get updated folder",
		})
	}

	return c.JSON(fiber.Map{
		"folder": updatedFolder,
	})
}

// DeleteFolder deletes a folder
// DELETE /api/folders/:id
func (h *FolderHandler) DeleteFolder(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can delete folders
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid folder ID",
		})
	}

	err = h.folderService.DeleteFolder(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete folder",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Folder deleted successfully",
	})
}

// ToggleFolder enables/disables a folder
// PUT /api/folders/:id/toggle
func (h *FolderHandler) ToggleFolder(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can toggle folders
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid folder ID",
		})
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = h.folderService.ToggleFolder(id, req.Enabled)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to toggle folder",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Folder toggled successfully",
	})
}

// ScanFolder triggers a scan of a specific folder
// POST /api/folders/:id/scan
func (h *FolderHandler) ScanFolder(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can trigger scans
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid folder ID",
		})
	}

	// Run scan in background
	go func() {
		if err := h.scannerService.ScanFolder(id); err != nil {
			// Log error but don't fail the request
		}
	}()

	return c.JSON(fiber.Map{
		"message": "Folder scan started",
	})
}

// ListFilesInFolder lists all files in a folder
// GET /api/folders/:id/files
func (h *FolderHandler) ListFilesInFolder(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid folder ID",
		})
	}

	// Parse pagination parameters
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	files, err := h.folderService.ListFilesInFolder(id, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list files",
		})
	}

	count, _ := h.folderService.CountFilesInFolder(id)

	return c.JSON(fiber.Map{
		"files":  files,
		"total":  count,
		"limit":  limit,
		"offset": offset,
	})
}

// BrowseDirectoryTree browses the file system directory tree
// POST /api/folders/browse
func (h *FolderHandler) BrowseDirectoryTree(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can browse directories
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	var req struct {
		Path string `json:"path"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// If path is empty, use root directory
	if req.Path == "" {
		req.Path = "/"
	}

	directories, err := h.folderService.BrowseDirectory(req.Path)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to browse directory: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"path":        req.Path,
		"directories": directories,
	})
}
