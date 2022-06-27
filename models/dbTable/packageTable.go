package dbTable

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/pennsieve/pennsieve-go-api/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-api/models/packageInfo/packageState"
	"github.com/pennsieve/pennsieve-go-api/models/packageInfo/packageType"
	"log"
	"regexp"
	"strings"
	"time"
)

// Package is a representation of a container on Pennsieve that contains one or more sourceFiles
type Package struct {
	Id           int64                          `json:"id"`
	Name         string                         `json:"name"`
	PackageType  packageType.Type               `json:"type"`
	PackageState packageState.State             `json:"state"`
	NodeId       string                         `json:"node_id"`
	ParentId     sql.NullInt64                  `json:"parent_id"`
	DatasetId    int                            `json:"dataset_id"`
	OwnerId      int                            `json:"owner_id"`
	Size         sql.NullInt64                  `json:"size"`
	ImportId     sql.NullString                 `json:"import_id"`
	Attributes   []packageInfo.PackageAttribute `json:"attributes"`
	CreatedAt    time.Time                      `json:"created_at"`
	UpdatedAt    time.Time                      `json:"updated_at"`
}

// PackageParams is used as the input to create a package
type PackageParams struct {
	Name         string                         `json:"name"`
	PackageType  packageType.Type               `json:"type"`
	PackageState packageState.State             `json:"state"`
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

// Add adds multiple packages to the Pennsieve Postgres DB
func (p *Package) Add(db *sql.DB, records []PackageParams) ([]Package, error) {

	currentTime := time.Now()
	var vals []interface{}
	var inserts []string

	// All records have the same datasetID
	datasetId := records[0].DatasetId

	// CHECK EXISTING FILES IN FOLDER AND UPDATE NAME IF NECESSARY

	// Group files by parentID so we can combine SQL queries for children of the parent.
	parentIdMap := map[int64][]PackageParams{}
	for _, r := range records {
		parentIdMap[r.ParentId] = append(parentIdMap[r.ParentId], r)
	}

	// Iterate over map of parentIDs and get children that have names like the ones uploaded.
	for key, value := range parentIdMap {
		var names []string
		for _, v := range value {
			r := regexp.MustCompile(`(?P<FileName>[^\.]*)?\.?(?P<Extension>.*)`)
			pathParts := r.FindStringSubmatch(v.Name)
			fName := pathParts[r.SubexpIndex("FileName")]
			names = append(names, fmt.Sprintf("'%s%%'", fName))
		}
		arrayString := strings.Join(names, ",")

		sqlString := fmt.Sprintf("SELECT name FROM packages WHERE dataset_id=%d AND parent_id=%d AND name LIKE ANY (ARRAY[%s]);", datasetId, key, arrayString)
		if key == -1 {
			sqlString = fmt.Sprintf("SELECT name FROM packages WHERE dataset_id=%d AND parent_id IS NULL AND name LIKE ANY (ARRAY[%s]);", datasetId, arrayString)
		}

		stmt, err := db.Prepare(sqlString)
		if err != nil {
			log.Fatalln("Something is wrong", err)
		}

		// format all vals at once
		var allNames []string
		rows, _ := stmt.Query(vals...)
		for rows.Next() {
			var currentFile string
			err = rows.Scan(
				&currentFile,
			)
			allNames = append(allNames, currentFile)
		}

		// Update names if suggested name exists.
		for i, _ := range records {
			checkUpdateName(&records[i], 1, "", allNames)
		}

	}

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
		"dataset_id, owner_id, size, import_id, created_at, updated_at"

	sqlInsert = sqlInsert + strings.Join(inserts, ",") + fmt.Sprintf("RETURNING %s;", returnRows)

	//prepare the statement
	stmt, err := db.Prepare(sqlInsert)
	if err != nil {
		log.Fatalln("ERROR: ", err)
	}
	defer stmt.Close()

	// format all vals at once
	var allInsertedPackages []Package
	rows, err := stmt.Query(vals...)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			log.Println(pqErr)
		}
		return nil, err
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
			&currentRecord.ImportId,
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
		folderFilter = fmt.Sprintf("AND type = '%s'", packageType.Collection.String())
	}

	// Return children for specific dataset in specific org with specific parent.
	// Do NOT return any packages that are in DELETE State
	queryRows := "id, name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, created_at, updated_at"

	queryStr := fmt.Sprintf("SELECT %s FROM packages WHERE dataset_id = %d AND parent_id = %d AND state != '%s' %s;",
		queryRows, datasetId, parent.Id, packageState.Deleting.String(), folderFilter)

	// If parent is empty => return children of root of dataset.
	if parent.NodeId == "" {
		queryStr = fmt.Sprintf("SELECT %s FROM packages WHERE dataset_id = %d AND parent_id IS NULL AND state != '%s' %s;",
			queryRows, datasetId, packageState.Deleting.String(), folderFilter)
	}

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

// checkUpdateName Recursively checks name and append integer if name exists.
func checkUpdateName(item *PackageParams, index int, newName string, namesInFolder []string) {

	if newName == "" {
		newName = item.Name
	}

	for _, n := range namesInFolder {
		if newName == n {
			r := regexp.MustCompile(`(?P<FileName>[^\.]*)?\.?(?P<Extension>.*)`)
			pathParts := r.FindStringSubmatch(item.Name)

			name := pathParts[r.SubexpIndex("FileName")]
			extension := pathParts[r.SubexpIndex("Extension")]

			index++

			updatedName := ""
			if extension != "" {
				updatedName = fmt.Sprintf("%s (%d).%s", name, index, extension)
			} else {
				updatedName = fmt.Sprintf("%s (%d)", name, index)
			}

			// Recursively call this function to check if updated name also exists.
			checkUpdateName(item, index, updatedName, namesInFolder)
			return
		}
	}

	// Update name to new name
	item.Name = newName
}
