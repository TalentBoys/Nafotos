package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"awesome-sharing/internal/middleware"
	"awesome-sharing/internal/services"
)

type UserHandler struct {
	authService *services.AuthService
}

func NewUserHandler(authService *services.AuthService) *UserHandler {
	return &UserHandler{
		authService: authService,
	}
}

type UpdateUserRequest struct {
	Email   *string `json:"email"`
	Role    *string `json:"role"`
	Enabled *bool   `json:"enabled"`
}

// ListUsers returns all users (admin only)
// GET /api/users
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	// Check for pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 25)
	search := c.Query("search", "")
	role := c.Query("role", "")

	// Use paginated version if parameters are provided
	if page > 1 || limit != 25 || search != "" || role != "" {
		return h.ListUsersPaginated(c)
	}

	// Otherwise use original behavior for backward compatibility
	users, err := h.authService.ListUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	return c.JSON(fiber.Map{
		"users": users,
		"total": len(users),
	})
}

// ListUsersPaginated returns users with pagination, search, and filters (admin only)
// GET /api/users?page=1&limit=25&search=query&role=admin
func (h *UserHandler) ListUsersPaginated(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 25)
	search := c.Query("search", "")
	role := c.Query("role", "")

	users, total, err := h.authService.ListUsersPaginated(page, limit, search, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(fiber.Map{
		"users":       users,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// GetUser returns a specific user by ID (admin only)
// GET /api/users/:id
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	user, err := h.authService.GetUserByID(id)
	if err != nil {
		if err == services.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

// CreateUser creates a new user (admin only)
// POST /api/users
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

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

	// Default role to 'user' if not specified
	if req.Role == "" {
		req.Role = "user"
	}

	// Validate role
	if req.Role != "admin" && req.Role != "user" && req.Role != "server_owner" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Role must be 'admin', 'user', or 'server_owner'",
		})
	}

	// Prevent creating server_owner accounts (only allowed during initialization)
	if req.Role == "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Cannot create server_owner accounts. Server owner is created during initialization only.",
		})
	}

	// Get current user
	currentUser := middleware.GetUser(c)

	// Admin cannot create other admin accounts
	if req.Role == "admin" && currentUser != nil && currentUser.Role == "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin users cannot create other admin accounts",
		})
	}

	user, err := h.authService.CreateUser(req.Username, req.Password, req.Email, req.Role)
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
		"user": user,
	})
}

// UpdateUser updates a user (admin only)
// PUT /api/users/:id
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get target user
	targetUser, err := h.authService.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Prevent modifying server_owner role
	if targetUser.Role == "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Cannot modify server_owner user. Server owner role is immutable.",
		})
	}

	// Get current user
	currentUser := middleware.GetUser(c)

	// Admin cannot modify other admin accounts
	if targetUser.Role == "admin" && currentUser != nil && currentUser.Role == "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin users cannot modify other admin accounts",
		})
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Role != nil {
		// Validate role
		if *req.Role != "admin" && *req.Role != "user" && *req.Role != "server_owner" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Role must be 'admin', 'user', or 'server_owner'",
			})
		}

		// Prevent promoting to server_owner
		if *req.Role == "server_owner" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Cannot promote users to server_owner. Server owner is created during initialization only.",
			})
		}

		updates["role"] = *req.Role
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No fields to update",
		})
	}

	err = h.authService.UpdateUser(id, updates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	// Return updated user
	user, err := h.authService.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch updated user",
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

// DeleteUser deletes a user (admin only)
// DELETE /api/users/:id
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Prevent self-deletion
	currentUser := middleware.GetUser(c)
	if currentUser != nil && currentUser.ID == id {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot delete your own account",
		})
	}

	// Get target user
	targetUser, err := h.authService.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Prevent deleting server_owner
	if targetUser.Role == "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Cannot delete server_owner user. Server owner is permanent.",
		})
	}

	// Admin cannot delete other admin accounts
	if targetUser.Role == "admin" && currentUser != nil && currentUser.Role == "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin users cannot delete other admin accounts",
		})
	}

	err = h.authService.DeleteUser(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

// ToggleUser enables or disables a user (admin only)
// PUT /api/users/:id/toggle
func (h *UserHandler) ToggleUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Prevent self-disable
	currentUser := middleware.GetUser(c)
	if currentUser != nil && currentUser.ID == id && !req.Enabled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot disable your own account",
		})
	}

	// Get target user
	targetUser, err := h.authService.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Prevent toggling server_owner
	if targetUser.Role == "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Cannot enable/disable server_owner user.",
		})
	}

	// Admin cannot toggle other admin accounts
	if targetUser.Role == "admin" && currentUser != nil && currentUser.Role == "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin users cannot enable/disable other admin accounts",
		})
	}

	err = h.authService.UpdateUser(id, map[string]interface{}{
		"enabled": req.Enabled,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to toggle user",
		})
	}

	// Log activity
	action := "enabled"
	if !req.Enabled {
		action = "disabled"
	}
	h.authService.LogUserActivity(id, currentUser.ID, action, "", c.IP())

	user, err := h.authService.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch updated user",
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

