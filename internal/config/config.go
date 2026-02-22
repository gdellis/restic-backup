package config

import (
	"errors"
	"os"
	"path/filepath"
)

var (
	ErrNoIdentity    = errors.New("no age identity found")
	ErrDecryptFailed = errors.New("failed to decrypt config")
)

type Identity struct {
	Recipients []string
	Passphrase string
}

type Config struct {
	Version      int           `json:"version"`
	Identity     *Identity     `json:"identity,omitempty"`
	Repositories []Repository  `json:"repositories"`
	Settings     *Settings     `json:"settings,omitempty"`
}

type Repository struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Backend     string            `json:"backend"`
	Path        string            `json:"path"`
	Password    string            `json:"password,omitempty"`
	PasswordEnv string            `json:"password_env,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
	BackupPaths []string         `json:"backup_paths,omitempty"`
	Exclude     []string         `json:"exclude,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
	Schedule    string           `json:"schedule,omitempty"`
	Retention   *RetentionPolicy `json:"retention,omitempty"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

type RetentionPolicy struct {
	KeepLast     int `json:"keep_last"`
	KeepDaily    int `json:"keep_daily"`
	KeepWeekly   int `json:"keep_weekly"`
	KeepMonthly  int `json:"keep_monthly"`
	KeepYearly   int `json:"keep_yearly"`
	KeepTags     []string `json:"keep_tags"`
}

type Settings struct {
	ResticPath      string `json:"restic_path,omitempty"`
	DefaultBackend  string `json:"default_backend"`
	Notifications  *Notifications `json:"notifications,omitempty"`
	Theme          string `json:"theme"`
}

type Notifications struct {
	Enabled    bool     `json:"enabled"`
	OnSuccess  bool     `json:"on_success"`
	OnFailure  bool     `json:"on_failure"`
	WebhookURL string   `json:"webhook_url,omitempty"`
	SMTP       *SMTPConfig `json:"smtp,omitempty"`
}

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	From     string `json:"from"`
	To       string `json:"to"`
}

func GetConfigDir() string {
	home, _ := os.UserHomeDir()
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "restic-client")
	}
	return filepath.Join(home, ".config", "restic-client")
}

func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.json")
}

func GetIdentityPath() string {
	return filepath.Join(GetConfigDir(), "identity.txt")
}

func EnsureConfigDir() error {
	dir := GetConfigDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0700)
	}
	return nil
}
