package manifestFile

type Status int64

const (
	Initiated Status = iota // set when client creates manifest (local only)
	Synced                  // set when client syncs with server (local, server)
	Imported                // set when uploader imported file (local, server)
	Finalized               // set when importer moves file to final destination (local, server)
	Verified                // set when client is informed of finalized status (local, server)
	Failed                  // set when importer fails to import (local, server)
	Removed                 // set when client removes file locally (local only)
	Unknown                 // set when sync failed (local only)
)

// String returns string version of FileStatus object.
func (s Status) String() string {
	switch s {
	case Initiated:
		return "Initiated"
	case Synced:
		return "Synced"
	case Imported:
		return "Imported"
	case Finalized:
		return "Finalized"
	case Verified:
		return "Verified"
	case Failed:
		return "Failed"
	case Removed:
		return "Removed"
	case Unknown:
		return "Unknown"
	default:
		return "Initiated"
	}
}

// ManifestFileStatusMap maps string values to FileStatus objects.
func (s Status) ManifestFileStatusMap(value string) Status {
	switch value {
	case "Initiated":
		return Initiated
	case "Synced":
		return Synced
	case "Imported":
		return Imported
	case "Finalized":
		return Finalized
	case "Verified":
		return Verified
	case "Removed":
		return Removed
	case "Failed":
		return Failed
	case "Unknown":
		return Unknown
	}
	return Initiated
}

// FileDTO used to transfer file object in API requests.
type FileDTO struct {
	UploadID   string `json:"upload_id"`
	S3Key      string `json:"s3_key"`
	TargetPath string `json:"target_path"`
	TargetName string `json:"target_name"`
	Status     Status `json:"status"`
}

// FileStatusDTO used to transfer status information in API requests.
type FileStatusDTO struct {
	UploadId string `json:"upload_id"`
	Status   Status `json:"status"`
}
