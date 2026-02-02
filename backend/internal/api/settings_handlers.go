package api

import (
	"github.com/gofiber/fiber/v2"

	"awesome-sharing/internal/services"
)

type SettingsHandler struct {
	settingsService *services.SettingsService
}

func NewSettingsHandler(settingsService *services.SettingsService) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
	}
}

// GetSettings returns all system settings (admin only)
// GET /api/settings
func (h *SettingsHandler) GetSettings(c *fiber.Ctx) error {
	settings, err := h.settingsService.GetAllSettings()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch settings",
		})
	}

	return c.JSON(fiber.Map{
		"settings": settings,
	})
}

// GetPublicSettings returns public settings (no auth required)
// GET /api/settings/public
func (h *SettingsHandler) GetPublicSettings(c *fiber.Ctx) error {
	siteName, _ := h.settingsService.GetSiteName()
	allowRegistration, _ := h.settingsService.IsRegistrationAllowed()

	return c.JSON(fiber.Map{
		"site_name":          siteName,
		"allow_registration": allowRegistration,
	})
}

// UpdateSettings updates system settings (admin only)
// PUT /api/settings
func (h *SettingsHandler) UpdateSettings(c *fiber.Ctx) error {
	var req map[string]string

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No settings to update",
		})
	}

	err := h.settingsService.SetSettings(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update settings",
		})
	}

	// Return updated settings
	settings, err := h.settingsService.GetAllSettings()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch updated settings",
		})
	}

	return c.JSON(fiber.Map{
		"settings": settings,
	})
}

// GetDomain returns the configured domain
// GET /api/settings/domain
func (h *SettingsHandler) GetDomain(c *fiber.Ctx) error {
	domain, err := h.settingsService.GetDomain()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch domain",
		})
	}

	return c.JSON(fiber.Map{
		"domain": domain,
	})
}

// UpdateDomain updates the domain setting (admin only)
// PUT /api/settings/domain
func (h *SettingsHandler) UpdateDomain(c *fiber.Ctx) error {
	var req struct {
		Domain string `json:"domain"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Domain is required",
		})
	}

	err := h.settingsService.SetDomain(req.Domain)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update domain",
		})
	}

	return c.JSON(fiber.Map{
		"domain": req.Domain,
	})
}
