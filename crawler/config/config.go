package config

import (
	"errors"
	"os"
	"encoding/json"
)

var config Configuration

// Loads configuration from json file
func Load(path string) error {
	// Open file
	fp, err := os.Open(path)
	if err != nil {
		return err
	}

	// Set default values
	config.DisplayMatchedUrl = false

	// Decode json
	decoder := json.NewDecoder(fp)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}

	// Validate Configuration
	// There is nothing to validate now :(
	return nil
}

// Returns configuration object if loaded otherwise returns error
func Get() (Configuration, error) {
	if (Configuration {}) == config {
		return Configuration {}, errors.New("Configuration not loaded")
	}
	return config, nil
}