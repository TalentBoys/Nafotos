package api

import (
	"awesome-sharing/internal/database"
	"awesome-sharing/internal/middleware"
	"awesome-sharing/internal/models"
	"awesome-sharing/internal/services"
	"database/sql"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	db            *database.DB
	scanner       *services.FileScanner
	thumbService  *services.ThumbnailService
	validator     *services.FileValidatorService
	folderService *services.FolderService
	permService   *services.PermissionGroupService
}

func NewHandler(db *database.DB, scanner *services.FileScanner, thumbService *services.ThumbnailService, validator *services.FileValidatorService, folderService *services.FolderService, permService *services.PermissionGroupService) *Handler {
	return &Handler{
		db:            db,
		scanner:       scanner,
		thumbService:  thumbService,
		validator:     validator,
		folderService: folderService,
		permService:   permService,
	}
}

// GetFiles returns a list of files with pagination
func (h *Handler) GetFiles(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	fileType := c.Query("type", "")
	offset := (page - 1) * limit

	isServerOwner := user.Role == "server_owner"

	var query string
	args := []interface{}{}

	if isServerOwner {
		// Server owner can see all files
		query = `SELECT f.id, f.filename, f.file_type, f.size, f.created_at, f.updated_at,
		                pm.width, pm.height, pm.taken_at
		         FROM files f
		         LEFT JOIN photo_metadata pm ON f.id = pm.file_id
		         WHERE 1=1`
	} else {
		// Regular users can only see files they have permission for through permission groups
		query = `SELECT DISTINCT f.id, f.filename, f.file_type, f.size, f.created_at, f.updated_at,
		                pm.width, pm.height, pm.taken_at
		         FROM files f
		         LEFT JOIN photo_metadata pm ON f.id = pm.file_id
		         JOIN file_folder_mappings ffm ON f.id = ffm.file_id
		         JOIN permission_group_folders pgf ON ffm.folder_id = pgf.folder_id
		         JOIN permission_group_permissions pgp ON pgf.permission_group_id = pgp.permission_group_id
		         WHERE pgp.user_id = ?`
		args = append(args, user.ID)
	}

	if fileType != "" {
		query += " AND f.file_type = ?"
		args = append(args, fileType)
	}

	query += " ORDER BY pm.taken_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	files := []models.File{}
	for rows.Next() {
		var f models.File
		var width, height sql.NullInt32
		var takenAt sql.NullTime
		if err := rows.Scan(&f.ID, &f.Filename, &f.FileType, &f.Size, &f.CreatedAt, &f.UpdatedAt,
			&width, &height, &takenAt); err != nil {
			log.Printf("Error scanning file: %v", err)
			continue
		}
		// Populate photo metadata fields if present
		if width.Valid {
			f.Width = int(width.Int32)
		}
		if height.Valid {
			f.Height = int(height.Int32)
		}
		if takenAt.Valid {
			f.TakenAt = &takenAt.Time
		}
		f.ThumbnailURL = "/api/files/" + strconv.FormatInt(f.ID, 10) + "/thumbnail"
		files = append(files, f)
	}

	// Validate files and filter out deleted ones, also resolves absolute_path
	files = h.validator.ValidateFiles(files)

	return c.JSON(fiber.Map{
		"files": files,
		"page":  page,
		"limit": limit,
	})
}

