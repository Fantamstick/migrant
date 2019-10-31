package app

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Fantamstick/migrant/input"
	"github.com/Fantamstick/migrant/migrate"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// size of indent after description on migrations printout
const INDENT = 10

// current version number
const VERSION = "1.0"

// Commands
var (
	command = &cobra.Command{
		Use:     "migrant",
		Short:   "relive the migrant experience through database schema management",
		Version: VERSION,
	}

	genCommand = &cobra.Command{
		Use:   "gen",
		Short: "generate a new migration",
		Run:   gen,
		Args:  cobra.MinimumNArgs(1),
	}

	upCommand = &cobra.Command{
		Use:   "up",
		Short: "apply migrations to the database",
		Run:   up,
	}

	seedCommand = &cobra.Command{
		Use:   "seed",
		Short: "seed target database",
		Run:   seed,
		Args:  cobra.MaximumNArgs(1),
	}

	resetCommand = &cobra.Command{
		Use:   "reset",
		Short: "reapply ALL migrations to database",
		Run:   reset,
	}

	truncateCommand = &cobra.Command{
		Use:   "truncate",
		Short: "truncate all tables in the database",
		Run:   truncate,
	}
)

// Parameters
var (
	configFileName string
	targetDatabase string
)

func init() {
	command.PersistentFlags().StringVarP(&configFileName, "config", "c", "./config.yml", "the name of the config file")
	command.PersistentFlags().StringVarP(&targetDatabase, "database", "d", "default!", "which database to target (or use default db)")

	command.AddCommand(genCommand)
	command.AddCommand(upCommand)
	command.AddCommand(seedCommand)
	command.AddCommand(resetCommand)
	command.AddCommand(truncateCommand)

	// defaults for config
	viper.SetDefault("migrations", "./migrations")
}

// Run the main command
func Run() {
	if err := command.Execute(); err != nil {
		log.Fatal(err)
	}
}

// generate a new migration file
func gen(cmd *cobra.Command, args []string) {
	MustLoadConfig(configFileName)
	dbConfig := MustFindDBConfig(targetDatabase)
	migrationPath := mustFindMigrationsPath(dbConfig)
	migrationDesc := args[0]

	err := migrate.GenerateMigration(migrationPath, migrationDesc)

	if err != nil {
		color.Red(fmt.Sprintf("Error generating migration: %s", err.Error()))
		return
	}

	color.Green(fmt.Sprintf("Generated migration in %s", migrationPath))
}

// apply migrations to the database if they are not in the migrations table.
func up(cmd *cobra.Command, args []string) {
	MustLoadConfig(configFileName)
	MustLoadSecrets()
	dbConfig := MustFindDBConfig(targetDatabase)
	db := MustConnect(dbConfig)
	defer db.Close()
	mustHaveOrCreatedMigrationTable(db)
	migrationsPath := mustFindMigrationsPath(dbConfig)
	migrations := migrate.CheckMigrations(db, migrationsPath)
	indent := strconv.Itoa(FindLongestDesc(migrations) + INDENT)
	willApply := 0

	for m := range migrations {
		if migrations[m].Applied {
			color.Green(fmt.Sprintf("%s %-"+indent+"s [APPLIED]\n", migrations[m].Prefix, migrations[m].Desc))
		} else {
			color.Red(fmt.Sprintf("%s %-"+indent+"s [NOT APPLIED]\n", migrations[m].Prefix, migrations[m].Desc))
			willApply++
		}
	}

	if willApply == 0 {
		fmt.Printf("No migrations to apply. All done 😎")
		return
	}

	fmt.Printf("Will apply %d migrations", willApply)

	if !input.Confirm() {
		fmt.Print("No further actions will take place.")
		return
	}

	err := migrate.ApplyMigrations(db, migrations)

	if err != nil {
		color.Red(fmt.Sprintf("was not able to apply migrations: %s", err.Error()))
	}

	color.Green("All done 😎")
}

// seed the selected database
func seed(cmd *cobra.Command, args []string) {
	MustLoadConfig(configFileName)
	MustLoadSecrets()
	dbConfig := MustFindDBConfig(targetDatabase)
	db := MustConnect(dbConfig)
	defer db.Close()

	color.Red("*********************************************************")
	color.Red("* This will destroy all data and replace with seed data *")
	color.Red("*********************************************************")

	if !input.Confirm() {
		fmt.Print("No further actions will take place.")
		return
	}

	err := migrate.TruncateTables(db)

	if err != nil {
		color.Red(fmt.Sprintf("Error during initial truncate stage: %s", err.Error()))
	}

	files := MustFindSeedFiles(args)
	migrate.ApplySeeds(db, files)
	color.Green("...all done 😎")
}

// destroy all tables in database and reapply all migrations
func reset(cmd *cobra.Command, args []string) {
	var err error

	MustLoadConfig(configFileName)
	MustLoadSecrets()
	dbConfig := MustFindDBConfig(targetDatabase)

	db := MustConnect(dbConfig)
	defer db.Close()

	migrationsPath := mustFindMigrationsPath(dbConfig)

	color.Red("**********************************************************")
	color.Red("* This will destroy all data and re-apply all migrations *")
	color.Red("**********************************************************")

	if !input.Confirm() {
		fmt.Print("No further actions will take place.")
		return
	}

	err = migrate.DropAllTables(db)

	if err != nil {
		color.Red("Was not able to drop tables before reset - your database is probably in a dire state.")
		color.Red(fmt.Sprintf("Received error: %s", err.Error()))
		return
	}

	migrate.InitMigrationTable(db)
	migrations := migrate.CheckMigrations(db, migrationsPath)
	err = migrate.ApplyMigrations(db, migrations)

	if err != nil {
		color.Red("Was not able to complete migrations - your database is probably in a dire state.")
		color.Red(fmt.Sprintf("Received error: %s", err.Error()))
		return
	}

	color.Green("All done 😎")
}

// truncate all database tables.
func truncate(cmd *cobra.Command, args []string) {
	MustLoadConfig(configFileName)
	MustLoadSecrets()
	dbConfig := MustFindDBConfig(targetDatabase)
	db := MustConnect(dbConfig)

	defer db.Close()

	color.Red("******************************")
	color.Red("* This will destroy all data *")
	color.Red("******************************")

	if !input.ConfirmByTyping("destroy") {
		fmt.Print("No further actions will take place.")
		return
	}

	migrate.TruncateTables(db)

	color.Green("...all done 😎")
}
