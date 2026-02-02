package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"awesome-sharing/internal/models"
)

var (
	ErrShareNotFound   = errors.New("share not found")
	ErrShareExpired    = errors.New("share has expired")
	ErrShareDisabled   = errors.New("share is disabled")
	ErrMaxViewsReached = errors.New("maximum views reached")
	ErrInvalidPassword = errors.New("invalid password")
	ErrAccessDenied    = errors.New("access denied")
)

type ShareService struct {
	db *sql.DB
}

func NewShareService(db *sql.DB) *ShareService {
	return &ShareService{db: db}
}

// CreateShare creates a new share link
func (s *ShareService) CreateShare(shareType string, resourceID, ownerID int64, accessType string, password string, requiresAuth bool, expiresAt *time.Time, maxViews *int) (*models.Share, error) {
	// Generate short share ID
	shareID := generateShortID(8)

	var passwordHash string
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		passwordHash = string(hash)
	}

	_, err := s.db.Exec(`
		INSERT INTO shares (id, share_type, resource_id, owner_id, access_type, password_hash, requires_auth, expires_at, max_views, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
	`, shareID, shareType, resourceID, ownerID, accessType, passwordHash, requiresAuth, expiresAt, maxViews)
	if err != nil {
		return nil, err
	}

	return s.GetShare(shareID)
}

// GetShare retrieves a share by ID
func (s *ShareService) GetShare(id string) (*models.Share, error) {
	var share models.Share
	var passwordHash sql.NullString

	err := s.db.QueryRow(`
		SELECT id, share_type, resource_id, owner_id, access_type, password_hash, requires_auth, expires_at, max_views, view_count, enabled, created_at
		FROM shares WHERE id = ?
	`, id).Scan(&share.ID, &share.ShareType, &share.ResourceID, &share.OwnerID,
		&share.AccessType, &passwordHash, &share.RequiresAuth, &share.ExpiresAt, &share.MaxViews,
		&share.ViewCount, &share.Enabled, &share.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrShareNotFound
	}
	if err != nil {
		return nil, err
	}

	if passwordHash.Valid && passwordHash.String != "" {
		share.PasswordHash = passwordHash.String
		share.HasPassword = true
	}

	return &share, nil
}

// ValidateShareAccess validates if a share can be accessed
func (s *ShareService) ValidateShareAccess(shareID, password string, userID *int64) (*models.Share, error) {
	share, err := s.GetShare(shareID)
	if err != nil {
		return nil, err
	}

	// Check if enabled
	if !share.Enabled {
		return nil, ErrShareDisabled
	}

	// Check expiration
	if share.ExpiresAt != nil && time.Now().After(*share.ExpiresAt) {
		return nil, ErrShareExpired
	}

	// Check max views
	if share.MaxViews != nil && share.ViewCount >= *share.MaxViews {
		return nil, ErrMaxViewsReached
	}

	// Check if authentication is required
	if share.RequiresAuth && userID == nil {
		return nil, ErrAccessDenied
	}

	// Check password if set
	if share.PasswordHash != "" {
		if password == "" {
			return nil, ErrInvalidPassword
		}
		if err := bcrypt.CompareHashAndPassword([]byte(share.PasswordHash), []byte(password)); err != nil {
			return nil, ErrInvalidPassword
		}
	}

	// Check private share permissions
	if share.AccessType == "private" {
		if userID == nil {
			return nil, ErrAccessDenied
		}
		// Check if user has permission
		hasPermission, err := s.CheckSharePermission(shareID, *userID)
		if err != nil {
			return nil, err
		}
		if !hasPermission && share.OwnerID != *userID {
			return nil, ErrAccessDenied
		}
	}

	return share, nil
}

// LogAccess logs a share access
func (s *ShareService) LogAccess(shareID string, userID *int64, ipAddress, userAgent string) error {
	// Increment view count
	_, err := s.db.Exec("UPDATE shares SET view_count = view_count + 1 WHERE id = ?", shareID)
	if err != nil {
		return err
	}

	// Log access
	_, err = s.db.Exec(`
		INSERT INTO share_access_log (share_id, accessed_by, ip_address, user_agent)
		VALUES (?, ?, ?, ?)
	`, shareID, userID, ipAddress, userAgent)
	return err
}

