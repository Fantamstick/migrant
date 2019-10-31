package app

import (
	"log"
	"os"

	"github.com/Fantamstick/migrant/migrate"
)

// check each path to see if it's a file and return as array of SeedFile objects
func MustFindSeedFiles(paths []string) []migrate.SeedFile {
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
