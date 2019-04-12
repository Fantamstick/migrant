package migrate_test

import (
	"migrant/migrate"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDropAllTables(t *testing.T) {
	dropTestTables := mustHaveTestTables()
	defer dropTestTables()

	t.Run("it drops all tables", func(t *testing.T) {
		assert.Equal(t, 2, countTables(), "there should be two tables")

		err := migrate.DropAllTables(db)

		assert.Nil(t, err, "should not return any errors")
		assert.Equal(t, 0, countTables(), "should be no tables left")
	})
}
