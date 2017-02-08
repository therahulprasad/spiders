package main

import "github.com/therahulprasad/spiderman/crawler"

func main() {
	config_path := "config.json"
	//crawler.Process("https://www.sampada.net/ಮಲೆನಾಡಿನ-ಮಾಳ-ಕಾವಲು-ಕೊನೆಗೆ-ಕಂಬಳ-ಭಾಗ-1/47442", config_path)
	ch_exit := make(chan bool)

	crawler.Process(config_path, ch_exit)

	<-ch_exit
}

