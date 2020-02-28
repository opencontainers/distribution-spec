package conformance

import (
	"encoding/json"
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test08ManifestDelete = func() {
	g.Context("Manifest Delete", func() {
		g.Specify("DELETE request to manifest should return 202, 400, or 405", func() {
			req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
				reggie.WithReference(firstTag))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(SatisfyAny(
				Equal(http.StatusBadRequest),
				Equal(http.StatusMethodNotAllowed),
				Equal(http.StatusAccepted)))
			if resp.StatusCode() == http.StatusBadRequest {
				errorResponses, err := resp.Errors()
				Expect(err).To(BeNil())
				Expect(errorResponses).ToNot(BeEmpty())
				Expect(errorResponses[0].Code).To(Equal(errorCodes[UNSUPPORTED]))
			}
		})

		g.Specify("DELETE request to manifest (digest) should yield 202 response unless delete disallowed or already deleted", func() {
			req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifestDigest))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			// In the case that the previous request was accepted, this may or may not fail (which is ok)
			Expect(resp.StatusCode()).To(SatisfyAny(
				Equal(http.StatusBadRequest),
				Equal(http.StatusMethodNotAllowed),
				Equal(http.StatusAccepted),
				Equal(http.StatusNotFound),
			))
			if resp.StatusCode() == http.StatusBadRequest {
				errorResponses, err := resp.Errors()
				Expect(err).To(BeNil())
				Expect(errorResponses).ToNot(BeEmpty())
				Expect(errorResponses[0].Code).To(Equal(errorCodes[UNSUPPORTED]))
			}
		})

		g.Specify("GET request to deleted manifest URL should yield 404 response, unless delete is disallowed", func() {
			req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifestDigest))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(SatisfyAny(
				Equal(http.StatusNotFound),
				Equal(http.StatusOK),
			))
		})

		g.Specify("GET request to tags list should reflect manifest deletion", func() {
			req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list")
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			tagList := &TagList{}
			jsonData := []byte(resp.String())
			err = json.Unmarshal(jsonData, tagList)
			Expect(err).To(BeNil())
			Expect(len(tagList.Tags)).To(BeNumerically("<", numTags))
		})
	})
}
