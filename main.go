package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

	upCommand = &cobra.Command{
		Use:   "up",
		Short: "apply migrations to the database",
		Run:   up,
	}

	genCommand = &cobra.Command{
		Use:   "gen",
		Short: "generate a new migration",
		Run:   gen,
	}
)

// Parameters
var (
	configFileName string
	targetDatabase string
)

func init() {
	command.PersistentFlags().StringVarP(&configFileName, "config", "c", "./config.yml", "the name of the config file")

	command.AddCommand(upCommand)
	upCommand.Flags().StringVarP(&targetDatabase, "database", "d", "default!", "which database to target (or use default db)")

	command.AddCommand(genCommand)
	genCommand.Flags().StringVarP(&targetDatabase, "database", "d", "default!", "which database to generate a migration for (or use default db)")

	viper.SetDefault("migrations", "./migrations")
}

func main() {
	if err := command.Execute(); err != nil {
		log.Fatal(err)
	}
}

// apply migrations to the database if they are not in the migrations table.
func up(cmd *cobra.Command, args []string) {
	mustLoadConfig(configFileName)
	dbConfig := mustFindDatabase(targetDatabase)
	db := mustConnectTo(dbConfig)
	defer db.Close()

	fmt.Printf("%+v", dbConfig)

	migrationsPath := mustFindMigrationsPath(dbConfig)

	fmt.Print(migrationsPath)

	//
	// migrate.CheckMigrations(db)
}

// generate a new migration file
func gen(cmd *cobra.Command, args []string) {

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
