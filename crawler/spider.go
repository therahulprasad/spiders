package crawler

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/therahulprasad/spiders/crawler/config"
	"github.com/therahulprasad/spiders/crawler/db"
)

func terminateLinkProcessor(killLinkProcessor, killLinkProcessorAck chan bool) {
	fmt.Println("Terminating link processor")

	// Kill link processor
	killLinkProcessor <- true

	// Wait for goroutines to be killed
	<-killLinkProcessorAck
	//
	// // Exit
	// chWaitForExit <- true
}

func sigKillHandller(chKill, killLinkProcessor, killLinkProcessorAck, killAllWorker, killAllWorkerAck, chWaitForExit chan bool) {
	<-chKill

	fmt.Println("Sigterm received, Killing all workers")
	killAllWorker <- true // This will kill all workers and then terminate link processor

	// wait for all workers to exit
	<-killAllWorkerAck

	terminateLinkProcessor(killLinkProcessor, killLinkProcessorAck)

	// Send chWaitForExit (This will terminate the program)
	// This should only ne sent when all groutines has been killed
	chWaitForExit <- true
}

// Initialize Initializes Crawler
func Initialize(path string, chWaitForExit, chKill chan bool, resume bool) {
	// Setup system
	configuration := Setup(path, resume)

	// Kill switch for link processor
	killLinkProcessor := make(chan bool)
	killLinkProcessorAck := make(chan bool)

	// Single switch to kill all workers
	killAllWorker := make(chan bool)
	killAllWorkerAck := make(chan bool)

	// Kill workers individually
	killWorker := make(chan int)
	// Worker death acknowledgement
	// chWorkerCount := make(chan int)

	// Handle Ctrl + C
	go sigKillHandller(chKill, killLinkProcessor, killLinkProcessorAck, killAllWorker, killAllWorkerAck, chWaitForExit)

	// Channel for passing Documents
	docs_channel := make(chan *goquery.Document)

	// Handle resuming on completed project
	if resume == true && db.CountRemainingRows() == 0 {
		fmt.Println("Project is complete. There is no new link to crawl in database.")
		os.Exit(0)
	}

	processDoc := true
	// if project_type is crawl then push root URL otherwise just ignore it Link_processor will take care of batches
	if configuration.ProjectType == config.PROJECT_TYPE_CRAWL {
		// First element is required only when starting a new project so ignore when a project is being resumed
		if resume == false {
			_, err := db.Push(configuration.RootURL, 0)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	} else if configuration.ProjectType == config.PROJECT_TYPE_BACTH {
		if resume != true {
			push_all_batch_links(configuration)
		}
		processDoc = false
		// go fakeLinkProcessor(killLinkProcessor, killLinkProcessorAck)
	}

	// Start 1 link processor goroutine
	go linkProcessor(docs_channel, configuration, killLinkProcessor, killLinkProcessorAck, processDoc)

	// Start workers for processing pages
	numWorkers := configuration.WebCount
	for i := 0; i < numWorkers; i++ {
		go processPageWorker(configuration, docs_channel, killWorker)
	}
	go workerManager(killWorker, numWorkers, killAllWorker, killAllWorkerAck)
	// go processFirstPage(configuration, docs_channel)
}

// If crawling for the first time, DB will just contain single link
// TODO: Is it needed ? Why wont the workers start processing first link ?
// func processFirstPage(configuration config.Configuration, chDocs chan *goquery.Document) {
// 	// Process first page
// 	row, err := db.Pop()
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
//
// 	err = pageProcessor(row.Id, row.Link, configuration, chDocs)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// }

func workerManager(killWorker chan int, workerCount int, killAllWorker, killAllWorkerAck chan bool) {
	for {
		select {
		case <-killAllWorker:
			// Request to kill all Workers received
			// Kill remaining workers
			fmt.Println(workerCount)
			for i := workerCount; i > 0; i-- {
				// Kill a worker
				killWorker <- 1

				// // Get the acknowledgement
				// <-chWorkerCount
			}

			// All workers are killed. Now return
			killAllWorkerAck <- true
			return
			// // When all workers are killed, Terminate
			// terminateLinkProcessor(killLinkProcessor, killLinkProcessorAck, chWaitForExit)
		}
	}
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}
func processPageWorker(configuration config.Configuration, chLinkProcessor chan *goquery.Document, killWorker chan int) {
	for {
		select {
		case <-killWorker:
			if configuration.Debug == true {
				fmt.Println("Killed")
			}
			return
		default:
			// Wait for random time betweeb 1 - 5 seconds
			time.Sleep(time.Duration(random(1, 5000)) * time.Millisecond)

			node, err := db.Pop()
			if err != nil {
				log.Println("Error while db.Pop @ process_page: " + err.Error())
			}

			// if empty value is returned. Wait for 10 seconds and try again
			if node == (db.Node{}) {
				if configuration.Debug == true {
					fmt.Println("Empty Node :(")
				}

				// // Check if database contains 0 new nodes
				// if db.CountRemainingRows() == 0 {
				// 	killWorker <- 1
				// }

				// Wait for random time
				time.Sleep(time.Duration(random(1, 5000)) * time.Millisecond)
			} else {
				pageProcessor(node.Id, node.Link, configuration, chLinkProcessor)
			}
		}
	}
}

func pageProcessor(id int64, url string, configuration config.Configuration, chLinkProcessor chan *goquery.Document) error {
	fmt.Println("Processing page: " + url)
	// Open the url
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return err
	}

	// pageValidator := configuration.PageValidator
	// if (configuration.PageValidator) == "" {
	// 	pageValidator = ".*"
	// }

	// Check if page contains article
	article := doc.Find(configuration.PageValidator)

	if article.Length() != 0 {
		// Article found
		if configuration.ContentHolder == config.CONTENT_HOLDER_TEXT {
			// copy text
			scrapText(id, doc, configuration)
		} else if configuration.ContentHolder == config.CONTENT_HOLDER_ATTR {
			// copy attribute details
			scrapTagAttr(id, doc, configuration)
		} else {
			log.Fatal("Should not reach here because configuration.ContentHolder is already validated when config was loaded")
		}

	} else {
		db.Update(id, 0, db.ValidationFailed, "")
	}
	// Process all links in the doc
	chLinkProcessor <- doc
	//ls_url(doc, configuration)

	return nil
}

