package db

import (
	"encoding/json"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/config"
	"github.com/pennsieve/pennsieve-go-api/models"
	"github.com/pennsieve/pennsieve-go-api/models/packageInfo"
	"log"
	"strconv"
	"strings"
	"time"
)

// Package is a representation of a container on Pennsieve that contains one or more sourceFiles
type Package struct {
	Id           int                    `json:"id"`
	Name         string                 `json:"name"`
	PackageType  packageInfo.Type       `json:"type"`
	PackageState packageInfo.State      `json:"state"`
	NodeId       string                 `json:"node_id"`
	ParentId     int                    `json:"parent_id"`
	DatasetId    int                    `json:"dataset_id"`
	OwnerId      int                    `json:"owner_id"`
	Size         int64                  `json:"size"`
	ImportId     string                 `json:"import_id"`
	Attributes   []models.FileAttribute `json:"attributes"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type PackageParams struct {
	Name         string                 `json:"name"`
	PackageType  packageInfo.Type       `json:"type"`
	PackageState packageInfo.State      `json:"state"`
	NodeId       string                 `json:"node_id"`
	ParentId     int                    `json:"parent_id"`
	DatasetId    int                    `json:"dataset_id"`
	OwnerId      int                    `json:"owner_id"`
	Size         int64                  `json:"size"`
	ImportId     string                 `json:"import_id"`
	Attributes   []models.FileAttribute `json:"attributes"`
}

type FolderUpload struct {
	Id       string
	Name     string
	ParentId string
	depth    int
}

// Add adds multiple packages to the Pennsieve Postgres DB
func (*Package) Add(organizationId int, records []PackageParams) error {

	currentTime := time.Now()
	var vals []interface{}
	var inserts []string

	schemaTable := "\"" + strconv.FormatInt(int64(organizationId), 10) + "\".packages"

	sqlInsert := fmt.Sprintf("INSERT INTO %s(name, type, state, node_id, dataset_id, owner_id, "+
		"size, import_id, attributes, created_at, updated_at) VALUES ", schemaTable)

	for index, row := range records {
		inserts = append(inserts, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			index*11+1,
			index*11+2,
			index*11+3,
			index*11+4,
			index*11+5,
			index*11+6,
			index*11+7,
			index*11+8,
			index*11+9,
			index*11+10,
			index*11+11,
		))

		attributeJson, err := json.Marshal(row.Attributes)
		if err != nil {
			log.Println(err)
		}

		vals = append(vals, row.Name, row.PackageType.String(), row.PackageState.String(), row.NodeId, row.DatasetId,
			row.OwnerId, row.Size, row.ImportId, string(attributeJson), currentTime, currentTime)
	}
	sqlInsert = sqlInsert + strings.Join(inserts, ",")

	//prepare the statement
	stmt, err := config.DB.Prepare(sqlInsert)
	if err != nil {
		log.Fatalln("ERROR: ", err)
	}
	defer stmt.Close()

	// format all vals at once
	_, err = stmt.Exec(vals...)
	if err != nil {
		log.Println(err)
	}

	return nil
}
