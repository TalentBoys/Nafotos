package services

import (
	"awesome-sharing/internal/database"
	"awesome-sharing/pkg/exif"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileScanner struct {
	db            *database.DB
	folderService *FolderService
	thumbsDir     string
}

func NewFileScanner(db *database.DB, folderService *FolderService, thumbsDir string) *FileScanner {
	return &FileScanner{
		db:            db,
		folderService: folderService,
		thumbsDir:     thumbsDir,
	}
}

// ScanFolder scans a specific folder
func (fs *FileScanner) ScanFolder(folderID int64) error {
	// Get folder information
	folder, err := fs.folderService.GetFolder(folderID)
	if err != nil {
		return err
	}

	log.Printf("Starting scan of folder: %s (%s)", folder.Name, folder.AbsolutePath)

	if err := fs.scanDirectory(folder.ID, folder.AbsolutePath, folder.AbsolutePath); err != nil {
		return err
	}

	log.Printf("Completed scan of folder: %s", folder.Name)
	return nil
}

// ScanAllFolders scans all enabled folders
func (fs *FileScanner) ScanAllFolders() {
	log.Println("Starting scan of all folders...")

	// Get all enabled folders (admin view)
	rows, err := fs.db.Query("SELECT id, name, absolute_path FROM folders WHERE enabled = 1")
	if err != nil {
		log.Printf("Error querying folders: %v", err)
		return
	}
	defer rows.Close()

	foldersScanned := 0
	for rows.Next() {
		var folderID int64
		var name, absolutePath string
		if err := rows.Scan(&folderID, &name, &absolutePath); err != nil {
			log.Printf("Error reading folder: %v", err)
			continue
		}

		log.Printf("Scanning folder: %s (%s)", name, absolutePath)
		if err := fs.scanDirectory(folderID, absolutePath, absolutePath); err != nil {
			log.Printf("Error scanning folder %s: %v", name, err)
		}
		foldersScanned++
	}

	log.Printf("Scan completed. %d folders scanned.", foldersScanned)
}

// scanDirectory recursively scans a directory
func (fs *FileScanner) scanDirectory(folderID int64, rootPath, currentPath string) error {
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(currentPath, entry.Name())

		// Skip hidden files and directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip thumbnails directory
		if fs.thumbsDir != "" {
			absThumbsDir, _ := filepath.Abs(fs.thumbsDir)
			absFullPath, _ := filepath.Abs(fullPath)
			if strings.HasPrefix(absFullPath, absThumbsDir) {
				continue
			}
		}

		if entry.IsDir() {
			// Recursively scan subdirectories
			if err := fs.scanDirectory(folderID, rootPath, fullPath); err != nil {
				log.Printf("Error scanning directory %s: %v", fullPath, err)
			}
			continue
		}

		// Process file
		if fs.isMediaFile(entry.Name()) {
			if err := fs.indexFile(folderID, rootPath, fullPath); err != nil {
				log.Printf("Error indexing file %s: %v", fullPath, err)
			}
		}
	}

	return nil
}

// isMediaFile checks if the file is an image or video
func (fs *FileScanner) isMediaFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".heic", ".heif", ".tif", ".tiff"}
	videoExts := []string{".mp4", ".mov", ".avi", ".mkv", ".webm", ".m4v"}

	for _, e := range imageExts {
		if ext == e {
			return true
		}
	}
	for _, e := range videoExts {
		if ext == e {
			return true
		}
	}
	return false
}

// indexFile adds or updates a file in the database
func (fs *FileScanner) indexFile(folderID int64, rootPath, filePath string) error {
	// Calculate relative path
	relativePath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		return err
	}

	// Check if file already exists in this folder
	var existingID int64
	err = fs.db.QueryRow(`
		SELECT f.id FROM files f
		INNER JOIN file_folder_mappings ffm ON f.id = ffm.file_id
		WHERE ffm.folder_id = ? AND ffm.relative_path = ?
	`, folderID, relativePath).Scan(&existingID)

	if err == nil {
		// File already indexed - check if dimensions are missing
		if err := fs.fixMissingDimensions(existingID, filePath); err != nil {
			log.Printf("Warning: Failed to fix missing dimensions for file %d: %v", existingID, err)
		}
		return nil
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	fileType := "image"
	ext := strings.ToLower(filepath.Ext(filePath))
	if strings.Contains(".mp4.mov.avi.mkv.webm.m4v", ext) {
		fileType = "video"
	}

	// Insert file into database WITHOUT photo-specific fields
	result, err := fs.db.Exec(`
		INSERT INTO files (filename, file_type, size, is_thumbnail, parent_file_id)
		VALUES (?, ?, ?, 0, NULL)`,
		filepath.Base(filePath), fileType, info.Size())

	if err != nil {
		return err
	}

	fileID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Extract and save EXIF data for images
	if fileType == "image" {
		if err := fs.savePhotoMetadata(fileID, filePath, info.ModTime()); err != nil {
			log.Printf("Warning: Failed to save photo metadata for file %d: %v", fileID, err)
			// Don't fail indexing if EXIF extraction fails
		}
	}

	// Create file-folder mapping
	if err := fs.folderService.AddFileMapping(fileID, folderID, relativePath); err != nil {
		log.Printf("Warning: Failed to create mapping for file %d to folder %d: %v", fileID, folderID, err)
		return err
	}

	log.Printf("Indexed: %s (folder ID: %d)", filePath, folderID)
	return nil
}

