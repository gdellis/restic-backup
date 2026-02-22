package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Notifier interface {
	Send(subject, body string) error
}

type Config struct {
	Enabled    bool
	OnSuccess  bool
	OnFailure  bool
	WebhookURL string
	SMTP       *SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       string
}

type WebhookNotifier struct {
	URL    string
	Method string
}

func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL:    url,
		Method: "POST",
	}
}

func (n *WebhookNotifier) Send(subject, body string) error {
	payload := map[string]string{
		"subject": subject,
		"body":    body,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequest(n.Method, n.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status: %d", resp.StatusCode)
	}
	
	return nil
}

type SMTPNotifier struct {
	config *SMTPConfig
}

func NewSMTPNotifier(cfg *SMTPConfig) *SMTPNotifier {
	return &SMTPNotifier{
		config: cfg,
	}
}

func (n *SMTPNotifier) Send(subject, body string) error {
	return fmt.Errorf("SMTP notifier not implemented: use webhook for now")
}

type MultiNotifier struct {
	notifiers []Notifier
}

func NewMultiNotifier(configs []Notifier) *MultiNotifier {
	return &MultiNotifier{
		notifiers: configs,
	}
}

func (n *MultiNotifier) Send(subject, body string) error {
	var errs []string
	for _, notifier := range n.notifiers {
		if err := notifier.Send(subject, body); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func SendNotification(event string, config *Config) error {
	if config == nil || !config.Enabled {
		return nil
	}
	
	if event == "success" && !config.OnSuccess {
		return nil
	}
	if event == "failure" && !config.OnFailure {
		return nil
	}
	
	subject := fmt.Sprintf("Restic Backup: %s", event)
	body := fmt.Sprintf("Backup %s at %s", event, os.Getenv("HOSTNAME"))
	
	if config.WebhookURL != "" {
		notifier := NewWebhookNotifier(config.WebhookURL)
		return notifier.Send(subject, body)
	}
	
	return nil
}
