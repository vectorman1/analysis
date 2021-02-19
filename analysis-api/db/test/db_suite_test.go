package test_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestCurrencyRepository(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CurrencyRepository")
}
