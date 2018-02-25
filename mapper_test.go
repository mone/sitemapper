package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/mone/sitemapper"
	"net/url"
)

var _ = Describe("Mapper", func() {

	var (
		pageUrl *url.URL
		monzoUrl *url.URL
		aboutPageUrl *url.URL
		otherPageUrl *url.URL
		lastPageUrl *url.URL

		addressChan chan url.URL
		linksChan chan HtmlPageLinks
	)

	BeforeEach(func() {
		pageUrl, _ = url.Parse("https://www.google.com/")
		monzoUrl, _ = url.Parse("https://www.monzo.com/")
		aboutPageUrl, _ = url.Parse("https://www.google.com/about")
		otherPageUrl, _ = url.Parse("https://www.google.com/other")
		lastPageUrl, _ = url.Parse("https://www.google.com/42")

		addressChan = make(chan url.URL)
		linksChan = make(chan HtmlPageLinks)
	})

	It("should retrieve the full tree", func(done Done) {

		go func() {
			res := MapSite(*pageUrl, addressChan, linksChan)

			expectedMap := map[url.URL][]url.URL {
				*pageUrl: { *aboutPageUrl, *otherPageUrl },
				*aboutPageUrl: {},
				*otherPageUrl: { *lastPageUrl },
				*lastPageUrl: {},
			}

			Expect(map[url.URL][]url.URL(res)).To(Equal(expectedMap))

			close(done)
		}()

		Eventually(addressChan).Should(Receive(Equal(*pageUrl)))

		linksChan <- HtmlPageLinks{
			*pageUrl,
			[]url.URL{*aboutPageUrl, *otherPageUrl},
		}

		Eventually(addressChan).Should(Receive(Equal(*aboutPageUrl)))
		Eventually(addressChan).Should(Receive(Equal(*otherPageUrl)))

		linksChan <- HtmlPageLinks{
			*aboutPageUrl,
			[]url.URL{},
		}

		linksChan <- HtmlPageLinks{
			*otherPageUrl,
			[]url.URL{*lastPageUrl},
		}

		Eventually(addressChan).Should(Receive(Equal(*lastPageUrl)))

		linksChan <- HtmlPageLinks{
			*lastPageUrl,
			[]url.URL{},
		}

		Eventually(addressChan).Should(BeClosed())

		close(linksChan)

	})

	It("should ignore addresses outside the root's host", func(done Done) {
		go func() {
			res := MapSite(*pageUrl, addressChan, linksChan)

			expectedMap := map[url.URL][]url.URL {
				*pageUrl: { *monzoUrl },
			}

			Expect(map[url.URL][]url.URL(res)).To(Equal(expectedMap))

			close(done)
		}()

		Eventually(addressChan).Should(Receive(Equal(*pageUrl)))

		linksChan <- HtmlPageLinks{
			*pageUrl,
			[]url.URL{*monzoUrl},
		}

		Consistently(addressChan).ShouldNot(Receive())

		Eventually(addressChan).Should(BeClosed())

		close(linksChan)
	})

	It("should fetch each address once and avoid infinite loops", func(done Done) {
		go func() {
			res := MapSite(*pageUrl, addressChan, linksChan)

			expectedMap := map[url.URL][]url.URL {
				*pageUrl: { *aboutPageUrl },
				*aboutPageUrl: { *pageUrl },
			}

			Expect(map[url.URL][]url.URL(res)).To(Equal(expectedMap))

			close(done)
		}()

		Eventually(addressChan).Should(Receive(Equal(*pageUrl)))

		linksChan <- HtmlPageLinks{
			*pageUrl,
			[]url.URL{*aboutPageUrl},
		}

		Eventually(addressChan).Should(Receive(Equal(*aboutPageUrl)))

		linksChan <- HtmlPageLinks{
			*aboutPageUrl,
			[]url.URL{*pageUrl},
		}

		Consistently(addressChan).ShouldNot(Receive())

		Eventually(addressChan).Should(BeClosed())

		close(linksChan)
	})

})