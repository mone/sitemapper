package main

import (
	"net/url"
	"bytes"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// Simple struct used to deliver results downstream
type HtmlPageLinks struct {
	Address url.URL
	LinksTo []url.URL
}

// Given a html page it will parse it, extract the links and send them downstream
func extractLinks(page HtmlPage, output chan HtmlPageLinks) {
	log.Debug("Parsing document ", page.Address)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(page.Bytes))

	if err != nil {
		log.Error("Can't parse document", page.Address, err)
		output <- HtmlPageLinks{page.Address, make([]url.URL, 0)}
		return
	}

	log.Debug("Document parsed, extracting links ", page.Address)

	// grab all the href values from the document save them in a "Set" to avoid duplicates
	linksSet := make(map[url.URL]bool, 0)
	doc.Find("a").Each(func(_ int, elem *goquery.Selection) {
		value, ok := elem.Attr("href")
		if ok {
			asUrl, error := url.Parse(value)
			if error != nil {
				log.Warn("Can't parse address ", value, error)
			} else {
				// makes the address absolute (if necessary) and appends it to our set
				linksSet[*page.Address.ResolveReference(asUrl)] = true
			}
		}
	})

	// convert our list in an array
	links := make([]url.URL, 0, len(linksSet))
	for  address := range linksSet {
		links = append(links, address)
	}

	log.Debug("Links extracted ", page.Address, " ", links)

	output <- HtmlPageLinks{page.Address, links}

}

// Reads pages from the given chan and outputs contained links on the
// returned chan
func StartLinkExtractor(requests chan HtmlPage) chan HtmlPageLinks {

	respChan := make(chan HtmlPageLinks)

	go func() {
		for toParse := range requests {
			// I expect extractLinks to be much faster than the http fetcher,
			// so using a dedicated go routine should not be necessary here
			extractLinks(toParse, respChan)
		}

		log.Info("Upstream closed, closing downstream")
		// once the upstream channel is closed we can close the downstream one
		close(respChan)
	}()

	return respChan

}