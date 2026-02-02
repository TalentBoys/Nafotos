# AwesomeSharing Backend

[中文文档](./README-ZH_CN.md)

AwesomeSharing is a Go-based photo and file sharing system backend service that supports file management, albums, sharing links, permission management, and more.

## Tech Stack

- **Language**: Go 1.24.0
- **Web Framework**: Fiber v2
- **Database**: SQLite
- **Image Processing**: imaging, goexif
- **Encryption**: golang.org/x/crypto

## Project Structure

```
backend/
├── cmd/
│   └── server/          # Server entry point
│       └── main.go      # Main program
├── internal/
│   ├── api/             # API handlers and routes
│   │   ├── routes_v2.go         # V2 routes (with authentication)
│   │   ├── auth_handlers.go     # Authentication handlers
│   │   ├── user_handlers.go     # User management handlers
│   │   ├── folder_handlers.go   # Folder handlers
│   │   ├── album_handlers.go    # Album handlers
│   │   ├── share_handlers.go    # Share handlers
│   │   └── ...
│   ├── services/        # Business logic services
│   │   ├── auth.go              # Authentication service
│   │   ├── folder.go            # Folder service
│   │   ├── album.go             # Album service
│   │   ├── share.go             # Share service
│   │   ├── scanner.go           # File scanner service
│   │   ├── thumbnail.go         # Thumbnail service
│   │   └── ...
│   ├── middleware/      # Middleware
│   │   └── auth.go              # Authentication middleware
│   ├── models/          # Data models
│   │   └── models.go
│   ├── database/        # Database initialization and migration
│   │   ├── database.go
│   │   └── schema_v3.go
│   ├── config/          # Configuration management
│   │   └── config.go
│   └── initialization/  # Initialization logic
│       └── init.go
├── pkg/                 # Reusable packages
│   ├── utils/
│   └── exif/
├── config/              # Config directory (generated at runtime)
├── upload/              # Upload directory (generated at runtime)
├── go.mod               # Go module dependencies
└── run-local-v2.sh      # Local startup script

```

## Quick Start

### Prerequisites

- Go 1.24.0 or higher
- Git

### Install Dependencies

```bash
cd backend
go mod download
```

### Local Startup

#### Option 1: Using Startup Script (Recommended)

```bash
./run-local-v2.sh
```

The script automatically loads port configuration from `../.env.local` if it exists.

#### Option 2: Direct Run

```bash
# Set environment variables
export CONFIG_DIR="./config"
export UPLOAD_DIR="./upload"
export PORT="8080"
export ALLOWED_ORIGIN="http://localhost:3000"

# Start server
go run cmd/server/main.go
```

### Configuring Development Ports

For local development, you can configure custom ports by creating a `.env.local` file in the project root:

```bash
# .env.local
BACKEND_PORT=8080
FRONTEND_PORT=3000
```

The backend script (`run-local-v2.sh`) and frontend Vite configuration will automatically use these ports. This allows different developers to use different ports without modifying scripts or configuration files.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `CONFIG_DIR` | `/config` | Config directory path (stores database and thumbnails) |
| `UPLOAD_DIR` | `/upload` | Upload directory path |
| `ALLOWED_ORIGIN` | `*` | CORS allowed origin (recommend setting specific domain in production) |
| `DISABLE_FILE_VALIDATION` | `false` | Disable file validation (set to `true` to disable) |

### First Startup

On first startup, the server will automatically:

1. Create SQLite database (`config/awesome-sharing.db`)
2. Initialize database schema
3. Create default admin account:
   - Username: `admin`
   - Password: `admin`
4. Start background file scanner (scans every 30 minutes)
5. Start file validation service (cleans up invalid files every 6 hours)

### Verify Service

Access health check endpoint:

```bash
curl http://localhost:8080/api/health
```

Expected response:

```json
{"status":"ok"}
```

## API Documentation

### Public Endpoints (No Authentication Required)

#### Health Check

```
GET /api/health
```

#### Public Settings

```
GET /api/settings/public
```

#### Access Share Link

```
GET /api/s/:id
```

### Authentication Endpoints

#### Login

```
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin"
}
```

#### Register (Optional Restriction)

```
POST /api/auth/register
Content-Type: application/json

{
  "username": "newuser",
  "password": "password123",
  "email": "user@example.com"
}
```

#### Logout

```
POST /api/auth/logout
Authorization: Bearer <token>
```

#### Get Current User Info

```
GET /api/auth/me
Authorization: Bearer <token>
```

#### Change Password

```
POST /api/auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "oldpass",
  "new_password": "newpass"
}
```

### User Management Endpoints (Admin Only)

```
GET    /api/users                      # List users
GET    /api/users/search               # Search users
GET    /api/users/stats                # User statistics
POST   /api/users                      # Create user
GET    /api/users/:id                  # Get user details
PUT    /api/users/:id                  # Update user
DELETE /api/users/:id                  # Delete user
PUT    /api/users/:id/toggle           # Enable/disable user
POST   /api/users/:id/reset-password   # Reset password
GET    /api/users/:id/activity-logs    # User activity logs
POST   /api/users/export               # Export users
POST   /api/users/bulk/enable-disable  # Bulk enable/disable
POST   /api/users/bulk/delete          # Bulk delete
```

### Folder Management Endpoints

