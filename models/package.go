package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/models/packageInfo"
	"log"
	"strings"
	"time"
)

// Package is a representation of a container on Pennsieve that contains one or more sourceFiles
type Package struct {
	Id           int64                          `json:"id"`
	Name         string                         `json:"name"`
	PackageType  packageInfo.Type               `json:"type"`
	PackageState packageInfo.State              `json:"state"`
	NodeId       string                         `json:"node_id"`
	ParentId     sql.NullInt64                  `json:"parent_id"`
	DatasetId    int                            `json:"dataset_id"`
	OwnerId      int                            `json:"owner_id"`
	Size         sql.NullInt64                  `json:"size"`
	ImportId     sql.NullInt64                  `json:"import_id"`
	Attributes   []packageInfo.PackageAttribute `json:"attributes"`
	CreatedAt    time.Time                      `json:"created_at"`
	UpdatedAt    time.Time                      `json:"updated_at"`
}

type PackageParams struct {
	Name         string                         `json:"name"`
	PackageType  packageInfo.Type               `json:"type"`
	PackageState packageInfo.State              `json:"state"`
	NodeId       string                         `json:"node_id"`
	ParentId     int64                          `json:"parent_id"`
	DatasetId    int                            `json:"dataset_id"`
	OwnerId      int                            `json:"owner_id"`
	Size         int64                          `json:"size"`
	ImportId     sql.NullString                 `json:"import_id"`
	Attributes   []packageInfo.PackageAttribute `json:"attributes"`
}

// PackageMap maps path to models.Package
type PackageMap = map[string]Package

// getSchemaTable returns a string with the table name prepended with the schema name.
//func (*Package) getSchemaTable(organizationId int) string {
//	return "\"" + strconv.FormatInt(int64(organizationId), 10) + "\".packages"
//}

// Add adds multiple packages to the Pennsieve Postgres DB
func (p *Package) Add(db *sql.DB, organizationId int, records []PackageParams) ([]Package, error) {

	currentTime := time.Now()
	var vals []interface{}
	var inserts []string

	sqlInsert := "INSERT INTO packages(name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, import_id, attributes, created_at, updated_at) VALUES "

	for index, row := range records {
		inserts = append(inserts, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			index*12+1,
			index*12+2,
			index*12+3,
			index*12+4,
			index*12+5,
			index*12+6,
			index*12+7,
			index*12+8,
			index*12+9,
			index*12+10,
			index*12+11,
			index*12+12,
		))

		attributeJson, err := json.Marshal(row.Attributes)
		if err != nil {
			log.Println(err)
		} else if string(attributeJson) == "null" {
			attributeJson = []byte("[]")
		}

		sqlParentId := sql.NullInt64{Valid: false}
		if row.ParentId >= 0 {
			sqlParentId = sql.NullInt64{
				Int64: row.ParentId,
				Valid: true,
			}
		}

		vals = append(vals, row.Name, row.PackageType.String(), row.PackageState.String(), row.NodeId, sqlParentId, row.DatasetId,
			row.OwnerId, row.Size, row.ImportId, string(attributeJson), currentTime, currentTime)
	}

	returnRows := "id, name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, created_at, updated_at"

	sqlInsert = sqlInsert + strings.Join(inserts, ",") + fmt.Sprintf("RETURNING %s;", returnRows)

	//prepare the statement
	stmt, err := db.Prepare(sqlInsert)
	if err != nil {
		log.Fatalln("ERROR: ", err)
	}
	defer stmt.Close()

	// format all vals at once
	var allInsertedPackages []Package
	rows, _ := stmt.Query(vals...)
	for rows.Next() {
		var currentRecord Package
		err = rows.Scan(
			&currentRecord.Id,
			&currentRecord.Name,
			&currentRecord.PackageType,
			&currentRecord.PackageState,
			&currentRecord.NodeId,
			&currentRecord.ParentId,
			&currentRecord.DatasetId,
			&currentRecord.OwnerId,
			&currentRecord.Size,
			&currentRecord.CreatedAt,
			&currentRecord.UpdatedAt,
		)

		if err != nil {
			log.Println("ERROR: ", err)
		}

		allInsertedPackages = append(allInsertedPackages, currentRecord)

	}

	if err != nil {
		log.Println(err)
	}

	return allInsertedPackages, err
}

// Children returns an array of Packages that have a specific parent package or root.
func (p *Package) Children(db *sql.DB, organizationId int, parent *Package, datasetId int, onlyFolders bool) ([]Package, error) {

	folderFilter := ""
	if onlyFolders {
		folderFilter = fmt.Sprintf("AND type = '%s'", packageInfo.Collection.String())
	}

	// Return children for specific dataset in specific org with specific parent.
	// Do NOT return any packages that are in DELETE State
	queryRows := "id, name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, created_at, updated_at"

	queryStr := fmt.Sprintf("SELECT %s FROM packages WHERE dataset_id = %d AND parent_id = %d AND state != '%s' %s;",
		queryRows, datasetId, parent.Id, packageInfo.Deleting.String(), folderFilter)

	// If parent is empty => return children of root of dataset.
	if parent.NodeId == "" {
		fmt.Println("Getting ROOT FOLDERS FOR", parent.Name)
		queryStr = fmt.Sprintf("SELECT %s FROM packages WHERE dataset_id = %d AND parent_id IS NULL AND state != '%s' %s;",
			queryRows, datasetId, packageInfo.Deleting.String(), folderFilter)
	}

	log.Println(queryStr)

	rows, err := db.Query(queryStr)
	var allPackages []Package
	if err != nil {
		log.Fatalln("ERROR: ", err)
	}

	for rows.Next() {
		var currentRecord Package
		err = rows.Scan(
			&currentRecord.Id,
			&currentRecord.Name,
			&currentRecord.PackageType,
			&currentRecord.PackageState,
			&currentRecord.NodeId,
			&currentRecord.ParentId,
			&currentRecord.DatasetId,
			&currentRecord.OwnerId,
			&currentRecord.Size,
			&currentRecord.CreatedAt,
			&currentRecord.UpdatedAt,
		)

		if err != nil {
			log.Println("ERROR: ", err)
		}

		allPackages = append(allPackages, currentRecord)
	}
	return allPackages, err
}
