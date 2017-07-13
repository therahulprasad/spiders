package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var config Configuration

// Loads configuration from json file
func Load(path string) error {
	// Open file
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Set default values
	config.DisplayMatchedUrl = false

	// Decode YAML
	yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return err
	}

	// Validate Configuration
	if !(config.ProjectType == PROJECT_TYPE_CRAWL || config.ProjectType == PROJECT_TYPE_BACTH)  {
		log.Fatal("Config:project_type - Only " + PROJECT_TYPE_BACTH + " & " + PROJECT_TYPE_CRAWL + " is supported.")
	}

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