package pkg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/pennsieve/pennsieve-go-api/models"
	"github.com/pennsieve/pennsieve-go-api/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-api/models/uploadFile"
	"github.com/pennsieve/pennsieve-go-api/models/uploadFolder"
	"log"
	"sort"
)

// UploadSession contains the information that is shared based on the upload session ID
type UploadSession struct {
	organizationId  int    `json:"organization_id"`
	datasetId       int    `json:"dataset_id"`
	ownerId         int    `json:"owner_id"`
	targetPackageId string `json:"target_package_id"`
	db              *sql.DB
}

// Close closes the database connection associated with the session.
func (s *UploadSession) Close() {
	err := s.db.Close()
	if err != nil {
		log.Println("Unable to close DB connection from Lambda function.")
		return
	}

}

// CreateUploadSession returns an authenticated object based on the uploadSession UUID
func (*UploadSession) CreateUploadSession(uploadSessionId string) (*UploadSession, error) {

	//TODO: Replace by checking DB
	// This function should check and validate 'credentials' based on the uploadSessionID
	// The methods associated with type UploadSession can then safely interact with the DB.

	organizationId := 19

	s := UploadSession{
		organizationId: organizationId, // Pennsieve Test
		datasetId:      1682,           // Test Upload
		ownerId:        24,             // Joost
	}

	db, err := s.connectRDS()
	s.db = db
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetCreateUploadFolders creates new folders in the database.
// It updates UploadFolders with real folder ID for folders that already exist.
// Assumes map keys are absolute paths in the dataset
func (s *UploadSession) GetCreateUploadFolders(folders uploadFolder.UploadFolderMap) models.PackageMap {

	// Create map to map parentID to array of children

	// Get Root Folders
	p := models.Package{}
	rootChildren, _ := p.Children(s.db, s.organizationId, &p, s.datasetId, true)

	// Map NodeId to Packages for folders that exist in DB
	existingFolders := models.PackageMap{}
	for _, k := range rootChildren {
		existingFolders[k.Name] = k
	}

	// Sort the keys of the map so we can iterate over the sorted map
	pathKeys := make([]string, 0)
	for k, _ := range folders {
		pathKeys = append(pathKeys, k)
	}
	sort.Strings(pathKeys)

	// Iterate over the sorted map
	for _, path := range pathKeys {

		if folder, ok := existingFolders[path]; ok {

			// Use existing folder
			folders[path].NodeId = folder.NodeId
			folders[path].Id = folder.Id

			// Iterate over map and update values that have identified current folder as parent.
			for _, childFolder := range folders[path].Children {
				childFolder.ParentId = folder.Id
				childFolder.ParentNodeId = folder.NodeId
			}

			// Add children of current folder to existing folders
			children, _ := p.Children(s.db, s.organizationId, &folder, s.datasetId, true)
			for _, k := range children {
				p := fmt.Sprintf("%s/%s", path, k.Name)
				existingFolders[p] = k
			}

		} else {
			// Create folder
			pkgParams := models.PackageParams{
				Name:         folders[path].Name,
				PackageType:  packageInfo.Collection,
				PackageState: packageInfo.Ready,
				NodeId:       folders[path].NodeId,
				ParentId:     folders[path].ParentId,
				DatasetId:    s.datasetId,
				OwnerId:      s.ownerId,
				Size:         0,
				Attributes:   nil,
			}

			result, _ := p.Add(s.db, s.organizationId, []models.PackageParams{pkgParams})
			folders[path].Id = result[0].Id
			existingFolders[path] = result[0]

			for _, childFolder := range folders[path].Children {
				childFolder.ParentId = result[0].Id
				childFolder.ParentNodeId = result[0].NodeId
			}
		}
	}

	return existingFolders

}

// GetPackageParams returns an array of PackageParams to insert in the Packages Table.
func (s *UploadSession) GetPackageParams(uploadFiles []uploadFile.UploadFile, packageMap models.PackageMap) ([]models.PackageParams, error) {
	var pkgParams []models.PackageParams

	for _, file := range uploadFiles {
		packageID := fmt.Sprintf("N:package:%s", uuid.New().String())

		parentId := int64(-1)
		if file.Path != "" {
			parentId = packageMap[file.Path].Id
		}

		//TODO: Replace by s3Key when mapped
		uploadId := sql.NullString{
			String: uuid.New().String(),
			Valid:  true,
		}

		// Set Default attributes for File ==> Subtype and Icon
		var attributes []packageInfo.PackageAttribute
		attributes = append(attributes, packageInfo.PackageAttribute{
			Key:      "subType",
			Fixed:    false,
			Value:    file.SubType,
			Hidden:   true,
			Category: "Pennsieve",
			DataType: "string",
		}, packageInfo.PackageAttribute{
			Key:      "icon",
			Fixed:    false,
			Value:    file.Icon.String(),
			Hidden:   true,
			Category: "Pennsieve",
			DataType: "string",
		})

		pkgParam := models.PackageParams{
			Name:         file.Name,
			PackageType:  file.Type,
			PackageState: packageInfo.Uploaded,
			NodeId:       packageID,
			ParentId:     parentId,
			DatasetId:    s.datasetId,
			OwnerId:      s.ownerId,
			Size:         file.Size,
			ImportId:     uploadId,
			Attributes:   attributes,
		}

		pkgParams = append(pkgParams, pkgParam)
	}

	fmt.Println("PKGPARAMS")
	for _, pkg := range pkgParams {
		fmt.Printf("Name: %s, ParentId: %d, NodeId: %s", pkg.Name, pkg.ParentId, pkg.NodeId)
	}

	return pkgParams, nil

}

// ImportFiles is the wrapper function to import files from a single upload-session.
// A single upload session implies that all files belong to the same organization, dataset and owner.
func (s *UploadSession) ImportFiles(files []uploadFile.UploadFile) {

	// Sort files by the length of their path
	// First element closest to root.
	var f uploadFile.UploadFile
	f.Sort(files)

	// Iterate over files and return map of folders and subfolders
	folderMap := f.GetUploadFolderMap(files, "")

	// Iterate over folders and create them if they do not exist in database
	packageMap := s.GetCreateUploadFolders(folderMap)

	// 3. Create Package Params to add files to packages table.
	pkgParams, _ := s.GetPackageParams(files, packageMap)

	var packageTable models.Package
	packageTable.Add(s.db, s.organizationId, pkgParams)

}

// connectRDS returns a DB instance.
// The Lambda function leverages IAM roles to gain access to the DB Proxy.
// The function does NOT set the search_path to the organization schema as multiple
// concurrent upload session can be handled across multiple organizations.
func (s *UploadSession) connectRDS() (*sql.DB, error) {

	var dbName string = "pennsieve_postgres"
	var dbUser string = "dev_rds_proxy_user"
	var dbHost string = "dev-pennsieve-postgres-use1-proxy.proxy-ctkakwd4msv8.us-east-1.rds.amazonaws.com"
	var dbPort int = 5432
	var dbEndpoint string = fmt.Sprintf("%s:%d", dbHost, dbPort)
	var region string = "us-east-1"

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error: " + err.Error())
	}

	authenticationToken, err := auth.BuildAuthToken(
		context.TODO(), dbEndpoint, region, dbUser, cfg.Credentials)
	if err != nil {
		panic("failed to create authentication token: " + err.Error())
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		dbHost, dbPort, dbUser, authenticationToken, dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// Set Search Path to organization
	_, err = db.Exec(fmt.Sprintf("SET search_path = \"%d\";", s.organizationId))
	if err != nil {
		log.Println(fmt.Sprintf("Unable to set search_path to %d.", s.organizationId))
		err := db.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	return db, err
}
