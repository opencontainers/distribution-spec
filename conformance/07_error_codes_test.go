package conformance

import (
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test07ErrorCodes = func() {
	g.Context("Error codes", func() {
		g.Specify("400 response body should contain OCI-conforming JSON message", func() {
			req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
				reggie.WithReference("sha256:totallywrong")).
				SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
				SetBody(invalidManifestContent)
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusBadRequest))

			errorResponses, err := resp.Errors()
			Expect(err).To(BeNil())

			Expect(errorResponses).ToNot(BeEmpty())
			Expect(errorCodes).To(ContainElement(errorResponses[0].Code))
		})

		g.Specify("Response from DELETE request to manifest URL (tag) should have OCI-conforming code", func() {
			req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
				reggie.WithReference(firstTag))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())

			errorResponses, err := resp.Errors()
			Expect(err).To(BeNil())

			Expect(errorResponses).ToNot(BeEmpty())
			Expect(errorCodes).To(ContainElement(errorResponses[0].Code))
		})
	})
}
