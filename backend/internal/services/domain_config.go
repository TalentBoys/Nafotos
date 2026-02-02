package services

import (
	"database/sql"
	"errors"
	"time"

	"awesome-sharing/internal/database"
	"awesome-sharing/internal/models"
)

type DomainConfigService struct {
	db *database.DB
}

func NewDomainConfigService(db *database.DB) *DomainConfigService {
	return &DomainConfigService{db: db}
}

// GetConfig retrieves the current domain configuration
func (s *DomainConfigService) GetConfig() (*models.DomainConfig, error) {
	var config models.DomainConfig

	err := s.db.QueryRow(`
		SELECT id, protocol, domain, port, updated_by, updated_at
		FROM domain_config
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&config.ID, &config.Protocol, &config.Domain, &config.Port, &config.UpdatedBy, &config.UpdatedAt)

	if err == sql.ErrNoRows {
		// Return default config if none exists
		return &models.DomainConfig{
			Protocol: "http",
			Domain:   "localhost",
			Port:     "8080",
		}, nil
	}

	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves or updates the domain configuration
func (s *DomainConfigService) SaveConfig(protocol, domain, port string, userID int64) (*models.DomainConfig, error) {
	// Validate inputs
	if protocol != "http" && protocol != "https" {
		return nil, errors.New("protocol must be either 'http' or 'https'")
	}

	if domain == "" {
		return nil, errors.New("domain cannot be empty")
	}

	if port == "" {
		return nil, errors.New("port cannot be empty")
	}

	// Insert new configuration
	result, err := s.db.Exec(`
		INSERT INTO domain_config (protocol, domain, port, updated_by, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, protocol, domain, port, userID, time.Now())

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Return the saved config
	return &models.DomainConfig{
		ID:        id,
		Protocol:  protocol,
		Domain:    domain,
		Port:      port,
		UpdatedBy: &userID,
		UpdatedAt: time.Now(),
	}, nil
}

// GetFullURL returns the full URL based on the current configuration
func (s *DomainConfigService) GetFullURL() (string, error) {
	config, err := s.GetConfig()
	if err != nil {
		return "", err
	}

	url := config.Protocol + "://" + config.Domain

	// Only add port if it's not the default port for the protocol
	if (config.Protocol == "http" && config.Port != "80") ||
		(config.Protocol == "https" && config.Port != "443") {
		url += ":" + config.Port
	}

	return url, nil
}
