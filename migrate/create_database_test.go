package migrate_test

import (
	"testing"

	"github.com/Fantamstick/migrant/migrate"
	"github.com/stretchr/testify/assert"
)

func TestCreateDatabase(t *testing.T) {
	defer mustExec("DROP DATABASE IF EXISTS test2", "USE test")

	t.Run("it creates a database", func(t *testing.T) {
		_, err := db.Exec("USE test2")
		assert.NotNil(t, err, "should throw error when switching to db that doesn't exist")

		migrate.CreateDatabase(db, "test2")

		_, err = db.Exec("USE test2")
		assert.Nil(t, err, "should not throw error after creating database")
	})
}
