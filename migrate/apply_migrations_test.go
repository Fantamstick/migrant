package migrate_test

import (
	"migrant/migrate"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyMigrations(t *testing.T) {
	closeMigrations := mustAddMigrations()
	defer closeMigrations()
	defer mustExec("DROP TABLE IF EXISTS test_table_1", "DROP TABLE IF EXISTS test_table_2")

	migrations := []migrate.MigrationFile{
		{
			Prefix:  "20190101001122",
			Path:    "../fixtures/migrations1/20190101001122_test_1.sql",
			Desc:    "test 1",
			Applied: true, // this migration should not be run
		},
		{
			Prefix:  "20190102001122",
			Path:    "../fixtures/migrations1/20190102001122_test_2.sql",
			Desc:    "test 2",
			Applied: false, // this migration should be run
		},
	}

	t.Run("it applies unapplied migrations", func(t *testing.T) {
		err := migrate.ApplyMigrations(db, migrations)
		assert.Nil(t, err, "should return no error")

		_, err = db.Exec("SELECT count(*) FROM test_table_1")
		assert.NotNil(t, err, "test table 1 should not have been created")

		_, err = db.Exec("SELECT count(*) FROM test_table_2")
		assert.Nil(t, err, "test table 2 should have been created")
	})
}
