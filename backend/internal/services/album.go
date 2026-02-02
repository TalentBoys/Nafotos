package services

import (
	"database/sql"
	"errors"
	"time"

	"awesome-sharing/internal/models"
)

var (
	ErrAlbumNotFound = errors.New("album not found")
)

type AlbumService struct {
	db *sql.DB
}

func NewAlbumService(db *sql.DB) *AlbumService {
	return &AlbumService{db: db}
}

// CreateAlbum creates a new album
func (s *AlbumService) CreateAlbum(name, description string, ownerID int64) (*models.Album, error) {
	result, err := s.db.Exec(`
		INSERT INTO albums_v2 (name, description, owner_id)
		VALUES (?, ?, ?)
	`, name, description, ownerID)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetAlbum(id)
}

// GetAlbum retrieves an album by ID
func (s *AlbumService) GetAlbum(id int64) (*models.Album, error) {
	var album models.Album
	err := s.db.QueryRow(`
		SELECT id, name, description, owner_id, cover_file_id, created_at, updated_at
		FROM albums_v2 WHERE id = ?
	`, id).Scan(&album.ID, &album.Name, &album.Description, &album.OwnerID,
		&album.CoverFileID, &album.CreatedAt, &album.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrAlbumNotFound
	}
	if err != nil {
		return nil, err
	}

	return &album, nil
}

// ListAlbums retrieves all albums for a user
func (s *AlbumService) ListAlbums(ownerID int64) ([]models.Album, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, owner_id, cover_file_id, created_at, updated_at
		FROM albums_v2 WHERE owner_id = ?
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []models.Album
	for rows.Next() {
		var album models.Album
		if err := rows.Scan(&album.ID, &album.Name, &album.Description, &album.OwnerID,
			&album.CoverFileID, &album.CreatedAt, &album.UpdatedAt); err != nil {
			return nil, err
		}
		albums = append(albums, album)
	}

	return albums, nil
}

// UpdateAlbum updates album information
func (s *AlbumService) UpdateAlbum(id int64, name, description string, coverFileID *int64) error {
	_, err := s.db.Exec(`
		UPDATE albums_v2
		SET name = ?, description = ?, cover_file_id = ?, updated_at = ?
		WHERE id = ?
	`, name, description, coverFileID, time.Now(), id)
	return err
}

// DeleteAlbum deletes an album and its items
func (s *AlbumService) DeleteAlbum(id int64) error {
	_, err := s.db.Exec("DELETE FROM albums_v2 WHERE id = ?", id)
	return err
}

// ListItemsWithFiles retrieves album files directly from file_folder_mappings
// based on album folder configurations (dynamic query, no album_items table)
func (s *AlbumService) ListItemsWithFiles(albumID int64, sortOrder string) ([]models.File, error) {
	// Get all folder configurations for this album
	folderConfigs, err := s.ListAlbumFolders(albumID)
	if err != nil {
		return nil, err
	}

	if len(folderConfigs) == 0 {
		return []models.File{}, nil
	}

	// Build dynamic query to get all matching files
	// Use UNION to combine results from all folder configurations
	// LEFT JOIN photo_metadata to get photo-specific fields (width, height, taken_at)
	var queryParts []string
	var args []interface{}

	for _, config := range folderConfigs {
		if config.PathPrefix == "" {
			// Empty prefix means entire folder
			queryParts = append(queryParts, `
				SELECT DISTINCT f.id, f.filename, f.file_type, f.size,
					COALESCE(pm.width, 0) as width, COALESCE(pm.height, 0) as height,
					pm.taken_at, f.created_at, f.updated_at, f.is_thumbnail, f.parent_file_id
				FROM files f
				INNER JOIN file_folder_mappings ffm ON f.id = ffm.file_id
				LEFT JOIN photo_metadata pm ON f.id = pm.file_id
				WHERE ffm.folder_id = ?
			`)
			args = append(args, config.FolderID)
		} else {
			// Match files with path prefix
			queryParts = append(queryParts, `
				SELECT DISTINCT f.id, f.filename, f.file_type, f.size,
					COALESCE(pm.width, 0) as width, COALESCE(pm.height, 0) as height,
					pm.taken_at, f.created_at, f.updated_at, f.is_thumbnail, f.parent_file_id
				FROM files f
				INNER JOIN file_folder_mappings ffm ON f.id = ffm.file_id
				LEFT JOIN photo_metadata pm ON f.id = pm.file_id
				WHERE ffm.folder_id = ? AND ffm.relative_path LIKE ?
			`)
			args = append(args, config.FolderID, config.PathPrefix+"%")
		}
	}

	// Combine all queries with UNION
	query := "SELECT * FROM (" + queryParts[0]
	for i := 1; i < len(queryParts); i++ {
		query += " UNION " + queryParts[i]
	}
	query += ")"

	// Add ORDER BY based on sortOrder parameter
	// Default to taken_at DESC if not specified
	if sortOrder == "" {
		sortOrder = "taken_at DESC"
	}
	query += " ORDER BY " + sortOrder

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.Filename, &f.FileType, &f.Size, &f.Width, &f.Height,
			&f.TakenAt, &f.CreatedAt, &f.UpdatedAt, &f.IsThumbnail, &f.ParentFileID); err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	return files, nil
}

