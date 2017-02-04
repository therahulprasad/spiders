package crawler

import (
	"regexp"
	"github.com/therahulprasad/spiderman/crawler/db"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/therahulprasad/spiderman/crawler/config"
	"net/url"
)

func link_processor(docs chan *goquery.Document, configuration config.Configuration, kill, killLinkProcessorAck chan bool) {
	// TODO: How do I persist all the details before killing
	for {
		select {
		case doc := <-docs:
			ls_url(doc, configuration)
		case <- kill:
			killLinkProcessorAck <- true
			break
		}
	}
}

func ls_url(doc *goquery.Document, configuration config.Configuration) {
	if configuration.Debug {
		fmt.Println("ls_url")
	}

	// Find all the links and queue it
	doc.Find("a").Each(func (i int, s *goquery.Selection) {
		link, ok := s.Attr("href")
		if ok {
			err := queue_url(link, configuration)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	})
}

func queue_url(link string, configuration config.Configuration) error {
	if configuration.Debug {
		fmt.Println("queue_url")
	}

	// If link is empty do nothing
	if link == "" { return nil }

	// Find out domain of the url
	rootUrl, err := url.Parse(configuration.RootURL)
	if err != nil { return err }

	domainUrl := ""
	domainUrl += rootUrl.Scheme + "://"
	domainUrl += rootUrl.Host

	// Check if link is relative or absolute
	isAbsoluteUrl := false
	if len(link) >= 4 && link[0:4] == "http" {
		isAbsoluteUrl = true
	}

	// Build final link to be added to database
	finalLink := link
	if !isAbsoluteUrl {
		if link[0:1] == "/" {
			finalLink = domainUrl + link
		} else {
			finalLink = domainUrl + "/" + link
		}
	}

	// Check if URL matches configuration regex
	matched := true
	if configuration.LinkValidator != "" {
		matched, err =  regexp.MatchString(configuration.LinkValidator, finalLink)
		if err != nil {return err}
	}

	// Push URL duplicated will be ignored
	if matched {
		_, err := db.Push(finalLink, 0)
		if err != nil {
			fmt.Println("Push Error: " + finalLink + " : " + err.Error())
		}
	}

	return nil
}