// fixMissingDimensions checks if a file has missing width/height and attempts to fix it
func (fs *FileScanner) fixMissingDimensions(fileID int64, filePath string) error {
	// Check if this is an image file
	ext := strings.ToLower(filepath.Ext(filePath))
	if strings.Contains(".mp4.mov.avi.mkv.webm.m4v", ext) {
		// Skip video files for now
		return nil
	}

	// Check current dimensions in photo_metadata
	var width, height int
	err := fs.db.QueryRow(`
		SELECT width, height FROM photo_metadata WHERE file_id = ?
	`, fileID).Scan(&width, &height)

	if err != nil {
		// No metadata record exists, try to create one
		info, err := os.Stat(filePath)
		if err != nil {
			return err
		}
		return fs.savePhotoMetadata(fileID, filePath, info.ModTime())
	}

	// If dimensions are valid, no need to fix
	if width > 0 && height > 0 {
		return nil
	}

	// Dimensions are missing, try to get them
	log.Printf("Fixing missing dimensions for file %d: %s", fileID, filePath)

	newWidth, newHeight := 0, 0

	// Try EXIF first
	exifData, err := exif.ExtractEXIF(filePath)
	if err == nil && exifData.Width > 0 && exifData.Height > 0 {
		newWidth, newHeight = exifData.Width, exifData.Height
	}

	// Fallback to GetDimensions
	if newWidth == 0 || newHeight == 0 {
		if w, h, err := GetDimensions(filePath); err == nil {
			newWidth, newHeight = w, h
		} else {
			log.Printf("Failed to get dimensions for %s: %v", filePath, err)
			return err
		}
	}

	// Update the database
	if newWidth > 0 && newHeight > 0 {
		_, err = fs.db.Exec(`
			UPDATE photo_metadata SET width = ?, height = ? WHERE file_id = ?
		`, newWidth, newHeight, fileID)
		if err != nil {
			return err
		}
		log.Printf("Fixed dimensions for file %d: %dx%d", fileID, newWidth, newHeight)
	}

	return nil
}

// savePhotoMetadata extracts EXIF data and saves it to photo_metadata table
func (fs *FileScanner) savePhotoMetadata(fileID int64, filePath string, modTime time.Time) error {
	// Default values
	takenAt := modTime
	width, height := 0, 0

	// Try to extract EXIF
	exifData, err := exif.ExtractEXIF(filePath)
	if err == nil {
		if !exifData.DateTime.IsZero() {
			takenAt = exifData.DateTime
		}
		width = exifData.Width
		height = exifData.Height

		// If EXIF dimensions are missing/zero, use GetDimensions as fallback
		if width == 0 || height == 0 {
			log.Printf("EXIF dimensions missing for %s, using GetDimensions fallback", filepath.Base(filePath))
			if w, h, err := GetDimensions(filePath); err == nil {
				width, height = w, h
				log.Printf("GetDimensions success: %dx%d for %s", width, height, filepath.Base(filePath))
			} else {
				log.Printf("GetDimensions failed for %s: %v", filepath.Base(filePath), err)
			}
		} else {
			log.Printf("EXIF dimensions found: %dx%d for %s", width, height, filepath.Base(filePath))
		}

		// Insert with all EXIF fields
		_, err = fs.db.Exec(`
			INSERT INTO photo_metadata (
				file_id, width, height, taken_at,
				make, model, latitude, longitude, altitude,
				iso, aperture, shutter_speed, focal_length, orientation
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			fileID, width, height, takenAt,
			exifData.Make, exifData.Model,
			exifData.Latitude, exifData.Longitude, exifData.Altitude,
			exifData.ISO, exifData.Aperture, exifData.ShutterSpeed,
			exifData.FocalLength, exifData.Orientation)

		return err
	}

	// If EXIF extraction failed, use GetDimensions as fallback
	log.Printf("EXIF extraction failed for %s: %v, using GetDimensions fallback", filepath.Base(filePath), err)
	if w, h, err := GetDimensions(filePath); err == nil {
		width, height = w, h
		log.Printf("GetDimensions success: %dx%d for %s", width, height, filepath.Base(filePath))
	} else {
		log.Printf("GetDimensions failed for %s: %v", filepath.Base(filePath), err)
	}

	// Insert minimal metadata
	_, err = fs.db.Exec(`
		INSERT INTO photo_metadata (file_id, width, height, taken_at)
		VALUES (?, ?, ?, ?)`,
		fileID, width, height, takenAt)

	return err
}

// ScanPeriodically runs scan at regular intervals
func (fs *FileScanner) ScanPeriodically(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial scan
	fs.ScanAllFolders()

	for range ticker.C {
		fs.ScanAllFolders()
	}
}
