package crawler

import (
	"fmt"
	"net/url"
	"path"
	"regexp"

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

// Wait for kill signal and
// Process document to find all eligible links and add it to database
func linkProcessor(docs chan *goquery.Document, configuration config.Configuration, kill, killLinkProcessorAck chan bool, processDoc bool) {
	if configuration.Debug == true {
		fmt.Println("linkProcessor started")
	}
	// TODO: How do I persist all the details before killing
	for {
		select {
		case <-kill:
			killLinkProcessorAck <- true
			return
		case doc := <-docs:
			if processDoc == true {
				lsURL(doc, configuration)
			}
		}
	}
}

// Process document to find all eligible links and add it to database
func lsURL(doc *goquery.Document, configuration config.Configuration) {
	if configuration.Debug {
		fmt.Println("ls_url")
	}

	var finalLinks []string
	// Find all the links and queue it
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, ok := s.Attr("href")
		if ok {
			finalLink, ok, err := validateURL(link, configuration, doc.Url)
			if err != nil {
				fmt.Println(err.Error())
			}

			if ok {
				//finalLink = configuration.ProxyAPI + finalLink
				if configuration.DisplayMatchedURL {
					fmt.Println("Valid Link found: " + finalLink)
				}
				finalLinks = append(finalLinks, finalLink)
			} else {
				if configuration.Debug == true {
					fmt.Println("Invalid URL: " + finalLink)
				}
			}
		}
	})

	if len(finalLinks) > 0 {
		_, err := db.PushMulti(finalLinks, 0)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func validateURL(link string, configuration config.Configuration, currentURL *url.URL) (string, bool, error) {
	if configuration.Debug {
		fmt.Println("validate_url: " + link)
	}

	var err error

	// If link is empty do nothing
	if link == "" {
		return "", false, nil
	}

	// Find out domain of the url
	rootUrl, err := url.Parse(configuration.RootURL)
	//fmt.Println(rootUrl)
	//fmt.Println(rootUrl.Scheme)
	//fmt.Println(rootUrl.Host)
	//fmt.Println(err)
	//if err != nil {return "", false, err}

	//configuration.RootURL

	domainURL := ""
	if configuration.ProxyAPI == "" {
		domainURL += currentURL.Scheme + "://"
		domainURL += currentURL.Host
	} else {
		domainURL += rootUrl.Scheme + "://"
		domainURL += rootUrl.Host
	}

	currentPath := currentURL.Path

	// Check if link is relative or absolute
	isAbsoluteURL := false
	if len(link) >= 4 && link[0:4] == "http" {
		isAbsoluteURL = true
	}

	// Build final link to be added to database
	finalLink := link
	if !isAbsoluteURL {
		// If link is relative to root
		if link[0:1] == "/" {
			finalLink = domainURL + link
		} else {
			if configuration.ProxyAPI == "" {
				finalPath := path.Join(path.Dir(currentPath), link)
				finalLink = domainURL + finalPath
			} else {
				if configuration.Debug == true {
					fmt.Println("The case of URL not relative to ROOT with ProxyAPI is not handled")
				}
			}
		}
	}

	// Check if URL matches configuration regex
	matched := true
	if configuration.LinkValidator != "" {
		matched, err = regexp.MatchString(configuration.LinkValidator, finalLink)
		if err != nil {
			return "", false, err
		}
	}

	// Check if links needs to be sanitized
	if matched && configuration.LinkSanitizer != "" {
		re := regexp.MustCompile(configuration.LinkSanitizer)
		finalLink = re.ReplaceAllString(finalLink, configuration.LinkSanitizerReplacement)
	}

	// fmt.Println("validate_url > link : " + link)
	// fmt.Println("validate_url > finalLink after sanitizarion : " + finalLink)
	// if matched {
	// 	// fmt.Println("validate_url > matched : yes")

	// } else {
	// 	// fmt.Println("validate_url > matched : no")
	// }

	return finalLink, matched, nil
	// Push URL duplicated will be ignored
	//if matched {
	//	_, err := db.Push(finalLink, 0)
	//	if err != nil {
	//		fmt.Println("Push Error: " + finalLink + " : " + err.Error())
	//	}
	//}
}
