# Site Mapper

Simple go site mapper: it will construct a site map starting from a specified address.

Addresses having a different host than the initial root are reported but not expanded.

## Build

`go build sitemapper.go httpfetch.go linkextractor.go mapper.go`

### Dependencies

```
go get github.com/sirupsen/logrus
go get github.com/PuerkitoBio/goquery
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega/...
```
## Run

once built
`./sitemapper http://www.example.com/`
