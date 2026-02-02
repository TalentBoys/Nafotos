package models

import "time"

// User represents a system user
type User struct {
	ID                int64      `json:"id"`
	Username          string     `json:"username"`
	PasswordHash      string     `json:"-"` // Never expose password hash
	Email             string     `json:"email,omitempty"`
	Role              string     `json:"role"` // 'server_owner', 'admin', or 'user'
	Enabled           bool       `json:"enabled"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	PasswordChangedAt *time.Time `json:"password_changed_at,omitempty"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// UserActivityLog represents an audit log entry for user management actions
type UserActivityLog struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`       // User being acted upon
	PerformedBy int64     `json:"performed_by"`  // User performing the action
	Action      string    `json:"action"`        // 'created', 'updated', 'deleted', 'password_reset', 'enabled', 'disabled'
	Details     string    `json:"details"`       // JSON metadata
	IPAddress   string    `json:"ip_address"`
	CreatedAt   time.Time `json:"created_at"`
}

// Folder represents a folder in the file system (文件夹)
type Folder struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	AbsolutePath string    `json:"absolute_path"`
	Enabled      bool      `json:"enabled"`
	CreatedBy    int64     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FileFolderMapping represents the mapping between files and folders (文件到文件夹的映射)
type FileFolderMapping struct {
	FileID       int64     `json:"file_id"`
	FolderID     int64     `json:"folder_id"`
	RelativePath string    `json:"relative_path"`
	CreatedAt    time.Time `json:"created_at"`
}

// PermissionGroup represents a group of folders for access control (权限组)
type PermissionGroup struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PermissionGroupFolder represents folders in a permission group (权限组包含的文件夹)
type PermissionGroupFolder struct {
	PermissionGroupID int64     `json:"permission_group_id"`
	FolderID          int64     `json:"folder_id"`
	AddedAt           time.Time `json:"added_at"`
}

// PermissionGroupPermission represents user permissions for a permission group (权限组的用户权限)
type PermissionGroupPermission struct {
	ID                int64     `json:"id"`
	PermissionGroupID int64     `json:"permission_group_id"`
	UserID            int64     `json:"user_id"`
	Permission        string    `json:"permission"` // 'read' or 'write'
	GrantedAt         time.Time `json:"granted_at"`
}

// File represents a file in the system (文件)
type File struct {
	ID            int64      `json:"id"`
	Filename      string     `json:"filename"`
	FileType      string     `json:"file_type"` // image, video
	Size          int64      `json:"size"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	IsThumbnail   bool       `json:"is_thumbnail"`
	ParentFileID  *int64     `json:"parent_file_id,omitempty"`
	ThumbnailURL  string     `json:"thumbnail_url,omitempty"`
	AbsolutePath  string     `json:"absolute_path,omitempty"` // Computed field, not stored in DB

	// Photo-specific fields (joined from photo_metadata table for images)
	// These fields will be populated via LEFT JOIN for backward compatibility in API responses
	Width         int        `json:"width,omitempty"`
	Height        int        `json:"height,omitempty"`
	TakenAt       *time.Time `json:"taken_at,omitempty"`
}

// PhotoMetadata represents photo-specific metadata extracted from EXIF
type PhotoMetadata struct {
	ID       int64     `json:"id"`
	FileID   int64     `json:"file_id"`

	// Dimensions
	Width    int       `json:"width"`
	Height   int       `json:"height"`

	// DateTime
	TakenAt  *time.Time `json:"taken_at,omitempty"`

	// Camera info
	Make     string    `json:"make,omitempty"`
	Model    string    `json:"model,omitempty"`

	// GPS location
	Latitude  *float64  `json:"latitude,omitempty"`
	Longitude *float64  `json:"longitude,omitempty"`
	Altitude  *float64  `json:"altitude,omitempty"`

	// Camera settings
	ISO          *int     `json:"iso,omitempty"`
	Aperture     *float64 `json:"aperture,omitempty"`
	ShutterSpeed string   `json:"shutter_speed,omitempty"`
	FocalLength  *float64 `json:"focal_length,omitempty"`

	// Orientation
	Orientation int       `json:"orientation"`

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ImageThumbnail represents a generated thumbnail for an image
type ImageThumbnail struct {
	ID        int64     `json:"id"`
	FileID    int64     `json:"file_id"`
	SizeType  string    `json:"size_type"` // small, medium, large
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	FileSize  int64     `json:"file_size"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// Album represents a collection of files with soft links
type Album struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	OwnerID     int64     `json:"owner_id"`
	CoverFileID *int64    `json:"cover_file_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AlbumItem represents a soft link to a file via folder + relative path
type AlbumItem struct {
	ID           int64     `json:"id"`
	AlbumID      int64     `json:"album_id"`
	FolderID     int64     `json:"folder_id"`
	RelativePath string    `json:"relative_path"`
	FileID       *int64    `json:"file_id,omitempty"` // Current resolved file
	AddedAt      time.Time `json:"added_at"`
}

// AlbumFolder represents a folder configuration for an album
type AlbumFolder struct {
	ID         int64     `json:"id"`
	AlbumID    int64     `json:"album_id"`
	FolderID   int64     `json:"folder_id"`
	PathPrefix string    `json:"path_prefix"` // e.g., "2024/", "vacation/", or "" for entire folder
	AddedAt    time.Time `json:"added_at"`
}

// Tag represents a label for files
type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// FileTag represents the many-to-many relationship between files and tags
type FileTag struct {
	FileID int64 `json:"file_id"`
	TagID  int64 `json:"tag_id"`
}

// SystemSetting represents a system configuration setting
type SystemSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Share represents a shareable link
type Share struct {
	ID           string     `json:"id"` // Short ID
	ShareType    string     `json:"share_type"` // 'file' or 'album'
	ResourceID   int64      `json:"resource_id"`
	OwnerID      int64      `json:"owner_id"`
	AccessType   string     `json:"access_type"` // 'public' or 'private'
	PasswordHash string     `json:"-"` // Optional password (not exposed to frontend)
	HasPassword  bool       `json:"has_password"` // Whether password is set (for frontend display)
	RequiresAuth bool       `json:"requires_auth"` // Whether authentication is required
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	MaxViews     *int       `json:"max_views,omitempty"`
	ViewCount    int        `json:"view_count"`
	Enabled      bool       `json:"enabled"`
	CreatedAt    time.Time  `json:"created_at"`
}

// SharePermission represents user access to a private share
type SharePermission struct {
	ID        int64     `json:"id"`
	ShareID   string    `json:"share_id"`
	UserID    int64     `json:"user_id"`
	GrantedAt time.Time `json:"granted_at"`
}

// ShareAccessLog represents an access log entry for a share
type ShareAccessLog struct {
	ID         int64      `json:"id"`
	ShareID    string     `json:"share_id"`
	AccessedBy *int64     `json:"accessed_by,omitempty"` // NULL for anonymous
	IPAddress  string     `json:"ip_address,omitempty"`
	UserAgent  string     `json:"user_agent,omitempty"`
	AccessedAt time.Time  `json:"accessed_at"`
}

// DomainConfig represents the domain configuration for generating share links
type DomainConfig struct {
	ID        int64      `json:"id"`
	Protocol  string     `json:"protocol"`  // http or https
	Domain    string     `json:"domain"`    // example.com or IP address
	Port      string     `json:"port"`      // 80, 443, 8080, etc.
	UpdatedBy *int64     `json:"updated_by,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

