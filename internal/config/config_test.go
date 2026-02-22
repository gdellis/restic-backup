package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Unsetenv("HOME")
		os.Unsetenv("XDG_CONFIG_HOME")
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		}
	}()

	tests := []struct {
		name           string
		xdgConfigHome  string
		expectedSuffix string
	}{
		{
			name:           "default XDG config home",
			xdgConfigHome:  "/custom/xdg",
			expectedSuffix: "restic-client",
		},
		{
			name:           "fallback to home",
			xdgConfigHome:  "",
			expectedSuffix: ".config/restic-client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.xdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
			os.Unsetenv("HOME")

			result := GetConfigDir()
			if tt.xdgConfigHome != "" {
				if result != filepath.Join(tt.xdgConfigHome, "restic-client") {
					t.Errorf("GetConfigDir() = %v, want %v", result, filepath.Join(tt.xdgConfigHome, "restic-client"))
				}
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	result := GetConfigPath()
	expectedSuffix := filepath.Join("restic-client", "config.json")
	if filepath.Base(filepath.Dir(result)) != "restic-client" {
		t.Errorf("GetConfigPath() = %v, expected to end with %v", result, expectedSuffix)
	}
}

func TestGetIdentityPath(t *testing.T) {
	result := GetIdentityPath()
	if filepath.Base(filepath.Dir(result)) != "restic-client" {
		t.Errorf("GetIdentityPath() = %v, expected to end with restic-client/identity.txt", result)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if origHome != "" {
			os.Setenv("HOME", origHome)
		} else {
			os.Unsetenv("HOME")
		}
		if origXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	os.Unsetenv("XDG_CONFIG_HOME")
	os.Setenv("HOME", tmpDir)

	err := EnsureConfigDir()
	if err != nil {
		t.Errorf("EnsureConfigDir() error = %v", err)
	}

	expectedDir := filepath.Join(tmpDir, ".config", "restic-client")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("EnsureConfigDir() did not create directory at %s", expectedDir)
	}
}

func TestEnsureConfigDirAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "restic-client")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	origHome := os.Getenv("HOME")
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if origHome != "" {
			os.Setenv("HOME", origHome)
		} else {
			os.Unsetenv("HOME")
		}
		if origXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	os.Unsetenv("XDG_CONFIG_HOME")
	os.Setenv("HOME", tmpDir)

	err := EnsureConfigDir()
	if err != nil {
		t.Errorf("EnsureConfigDir() error = %v", err)
	}
}
