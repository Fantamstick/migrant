package migrate

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"text/template"

	yaml "gopkg.in/yaml.v2"
)

type SeedFile struct {
	Path string
}

type SeedInput struct {
	Vars  map[string]string
	Seeds []Seed
}

type Seed struct {
	Table  string
	Insert []map[string]string
}

// ApplySeeds reads an array of seed files and applies them to the database.
func ApplySeeds(db *sql.DB, seedFiles []SeedFile) {

	// read yaml file
	for s := range seedFiles {
		file, err := ioutil.ReadFile(seedFiles[s].Path)

		if err != nil {
			log.Fatal(err)
		}

		var input SeedInput
		err = yaml.Unmarshal(file, &input)

		if err != nil {
			log.Fatal(err)
		}

		collectedIds := make(map[string][]int64) //  will hold the ids that are generated during the seed process

		f := template.FuncMap{
			"id": func(source string, index int) string {
				return fmt.Sprint(collectedIds[source][index])
			},
			"var": func(source string) string {
				return fmt.Sprint(input.Vars[source])
			},
		}

		for set := range input.Seeds {
			table := input.Seeds[set].Table

			// make sure there's an array to collect ids
			if _, ok := collectedIds[table]; !ok {
				collectedIds[table] = make([]int64, 0)
			}

			if err != nil {
				log.Fatal(err)
			}

			// for each insert, collect the cols and vals. Run each val as a template to get its computed value.
			for i := range input.Seeds[set].Insert {
				insert := input.Seeds[set].Insert[i]

				cols := make([]string, 0)
				vals := make([]interface{}, 0)
				qs := make([]string, 0)

				for colName, valTemplate := range insert {
					t := template.Must(template.New(colName).Funcs(f).Parse(valTemplate))
					var buf bytes.Buffer
					err := t.Execute(&buf, nil)

					if err != nil {
						log.Fatal(err)
					}

					// store the column name and computer value
					cols = append(cols, colName)
					vals = append(vals, buf.String())
					qs = append(qs, "?")
				}

				q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ", "), strings.Join(qs, ", "))
				res, err := db.Exec(q, vals...)

				if err != nil {
					log.Fatal(err)
				}

				lastId, err := res.LastInsertId()

				if err == nil {
					collectedIds[table] = append(collectedIds[table], lastId)
				}
			}
		}
	}
}
