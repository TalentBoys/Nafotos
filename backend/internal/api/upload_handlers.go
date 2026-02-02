package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"

	"awesome-sharing/internal/middleware"
	"awesome-sharing/internal/services"
)

type UploadHandler struct {
	folderService  *services.FolderService
	scannerService *services.FileScanner
}

func NewUploadHandler(folderService *services.FolderService, scannerService *services.FileScanner) *UploadHandler {
	return &UploadHandler{
		folderService:  folderService,
		scannerService: scannerService,
	}
}

// UploadFiles handles file uploads
// POST /api/upload
func (h *UploadHandler) UploadFiles(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Get target path from form
	targetPath := c.FormValue("target_path")
	if targetPath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Target path is required",
		})
	}

	// Validate target path is absolute
	if !filepath.IsAbs(targetPath) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Target path must be absolute",
		})
	}

	// Clean the path
	targetPath = filepath.Clean(targetPath)

	// Check if target directory exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Target directory does not exist",
		})
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse form data",
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No files provided",
		})
	}

	// Supported image extensions
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".heic": true,
		".heif": true,
		".tif":  true,
		".tiff": true,
	}

	var uploadedFiles []string
	var failedFiles []map[string]string

	for _, file := range files {
		// Check file extension
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !supportedExts[ext] {
			failedFiles = append(failedFiles, map[string]string{
				"filename": file.Filename,
				"error":    "Unsupported file format",
			})
			continue
		}

		// Generate destination path
		destPath := filepath.Join(targetPath, file.Filename)

		// Check if file already exists
		if _, err := os.Stat(destPath); err == nil {
			failedFiles = append(failedFiles, map[string]string{
				"filename": file.Filename,
				"error":    "File already exists",
			})
			continue
		}

		// Open uploaded file
		src, err := file.Open()
		if err != nil {
			failedFiles = append(failedFiles, map[string]string{
				"filename": file.Filename,
				"error":    fmt.Sprintf("Failed to open file: %v", err),
			})
			continue
		}

		// Create destination file
		dst, err := os.Create(destPath)
		if err != nil {
			src.Close()
			failedFiles = append(failedFiles, map[string]string{
				"filename": file.Filename,
				"error":    fmt.Sprintf("Failed to create file: %v", err),
			})
			continue
		}

		// Copy file contents
		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			os.Remove(destPath) // Clean up partial file
			failedFiles = append(failedFiles, map[string]string{
				"filename": file.Filename,
				"error":    fmt.Sprintf("Failed to save file: %v", err),
			})
			continue
		}

		src.Close()
		dst.Close()

		uploadedFiles = append(uploadedFiles, file.Filename)
	}

	// Trigger scan of the target directory
	// Find folder ID for the target path
	go func() {
		// This will scan all folders, but it's a background task
		// In a real implementation, you might want to scan only the specific folder
		h.scannerService.ScanAllFolders()
	}()

	return c.JSON(fiber.Map{
		"message":        "Upload completed",
		"uploaded":       uploadedFiles,
		"uploaded_count": len(uploadedFiles),
		"failed":         failedFiles,
		"failed_count":   len(failedFiles),
		"total":          len(files),
	})
}

// CreateDirectory creates a new directory in the file system
// POST /api/upload/create-directory
func (h *UploadHandler) CreateDirectory(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req struct {
		ParentPath    string `json:"parent_path"`
		DirectoryName string `json:"directory_name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.ParentPath == "" || req.DirectoryName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Parent path and directory name are required",
		})
	}

	// Validate parent path is absolute
	if !filepath.IsAbs(req.ParentPath) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Parent path must be absolute",
		})
	}

	// Clean paths
	parentPath := filepath.Clean(req.ParentPath)
	dirName := filepath.Clean(req.DirectoryName)

	// Prevent directory traversal
	if strings.Contains(dirName, "..") || strings.Contains(dirName, "/") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid directory name",
		})
	}

	// Create full path
	fullPath := filepath.Join(parentPath, dirName)

	// Check if directory already exists
	if _, err := os.Stat(fullPath); err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Directory already exists",
		})
	}

	// Create directory
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create directory: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Directory created successfully",
		"path":    fullPath,
	})
}

// BrowseUploadTarget browses directories and shows files for upload target selection
// POST /api/upload/browse
func (h *UploadHandler) BrowseUploadTarget(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req struct {
		Path     string `json:"path"`
		FolderID *int64 `json:"folder_id"` // Optional: browse within a specific folder
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var browsePath string

	// If folder ID is provided, get folder path
	if req.FolderID != nil {
		folder, err := h.folderService.GetFolder(*req.FolderID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Folder not found",
			})
		}
		browsePath = folder.AbsolutePath
	} else if req.Path != "" {
		browsePath = req.Path
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Path or folder_id is required",
		})
	}

	// Validate path
	if !filepath.IsAbs(browsePath) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Path must be absolute",
		})
	}

	browsePath = filepath.Clean(browsePath)

	// Check if directory exists
	info, err := os.Stat(browsePath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Directory not found",
		})
	}
	if !info.IsDir() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Path is not a directory",
		})
	}

	// Read directory contents
	entries, err := os.ReadDir(browsePath)
	if err != nil {
		// Permission denied or other error
		return c.JSON(fiber.Map{
			"path":        browsePath,
			"directories": []services.DirectoryInfo{},
		})
	}

	// Filter and collect directories
	var directories []services.DirectoryInfo
	for _, entry := range entries {
		if entry.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(browsePath, entry.Name())
			directories = append(directories, services.DirectoryInfo{
				Name:        entry.Name(),
				Path:        fullPath,
				IsDirectory: true,
			})
		}
	}

	return c.JSON(fiber.Map{
		"path":        browsePath,
		"directories": directories,
	})
}
