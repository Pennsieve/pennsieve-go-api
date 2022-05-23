package pkg

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/pennsieve/pennsieve-go-api/models/fileInfo"
	"github.com/pennsieve/pennsieve-go-api/models/iconInfo"
	"github.com/pennsieve/pennsieve-go-api/models/packageInfo"
	"log"
	"regexp"
	"strings"
)

type FolderUpload struct {
	Id       string
	Name     string
	ParentId string
	depth    int
}

// Uploadfile is the parsed and cleaned representation of the SQS S3 Put Event
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
