package app

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"

	"bitbucket.org/fantamstick/migrant/input"
	"bitbucket.org/fantamstick/migrant/migrate"

	"github.com/spf13/viper"
)

// mustFindMigrationsPath looks for the migration path specified in the config and dies if it is not present.
func mustFindMigrationsPath(config DatabaseConfig) string {
	checkPath := path.Join(viper.GetString("migrations"), config.Name)

	info, err := os.Stat(checkPath)

	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		log.Fatal("target migrations location is not a folder")
	}

	return checkPath
}

// HasMigrationTable will check to see if a migration table exists. If not, it will ask the user to make
// one. If the user declines, the app will exit.
func mustHaveOrCreatedMigrationTable(db *sql.DB) {
	info, err := migrate.Stat(db)

	if err != nil {
		log.Fatal(err)
	}

	if !info.HasMigrationTable() {
		if !input.Confirm("There is no migration table. Do you want to create one? (This will alter your database)") {
			fmt.Print("Cannot run migrations without a migration table. No further actions will take place.")
			os.Exit(1)
		}

		migrate.InitMigrationTable(db)
	}
}
