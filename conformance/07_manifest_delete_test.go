package conformance

import (
	"encoding/json"
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test07ManifestDelete = func() {
	g.Context("Manifest Delete", func() {
		g.Specify("DELETE request to manifest URL should yield 202 response", func() {
			req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifestDigest))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusAccepted))
		})

		g.Specify("GET request to deleted manifest URL should yield 404 response", func() {
			req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifestDigest))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
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
