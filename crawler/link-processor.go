package crawler

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/therahulprasad/spiders/crawler/config"
	"github.com/therahulprasad/spiders/crawler/db"
)

// TODO: Find a work around
// Fake function to process kill signal for quit handler
// func fakeLinkProcessor(kill, killLinkProcessorAck chan bool) {
// 	<-kill
// 	killLinkProcessorAck <- true
// }

func linkProcessor(docs chan *goquery.Document, configuration config.Configuration, kill, killLinkProcessorAck chan bool, processDoc bool) {
	if configuration.Debug {
		log.Println("Link processor started")
	}

	for {
		select {
		case <-kill:
			if configuration.Debug {
				log.Println("Link processor terminated")
			}
			killLinkProcessorAck <- true
			return
		case doc := <-docs:
			if processDoc {
				if err := extractAndQueueLinks(doc, configuration); err != nil {
					log.Printf("Error processing links from document: %v", err)
				}
			}
		}
	}
}

func extractAndQueueLinks(doc *goquery.Document, configuration config.Configuration) error {
	if configuration.Debug {
		log.Println("Extracting links from document")
	}

	var validLinks []string

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if !exists {
			return
		}

		validatedLink, isValid, err := validateURL(link, configuration, doc.Url)
		if err != nil {
			log.Printf("Error validating URL %s: %v", link, err)
			return
		}

		if isValid {
			if configuration.DisplayMatchedURL {
				log.Printf("Valid link found: %s", validatedLink)
			}
			validLinks = append(validLinks, validatedLink)
		} else if configuration.Debug {
			log.Printf("Invalid URL rejected: %s", validatedLink)
		}
	})

	if len(validLinks) > 0 {
		_, err := db.PushMulti(validLinks, 0)
		if err != nil {
			return fmt.Errorf("failed to queue %d links: %w", len(validLinks), err)
		}
		log.Printf("Queued %d new links", len(validLinks))
	}

	return nil
}

func validateURL(link string, configuration config.Configuration, currentURL *url.URL) (string, bool, error) {
	if configuration.Debug {
		log.Printf("Validating URL: %s", link)
	}

	if link == "" {
		return "", false, nil
	}

	rootURL, err := url.Parse(configuration.RootURL)
	if err != nil {
		return "", false, fmt.Errorf("invalid root URL %s: %w", configuration.RootURL, err)
	}

	finalLink := buildAbsoluteURL(link, currentURL, rootURL, configuration.ProxyAPI)

	if !isURLAllowed(finalLink, configuration) {
		return finalLink, false, nil
	}

	finalLink = sanitizeURL(finalLink, configuration)

	return finalLink, true, nil
}

func buildAbsoluteURL(link string, currentURL, rootURL *url.URL, proxyAPI string) string {
	if strings.HasPrefix(link, "http") {
		return link
	}

	var baseURL *url.URL
	if proxyAPI == "" {
		baseURL = currentURL
	} else {
		baseURL = rootURL
	}

	if strings.HasPrefix(link, "/") {
		return fmt.Sprintf("%s://%s%s", baseURL.Scheme, baseURL.Host, link)
	}

	if proxyAPI == "" {
		finalPath := path.Join(path.Dir(currentURL.Path), link)
		return fmt.Sprintf("%s://%s%s", baseURL.Scheme, baseURL.Host, finalPath)
	}

	return link
}

func isURLAllowed(finalLink string, configuration config.Configuration) bool {
	if configuration.LinkValidator != "" {
		matched, err := regexp.MatchString(configuration.LinkValidator, finalLink)
		if err != nil {
			log.Printf("Error matching link validator regex: %v", err)
			return false
		}
		if !matched {
			return false
		}
	}

	for _, invalidator := range configuration.LinkInvalidators {
		if invalidator == "" {
			continue
		}
		matched, err := regexp.MatchString(invalidator, finalLink)
		if err != nil {
			log.Printf("Error matching link invalidator regex: %v", err)
			continue
		}
		if matched {
			return false
		}
	}

	return true
}

func sanitizeURL(finalLink string, configuration config.Configuration) string {
	if configuration.LinkSanitizer != "" {
		re := regexp.MustCompile(configuration.LinkSanitizer)
		finalLink = re.ReplaceAllString(finalLink, configuration.LinkSanitizerReplacement)
	}
	return finalLink
}
