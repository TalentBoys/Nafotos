package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

// Initialize creates database connection and tables
func Initialize(dbPath string) (*DB, error) {
	// Add connection parameters for better concurrency handling
	dbPath = dbPath + "?_busy_timeout=5000&_journal_mode=WAL"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	// With WAL mode, SQLite can handle multiple concurrent readers and one writer
	// Increase connection pool to allow concurrent read operations
	db.SetMaxOpenConns(10) // Allow up to 10 concurrent connections (WAL mode supports this)
	db.SetMaxIdleConns(2)  // Keep 2 idle connections ready
	db.SetConnMaxLifetime(0) // Connections never expire

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

	database := &DB{db}

	// Create legacy tables first for backwards compatibility
	if err := database.createLegacyTables(); err != nil {
		return nil, err
	}

	// Run migrations for new features
	if err := database.runMigrations(); err != nil {
		return nil, err
	}

	return database, nil
}

func (db *DB) createLegacyTables() error {
	// This function is kept for reference but no longer used
	// The new schema v3 is applied directly
	return nil
}

func (db *DB) runMigrations() error {
	// Check current schema version
	currentVersion := db.getSchemaVersion()
	targetVersion := 5

	if currentVersion >= targetVersion {
		log.Printf("Database is already at version %d, skipping migration", currentVersion)
		// Ensure domain_config table exists (added after v3)
		db.Exec(`CREATE TABLE IF NOT EXISTS domain_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			protocol TEXT NOT NULL DEFAULT 'http',
			domain TEXT NOT NULL,
			port TEXT NOT NULL DEFAULT '80',
			updated_by INTEGER,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
		)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_domain_config_updated_by ON domain_config(updated_by)`)
		log.Println("✓ Ensured domain_config table exists")

		// Check if requires_auth column exists in shares table
		var columnExists int
		err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('shares') WHERE name='requires_auth'`).Scan(&columnExists)
		if err == nil && columnExists == 0 {
			log.Println("Adding requires_auth column to shares table...")
			_, err := db.Exec(`ALTER TABLE shares ADD COLUMN requires_auth BOOLEAN DEFAULT 0`)
			if err != nil {
				log.Printf("Warning: Failed to add requires_auth column: %v", err)
			} else {
				log.Println("✓ Added requires_auth column to shares table")
			}
		}

		return nil
	}

	// If database is at v3 or v4, run v5 migration
	if currentVersion >= 3 && currentVersion < 5 {
		log.Println("Running migration from v3/v4 to v5...")
		if _, err := db.Exec(migrationV4ToV5); err != nil {
			log.Printf("Error running migration to schema v5: %v", err)
			return err
		}
		db.setSchemaVersion(5)
		log.Println("✓ Migration to v5 completed successfully")
		return nil
	}

	log.Printf("Migrating database from version %d to version %d...", currentVersion, targetVersion)

	// Temporarily disable foreign key constraints for migration
	log.Println("Disabling foreign key constraints...")
	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return err
	}

	// Drop old deprecated tables
	log.Println("Dropping deprecated tables...")
	db.Exec("DROP TABLE IF EXISTS file_library_mappings")
	db.Exec("DROP TABLE IF EXISTS library_permissions")
	db.Exec("DROP TABLE IF EXISTS library_paths")
	db.Exec("DROP TABLE IF EXISTS file_libraries")
	db.Exec("DROP TABLE IF EXISTS mount_points")
	db.Exec("DROP TABLE IF EXISTS album_files")
	db.Exec("DROP TABLE IF EXISTS albums")

	// Drop tables that need restructuring (in reverse order of dependencies)
	log.Println("Dropping tables for restructuring...")
	db.Exec("DROP TABLE IF EXISTS file_tags")
	db.Exec("DROP TABLE IF EXISTS file_thumbnails")
	db.Exec("DROP TABLE IF EXISTS album_items")
	db.Exec("DROP TABLE IF EXISTS albums_v2")
	db.Exec("DROP TABLE IF EXISTS file_folder_mappings")
	db.Exec("DROP TABLE IF EXISTS folders")
	db.Exec("DROP TABLE IF EXISTS permission_group_permissions")
	db.Exec("DROP TABLE IF EXISTS permission_group_folders")
	db.Exec("DROP TABLE IF EXISTS permission_groups")
	db.Exec("DROP TABLE IF EXISTS files")

	// Execute new schema v3
	log.Println("Creating schema v3 tables...")
	if _, err := db.Exec(schemaV3); err != nil {
		log.Printf("Error running migrations to schema v3: %v", err)
		return err
	}

	// Re-enable foreign key constraints
	log.Println("Re-enabling foreign key constraints...")
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}

	// Insert default system settings
	defaultSettings := []struct {
		key   string
		value string
	}{
		{"domain", "localhost:8080"},
		{"site_name", "AwesomeSharing"},
		{"allow_registration", "false"},
	}

	for _, setting := range defaultSettings {
		db.Exec(`INSERT OR IGNORE INTO system_settings (key, value) VALUES (?, ?)`,
			setting.key, setting.value)
	}

	// Update schema version to v3
	db.setSchemaVersion(3)

	log.Println("Database migration to schema v3 completed successfully")
	log.Println("NOTE: All file and album data has been cleared. Please add folders and re-scan.")

	// Now run migration from v3 to v5
	log.Println("Running migration from v3 to v5...")
	if _, err := db.Exec(migrationV4ToV5); err != nil {
		log.Printf("Error running migration to schema v5: %v", err)
		return err
	}
	db.setSchemaVersion(5)
	log.Println("✓ Migration to v5 completed successfully")

	return nil
}

// getSchemaVersion retrieves the current schema version from the database
func (db *DB) getSchemaVersion() int {
	// Create schema_version table if it doesn't exist
	db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	var version int
	err := db.QueryRow("SELECT version FROM schema_version ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil {
		// No version found, return 0
		return 0
	}
	return version
}

// setSchemaVersion sets the current schema version
func (db *DB) setSchemaVersion(version int) error {
	_, err := db.Exec("INSERT INTO schema_version (version) VALUES (?)", version)
	return err
}
