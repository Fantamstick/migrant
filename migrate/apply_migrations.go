package migrate

import (
	"database/sql"
	"io/ioutil"
)

// ApplyMigrations takes an array of migration files. If the file is not yet apply
// it will run the contents against the current db.
func ApplyMigrations(db *sql.DB, migrations []MigrationFile) error {
	for m := range migrations {
		if migrations[m].Applied {
			continue
		}

		sql, err := ioutil.ReadFile(migrations[m].Path)

		if err != nil {
			return err
		}

		_, err = db.Exec(string(sql))

		if err != nil {
			return err
		}
	}

	return nil
}
