package app

import (
	"fmt"
	"strings"

	"github.com/Fantamstick/migrant/migrate"
	"github.com/fatih/color"
)

// FindLongestDesc returns the longest description in an array of migration files
func FindLongestDesc(migrations []migrate.MigrationFile) int {
	longest := 0
	for m := range migrations {
		if l := len(migrations[m].Desc); l > longest {
			longest = l
		}
	}
	return longest
}

// PrintFlag prints the given text in a fancy box. Not being used? How sad...
func PrintFlag(text string, col *color.Color) {

	length := len(text)
	length += 4
	hr := strings.Repeat("*", length)

	col.Println(hr)
	col.Println(fmt.Sprintf("* %s *", text))
	col.Println(hr)
}
