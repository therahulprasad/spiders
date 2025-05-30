package crawler

import (
	"bufio"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/therahulprasad/spiders/crawler/config"
	"github.com/therahulprasad/spiders/crawler/db"
)

func push_all_batch_links(configuration config.Configuration) {

	// To store list of all the processed links
	var finalLinks []string

	if configuration.BatchURL != "" {
		// https://www.example.com/[$01-$999]/[$1-$9]/sjkn
		pattern := `http.*?\[\$(\d+)-\$(\d+)\](.+?\[\$(\d+)-\$(\d+)\].*?)?`
		// $1 - $2   &   $4 - $5

		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(configuration.BatchURL)

		if matches[1] != "" && matches[2] != "" {
			// Process URL

			// Find start and end number
			start, _ := strconv.Atoi(matches[1])
			end, _ := strconv.Atoi(matches[2])

			// Find length of starting number for padding
			start_len := len(matches[1])
			padding_limit := math.Pow10(start_len - 1)

			// Exit on invalid config
			if end < start {
				log.Fatal("Invalid batch_url format. Start must be less than end")
			}

			// Create list of links
			for i := start; i <= end; i++ {
				// Push every link to db
				from := "[$" + matches[1] + "-$" + matches[2] + "]"

				// Add padding of 0s
				padding := ""
				if i < int(padding_limit) {
					digit_length := math.Floor(math.Log10(float64(i)))
					for j := int(digit_length); j < start_len-1; j++ {
						padding += "0"
					}
				}
				to := padding + strconv.Itoa(i)
				finalLink := strings.Replace(configuration.BatchURL, from, to, 1)
				finalLinks = append(finalLinks, finalLink)
			}
		} else {
			log.Fatal("Invalid batch_url format")
		}
	} else if configuration.BatchFile != "" {
		file, err := os.Open(configuration.BatchFile)
		if err != nil {
			log.Fatalf("Failed to open batch file: %v", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		urlPattern := regexp.MustCompile(`^https?://[^\s]+$`)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if urlPattern.MatchString(line) {
				finalLinks = append(finalLinks, line)
			} else if line != "" {
				log.Printf("Ignoring non-URL line in batch file: %s", line)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("Error reading batch file: %v", err)
		}
	} else {
		log.Fatal("batch_url and batch_file both cant be empty for project_type: batch")
	}

	if _, err := db.PushMulti(finalLinks, -1); err != nil {
		log.Fatalf("Failed to push batch links: %v", err)
	}
}
