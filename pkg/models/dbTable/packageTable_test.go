package dbTable

import (
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/pkg/core"
	"github.com/stretchr/testify/assert"
	"os"
	"time"

	"github.com/pennsieve/pennsieve-go-api/pkg/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/packageInfo/packageState"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/packageInfo/packageType"
	"testing"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// TestPackage is main testing function that sets up tests and runs sub-tests.
func TestPackage(t *testing.T) {
	// <setup code>

	pkgAttr1 := packageInfo.PackageAttribute{
		Key:      "subtype",
		Fixed:    false,
		Value:    "Image",
		Hidden:   false,
		Category: "Pennsieve",
		DataType: "string",
	}
	pkgAttr2 := packageInfo.PackageAttribute{
		Key:      "subtype",
		Fixed:    false,
		Value:    "Image",
		Hidden:   false,
		Category: "Pennsieve",
		DataType: "string",
	}

	singlePkg := Package{
		Id:           1,
		Name:         "knuth.jpg",
		PackageType:  1,
		PackageState: 1,
		NodeId:       "N:package:12345678-1111-1111-1111-123456789ABC",
		ParentId: sql.NullInt64{
			Int64: 1,
			Valid: true,
		},
		DatasetId: 1,
		OwnerId:   1,
		Size: sql.NullInt64{
			Int64: 123456,
			Valid: true,
		},
		ImportId: sql.NullString{
			String: "12345678-1111-1111-1111-123456789ABC",
			Valid:  true,
		},
		Attributes: nil,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	singlePkgParams := []PackageParams{
		{
			Name:         "knuth.jpg",
			PackageType:  packageType.Image,
			PackageState: packageState.Uploaded,
			NodeId:       "N:package:12345678-1111-1111-1111-123456789ABC",
			DatasetId:    1,
			ParentId:     1,
			OwnerId:      1,
			Size:         123456789,
			ImportId: sql.NullString{
				String: "12345678-1111-1111-1111-123456789ABC",
				Valid:  true,
			},
			Attributes: []packageInfo.PackageAttribute{pkgAttr1, pkgAttr2},
		},
	}

	leePkg := Package{
		Id:           2,
		Name:         "lee",
		PackageType:  11,
		PackageState: 1,
		NodeId:       "N:package:12345678-2222-2222-2222-123456789ABC",
		ParentId: sql.NullInt64{
			Int64: 5,
			Valid: true,
		},
		DatasetId: 1,
		OwnerId:   15,
		Size: sql.NullInt64{
			Int64: 123456,
			Valid: true,
		},
		ImportId: sql.NullString{
			String: "12345678-2222-2222-2222-123456789ABC",
			Valid:  true,
		},
		Attributes: nil,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	leePkgParams := []PackageParams{
		{
			Name:         "lee.jpg",
			PackageType:  packageType.Collection,
			PackageState: packageState.Uploaded,
			NodeId:       "N:package:12345678-2222-2222-2222-123456789ABC",
			DatasetId:    1,
			ParentId:     2,
			OwnerId:      15,
			Size:         123456789,
			ImportId: sql.NullString{
				String: "12345678-2222-2222-2222-123456789ABC",
				Valid:  true,
			},
			Attributes: []packageInfo.PackageAttribute{pkgAttr1, pkgAttr2},
		},
	}

	const (
		host     = "pennsieve-go-api-postgres-1"
		port     = 5432
		user     = "postgres"
		password = "password"
		dbname   = "postgres"
		orgId    = 1
	)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	fmt.Println("Connect to Postgres")
	db, err := sql.Open("postgres", psqlInfo)

	nullParentStmt, err := db.Prepare("create unique index packages_name_dataset_id__parent_id_null_idx on \"1\".packages (name,dataset_id,\"type\") where parent_id is null;")
	defer nullParentStmt.Close()
	_, err = nullParentStmt.Query()

	nonNullParentStmt, err := db.Prepare("create unique index packages_name_dataset_id_parent_id__parent_id_not_null_idx on \"1\".packages (name,dataset_id,\"type\",parent_id) where parent_id is NOT null;")
	defer nonNullParentStmt.Close()
	_, err = nonNullParentStmt.Query()

	if err != nil {
		panic(err)
	}
	defer db.Close()
	//
	t.Run("AddPackages", func(t *testing.T) {
		_, err = db.Exec(fmt.Sprintf("SET search_path = \"%d\";", orgId))
		testSinglePackageInsert(t, singlePkg, db, singlePkgParams)
		testFailDuplicateNames(t, leePkg, db, leePkgParams)
	})
	// <tear-down code>
}

// Basic sanity test
func testSinglePackageInsert(t *testing.T, p Package, db core.PostgresAPI, pkgParams []PackageParams) {
	result, err := p.Add(db, pkgParams)
	if err != nil {
		fmt.Println("ERROR: " + err.Error())
	}
	assert.Equal(t, err, nil)
	assert.Equal(t, len(result), 1)

}
func testFailDuplicateNames(t *testing.T, p Package, db core.PostgresAPI, pkgParams []PackageParams) {
	initialResult, err1 := p.Add(db, pkgParams)

	pkgParams[0].ImportId = sql.NullString{
		String: "",
		Valid:  false,
	}
	originalNodeID := pkgParams[0].NodeId
	pkgParams[0].NodeId = "N:package:DUPLICATE-1111-1111-1111-123456789ABC"

	duplicatedResult, err2 := p.Add(db, pkgParams)

	assert.Equal(t, err1, nil)
	assert.Equal(t, err2, nil)

	assert.Equal(t, len(initialResult), 1)
	assert.Equal(t, len(duplicatedResult), 1)

	// Check that we get the original node id back
	assert.Equal(t, duplicatedResult[0].NodeId, originalNodeID)

}
