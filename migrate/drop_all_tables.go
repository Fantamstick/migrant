package migrate

import (
	"database/sql"
	"log"
)

// DropAllTables goes ahead and drops all the tables in the current database.
func DropAllTables(db *sql.DB) error {
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
		_, err := db.Exec("DROP TABLE " + tables[t])

		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
