package migrate_test

import (
	"migrant/migrate"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateTables(t *testing.T) {

	dropTestTables := mustHaveTestTables()
	defer dropTestTables()

	dropMigrations := mustAddMigrations()
	defer dropMigrations()

	mustExec(`INSERT INTO test_table_1 (name) VALUES ("foo"), ("bar"), ("baz");`)
	mustExec(`INSERT INTO migrations (name) VALUES ("test1"), ("test2")`)

	t.Run("it truncates table contents", func(t *testing.T) {
		var count int64
		count = getRowCount("test_table_1")
		assert.Equal(t, int64(3), count, "should have three rows in the database")

		migrate.TruncateTables(db)

		count = getRowCount("test_table_1")
		assert.Equal(t, int64(0), count, "should have no rows now")

		count = getRowCount("migrations")
		assert.Equal(t, int64(2), count, "it should not delete migrations")
	})
}
