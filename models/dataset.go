package models

import (
	"github.com/pennsieve/pennsieve-go-api/config"
	"log"
	"strconv"
)

type Dataset struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

// getSchemaTable returns a string with the table name prepended with the schema name.
func (*Dataset) getSchemaTable(organizationId int) string {
	return "\"" + strconv.FormatInt(int64(organizationId), 10) + "\".datasets"
}

// GetAll returns all rows in the Upload Record Table
func (d *Dataset) GetAll(organizationId int) ([]Dataset, error) {
	queryStr := "SELECT (name, state) FROM " + d.getSchemaTable(organizationId)

	rows, err := config.DB.Query(queryStr)
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
