package manifest

type ManifestStatus int64

const (
	Initiated ManifestStatus = iota
	Uploading
	Completed
	Cancelled
)

func (s ManifestStatus) String() string {
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

func (s ManifestStatus) ManifestStatusMap(value string) ManifestStatus {
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

type ManifestFileStatus int64

const (
	FileInitiated ManifestFileStatus = iota
	FileSynced
	FileUploaded
	FileVerified
	FileFailed
	FileRemoved
)

func (s ManifestFileStatus) String() string {
	switch s {
	case FileInitiated:
		return "Initiated"
	case FileSynced:
		return "Synced"
	case FileUploaded:
		return "Uploaded"
	case FileVerified:
		return "Verified"
	case FileFailed:
		return "Failed"
	case FileRemoved:
		return "Removed"
	default:
		return "Initiated"
	}
}

func (s ManifestFileStatus) ManifestFileStatusMap(value string) ManifestFileStatus {
	switch value {
	case "Initiated":
		return FileInitiated
	case "Synced":
		return FileSynced
	case "Uploaded":
		return FileUploaded
	case "Verified":
		return FileVerified
	case "Removed":
		return FileRemoved
	case "Failed":
		return FileFailed
	}
	return FileInitiated
}

type FileDTO struct {
	UploadID   string             `json:"upload_id"`
	S3Key      string             `json:"s3_key"`
	TargetPath string             `json:"target_path"`
	TargetName string             `json:"target_name"`
	Status     ManifestFileStatus `json:"status"`
}

type DTO struct {
	ID        string         `json:"id"`
	DatasetId string         `json:"dataset_id"`
	Files     []FileDTO      `json:"files"`
	Status    ManifestStatus `json:"status"`
}

type PostResponse struct {
	ManifestNodeId string   `json:"manifest_node_id"'`
	NrFilesUpdated int      `json:"nr_files_updated"`
	NrFilesRemoved int      `json:"nr_files_removed"`
	FailedFiles    []string `json:"failed_files"`
}

// AddFilesStats object that is returned to the client.
type AddFilesStats struct {
	NrFilesUpdated int
	NrFilesRemoved int
	FailedFiles    []string
}
