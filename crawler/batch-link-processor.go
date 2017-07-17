package crawler

import (
	"github.com/therahulprasad/spiders/crawler/config"
	"regexp"
	"log"
	"strconv"
	"strings"
	"math"
	"github.com/therahulprasad/spiders/crawler/db"
)

func push_all_batch_links(configuration config.Configuration) {
	// To store list of all the processed links
	var finalLinks []string

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
		padding_limit := math.Pow10(start_len-1)

		// Exit on invalid config
		if end < start {
			log.Fatal("Invalid batch_url format. Start must be less than end")
		}

		// Create list of links
		for i:=start; i<=end; i++ {
			// Push every link to db
			from := "[$"+matches[1]+"-$"+matches[2]+"]"

			// Add padding of 0s
			padding := ""
			if i < int(padding_limit) {
				digit_length := math.Floor(math.Log10(float64(i)))
				for j:=int(digit_length); j<start_len-1; j++ {
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

	//os.Exit(0)
	db.PushMulti(finalLinks, -1)
}