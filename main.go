package main

import (
	"github.com/therahulprasad/spiderman/crawler"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"crypto/md5"
)

const VERSION = "0.2"
const MAJOR_CHANGE = "Initial Release"
func main() {
	//test()

	verFlag := flag.Bool("version", false, "To get current versiion")
	configPathFlag := flag.String("config", "config.json", "Path of json config file")
	flag.Parse()
	if *verFlag == true {
		fmt.Println("Version: " + VERSION)
		fmt.Println(MAJOR_CHANGE)
		os.Exit(0)
	}

	config_path := *configPathFlag

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
	crawler.Process(config_path, ch_exit_wait, chKill)

	// Wait for Crawler to end
	<-ch_exit_wait
}


func test() {
	sum := md5.Sum([]byte("hello"))
	sumStr := string(sum[:])
	sumStr2 := fmt.Sprintf("%x", sum)

	fmt.Printf("%x", sum)
	fmt.Println(sumStr)
	fmt.Println(sumStr2)

	os.Exit(0)
}
