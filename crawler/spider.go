package crawler

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/therahulprasad/spiders/crawler/config"
	"github.com/therahulprasad/spiders/crawler/db"
)

func terminateLinkProcessor(killLinkProcessor, killLinkProcessorAck chan bool) {
	log.Println("Terminating link processor")
	killLinkProcessor <- true
	<-killLinkProcessorAck
}

func shutdownHandler(chKill, killLinkProcessor, killLinkProcessorAck, killAllWorker, killAllWorkerAck, chWaitForExit chan bool) {
	<-chKill

	log.Println("Shutdown signal received, stopping all workers")
	killAllWorker <- true

	<-killAllWorkerAck
	terminateLinkProcessor(killLinkProcessor, killLinkProcessorAck)

	chWaitForExit <- true
}

type Crawler struct {
	config     config.Configuration
	docsChan   chan *goquery.Document
	killWorker chan int
	numWorkers int
}

func NewCrawler(cfg config.Configuration) *Crawler {
	return &Crawler{
		config:     cfg,
		docsChan:   make(chan *goquery.Document),
		killWorker: make(chan int),
		numWorkers: cfg.WebCount,
	}
}

func (c *Crawler) initializeURLs(resume bool) error {
	if c.config.ProjectType == config.PROJECTTYPECRAWL {
		if !resume {
			_, err := db.Push(c.config.RootURL, 0)
			if err != nil {
				return fmt.Errorf("failed to push root URL: %w", err)
			}
		}
	} else if c.config.ProjectType == config.PROJECTTYPEBATCH {
		if !resume {
			if err := pushAllBatchLinks(c.config); err != nil {
				return fmt.Errorf("failed to push batch URLs: %w", err)
			}
		}
	}
	return nil
}

func (c *Crawler) startWorkers(killWorker chan int) {
	for i := 0; i < c.numWorkers; i++ {
		go processPageWorker(c.config, c.docsChan, killWorker)
	}
}

func Initialize(configPath string, chWaitForExit, chKill chan bool, resume bool) {
	configuration := Setup(configPath, resume)

	killLinkProcessor := make(chan bool)
	killLinkProcessorAck := make(chan bool)
	killAllWorker := make(chan bool)
	killAllWorkerAck := make(chan bool)

	go shutdownHandler(chKill, killLinkProcessor, killLinkProcessorAck, killAllWorker, killAllWorkerAck, chWaitForExit)

	if resume && db.CountRemainingRows() == 0 {
		log.Println("Project is complete. No remaining URLs to crawl.")
		os.Exit(0)
	}

	crawler := NewCrawler(configuration)
	if err := crawler.initializeURLs(resume); err != nil {
		log.Fatal(err)
	}

	processDoc := configuration.ProjectType == config.PROJECTTYPECRAWL
	go linkProcessor(crawler.docsChan, configuration, killLinkProcessor, killLinkProcessorAck, processDoc)

	crawler.startWorkers(crawler.killWorker)
	go workerManager(crawler.killWorker, crawler.numWorkers, killAllWorker, killAllWorkerAck)

	// Start completion monitor for batch processing
	if configuration.ProjectType == config.PROJECTTYPEBATCH {
		go completionMonitor(configuration, killLinkProcessor, killLinkProcessorAck, killAllWorker, killAllWorkerAck, chWaitForExit)
	}
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
			log.Printf("Stopping %d workers", workerCount)
			
			// Send kill signals with timeout to prevent hanging
			for i := 0; i < workerCount; i++ {
				select {
				case killWorker <- 1:
					// Successfully sent kill signal
				case <-time.After(1 * time.Second):
					// Worker might have already terminated, continue
					log.Printf("Worker %d may have already terminated", i+1)
				}
			}
			
			// Give workers a moment to clean up
			time.Sleep(100 * time.Millisecond)
			log.Println("Worker termination signals sent")
			
			killAllWorkerAck <- true
			return
		}
	}
}

func randomDelay(min, max int) int {
	return rand.Intn(max-min) + min
}
func processPageWorker(configuration config.Configuration, chLinkProcessor chan *goquery.Document, killWorker chan int) {
	emptyQueueCount := 0
	maxEmptyChecks := 3 // Number of consecutive empty queue checks before considering work done

	for {
		select {
		case <-killWorker:
			if configuration.Debug {
				log.Println("Worker terminated")
			}
			return
		default:
			time.Sleep(time.Duration(randomDelay(1000, 5000)) * time.Millisecond)

			node, err := db.Pop()
			if err != nil {
				log.Printf("Error retrieving URL from queue: %v", err)
				continue
			}

			if node == (db.Node{}) {
				emptyQueueCount++
				if configuration.Debug {
					log.Printf("No URLs available in queue (check %d/%d)", emptyQueueCount, maxEmptyChecks)
				}

				// For batch processing, check if work is truly complete
				if configuration.ProjectType == config.PROJECTTYPEBATCH && emptyQueueCount >= maxEmptyChecks {
					remainingCount := db.CountRemainingRows()
					if remainingCount == 0 {
						log.Println("Batch processing complete, worker shutting down")
						return
					}
				}

				time.Sleep(time.Duration(randomDelay(1000, 5000)) * time.Millisecond)
				continue
			}

			// Reset empty queue counter when work is found
			emptyQueueCount = 0

			if err := pageProcessor(node.ID, node.Link, configuration, chLinkProcessor); err != nil {
				log.Printf("Error processing page %s: %v", node.Link, err)
			}
		}
	}
}

