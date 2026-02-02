# AwesomeSharing - Implementation Summary

[中文文档](./README-ZH_CN.md)

## Project Overview

AwesomeSharing is a self-hosted photo and file management system that supports multi-directory management, permission control, album organization, and file sharing features.

## Tech Stack

### Backend
- **Language**: Go 1.24.0
- **Web Framework**: Fiber v2
- **Database**: SQLite
- **Image Processing**: disintegration/imaging
- **EXIF Extraction**: rwcarlsen/goexif
- **Password Encryption**: bcrypt (golang.org/x/crypto)

### Frontend
- **Framework**: React 18 + TypeScript
- **Styling**: Tailwind CSS + Shadcn UI
- **Routing**: React Router v6
- **Internationalization**: react-i18next
- **Build Tool**: Vite

## Core Features Implemented ✅

### 1. User Management and Permission Control
**Location**: `backend/internal/services/auth.go`, `backend/internal/api/auth_handlers.go`, `backend/internal/api/user_handlers.go`

**Implemented Features**:
- User registration and login (with configurable registration toggle)
- Session-based authentication (7-day validity)
- Three user roles: `server_owner` (super admin), `admin` (administrator), `user` (regular user)
- User CRUD operations (create, query, update, delete)
- User enable/disable functionality
- Password change and reset
- Bulk user operations (bulk enable/disable, bulk delete)
- User activity logs (recording user management actions)
- User statistics and search functionality
- User export functionality

**Database Tables**:
- `users` - User information
- `sessions` - Session management
- `user_activity_logs` - User activity audit logs

**Frontend Implementation**:
- Login page: `frontend/src/pages/Login.tsx`
- User management page: `frontend/src/pages/UserManagement.tsx`
- Auth context: `frontend/src/contexts/AuthContext.tsx`
- Route protection: `frontend/src/components/ProtectedRoute.tsx`

### 2. Folder System
**Location**: `backend/internal/services/folder.go`, `backend/internal/api/folder_handlers.go`

**Core Features**:
- Admins can create folders pointing to any absolute path in the filesystem
- Each folder contains: name, absolute path, enabled status, creator
- Folders can be enabled/disabled
- Support for manually triggering folder scans
- Query file lists within folders (with permission filtering)

**Database Tables**:
- `folders` - Folder metadata
- `file_folder_mappings` - File-to-folder mapping relationships (storing relative paths)

**Design Highlights**:
- Users can configure any path in the system, not limited to `/config` or `/upload`
- Files are stored by relative path, allowing re-association even when folders move

**Frontend Implementation**:
- Folder management page: `frontend/src/pages/FolderManagement.tsx`
- Folder browsing page: `frontend/src/pages/Folders.tsx`

### 3. Permission Group System
**Location**: `backend/internal/services/permission_group.go`, `backend/internal/api/permission_group_handlers.go`

**Core Functionality**:
- Create permission groups (containing multiple folders)
- Assign users to permission groups with permissions (`read` or `write`)
- Manage folders within permission groups (add/remove)
- Query user's permission groups and access rights

**Permission Logic**:
- `admin` and `server_owner` roles automatically have all permissions
- Regular users gain access to specific folders through permission groups
- Supports read permission (view only) and write permission (can modify)

**Database Tables**:
- `permission_groups` - Permission group metadata
- `permission_group_folders` - Folders contained in permission groups
- `permission_group_permissions` - User permissions for permission groups

**Frontend Implementation**:
- Permission group management page: `frontend/src/pages/PermissionGroupManagement.tsx`

### 4. Enhanced Albums (V2) System
**Location**: `backend/internal/services/album.go`, `backend/internal/api/album_handlers.go`

**Core Innovation - Soft Links**:
- Album items are stored as `(folder_id, relative_path)` instead of directly storing `file_id`
- When folder paths change, update the folder configuration and rescan
- Albums automatically resolve files through relative paths, avoiding album invalidation

**Implemented Features**:
- Create/update/delete albums
- Add/remove album items (via folder_id + relative_path)
- Query albums and all their items
- Set album covers
- Resolve album items (re-associate files after moving)
- Batch resolve all albums

