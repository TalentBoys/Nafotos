package initialization

import (
	"database/sql"
	"log"
	"os"

	"awesome-sharing/internal/services"
)

// InitializeDefaultData creates default server_owner user
func InitializeDefaultData(db *sql.DB) error {
	authService := services.NewAuthService(db)

	// Check if server_owner already exists
	var serverOwnerCount int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'server_owner'").Scan(&serverOwnerCount)
	if err != nil {
		log.Printf("Error checking server_owner count: %v", err)
		return err
	}

	// Create server_owner if not exists
	if serverOwnerCount == 0 {
		log.Println("No server_owner found. Creating server_owner user...")

		// Get username and password from environment variables
		username := os.Getenv("SERVER_OWNER_USERNAME")
		password := os.Getenv("SERVER_OWNER_PASSWORD")

		// Use default values if environment variables are not set
		if username == "" {
			username = "server-owner"
		}
		if password == "" {
			password = "server-owner"
		}

		// Create server_owner
		_, err := authService.CreateUser(username, password, "owner@localhost", "server_owner")
		if err != nil {
			log.Printf("Error creating server_owner: %v", err)
			return err
		}
		log.Printf("✓ Server owner user created (username: %s)", username)
		log.Println("⚠️  IMPORTANT: Please change the default password immediately!")
		log.Println("⚠️  server_owner has full system access and can manage all users")
	}

	return nil
}


// CleanupExpiredSessions removes expired sessions periodically
func CleanupExpiredSessions(db *sql.DB) {
	authService := services.NewAuthService(db)
	err := authService.CleanupExpiredSessions()
	if err != nil {
		log.Printf("Error cleaning up expired sessions: %v", err)
	}
}
