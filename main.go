package main

import "github.com/therahulprasad/spiderman/crawler"

func main() {
	config_path := "config.json"

	// Channel which waits for Crawler to end
	ch_exit := make(chan bool)

	// Start the crawler
	crawler.Process(config_path, ch_exit)

	// Wait for Crawler to end
	<-ch_exit
}

