package conformance

import (
	"encoding/json"
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test04ContentManagement = func() {
	g.Context("Content Management - Requires push and delete actions", func() {

		const defaultTagName = "tagTest0"
		var tagToDelete string
		var numTags int

		g.Context("Setup", func() {
			g.Specify("Push - push a manifest with associated tags", func() {
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, _ := client.Do(req)

				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetHeader("Content-Length", configContentLength).
					SetHeader("Content-Type", "application/octet-stream").
					SetQueryParam("digest", blobDigest).
					SetBody(configContent)
				resp, _ = client.Do(req)

				tagToDelete = defaultTagName
				req = client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(tagToDelete)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(manifestContent)
				resp, _ = client.Do(req)
			})

			g.Specify("Discovery - check how many tags there are before anything gets deleted", func() {
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				tagList := &TagList{}
				jsonData := []byte(resp.String())
				err = json.Unmarshal(jsonData, tagList)
				numTags = len(tagList.Tags)
			})
		})

		g.Context("Manifest delete", func() {
			g.Specify("DELETE request to manifest tag should return 202, unless tag deletion is disallowed (400)", func() {
				SkipIfDisabled(contentManagement)
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

			g.Specify("DELETE request to manifest (digest) should yield 202 response unless already deleted", func() {
				SkipIfDisabled(contentManagement)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifestDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				// In the case that the previous request was accepted, this may or may not fail (which is ok)
				Expect(resp.StatusCode()).To(SatisfyAny(
					Equal(http.StatusAccepted),
					Equal(http.StatusNotFound),
				))
			})

			g.Specify("GET request to deleted manifest URL should yield 404 response, unless delete is disallowed", func() {
				SkipIfDisabled(contentManagement)
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifestDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					Equal(http.StatusNotFound),
					Equal(http.StatusOK),
				))
			})

			g.Specify("GET request to tags list should reflect manifest deletion", func() {
				SkipIfDisabled(contentManagement)
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

		g.Context("Blob delete", func() {
			g.Specify("DELETE request to blob URL should yield 202 response", func() {
				SkipIfDisabled(contentManagement)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(blobDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusAccepted))
			})

			g.Specify("GET request to deleted blob URL should yield 404 response", func() {
				SkipIfDisabled(contentManagement)
				req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>", reggie.WithDigest(blobDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})
		})
	})
}
