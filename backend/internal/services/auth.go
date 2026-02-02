package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"awesome-sharing/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrUserExists         = errors.New("username already exists")
)

type AuthService struct {
	db *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

// HashPassword hashes a plain password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword verifies a password against its hash
func (s *AuthService) CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// CreateUser creates a new user
func (s *AuthService) CreateUser(username, password, email, role string) (*models.User, error) {
	// Check if user exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// Hash password
	passwordHash, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Insert user
	result, err := s.db.Exec(`
		INSERT INTO users (username, password_hash, email, role, enabled)
		VALUES (?, ?, ?, ?, 1)
	`, username, passwordHash, email, role)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Return user
	return s.GetUserByID(id)
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(username, password string) (*models.User, *models.Session, error) {
	// Get user
	var user models.User
	var passwordHash string
	err := s.db.QueryRow(`
		SELECT id, username, password_hash, email, role, enabled, created_at, updated_at, last_login_at
		FROM users WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &passwordHash, &user.Email, &user.Role,
		&user.Enabled, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)

	if err == sql.ErrNoRows {
		return nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, err
	}

	// Check if user is enabled
	if !user.Enabled {
		return nil, nil, ErrUserDisabled
	}

	// Verify password
	if err := s.CheckPassword(password, passwordHash); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Update last login time
	now := time.Now()
	_, err = s.db.Exec("UPDATE users SET last_login_at = ? WHERE id = ?", now, user.ID)
	if err != nil {
		return nil, nil, err
	}
	user.LastLoginAt = &now

	// Create session
	session, err := s.CreateSession(user.ID, 24*time.Hour*7) // 7 days
	if err != nil {
		return nil, nil, err
	}

	return &user, session, nil
}

// CreateSession creates a new session for a user
func (s *AuthService) CreateSession(userID int64, duration time.Duration) (*models.Session, error) {
	// Generate random session ID
	sessionID, err := generateRandomID(32)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(duration)

	_, err = s.db.Exec(`
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES (?, ?, ?)
	`, sessionID, userID, expiresAt)
	if err != nil {
		return nil, err
	}

	return &models.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// ValidateSession validates a session and returns the associated user
func (s *AuthService) ValidateSession(sessionID string) (*models.User, error) {
	var session models.Session
	err := s.db.QueryRow(`
		SELECT id, user_id, expires_at, created_at
		FROM sessions WHERE id = ?
	`, sessionID).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("invalid session")
	}
	if err != nil {
		return nil, err
	}

	// Check if session expired
	if time.Now().After(session.ExpiresAt) {
		s.DeleteSession(sessionID)
		return nil, errors.New("session expired")
	}

	// Get user
	return s.GetUserByID(session.UserID)
}

// DeleteSession deletes a session (logout)
func (s *AuthService) DeleteSession(sessionID string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(id int64) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, username, email, role, enabled, created_at, updated_at, last_login_at, password_changed_at
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Username, &user.Email, &user.Role,
		&user.Enabled, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.PasswordChangedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (s *AuthService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, username, email, role, enabled, created_at, updated_at, last_login_at, password_changed_at
		FROM users WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.Email, &user.Role,
		&user.Enabled, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.PasswordChangedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ListUsers retrieves all users (admin only)
func (s *AuthService) ListUsers() ([]models.User, error) {
	rows, err := s.db.Query(`
		SELECT id, username, email, role, enabled, created_at, updated_at, last_login_at, password_changed_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role,
			&user.Enabled, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.PasswordChangedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUser updates user information
func (s *AuthService) UpdateUser(id int64, updates map[string]interface{}) error {
	// Build dynamic update query
	// For simplicity, we'll handle specific fields
	if email, ok := updates["email"]; ok {
		_, err := s.db.Exec("UPDATE users SET email = ?, updated_at = ? WHERE id = ?",
			email, time.Now(), id)
		if err != nil {
			return err
		}
	}

	if role, ok := updates["role"]; ok {
		_, err := s.db.Exec("UPDATE users SET role = ?, updated_at = ? WHERE id = ?",
			role, time.Now(), id)
		if err != nil {
			return err
		}
	}

	if enabled, ok := updates["enabled"]; ok {
		_, err := s.db.Exec("UPDATE users SET enabled = ?, updated_at = ? WHERE id = ?",
			enabled, time.Now(), id)
		if err != nil {
			return err
		}
	}

	if password, ok := updates["password"]; ok {
		passwordHash, err := s.HashPassword(password.(string))
		if err != nil {
			return err
		}
		_, err = s.db.Exec("UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?",
			passwordHash, time.Now(), id)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteUser deletes a user
func (s *AuthService) DeleteUser(id int64) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

// CleanupExpiredSessions removes expired sessions
func (s *AuthService) CleanupExpiredSessions() error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}

// generateRandomID generates a random hex string of given length
func generateRandomID(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ListUsersPaginated retrieves users with pagination, search, and filtering
func (s *AuthService) ListUsersPaginated(page, limit int, search, role string) ([]models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}

	offset := (page - 1) * limit

	// Build query
	query := `SELECT id, username, email, role, enabled, created_at, updated_at, last_login_at, password_changed_at FROM users WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`
	args := []interface{}{}

	// Add search filter
	if search != "" {
		query += ` AND (username LIKE ? OR email LIKE ?)`
		countQuery += ` AND (username LIKE ? OR email LIKE ?)`
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Add role filter
	if role != "" {
		query += ` AND role = ?`
		countQuery += ` AND role = ?`
		args = append(args, role)
	}

	// Get total count
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add ordering and pagination
	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	// Execute query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role,
			&user.Enabled, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.PasswordChangedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

// ResetUserPassword resets a user's password (admin function)
func (s *AuthService) ResetUserPassword(userID int64, newPassword string) error {
	passwordHash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = s.db.Exec(`
		UPDATE users
		SET password_hash = ?, password_changed_at = ?, updated_at = ?
		WHERE id = ?
	`, passwordHash, now, now, userID)

	return err
}

// BulkEnableDisableUsers enables or disables multiple users
func (s *AuthService) BulkEnableDisableUsers(userIDs []int64, enabled bool) error {
	if len(userIDs) == 0 {
		return nil
	}

	// Use transaction for atomicity
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE users SET enabled = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, id := range userIDs {
		if _, err := stmt.Exec(enabled, now, id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BulkDeleteUsers deletes multiple users
func (s *AuthService) BulkDeleteUsers(userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	// Use transaction for atomicity
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`DELETE FROM users WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range userIDs {
		if _, err := stmt.Exec(id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LogUserActivity logs user management actions
func (s *AuthService) LogUserActivity(userID, performedBy int64, action, details, ipAddress string) error {
	_, err := s.db.Exec(`
		INSERT INTO user_activity_logs (user_id, performed_by, action, details, ip_address)
		VALUES (?, ?, ?, ?, ?)
	`, userID, performedBy, action, details, ipAddress)
	return err
}

// GetUserActivityLogs retrieves activity logs for a user with pagination
func (s *AuthService) GetUserActivityLogs(userID int64, page, limit int) ([]models.UserActivityLog, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get total count
	var total int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM user_activity_logs WHERE user_id = ?`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get logs
	rows, err := s.db.Query(`
		SELECT id, user_id, performed_by, action, details, ip_address, created_at
		FROM user_activity_logs
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.UserActivityLog
	for rows.Next() {
		var log models.UserActivityLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.PerformedBy, &log.Action,
			&log.Details, &log.IPAddress, &log.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

// ExportUsers exports user data to CSV format
func (s *AuthService) ExportUsers() ([]byte, error) {
	users, err := s.ListUsers()
	if err != nil {
		return nil, err
	}

	// Build CSV
	csv := "ID,Username,Email,Role,Status,Created,Last Login\n"
	for _, user := range users {
		status := "Enabled"
		if !user.Enabled {
			status = "Disabled"
		}

		lastLogin := "Never"
		if user.LastLoginAt != nil {
			lastLogin = user.LastLoginAt.Format("2006-01-02 15:04:05")
		}

		csv += fmt.Sprintf("%d,\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
			user.ID,
			user.Username,
			user.Email,
			user.Role,
			status,
			user.CreatedAt.Format("2006-01-02 15:04:05"),
			lastLogin,
		)
	}

	return []byte(csv), nil
}
