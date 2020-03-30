package conformance

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test01Pull = func() {
	g.Context("Pull", func() {

		var tagResponse *reggie.Response

		g.Context("Setup", func() {
			g.Specify("Populate registry with test blob", func() {
				RunOnlyIf(runPullSetup)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, _ := client.Do(req)
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", blobDigest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", fmt.Sprintf("%d", len(configContent))).
					SetBody(configContent)
				client.Do(req)
			})

			g.Specify("Populate registry with test manifest", func() {
				RunOnlyIf(runPullSetup)
				tag := testTagName
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(tag)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(manifestContent)
				client.Do(req)
			})

			g.Specify("Retrieve list of current tags", func() {
				RunOnlyIf(runPullSetup)
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list")
				resp, _ := client.Do(req)
				tagResponse = resp
				tagList := &TagList{}
				jsonData := []byte(resp.String())
				json.Unmarshal(jsonData, tagList)
			})

			g.Specify("Create mock registry response", func() {
				RunOnlyIfNot(runPullSetup)
				// TODO: handle
				// OCI_MANIFEST_DIGEST=<digest>
				// OCI_TAG_NAME=<tag>
				// OCI_BLOB_DIGEST=<digest>
				//tagResponse = &reggie.Response{}
				//tagResponse.RawResponse = &http.Response{}
				//tagResponse.RawResponse.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(``)))
			})
		})

		g.Context("Pull blobs", func() {
			g.Specify("GET nonexistent blob should result in 404 response", func() {
				req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>",
					reggie.WithDigest(dummyDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("GET request to existing blob URL should yield 200", func() {
				req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>", reggie.WithDigest(blobDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})

		g.Context("Pull manifests", func() {
			g.Specify("GET nonexistent manifest should return 404", func() {
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(nonexistentManifest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("GET request to manifest path (digest) should yield 200 response", func() {
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifestDigest)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})

			g.Specify("GET request to manifest path (tag) should yield 200 response", func() {
				tag := getTagName(tagResponse)
				Expect(tag).ToNot(BeEmpty())
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>", reggie.WithReference(tag)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})

		g.Context("Error codes", func() {
			g.Specify("400 response body should contain OCI-conforming JSON message", func() {
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference("sha256:totallywrong")).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(invalidManifestContent)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					Equal(http.StatusBadRequest),
					Equal(http.StatusNotFound)))
				if resp.StatusCode() == http.StatusBadRequest {
					errorResponses, err := resp.Errors()
					Expect(err).To(BeNil())

					Expect(errorResponses).ToNot(BeEmpty())
					Expect(errorCodes).To(ContainElement(errorResponses[0].Code))
				}
			})
		})

		g.Context("Teardown", func() {
			g.Specify("Delete manifest created in setup", func() {
				RunOnlyIf(runPullSetup)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifestDigest))
				client.Do(req)
			})

			g.Specify("Delete blob created in setup", func() {
				RunOnlyIf(runPullSetup)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(blobDigest))
				client.Do(req)
			})
		})
	})
}
