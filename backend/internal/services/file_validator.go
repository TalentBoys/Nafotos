package services

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"awesome-sharing/internal/models"
)

// FileValidatorService handles validation and cleanup of deleted files
type FileValidatorService struct {
	db            *sql.DB
	folderService *FolderService
	mu            sync.Mutex
	cleanupCache  map[int64]bool // Cache to avoid repeated cleanup attempts
}

func NewFileValidatorService(db *sql.DB, folderService *FolderService) *FileValidatorService {
	return &FileValidatorService{
		db:            db,
		folderService: folderService,
		cleanupCache:  make(map[int64]bool),
	}
}

// ValidateFiles checks if files exist and returns only valid ones
// Also marks invalid files for cleanup
func (s *FileValidatorService) ValidateFiles(files []models.File) []models.File {
	validFiles := make([]models.File, 0, len(files))
	invalidIDs := make([]int64, 0)

	for _, file := range files {
		// Resolve absolute path from folder mapping
		absolutePath, err := s.folderService.ResolveAbsolutePath(file.ID)
		if err != nil || !s.fileExists(absolutePath) {
			invalidIDs = append(invalidIDs, file.ID)
		} else {
			// Set the absolute path for display
			file.AbsolutePath = absolutePath
			validFiles = append(validFiles, file)
		}
	}

	// Cleanup invalid files in background
	if len(invalidIDs) > 0 {
		go s.cleanupFiles(invalidIDs)
	}

	return validFiles
}

// fileExists checks if a file exists on the filesystem with timeout protection
func (s *FileValidatorService) fileExists(path string) bool {
	// Use a channel to implement timeout for file check
	result := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	startTime := time.Now()
	go func() {
		_, err := os.Stat(path)
		result <- (err == nil)
	}()

	select {
	case exists := <-result:
		elapsed := time.Since(startTime)
		// Log slow file checks (over 1 second)
		if elapsed > time.Second {
			log.Printf("Warning: slow file check (%v): %s", elapsed, path)
		}
		return exists
	case <-ctx.Done():
		// Timeout - assume file doesn't exist or path is inaccessible
		log.Printf("ERROR: timeout (5s) checking file existence: %s", path)
		return false
	}
}

// cleanupFiles removes file records from database
func (s *FileValidatorService) cleanupFiles(fileIDs []int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanedCount := 0
	totalToClean := len(fileIDs)

	for i, id := range fileIDs {
		// Check cache to avoid repeated cleanup
		if s.cleanupCache[id] {
			continue
		}

		// Delete file record (this will cascade delete thumbnails and mappings via foreign key)
		_, err := s.db.Exec("DELETE FROM files WHERE id = ?", id)
		if err != nil {
			log.Printf("Error deleting file record %d: %v", id, err)
			continue
		}

		// Delete associated thumbnails from filesystem
		s.deleteFileThumbnails(id)

		// Mark as cleaned up
		s.cleanupCache[id] = true
		cleanedCount++

		// Log progress for large cleanups
		if totalToClean > 10 && (i+1)%10 == 0 {
			log.Printf("Cleanup progress: %d/%d files removed", i+1, totalToClean)
		}
	}

	if cleanedCount > 0 {
		log.Printf("Successfully cleaned up %d file records", cleanedCount)
	}
}

// deleteFileThumbnails deletes thumbnail files from filesystem
func (s *FileValidatorService) deleteFileThumbnails(fileID int64) {
	// Query file_thumbnails table to get thumbnail paths
	rows, err := s.db.Query("SELECT path FROM file_thumbnails WHERE file_id = ?", fileID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			continue
		}
		// Delete thumbnail file
		if err := os.Remove(path); err != nil {
			log.Printf("Error deleting thumbnail file %s: %v", path, err)
		}
	}
}

// CleanupAllInvalidFiles scans entire database and removes invalid file records
func (s *FileValidatorService) CleanupAllInvalidFiles() (int, error) {
	log.Println("Starting full file validation and cleanup...")

	// First, get count of files to validate
	var fileCount int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM files f
		JOIN file_folder_mappings ffm ON f.id = ffm.file_id
		JOIN folders fo ON ffm.folder_id = fo.id
		WHERE f.is_thumbnail IS NULL OR f.is_thumbnail = 0
	`).Scan(&fileCount)
	if err != nil {
		log.Printf("Error getting file count: %v", err)
	} else {
		log.Printf("Found %d files to validate", fileCount)
	}

	// Get all file-folder mappings
	log.Println("Querying database for all file-folder mappings...")
	rows, err := s.db.Query(`
		SELECT f.id, fo.absolute_path, ffm.relative_path
		FROM files f
		JOIN file_folder_mappings ffm ON f.id = ffm.file_id
		JOIN folders fo ON ffm.folder_id = fo.id
		WHERE f.is_thumbnail IS NULL OR f.is_thumbnail = 0
	`)
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return 0, err
	}
	defer rows.Close()
	log.Println("Database query completed, starting validation...")

	invalidIDs := make([]int64, 0)
	total := 0
	checked := 0
	progressInterval := 10 // Log progress every 10 files for better debugging

	for rows.Next() {
		var id int64
		var folderPath, relativePath string
		if err := rows.Scan(&id, &folderPath, &relativePath); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		total++
		checked++

		// Construct absolute path
		absolutePath := filepath.Join(folderPath, relativePath)

		// Log first few files for debugging
		if checked <= 5 {
			log.Printf("Checking file %d: %s", id, absolutePath)
		}

		// Log progress periodically
		if checked%progressInterval == 0 {
			log.Printf("Validation progress: checked %d files, found %d invalid so far...", checked, len(invalidIDs))
		}

		exists := s.fileExists(absolutePath)
		if !exists {
			invalidIDs = append(invalidIDs, id)
			if len(invalidIDs) <= 5 {
				log.Printf("File %d marked as invalid: %s", id, absolutePath)
			}
		}
	}

	log.Printf("Validation scan complete: total %d files checked", total)

	// Cleanup invalid files
	if len(invalidIDs) > 0 {
		log.Printf("Cleaning up %d invalid files...", len(invalidIDs))
		s.cleanupFiles(invalidIDs)
	}

	log.Printf("File validation complete: checked %d files, cleaned up %d invalid files", total, len(invalidIDs))
	return len(invalidIDs), nil
}

// CheckFileExists checks if a specific file exists
func (s *FileValidatorService) CheckFileExists(fileID int64) (bool, error) {
	absolutePath, err := s.folderService.ResolveAbsolutePath(fileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return s.fileExists(absolutePath), nil
}
