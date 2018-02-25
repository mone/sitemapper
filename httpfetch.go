package main

import (
	"net/http"
	"net/url"
	"io/ioutil"
	"sync"
	log "github.com/sirupsen/logrus"
)

// Simple struct used to deliver results downstream
type HtmlPage struct {
	Address url.URL
	Bytes []byte
}

// Abstracting access to network in order to mock it during tests
type HttpClient interface {
	Get (address url.URL) (resp *http.Response, err error)
}

// Default implementation of HttpClient uses the DefaultClient of the http package
type DefaultHttpClient struct {}

func (client *DefaultHttpClient) Get (address url.URL) (resp *http.Response, err error) {
	return http.Get(address.String())
}

// This function uses the given client to fetch the page at the given address and
// sends the output in the form of a HtmlPage downstream
func httpFetch(
	client HttpClient,
	address url.URL,
	output chan HtmlPage,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	log.Debug("Hitting network for ", address)

	// isolated in order to unify calls to the chan and to eventually
	// implement retries
	fetch := func() ([]byte, error) {
		resp, err := client.Get(address)
		if err != nil {
			return make([]byte, 0), err
		}
		// close the response once we've read and published it
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return make([]byte, 0), err
		}

		return body, nil
	}

	html, err := fetch()
	if err != nil {
		// currently in case of error we just skip the page
		// TODO wait & retry
		log.Error("Could not read ", address.String(), err)
	}

	log.Debug("Page retrieved ", address)
	output <- HtmlPage{ address, html }
}

/**
 * Starts listening on the provided urlChannel and spins up a new go routine every time
 * an URL is published on it. The new go routine will fetch the page from the web and will publish
 * a HtmlPage to the chan that is returned by this function
 */
func StartHttpFetchers(
	urlChannel chan url.URL,
	httpClient HttpClient,
) chan HtmlPage {

	respChan := make(chan HtmlPage)

	var wg sync.WaitGroup

	go func() {
		for toFetch := range urlChannel {
			wg.Add(1)
			// TODO evaluate the opportunity to introduce a pool of go-routines
			go httpFetch(httpClient, toFetch, respChan, &wg)
		}

		log.Info("Upstream channel down, preparing shutdown")

		// when the upstream channel is closed, we wait for any pending go routines and
		// then we close the downstream channel (note: considering the way this class is used
		// in this exercise, when upstream closes we know there are no pending go routines,
		// it still makes sense to implement the correct shutdown procedure anyway)
		wg.Wait()

		log.Info("All routines completed, closing downstream")
		close(respChan)
	}()

	return respChan

}