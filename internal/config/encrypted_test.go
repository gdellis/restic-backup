package config

import (
	"testing"
)

func TestAddRepository(t *testing.T) {
	cfg := &Config{
		Version:      1,
		Repositories: []Repository{},
	}

	repo := Repository{
		ID:   "",
		Name: "test-repo",
		Path: "/backups/test",
	}

	err := AddRepository(cfg, repo)
	if err != nil {
		t.Errorf("AddRepository() error = %v", err)
	}

	if len(cfg.Repositories) != 1 {
		t.Errorf("expected 1 repository, got %d", len(cfg.Repositories))
	}

	if cfg.Repositories[0].Name != "test-repo" {
		t.Errorf("expected repository name 'test-repo', got %s", cfg.Repositories[0].Name)
	}

	if cfg.Repositories[0].ID == "" {
		t.Error("expected repository ID to be set")
	}

	if cfg.Repositories[0].CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}
}

func TestAddRepositoryMultiple(t *testing.T) {
	cfg := &Config{
		Version:      1,
		Repositories: []Repository{},
	}

	for i := 0; i < 3; i++ {
		repo := Repository{Name: "test-repo"}
		if err := AddRepository(cfg, repo); err != nil {
			t.Errorf("AddRepository() error = %v", err)
		}
	}

	if len(cfg.Repositories) != 3 {
		t.Errorf("expected 3 repositories, got %d", len(cfg.Repositories))
	}
}

func TestUpdateRepository(t *testing.T) {
	cfg := &Config{
		Repositories: []Repository{
			{ID: "repo-1", Name: "original"},
		},
	}

	err := UpdateRepository(cfg, "repo-1", func(r *Repository) error {
		r.Name = "updated"
		return nil
	})
	if err != nil {
		t.Errorf("UpdateRepository() error = %v", err)
	}

	if cfg.Repositories[0].Name != "updated" {
		t.Errorf("expected name 'updated', got %s", cfg.Repositories[0].Name)
	}

	if cfg.Repositories[0].UpdatedAt == "" {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestUpdateRepositoryNotFound(t *testing.T) {
	cfg := &Config{
		Repositories: []Repository{
			{ID: "repo-1", Name: "original"},
		},
	}

	err := UpdateRepository(cfg, "nonexistent", func(r *Repository) error {
		r.Name = "updated"
		return nil
	})
	if err == nil {
		t.Error("expected error for nonexistent repository")
	}
}

func TestRemoveRepository(t *testing.T) {
	cfg := &Config{
		Repositories: []Repository{
			{ID: "repo-1", Name: "first"},
			{ID: "repo-2", Name: "second"},
		},
	}

	err := RemoveRepository(cfg, "repo-1")
	if err != nil {
		t.Errorf("RemoveRepository() error = %v", err)
	}

	if len(cfg.Repositories) != 1 {
		t.Errorf("expected 1 repository, got %d", len(cfg.Repositories))
	}

	if cfg.Repositories[0].ID != "repo-2" {
		t.Errorf("expected remaining repository to be repo-2, got %s", cfg.Repositories[0].ID)
	}
}

func TestRemoveRepositoryNotFound(t *testing.T) {
	cfg := &Config{
		Repositories: []Repository{
			{ID: "repo-1", Name: "first"},
		},
	}

	err := RemoveRepository(cfg, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent repository")
	}
}

func TestGetRepositoryByID(t *testing.T) {
	cfg := &Config{
		Repositories: []Repository{
			{ID: "repo-1", Name: "first"},
			{ID: "repo-2", Name: "second"},
		},
	}

	repo := GetRepositoryByID(cfg, "repo-1")
	if repo == nil {
		t.Error("expected to find repository")
	}
	if repo.Name != "first" {
		t.Errorf("expected 'first', got %s", repo.Name)
	}

	notFound := GetRepositoryByID(cfg, "nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent repository")
	}
}

func TestRepositoryGetPasswordSource(t *testing.T) {
	tests := []struct {
		name     string
		repo     Repository
		expected string
	}{
		{
			name:     "env password",
			repo:     Repository{PasswordEnv: "MY_PASSWORD"},
			expected: "env:MY_PASSWORD",
		},
		{
			name:     "1Password",
			repo:     Repository{PasswordOP: "op://vault/item"},
			expected: "1Password:op://vault/item",
		},
		{
			name:     "encrypted",
			repo:     Repository{Password: "secret"},
			expected: "encrypted",
		},
		{
			name:     "priority env over password",
			repo:     Repository{PasswordEnv: "MY_PASSWORD", Password: "secret"},
			expected: "env:MY_PASSWORD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.repo.GetPasswordSource()
			if result != tt.expected {
				t.Errorf("GetPasswordSource() = %v, want %v", result, tt.expected)
			}
		})
	}
}
