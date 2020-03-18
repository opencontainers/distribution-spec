package conformance

import (
	"encoding/json"
	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"os"
)

const(
	testTagName = "tagTest0"
)

var contentManagementTest = func() {
	g.Context("Content Management", func() {
		g.Context("Setup", func() {
			g.Specify("Push - push a manifest with associated tags", func() {
				if userDisabled(push) {
					tagToDelete = os.Getenv(envVarTagToDelete)
				}
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusAccepted))
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetHeader("Content-Length", configContentLength).
					SetHeader("Content-Type", "application/octet-stream").
					SetQueryParam("digest", blobDigest).
					SetBody(configContent)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				location := resp.Header().Get("Location")
				Expect(location).ToNot(BeEmpty())
				Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
				tagToDelete = testTagName
				req = client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(tagToDelete)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(manifestContent)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				location = resp.Header().Get("Location")
				Expect(location).ToNot(BeEmpty())
				Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
			})
		})

		g.Context("Test deletion endpoints", func() {
			g.Specify("DELETE request to manifest tag should return 202 or 400", func() {
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(tagToDelete))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					Equal(http.StatusBadRequest),
					Equal(http.StatusAccepted)))
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
				SkipIfDisabled(discovery)
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
	})
}
