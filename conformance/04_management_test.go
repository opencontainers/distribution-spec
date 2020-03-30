package conformance

import (
	"encoding/json"
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test04Management = func() {
	g.Context("Management", func() {

		const defaultTagName = "tagTest0"
		var tagToDelete string
		var numTags int

		g.Context("Setup", func() {
			g.Specify("Populate registry with test blob", func() {
				RunOnlyIf(runManagementSetup)
				SkipIfDisabled(management)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, _ := client.Do(req)

				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetHeader("Content-Length", configContentLength).
					SetHeader("Content-Type", "application/octet-stream").
					SetQueryParam("digest", blobDigest).
					SetBody(configContent)
				_, err := client.Do(req)
				_ = err
			})

			g.Specify("Populate registry with test tag", func() {
				RunOnlyIf(runManagementSetup)
				SkipIfDisabled(management)
				tagToDelete = defaultTagName
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(tagToDelete)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(manifestContent)
				_, err := client.Do(req)
				_ = err
			})

			g.Specify("Check how many tags there are before anything gets deleted", func() {
				RunOnlyIf(runManagementSetup)
				SkipIfDisabled(management)
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				tagList := &TagList{}
				jsonData := []byte(resp.String())
				err = json.Unmarshal(jsonData, tagList)
				Expect(err).To(BeNil())
				numTags = len(tagList.Tags)
			})
		})

		g.Context("Manifest delete", func() {
			g.Specify("DELETE request to manifest tag should return 202, unless tag deletion is disallowed (400)", func() {
				SkipIfDisabled(management)
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
				SkipIfDisabled(management)
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
				SkipIfDisabled(management)
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifestDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					Equal(http.StatusNotFound),
					Equal(http.StatusOK),
				))
			})

			g.Specify("GET request to tags list should reflect manifest deletion", func() {
				SkipIfDisabled(management)
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
				RunOnlyIf(runManagementSetup)
				SkipIfDisabled(management)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(blobDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusAccepted))
			})

			g.Specify("GET request to deleted blob URL should yield 404 response", func() {
				RunOnlyIf(runManagementSetup)
				SkipIfDisabled(management)
				req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>", reggie.WithDigest(blobDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})
		})

		g.Context("Teardown", func() {
			// TODO: delete blob+tag?
			// No teardown required at this time for content management tests
		})
	})
}
