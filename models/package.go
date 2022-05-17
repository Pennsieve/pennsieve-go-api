package models

import (
	"fmt"
	pennsievePackage "github.com/pennsieve/pennsieve-go-api/models/packageType"
	"github.com/pennsieve/pennsieve-go-api/pkg/config"
	"log"
	"strconv"
	"strings"
	"time"
)

// Package is a representation of a container on Pennsieve that contains one or more sourceFiles
type Package struct {
	Id           int                    `json:"id"`
	Name         string                 `json:"name"`
	PackageType  pennsievePackage.Type  `json:"type"`
	PackageState pennsievePackage.State `json:"state"`
	NodeId       string                 `json:"node_id"`
	ParentId     int                    `json:"parent_id"`
	DatasetId    int                    `json:"dataset_id"`
	OwnerId      int                    `json:"owner_id"`
	Size         int64                  `json:"size"`
	ImportId     string                 `json:"import_id"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type PackageParams struct {
	Name         string                 `json:"name"`
	PackageType  pennsievePackage.Type  `json:"type"`
	PackageState pennsievePackage.State `json:"state"`
	NodeId       string                 `json:"node_id"`
	ParentId     int                    `json:"parent_id"`
	DatasetId    int                    `json:"dataset_id"`
	OwnerId      int                    `json:"owner_id"`
	Size         int64                  `json:"size"`
	ImportId     string                 `json:"import_id"`
}

// Add adds multiple packages to the Pennsieve Postgres DB
func (*Package) Add(organizationId int, records []PackageParams) error {

	currentTime := time.Now()
	const rowSQL = "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	var vals []interface{}
	var inserts []string

	schemaTable := "\"" + strconv.FormatInt(int64(organizationId), 10) + "\".packages"

	sqlInsert := fmt.Sprintf("INSERT INTO %s (name, type, state, node_id, parent_id, dataset_id, owner_id, "+
		"size, import_id, created_at, updated_at) VALUES ", schemaTable)
	for _, row := range records {
		inserts = append(inserts, rowSQL)
		vals = append(vals, row.Name, row.PackageType.String(), row.PackageState.String(), row.NodeId,
			row.ParentId, row.OwnerId, row.Size, row.ImportId, currentTime, currentTime)
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
