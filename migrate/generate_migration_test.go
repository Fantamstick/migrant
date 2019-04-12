package migrate_test

import (
	"io/ioutil"
	"migrant/migrate"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateMigration(t *testing.T) {

	t.Run("it generates a new database migration", func(t *testing.T) {
		os.Mkdir("../.test", 0777)

		defer func() {
			os.RemoveAll("../.test/")
		}()

		err := migrate.GenerateMigration("../.test", "foo bar baz")
		assert.Nil(t, err, "should return no errors")

		dirInfo, err := ioutil.ReadDir("../.test")
		assert.Nil(t, err, "should be able to read test dir")
		assert.Len(t, dirInfo, 1, "should only have 1 generated file")

		// // look for files like: 20191212012345_name_of_migration.sql
		assert.Regexp(t, regexp.MustCompile(`^\d{14}_.*\.sql$`), dirInfo[0].Name(), "name of file should match pattern")
		assert.NotContains(t, dirInfo[0].Name(), " ", "should not contain any white space")
	})
}
