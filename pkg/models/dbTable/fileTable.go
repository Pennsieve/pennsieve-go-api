package dbTable

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pennsieve/pennsieve-go-api/pkg/core"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/fileInfo/objectType"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/fileInfo/processingState"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/fileInfo/uploadState"
	"log"
	"strings"
	"time"
)

type File struct {
	Id              string                          `json:"id"`
	PackageId       int                             `json:"package_id"`
	Name            string                          `json:"name"`
	FileType        fileType.Type                   `json:"file_type"`
	S3Bucket        string                          `json:"s3_bucket"`
	S3Key           string                          `json:"s3_key"`
	ObjectType      objectType.ObjectType           `json:"object_type"`
	Size            int64                           `json:"size"`
	CheckSum        string                          `json:"checksum"`
	UUID            uuid.UUID                       `json:"uuid"`
	ProcessingState processingState.ProcessingState `json:"processing_state"`
	UploadedState   uploadState.UploadedState       `json:"uploaded_state"`
	CreatedAt       time.Time                       `json:"created_at"`
	UpdatedAt       time.Time                       `json:"updated_at"`
}

type FileParams struct {
	PackageId  int                   `json:"package_id"`
	Name       string                `json:"name"`
	FileType   fileType.Type         `json:"file_type"`
	S3Bucket   string                `json:"s3_bucket"`
	S3Key      string                `json:"s3_key"`
	ObjectType objectType.ObjectType `json:"object_type"`
	Size       int64                 `json:"size"`
	CheckSum   string                `json:"checksum"`
	UUID       uuid.UUID             `json:"uuid"`
}

func (p *File) Add(db core.PostgresAPI, files []FileParams) ([]File, error) {

	currentTime := time.Now()
	var vals []interface{}
	var inserts []string

	for index, row := range files {
		inserts = append(inserts, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			index*13+1,
			index*13+2,
			index*13+3,
			index*13+4,
			index*13+5,
			index*13+6,
			index*13+7,
			index*13+8,
			index*13+9,
			index*13+10,
			index*13+11,
			index*13+12,
			index*13+13,
		))

		etag := fmt.Sprintf("{\"checksum\": \"%s\", \"chunkSize\": \"%s\"}", "1234z", "5242880")

		vals = append(vals, row.PackageId, row.Name, row.FileType.String(), row.S3Bucket, row.S3Key,
			row.ObjectType.String(), row.Size, etag, row.UUID.String(), processingState.Unprocessed.String(),
			uploadState.Uploaded.String(), currentTime, currentTime)

	}

	sqlInsert := "INSERT INTO files(package_id, name, file_type, s3_bucket, s3_key, " +
		"object_type, size, checksum, uuid, processing_state, uploaded_state, created_at, updated_at) VALUES "

	returnRows := "id, package_id, name, file_type, s3_bucket, s3_key, " +
		"object_type, size, checksum, uuid, processing_state, uploaded_state, created_at, updated_at"

	sqlInsert = sqlInsert + strings.Join(inserts, ",") + fmt.Sprintf("RETURNING %s;", returnRows)

	//prepare the statement
	stmt, err := db.Prepare(sqlInsert)
	if err != nil {
		log.Fatalln("ERROR: ", err)
	}
	defer stmt.Close()

	// format all vals at once
	var allInsertedFiles []File
	rows, err := stmt.Query(vals...)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			log.Println(pqErr)
		}
		return nil, err
	}

	for rows.Next() {
		var currentRecord File

		var fType string
		var oType string
		var pState string
		var uState string

		err = rows.Scan(
			&currentRecord.Id,
			&currentRecord.PackageId,
			&currentRecord.Name,
			&fType,
			&currentRecord.S3Bucket,
			&currentRecord.S3Key,
			&oType,
			&currentRecord.Size,
			&currentRecord.CheckSum,
			&currentRecord.UUID,
			&pState,
			&uState,
			&currentRecord.CreatedAt,
			&currentRecord.UpdatedAt,
		)

		currentRecord.FileType = fileType.Dict[fType]
		currentRecord.ObjectType = objectType.Dict[oType]
		currentRecord.ProcessingState = processingState.Dict[pState]
		currentRecord.UploadedState = uploadState.Dict[uState]

		if err != nil {
			log.Println("ERROR: ", err)
		}

		allInsertedFiles = append(allInsertedFiles, currentRecord)

	}

	if err != nil {
		log.Println(err)
	}

	return allInsertedFiles, err
}

// UpdateBucket updates the storage bucket as part of upload process and sets Status
func (p *File) UpdateBucket(db core.PostgresAPI, uploadId string, bucket string, organizationId int64) error {

	queryStr := fmt.Sprintf("UPDATE \"%d\".files SET s3_bucket=$1 WHERE UUID=$2;", organizationId)

	result, err := db.Exec(queryStr, bucket, uploadId)
	if err != nil {
		log.Println("Error updating the bucket location: ", err)
		return err
	}

	affectedRows, err := result.RowsAffected()
	if affectedRows != 1 {
		log.Println("UpdateBucket: Unexpected number of updated rows!", affectedRows)
		if affectedRows == 0 {
			return errors.New("zero rows updated when 1 expected``")
		}
	}

	return nil

}
