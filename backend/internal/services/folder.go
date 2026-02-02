package services

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"awesome-sharing/internal/models"
)

var (
	ErrFolderNotFound       = errors.New("folder not found")
	ErrFolderPathConflict   = errors.New("folder path conflicts with existing folder")
	ErrFolderPathNotAbsolute = errors.New("folder path must be absolute")
)

type FolderService struct {
	db *sql.DB
}

func NewFolderService(db *sql.DB) *FolderService {
	return &FolderService{db: db}
}

// CreateFolder creates a new folder with path validation
func (s *FolderService) CreateFolder(name, absolutePath string, createdBy int64) (*models.Folder, error) {
	// Validate path
	if !filepath.IsAbs(absolutePath) {
		return nil, ErrFolderPathNotAbsolute
	}

	// Clean the path
	absolutePath = filepath.Clean(absolutePath)

	// Check for conflicts
	if err := s.ValidateFolderPath(absolutePath); err != nil {
		return nil, err
	}

	result, err := s.db.Exec(`
		INSERT INTO folders (name, absolute_path, enabled, created_by)
		VALUES (?, ?, 1, ?)
	`, name, absolutePath, createdBy)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetFolder(id)
}

// ValidateFolderPath checks if a path conflicts with existing folders
// Returns error if path is parent or child of any existing folder
func (s *FolderService) ValidateFolderPath(path string) error {
	// Clean the path
	path = filepath.Clean(path)

	// Get all existing folder paths
	rows, err := s.db.Query("SELECT absolute_path FROM folders")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var existingPath string
		if err := rows.Scan(&existingPath); err != nil {
			continue
		}

		// Check if new path is parent of existing path
		if strings.HasPrefix(existingPath, path+string(filepath.Separator)) {
			return ErrFolderPathConflict
		}

		// Check if new path is child of existing path
		if strings.HasPrefix(path, existingPath+string(filepath.Separator)) {
			return ErrFolderPathConflict
		}

		// Check if paths are identical
		if path == existingPath {
			return ErrFolderPathConflict
		}
	}

	return nil
}

// GetFolder retrieves a folder by ID
func (s *FolderService) GetFolder(id int64) (*models.Folder, error) {
	var folder models.Folder
	err := s.db.QueryRow(`
		SELECT id, name, absolute_path, enabled, created_by, created_at, updated_at
		FROM folders WHERE id = ?
	`, id).Scan(&folder.ID, &folder.Name, &folder.AbsolutePath, &folder.Enabled,
		&folder.CreatedBy, &folder.CreatedAt, &folder.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrFolderNotFound
	}
	if err != nil {
		return nil, err
	}

	return &folder, nil
}

