package packageinfo

import (
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/pkg/config"
	"log"
	"strconv"
	"strings"
	"time"
)

// PackageState is an enum indicating the state of the Package
type PackageState int64

const (
	Unavailable PackageState = iota
	Uploaded
	Deleting
	Infected
	UploadFailed
	Processing
	Ready
	ProcessingFailed
)

func (s PackageState) String() string {
	switch s {
	case Unavailable:
		return "UNAVAILABLE"
	case Uploaded:
		return "UPLOADED"
	case Deleting:
		return "DELETING"
	case Infected:
		return "INFECTED"
	case UploadFailed:
		return "UPLOAD_FAILED"
	case Processing:
		return "PROCESSING"
	case Ready:
		return "READY"
	case ProcessingFailed:
		return "PROCESSING_FAILED"
	}
	return "UNKNOWN"
}

// PackageType is an enum indicating the type of the Package
type PackageType int64

const (
	Image PackageType = iota
	MRI
	Slide
	ExternalFile
	MSWord
	PDF
	CSV
	Tabular
	TimeSeries
	Video
	Unknown
	Collection
	Text
	Unsupported
	HDF5
	ZIP
)

func (s PackageType) String() string {
	switch s {
	case Image:
		return "Image"
	case MRI:
		return "MRI"
	case Slide:
		return "Slide"
	case ExternalFile:
		return "ExternalFile"
	case MSWord:
		return "MSWord"
	case PDF:
		return "PDF"
	case CSV:
		return "CSV"
	case Tabular:
		return "Tabular"
	case TimeSeries:
		return "TimeSeries"
	case Video:
		return "Video"
	case Unknown:
		return "Unknown"
	case Collection:
		return "Collection"
	case Text:
		return "Text"
	case Unsupported:
		return "Unsupported"
	case HDF5:
		return "HDF5"
	case ZIP:
		return "ZIP"
	}
	return "Unknown"
}

// Package is a representation of a container on Pennsieve that contains one or more sourceFiles
type Package struct {
	Id           int          `json:"id"`
	Name         string       `json:"name"`
	PackageType  PackageType  `json:"type"`
	PackageState PackageState `json:"state"`
	NodeId       string       `json:"node_id"`
	ParentId     int          `json:"parent_id"`
	DatasetId    int          `json:"dataset_id"`
	OwnerId      int          `json:"owner_id"`
	Size         int64        `json:"size"`
	ImportId     string       `json:"import_id"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type PackageParams struct {
	Name         string       `json:"name"`
	PackageType  PackageType  `json:"type"`
	PackageState PackageState `json:"state"`
	NodeId       string       `json:"node_id"`
	ParentId     int          `json:"parent_id"`
	DatasetId    int          `json:"dataset_id"`
	OwnerId      int          `json:"owner_id"`
	Size         int64        `json:"size"`
	ImportId     string       `json:"import_id"`
}

// Add adds multiple packages to the Pennsieve Postgres DB
func (*Package) Add(organizationId int, records []PackageParams) error {

	currentTime := time.Now()
	const rowSQL = "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	var vals []interface{}
	var inserts []string

	schemaTable := "\"" + strconv.FormatInt(int64(organizationId), 10) + "\".packages"

	sqlInsert := fmt.Sprintf("INSERT INTO %s (name, type, state, node_id, parent_id, dataset_id, owner_id, "+
		"size, import_id, created_at, updated_at) VALUES ", schemaTable)
	for _, row := range records {
		inserts = append(inserts, rowSQL)
		vals = append(vals, row.Name, row.PackageType.String(), row.PackageState.String(), row.NodeId,
			row.ParentId, row.OwnerId, row.Size, row.ImportId, currentTime, currentTime)
	}
	sqlInsert = sqlInsert + strings.Join(inserts, ",")

	//prepare the statement
	stmt, err := config.DB.Prepare(sqlInsert)
	if err != nil {
		log.Fatalln("ERROR: ", err)
	}
	defer stmt.Close()

	// format all vals at once
	_, err = stmt.Exec(vals...)
	if err != nil {
		log.Println(err)
	}

	return nil
}
