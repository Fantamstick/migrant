package migrate

import "database/sql"

// CreateDatabase will perform the ritual for summoning a new demon.
func CreateDatabase(db *sql.DB, name string) error {
	_, err := db.Exec("CREATE DATABASE " + name)
	return err
}
