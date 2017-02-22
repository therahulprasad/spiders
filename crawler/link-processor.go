package crawler

import (
	"regexp"
	"github.com/therahulprasad/spiderman/crawler/db"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/therahulprasad/spiderman/crawler/config"
	"net/url"
	"github.com/go-playground/log"
	"errors"
	"path"
)

func link_processor(docs chan *goquery.Document, configuration config.Configuration, kill, killLinkProcessorAck chan bool) {
	// TODO: How do I persist all the details before killing
	for {
		select {
		case <- kill:
			killLinkProcessorAck <- true
			return
		case doc := <-docs:
			ls_url(doc, configuration)
		}
	}
}

func ls_url(doc *goquery.Document, configuration config.Configuration) {
	if configuration.Debug {
		fmt.Println("ls_url")
	}

	var finalLinks []string
	// Find all the links and queue it
	doc.Find("a").Each(func (i int, s *goquery.Selection) {
		link, ok := s.Attr("href")
		if ok {
			finalLink, ok, err := validate_url(link, configuration, doc.Url)
			if err != nil { fmt.Println(err.Error()) }

			if ok {
				if configuration.DisplayMatchedUrl {
					fmt.Println("Valid Link found: " + finalLink)
				}
				finalLinks = append(finalLinks, finalLink)
			}
		}
	})

	if len(finalLinks) > 0 {
		_, err := db.PushMulti(finalLinks, 0)
		if err != nil {log.Println(err.Error())}
	}
}

func validate_url(link string, configuration config.Configuration, currentUrl *url.URL) (string, bool, error) {
	if configuration.Debug {
		fmt.Println("validate_url")
	}

	var err error

	// If link is empty do nothing
	if link == "" { return "", false, errors.New("Empty link") }

	// Find out domain of the url
	//rootUrl, err := url.Parse(configuration.RootURL)
	//if err != nil {return "", false, err}

	domainUrl := ""
	domainUrl += currentUrl.Scheme + "://"
	domainUrl += currentUrl.Host

	currentPath := currentUrl.Path

	// Check if link is relative or absolute
	isAbsoluteUrl := false
	if len(link) >= 4 && link[0:4] == "http" {
		isAbsoluteUrl = true
	}

	// Build final link to be added to database
	finalLink := link
	if !isAbsoluteUrl {
		finalPath := path.Join(path.Dir(currentPath), link)
		finalLink = domainUrl + finalPath
	}

	// Check if URL matches configuration regex
	matched := true
	if configuration.LinkValidator != "" {
		matched, err =  regexp.MatchString(configuration.LinkValidator, finalLink)
		if err != nil {return "", false, err}
	}

	// Check if links needs to be sanitized
	if matched && configuration.LinkSanitizer != "" {
		re := regexp.MustCompile(configuration.LinkSanitizer)
		finalLink = re.ReplaceAllString(finalLink, configuration.LinkSanitizerReplacement)
	}

	return finalLink, matched, nil
	// Push URL duplicated will be ignored
	//if matched {
	//	_, err := db.Push(finalLink, 0)
	//	if err != nil {
	//		fmt.Println("Push Error: " + finalLink + " : " + err.Error())
	//	}
	//}
}
