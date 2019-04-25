package migrate_test

import (
	"testing"

	"bitbucket.org/fantamstick/migrant/migrate"
	"github.com/stretchr/testify/assert"
)

func TestCheckMigrations(t *testing.T) {
	closeMigrations := mustAddMigrations()
	defer closeMigrations()

	t.Run("it returns an empty list if there are no migrations", func(t *testing.T) {
		files := migrate.CheckMigrations(db, "../fixtures/migrations0")
		assert.Len(t, files, 0, "it should not return files")
	})

	t.Run("it returns list of unapplied migrations", func(t *testing.T) {
		files := migrate.CheckMigrations(db, "../fixtures/migrations1")
		assert.Len(t, files, 2, "should return 2 migration")

		assertMigration(t, &files[0], "20190101001122", "test 1", false)
		assertMigration(t, &files[1], "20190102001122", "test 2", false)
	})

	// artificially apply first migration
	mustExec("INSERT INTO migrations (name) VALUES (20190101001122)")

	t.Run("it returns list of applied migrations", func(t *testing.T) {
		files := migrate.CheckMigrations(db, "../fixtures/migrations1")
		assert.Len(t, files, 2, "should return 2 migration")

		assertMigration(t, &files[0], "20190101001122", "test 1", true)
		assertMigration(t, &files[1], "20190102001122", "test 2", false)
	})
}

func assertMigration(t *testing.T, m *migrate.MigrationFile, prefix, desc string, applied bool) {
	assert.Equal(t, prefix, m.Prefix, "should have correct prefix")
	assert.Equal(t, desc, m.Desc, "should have correct description")
	assert.Equal(t, applied, m.Applied, "does not have correct applied status")
}
