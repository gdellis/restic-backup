package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"filippo.io/age"
)

type EncryptedStorage struct {
	identity age.Identity
	recipient age.Recipient
}

func NewEncryptedStorage(identity age.Identity) *EncryptedStorage {
	return &EncryptedStorage{
		identity: identity,
	}
}

func (e *EncryptedStorage) SetRecipient(recipient age.Recipient) {
	e.recipient = recipient
}

func (e *EncryptedStorage) EncryptAndSave(cfg *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var buf bytes.Buffer
	writer, err := age.Encrypt(&buf, e.recipient)
	if err != nil {
		return fmt.Errorf("failed to create encryptor: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to finalize encryption: %w", err)
	}

	return os.WriteFile(GetConfigPath(), buf.Bytes(), 0600)
}

func (e *EncryptedStorage) LoadAndDecrypt() (*Config, error) {
	data, err := os.ReadFile(GetConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Version: 1,
				Repositories: []Repository{},
				Settings: &Settings{
					Theme: "dark",
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if len(data) == 0 {
		return &Config{Version: 1, Repositories: []Repository{}}, nil
	}

	reader, err := age.Decrypt(bytes.NewReader(data), e.identity)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	decrypted, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decrypted data: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(decrypted, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.Settings == nil {
		cfg.Settings = &Settings{Theme: "dark"}
	}

	return &cfg, nil
}

func LoadIdentity() (age.Identity, error) {
	identPath := GetIdentityPath()
	data, err := os.ReadFile(identPath)
	if err != nil {
		if os.IsNotExist(err) {
			identity, genErr := age.GenerateX25519Identity()
			if genErr != nil {
				return nil, genErr
			}
			identStr := identity.String()
			if err := os.WriteFile(identPath, []byte(identStr), 0600); err != nil {
				return nil, err
			}
			return identity, nil
		}
		return nil, err
	}
	return age.ParseX25519Identity(string(bytes.TrimSpace(data)))
}

func LoadRecipient() (age.Recipient, error) {
	identPath := GetIdentityPath()
	data, err := os.ReadFile(identPath)
	if err != nil {
		return nil, err
	}
	
	identity, err := age.ParseX25519Identity(string(bytes.TrimSpace(data)))
	if err != nil {
		return nil, err
	}
	
	return identity.Recipient(), nil
}

func AddRepository(cfg *Config, repo Repository) error {
	repo.ID = fmt.Sprintf("repo-%d", time.Now().Unix())
	repo.CreatedAt = time.Now().Format(time.RFC3339)
	repo.UpdatedAt = repo.CreatedAt
	cfg.Repositories = append(cfg.Repositories, repo)
	return nil
}

func UpdateRepository(cfg *Config, id string, updateFn func(*Repository) error) error {
	for i, repo := range cfg.Repositories {
		if repo.ID == id {
			if err := updateFn(&cfg.Repositories[i]); err != nil {
				return err
			}
			cfg.Repositories[i].UpdatedAt = time.Now().Format(time.RFC3339)
			return nil
		}
	}
	return fmt.Errorf("repository not found: %s", id)
}

func RemoveRepository(cfg *Config, id string) error {
	for i, repo := range cfg.Repositories {
		if repo.ID == id {
			cfg.Repositories = append(cfg.Repositories[:i], cfg.Repositories[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("repository not found: %s", id)
}

func GetRepositoryByID(cfg *Config, id string) *Repository {
	for _, repo := range cfg.Repositories {
		if repo.ID == id {
			return &repo
		}
	}
	return nil
}

func (r *Repository) GetPassword() (string, error) {
	if r.PasswordEnv != "" {
		pw := os.Getenv(r.PasswordEnv)
		if pw == "" {
			return "", fmt.Errorf("environment variable %s is not set", r.PasswordEnv)
		}
		return pw, nil
	}
	return r.Password, nil
}
