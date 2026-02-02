package services

import (
	"database/sql"
	"errors"
	"time"

	"awesome-sharing/internal/models"
)

var (
	ErrPermissionGroupNotFound = errors.New("permission group not found")
	ErrPermissionDenied        = errors.New("permission denied")
)

type PermissionGroupService struct {
	db *sql.DB
}

func NewPermissionGroupService(db *sql.DB) *PermissionGroupService {
	return &PermissionGroupService{db: db}
}

// CreatePermissionGroup creates a new permission group
func (s *PermissionGroupService) CreatePermissionGroup(name, description string, createdBy int64) (*models.PermissionGroup, error) {
	result, err := s.db.Exec(`
		INSERT INTO permission_groups (name, description, created_by)
		VALUES (?, ?, ?)
	`, name, description, createdBy)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Automatically grant write permission to the creator
	err = s.GrantPermission(id, createdBy, "write")
	if err != nil {
		return nil, err
	}

	return s.GetPermissionGroup(id)
}

// GetPermissionGroup retrieves a permission group by ID
func (s *PermissionGroupService) GetPermissionGroup(id int64) (*models.PermissionGroup, error) {
	var pg models.PermissionGroup
	err := s.db.QueryRow(`
		SELECT id, name, description, created_by, created_at, updated_at
		FROM permission_groups WHERE id = ?
	`, id).Scan(&pg.ID, &pg.Name, &pg.Description, &pg.CreatedBy,
		&pg.CreatedAt, &pg.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrPermissionGroupNotFound
	}
	if err != nil {
		return nil, err
	}

	return &pg, nil
}

