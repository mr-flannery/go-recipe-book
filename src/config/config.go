package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/utils"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DB struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
		SSLMode  string `yaml:"sslmode"`
		Admin    struct {
			Username string `yaml:"username"`
			Email    string `yaml:"email"`
			Password string `yaml:"password"`
		} `yaml:"admin"`
	} `yaml:"db"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Environment struct {
		Mode string `yaml:"mode"`
	} `yaml:"environment"`
	Api struct {
		Keys []string `yaml:"keys"`
	} `yaml:"api"`
	Mail struct {
		Domain string `yaml:"domain"`
		ApiKey string `yaml:"api_key"`
	} `yaml:"mail"`
}

var (
	config Config
)

func GetConfig() Config {
	var err error

	if !reflect.DeepEqual(config, Config{}) {
		return config
	}

	cfg := Config{}

	configPath := filepath.Join(utils.GetBasePath(), "config.yaml")
	file, err := os.Open(configPath)
	if err != nil {
		slog.Warn("Could not open config file, using environment variables only", "path", configPath, "error", err)
	} else {
		decoder := yaml.NewDecoder(file)
		err = decoder.Decode(&cfg)
		if err != nil {
			slog.Warn("Failed to decode config file, using environment variables only", "error", err)
		}
		file.Close()
	}

	applyEnvOverrides(&cfg)

	config = cfg
	return config
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.DB.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.DB.Port = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.DB.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.DB.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.DB.Name = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		cfg.DB.SSLMode = v
	}
	if v := os.Getenv("DB_ADMIN_USERNAME"); v != "" {
		cfg.DB.Admin.Username = v
	}
	if v := os.Getenv("DB_ADMIN_EMAIL"); v != "" {
		cfg.DB.Admin.Email = v
	}
	if v := os.Getenv("DB_ADMIN_PASSWORD"); v != "" {
		cfg.DB.Admin.Password = v
	}
	if v := os.Getenv("ENVIRONMENT_MODE"); v != "" {
		cfg.Environment.Mode = v
	}
	if v := os.Getenv("API_KEYS"); v != "" {
		cfg.Api.Keys = strings.Split(v, ",")
	}
	if v := os.Getenv("MAIL_DOMAIN"); v != "" {
		cfg.Mail.Domain = v
	}
	if v := os.Getenv("MAIL_API_KEY"); v != "" {
		cfg.Mail.ApiKey = v
	}
}
