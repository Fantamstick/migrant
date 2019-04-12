package migrate_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db *sql.DB
)

func TestMain(m *testing.M) {
	var err error
	db, err = sql.Open("mysql", "root:secret@tcp(127.0.0.1:33061)/?parseTime=true")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	mustExec("DROP DATABASE IF EXISTS test", "CREATE DATABASE test", "USE test")

	res := m.Run()
	os.Exit(res)
}

// exec all the statements or die
func mustExec(q ...string) {
	for i := range q {
		_, err := db.Exec(q[i])

		if err != nil {
			log.Fatal(err)
		}
	}
}

// add migration table, return a function that removes it
func mustAddMigrations() func() {
	mustExec(`
        CREATE TABLE migrations(
            name VARCHAR(14) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW()
        );
	`)

	return func() {
		mustExec(`DROP TABLE migrations`)
	}
}

// add test tables, return a function to remove them
func mustHaveTestTables() func() {
	mustExec(`
		CREATE TABLE test_table_1 (
			id INT AUTO_INCREMENT,
			name VARCHAR(32),
			PRIMARY KEY (id)
		);
	`, `
		CREATE TABLE link_table_1 (
			id INT AUTO_INCREMENT,
			test_table_id INT NOT NULL,
			foo VARCHAR(32),
			PRIMARY KEY (id),
			FOREIGN KEY (test_table_id) REFERENCES test_table_1 (id)
		);
	`)

	return func() {
		mustExec(`DROP TABLE IF EXISTS link_table_1;`, `DROP TABLE IF EXISTS test_table_1;`)
	}
}

// check how many rows are in a table
func getRowCount(table string) int64 {
	rows, err := db.Query("SELECT count(*) FROM " + table)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	var count int64
	rows.Next()
	err = rows.Scan(&count)

	if err != nil {
		log.Fatal(err)
	}

	return count
}

// count how many tables exist in the database
func countTables() int {
	rows, err := db.Query("SHOW TABLES")

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	count := 0

	for rows.Next() {
		count++
	}

	return count
}