**Database Tables**:
- `albums_v2` - Album metadata (includes owner_id)
- `album_items` - Album items (storing folder_id + relative_path + current file_id)

**Usage Flow Example**:
```
1. Folder X points to path: /photos/vacation
2. User adds /photos/vacation/beach.jpg to Album A
3. System stores: (folder_id=X, relative_path="vacation/beach.jpg")
4. Admin moves folder: /photos/vacation → /photos/2024-vacation
5. Admin updates Folder X path to /photos/2024-vacation
6. System rescans
7. Album A automatically resolves, beach.jpg still available!
```

**Frontend Implementation**:
- Albums page: `frontend/src/pages/Albums.tsx`

### 5. File Sharing System
**Location**: `backend/internal/services/share.go`, `backend/internal/api/share_handlers.go`

**Share Types**:
- File sharing (single file)
- Album sharing (entire album)

**Access Control**:
- **Public shares**: Anyone can access via link (anonymous)
- **Private shares**: Only specified users can access (requires login)

**Advanced Features**:
- Password protection (optional)
- Expiration settings:
  - By hours (e.g., 24 hours)
  - By days (e.g., 7 days)
  - Permanent (no expiration)
- Maximum view count limit
- Access counter
- Enable/disable shares
- Access logs (recording visitors, IP, UserAgent, access time)

**Management Features**:
- View all your shares
- View share access logs
- Extend share expiration time
- Bulk delete expired shares
- Grant/revoke user permissions for private shares

**Share Link Format**: `/api/s/:shareId`

**Database Tables**:
- `shares` - Share metadata and settings
- `share_permissions` - User permissions for private shares
- `share_access_log` - Access audit logs

### 6. Domain Configuration
**Location**: `backend/internal/services/domain_config.go`, `backend/internal/api/domain_config_handlers.go`

**Functionality**:
- Configure system access domain/IP (e.g., `qjkobe.online:1234`)
- Separately configure protocol (http/https), domain, and port
- Used for generating complete share links
- Admin-only functionality

**Database Table**:
- `domain_config` - Domain configuration

**Frontend Implementation**:
- Domain config page: `frontend/src/pages/DomainConfig.tsx`

### 7. System Settings
**Location**: `backend/internal/services/settings.go`, `backend/internal/api/settings_handlers.go`

**Configurable Items**:
- Site name
- Registration toggle (whether to allow new user registration)
- Other system-level configurations (Key-Value storage)

**Database Table**:
- `system_settings` - System configuration (Key-Value)

**Frontend Implementation**:
- Settings page: `frontend/src/pages/Settings.tsx`

### 8. File Scanning and Thumbnail Generation
**Location**: `backend/internal/services/scanner.go`, `backend/internal/services/thumbnail.go`

**Scanning Service**:
- Automatically scan all enabled folders
- Extract file metadata (EXIF, capture date, dimensions, etc.)
- Establish file-to-folder mapping relationships
- Scheduled scanning (every 30 minutes)
- Support for manually triggering scans

**Thumbnail Generation**:
- Automatically generate thumbnails for images (multiple sizes)
- Supported formats: JPEG, PNG, HEIC, TIFF, etc.
- Thumbnails stored in `config/thumbs/` directory
- Lazy loading generation (generated on first access)

**File Validation Service**:
- Periodically validate whether files in the database still exist in the filesystem
- Clean up invalid file records
- Scheduled runs (every 6 hours)
- Can be disabled via environment variable `DISABLE_FILE_VALIDATION=true`

**Database Tables**:
- `files` - File metadata
- `file_thumbnails` - Thumbnail information

### 9. Authentication Middleware
**Location**: `backend/internal/middleware/auth.go`

