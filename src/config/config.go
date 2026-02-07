package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

	"gopkg.in/yaml.v3"
)

// getPackageDir returns the directory containing this source file
func getPackageDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}
	return filepath.Dir(filename)
}

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

	// config.yaml is located at the project root (parent of src/)
	configPath := filepath.Join(getPackageDir(), "..", "..", "config.yaml")
	file, err := os.Open(configPath)
	if err != nil {
		panic("Failed to open config file: " + err.Error())
	}

	decoder := yaml.NewDecoder(file)
	cfg := Config{}
	err = decoder.Decode(&cfg)
	if err != nil {
		panic("Failed to decode config file: " + err.Error())
	}

	config = cfg

	err = file.Close()
	if err != nil {
		slog.Warn("Failed to close config file", "error", err)
	}

	return config
}