// ListPermissionGroups retrieves permission groups accessible to a user
func (s *PermissionGroupService) ListPermissionGroups(userID int64, isAdmin bool) ([]models.PermissionGroup, error) {
	var rows *sql.Rows
	var err error

	if isAdmin {
		// Admin can see all permission groups
		rows, err = s.db.Query(`
			SELECT id, name, description, created_by, created_at, updated_at
			FROM permission_groups
			ORDER BY created_at DESC
		`)
	} else {
		// Regular users can only see permission groups they have access to
		rows, err = s.db.Query(`
			SELECT DISTINCT pg.id, pg.name, pg.description, pg.created_by, pg.created_at, pg.updated_at
			FROM permission_groups pg
			INNER JOIN permission_group_permissions pgp ON pg.id = pgp.permission_group_id
			WHERE pgp.user_id = ?
			ORDER BY pg.created_at DESC
		`, userID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.PermissionGroup
	for rows.Next() {
		var pg models.PermissionGroup
		if err := rows.Scan(&pg.ID, &pg.Name, &pg.Description, &pg.CreatedBy,
			&pg.CreatedAt, &pg.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, pg)
	}

	return groups, nil
}

// UpdatePermissionGroup updates permission group information
func (s *PermissionGroupService) UpdatePermissionGroup(id int64, name, description string) error {
	_, err := s.db.Exec(`
		UPDATE permission_groups
		SET name = ?, description = ?, updated_at = ?
		WHERE id = ?
	`, name, description, time.Now(), id)
	return err
}

// DeletePermissionGroup deletes a permission group
func (s *PermissionGroupService) DeletePermissionGroup(id int64) error {
	_, err := s.db.Exec("DELETE FROM permission_groups WHERE id = ?", id)
	return err
}

// AddFolder adds a folder to a permission group
func (s *PermissionGroupService) AddFolder(groupID, folderID int64) error {
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO permission_group_folders (permission_group_id, folder_id)
		VALUES (?, ?)
	`, groupID, folderID)
	return err
}

// RemoveFolder removes a folder from a permission group
func (s *PermissionGroupService) RemoveFolder(groupID, folderID int64) error {
	_, err := s.db.Exec(`
		DELETE FROM permission_group_folders
		WHERE permission_group_id = ? AND folder_id = ?
	`, groupID, folderID)
	return err
}

// ListFoldersInGroup retrieves all folders in a permission group
func (s *PermissionGroupService) ListFoldersInGroup(groupID int64) ([]models.Folder, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.name, f.absolute_path, f.enabled, f.created_by, f.created_at, f.updated_at
		FROM folders f
		INNER JOIN permission_group_folders pgf ON f.id = pgf.folder_id
		WHERE pgf.permission_group_id = ?
		ORDER BY f.created_at DESC
	`, groupID)
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

// GrantPermission grants a user permission to a permission group
func (s *PermissionGroupService) GrantPermission(groupID, userID int64, permission string) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO permission_group_permissions (permission_group_id, user_id, permission)
		VALUES (?, ?, ?)
	`, groupID, userID, permission)
	return err
}

// RevokePermission revokes a user's permission to a permission group
func (s *PermissionGroupService) RevokePermission(groupID, userID int64) error {
	_, err := s.db.Exec(`
		DELETE FROM permission_group_permissions
		WHERE permission_group_id = ? AND user_id = ?
	`, groupID, userID)
	return err
}

// ListUsersWithAccess retrieves all users with access to a permission group
func (s *PermissionGroupService) ListUsersWithAccess(groupID int64) ([]struct {
	User       models.User
	Permission string
	GrantedAt  time.Time
}, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.username, u.email, u.role, u.enabled,
		       u.created_at, u.updated_at, u.last_login_at, u.password_changed_at,
		       pgp.permission, pgp.granted_at
		FROM users u
		INNER JOIN permission_group_permissions pgp ON u.id = pgp.user_id
		WHERE pgp.permission_group_id = ?
		ORDER BY pgp.granted_at DESC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		User       models.User
		Permission string
		GrantedAt  time.Time
	}

	for rows.Next() {
		var result struct {
			User       models.User
			Permission string
			GrantedAt  time.Time
		}
		if err := rows.Scan(
			&result.User.ID, &result.User.Username, &result.User.Email,
			&result.User.Role, &result.User.Enabled,
			&result.User.CreatedAt, &result.User.UpdatedAt,
			&result.User.LastLoginAt, &result.User.PasswordChangedAt,
			&result.Permission, &result.GrantedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// CheckPermission checks if a user has permission to access a permission group
func (s *PermissionGroupService) CheckPermission(groupID, userID int64, requiredPermission string) (bool, error) {
	var permission string
	err := s.db.QueryRow(`
		SELECT permission FROM permission_group_permissions
		WHERE permission_group_id = ? AND user_id = ?
	`, groupID, userID).Scan(&permission)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// 'write' permission includes 'read'
	if requiredPermission == "read" && (permission == "read" || permission == "write") {
		return true, nil
	}
	if requiredPermission == "write" && permission == "write" {
		return true, nil
	}

	return false, nil
}

// CheckFileAccess checks if a user has access to a specific file through permission groups
func (s *PermissionGroupService) CheckFileAccess(userID, fileID int64, isAdmin bool) (bool, error) {
	// Admin always has access
	if isAdmin {
		return true, nil
	}

	// Check if user has permission to any permission group that contains a folder with this file
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(DISTINCT pgp.permission_group_id)
		FROM permission_group_permissions pgp
		INNER JOIN permission_group_folders pgf ON pgp.permission_group_id = pgf.permission_group_id
		INNER JOIN file_folder_mappings ffm ON pgf.folder_id = ffm.folder_id
		WHERE pgp.user_id = ? AND ffm.file_id = ?
	`, userID, fileID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CheckFolderAccess checks if a user has access to a specific folder through permission groups
func (s *PermissionGroupService) CheckFolderAccess(userID, folderID int64, isAdmin bool) (bool, error) {
	// Admin always has access
	if isAdmin {
		return true, nil
	}

	// Check if user has permission to any permission group containing this folder
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(DISTINCT pgp.permission_group_id)
		FROM permission_group_permissions pgp
		INNER JOIN permission_group_folders pgf ON pgp.permission_group_id = pgf.permission_group_id
		WHERE pgp.user_id = ? AND pgf.folder_id = ?
	`, userID, folderID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetPermissionGroupsForFolder retrieves all permission groups that contain a specific folder
func (s *PermissionGroupService) GetPermissionGroupsForFolder(folderID int64) ([]models.PermissionGroup, error) {
	rows, err := s.db.Query(`
		SELECT pg.id, pg.name, pg.description, pg.created_by, pg.created_at, pg.updated_at
		FROM permission_groups pg
		INNER JOIN permission_group_folders pgf ON pg.id = pgf.permission_group_id
		WHERE pgf.folder_id = ?
		ORDER BY pg.created_at DESC
	`, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.PermissionGroup
	for rows.Next() {
		var pg models.PermissionGroup
		if err := rows.Scan(&pg.ID, &pg.Name, &pg.Description, &pg.CreatedBy,
			&pg.CreatedAt, &pg.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, pg)
	}

	return groups, nil
}
