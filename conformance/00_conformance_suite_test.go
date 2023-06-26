package conformance

import (
	"testing"

	g "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestConformance(t *testing.T) {
	g.Describe(suiteDescription, func() {
		test01Pull()
		test02Push()
		test03ContentDiscovery()
		test04ContentManagement()
	})

	RegisterFailHandler(g.Fail)
	reporters := []g.Reporter{newHTMLReporter(reportHTMLFilename), reporters.NewJUnitReporter(reportJUnitFilename)}
	g.RunSpecsWithDefaultAndCustomReporters(t, suiteDescription, reporters)
}
