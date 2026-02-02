package middleware

import (
	"github.com/gofiber/fiber/v2"

	"awesome-sharing/internal/models"
	"awesome-sharing/internal/services"
)

const (
	UserContextKey = "user"
)

// AuthMiddleware creates a middleware that validates session and injects user into context
func AuthMiddleware(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get session ID from cookie
		sessionID := c.Cookies("session_id")
		if sessionID == "" {
			// Also check Authorization header
			sessionID = c.Get("Authorization")
			if sessionID != "" && len(sessionID) > 7 && sessionID[:7] == "Bearer " {
				sessionID = sessionID[7:]
			}
		}

		if sessionID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No session provided",
			})
		}

		// Validate session
		user, err := authService.ValidateSession(sessionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		}

		// Check if user is enabled
		if !user.Enabled {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "User is disabled",
			})
		}

		// Store user in context
		c.Locals(UserContextKey, user)

		return c.Next()
	}
}

// OptionalAuthMiddleware is like AuthMiddleware but doesn't fail if no session
func OptionalAuthMiddleware(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session_id")
		if sessionID == "" {
			sessionID = c.Get("Authorization")
			if sessionID != "" && len(sessionID) > 7 && sessionID[:7] == "Bearer " {
				sessionID = sessionID[7:]
			}
		}

		if sessionID != "" {
			user, err := authService.ValidateSession(sessionID)
			if err == nil && user.Enabled {
				c.Locals(UserContextKey, user)
			}
		}

		return c.Next()
	}
}

// AdminOnlyMiddleware ensures the user is an admin
func AdminOnlyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		if user.Role != "admin" && user.Role != "server_owner" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}

		return c.Next()
	}
}

// ServerOwnerOnlyMiddleware ensures the user is the server owner
func ServerOwnerOnlyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		if user.Role != "server_owner" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Server owner access required",
			})
		}

		return c.Next()
	}
}

// AdminOrOwnerMiddleware allows both admin and server owner
func AdminOrOwnerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		if user.Role != "admin" && user.Role != "server_owner" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin or server owner access required",
			})
		}

		return c.Next()
	}
}

// GetUser retrieves the user from the fiber context
func GetUser(c *fiber.Ctx) *models.User {
	user := c.Locals(UserContextKey)
	if user == nil {
		return nil
	}
	return user.(*models.User)
}

// IsAdmin checks if the current user is an admin
func IsAdmin(c *fiber.Ctx) bool {
	user := GetUser(c)
	return user != nil && (user.Role == "admin" || user.Role == "server_owner")
}

// IsServerOwner checks if the current user is the server owner
func IsServerOwner(c *fiber.Ctx) bool {
	user := GetUser(c)
	return user != nil && user.Role == "server_owner"
}

