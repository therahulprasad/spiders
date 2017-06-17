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
	"time"
	"crypto/md5"
)

func terminate(killLinkProcessor, killLinkProcessorAck, ch_exit chan bool) {
	fmt.Println("Waiting for all workers to die")

	// Kill link processor
	killLinkProcessor<- true

	// Wait for goroutines to be killed
	<-killLinkProcessorAck

	// Exit
	ch_exit<- true
}

func handle_sig_kill(chQuit, killLinkProcessor, killLinkProcessorAck, ch_exit chan bool, killAllWorker chan int) {
	<-chQuit
	fmt.Println("Sigterm received, Killing all workers")
	terminate(killLinkProcessor, killLinkProcessorAck, ch_exit)
}

func Initialize(path string, chWaitForExit, ch_kill chan bool, resume bool) {
	// Setup system
	configuration := Setup(path, resume)

	// Kill switch for link processor
	killLinkProcessor := make(chan bool)
	killLinkProcessorAck := make(chan bool)

	// Single switch to kill all workers
	killAllWorker := make(chan int)

	// Kill workers individually
	killWorker := make(chan int)
	// Worker death acknowledgement
	chWorkerCount := make(chan int)

	// Handle Ctrl + C
	go handle_sig_kill(ch_kill, killLinkProcessor, killLinkProcessorAck, chWaitForExit, killAllWorker)

	if resume == false {
		_, err := db.Push(configuration.RootURL, 0)
		if err != nil { log.Fatal(err.Error())}
	}

	// Start 1 link processor goroutine
	docs_channel := make(chan *goquery.Document)
	go link_processor(docs_channel, configuration, killLinkProcessor, killLinkProcessorAck)

	// Process first page
	node, err := db.Pop()
	if err != nil { log.Fatal(err.Error())}

	err = page_processor(node.Id, node.Link, configuration, docs_channel)
	if err != nil { log.Fatal(err.Error())}

	// Start workers for processing pages
	num_workers := configuration.WebCount
	for i:=0; i<num_workers;i++ {
		go worker_process_page(configuration, docs_channel, chWorkerCount, killWorker)
	}

	go worker_manager(chWorkerCount, killWorker, killAllWorker, num_workers, killLinkProcessor, killLinkProcessorAck, chWaitForExit)
}

func worker_manager(chWorkerCount, killWorker, killAllWorker chan int, num_workers int, killLinkProcessor, killLinkProcessorAck, ch_exit chan bool) {
	for {
		select {
		case <-chWorkerCount:
			fmt.Println("One worker is dead")
			// For every worker who died decrease worker count
			num_workers--

			// If all workers are dead, then terminate program
			if num_workers == 0 {
				terminate(killLinkProcessor, killLinkProcessorAck, ch_exit)
			}
		case <-killAllWorker:
			// Request to kill all Workers received
			// Kill remaining workers
			for i:=num_workers; i<=0; i-- {
				// Kill a worker
				killWorker<- 1

				// Get the acknowledgement
				<-chWorkerCount
			}

			// When all workers are killed, Terminate
			terminate(killLinkProcessor, killLinkProcessorAck, ch_exit)
		}
	}
}

func worker_process_page(configuration config.Configuration, ch_link_processor chan *goquery.Document, chWorkerCount, killWorker chan int) {
	emptyCount := 0
	for {
		select {
		case <-killWorker:
			return
		default:
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
	} else {
		db.Update(id, 0, db.ValidationFailed, "")
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
	strAll := ""
	selections := doc.Find(configuration.ContentSelector)
	selections.Each(func(i int, s *goquery.Selection) {
		count++
		str := s.Text()
		strAll = strAll + str
		ioutil.WriteFile(configuration.DataDir() + "/" + strconv.FormatInt(id, 10) + "_" + strconv.FormatInt(count, 10) + ".txt", []byte(str), os.FileMode(0777))
	})

	md5hash := md5.Sum([]byte(strAll))

	// Update database that the link is scrapped
	err := db.Update(id, count, db.Success, fmt.Sprintf("%x", md5hash))
	if err != nil {log.Println(err.Error())}
}