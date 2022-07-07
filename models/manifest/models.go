package manifest

import "github.com/pennsieve/pennsieve-go-api/models/manifest/manifestFile"

type Status int64

const (
	Initiated Status = iota
	Uploading
	Completed
	Cancelled
)

func (s Status) String() string {
	switch s {
	case Initiated:
		return "Initiated"
	case Uploading:
		return "Uploading"
	case Completed:
		return "Completed"
	case Cancelled:
		return "Cancelled"
	default:
		return "Initiated"
	}
}

func (s Status) ManifestStatusMap(value string) Status {
	switch value {
	case "Initiated":
		return Initiated
	case "Uploading":
		return Uploading
	case "Completed":
		return Completed
	case "Cancelled":
		return Cancelled
	}
	return Initiated
}

type DTO struct {
	ID        string                 `json:"id"`
	DatasetId string                 `json:"dataset_id"`
	Files     []manifestFile.FileDTO `json:"files"`
	Status    Status                 `json:"status"`
}

type PostResponse struct {
	ManifestNodeId string                       `json:"manifest_node_id"'`
	NrFilesUpdated int                          `json:"nr_files_updated"`
	NrFilesRemoved int                          `json:"nr_files_removed"`
	UpdatedFiles   []manifestFile.FileStatusDTO `json:"updated_files"`
	FailedFiles    []string                     `json:"failed_files"`
}

// AddFilesStats object that is returned to the client.
type AddFilesStats struct {
	NrFilesUpdated int
	NrFilesRemoved int
	FileStatus     []manifestFile.FileStatusDTO
	FailedFiles    []string
}
