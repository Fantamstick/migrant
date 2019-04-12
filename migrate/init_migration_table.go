package migrate

import (
	"database/sql"
	"log"
)

// InitMigrationTable checks to see if the migration table exists, and if not, creates it.
func InitMigrationTable(db *sql.DB) {
	_, err := db.Exec("SELECT count(*) FROM migrations")

	if err == nil {
		return
	}

	_, err = db.Exec(`
		CREATE TABLE migrations(
			name VARCHAR(14) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`)

	if err != nil {
		log.Fatal(err)
	}
}