// ListFolders retrieves folders accessible to a user
func (s *FolderService) ListFolders(userID int64, isAdmin bool) ([]models.Folder, error) {
	var rows *sql.Rows
	var err error

	if isAdmin {
		// Admin can see all folders
		rows, err = s.db.Query(`
			SELECT id, name, absolute_path, enabled, created_by, created_at, updated_at
			FROM folders
			ORDER BY created_at DESC
		`)
	} else {
		// Regular users can only see folders they have permission for through permission groups
		rows, err = s.db.Query(`
			SELECT DISTINCT f.id, f.name, f.absolute_path, f.enabled, f.created_by, f.created_at, f.updated_at
			FROM folders f
			INNER JOIN permission_group_folders pgf ON f.id = pgf.folder_id
			INNER JOIN permission_group_permissions pgp ON pgf.permission_group_id = pgp.permission_group_id
			WHERE pgp.user_id = ?
			ORDER BY f.created_at DESC
		`, userID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []models.Folder
	for rows.Next() {
		var folder models.Folder
		if err := rows.Scan(&folder.ID, &folder.Name, &folder.AbsolutePath, &folder.Enabled,
			&folder.CreatedBy, &folder.CreatedAt, &folder.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}

	return folders, nil
}

// UpdateFolder updates folder information
func (s *FolderService) UpdateFolder(id int64, name, absolutePath string) error {
	// Validate new path if it's being changed
	if absolutePath != "" {
		if !filepath.IsAbs(absolutePath) {
			return ErrFolderPathNotAbsolute
		}

		absolutePath = filepath.Clean(absolutePath)

		// Get current folder
		currentFolder, err := s.GetFolder(id)
		if err != nil {
			return err
		}

		// Only validate if path is actually changing
		if absolutePath != currentFolder.AbsolutePath {
			// Temporarily exclude this folder from validation
			// We'll validate against all other folders
			rows, err := s.db.Query("SELECT absolute_path FROM folders WHERE id != ?", id)
			if err != nil {
				return err
			}
			defer rows.Close()

			for rows.Next() {
				var existingPath string
				if err := rows.Scan(&existingPath); err != nil {
					continue
				}

				if strings.HasPrefix(existingPath, absolutePath+string(filepath.Separator)) ||
					strings.HasPrefix(absolutePath, existingPath+string(filepath.Separator)) ||
					absolutePath == existingPath {
					return ErrFolderPathConflict
				}
			}
		}

		_, err = s.db.Exec(`
			UPDATE folders
			SET name = ?, absolute_path = ?, updated_at = ?
			WHERE id = ?
		`, name, absolutePath, time.Now(), id)
		return err
	}

	_, err := s.db.Exec(`
		UPDATE folders
		SET name = ?, updated_at = ?
		WHERE id = ?
	`, name, time.Now(), id)
	return err
}

// DeleteFolder deletes a folder
func (s *FolderService) DeleteFolder(id int64) error {
	_, err := s.db.Exec("DELETE FROM folders WHERE id = ?", id)
	return err
}

// ToggleFolder enables/disables a folder
func (s *FolderService) ToggleFolder(id int64, enabled bool) error {
	_, err := s.db.Exec("UPDATE folders SET enabled = ?, updated_at = ? WHERE id = ?",
		enabled, time.Now(), id)
	return err
}

// GetFolderForFile retrieves the folder(s) containing a file
func (s *FolderService) GetFolderForFile(fileID int64) ([]models.Folder, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.name, f.absolute_path, f.enabled, f.created_by, f.created_at, f.updated_at
		FROM folders f
		INNER JOIN file_folder_mappings ffm ON f.id = ffm.folder_id
		WHERE ffm.file_id = ?
	`, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []models.Folder
	for rows.Next() {
		var folder models.Folder
		if err := rows.Scan(&folder.ID, &folder.Name, &folder.AbsolutePath, &folder.Enabled,
			&folder.CreatedBy, &folder.CreatedAt, &folder.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}

	return folders, nil
}

// ResolveAbsolutePath calculates the absolute path for a file
// Returns the first mapping found (a file may be in multiple folders)
func (s *FolderService) ResolveAbsolutePath(fileID int64) (string, error) {
	var folderPath, relativePath string
	err := s.db.QueryRow(`
		SELECT f.absolute_path, ffm.relative_path
		FROM file_folder_mappings ffm
		INNER JOIN folders f ON ffm.folder_id = f.id
		WHERE ffm.file_id = ?
		LIMIT 1
	`, fileID).Scan(&folderPath, &relativePath)

	if err == sql.ErrNoRows {
		return "", errors.New("file not found in any folder")
	}
	if err != nil {
		return "", err
	}

	return filepath.Join(folderPath, relativePath), nil
}

// AddFileMapping adds a file-folder mapping
func (s *FolderService) AddFileMapping(fileID, folderID int64, relativePath string) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO file_folder_mappings (file_id, folder_id, relative_path)
		VALUES (?, ?, ?)
	`, fileID, folderID, relativePath)
	return err
}

// RemoveFileMapping removes a specific file-folder mapping
func (s *FolderService) RemoveFileMapping(fileID, folderID int64) error {
	_, err := s.db.Exec(`
		DELETE FROM file_folder_mappings
		WHERE file_id = ? AND folder_id = ?
	`, fileID, folderID)
	return err
}

// ListFilesInFolder retrieves all files in a folder
func (s *FolderService) ListFilesInFolder(folderID int64, limit, offset int) ([]models.File, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.filename, f.file_type, f.size, f.width, f.height, f.taken_at,
		       f.created_at, f.updated_at, f.is_thumbnail, f.parent_file_id
		FROM files f
		INNER JOIN file_folder_mappings ffm ON f.id = ffm.file_id
		WHERE ffm.folder_id = ? AND (f.is_thumbnail IS NULL OR f.is_thumbnail = 0)
		ORDER BY f.taken_at DESC
		LIMIT ? OFFSET ?
	`, folderID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(&file.ID, &file.Filename, &file.FileType,
			&file.Size, &file.Width, &file.Height, &file.TakenAt,
			&file.CreatedAt, &file.UpdatedAt, &file.IsThumbnail,
			&file.ParentFileID); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

// CountFilesInFolder counts files in a folder
func (s *FolderService) CountFilesInFolder(folderID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM files f
		INNER JOIN file_folder_mappings ffm ON f.id = ffm.file_id
		WHERE ffm.folder_id = ? AND (f.is_thumbnail IS NULL OR f.is_thumbnail = 0)
	`, folderID).Scan(&count)
	return count, err
}

// DirectoryInfo represents a directory in the file system
type DirectoryInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IsDirectory bool   `json:"is_directory"`
}

// BrowseDirectory lists subdirectories in a given path
func (s *FolderService) BrowseDirectory(path string) ([]DirectoryInfo, error) {
	// Validate path
	if !filepath.IsAbs(path) {
		return nil, errors.New("path must be absolute")
	}

	// Clean the path
	path = filepath.Clean(path)

	// Check if directory exists and is accessible
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("path is not a directory")
	}

	// Read directory contents
	entries, err := os.ReadDir(path)
	if err != nil {
		// Return empty array instead of error for permission denied
		// This prevents crashes and allows graceful handling
		return []DirectoryInfo{}, nil
	}

	// Filter and collect directories only
	var directories []DirectoryInfo
	for _, entry := range entries {
		if entry.IsDir() {
			// Skip hidden directories (starting with .)
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(path, entry.Name())
			directories = append(directories, DirectoryInfo{
				Name:        entry.Name(),
				Path:        fullPath,
				IsDirectory: true,
			})
		}
	}

	// Sort directories by name
	sort.Slice(directories, func(i, j int) bool {
		return directories[i].Name < directories[j].Name
	})

	// Return empty array if no directories found (instead of nil)
	if directories == nil {
		return []DirectoryInfo{}, nil
	}

	return directories, nil
}
