package db

import (
	"github.com/pennsieve/pennsieve-go-api/config"
	"log"
	"strconv"
)

type Dataset struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

// GetAll returns all rows in the Upload Record Table
func (*Dataset) GetAll(datasetId int32) ([]Dataset, error) {
	schemaTable := "\"" + strconv.FormatInt(int64(datasetId), 10) + "\".datasets"
	queryStr := "SELECT (name, state) FROM " + schemaTable

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
