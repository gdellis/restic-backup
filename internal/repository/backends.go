package repository

const (
	BackendLocal     = "local"
	BackendSFTP      = "sftp"
	BackendS3        = "s3"
	BackendMinIO     = "minio"
	BackendWasabi    = "wasabi"
	BackendBackblaze = "b2"
	BackendAzure     = "azure"
	BackendGCS       = "gcs"
	BackendSwift     = "swift"
	BackendRest      = "rest"
	BackendRclone    = "rclone"
)

var BackendDescriptions = map[string]string{
	BackendLocal:     "Local directory",
	BackendSFTP:      "SFTP server",
	BackendS3:        "Amazon S3",
	BackendMinIO:     "MinIO server",
	BackendWasabi:    "Wasabi Cloud Storage",
	BackendBackblaze: "Backblaze B2",
	BackendAzure:     "Microsoft Azure Blob Storage",
	BackendGCS:       "Google Cloud Storage",
	BackendSwift:     "OpenStack Swift",
	BackendRest:      "REST server",
	BackendRclone:    "rclone remote",
}

var BackendRequiresPassword = map[string]bool{
	BackendLocal:     false,
	BackendSFTP:      true,
	BackendS3:        true,
	BackendMinIO:     true,
	BackendWasabi:    true,
	BackendBackblaze: true,
	BackendAzure:     true,
	BackendGCS:       true,
	BackendSwift:     true,
	BackendRest:      true,
	BackendRclone:    true,
}

var BackendRequiresEnvPassword = map[string]bool{
	BackendS3:        true,
	BackendMinIO:     true,
	BackendWasabi:    true,
	BackendBackblaze: true,
	BackendAzure:     true,
	BackendGCS:       true,
	BackendSwift:     true,
	BackendRest:      true,
	BackendRclone:    true,
}