**Middleware Types**:
- `AuthMiddleware` - Requires valid session (enforces login)
- `OptionalAuthMiddleware` - Optional authentication (injects user info but doesn't enforce login)
- `AdminOnlyMiddleware` - Only allows admin and server_owner access
- `AdminOrOwnerMiddleware` - Only allows admin and server_owner access

**Session Sources**:
- session_id in Cookie
- Bearer token in Authorization header

### 10. Tags System
**Location**: `backend/internal/api/handlers.go` (to be enhanced)

**Basic Functionality**:
- Create tags (with colors)
- Add tags to files
- Query tag list

**Database Tables**:
- `tags` - Tag metadata
- `file_tags` - Many-to-many relationship between files and tags

## Database Architecture (Schema V3)

**Core Table Structure**:
```
users                        # Users
sessions                     # Sessions
user_activity_logs           # User activity logs
files                        # File metadata
folders                      # Folder configuration
file_folder_mappings         # File-folder mappings (relative paths)
permission_groups            # Permission groups
permission_group_folders     # Permission group-folder associations
permission_group_permissions # Permission group-user permissions
albums_v2                    # Albums (V2 version)
album_items                  # Album items (soft links)
tags                         # Tags
file_tags                    # File-tag associations
shares                       # Shares
share_permissions            # Private share permissions
share_access_log             # Share access logs
system_settings              # System settings
domain_config                # Domain configuration
file_thumbnails              # Thumbnails
```

**Database Features**:
- Foreign key constraints with cascade deletes
- Performance-optimized indexes
- Automatic timestamps
- Data integrity guarantees

## API Route Architecture

**Public Routes (No Authentication Required)**:
- `GET /api/health` - Health check
- `GET /api/settings/public` - Public settings
- `GET /api/s/:id` - Access share link

**Authentication Routes**:
- `POST /api/auth/login` - Login
- `POST /api/auth/register` - Register
- `POST /api/auth/logout` - Logout
- `GET /api/auth/me` - Get current user info
- `POST /api/auth/change-password` - Change password

**Protected Routes (Authentication Required)**:
- `/api/users/*` - User management (admin only)
- `/api/folders/*` - Folder management
- `/api/permission-groups/*` - Permission group management
- `/api/albums-v2/*` - Album management (V2)
- `/api/shares/*` - Share management
- `/api/settings/*` - System settings (admin only)
- `/api/domain-config/*` - Domain configuration (admin only)
- `/api/files/*` - File access (backward compatibility)
- `/api/timeline` - Timeline view
- `/api/search` - File search
- `/api/scan` - Trigger scan
- `/api/cleanup` - Clean up invalid files
- `/api/tags/*` - Tag management

## Project Structure

```
AwesomeSharing/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Server entry point
│   ├── internal/
│   │   ├── api/
│   │   │   ├── routes_v2.go         # Route configuration (V2)
│   │   │   ├── auth_handlers.go     # Auth handlers
│   │   │   ├── user_handlers.go     # User management handlers
│   │   │   ├── folder_handlers.go   # Folder handlers
│   │   │   ├── permission_group_handlers.go  # Permission group handlers
│   │   │   ├── album_handlers.go    # Album handlers
│   │   │   ├── share_handlers.go    # Share handlers
│   │   │   ├── settings_handlers.go # Settings handlers
│   │   │   ├── domain_config_handlers.go # Domain config handlers
│   │   │   └── handlers.go          # Other handlers
│   │   ├── services/
│   │   │   ├── auth.go              # Auth service
│   │   │   ├── folder.go            # Folder service
│   │   │   ├── permission_group.go  # Permission group service
│   │   │   ├── album.go             # Album service
│   │   │   ├── share.go             # Share service
│   │   │   ├── settings.go          # Settings service
│   │   │   ├── domain_config.go     # Domain config service
│   │   │   ├── scanner.go           # File scanner service
│   │   │   ├── thumbnail.go         # Thumbnail service
│   │   │   └── file_validator.go    # File validation service
│   │   ├── middleware/
│   │   │   └── auth.go              # Auth middleware
│   │   ├── models/
│   │   │   └── models.go            # Data models
│   │   ├── database/
│   │   │   ├── database.go          # Database initialization
│   │   │   └── schema_v3.go         # Database schema V3
│   │   ├── config/
│   │   │   └── config.go            # Configuration management
│   │   └── initialization/
│   │       └── init.go              # System initialization
│   ├── pkg/
│   │   ├── utils/                   # Utility functions
│   │   └── exif/                    # EXIF utilities
│   ├── go.mod
│   └── run-local-v2.sh              # Local startup script
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Layout.tsx           # Layout component
│   │   │   ├── ProtectedRoute.tsx   # Route protection
│   │   │   ├── Modal.tsx            # Modal
│   │   │   └── FileGrid.tsx         # File grid
│   │   ├── pages/
│   │   │   ├── Login.tsx            # Login page
│   │   │   ├── UserManagement.tsx   # User management
│   │   │   ├── FolderManagement.tsx # Folder management
│   │   │   ├── Folders.tsx          # Folder browsing
│   │   │   ├── PermissionGroupManagement.tsx # Permission group management
│   │   │   ├── Albums.tsx           # Albums
│   │   │   ├── Timeline.tsx         # Timeline
│   │   │   ├── Settings.tsx         # Settings
│   │   │   ├── DomainConfig.tsx     # Domain config
│   │   │   └── FileDetail.tsx       # File details
│   │   ├── contexts/
│   │   │   └── AuthContext.tsx      # Auth context
│   │   ├── locales/                 # i18n translations
│   │   ├── types/                   # TypeScript types
│   │   └── App.tsx
│   ├── package.json
│   └── vite.config.ts
├── config/                          # Config directory (generated at runtime)
│   ├── awesome-sharing.db           # SQLite database
│   └── thumbs/                      # Thumbnails directory
├── upload/                          # Upload directory (generated at runtime)
├── .env.example
└── README.md
```

## First-Time Startup Process

1. Server automatically executes on startup:
   - Create SQLite database file
   - Initialize database schema (Schema V3)
   - Create default admin account:
     - Username: `admin`
     - Password: `admin`
     - Role: `server_owner`
   - Start background file scanning service (every 30 minutes)
   - Start file validation service (every 6 hours)

2. After admin login:
   - Change default password (security recommendation)
   - Create folders (pointing to filesystem paths)
   - Create permission groups and assign folders
   - Create regular users and grant permissions
   - Configure domain (for share links)
   - Configure system settings

3. Regular user workflow:
   - Login to system
   - Browse folders with permissions
   - Create albums and add files
   - Create share links (public or private)
   - Manage own shares

## Key Design Decisions

### 1. Folder Replaces Library
- **Original design**: Each Library can contain multiple paths
- **Current design**: Each Folder corresponds to one absolute path
- **Advantage**: Simpler and more direct, each path managed independently

### 2. Soft Links Solve File Moving Problem
Traditional systems break albums when files move. This system uses `(folder_id, relative_path)` to implement soft links:
- Album items don't directly bind to file IDs
- After folder path changes, rescanning restores associations
- Significantly improves file management flexibility

### 3. Three-Layer Permission Control
- **Role layer**: server_owner > admin > user
- **Permission group layer**: Control folder access through permission groups
- **Share layer**: Public shares (anonymous) vs Private shares (specific users)

### 4. Session Authentication (Not JWT)
- Simpler implementation
- Easier to revoke (just delete session record)
- More suitable for small self-hosted applications

### 5. Complete Sharing Features
- Expiration control prevents permanent link leaks
- View count limits prevent abuse
- Access logs for auditing
- Password protection enhances security
- Private shares enable team collaboration

## Local Development Setup

### Configuring Development Ports

For local development, you can configure custom ports by creating a `.env.local` file in the project root:

```bash
# .env.local
BACKEND_PORT=8080
FRONTEND_PORT=3000
```

Both the backend script (`backend/run-local-v2.sh`) and frontend Vite configuration will automatically use these ports. This allows different developers to use different ports without modifying scripts or configuration files.

**To start development:**

```bash
# Terminal 1 - Start Backend
cd backend
./run-local-v2.sh

# Terminal 2 - Start Frontend
cd frontend
npm run dev
```

Both services will use the ports configured in `.env.local` (or default to 8080/3000 if not configured).

## Environment Variables Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port (or set `BACKEND_PORT` in `.env.local` for local development) |
| `CONFIG_DIR` | `/config` | Config directory path (stores database and thumbnails) |
| `UPLOAD_DIR` | `/upload` | Upload directory path |
| `ALLOWED_ORIGIN` | `*` | CORS allowed origin (recommend setting specific domain in production) |
| `DISABLE_FILE_VALIDATION` | `false` | Disable file validation (set to `true` to disable) |
| `BACKEND_PORT` | `8080` | Local development backend port (set in `.env.local`) |
| `FRONTEND_PORT` | `3000` | Local development frontend port (set in `.env.local`) |

## Background Services

After server startup, the following background tasks run automatically:

1. **File Scanner**: Scans all enabled folders every 30 minutes to discover new files
2. **File Validator**: Validates files in database every 6 hours, cleans up invalid records
3. **Session Cleanup**: Cleans up expired sessions every 1 hour

## Implementation Status

✅ **Backend Completion: ~95%**
- All core services implemented
- All API handlers implemented
- Complete database architecture
- Complete middleware
- Complete background tasks

✅ **Frontend Completion: ~80%**
- Complete authentication system (login, route protection, context)
- Complete user management page
- Complete folder management page
- Complete permission group management page
- Complete basic album functionality
- Complete domain config page
- Complete settings page

⚠️ **Features To Be Enhanced**:
- Share management frontend page (currently backend API only)
- Enhanced file detail page features
- Drag-and-drop sorting for album items
- Bulk file operations
- Advanced search filtering
- User avatar upload

## Security Recommendations

1. **Change Default Password**: Immediately change `admin` account password after first login
2. **Configure CORS**: Set `ALLOWED_ORIGIN` to specific domain in production
3. **Enable HTTPS**: Enable HTTPS via reverse proxy (e.g., Nginx) in production
4. **Regular Backups**: Regularly backup `config/awesome-sharing.db` database file
5. **File Permissions**: Ensure database and config files are not accessible by unauthorized users

## Technical Features

- **bcrypt** password hashing (secure and slow, prevents brute force)
- **Session authentication** (simple and easy to revoke)
- **SQLite** database (no separate database server needed)
- **Fiber** framework (fast, Express-style)
- **React + TypeScript** (type-safe frontend)
- **Responsive design** (supports desktop and mobile)
- **Internationalization support** (Chinese/English)

## Documentation

- Complete backend API documentation: `backend/README.md`
- Backend Chinese documentation: `backend/README-ZH_CN.md`
- Frontend documentation: `frontend/README.md`
- Frontend Chinese documentation: `frontend/README-ZH_CN.md`

## Usage Examples

### Admin Initial Configuration
```bash
1. Start server
2. Login with admin/admin
3. Change admin password
4. Create folders (e.g., /photos/vacation)
5. Create permission groups (e.g., Family Photos)
6. Add folders to permission groups
7. Create regular users
8. Grant permission group access to users
9. Configure domain (for share links)
```

### Regular User Usage
```bash
1. Login to system
2. Browse folders with permissions
3. Create albums, add favorite photos
4. Create share links (set expiration, password, etc.)
5. Send share links to friends
6. View access logs to see who accessed shares
```

### File Moving Scenario
```bash
1. Admin moves folder in filesystem: /photos/2023 → /photos/archived/2023
2. Admin updates Folder config: change path to /photos/archived/2023
3. Trigger scan
4. All albums automatically resolve, photo associations restored
5. Users unaware, albums work normally
```

## Project Status

🎉 **All Core Features Implemented and Functional!**

Current system capabilities:
- Complete user management and permission control
- Multi-folder management and scanning
- Permission groups and granular access control
- Album creation and management (with soft links)
- Share link generation and management
- Domain configuration and system settings
- Automatic thumbnail generation
- File validation and cleanup
- Complete audit logs

## License

MIT License
