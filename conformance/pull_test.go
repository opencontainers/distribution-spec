package conformance

import (
	"encoding/json"
	. "github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"os"
)

var pullTest = func() {
	g.Context("Test pull endpoints", func() {
		g.Specify("GET nonexistent blob should result in 404 response", func() {
			req := client.NewRequest(GET, "/v2/<name>/blobs/<digest>",
				WithDigest(dummyDigest))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
		})

		g.Specify("PATCH request with blob in body should yield 202 response", func() {
			SkipIfNotEnabled(push)
			req := client.NewRequest(POST, "/v2/<name>/blobs/uploads/")
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			location := resp.Header().Get("Location")
			Expect(location).ToNot(BeEmpty())

			req = client.NewRequest(PATCH, resp.GetRelativeLocation()).
				SetHeader("Content-Type", "application/octet-stream").
				SetBody(blobA)
			resp, err = client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusAccepted))
			lastResponse = resp
		})

		g.Specify("PUT request to session URL with digest should yield 201 response", func() {
			SkipIfNotEnabled(push)
			req := client.NewRequest(PUT, lastResponse.GetRelativeLocation()).
				SetQueryParam("digest", blobADigest).
				SetHeader("Content-Type", "application/octet-stream").
				SetHeader("Content-Length", blobALength)
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
			location := resp.Header().Get("Location")
			Expect(location).ToNot(BeEmpty())
		})

		g.Specify("GET request to blob URL from prior request should yield 200", func() {
			req := client.NewRequest(GET, "/v2/<name>/blobs/<digest>", WithDigest(blobDigest))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusOK))
		})

		g.Specify("GET nonexistent manifest should return 404", func() {
			req := client.NewRequest(GET, "/v2/<name>/manifests/<reference>",
				WithReference(nonexistentManifest))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
		})

		g.Specify("GET request to manifest URL (digest) should yield 200 response", func() {
			req := client.NewRequest(GET, "/v2/<name>/manifests/<digest>", WithDigest(manifestDigest)).
				SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusOK))
		})

		g.Specify("GET request to list tags should yield 200 response", func() {
			SkipIfNotEnabled(discovery)
			req := client.NewRequest(GET, "/v2/<name>/tags/list")
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			lastResponse = resp
			tagList := &TagList{}
			jsonData := []byte(resp.String())
			err = json.Unmarshal(jsonData, tagList)
			Expect(err).To(BeNil())
			numTags = len(tagList.Tags)
		})

		g.Specify("GET request to manifest URL (tag) should yield 200 response", func() {
			tag := tagName()
			Expect(tag).ToNot(BeEmpty())
			req := client.NewRequest(GET, "/v2/<name>/manifests/<reference>", WithReference(tag)).
				SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusOK))
		})
	})
}

func tagName() string {
	tl := &TagList{}
	if lastResponse != nil {
		jsonData := lastResponse.Body()
		err := json.Unmarshal(jsonData, tl)
		if err != nil && len(tl.Tags) > 0 {
			return tl.Tags[0]
		}
	}

	return os.Getenv(envVarTagName)
}
