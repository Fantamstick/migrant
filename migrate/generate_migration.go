package migrate

import (
	"io/ioutil"
	"path"
	"strings"
	"time"
)

// GenerateMigration creates a new empty sql file prefixed with a time stamp.
func GenerateMigration(dir, desc string) error {
	dateComponent := time.Now().Format("20060102150405")
	descComponent := strings.ReplaceAll(desc, " ", "_")
	fileName := dateComponent + "_" + descComponent + ".sql"
	filePath := path.Join(dir, fileName)
	err := ioutil.WriteFile(filePath, []byte("-- Write your migration here"), 0644)
	return err
}
