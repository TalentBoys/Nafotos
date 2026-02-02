package database

const schemaV3 = `
-- Users and Authentication
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    email TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    enabled BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login_at DATETIME,
    password_changed_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- User Activity Logs
CREATE TABLE IF NOT EXISTS user_activity_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    performed_by INTEGER NOT NULL,
    action TEXT NOT NULL,
    details TEXT,
    ip_address TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (performed_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_activity_user ON user_activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_performed ON user_activity_logs(performed_by);
CREATE INDEX IF NOT EXISTS idx_activity_time ON user_activity_logs(created_at);

-- Files (文件)
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT NOT NULL,
    file_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    width INTEGER DEFAULT 0,
    height INTEGER DEFAULT 0,
    taken_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_thumbnail BOOLEAN DEFAULT 0,
    parent_file_id INTEGER,
    FOREIGN KEY (parent_file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_files_taken_at ON files(taken_at);
CREATE INDEX IF NOT EXISTS idx_files_type ON files(file_type);
CREATE INDEX IF NOT EXISTS idx_files_is_thumbnail ON files(is_thumbnail);
CREATE INDEX IF NOT EXISTS idx_files_parent_file_id ON files(parent_file_id);

-- Folders (文件夹)
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    absolute_path TEXT NOT NULL UNIQUE,
    enabled BOOLEAN DEFAULT 1,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_folders_created_by ON folders(created_by);
CREATE INDEX IF NOT EXISTS idx_folders_enabled ON folders(enabled);
CREATE INDEX IF NOT EXISTS idx_folders_absolute_path ON folders(absolute_path);

-- File to Folder Mappings (文件到文件夹的映射)
CREATE TABLE IF NOT EXISTS file_folder_mappings (
    file_id INTEGER NOT NULL,
    folder_id INTEGER NOT NULL,
    relative_path TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (file_id, folder_id),
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_file_folder_mappings_file ON file_folder_mappings(file_id);
CREATE INDEX IF NOT EXISTS idx_file_folder_mappings_folder ON file_folder_mappings(folder_id);
CREATE INDEX IF NOT EXISTS idx_file_folder_mappings_folder_path ON file_folder_mappings(folder_id, relative_path);
CREATE INDEX IF NOT EXISTS idx_file_folder_mappings_relative_path ON file_folder_mappings(relative_path);

-- Permission Groups (权限组)
CREATE TABLE IF NOT EXISTS permission_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_permission_groups_created_by ON permission_groups(created_by);

-- Permission Group Folders (权限组包含的文件夹)
CREATE TABLE IF NOT EXISTS permission_group_folders (
    permission_group_id INTEGER NOT NULL,
    folder_id INTEGER NOT NULL,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (permission_group_id, folder_id),
    FOREIGN KEY (permission_group_id) REFERENCES permission_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_permission_group_folders_group ON permission_group_folders(permission_group_id);
CREATE INDEX IF NOT EXISTS idx_permission_group_folders_folder ON permission_group_folders(folder_id);

-- Permission Group Permissions (权限组的用户权限)
CREATE TABLE IF NOT EXISTS permission_group_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    permission_group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    permission TEXT NOT NULL DEFAULT 'read',
    granted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (permission_group_id) REFERENCES permission_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(permission_group_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_permission_group_perms_group ON permission_group_permissions(permission_group_id);
CREATE INDEX IF NOT EXISTS idx_permission_group_perms_user ON permission_group_permissions(user_id);

-- Albums (相册)
CREATE TABLE IF NOT EXISTS albums_v2 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    owner_id INTEGER NOT NULL,
    cover_file_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (cover_file_id) REFERENCES files(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_albums_v2_owner ON albums_v2(owner_id);

-- Album Folders (相册文件夹配置 - 通过文件夹和路径前缀构建相册)
CREATE TABLE IF NOT EXISTS album_folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    album_id INTEGER NOT NULL,
    folder_id INTEGER NOT NULL,
    path_prefix TEXT NOT NULL DEFAULT '',
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (album_id) REFERENCES albums_v2(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE,
    UNIQUE(album_id, folder_id, path_prefix)
);

CREATE INDEX IF NOT EXISTS idx_album_folders_album ON album_folders(album_id);
CREATE INDEX IF NOT EXISTS idx_album_folders_folder ON album_folders(folder_id);
CREATE INDEX IF NOT EXISTS idx_album_folders_album_folder ON album_folders(album_id, folder_id);

-- Tags (标签)
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    color TEXT DEFAULT '#3b82f6',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS file_tags (
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (file_id, tag_id),
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- System Settings
CREATE TABLE IF NOT EXISTS system_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Domain Configuration (域名配置)
CREATE TABLE IF NOT EXISTS domain_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    protocol TEXT NOT NULL DEFAULT 'http',
    domain TEXT NOT NULL,
    port TEXT NOT NULL DEFAULT '80',
    updated_by INTEGER,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_domain_config_updated_by ON domain_config(updated_by);

-- Shares (分享)
CREATE TABLE IF NOT EXISTS shares (
    id TEXT PRIMARY KEY,
    share_type TEXT NOT NULL,
    resource_id INTEGER NOT NULL,
    owner_id INTEGER NOT NULL,
    access_type TEXT NOT NULL DEFAULT 'public',
    password_hash TEXT,
    requires_auth BOOLEAN DEFAULT 0,
    expires_at DATETIME,
    max_views INTEGER,
    view_count INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_shares_owner ON shares(owner_id);
CREATE INDEX IF NOT EXISTS idx_shares_expires ON shares(expires_at);
CREATE INDEX IF NOT EXISTS idx_shares_type_resource ON shares(share_type, resource_id);

CREATE TABLE IF NOT EXISTS share_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    share_id TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    granted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (share_id) REFERENCES shares(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(share_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_share_perms_share ON share_permissions(share_id);
CREATE INDEX IF NOT EXISTS idx_share_perms_user ON share_permissions(user_id);

CREATE TABLE IF NOT EXISTS share_access_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    share_id TEXT NOT NULL,
    accessed_by INTEGER,
    ip_address TEXT,
    user_agent TEXT,
    accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (share_id) REFERENCES shares(id) ON DELETE CASCADE,
    FOREIGN KEY (accessed_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_share_access_share ON share_access_log(share_id);
CREATE INDEX IF NOT EXISTS idx_share_access_time ON share_access_log(accessed_at);

-- File Thumbnails (文件缩略图)
CREATE TABLE IF NOT EXISTS file_thumbnails (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    size_type TEXT NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    file_size INTEGER NOT NULL,
    path TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    UNIQUE(file_id, size_type)
);

CREATE INDEX IF NOT EXISTS idx_file_thumbnails_file ON file_thumbnails(file_id);
CREATE INDEX IF NOT EXISTS idx_file_thumbnails_size_type ON file_thumbnails(size_type);
`
