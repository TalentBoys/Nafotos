package services

import (
	"database/sql"
	"time"

	"awesome-sharing/internal/models"
)

type SettingsService struct {
	db *sql.DB
}

func NewSettingsService(db *sql.DB) *SettingsService {
	return &SettingsService{db: db}
}

// GetSetting retrieves a system setting by key
func (s *SettingsService) GetSetting(key string) (*models.SystemSetting, error) {
	var setting models.SystemSetting
	err := s.db.QueryRow(`
		SELECT key, value, updated_at
		FROM system_settings WHERE key = ?
	`, key).Scan(&setting.Key, &setting.Value, &setting.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &setting, nil
}

// GetAllSettings retrieves all system settings
func (s *SettingsService) GetAllSettings() (map[string]string, error) {
	rows, err := s.db.Query("SELECT key, value FROM system_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}

	return settings, nil
}

// SetSetting sets or updates a system setting
func (s *SettingsService) SetSetting(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO system_settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`, key, value, time.Now(), value, time.Now())
	return err
}

// SetSettings sets or updates multiple system settings
func (s *SettingsService) SetSettings(settings map[string]string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO system_settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for key, value := range settings {
		_, err = stmt.Exec(key, value, now, value, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeleteSetting deletes a system setting
func (s *SettingsService) DeleteSetting(key string) error {
	_, err := s.db.Exec("DELETE FROM system_settings WHERE key = ?", key)
	return err
}

// GetDomain retrieves the configured domain
func (s *SettingsService) GetDomain() (string, error) {
	setting, err := s.GetSetting("domain")
	if err != nil {
		return "", err
	}
	if setting == nil {
		return "localhost:8080", nil
	}
	return setting.Value, nil
}

// SetDomain sets the configured domain
func (s *SettingsService) SetDomain(domain string) error {
	return s.SetSetting("domain", domain)
}

// GetSiteName retrieves the site name
func (s *SettingsService) GetSiteName() (string, error) {
	setting, err := s.GetSetting("site_name")
	if err != nil {
		return "", err
	}
	if setting == nil {
		return "AwesomeSharing", nil
	}
	return setting.Value, nil
}

// IsRegistrationAllowed checks if registration is allowed
func (s *SettingsService) IsRegistrationAllowed() (bool, error) {
	setting, err := s.GetSetting("allow_registration")
	if err != nil {
		return false, err
	}
	if setting == nil {
		return false, nil
	}
	return setting.Value == "true", nil
}
