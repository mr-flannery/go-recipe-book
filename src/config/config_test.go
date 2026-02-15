package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfig_StructHasExpectedFields(t *testing.T) {
	cfg := Config{}

	if cfg.DB.Host != "" {
		t.Error("DB.Host should be empty by default")
	}
	if cfg.Server.Port != 0 {
		t.Error("Server.Port should be 0 by default")
	}
	if cfg.Environment.Mode != "" {
		t.Error("Environment.Mode should be empty by default")
	}
	if cfg.Api.Keys != nil {
		t.Error("Api.Keys should be nil by default")
	}
	if cfg.Mail.Domain != "" {
		t.Error("Mail.Domain should be empty by default")
	}
}

func TestConfig_YAMLDecoding_ParsesValidConfig(t *testing.T) {
	yamlContent := `
db:
  host: localhost
  port: "5432"
  user: testuser
  password: testpass
  name: testdb
  sslmode: disable
  admin:
    username: admin
    email: admin@test.com
    password: adminpass
server:
  port: 8080
environment:
  mode: development
api:
  keys:
    - key1
    - key2
mail:
  domain: mail.example.com
  api_key: mailkey123
`

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlContent), &cfg)
	if err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	if cfg.DB.Host != "localhost" {
		t.Errorf("expected DB.Host = localhost, got %s", cfg.DB.Host)
	}
	if cfg.DB.Port != "5432" {
		t.Errorf("expected DB.Port = 5432, got %s", cfg.DB.Port)
	}
	if cfg.DB.User != "testuser" {
		t.Errorf("expected DB.User = testuser, got %s", cfg.DB.User)
	}
	if cfg.DB.Password != "testpass" {
		t.Errorf("expected DB.Password = testpass, got %s", cfg.DB.Password)
	}
	if cfg.DB.Name != "testdb" {
		t.Errorf("expected DB.Name = testdb, got %s", cfg.DB.Name)
	}
	if cfg.DB.SSLMode != "disable" {
		t.Errorf("expected DB.SSLMode = disable, got %s", cfg.DB.SSLMode)
	}
	if cfg.DB.Admin.Username != "admin" {
		t.Errorf("expected DB.Admin.Username = admin, got %s", cfg.DB.Admin.Username)
	}
	if cfg.DB.Admin.Email != "admin@test.com" {
		t.Errorf("expected DB.Admin.Email = admin@test.com, got %s", cfg.DB.Admin.Email)
	}
	if cfg.DB.Admin.Password != "adminpass" {
		t.Errorf("expected DB.Admin.Password = adminpass, got %s", cfg.DB.Admin.Password)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected Server.Port = 8080, got %d", cfg.Server.Port)
	}
	if cfg.Environment.Mode != "development" {
		t.Errorf("expected Environment.Mode = development, got %s", cfg.Environment.Mode)
	}
	if len(cfg.Api.Keys) != 2 {
		t.Errorf("expected 2 API keys, got %d", len(cfg.Api.Keys))
	}
	if cfg.Api.Keys[0] != "key1" {
		t.Errorf("expected first API key = key1, got %s", cfg.Api.Keys[0])
	}
	if cfg.Mail.Domain != "mail.example.com" {
		t.Errorf("expected Mail.Domain = mail.example.com, got %s", cfg.Mail.Domain)
	}
	if cfg.Mail.ApiKey != "mailkey123" {
		t.Errorf("expected Mail.ApiKey = mailkey123, got %s", cfg.Mail.ApiKey)
	}
}

func TestConfig_YAMLDecoding_HandlesPartialConfig(t *testing.T) {
	yamlContent := `
db:
  host: localhost
server:
  port: 3000
`

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlContent), &cfg)
	if err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	if cfg.DB.Host != "localhost" {
		t.Errorf("expected DB.Host = localhost, got %s", cfg.DB.Host)
	}
	if cfg.DB.Port != "" {
		t.Errorf("expected DB.Port to be empty, got %s", cfg.DB.Port)
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("expected Server.Port = 3000, got %d", cfg.Server.Port)
	}
	if cfg.Environment.Mode != "" {
		t.Errorf("expected Environment.Mode to be empty, got %s", cfg.Environment.Mode)
	}
}

func TestConfig_YAMLDecoding_HandlesEmptyConfig(t *testing.T) {
	yamlContent := ``

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlContent), &cfg)
	if err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	if cfg.DB.Host != "" {
		t.Errorf("expected DB.Host to be empty, got %s", cfg.DB.Host)
	}
	if cfg.Server.Port != 0 {
		t.Errorf("expected Server.Port to be 0, got %d", cfg.Server.Port)
	}
}

func TestConfig_YAMLFileDecoding_ReadsFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	yamlContent := `
db:
  host: filehost
  port: "5433"
server:
  port: 9090
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	file, err := os.Open(configPath)
	if err != nil {
		t.Fatalf("failed to open test config: %v", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	var cfg Config
	err = decoder.Decode(&cfg)
	if err != nil {
		t.Fatalf("failed to decode config: %v", err)
	}

	if cfg.DB.Host != "filehost" {
		t.Errorf("expected DB.Host = filehost, got %s", cfg.DB.Host)
	}
	if cfg.DB.Port != "5433" {
		t.Errorf("expected DB.Port = 5433, got %s", cfg.DB.Port)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected Server.Port = 9090, got %d", cfg.Server.Port)
	}
}