// Scraps data and create a text files for each selector
func scrapText(id int64, doc *goquery.Document, configuration config.Configuration) {
	if configuration.Debug {
		fmt.Println("scrapText")
	}

	var count int64 = 0
	strAll := ""
	selections := doc.Find(configuration.ContentSelector)
	selections.Each(func(i int, s *goquery.Selection) {
		count++
		str := s.Text()
		str = strings.TrimSpace(str)
		strAll = strAll + str
		ioutil.WriteFile(configuration.DataDir()+"/"+strconv.FormatInt(id, 10)+"_"+strconv.FormatInt(count, 10)+".txt", []byte(str), os.FileMode(0777))
	})

	md5hash := md5.Sum([]byte(strAll))

	// Update database that the link is scrapped
	err := db.Update(id, count, db.Success, fmt.Sprintf("%x", md5hash))
	if err != nil {
		log.Println(err.Error())
	}
}

// Scraps attributes and create a text file for each page
func scrapTagAttr(id int64, doc *goquery.Document, configuration config.Configuration) {
	if configuration.Debug {
		fmt.Println("scrapTag")
	}

	var count int64 = 0
	strAll := ""
	selections := doc.Find(configuration.ContentSelector)
	selections.Each(func(i int, s *goquery.Selection) {
		count++
		str, exists := s.Attr(configuration.ContentTagAttr)
		if exists {
			// Append links using newline
			if strAll == "" {
				strAll = str
			} else {
				strAll = strAll + "\n" + str
			}
		}
	})

	// Create a single file for each page
	ioutil.WriteFile(configuration.DataDir()+"/"+strconv.FormatInt(id, 10)+".txt", []byte(strAll), os.FileMode(0777))
	md5hash := md5.Sum([]byte(strAll))

	// Update database that the link is scrapped
	err := db.Update(id, count, db.Success, fmt.Sprintf("%x", md5hash))
	if err != nil {
		log.Println(err.Error())
	}
}