// GetAlbumFileCount returns the number of files in an album (dynamic count)
func (s *AlbumService) GetAlbumFileCount(albumID int64) (int, error) {
	// Get all folder configurations for this album
	folderConfigs, err := s.ListAlbumFolders(albumID)
	if err != nil {
		return 0, err
	}

	if len(folderConfigs) == 0 {
		return 0, nil
	}

	// Build dynamic query to count all matching files
	var queryParts []string
	var args []interface{}

	for _, config := range folderConfigs {
		if config.PathPrefix == "" {
			queryParts = append(queryParts, `
				SELECT DISTINCT ffm.file_id
				FROM file_folder_mappings ffm
				WHERE ffm.folder_id = ?
			`)
			args = append(args, config.FolderID)
		} else {
			queryParts = append(queryParts, `
				SELECT DISTINCT ffm.file_id
				FROM file_folder_mappings ffm
				WHERE ffm.folder_id = ? AND ffm.relative_path LIKE ?
			`)
			args = append(args, config.FolderID, config.PathPrefix+"%")
		}
	}

	// Combine with UNION and count
	query := "SELECT COUNT(*) FROM (" + queryParts[0]
	for i := 1; i < len(queryParts); i++ {
		query += " UNION " + queryParts[i]
	}
	query += ")"

	var count int
	err = s.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// FolderConfig represents a folder configuration for an album
type FolderConfig struct {
	FolderID   int64  `json:"folder_id"`
	PathPrefix string `json:"path_prefix"`
}

// AddFolders adds folder configurations to an album
func (s *AlbumService) AddFolders(albumID int64, folderConfigs []FolderConfig) error {
	for _, config := range folderConfigs {
		_, err := s.db.Exec(`
			INSERT OR IGNORE INTO album_folders (album_id, folder_id, path_prefix)
			VALUES (?, ?, ?)
		`, albumID, config.FolderID, config.PathPrefix)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveFolder removes a folder configuration from an album
func (s *AlbumService) RemoveFolder(albumID, folderID int64, pathPrefix string) error {
	_, err := s.db.Exec(`
		DELETE FROM album_folders
		WHERE album_id = ? AND folder_id = ? AND path_prefix = ?
	`, albumID, folderID, pathPrefix)
	return err
}

// ListAlbumFolders retrieves all folder configurations for an album
func (s *AlbumService) ListAlbumFolders(albumID int64) ([]models.AlbumFolder, error) {
	rows, err := s.db.Query(`
		SELECT id, album_id, folder_id, path_prefix, added_at
		FROM album_folders
		WHERE album_id = ?
		ORDER BY added_at DESC
	`, albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []models.AlbumFolder
	for rows.Next() {
		var folder models.AlbumFolder
		if err := rows.Scan(&folder.ID, &folder.AlbumID, &folder.FolderID,
			&folder.PathPrefix, &folder.AddedAt); err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}

	return folders, nil
}

