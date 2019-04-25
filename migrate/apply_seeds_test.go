package migrate_test

import (
	"database/sql"
	"log"
	"testing"

	"bitbucket.org/fantamstick/migrant/migrate"
	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
)

func TestApplySeeds(t *testing.T) {
	mustExec(`
		CREATE TABLE test_table_1 (
			id INT AUTO_INCREMENT,
			name VARCHAR(32),
			PRIMARY KEY (id)
		);
	`, `
		CREATE TABLE link_table_1 (
			id INT AUTO_INCREMENT,
			test_table_id INT NOT NULL,
			foo VARCHAR(32),
			bcrypt BINARY(60),
			PRIMARY KEY (id),
			FOREIGN KEY (test_table_id) REFERENCES test_table_1 (id)
		);
	`)

	defer mustExec("DROP TABLE IF EXISTS link_table_1", "DROP TABLE IF EXISTS test_table_1")

	type testTable struct {
		ID   int64
		Name string
	}

	scanTests := func(db *sql.DB) []testTable {
		rows, err := db.Query("SELECT id, name FROM test_table_1")
		assert.Nil(t, err, "should not have returned an error")
		defer rows.Close()

		testTables := make([]testTable, 0)

		for rows.Next() {
			t := testTable{}
			if err := rows.Scan(&t.ID, &t.Name); err != nil {
				log.Fatal(err)
			}
			testTables = append(testTables, t)
		}

		rows.Close()
		return testTables
	}

	type linkTable struct {
		ID          int64
		TestTableID int64
		Foo         string
		Bcrypt      []byte
	}

	scanLinks := func(db *sql.DB) []linkTable {
		rows, err := db.Query("SELECT id, test_table_id, foo, bcrypt FROM link_table_1")
		assert.Nil(t, err, "should not have returned an error")
		defer rows.Close()

		linkTables := make([]linkTable, 0)

		for rows.Next() {
			l := linkTable{}
			if err := rows.Scan(&l.ID, &l.TestTableID, &l.Foo, &l.Bcrypt); err != nil {
				log.Fatal(err)
			}
			linkTables = append(linkTables, l)
		}

		rows.Close()
		return linkTables
	}

	seeds := []migrate.SeedFile{
		{
			Path: "../fixtures/seeds0/20190101001122_seed_1.yaml",
		},
	}

	t.Run("it seeds databases", func(t *testing.T) {
		migrate.ApplySeeds(db, seeds)

		testTables := scanTests(db)
		assert.Len(t, testTables, 1, "should have seeded 1 value")
		assert.Equal(t, "test 1", testTables[0].Name, "should have seeded correct name")

		linkTables := scanLinks(db)
		assert.Len(t, linkTables, 2)
		assert.Equal(t, testTables[0].ID, linkTables[0].TestTableID, "should seed correct reference id")
		assert.Equal(t, "hoge", linkTables[1].Foo, "should seed variable")
		assert.Nil(t, bcrypt.CompareHashAndPassword(linkTables[1].Bcrypt, []byte("secret")), "should decrypt hashed string")
	})
}
