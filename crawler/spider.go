package crawler

import (
	"fmt"
	"github.com/therahulprasad/spiderman/crawler/config"
	"github.com/therahulprasad/spiderman/crawler/db"
	"github.com/PuerkitoBio/goquery"
	"os"
	"io/ioutil"
	"strconv"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func terminate(killLinkProcessor, killLinkProcessorAck, ch_exit chan bool) {
	// Kill link processor
	killLinkProcessor<- true

	// Wait for goroutes to be killed
	fmt.Println("Waiting for all workers to die")
	<-killLinkProcessorAck

	// Exit
	ch_exit<- true
}

func handle_sig_term(chQuit chan os.Signal, killLinkProcessor, killLinkProcessorAck, ch_exit chan bool) {
	<-chQuit
	fmt.Println("Sigterm received, Killing all workers")
	terminate(killLinkProcessor, killLinkProcessorAck, ch_exit)
}

func Process(path string, ch_exit chan bool) {
	// Setup system
	configuration := Setup(path)

	// Kill switch for link processor
	killLinkProcessor := make(chan bool)
	killLinkProcessorAck := make(chan bool)

	// Handle Ctrl + C
	chQuit := make(chan os.Signal, 2)
	signal.Notify(chQuit, os.Interrupt, syscall.SIGTERM)
	go handle_sig_term(chQuit, killLinkProcessor, killLinkProcessorAck, ch_exit)

	// TODO: In case of resume operation, skip this
	_, err := db.Push(configuration.RootURL, 0)
	if err != nil { log.Fatal(err.Error())}

	// Start 1 link processor goroutine
	docs_channel := make(chan *goquery.Document)
	go link_processor(docs_channel, configuration, killLinkProcessor, killLinkProcessorAck)

	// Process first page
	node, err := db.Pop()
	if err != nil { log.Fatal(err.Error())}

	err = page_processor(node.Id, node.Link, configuration, docs_channel)
	if err != nil { log.Fatal(err.Error())}

	// Start 10 workers for processing pages
	chWorkerCount := make(chan int)

	num_workers := 10
	for i:=0; i<num_workers;i++ {
		go process_page(configuration, docs_channel, chWorkerCount)
	}

	for range chWorkerCount {
		num_workers--

		if num_workers == 0 {
			terminate(killLinkProcessor, killLinkProcessorAck, ch_exit)
		}
	}
	// Pop new url from queue
	// Send the url to workers
	// Check if exit condition is reached
}

func process_page(configuration config.Configuration, ch_link_processor chan *goquery.Document, chWorkerCount chan int) {
	emptyCount := 0
	for {
		node, err := db.Pop()
		if err != nil {
			log.Println("Error while db.Pop @ process_page: " + err.Error())
			chWorkerCount<- 1
			return
		}

		// if empty value is returned. Wait for 10 seconds and try again
		if node == (db.Node{}) {
			fmt.Println("Empty Node :(")
			emptyCount++
			time.Sleep(time.Duration(10) * time.Second)
		} else {
			emptyCount = 0
			page_processor(node.Id, node.Link, configuration, ch_link_processor)
		}

		// If value is empty for 10 consecutive period then return
		if emptyCount == 10 {
			chWorkerCount<- 1
			return
		}
	}
}

func page_processor(id int64, url string, configuration config.Configuration, ch_link_processor chan *goquery.Document) (error) {
	fmt.Println("Processing page: " + url)
	// Open the url
	doc, err := goquery.NewDocument(url)
	if err != nil {return err}

	// Check if page contains article
	article:=doc.Find(configuration.PageValidator)

	if article.Length() != 0 {
		// Article found copy text
		scrap(id, doc, configuration)
	}

	// Process all links in the doc
	ch_link_processor <- doc
	//ls_url(doc, configuration)

	return nil
}

// Scraps data and create a text files
func scrap(id int64, doc *goquery.Document, configuration config.Configuration) {
	if configuration.Debug {
		fmt.Println("scrap")
	}

	var count int64 = 0
	selections := doc.Find(configuration.ContentSelector)
	selections.Each(func(i int, s *goquery.Selection) {
		count++
		str := s.Text()
		ioutil.WriteFile(configuration.DataDir() + "/" + strconv.FormatInt(id, 10) + "_" + strconv.FormatInt(count, 10) + ".txt", []byte(str), os.FileMode(0777))
	})

	// Update database that the link is scrapped
	db.Update(id, count, "success")
}