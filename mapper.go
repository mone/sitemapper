package main

import (
	"net/url"
	log "github.com/sirupsen/logrus"
	"fmt"
)

// Utility function, checks if two addresses pertain to the same host
func isSameHost(root *url.URL, other *url.URL) bool {
	return other.Host == root.Host
}

// The MapSite will start by pushing the specified root down the addressChan,
// will wait the related links on the link chan and (if those links are from
// the same host as the root and not already processed) will send those on the addressChan too.
// Once it has retrieved all the pages in the tree for the specified root, it will return a map
// containing as key the various pages and as value the list of the pages it links to.
func MapSite(root url.URL, addressChan chan url.URL, linksChan chan HtmlPageLinks) PagesMap {
	state := initState()
	log.Info("Starting crawling from root ", root)
	addressChan <- root
	state.onRequested(root)

	for links := range linksChan {
		// update the state (mapper is single threaded, no sync needed)
		state.onRetrieved(links.Address, links.LinksTo)

		for _, link := range links.LinksTo {
			if isSameHost(&root, &link) && state.shouldBeRequested(link) {
				log.Debug("Requesting ", link)
				addressChan <- link
				state.onRequested(link)
			} else {
				log.Debug("Skipping ", link)
			}
		}

		if !state.hasPending() {
			// all that we pushed down the addressChan has come back
			// through the linksChan, there is nothing else for us to do
			log.Info("Fetching completed")
			// we completed our crawling, let's close the channel we write to,
			// this will generate a chain reaction, leading to the closure of the
			// linksChan, that will allow us to exit
			close(addressChan)
		}

	}

	return state.retrieved

}

type PendingMap map[url.URL]bool
type PagesMap map[url.URL][]url.URL

// Stores the current state of the mapper
type State struct {
	pending PendingMap
	retrieved PagesMap
}

func initState() State {
	return State{
		make(map[url.URL]bool),
		make(map[url.URL][]url.URL),
	}
}

func (state *State) onRequested(url url.URL) {
	log.Print("Fetching ", url.String())
	state.pending[url] = true
}

func (state *State) onRetrieved(url url.URL, links []url.URL) {
	log.Print("Fetched ", len(links), " ", url.String())
	delete(state.pending, url)
	state.retrieved[url] = links
}

func (state *State) shouldBeRequested(url url.URL) bool {
	_, isPending := state.pending[url]
	_, isRetrieved := state.retrieved[url]
	return !isPending && !isRetrieved
}

func (state *State) hasPending() bool {
	return len(state.pending) != 0
}

// struct used to simulate the recursion stack
type StackElement struct {
	address url.URL
	level int
}

// Prints the sitemap
func (pages PagesMap) Print(root url.URL) {
	// recursive version might be more concise, but if I understood correctly
	// go does not interpret tail recursion so it would risk a stack overflow,
	// let's iterate (assuming max slice size > max stack size)

	printed := make(map[url.URL]bool, 0)
	stack := make([]StackElement, 0)

	// start from the root links and go through the map
	stack = append(stack, StackElement{root, 0})

	for len(stack) > 0 {
		// pop the head
		toPrint := stack[0]
		stack = stack[1:]

		// setup some decoration
		for i := 0; i < toPrint.level; i++ {
			fmt.Print("  ")
		}
		if toPrint.level > 0 {
			fmt.Print("|-")
		}

		_, alreadyPrinted := printed[toPrint.address]

		if !alreadyPrinted {
			fmt.Println(toPrint.address.String())
			printed[toPrint.address] = true

			nextLevel := toPrint.level + 1

			for _, child := range pages[toPrint.address] {
				// push the stack
				stack = append([]StackElement{{child, nextLevel}}, stack...)
			}

		} else {
			// we already printed the tree starting from this page, just add a line
			// to reference the previously printed branch
			fmt.Println(toPrint.address.String(), "--> see above")
		}

	}

}