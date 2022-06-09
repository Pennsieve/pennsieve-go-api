package dbTable

import (
	"github.com/aws/smithy-go/rand"
	"github.com/pennsieve/pennsieve-go-api/models/fileInfo"
	"time"
)

type File struct {
	Id              string        `json:"id"`
	PackageId       int           `json:"package_id"`
	Name            string        `json:"name"`
	FileType        fileInfo.Type `json:"file_type"`
	S3Bucket        string        `json:"s3_bucket"`
	S3Key           string        `json:"s3_key"`
	ObjectType      string        `json:"object_type"`
	Size            int64         `json:"size"`
	CheckSum        string        `json:"checksum"`
	UUID            rand.UUID     `json:"uuid"`
	ProcessingState string        `json:"processing_state"`
	UploadedState   string        `json:"uploaded_state"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

type FileParams struct {
	PackageId  int           `json:"package_id"`
	Name       string        `json:"name"`
	FileType   fileInfo.Type `json:"file_type"`
	S3Bucket   string        `json:"s3_bucket"`
	S3Key      string        `json:"s3_key"`
	ObjectType string        `json:"object_type"`
	Size       int64         `json:"size"`
	CheckSum   string        `json:"checksum"`
	UUID       rand.UUID     `json:"uuid"`
}

func (p *File) Add(organizationId int, files []FileParams) error {
	return nil
}
