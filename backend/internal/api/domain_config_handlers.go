package api

import (
	"awesome-sharing/internal/middleware"
	"awesome-sharing/internal/services"
	"github.com/gofiber/fiber/v2"
)

type DomainConfigHandlers struct {
	service *services.DomainConfigService
}

func NewDomainConfigHandlers(service *services.DomainConfigService) *DomainConfigHandlers {
	return &DomainConfigHandlers{
		service: service,
	}
}

// GetDomainConfig godoc
// @Summary Get domain configuration
// @Description Get the current domain configuration (Admin only)
// @Tags domain-config
// @Produce json
// @Success 200 {object} models.DomainConfig
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/domain-config [get]
func (h *DomainConfigHandlers) GetDomainConfig(c *fiber.Ctx) error {
	config, err := h.service.GetConfig()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve domain configuration",
		})
	}

	return c.JSON(config)
}

// SaveDomainConfig godoc
// @Summary Save domain configuration
// @Description Save or update the domain configuration (Admin only)
// @Tags domain-config
// @Accept json
// @Produce json
// @Param config body SaveDomainConfigRequest true "Domain configuration"
// @Success 200 {object} models.DomainConfig
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/domain-config [post]
func (h *DomainConfigHandlers) SaveDomainConfig(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req SaveDomainConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	config, err := h.service.SaveConfig(req.Protocol, req.Domain, req.Port, user.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(config)
}

type SaveDomainConfigRequest struct {
	Protocol string `json:"protocol"` // http or https
	Domain   string `json:"domain"`   // example.com or IP address
	Port     string `json:"port"`     // 80, 443, 8080, etc.
}
