package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/therahulprasad/spiders/crawler"
)

// VERSION it will be printed when -version flag is used
const VERSION = "0.4"

// MAJORCHANGE Whats new in this version, it will be printed when -version flag is used
const MAJORCHANGE = "Batch Processing"

// const DEFAULT_CONFIG_FILENAME = "config.yaml"

func main() {
	verFlag := flag.Bool("v", false, "To get current versiion")
	configPathFlag := flag.String("c", "", "Path of json config file")
	resumeFlag := flag.Bool("r", false, "Resume incomplete project")
	flag.Parse()
	if *configPathFlag == "" {
		fmt.Println("Config flag is mising")
		os.Exit(1)
	}
	if *verFlag == true {
		fmt.Println("Multi threaded web crawler to collect text")
		fmt.Println("Version: " + VERSION)
		os.Exit(0)
	}

	configPath := *configPathFlag
	resume := *resumeFlag

	// Channel which waits for Crawler to end
	chExitWait := make(chan bool)

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
	crawler.Initialize(configPath, chExitWait, chKill, resume)

	// Wait for Crawler to end
	<-chExitWait

	fmt.Println("Bye")
}