// ListSharesByOwner retrieves all shares created by a user
func (s *ShareService) ListSharesByOwner(ownerID int64) ([]models.Share, error) {
	rows, err := s.db.Query(`
		SELECT id, share_type, resource_id, owner_id, access_type, password_hash, requires_auth, expires_at, max_views, view_count, enabled, created_at
		FROM shares WHERE owner_id = ?
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []models.Share
	for rows.Next() {
		var share models.Share
		var passwordHash sql.NullString
		if err := rows.Scan(&share.ID, &share.ShareType, &share.ResourceID, &share.OwnerID,
			&share.AccessType, &passwordHash, &share.RequiresAuth, &share.ExpiresAt, &share.MaxViews, &share.ViewCount,
			&share.Enabled, &share.CreatedAt); err != nil {
			return nil, err
		}
		if passwordHash.Valid && passwordHash.String != "" {
			share.HasPassword = true
		}
		shares = append(shares, share)
	}

	return shares, nil
}

// UpdateShare updates share settings
func (s *ShareService) UpdateShare(id string, updates map[string]interface{}) error {
	if expiresAt, ok := updates["expires_at"]; ok {
		_, err := s.db.Exec("UPDATE shares SET expires_at = ? WHERE id = ?", expiresAt, id)
		if err != nil {
			return err
		}
	}

	if enabled, ok := updates["enabled"]; ok {
		_, err := s.db.Exec("UPDATE shares SET enabled = ? WHERE id = ?", enabled, id)
		if err != nil {
			return err
		}
	}

	if maxViews, ok := updates["max_views"]; ok {
		_, err := s.db.Exec("UPDATE shares SET max_views = ? WHERE id = ?", maxViews, id)
		if err != nil {
			return err
		}
	}

	if requiresAuth, ok := updates["requires_auth"]; ok {
		_, err := s.db.Exec("UPDATE shares SET requires_auth = ? WHERE id = ?", requiresAuth, id)
		if err != nil {
			return err
		}
	}

	if password, ok := updates["password"]; ok {
		var passwordHash string
		if password != nil && password.(string) != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(password.(string)), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			passwordHash = string(hash)
		}
		_, err := s.db.Exec("UPDATE shares SET password_hash = ? WHERE id = ?", passwordHash, id)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteShare deletes a share
func (s *ShareService) DeleteShare(id string) error {
	_, err := s.db.Exec("DELETE FROM shares WHERE id = ?", id)
	return err
}

// DeleteExpiredShares deletes all expired shares
func (s *ShareService) DeleteExpiredShares() (int64, error) {
	result, err := s.db.Exec("DELETE FROM shares WHERE expires_at IS NOT NULL AND expires_at < ?", time.Now())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ExtendShare extends the expiration of a share
func (s *ShareService) ExtendShare(id string, duration time.Duration) error {
	share, err := s.GetShare(id)
	if err != nil {
		return err
	}

	var newExpiresAt time.Time
	if share.ExpiresAt != nil {
		newExpiresAt = share.ExpiresAt.Add(duration)
	} else {
		newExpiresAt = time.Now().Add(duration)
	}

	_, err = s.db.Exec("UPDATE shares SET expires_at = ? WHERE id = ?", newExpiresAt, id)
	return err
}

// GrantSharePermission grants a user access to a private share
func (s *ShareService) GrantSharePermission(shareID string, userID int64) error {
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO share_permissions (share_id, user_id)
		VALUES (?, ?)
	`, shareID, userID)
	return err
}

// RevokeSharePermission revokes a user's access to a private share
func (s *ShareService) RevokeSharePermission(shareID string, userID int64) error {
	_, err := s.db.Exec(`
		DELETE FROM share_permissions WHERE share_id = ? AND user_id = ?
	`, shareID, userID)
	return err
}

// CheckSharePermission checks if a user has permission to access a private share
func (s *ShareService) CheckSharePermission(shareID string, userID int64) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM share_permissions WHERE share_id = ? AND user_id = ?)
	`, shareID, userID).Scan(&exists)
	return exists, err
}

// ListSharePermissions retrieves all permissions for a share
func (s *ShareService) ListSharePermissions(shareID string) ([]models.SharePermission, error) {
	rows, err := s.db.Query(`
		SELECT id, share_id, user_id, granted_at
		FROM share_permissions WHERE share_id = ?
		ORDER BY granted_at DESC
	`, shareID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.SharePermission
	for rows.Next() {
		var perm models.SharePermission
		if err := rows.Scan(&perm.ID, &perm.ShareID, &perm.UserID, &perm.GrantedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// GetAccessLog retrieves access log for a share
func (s *ShareService) GetAccessLog(shareID string, limit int) ([]models.ShareAccessLog, error) {
	rows, err := s.db.Query(`
		SELECT id, share_id, accessed_by, ip_address, user_agent, accessed_at
		FROM share_access_log WHERE share_id = ?
		ORDER BY accessed_at DESC
		LIMIT ?
	`, shareID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.ShareAccessLog
	for rows.Next() {
		var log models.ShareAccessLog
		if err := rows.Scan(&log.ID, &log.ShareID, &log.AccessedBy, &log.IPAddress,
			&log.UserAgent, &log.AccessedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// generateShortID generates a short random ID for shares
func generateShortID(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	id := base64.URLEncoding.EncodeToString(bytes)
	id = strings.TrimRight(id, "=")
	if len(id) > length {
		id = id[:length]
	}
	return id
}

// GenerateAccessToken generates a temporary access token for a share
// Token format: shareID:timestamp:signature
func (s *ShareService) GenerateAccessToken(shareID string) (string, error) {
	share, err := s.GetShare(shareID)
	if err != nil {
		return "", err
	}

	// Generate a random token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := base64.URLEncoding.EncodeToString(tokenBytes)
	token = strings.TrimRight(token, "=")

	// Token format: shareID:resourceID:token
	accessToken := fmt.Sprintf("%s:%d:%s", shareID, share.ResourceID, token)
	return accessToken, nil
}

// ValidateAccessToken validates an access token and returns the share and resource ID
func (s *ShareService) ValidateAccessToken(token string) (string, int64, error) {
	// Parse token format: shareID:resourceID:token
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		return "", 0, errors.New("invalid token format")
	}

	shareID := parts[0]
	var resourceID int64
	fmt.Sscanf(parts[1], "%d", &resourceID)

	// Verify the share still exists and is valid
	share, err := s.GetShare(shareID)
	if err != nil {
		return "", 0, err
	}

	// Check if share is enabled
	if !share.Enabled {
		return "", 0, ErrShareDisabled
	}

	// Check expiration
	if share.ExpiresAt != nil && time.Now().After(*share.ExpiresAt) {
		return "", 0, ErrShareExpired
	}

	// Verify resource ID matches
	if share.ResourceID != resourceID {
		return "", 0, errors.New("resource ID mismatch")
	}

	return shareID, resourceID, nil
}
