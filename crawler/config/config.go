package config

import (
	"errors"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

var config Configuration

// Load configuration from json file
func Load(path string) error {
	// Open file
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Set default values
	config.DisplayMatchedURL = false

	// Decode YAML
	yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return err
	}

	// Validate supported Project Type
	if !(config.ProjectType == PROJECTTYPECRAWL || config.ProjectType == PROJECTTYPEBACTH) {
		log.Fatal("Config:project_type - Only " + PROJECTTYPEBACTH + " & " + PROJECTTYPECRAWL + " is supported.")
	}

	// Validate supported Content Holder
	if !(config.ContentHolder == CONTENTHOLDERTEXT || config.ContentHolder == CONTENTHOLDERATTR || config.ContentHolder == CONTENTHOLDERHTML) {
		log.Fatal("Config:content_holder - Only " + CONTENTHOLDERHTML + ", " + CONTENTHOLDERTEXT + " & " + CONTENTHOLDERATTR + " is supported.")
	}

	// There is nothing to validate now :(
	return nil
}

// Get gets configuration object if loaded otherwise returns error
func Get() (Configuration, error) {
	if (Configuration{}) == config {
		return Configuration{}, errors.New("Configuration not loaded")
	}
	return config, nil
}
