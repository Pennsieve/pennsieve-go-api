package pkg

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-api/models/dbTable"
	"github.com/pennsieve/pennsieve-go-api/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-api/models/uploadFile"
	"github.com/pennsieve/pennsieve-go-api/models/uploadFolder"
	"github.com/pennsieve/pennsieve-go-api/pkg/core"
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

// Close closes the organization connection associated with the session.
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

	db, err := core.ConnectRDSWithOrg(organizationId)
	s.db = db
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetCreateUploadFolders creates new folders in the organization.
// It updates UploadFolders with real folder ID for folders that already exist.
// Assumes map keys are absolute paths in the dataset
func (s *UploadSession) GetCreateUploadFolders(folders uploadFolder.UploadFolderMap) dbTable.PackageMap {

	// Create map to map parentID to array of children

	// Get Root Folders
	p := dbTable.Package{}
	rootChildren, _ := p.Children(s.db, s.organizationId, &p, s.datasetId, true)

	// Map NodeId to Packages for folders that exist in DB
	existingFolders := dbTable.PackageMap{}
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
			pkgParams := dbTable.PackageParams{
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

			result, _ := p.Add(s.db, s.organizationId, []dbTable.PackageParams{pkgParams})
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
func (s *UploadSession) GetPackageParams(uploadFiles []uploadFile.UploadFile, packageMap dbTable.PackageMap) ([]dbTable.PackageParams, error) {
	var pkgParams []dbTable.PackageParams

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

		pkgParam := dbTable.PackageParams{
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

	// Iterate over folders and create them if they do not exist in organization
	packageMap := s.GetCreateUploadFolders(folderMap)

	// 3. Create Package Params to add files to packages table.
	pkgParams, _ := s.GetPackageParams(files, packageMap)

	var packageTable dbTable.Package
	packageTable.Add(s.db, s.organizationId, pkgParams)

}
