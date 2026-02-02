package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"awesome-sharing/internal/middleware"
	"awesome-sharing/internal/services"
)

type PermissionGroupHandler struct {
	permissionGroupService *services.PermissionGroupService
}

func NewPermissionGroupHandler(permissionGroupService *services.PermissionGroupService) *PermissionGroupHandler {
	return &PermissionGroupHandler{
		permissionGroupService: permissionGroupService,
	}
}

// CreatePermissionGroup creates a new permission group
// POST /api/permission-groups
func (h *PermissionGroupHandler) CreatePermissionGroup(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can create permission groups
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	group, err := h.permissionGroupService.CreatePermissionGroup(req.Name, req.Description, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create permission group",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"group": group,
	})
}

// ListPermissionGroups lists all permission groups accessible to the user
// GET /api/permission-groups
func (h *PermissionGroupHandler) ListPermissionGroups(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	isAdmin := user.Role == "admin" || user.Role == "server_owner"
	groups, err := h.permissionGroupService.ListPermissionGroups(user.ID, isAdmin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list permission groups",
		})
	}

	return c.JSON(fiber.Map{
		"groups": groups,
		"total":  len(groups),
	})
}

// GetPermissionGroup retrieves a specific permission group
// GET /api/permission-groups/:id
func (h *PermissionGroupHandler) GetPermissionGroup(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid permission group ID",
		})
	}

	group, err := h.permissionGroupService.GetPermissionGroup(id)
	if err != nil {
		if err == services.ErrPermissionGroupNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Permission group not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get permission group",
		})
	}

	return c.JSON(fiber.Map{
		"group": group,
	})
}

// UpdatePermissionGroup updates a permission group
// PUT /api/permission-groups/:id
func (h *PermissionGroupHandler) UpdatePermissionGroup(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can update permission groups
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid permission group ID",
		})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	err = h.permissionGroupService.UpdatePermissionGroup(id, req.Name, req.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update permission group",
		})
	}

	updatedGroup, err := h.permissionGroupService.GetPermissionGroup(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get updated permission group",
		})
	}

	return c.JSON(fiber.Map{
		"group": updatedGroup,
	})
}

// DeletePermissionGroup deletes a permission group
// DELETE /api/permission-groups/:id
func (h *PermissionGroupHandler) DeletePermissionGroup(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can delete permission groups
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid permission group ID",
		})
	}

	err = h.permissionGroupService.DeletePermissionGroup(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete permission group",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Permission group deleted successfully",
	})
}

// AddFolderToGroup adds a folder to a permission group
// POST /api/permission-groups/:id/folders
func (h *PermissionGroupHandler) AddFolderToGroup(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can modify permission groups
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	groupID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid permission group ID",
		})
	}

	var req struct {
		FolderID int64 `json:"folder_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = h.permissionGroupService.AddFolder(groupID, req.FolderID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add folder to group",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Folder added to group successfully",
	})
}

// RemoveFolderFromGroup removes a folder from a permission group
// DELETE /api/permission-groups/:id/folders/:folderId
func (h *PermissionGroupHandler) RemoveFolderFromGroup(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can modify permission groups
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	groupID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid permission group ID",
		})
	}

	folderID, err := strconv.ParseInt(c.Params("folderId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid folder ID",
		})
	}

	err = h.permissionGroupService.RemoveFolder(groupID, folderID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove folder from group",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Folder removed from group successfully",
	})
}

// ListFoldersInGroup lists all folders in a permission group
// GET /api/permission-groups/:id/folders
func (h *PermissionGroupHandler) ListFoldersInGroup(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	groupID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid permission group ID",
		})
	}

	folders, err := h.permissionGroupService.ListFoldersInGroup(groupID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list folders in group",
		})
	}

	// Transform folders to match frontend expectations
	type FolderResponse struct {
		ID        int64  `json:"id"`
		FolderID  int64  `json:"folder_id"`
		Name      string `json:"folder_name"`
		Path      string `json:"folder_path"`
	}

	folderResponses := make([]FolderResponse, len(folders))
	for i, folder := range folders {
		folderResponses[i] = FolderResponse{
			ID:       folder.ID,
			FolderID: folder.ID,
			Name:     folder.Name,
			Path:     folder.AbsolutePath,
		}
	}

	return c.JSON(fiber.Map{
		"folders": folderResponses,
		"total":   len(folderResponses),
	})
}

// GrantPermission grants a user permission to a permission group
// POST /api/permission-groups/:id/permissions
func (h *PermissionGroupHandler) GrantPermission(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can modify permissions
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	groupID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid permission group ID",
		})
	}

	var req struct {
		UserID     int64  `json:"user_id"`
		Permission string `json:"permission"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Permission != "read" && req.Permission != "write" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Permission must be 'read' or 'write'",
		})
	}

	err = h.permissionGroupService.GrantPermission(groupID, req.UserID, req.Permission)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to grant permission",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Permission granted successfully",
	})
}

// RevokePermission revokes a user's permission to a permission group
// DELETE /api/permission-groups/:id/permissions/:userId
func (h *PermissionGroupHandler) RevokePermission(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Only admins can modify permissions
	if user.Role != "admin" && user.Role != "server_owner" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin privileges required",
		})
	}

	groupID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid permission group ID",
		})
	}

	userID, err := strconv.ParseInt(c.Params("userId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	err = h.permissionGroupService.RevokePermission(groupID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke permission",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Permission revoked successfully",
	})
}

// ListPermissions lists all permissions for a permission group
// GET /api/permission-groups/:id/permissions
func (h *PermissionGroupHandler) ListPermissions(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	groupID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid permission group ID",
		})
	}

	permissions, err := h.permissionGroupService.ListUsersWithAccess(groupID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list permissions",
		})
	}

	// Transform permissions to match frontend expectations
	type PermissionResponse struct {
		ID       int64  `json:"id"`
		UserID   int64  `json:"user_id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Permission string `json:"permission"`
		GrantedAt  string `json:"granted_at"`
	}

	permissionResponses := make([]PermissionResponse, len(permissions))
	for i, perm := range permissions {
		permissionResponses[i] = PermissionResponse{
			ID:         perm.User.ID,
			UserID:     perm.User.ID,
			Username:   perm.User.Username,
			Email:      perm.User.Email,
			Permission: perm.Permission,
			GrantedAt:  perm.GrantedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return c.JSON(fiber.Map{
		"permissions": permissionResponses,
		"total":       len(permissionResponses),
	})
}
