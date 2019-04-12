package migrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type Migration struct {
	Name      string
	CreatedAt time.Time
}

type MigrationFile struct {
	Path    string
	Prefix  string
	Desc    string
	Applied bool
}

// CheckMigrations returns a list of migrations in the specified folder, indicating which
// ones have already been applied to the database.
func CheckMigrations(db *sql.DB, migrationPath string) []MigrationFile {
	list := getList(migrationPath)
	migrations := getMigrations(db)

	// check to see if migrations are applied
	for m := range migrations {
		for f := range list {
			if list[f].Prefix == migrations[m].Name {
				list[f].Applied = true
				break
			}
		}
	}

	return list
}

// get the migrations in the db
func getMigrations(db *sql.DB) []Migration {
	rows, err := db.Query("SELECT name, created_at FROM migrations")

	if err != nil {
		fmt.Println("sql error")
		log.Fatal(err)
	}

	defer rows.Close()

	migrations := make([]Migration, 0)

	for rows.Next() {
		m := Migration{}
		err := rows.Scan(&m.Name, &m.CreatedAt)

		if err != nil {
			fmt.Println("scan error")
			log.Fatal(err)
		}

		migrations = append(migrations, m)
	}

	rows.Close()

	return migrations
}

// get the files in a directory
func getList(source string) []MigrationFile {
	info, err := os.Stat(source)

	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		log.Fatal("migration source is not a directory")
	}

	dir, err := ioutil.ReadDir(source)

	if err != nil {
		log.Fatal(err)
	}

	checker := regexp.MustCompile(`^\d{14}_.*\.sql$`)
	splitter := regexp.MustCompile(`^(\d*)_(.*)\.sql$`)
	list := make([]MigrationFile, 0)

	for file := range dir {
		if !checker.MatchString(dir[file].Name()) {
			continue
		}

		matches := splitter.FindStringSubmatch(dir[file].Name())

		if len(matches) < 3 {
			log.Fatal("not a valid migration file")
		}

		mf := MigrationFile{}
		mf.Path = path.Join(source, dir[file].Name())      // migration location
		mf.Prefix = matches[1]                             // the timestamp id thing on the front
		mf.Desc = strings.ReplaceAll(matches[2], "_", " ") // a more or less readable description
		list = append(list, mf)
	}

	return list
}