// GetTimeline returns files grouped by date
func (h *Handler) GetTimeline(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset := (page - 1) * limit

	isServerOwner := user.Role == "server_owner"

	var query string
	var args []interface{}

	if isServerOwner {
		// Server owner can see all files
		query = `SELECT f.id, f.filename, f.file_type, f.size, f.created_at, f.updated_at,
		                pm.width, pm.height, pm.taken_at
		         FROM files f
		         LEFT JOIN photo_metadata pm ON f.id = pm.file_id
		         WHERE pm.taken_at IS NOT NULL
		         ORDER BY pm.taken_at DESC
		         LIMIT ? OFFSET ?`
		args = []interface{}{limit, offset}
	} else {
		// Regular users can only see files they have permission for
		query = `SELECT DISTINCT f.id, f.filename, f.file_type, f.size, f.created_at, f.updated_at,
		                pm.width, pm.height, pm.taken_at
		         FROM files f
		         LEFT JOIN photo_metadata pm ON f.id = pm.file_id
		         JOIN file_folder_mappings ffm ON f.id = ffm.file_id
		         JOIN permission_group_folders pgf ON ffm.folder_id = pgf.folder_id
		         JOIN permission_group_permissions pgp ON pgf.permission_group_id = pgp.permission_group_id
		         WHERE pm.taken_at IS NOT NULL AND pgp.user_id = ?
		         ORDER BY pm.taken_at DESC
		         LIMIT ? OFFSET ?`
		args = []interface{}{user.ID, limit, offset}
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	files := []models.File{}
	for rows.Next() {
		var f models.File
		var width, height sql.NullInt32
		var takenAt sql.NullTime
		if err := rows.Scan(&f.ID, &f.Filename, &f.FileType, &f.Size, &f.CreatedAt, &f.UpdatedAt,
			&width, &height, &takenAt); err != nil {
			continue
		}
		// Populate photo metadata fields if present
		if width.Valid {
			f.Width = int(width.Int32)
		}
		if height.Valid {
			f.Height = int(height.Int32)
		}
		if takenAt.Valid {
			f.TakenAt = &takenAt.Time
		}
		f.ThumbnailURL = "/api/files/" + strconv.FormatInt(f.ID, 10) + "/thumbnail"
		files = append(files, f)
	}

	// Validate files and filter out deleted ones
	files = h.validator.ValidateFiles(files)

	return c.JSON(fiber.Map{
		"files": files,
		"page":  page,
		"limit": limit,
	})
}

// GetFileByID returns a single file's details
func (h *Handler) GetFileByID(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid file ID"})
	}

	// Check if user has access to this file
	isServerOwner := user.Role == "server_owner"
	if !isServerOwner {
		hasAccess, err := h.permService.CheckFileAccess(user.ID, id, isServerOwner)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}
	}

	var f models.File
	var width, height sql.NullInt32
	var takenAt sql.NullTime
	err = h.db.QueryRow(`
		SELECT f.id, f.filename, f.file_type, f.size, f.created_at, f.updated_at,
		       pm.width, pm.height, pm.taken_at
		FROM files f
		LEFT JOIN photo_metadata pm ON f.id = pm.file_id
		WHERE f.id = ?`, id).Scan(
		&f.ID, &f.Filename, &f.FileType, &f.Size, &f.CreatedAt, &f.UpdatedAt,
		&width, &height, &takenAt)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	// Populate photo metadata fields if present
	if width.Valid {
		f.Width = int(width.Int32)
	}
	if height.Valid {
		f.Height = int(height.Int32)
	}
	if takenAt.Valid {
		f.TakenAt = &takenAt.Time
	}

	// Resolve absolute path
	absolutePath, err := h.folderService.ResolveAbsolutePath(f.ID)
	if err == nil {
		f.AbsolutePath = absolutePath
	}

	f.ThumbnailURL = "/api/files/" + strconv.FormatInt(f.ID, 10) + "/thumbnail"

	return c.JSON(f)
}

// GetFileThumbnail serves thumbnail for a file
func (h *Handler) GetFileThumbnail(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid file ID"})
	}

	// Check if user has access to this file
	isServerOwner := user.Role == "server_owner"
	if !isServerOwner {
		hasAccess, err := h.permService.CheckFileAccess(user.ID, id, isServerOwner)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}
	}

	// Get size parameter (small, medium, large)
	sizeType := c.Query("size", "small")

	// Resolve absolute path through folder service
	filePath, err := h.folderService.ResolveAbsolutePath(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	thumbPath, err := h.thumbService.GetThumbnail(filePath, id, sizeType)
	if err != nil {
		log.Printf("Error getting thumbnail: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate thumbnail"})
	}

	return c.SendFile(thumbPath)
}

// DownloadFile sends the original file
func (h *Handler) DownloadFile(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid file ID"})
	}

	// Check if user has access to this file
	isServerOwner := user.Role == "server_owner"
	if !isServerOwner {
		hasAccess, err := h.permService.CheckFileAccess(user.ID, id, isServerOwner)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}
	}

	var filename string
	err = h.db.QueryRow("SELECT filename FROM files WHERE id = ?", id).Scan(&filename)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	// Resolve absolute path through folder service
	filePath, err := h.folderService.ResolveAbsolutePath(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.SendFile(filePath)
}

