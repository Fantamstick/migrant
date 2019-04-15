package main

import (
	"database/sql"
	"fmt"
	"log"
	"migrant/input"
	"migrant/migrate"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// size of indent after description on migrations printout
const INDENT = 10

// DatabaseConfig stores information about a target database
type DatabaseConfig struct {
	Name    string
	Driver  string
	Uri     string
	Default bool
}

// Commands
var (
	command = &cobra.Command{
		Use:   "migrant",
		Short: "relive the migrant experience through database schema management",
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

func main() {
	if err := command.Execute(); err != nil {
		log.Fatal(err)
	}
}

// generate a new migration file
func gen(cmd *cobra.Command, args []string) {
	mustLoadConfig(configFileName)
	dbConfig := mustFindDatabase(targetDatabase)
	migrationPath := mustFindMigrationsPath(dbConfig)
	migrationDesc := args[0]
	err := migrate.GenerateMigration(migrationPath, migrationDesc)

	if err != nil {
		color.Red(fmt.Sprintf("Error generating migration: %s", err.Error()))
	} else {
		color.Green(fmt.Sprintf("Generated migration in %s", migrationPath))
	}
}

// apply migrations to the database if they are not in the migrations table.
func up(cmd *cobra.Command, args []string) {
	mustLoadConfig(configFileName)
	dbConfig := mustFindDatabase(targetDatabase)
	db := mustConnectTo(dbConfig)
	defer db.Close()
	migrationsPath := mustFindMigrationsPath(dbConfig)
	migrations := migrate.CheckMigrations(db, migrationsPath)
	indent := strconv.Itoa(findLongestDesc(migrations) + INDENT)

	for m := range migrations {
		if migrations[m].Applied {
			color.Green(fmt.Sprintf("%s %-"+indent+"s [APPLIED]\n", migrations[m].Prefix, migrations[m].Desc))
		} else {
			color.Red(fmt.Sprintf("%s %-"+indent+"s [NOT APPLIED]\n", migrations[m].Prefix, migrations[m].Desc))
		}
	}

	reader := input.New()

	if !reader.Confirm() {
		fmt.Print("No further actions will take place.")
		return
	}

	err := migrate.ApplyMigrations(db, migrations)

	if err != nil {
		color.Red(fmt.Sprintf("was not able to apply migrations: %s", err.Error()))
	}

	color.Green("...all done 😎")
}

// seed the selected database
func seed(cmd *cobra.Command, args []string) {
	mustLoadConfig(configFileName)
	dbConfig := mustFindDatabase(targetDatabase)
	db := mustConnectTo(dbConfig)
	defer db.Close()
	reader := input.New()

	color.Red("*************************************************************")
	color.Red("* This will destroy all data and replace with initial seeds *")
	color.Red("*************************************************************")

	if !reader.Confirm() {
		fmt.Print("No further actions will take place.")
		return
	}

	err := migrate.TruncateTables(db)

	if err != nil {
		color.Red(fmt.Sprintf("Error during initial truncate stage: %s", err.Error()))
	}

	files := mustFindSeedFiles(args)
	migrate.ApplySeeds(db, files)
	color.Green("...all done 😎")
}

// destroy all tables in database and reapply all migrations
func reset(cmd *cobra.Command, args []string) {
	mustLoadConfig(configFileName)
	dbConfig := mustFindDatabase(targetDatabase)
	db := mustConnectTo(dbConfig)
	defer db.Close()
	migrationsPath := mustFindMigrationsPath(dbConfig)
	reader := input.New()

	color.Red("**********************************************************")
	color.Red("* This will destroy all data and re-apply all migrations *")
	color.Red("**********************************************************")

	if !reader.Confirm() {
		fmt.Print("No further actions will take place.")
		return
	}

	migrate.DropAllTables(db)
	migrate.InitMigrationTable(db)
	migrations := migrate.CheckMigrations(db, migrationsPath)
	err := migrate.ApplyMigrations(db, migrations)

	if err != nil {
		color.Red("Was not able to complete migrations - your database is probably in a dire state.")
		color.Red(fmt.Sprintf("Received error: %s", err.Error()))
		return
	}

	color.Green("...all done 😎")
}

// truncate all database tables.
func truncate(cmd *cobra.Command, args []string) {
	mustLoadConfig(configFileName)
	dbConfig := mustFindDatabase(targetDatabase)
	db := mustConnectTo(dbConfig)
	defer db.Close()

	reader := input.New()

	color.Red("******************************")
	color.Red("* This will destroy all data *")
	color.Red("******************************")

	if !reader.Confirm() {
		fmt.Print("No further actions will take place.")
		return
	}

	migrate.TruncateTables(db)

	color.Green("...all done 😎")
}

// the longest description
func findLongestDesc(migrations []migrate.MigrationFile) int {
	longest := 0
	for m := range migrations {
		if l := len(migrations[m].Desc); l > longest {
			longest = l
		}
	}
	return longest
}

// search for the config file and return an error if it doesn't exist
func mustLoadConfig(name string) {
	var suffix = filepath.Ext(name)
	name = strings.TrimSuffix(name, suffix)
	viper.AddConfigPath(".")             // search the current path
	viper.AddConfigPath("/etc/migrant/") // search in the etc path
	viper.SetConfigName(name)
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatal(err)
	}
}

// find the specified database and return the config
func mustFindDatabase(name string) DatabaseConfig {
	if name == "default!" {
		sm := viper.GetStringMap("databases")

		for k := range sm {
			if viper.GetBool("databases." + k + ".default") {
				c := DatabaseConfig{
					Name:    k,
					Driver:  viper.GetString("databases." + k + ".driver"),
					Uri:     viper.GetString("databases." + k + ".uri"),
					Default: true,
				}

				return c
			}
		}

		log.Fatal("default database not found")
	}

	if viper.Get("databases."+name) == nil {
		log.Fatal("database not found")
	}

	c := DatabaseConfig{
		Name:    name,
		Driver:  viper.GetString("databases." + name + ".driver"),
		Uri:     viper.GetString("databases." + name + ".uri"),
		Default: viper.GetBool("databases." + name + ".default"),
	}

	return c
}

// connect to the specified db
func mustConnectTo(config DatabaseConfig) *sql.DB {
	db, err := sql.Open(config.Driver, config.Uri)

	if err != nil {
		log.Fatal(err)
	}

	return db
}

// find where the migrations go for this database
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

// check each path to see if it's a file and return as array of SeedFile objects
func mustFindSeedFiles(paths []string) []migrate.SeedFile {
	files := make([]migrate.SeedFile, 0)

	for p := range paths {
		_, err := os.Stat(paths[p])

		if err != nil {
			log.Fatal(err)
		}

		files = append(files, migrate.SeedFile{Path: paths[p]})
	}

	return files
}
