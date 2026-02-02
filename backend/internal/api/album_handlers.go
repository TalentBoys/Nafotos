package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"awesome-sharing/internal/middleware"
	"awesome-sharing/internal/services"
)

type AlbumHandler struct {
	albumService *services.AlbumService
}

func NewAlbumHandler(albumService *services.AlbumService) *AlbumHandler {
	return &AlbumHandler{
		albumService: albumService,
	}
}

// ListAlbums returns all albums for the current user
// GET /api/albums
func (h *AlbumHandler) ListAlbums(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	albums, err := h.albumService.ListAlbums(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch albums",
		})
	}

	return c.JSON(fiber.Map{
		"albums": albums,
		"total":  len(albums),
	})
}

// GetAlbum returns a specific album
// GET /api/albums/:id
func (h *AlbumHandler) GetAlbum(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid album ID",
		})
	}

	album, err := h.albumService.GetAlbum(id)
	if err != nil {
		if err == services.ErrAlbumNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Album not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch album",
		})
	}

	// Check ownership
	if album.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(fiber.Map{
		"album": album,
	})
}

// CreateAlbum creates a new album
// POST /api/albums
func (h *AlbumHandler) CreateAlbum(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req struct {
		Name        string                        `json:"name"`
		Description string                        `json:"description"`
		Folders     []services.FolderConfig       `json:"folders"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Album name is required",
		})
	}

	album, err := h.albumService.CreateAlbum(req.Name, req.Description, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create album",
		})
	}

	// Add folder configurations if provided
	if len(req.Folders) > 0 {
		if err := h.albumService.AddFolders(album.ID, req.Folders); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to add folders to album",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"album": album,
	})
}

// UpdateAlbum updates an album
// PUT /api/albums/:id
func (h *AlbumHandler) UpdateAlbum(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid album ID",
		})
	}

	// Check ownership
	album, err := h.albumService.GetAlbum(id)
	if err != nil {
		if err == services.ErrAlbumNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Album not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch album",
		})
	}

	if album.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		CoverFileID *int64 `json:"cover_file_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Album name is required",
		})
	}

	err = h.albumService.UpdateAlbum(id, req.Name, req.Description, req.CoverFileID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update album",
		})
	}

	updatedAlbum, err := h.albumService.GetAlbum(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch updated album",
		})
	}

	return c.JSON(fiber.Map{
		"album": updatedAlbum,
	})
}

// DeleteAlbum deletes an album
// DELETE /api/albums/:id
func (h *AlbumHandler) DeleteAlbum(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid album ID",
		})
	}

	// Check ownership
	album, err := h.albumService.GetAlbum(id)
	if err != nil {
		if err == services.ErrAlbumNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Album not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch album",
		})
	}

	if album.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	err = h.albumService.DeleteAlbum(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete album",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Album deleted successfully",
	})
}

// ListAlbumItems returns all items in an album with file details
// GET /api/albums/:id/items
func (h *AlbumHandler) ListAlbumItems(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid album ID",
		})
	}

	// Check ownership
	album, err := h.albumService.GetAlbum(id)
	if err != nil {
		if err == services.ErrAlbumNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Album not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch album",
		})
	}

	if album.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Get sort order from query parameter (default: taken_at DESC)
	sortOrder := c.Query("sort", "taken_at DESC")

	files, err := h.albumService.ListItemsWithFiles(id, sortOrder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch album items",
		})
	}

	return c.JSON(fiber.Map{
		"files": files,
		"total": len(files),
	})
}

// AddAlbumFolders adds folder configurations to an album
// POST /api/albums/:id/folders
func (h *AlbumHandler) AddAlbumFolders(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid album ID",
		})
	}

	// Check ownership
	album, err := h.albumService.GetAlbum(id)
	if err != nil {
		if err == services.ErrAlbumNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Album not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch album",
		})
	}

	if album.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	var req struct {
		Folders []services.FolderConfig `json:"folders"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Folders) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one folder configuration is required",
		})
	}

	err = h.albumService.AddFolders(id, req.Folders)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add folders to album",
		})
	}

	// Get file count for the updated album
	count, _ := h.albumService.GetAlbumFileCount(id)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Folders added successfully",
		"count":   count,
	})
}

// ListAlbumFolders returns folder configurations for an album
// GET /api/albums/:id/folders
func (h *AlbumHandler) ListAlbumFolders(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid album ID",
		})
	}

	// Check ownership
	album, err := h.albumService.GetAlbum(id)
	if err != nil {
		if err == services.ErrAlbumNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Album not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch album",
		})
	}

	if album.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	folders, err := h.albumService.ListAlbumFolders(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch album folders",
		})
	}

	return c.JSON(fiber.Map{
		"folders": folders,
		"total":   len(folders),
	})
}

// RemoveAlbumFolder removes a folder configuration from an album
// DELETE /api/albums/:id/folders/:folderId
func (h *AlbumHandler) RemoveAlbumFolder(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid album ID",
		})
	}

	folderId, err := strconv.ParseInt(c.Params("folderId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid folder ID",
		})
	}

	// Check ownership
	album, err := h.albumService.GetAlbum(id)
	if err != nil {
		if err == services.ErrAlbumNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Album not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch album",
		})
	}

	if album.OwnerID != user.ID && user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Get path_prefix from query parameter
	pathPrefix := c.Query("path_prefix", "")

	err = h.albumService.RemoveFolder(id, folderId, pathPrefix)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove folder from album",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Folder removed successfully",
	})
}
