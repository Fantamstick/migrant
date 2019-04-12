package migrate

import (
	"database/sql"
	"log"
)

// TruncateTables momentarily disables foreign key checks, then truncates all
// tables in the database. It will not delete entries from the migration table.
func TruncateTables(db *sql.DB) error {
	rows, err := db.Query("SHOW TABLES")

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var tables []string

	for rows.Next() {
		var t string
		err := rows.Scan(&t)

		if err != nil {
			log.Fatal(err)
		}
		tables = append(tables, t)
	}
	rows.Close()

	db.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	defer db.Exec("SET FORGEIGN KEY CHECKS = 1;")

	for t := range tables {
		if tables[t] == "migrations" {
			continue
		}

		_, err := db.Exec("TRUNCATE " + tables[t])

		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
