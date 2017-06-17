package main

import (
	"github.com/therahulprasad/spiderman/crawler"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const VERSION = "0.2"
const MAJOR_CHANGE = "Initial Release"
const DEFAULT_CONFIG_FILENAME = "config.yaml"

func main() {
	verFlag := flag.Bool("version", false, "To get current versiion")
	configPathFlag := flag.String("config", DEFAULT_CONFIG_FILENAME, "Path of json config file")
	resumeFlag := flag.Bool("resume", false, "Resume incomplete project")
	flag.Parse()
	if *verFlag == true {
		fmt.Println("Version: " + VERSION)
		fmt.Println(MAJOR_CHANGE)
		os.Exit(0)
	}

	config_path := *configPathFlag
	resume := *resumeFlag

	// Channel which waits for Crawler to end
	ch_exit_wait := make(chan bool)

	// Kill signal for workers
	chKill := make(chan bool)

	chQuit := make(chan os.Signal, 2)
	signal.Notify(chQuit, os.Interrupt, syscall.SIGTERM)

	// Ctrl + C handler
	go func(chQuit chan os.Signal, ch_kill chan bool) {
		// Wait to Ctrl + C
		<-chQuit

		// send kill signal to processor
		chKill <- true
	}(chQuit, chKill)

	// Start the crawler
	crawler.Initialize(config_path, ch_exit_wait, chKill, resume)

	// Wait for Crawler to end
	<-ch_exit_wait
}