package main

import (
	"net/url"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {

	root, err := url.Parse(os.Args[1])
	if err != nil {
		log.Fatal("Can't parse root")
		panic(1)
	}

	// we'll push the addresses of the pages we want to map on this channel
	addressChan := make(chan url.URL)

	client := &DefaultHttpClient{}

	// the http fetchers will read the addresses, fetch the pages and push them down the pagesChan
	pagesChan := StartHttpFetchers(addressChan, client)
	// the link extractor will read the pages, parse and extract the contained links and push them down the linksChan
	linksChan := StartLinkExtractor(pagesChan)
	// the MapSite will act both as the first and the last link in the chain of channels
	// will push the root down the addressChan, wait other links on the links chan and
	// will send those on the addressChan, wash rinse repeat
	siteMap := MapSite(*root, addressChan, linksChan)

	siteMap.Print(*root)

}

