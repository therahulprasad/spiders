package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

var config Configuration

func Load(path string) error {
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	config.DisplayMatchedURL = false

	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return validateConfig()
}

func validateConfig() error {
	if config.ProjectType != PROJECTTYPECRAWL && config.ProjectType != PROJECTTYPEBATCH {
		return fmt.Errorf("invalid project_type '%s': must be '%s' or '%s'", 
			config.ProjectType, PROJECTTYPECRAWL, PROJECTTYPEBATCH)
	}

	validContentHolders := []string{CONTENTHOLDERTEXT, CONTENTHOLDERATTR, CONTENTHOLDERHTML}
	for _, valid := range validContentHolders {
		if config.ContentHolder == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid content_holder '%s': must be one of %v", 
		config.ContentHolder, validContentHolders)
}

func Get() (Configuration, error) {
	if config.ProjectType == "" {
		return Configuration{}, errors.New("configuration not loaded - call Load() first")
	}
	return config, nil
}
