package migrate_test

import (
	"testing"

	"github.com/Fantamstick/migrant/migrate"
	"github.com/stretchr/testify/assert"
)

func TestDropAllTables(t *testing.T) {
	dropTestTables := mustHaveTestTables()
	defer dropTestTables()

	t.Run("it drops all tables", func(t *testing.T) {
		assert.Greaterf(t, countTables(), 0, "there should be two tables")

		err := migrate.DropAllTables(db)

		assert.Nil(t, err, "should not return any errors")
		assert.Equal(t, 0, countTables(), "should be no tables left")
	})
}
