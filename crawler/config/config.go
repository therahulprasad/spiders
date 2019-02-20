package config

import (
	"errors"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
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

	// Validate supported Project Type
	if !(config.ProjectType == PROJECT_TYPE_CRAWL || config.ProjectType == PROJECT_TYPE_BACTH) {
		log.Fatal("Config:project_type - Only " + PROJECT_TYPE_BACTH + " & " + PROJECT_TYPE_CRAWL + " is supported.")
	}

	// Validate supported Content Holder
	if !(config.ContentHolder == CONTENT_HOLDER_TEXT || config.ContentHolder == CONTENT_HOLDER_ATTR) {
		log.Fatal("Config:content_holder - Only " + CONTENT_HOLDER_TEXT + " & " + CONTENT_HOLDER_ATTR + " is supported.")
	}

	// There is nothing to validate now :(
	return nil
}

// Returns configuration object if loaded otherwise returns error
func Get() (Configuration, error) {
	if (Configuration{}) == config {
		return Configuration{}, errors.New("Configuration not loaded")
	}
	return config, nil
}
