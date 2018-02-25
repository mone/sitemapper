package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/mone/sitemapper"
	"net/url"
)

var _ = Describe("StartLinkExtractor", func() {

	var (
		pageUrl *url.URL
		monzoUrl *url.URL
		aboutPageUrl *url.URL
	)

	BeforeEach(func() {
		pageUrl, _ = url.Parse("https://www.google.com/")
		monzoUrl, _ = url.Parse("https://www.monzo.com/")
		aboutPageUrl, _ = url.Parse("https://www.google.com/about")
	})


	It("should extract a link from a document", func(done Done) {
		documentWithOneLink := ([]byte)(`<a href="https://www.monzo.com/">link</a>`)

		pages := make(chan HtmlPage)

		output := StartLinkExtractor(pages)

		pages <- HtmlPage{*pageUrl, documentWithOneLink}

		res := <-output

		Expect(res).To(Equal(HtmlPageLinks{
			*pageUrl,
			[]url.URL{*monzoUrl},
		}))

		close(done)

	})

	It("should extract and resolve a relative link", func(done Done) {
		documentWithRelativeLink := ([]byte)(`<a href="/about">link</a>`)

		pages := make(chan HtmlPage)

		output := StartLinkExtractor(pages)

		pages <- HtmlPage{*pageUrl, documentWithRelativeLink}

		res := <-output

		Expect(res).To(Equal(HtmlPageLinks{
			*pageUrl,
			[]url.URL{*aboutPageUrl},
		}))

		close(done)
	})

	It("should extract all the links from a document", func(done Done) {
		documentWithMoreLinks := ([]byte)(`
			<a href="https://www.monzo.com/">link</a>
			<a href="https://www.google.com/">link</a>
			<a href="https://www.google.com/about">link</a>
		`)

		pages := make(chan HtmlPage)

		output := StartLinkExtractor(pages)

		pages <- HtmlPage{*pageUrl, documentWithMoreLinks}

		res := <-output

		Expect(res).To(Equal(HtmlPageLinks{
			*pageUrl,
			[]url.URL{*monzoUrl, *pageUrl, *aboutPageUrl},
		}))

		close(done)

	})

	It("should not extract duplicates", func(done Done) {
		documentWithMoreLinks := ([]byte)(`
			<a href="https://www.google.com/">link</a>
			<a href="https://www.google.com/">link</a>
		`)

		pages := make(chan HtmlPage)

		output := StartLinkExtractor(pages)

		pages <- HtmlPage{*pageUrl, documentWithMoreLinks}

		res := <-output

		Expect(res).To(Equal(HtmlPageLinks{
			*pageUrl,
			[]url.URL{*pageUrl},
		}))

		close(done)

	})


	It("should extract links nested in other elements", func(done Done) {
		documentWithNestedLink := ([]byte)(`
			<html><body><div><span>
				<a href="https://www.monzo.com/">link</a>
			</span></div></body></html>
			
		`)

		pages := make(chan HtmlPage)

		output := StartLinkExtractor(pages)

		pages <- HtmlPage{*pageUrl, documentWithNestedLink}

		res := <-output

		Expect(res).To(Equal(HtmlPageLinks{
			*pageUrl,
			[]url.URL{*monzoUrl},
		}))

		close(done)
	})

	It("should ignore commented links", func(done Done) {
		documentWithCommentedLink := ([]byte)(`
			<a href="https://www.monzo.com/">link</a>
			<!--
				<a href="https://www.google.com/">link</a>
			-->
		`)

		pages := make(chan HtmlPage)

		output := StartLinkExtractor(pages)

		pages <- HtmlPage{*pageUrl, documentWithCommentedLink}

		res := <-output

		Expect(res).To(Equal(HtmlPageLinks{
			*pageUrl,
			[]url.URL{*monzoUrl},
		}))

		close(done)

	})

	It("should not fail when no links are in the page", func(done Done) {
		documentWithNoLink := ([]byte)(`<div>empty</div>`)

		pages := make(chan HtmlPage)

		output := StartLinkExtractor(pages)

		pages <- HtmlPage{*pageUrl, documentWithNoLink}

		res := <-output

		Expect(res).To(Equal(HtmlPageLinks{
			*pageUrl,
			[]url.URL{},
		}))

		close(done)

	})

	It("should not fail when page is empty", func(done Done) {
		documentEmpty := make([]byte, 0)

		pages := make(chan HtmlPage)

		output := StartLinkExtractor(pages)

		pages <- HtmlPage{*pageUrl, documentEmpty}

		res := <-output

		Expect(res).To(Equal(HtmlPageLinks{
			*pageUrl,
			[]url.URL{},
		}))

		close(done)

	})

	It("should not fail when page is nil", func(done Done) {
		var documentNil []byte

		pages := make(chan HtmlPage)

		output := StartLinkExtractor(pages)

		pages <- HtmlPage{*pageUrl, documentNil}

		res := <-output

		Expect(res).To(Equal(HtmlPageLinks{
			*pageUrl,
			[]url.URL{},
		}))

		close(done)

	})

	It("should not fail when page can't be parsed", func(done Done) {
		documentNotParsable := ([]byte)(`<<>>`) // TODO find a page that won't parse

		pages := make(chan HtmlPage)

		output := StartLinkExtractor(pages)

		pages <- HtmlPage{*pageUrl, documentNotParsable}

		res := <-output

		Expect(res).To(Equal(HtmlPageLinks{
			*pageUrl,
			[]url.URL{},
		}))

		close(done)

	})

})


