package uploadFile

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-api/models/fileInfo"
	"github.com/pennsieve/pennsieve-go-api/models/iconInfo"
	"github.com/pennsieve/pennsieve-go-api/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-api/models/uploadFolder"
	"log"
	"regexp"
	"sort"
	"strings"
)

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

// Sort sorts []UploadFiles by the depth of the folder the file resides in.
func (f *UploadFile) Sort(files []UploadFile) {
	sort.Slice(files, func(i, j int) bool {
		//pathSlices1 := strings.Split(files[i].Path, "/")
		//pathSlices2 := strings.Split(files[j].Path, "/")
		return files[i].Path < files[j].Path
	})
}

// getFileInfo returns a FileTypeInfo for a particular extension.
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

// GetUploadFolderMap returns an object that maps path name to Folder object.
func (f *UploadFile) GetUploadFolderMap(sortedFiles []UploadFile, targetFolder string) uploadFolder.UploadFolderMap {

	// Mapping path from targetFolder to UploadFolder Object
	var folderNameMap = map[string]*uploadFolder.UploadFolder{}

	// Iterate over the files and create the UploadFolder objects.
	for index, f := range sortedFiles {

		if f.Path == "" {
			continue
		}

		fmt.Printf("File index: %d, File Path: %s\n ", index, f.Path)

		// Prepend the target-Folder if it exists
		p := f.Path
		if targetFolder != "" {
			p = strings.Join(
				[]string{
					targetFolder, p,
				}, "/")
		}

		// Iterate over path segments in a single file and identify folders.
		pathSegments := strings.Split(p, "/")
		absoluteSegment := "" // Current location in the path walker for current file.
		currentNodeId := ""
		currentFolderPath := ""
		for depth, segment := range pathSegments {

			fmt.Printf("Depth: %d, segment: %s, abs_segment: %s\n", depth, segment, absoluteSegment)

			parentNodeId := currentNodeId
			parentFolderPath := currentFolderPath

			// If depth > 0 ==> prepend the previous absoluteSegment to the current path name.
			if depth > 0 {
				absoluteSegment = strings.Join(
					[]string{

						absoluteSegment,
						segment,
					}, "/")
			} else {
				absoluteSegment = segment
			}

			// If folder already exists in map, add current folder as a child to the parent
			// folder (which will exist too at this point). If not, create new folder to the map and add to parent folder.

			folder, ok := folderNameMap[absoluteSegment]
			if ok {
				currentNodeId = folder.NodeId
				currentFolderPath = absoluteSegment

			} else {
				currentNodeId = fmt.Sprintf("N:collection:%s", uuid.New().String())
				currentFolderPath = absoluteSegment

				folder = &uploadFolder.UploadFolder{
					NodeId:       currentNodeId,
					Name:         segment,
					ParentNodeId: parentNodeId,
					ParentId:     -1,
					Depth:        depth,
				}
				folderNameMap[absoluteSegment] = folder

				fmt.Printf("Create Folder: %s with Name: %s\n", absoluteSegment, folder.Name)
			}

			// Add current segment to parent if exist
			if parentFolderPath != "" {
				folderNameMap[parentFolderPath].Children = append(folderNameMap[parentFolderPath].Children, folder)
			}

		}
	}

	return folderNameMap
}
