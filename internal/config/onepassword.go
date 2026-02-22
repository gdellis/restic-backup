package config

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type OPClient struct {
	vault string
}

type OPItem struct {
	ID   string
	JSON map[string]interface{}
}

func NewOPClient(vault string) *OPClient {
	return &OPClient{
		vault: vault,
	}
}

func (c *OPClient) GetItem(itemName string) (*OPItem, error) {
	cmd := exec.Command("op", "item", "get", itemName, "--vault", c.vault, "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("op item get failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return parseOPItem(output)
}

func (c *OPClient) GetPassword(itemName, field string) (string, error) {
	cmd := exec.Command("op", "item", "get", itemName, "--vault", c.vault, "--fields", field, "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("op item get failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (c *OPClient) ListItems() ([]string, error) {
	cmd := exec.Command("op", "item", "list", "--vault", c.vault, "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("op item list failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, fmt.Errorf("failed to parse items: %w", err)
	}

	names := make([]string, len(items))
	for i, item := range items {
		if title, ok := item["title"].(string); ok {
			names[i] = title
		}
	}

	return names, nil
}

func (c *OPClient) IsAvailable() bool {
	cmd := exec.Command("op", "--version")
	return cmd.Run() == nil
}

func parseOPItem(data []byte) (*OPItem, error) {
	var item map[string]interface{}
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("failed to parse item: %w", err)
	}

	id, _ := item["id"].(string)
	return &OPItem{
		ID:   id,
		JSON: item,
	}, nil
}
