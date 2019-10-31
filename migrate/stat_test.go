package migrate_test

import (
	"testing"

	"github.com/Fantamstick/migrant/migrate"
	"github.com/stretchr/testify/assert"
)

func TestStat(t *testing.T) {
	t.Run("it returns info about database", func(t *testing.T) {
		// should scan a database and check if the migrations table exists
		info, err := migrate.Stat(db)

		assert.Nil(t, err, "should not return error")
		assert.False(t, info.HasMigrationTable(), "should report migration table not present")

		// now add migration table
		closeMigrations := mustAddMigrations()
		defer closeMigrations()

		info, err = migrate.Stat(db)

		assert.Nil(t, err, "should not return error")
		assert.True(t, info.HasMigrationTable(), "should report migration table present")
	})
}
