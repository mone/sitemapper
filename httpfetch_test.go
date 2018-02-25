package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/mone/sitemapper"
	"net/url"
	"net/http/httptest"
	"bytes"
	"net/http"
	"errors"
)

type HttpClientMock struct {
	response map[url.URL]string
}

func (client *HttpClientMock) Get (address url.URL) (resp *http.Response, err error) {
	recorder := httptest.NewRecorder()
	recorder.Body = bytes.NewBufferString(client.response[address])
	return recorder.Result(), nil
}

type BrokenHttpClientMock struct {}

func (client *BrokenHttpClientMock) Get (address url.URL) (resp *http.Response, err error) {
	return nil, errors.New("error")
}

type ComposedHttpClientMock struct {
	clients map[url.URL]HttpClient
}

func (client *ComposedHttpClientMock) Get (address url.URL) (resp *http.Response, err error) {
	return client.clients[address].Get(address)
}


var _ = Describe("StartHttpFetchers", func() {

	var (
		url1 *url.URL
		url2 *url.URL
		page1 string
		page2 string
	)

	BeforeEach(func() {
		url1, _ = url.Parse("http://www.example.com/")
		url2, _ = url.Parse("https://www.monzo.com/")

		page1 = "<html>page1</html>"
		page2 = "<html>page2</html>"
	})

	It("should fetch a couple of addresses", func(done Done) {
		client := HttpClientMock {
			map[url.URL]string{
				*url1: page1,
				*url2: page2,
			},
		}

		// add a buffer so we can fill it at once and wait for the responses
		inChan := make(chan url.URL, 2)

		outChan := StartHttpFetchers(inChan, &client)

		inChan <- *url1
		inChan <- *url2

		res1 := <-outChan
		res2 := <-outChan
		Consistently(outChan).ShouldNot(Receive())

		// order of output is not guaranteed, so we put everything together
		res := []HtmlPage{res1, res2}

		Expect(res).To(ContainElement(HtmlPage{
			*url1, []byte(page1),
		}))
		Expect(res).To(ContainElement(HtmlPage{
			*url2, []byte(page2),
		}))

		close(inChan)
		Eventually(outChan).Should(BeClosed())

		// tell ginko we're done
		close(done)

	})

	It("keep going in case of failures", func(done Done) {
		okClient := HttpClientMock {
			map[url.URL]string{
				*url2: page2,
			},
		}

		client := ComposedHttpClientMock{
			map[url.URL]HttpClient{
				*url1: &BrokenHttpClientMock{},
				*url2: &okClient,
			},
		}

		// add a buffer so we can fill it at once and wait for the responses
		inChan := make(chan url.URL, 2)

		outChan := StartHttpFetchers(inChan, &client)

		inChan <- *url1
		inChan <- *url2

		// currently there is no retry mechanism, so the failed request will produce an empty page
		res1 := <-outChan
		res2 := <-outChan
		Consistently(outChan).ShouldNot(Receive())

		res := []HtmlPage{res1, res2}

		Expect(res).To(ContainElement(HtmlPage{
			*url1, make([]byte, 0),
		}))
		Expect(res).To(ContainElement(HtmlPage{
			*url2, []byte(page2),
		}))

		close(inChan)
		Eventually(outChan).Should(BeClosed())

		// tell ginko we're done
		close(done)

	})

})
