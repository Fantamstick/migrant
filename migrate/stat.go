package migrate

import "database/sql"

// DatabaseInfo returns information about a database.
type DatabaseInfo struct {
	hasMigrationTable bool
}

// HasMigrationTable returns true if the database has a migration table.
func (d *DatabaseInfo) HasMigrationTable() bool {
	return d.hasMigrationTable
}

// Stat returns information about the provided database.
func Stat(db *sql.DB) (*DatabaseInfo, error) {
	i := DatabaseInfo{}

	_, err := db.Exec("SELECT count(*) FROM migrations")

	if err == nil {
		i.hasMigrationTable = true
	}

	return &i, nil
}