```
GET    /api/folders                # List folders
POST   /api/folders                # Create folder (admin)
GET    /api/folders/:id            # Get folder details
PUT    /api/folders/:id            # Update folder (admin)
DELETE /api/folders/:id            # Delete folder (admin)
PUT    /api/folders/:id/toggle     # Enable/disable folder (admin)
POST   /api/folders/:id/scan       # Scan folder (admin)
GET    /api/folders/:id/files      # List files in folder
```

### Permission Group Endpoints

```
GET    /api/permission-groups                        # List permission groups
POST   /api/permission-groups                        # Create permission group (admin)
GET    /api/permission-groups/:id                    # Get permission group details
PUT    /api/permission-groups/:id                    # Update permission group (admin)
DELETE /api/permission-groups/:id                    # Delete permission group (admin)
GET    /api/permission-groups/:id/folders            # List folders in permission group
POST   /api/permission-groups/:id/folders            # Add folder to permission group (admin)
DELETE /api/permission-groups/:id/folders/:folderId  # Remove folder from permission group (admin)
GET    /api/permission-groups/:id/permissions        # List permissions
POST   /api/permission-groups/:id/permissions        # Grant permission (admin)
DELETE /api/permission-groups/:id/permissions/:userId # Revoke permission (admin)
```

### Album Endpoints (V2)

```
GET    /api/albums-v2                 # List albums
POST   /api/albums-v2                 # Create album
GET    /api/albums-v2/:id             # Get album details
PUT    /api/albums-v2/:id             # Update album
DELETE /api/albums-v2/:id             # Delete album
GET    /api/albums-v2/:id/items       # List album items
POST   /api/albums-v2/:id/items       # Add items to album
DELETE /api/albums-v2/:id/items/:itemId # Remove item from album
POST   /api/albums-v2/:id/resolve     # Resolve album items (admin)
POST   /api/albums-v2/resolve-all     # Resolve all albums (admin)
```

### Share Endpoints

```
GET    /api/shares                         # List shares
POST   /api/shares                         # Create share
GET    /api/shares/:id                     # Get share details
PUT    /api/shares/:id                     # Update share
DELETE /api/shares/:id                     # Delete share
POST   /api/shares/:id/extend              # Extend share expiration
GET    /api/shares/:id/access-log          # Get share access log
POST   /api/shares/:id/permissions         # Grant share permission (private shares)
DELETE /api/shares/:id/permissions/:userId # Revoke share permission
DELETE /api/shares/expired                 # Delete expired shares
```

### File Endpoints (Legacy Compatibility)

```
GET /api/files                  # Get file list
GET /api/files/:id              # Get file details
GET /api/files/:id/thumbnail    # Get file thumbnail
GET /api/files/:id/download     # Download file
GET /api/timeline               # Timeline view
GET /api/search                 # Search files
```

### System Management Endpoints (Admin Only)

```
GET  /api/settings              # Get system settings
PUT  /api/settings              # Update system settings
GET  /api/settings/domain       # Get domain configuration
PUT  /api/settings/domain       # Update domain configuration
GET  /api/domain-config         # Get domain config
POST /api/domain-config         # Save domain config
POST /api/scan                  # Manually trigger scan
POST /api/cleanup               # Cleanup deleted files
```

### Other Endpoints

```
GET  /api/tags                  # Get tag list
POST /api/tags                  # Create tag
GET  /api/mount-points          # Get mount points
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests for Specific Package

```bash
go test ./internal/services
go test ./internal/api
```

### Run Tests with Coverage

```bash
go test -cover ./...
```

### Generate Coverage Report

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Development Guide

### Code Standards

- Use `gofmt` to format code
- Follow Go standard library naming conventions
- Add comments for exported functions and types

### Adding New API Endpoints

1. Create or update handler in `internal/api/`
2. Implement business logic in `internal/services/`
3. Register route in `internal/api/routes_v2.go`
4. If database changes needed, update `internal/database/schema_v*.go`

### Database Migration

Database migrations are executed automatically on server startup. To add new migration:

1. Create new schema file: `internal/database/schema_v<new_version>.go`
2. Register new migration version in `internal/database/database.go`
3. Restart server to apply migration

## Background Services

After server startup, the following background tasks run:

1. **File Scanner**: Scans configured folders every 30 minutes to discover new files
2. **File Validator**: Validates files in database every 6 hours, cleans up invalid records
3. **Session Cleanup**: Cleans up expired sessions every 1 hour

File validation can be disabled with environment variable `DISABLE_FILE_VALIDATION=true`.

## Troubleshooting

### Database Lock Error

If you encounter "database is locked" error:

1. Ensure only one server instance is running
2. Check if other programs are accessing the database file
3. Wait for background scan tasks to complete

### Cannot Create Thumbnails

Ensure:

1. `config/thumbs` directory exists and is writable
2. Image file format is supported (JPEG, PNG, HEIC, TIFF, etc.)
3. System has enough disk space

### Permission Issues

Ensure the user running the server has read/write permissions for:

- `CONFIG_DIR` (default: `./config`)
- `UPLOAD_DIR` (default: `./upload`)

## Security Recommendations

1. **Change Default Admin Password**: Change `admin` account password immediately after first login
2. **Configure CORS**: Set `ALLOWED_ORIGIN` to specific domain in production
3. **Use HTTPS**: Enable HTTPS via reverse proxy (e.g., Nginx) in production
4. **Regular Backups**: Regularly backup `config/awesome-sharing.db` database file
5. **File Permissions**: Ensure database and config files are not accessible by unauthorized users

## License

Please refer to the LICENSE file in the project root directory.

## Contributing

Issues and Pull Requests are welcome!

## Contact

For questions or suggestions, please contact via GitHub Issues.
