package conformance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	. "github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	tagResponse *Response
)

var test01Pull = func() {
	g.Context("Pull", func() {
		g.Context("Setup", func() {
			g.Specify("Push", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				location := resp.Header().Get("Location")
				Expect(location).ToNot(BeEmpty())

				req = client.NewRequest(PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", blobDigest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", fmt.Sprintf("%d", len(configContent))).
					SetBody(configContent)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
			})

			g.Specify("Discovery", func() {
				SkipIfDisabled(discovery)
				req := client.NewRequest(GET, "/v2/<name>/tags/list")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				tagResponse = resp
				tagList := &TagList{}
				jsonData := []byte(resp.String())
				err = json.Unmarshal(jsonData, tagList)
				Expect(err).To(BeNil())
				numTags = len(tagList.Tags)
			})
		})

		g.Context("Test pull endpoints", func() {

			g.Specify("GET nonexistent blob should result in 404 response", func() {
				req := client.NewRequest(GET, "/v2/<name>/blobs/<digest>",
					WithDigest(dummyDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
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

			g.Specify("GET request to manifest URL (tag) should yield 200 response", func() {
				tag := tagName(lastResponse)
				Expect(tag).ToNot(BeEmpty())
				req := client.NewRequest(GET, "/v2/<name>/manifests/<reference>", WithReference(tag)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})
	})
}

func tagName(lastResponse *Response) string {
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
