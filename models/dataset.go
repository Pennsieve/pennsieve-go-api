package models

import (
	"database/sql"
	"log"
)

type Dataset struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

// getSchemaTable returns a string with the table name prepended with the schema name.
//func (*Dataset) getSchemaTable(organizationId int) string {
//	return "\"" + strconv.FormatInt(int64(organizationId), 10) + "\".datasets"
//}

// GetAll returns all rows in the Upload Record Table
func (d *Dataset) GetAll(db *sql.DB, organizationId int) ([]Dataset, error) {
	queryStr := "SELECT (name, state) FROM datasets"

	rows, err := db.Query(queryStr)
	var allDatasets []Dataset
	if err == nil {
		for rows.Next() {
			var currentRecord Dataset
			err = rows.Scan(
				&currentRecord.Name,
				&currentRecord.State)

			if err != nil {
				log.Println("ERROR: ", err)
			}

			allDatasets = append(allDatasets, currentRecord)
		}
		return allDatasets, err
	}
	return allDatasets, err
}
