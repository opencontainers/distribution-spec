package conformance

import (
	"testing"

	g "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestConformance(t *testing.T) {
	g.Describe(suiteDescription, func() {
		//test01BaseAPIRoute()
		//test02BlobUploadStreamed()
		//test03BlobUploadMonolithic()
		//test04BlobUploadChunked()
		//test05ManifestUpload()
		//test06TagsList()
		//test07ErrorCodes()
		//test08ManifestDelete()
		//test09BlobDelete()
		pullTest()
		pushTest()
	})
	RegisterFailHandler(g.Fail)
	reporters := []g.Reporter{newHTMLReporter(reportHTMLFilename), reporters.NewJUnitReporter(reportJUnitFilename)}
	g.RunSpecsWithDefaultAndCustomReporters(t, suiteDescription, reporters)
}
