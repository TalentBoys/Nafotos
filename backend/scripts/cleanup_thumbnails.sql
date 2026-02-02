-- Cleanup script to remove existing thumbnail entries from files table
-- This script identifies and removes files that are thumbnails based on their path

-- First, mark files in the thumbs directory as thumbnails
UPDATE files
SET is_thumbnail = 1
WHERE path LIKE '%/thumbs/%' OR path LIKE '%\\thumbs\\%';

-- Delete all files marked as thumbnails
-- (Alternatively, you can keep them marked if you want to preserve the records)
DELETE FROM files
WHERE is_thumbnail = 1;

-- Verify the cleanup
SELECT COUNT(*) as remaining_thumbnail_files
FROM files
WHERE is_thumbnail = 1;

-- Show sample of remaining files (should be original files only)
SELECT id, filename, path, file_type, is_thumbnail
FROM files
ORDER BY taken_at DESC
LIMIT 10;
