package migrate_test

import (
	"testing"

	"github.com/Fantamstick/migrant/migrate"
	"github.com/stretchr/testify/assert"
)

func TestInitMigrationTable(t *testing.T) {
	defer mustExec("DROP TABLE migrations")

	t.Run("it creates a migration table if none exists", func(t *testing.T) {
		migrate.InitMigrationTable(db)

		_, err := db.Exec("SELECT count(*) FROM migrations")
		assert.Nil(t, err)
	})
}
