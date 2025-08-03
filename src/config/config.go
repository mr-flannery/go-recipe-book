package config

import (
	"os"

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
	} `yaml:"db"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
}

var (
	config Config
)

func GetConfig() (Config, error) {
	var err error

	if config != (Config{}) {
		return config, nil
	}

	// unsure if the path makes sense once I start packaging this into a docker container
	file, err := os.Open("../config.yaml")
	if err != nil {
		return Config{}, err
	}

	decoder := yaml.NewDecoder(file)
	cfg := Config{}
	err = decoder.Decode(&cfg)
	if err != nil {
		return Config{}, err
	}

	config = cfg

	err = file.Close()
	if err != nil {
		return Config{}, err
	}

	return config, err
}
