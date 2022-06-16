package manifest

type ManifestStatus int64

const (
	ManifestInitiated ManifestStatus = iota
	ManifestUploading
	ManifestCompleted
	ManifestCancelled
)

func (s ManifestStatus) String() string {
	switch s {
	case ManifestInitiated:
		return "Initiated"
	case ManifestUploading:
		return "InProgress"
	case ManifestCompleted:
		return "Completed"
	case ManifestCancelled:
		return "Cancelled"
	default:
		return "Initiated"
	}
}

func (s ManifestStatus) ManifestStatusMap(value string) ManifestStatus {
	switch value {
	case "Initiated":
		return ManifestInitiated
	case "InProgress":
		return ManifestUploading
	case "Completed":
		return ManifestCompleted
	case "Cancelled":
		return ManifestCancelled
	}
	return ManifestInitiated
}

type ManifestFileStatus int64

const (
	FileRegistered ManifestFileStatus = iota
	FileSynced
	FileUploading
	FileCompleted
	FileVerified
	FileCancelled
)

func (s ManifestFileStatus) String() string {
	switch s {
	case FileRegistered:
		return "Indexed"
	case FileSynced:
		return "Synced"
	case FileUploading:
		return "Uploading"
	case FileCompleted:
		return "Completed"
	case FileVerified:
		return "Verified"
	case FileCancelled:
		return "Cancelled"
	default:
		return "Initiated"
	}
}

func (s ManifestFileStatus) ManifestFileStatusMap(value string) ManifestFileStatus {
	switch value {
	case "Indexed":
		return FileRegistered
	case "Synced":
		return FileSynced
	case "Uploading":
		return FileUploading
	case "Completed":
		return FileCompleted
	case "Verified":
		return FileVerified
	case "Cancelled":
		return FileCancelled
	}
	return FileRegistered
}

type FileDTO struct {
	UploadID   string `json:"uploadId"`
	S3Key      string `json:"s3Key"`
	TargetPath string `json:"targetPath"`
	TargetName string `json:"targetName"`
}

type DTO struct {
	ID        string    `json:"id"`
	DatasetId string    `json:"dataset_id"`
	Files     []FileDTO `json:"files"`
}

type PostResponse struct {
	ManifestNodeId string   `json:"manifest_node_id"'`
	NrFilesAdded   int      `json:"nrFilesAdded"`
	FailedFiles    []string `json:"failedFiles"`
}