// ResetPassword resets a user's password (admin only)
// POST /api/users/:id/reset-password
func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.NewPassword) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 8 characters",
		})
	}

	// Get current user
	currentUser := middleware.GetUser(c)

	// Get target user
	targetUser, err := h.authService.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Admin cannot reset password for server_owner
	if targetUser.Role == "server_owner" && currentUser != nil && currentUser.Role == "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin users cannot reset password for server_owner",
		})
	}

	// Admin cannot reset password for other admin accounts
	if targetUser.Role == "admin" && currentUser != nil && currentUser.Role == "admin" && targetUser.ID != currentUser.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin users cannot reset password for other admin accounts",
		})
	}

	// Reset password
	err = h.authService.ResetUserPassword(id, req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reset password",
		})
	}

	// Log activity
	h.authService.LogUserActivity(id, currentUser.ID, "password_reset", "", c.IP())

	return c.JSON(fiber.Map{
		"message": "Password reset successfully",
	})
}

// BulkEnableDisable enables or disables multiple users (admin only)
// POST /api/users/bulk/enable-disable
func (h *UserHandler) BulkEnableDisable(c *fiber.Ctx) error {
	var req struct {
		UserIDs []int64 `json:"user_ids"`
		Enabled bool    `json:"enabled"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No user IDs provided",
		})
	}

	if len(req.UserIDs) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot process more than 100 users at once",
		})
	}

	// Prevent self-disable
	currentUser := middleware.GetUser(c)
	if currentUser != nil && !req.Enabled {
		for _, id := range req.UserIDs {
			if id == currentUser.ID {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Cannot disable your own account",
				})
			}
		}
	}

	// Check if any target users are server_owner or admin (for admin users)
	for _, id := range req.UserIDs {
		targetUser, err := h.authService.GetUserByID(id)
		if err == nil {
			if targetUser.Role == "server_owner" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Cannot modify server_owner user in bulk operations",
				})
			}
			// Admin cannot bulk modify other admin accounts
			if targetUser.Role == "admin" && currentUser != nil && currentUser.Role == "admin" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Admin users cannot modify other admin accounts in bulk operations",
				})
			}
		}
	}

	// Perform bulk operation
	err := h.authService.BulkEnableDisableUsers(req.UserIDs, req.Enabled)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update users",
		})
	}

	// Log activities
	action := "enabled"
	if !req.Enabled {
		action = "disabled"
	}
	for _, id := range req.UserIDs {
		h.authService.LogUserActivity(id, currentUser.ID, action, "bulk operation", c.IP())
	}

	return c.JSON(fiber.Map{
		"message": "Users updated successfully",
		"count":   len(req.UserIDs),
	})
}

// BulkDelete deletes multiple users (admin only)
// POST /api/users/bulk/delete
func (h *UserHandler) BulkDelete(c *fiber.Ctx) error {
	var req struct {
		UserIDs []int64 `json:"user_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No user IDs provided",
		})
	}

	if len(req.UserIDs) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot delete more than 100 users at once",
		})
	}

	// Prevent self-deletion
	currentUser := middleware.GetUser(c)
	if currentUser != nil {
		for _, id := range req.UserIDs {
			if id == currentUser.ID {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Cannot delete your own account",
				})
			}
		}
	}

	// Check if any target users are server_owner or admin (for admin users)
	for _, id := range req.UserIDs {
		targetUser, err := h.authService.GetUserByID(id)
		if err == nil {
			if targetUser.Role == "server_owner" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Cannot delete server_owner user",
				})
			}
			// Admin cannot bulk delete other admin accounts
			if targetUser.Role == "admin" && currentUser != nil && currentUser.Role == "admin" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Admin users cannot delete other admin accounts in bulk operations",
				})
			}
		}
	}

	// Log activities before deletion
	for _, id := range req.UserIDs {
		h.authService.LogUserActivity(id, currentUser.ID, "deleted", "bulk operation", c.IP())
	}

	// Perform bulk deletion
	err := h.authService.BulkDeleteUsers(req.UserIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete users",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Users deleted successfully",
		"count":   len(req.UserIDs),
	})
}

// GetUserActivityLogs returns activity logs for a user (admin only)
// GET /api/users/:id/activity-logs
func (h *UserHandler) GetUserActivityLogs(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	logs, total, err := h.authService.GetUserActivityLogs(id, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch activity logs",
		})
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(fiber.Map{
		"logs":        logs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// ExportUsers exports all users to CSV (admin only)
// POST /api/users/export
func (h *UserHandler) ExportUsers(c *fiber.Ctx) error {
	csvData, err := h.authService.ExportUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export users",
		})
	}

	// Log activity
	currentUser := middleware.GetUser(c)
	h.authService.LogUserActivity(currentUser.ID, currentUser.ID, "exported_users", "", c.IP())

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=users_export.csv")
	return c.Send(csvData)
}

// GetUserStats returns user statistics (admin only)
// GET /api/users/stats
func (h *UserHandler) GetUserStats(c *fiber.Ctx) error {
	users, err := h.authService.ListUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	totalUsers := len(users)
	activeUsers := 0
	admins := 0
	disabledUsers := 0

	for _, user := range users {
		if user.Enabled {
			activeUsers++
		} else {
			disabledUsers++
		}
		if user.Role == "admin" || user.Role == "server_owner" {
			admins++
		}
	}

	return c.JSON(fiber.Map{
		"total_users":    totalUsers,
		"active_users":   activeUsers,
		"admins":         admins,
		"disabled_users": disabledUsers,
	})
}

// SearchUsers searches for users by username or email (admin only)
// GET /api/users/search?q=query&limit=10
func (h *UserHandler) SearchUsers(c *fiber.Ctx) error {
	query := c.Query("q", "")
	limit := c.QueryInt("limit", 10)

	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Search query is required",
		})
	}

	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	users, _, err := h.authService.ListUsersPaginated(1, limit, query, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search users",
		})
	}

	return c.JSON(fiber.Map{
		"users": users,
		"total": len(users),
	})
}