func pageProcessor(id int64, url string, configuration config.Configuration, chLinkProcessor chan *goquery.Document) error {
	log.Printf("Processing page: %s", url)

	doc, err := goquery.NewDocument(configuration.ProxyAPI + url)
	if err != nil {
		return fmt.Errorf("failed to fetch page %s: %w", url, err)
	}

	pageValidator := configuration.PageValidator
	if pageValidator == "" {
		pageValidator = "body"
	}

	article := doc.Find(pageValidator)
	if article.Length() == 0 {
		if configuration.Debug {
			log.Printf("Page validation failed for %s: no elements found matching '%s'", url, pageValidator)
		}
		return db.Update(id, 0, db.ValidationFailed, "")
	}

	switch configuration.ContentHolder {
	case config.CONTENTHOLDERTEXT, config.CONTENTHOLDERHTML:
		if err := extractTextContent(id, doc, configuration, url); err != nil {
			return fmt.Errorf("failed to extract text content: %w", err)
		}
	case config.CONTENTHOLDERATTR:
		if err := extractAttributeContent(id, doc, configuration); err != nil {
			return fmt.Errorf("failed to extract attribute content: %w", err)
		}
	default:
		return fmt.Errorf("invalid content holder type: %s", configuration.ContentHolder)
	}

	chLinkProcessor <- doc
	return nil
}

func extractTextContent(id int64, doc *goquery.Document, configuration config.Configuration, url string) error {
	if configuration.Debug {
		log.Println("Extracting text content")
	}

	selections := doc.Find(configuration.ContentSelector)
	if selections.Length() == 0 {
		if configuration.Debug {
			log.Printf("No content found matching selector '%s' on URL: %s", configuration.ContentSelector, url)
		}
		return nil
	}

	var count int64
	allContent := ""

	selections.Each(func(i int, s *goquery.Selection) {
		count++
		var content string

		if configuration.ContentHolder == config.CONTENTHOLDERTEXT {
			content = strings.TrimSpace(s.Text())
		} else if configuration.ContentHolder == config.CONTENTHOLDERHTML {
			content, _ = s.Html()
		}

		allContent += content
		filename := fmt.Sprintf("%s/%d_%d.txt", configuration.DataDir(), id, count)
		if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
			log.Printf("Failed to write content file %s: %v", filename, err)
		}
	})

	md5hash := md5.Sum([]byte(allContent))
	return db.Update(id, count, db.Success, fmt.Sprintf("%x", md5hash))
}

func extractAttributeContent(id int64, doc *goquery.Document, configuration config.Configuration) error {
	if configuration.Debug {
		log.Println("Extracting attribute content")
	}

	var count int64
	var attributes []string

	selections := doc.Find(configuration.ContentSelector)
	selections.Each(func(i int, s *goquery.Selection) {
		count++
		if attr, exists := s.Attr(configuration.ContentTagAttr); exists {
			attributes = append(attributes, attr)
		}
	})

	allContent := strings.Join(attributes, "\n")
	filename := fmt.Sprintf("%s/%d.txt", configuration.DataDir(), id)
	if err := ioutil.WriteFile(filename, []byte(allContent), 0644); err != nil {
		return fmt.Errorf("failed to write attribute file %s: %w", filename, err)
	}

	md5hash := md5.Sum([]byte(allContent))
	return db.Update(id, count, db.Success, fmt.Sprintf("%x", md5hash))
}

func completionMonitor(configuration config.Configuration, killLinkProcessor, killLinkProcessorAck, killAllWorker, killAllWorkerAck, chWaitForExit chan bool) {
	checkInterval := 5 * time.Second  // Reduced from 30 seconds to 5 seconds
	consecutiveEmptyChecks := 0
	maxConsecutiveChecks := 2  // Reduced from 3 to 2 for faster detection

	for {
		time.Sleep(checkInterval)
		
		remainingCount := db.CountRemainingRows()
		if configuration.Debug {
			log.Printf("Completion monitor: %d URLs remaining in queue", remainingCount)
		}
		
		if remainingCount == 0 {
			consecutiveEmptyChecks++
			log.Printf("Queue empty (check %d/%d)", consecutiveEmptyChecks, maxConsecutiveChecks)
			
			if consecutiveEmptyChecks >= maxConsecutiveChecks {
				log.Println("Batch processing completed. Initiating shutdown...")
				
				// Stop all workers first
				log.Println("Stopping workers...")
				killAllWorker <- true
				
				// Wait for workers to stop
				log.Println("Waiting for workers to terminate...")
				<-killAllWorkerAck
				log.Println("All workers stopped")
				
				// Stop link processor
				log.Println("Stopping link processor...")
				terminateLinkProcessor(killLinkProcessor, killLinkProcessorAck)
				log.Println("Link processor stopped")
				
				// Signal main to exit
				log.Println("Signaling main application to exit...")
				chWaitForExit <- true
				return
			}
		} else {
			consecutiveEmptyChecks = 0
		}
	}
}
