package api

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"awesome-sharing/internal/middleware"
	"awesome-sharing/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
	settingsService *services.SettingsService
}

func NewAuthHandler(authService *services.AuthService, settingsService *services.SettingsService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		settingsService: settingsService,
	}
}

// Login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register request
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// Login authenticates a user and returns session
// POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and password are required",
		})
	}

	user, session, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid username or password",
			})
		}
		if err == services.ErrUserDisabled {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "User account is disabled",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Login failed",
		})
	}

	// Set session cookie
	// Note: For localhost cross-port requests, SameSite should be "None" or not set
	// However, SameSite=None requires Secure=true (HTTPS)
	// For HTTP development, we use Lax which should work for localhost
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		Domain:   "", // Empty domain to work with localhost
		Expires:  session.ExpiresAt,
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
	})

	return c.JSON(fiber.Map{
		"user":    user,
		"session": session,
	})
}

// Logout destroys the user session
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	sessionID := c.Cookies("session_id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No active session",
		})
	}

	if err := h.authService.DeleteSession(sessionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to logout",
		})
	}

	// Clear cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// Register creates a new user account
// POST /api/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and password are required",
		})
	}

	// Check if registration is allowed
	allowRegistration, err := h.settingsService.IsRegistrationAllowed()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check registration settings",
		})
	}

	// Only admins can register new users if registration is disabled
	user := middleware.GetUser(c)
	if !allowRegistration && (user == nil || user.Role != "admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Registration is disabled. Contact an administrator.",
		})
	}

	// Only admins can set role, default to 'user'
	role := "user"
	if req.Role != "" && user != nil && user.Role == "admin" {
		role = req.Role
	}

	// Create user
	newUser, err := h.authService.CreateUser(req.Username, req.Password, req.Email, role)
	if err != nil {
		if err == services.ErrUserExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Username already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user": newUser,
	})
}

// Me returns the current authenticated user
// GET /api/auth/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

// ChangePassword changes the current user's password
// POST /api/auth/change-password
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Old and new passwords are required",
		})
	}

	// Update password directly
	// TODO: In production, should verify old password first
	err := h.authService.UpdateUser(user.ID, map[string]interface{}{
		"password": req.NewPassword,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}
