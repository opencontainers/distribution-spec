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
				req := client.NewRequest(POST, "/v2/<name>/blobs/uploads/")
				resp, _ := client.Do(req)

				req = client.NewRequest(PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", blobDigest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", fmt.Sprintf("%d", len(configContent))).
					SetBody(configContent)
				resp, _ = client.Do(req)

				tag := testTagName
				req = client.NewRequest(PUT, "/v2/<name>/manifests/<reference>",
					WithReference(tag)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(manifestContent)
				resp, _ = client.Do(req)
			})

			g.Specify("Discovery", func() {
				req := client.NewRequest(GET, "/v2/<name>/tags/list")
				resp, _ := client.Do(req)
				tagResponse = resp
				tagList := &TagList{}
				jsonData := []byte(resp.String())
				json.Unmarshal(jsonData, tagList)
				numTags = len(tagList.Tags)
			})
		})

		g.Context("Pull blobs", func() {
			g.Specify("GET nonexistent blob should result in 404 response", func() {
				req := client.NewRequest(GET, "/v2/<name>/blobs/<digest>",
					WithDigest(dummyDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("GET request to existing blob URL should yield 200", func() {
				req := client.NewRequest(GET, "/v2/<name>/blobs/<digest>", WithDigest(blobDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})

		g.Context("Pull manifests", func () {
			g.Specify("GET nonexistent manifest should return 404", func() {
				req := client.NewRequest(GET, "/v2/<name>/manifests/<reference>",
					WithReference(nonexistentManifest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("GET request to manifest path (digest) should yield 200 response", func() {
				req := client.NewRequest(GET, "/v2/<name>/manifests/<digest>", WithDigest(manifestDigest)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})

			g.Specify("GET request to manifest path (tag) should yield 200 response", func() {
				tag := tagName(lastResponse)
				Expect(tag).ToNot(BeEmpty())
				req := client.NewRequest(GET, "/v2/<name>/manifests/<reference>", WithReference(tag)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})

		g.Context("Teardown", func() {
			g.Specify("Delete manifest created during setup", func() {
				req := client.NewRequest(DELETE, "/v2/<name>/manifests/<digest>", WithDigest(manifestDigest))
				_, _ = client.Do(req)
			})

			g.Specify("Delete blobs created during setup", func() {
				req := client.NewRequest(DELETE, "/v2/<name>/blobs/<digest>", WithDigest(blobDigest))
				_, _ = client.Do(req)
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

	if tn := os.Getenv(envVarTagName); tn != "" {
		return tn
	}

	return testTagName
}
