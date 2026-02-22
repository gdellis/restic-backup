package repository

import "testing"

func TestBackendConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"BackendLocal", BackendLocal, "local"},
		{"BackendSFTP", BackendSFTP, "sftp"},
		{"BackendS3", BackendS3, "s3"},
		{"BackendMinIO", BackendMinIO, "minio"},
		{"BackendWasabi", BackendWasabi, "wasabi"},
		{"BackendBackblaze", BackendBackblaze, "b2"},
		{"BackendAzure", BackendAzure, "azure"},
		{"BackendGCS", BackendGCS, "gcs"},
		{"BackendSwift", BackendSwift, "swift"},
		{"BackendRest", BackendRest, "rest"},
		{"BackendRclone", BackendRclone, "rclone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestBackendDescriptions(t *testing.T) {
	expectedBackends := []string{
		BackendLocal, BackendSFTP, BackendS3, BackendMinIO,
		BackendWasabi, BackendBackblaze, BackendAzure, BackendGCS,
		BackendSwift, BackendRest, BackendRclone,
	}

	for _, backend := range expectedBackends {
		desc, ok := BackendDescriptions[backend]
		if !ok {
			t.Errorf("missing description for backend %s", backend)
		}
		if desc == "" {
			t.Errorf("empty description for backend %s", backend)
		}
	}
}

func TestBackendRequiresPassword(t *testing.T) {
	localRequires, ok := BackendRequiresPassword[BackendLocal]
	if !ok {
		t.Error("BackendRequiresPassword missing entry for local")
	}
	if localRequires {
		t.Error("Local backend should not require password")
	}

	sftpRequires := BackendRequiresPassword[BackendSFTP]
	if !sftpRequires {
		t.Error("SFTP backend should require password")
	}
}

func TestBackendRequiresEnvPassword(t *testing.T) {
	localRequires := BackendRequiresEnvPassword[BackendLocal]
	if localRequires {
		t.Error("Local backend should not require env password")
	}

	sftpRequires := BackendRequiresEnvPassword[BackendSFTP]
	if sftpRequires {
		t.Error("SFTP backend uses interactive password, not env")
	}

	expectedEnvBackends := []string{
		BackendS3, BackendMinIO, BackendWasabi, BackendBackblaze,
		BackendAzure, BackendGCS, BackendSwift, BackendRest, BackendRclone,
	}

	for _, backend := range expectedEnvBackends {
		if !BackendRequiresEnvPassword[backend] {
			t.Errorf("Backend %s should require env password", backend)
		}
	}
}

func TestBackendDescriptionsComplete(t *testing.T) {
	allBackends := []string{
		BackendLocal, BackendSFTP, BackendS3, BackendMinIO,
		BackendWasabi, BackendBackblaze, BackendAzure, BackendGCS,
		BackendSwift, BackendRest, BackendRclone,
	}

	if len(BackendDescriptions) != len(allBackends) {
		t.Errorf("BackendDescriptions has %d entries, expected %d",
			len(BackendDescriptions), len(allBackends))
	}

	if len(BackendRequiresPassword) != len(allBackends) {
		t.Errorf("BackendRequiresPassword has %d entries, expected %d",
			len(BackendRequiresPassword), len(allBackends))
	}

	if len(BackendRequiresEnvPassword) != len(allBackends)-2 {
		t.Errorf("BackendRequiresEnvPassword has %d entries, expected %d (all except local and sftp)",
			len(BackendRequiresEnvPassword), len(allBackends)-2)
	}
}
