package crawler

import (
	"os"
	"github.com/therahulprasad/spiderman/crawler/db"
	"github.com/therahulprasad/spiderman/crawler/config"
	"log"
)


// Implements basic system setup
func Setup(config_path string, resume bool) config.Configuration {
	// Load config file
	err := config.Load(config_path)
	if err != nil {
		// Die if there is an error
		log.Fatal(err.Error())
	}

	// get Configurations
	configuration, err := config.Get()
	if err != nil {
		log.Fatal(err.Error())
	}

	if configuration.Directory == "" {
		log.Fatal("Directory name should not be empty")
	}
	// Check if directory does not exists then create it
	if _, err = os.Stat(configuration.Directory); os.IsNotExist(err) {
		err := os.Mkdir(configuration.Directory, os.FileMode(0777))
		if err != nil {
			log.Fatal("Could not create directory: " + configuration.Directory)
		}
	}

	// If directory already exists die
	// TODO: If resume flag is set then don't check this
	if !resume && err == nil {
		log.Fatal("Directory already exists: " + configuration.Directory)
	}


	// Create "data" folder in the directory
	// TODO: If resume flag is set then data directory must already exists
	err = os.Mkdir(configuration.DataDir(), os.FileMode(0777))
	if !resume && err != nil {
		log.Fatal("Could not create data directory")
	}

	if configuration.WebCount <= 0 {
		log.Fatal("web_count must be more than 0")
	}

	// Initiate database in the directory
	// TODO: If resume flag is set then Database must already be present
	db.Setup(configuration, resume)

	return configuration
}
