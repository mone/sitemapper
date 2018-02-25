package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSitemapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sitemapper Suite")
}
