package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// configFileName is the name of the db configuration file stored in the user's home directory
	configFileName = ".gatorconfig.json"
	// configFileMode defines the file permissions (read/write for owner, read for group and others)
	configFileMode = 0644
)

// Config represents the application configuration structure
// that is serialized to and deserialized from the config file
type Config struct {
	DBUrl           string `json:"db_url"`            // Database connection URL
	CurrentUsername string `json:"current_user_name"` // Currently active username
}

// Read loads the configuration from the config file
// Returns a pointer to the Config struct and any error encountered
func Read() (*Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("cannot get path: %v", err)
	}
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("file cannot read (%s): %v", path, err)
	}

	var config Config

	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		return nil, fmt.Errorf("JSON Unmarshal error: %v", err)
	}

	return &config, nil
}

// SetUser updates the current username in the configuration
// and persists the changes to the config file
func (cfg *Config) SetUser(username string) error {
	cfg.CurrentUsername = username
	return write(cfg)
}

// write persists the configuration to the config file
// it serializes the Config struct to JSON with indentation
func write(cfg *Config) error {
	// Convert config to formatted JSON
	jsonDataIndent, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON MarshalIndent error: %v", err)
	}

	// Get the full path to the config file
	path, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("cannot get config file path: %v", err)
	}

	// Write the JSON data to the config file
	err = os.WriteFile(path, jsonDataIndent, configFileMode)
	if err != nil {
		return fmt.Errorf("cannot write config file: %v", err)
	}

	return nil
}

// getConfigFilePath returns the full path to the configuration file
// located in the user's home directory
func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get home directory: %v", err)
	}

	fullFilePath := filepath.Join(homeDir, configFileName)

	return fullFilePath, nil
}