// SearchFiles searches files by name or tags
func (h *Handler) SearchFiles(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	query := c.Query("q", "")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Search query is required"})
	}

	isServerOwner := user.Role == "server_owner"

	var sqlQuery string
	var args []interface{}

	if isServerOwner {
		// Server owner can search all files
		sqlQuery = `SELECT DISTINCT f.id, f.filename, f.file_type, f.size, f.created_at, f.updated_at,
		                   pm.width, pm.height, pm.taken_at
		            FROM files f
		            LEFT JOIN photo_metadata pm ON f.id = pm.file_id
		            LEFT JOIN file_tags ft ON f.id = ft.file_id
		            LEFT JOIN tags t ON ft.tag_id = t.id
		            WHERE f.filename LIKE ? OR t.name LIKE ?
		            ORDER BY pm.taken_at DESC
		            LIMIT 100`
		args = []interface{}{"%" + query + "%", "%" + query + "%"}
	} else {
		// Regular users can only search files they have permission for
		sqlQuery = `SELECT DISTINCT f.id, f.filename, f.file_type, f.size, f.created_at, f.updated_at,
		                   pm.width, pm.height, pm.taken_at
		            FROM files f
		            LEFT JOIN photo_metadata pm ON f.id = pm.file_id
		            LEFT JOIN file_tags ft ON f.id = ft.file_id
		            LEFT JOIN tags t ON ft.tag_id = t.id
		            JOIN file_folder_mappings ffm ON f.id = ffm.file_id
		            JOIN permission_group_folders pgf ON ffm.folder_id = pgf.folder_id
		            JOIN permission_group_permissions pgp ON pgf.permission_group_id = pgp.permission_group_id
		            WHERE (f.filename LIKE ? OR t.name LIKE ?)
		            AND pgp.user_id = ?
		            ORDER BY pm.taken_at DESC
		            LIMIT 100`
		args = []interface{}{"%" + query + "%", "%" + query + "%", user.ID}
	}

	rows, err := h.db.Query(sqlQuery, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	files := []models.File{}
	for rows.Next() {
		var f models.File
		var width, height sql.NullInt32
		var takenAt sql.NullTime
		if err := rows.Scan(&f.ID, &f.Filename, &f.FileType, &f.Size, &f.CreatedAt, &f.UpdatedAt,
			&width, &height, &takenAt); err != nil {
			continue
		}
		// Populate photo metadata fields if present
		if width.Valid {
			f.Width = int(width.Int32)
		}
		if height.Valid {
			f.Height = int(height.Int32)
		}
		if takenAt.Valid {
			f.TakenAt = &takenAt.Time
		}
		f.ThumbnailURL = "/api/files/" + strconv.FormatInt(f.ID, 10) + "/thumbnail"
		files = append(files, f)
	}

	// Validate files and filter out deleted ones
	files = h.validator.ValidateFiles(files)

	return c.JSON(fiber.Map{"files": files})
}

// GetMountPoints returns all mount points (deprecated, kept for compatibility)
func (h *Handler) GetMountPoints(c *fiber.Ctx) error {
	// Return empty list since mount points are deprecated
	return c.JSON(fiber.Map{"mount_points": []interface{}{}})
}

// TriggerScan triggers a file system scan
func (h *Handler) TriggerScan(c *fiber.Ctx) error {
	// Deprecated - use folder scan instead
	return c.Status(400).JSON(fiber.Map{
		"error": "Global scan is deprecated. Please use folder-specific scan via /api/folders/:id/scan",
	})
}

// CleanupDeletedFiles removes database records for files that no longer exist
func (h *Handler) CleanupDeletedFiles(c *fiber.Ctx) error {
	count, err := h.validator.CleanupAllInvalidFiles()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"message": "Cleanup completed",
		"deleted": count,
	})
}

// GetTags returns all tags
func (h *Handler) GetTags(c *fiber.Ctx) error {
	rows, err := h.db.Query("SELECT id, name, color, created_at FROM tags")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	tags := []models.Tag{}
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color, &t.CreatedAt); err != nil {
			continue
		}
		tags = append(tags, t)
	}

	return c.JSON(fiber.Map{"tags": tags})
}

// CreateTag creates a new tag
func (h *Handler) CreateTag(c *fiber.Ctx) error {
	var tag models.Tag
	if err := c.BodyParser(&tag); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.db.Exec("INSERT INTO tags (name, color) VALUES (?, ?)", tag.Name, tag.Color)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	id, _ := result.LastInsertId()
	tag.ID = id

	return c.Status(201).JSON(tag)
}

// GetAlbums returns all albums
func (h *Handler) GetAlbums(c *fiber.Ctx) error {
	rows, err := h.db.Query("SELECT id, name, description, cover_file_id, created_at, updated_at FROM albums")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	albums := []models.Album{}
	for rows.Next() {
		var a models.Album
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.CoverFileID, &a.CreatedAt, &a.UpdatedAt); err != nil {
			continue
		}
		albums = append(albums, a)
	}

	return c.JSON(fiber.Map{"albums": albums})
}

// CreateAlbum creates a new album
func (h *Handler) CreateAlbum(c *fiber.Ctx) error {
	var album models.Album
	if err := c.BodyParser(&album); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.db.Exec("INSERT INTO albums (name, description) VALUES (?, ?)", album.Name, album.Description)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	id, _ := result.LastInsertId()
	album.ID = id

	return c.Status(201).JSON(album)
}
