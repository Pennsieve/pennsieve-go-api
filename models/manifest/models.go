package manifest

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
