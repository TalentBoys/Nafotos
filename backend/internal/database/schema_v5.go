package database

const migrationV4ToV5 = `
-- ============================================================
-- Migration from Schema v4 to v5
-- Purpose: Separate photo metadata into dedicated table
-- ============================================================

BEGIN TRANSACTION;

-- Step 1: Create photo_metadata table
CREATE TABLE IF NOT EXISTS photo_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL UNIQUE,

    -- Core dimensions (migrated from files table)
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    taken_at DATETIME,

    -- Camera information
    make TEXT,
    model TEXT,

    -- GPS Location
    latitude REAL,
    longitude REAL,
    altitude REAL,

    -- Camera settings
    iso INTEGER,
    aperture REAL,
    shutter_speed TEXT,
    focal_length REAL,

    -- Orientation
    orientation INTEGER DEFAULT 1,

    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_photo_metadata_file_id ON photo_metadata(file_id);
CREATE INDEX IF NOT EXISTS idx_photo_metadata_taken_at ON photo_metadata(taken_at);
CREATE INDEX IF NOT EXISTS idx_photo_metadata_location ON photo_metadata(latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_photo_metadata_camera ON photo_metadata(make, model);

-- Step 2: Migrate existing data from files to photo_metadata
-- Only migrate image files (not videos)
INSERT INTO photo_metadata (file_id, width, height, taken_at, created_at)
SELECT id, width, height, taken_at, CURRENT_TIMESTAMP
FROM files
WHERE file_type = 'image'
AND (width > 0 OR height > 0 OR taken_at IS NOT NULL);

-- Step 3: Rename file_thumbnails to image_thumbnails
ALTER TABLE file_thumbnails RENAME TO image_thumbnails;

-- Step 4: Update indexes for image_thumbnails
DROP INDEX IF EXISTS idx_file_thumbnails_file;
DROP INDEX IF EXISTS idx_file_thumbnails_size_type;
CREATE INDEX IF NOT EXISTS idx_image_thumbnails_file ON image_thumbnails(file_id);
CREATE INDEX IF NOT EXISTS idx_image_thumbnails_size_type ON image_thumbnails(size_type);

-- Step 5: Create new files table without photo columns
CREATE TABLE files_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT NOT NULL,
    file_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_thumbnail BOOLEAN DEFAULT 0,
    parent_file_id INTEGER,
    FOREIGN KEY (parent_file_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Step 6: Copy data to new files table (excluding photo columns)
INSERT INTO files_new (id, filename, file_type, size, created_at, updated_at, is_thumbnail, parent_file_id)
SELECT id, filename, file_type, size, created_at, updated_at, is_thumbnail, parent_file_id
FROM files;

-- Step 7: Drop old files table and rename new one
DROP TABLE files;
ALTER TABLE files_new RENAME TO files;

-- Step 8: Recreate indexes on files table
CREATE INDEX IF NOT EXISTS idx_files_type ON files(file_type);
CREATE INDEX IF NOT EXISTS idx_files_is_thumbnail ON files(is_thumbnail);
CREATE INDEX IF NOT EXISTS idx_files_parent_file_id ON files(parent_file_id);

COMMIT;
`
