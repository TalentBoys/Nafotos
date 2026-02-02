package database

// Migration from v3 to v4: Remove album_items table and optimize indexes
const migrationV3ToV4 = `
-- Drop album_items table (no longer needed, using dynamic queries instead)
DROP TABLE IF EXISTS album_items;

-- Add index for efficient album queries on file_folder_mappings
-- This index is crucial for LIKE queries with path_prefix
CREATE INDEX IF NOT EXISTS idx_file_folder_mappings_relative_path ON file_folder_mappings(relative_path);

-- Optimize existing index for album queries
CREATE INDEX IF NOT EXISTS idx_album_folders_album_folder ON album_folders(album_id, folder_id);
`
