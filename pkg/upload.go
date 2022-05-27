package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/pennsieve/pennsieve-go-api/models"
	"github.com/pennsieve/pennsieve-go-api/models/fileInfo"
	"github.com/pennsieve/pennsieve-go-api/models/iconInfo"
	"github.com/pennsieve/pennsieve-go-api/models/packageInfo"
	"log"
	"regexp"
	"sort"
	"strings"
)

// UploadFolder represents a folder that is part of an upload session.
type UploadFolder struct {
	Id           int64           // Id of the folder
	NodeId       string          // NodeId of the folder
	Name         string          // Name of the folder
	ParentId     int64           // Id of the parent (-1 for root)
	ParentNodeId string          // NodeId for the parent ("" for root)
	Depth        int             // Depth of folder in relation to root
	Children     []*UploadFolder // Children contains folders that need to be created that have current folder as parent.
}

// UploadFile is the parsed and cleaned representation of the SQS S3 Put Event
type UploadFile struct {
	SessionId string           // SessionId is id for the entire upload session.
	Path      string           // Path to collection without file-name
	Name      string           // Name is the filename including extension(s)
	Extension string           // Extension of file (separated from name)
	Type      packageInfo.Type // Type of the Package.
	SubType   string           // SubType of the file
	Icon      iconInfo.Icon    // Icon for the file
	Size      int64            // Size of file
	ETag      string           // ETag provided by S3

}

type UploadFolderMap = map[string]*UploadFolder //Maps path to UploadFolder
type PackageMap = map[string]models.Package     // Maps path to models.Package

// String returns a json representation of the UploadFile object
func (f *UploadFile) String() string {
	str, _ := json.Marshal(f)
	return string(str)
}

// FromS3Event parses the S3 PUT event and creates an UploadFile object.
func (f *UploadFile) FromS3Event(event *events.S3Event) (UploadFile, error) {

	// 1. Regex Path
	/*
		Match Upload Session Token as first 36 characters
		Match Key as string after Upload Session Token.
	*/
	r := regexp.MustCompile(`(?P<Session>[a-z0-9-]{36})\/(?P<Key>.*)`)
	res := r.FindStringSubmatch(event.Records[0].S3.Object.Key)
	//TODO: Handle non-compliant paths and unknown sessions

	// 2. Split Path into name and path
	/*
		Match path as 0+ segments that end with a /
		Match Filename as set of characters up to the first .
		Match optional Extension as everything after the first . in Filename
	*/
	path := res[r.SubexpIndex("Key")]
	r2 := regexp.MustCompile(`(?P<Path>([^\/]*\/)*)(?P<FileName>[^\.]*)?\.(?P<Extension>.*)`)
	pathParts := r2.FindStringSubmatch(path)

	fileExtension := pathParts[r2.SubexpIndex("Extension")]
	str := []string{pathParts[r2.SubexpIndex("FileName")], fileExtension}
	fileName := strings.Join(str, ".")

	fileInfo := getFileInfo(fileExtension)

	// TODO: Handle grouping somewhere
	// Option 1: Reach out to manifest DB and check when potential grouping?

	// 4. Create UploadFile Object
	result := UploadFile{
		SessionId: res[r.SubexpIndex("Session")],
		Path:      pathParts[r2.SubexpIndex("Path")],
		Name:      fileName,
		Extension: fileExtension,
		Type:      fileInfo.PackageType,
		SubType:   fileInfo.PackageSubType,
		Icon:      fileInfo.Icon,
		Size:      event.Records[0].S3.Object.Size,
		ETag:      event.Records[0].S3.Object.ETag,
	}

	return result, nil
}

// getFileInfo updates the UploadFile object and sets the PackageType
func getFileInfo(extension string) packageInfo.FileTypeInfo {

	fileType, exists := fileInfo.FileExtensionDict[extension]
	if !exists {
		fileType = fileInfo.Unknown
	}

	packageType, exists := packageInfo.FileTypeDict[fileType]
	if !exists {
		log.Println("Unmatched filetype. ?!?")
		packageType = packageInfo.FileTypeDict[fileInfo.Unknown]
	}

	return packageType
}

// GetCreateUploadFolders creates new folders in the database.
// It updates UploadFolders with real folder ID for folders that already exist.
// Assumes map keys are absolute paths in the dataset
func GetCreateUploadFolders(organizationId int, datasetId int, ownerId int, folders UploadFolderMap) PackageMap {

	// Create map to map parentID to array of children

	// Get Root Folders
	p := models.Package{}
	rootChildren, _ := p.Children(organizationId, &p, datasetId, true)

	// Map NodeId to Packages for folders that exist in DB
	existingFolders := PackageMap{}
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
			children, _ := p.Children(organizationId, &folder, datasetId, true)
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
				DatasetId:    datasetId,
				OwnerId:      ownerId,
				Size:         0,
				Attributes:   nil,
			}

			result, _ := p.Add(organizationId, []models.PackageParams{pkgParams})
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
