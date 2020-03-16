package conformance

import (
	"fmt"
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test02Push = func() {
	g.Context("Push", func() {

		var lastResponse *reggie.Response

		g.Context("Setup", func() {

		})

		g.Context("Blob Upload Streamed", func() {
			g.Specify("PATCH request with blob in body should yield 202 response", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				location := resp.Header().Get("Location")
				Expect(location).ToNot(BeEmpty())

				req = client.NewRequest(reggie.PATCH, resp.GetRelativeLocation()).
					SetHeader("Content-Type", "application/octet-stream").
					SetBody(blobA)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusAccepted))
				lastResponse = resp
			})

			g.Specify("PUT request to session URL with digest should yield 201 response", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.PUT, lastResponse.GetRelativeLocation()).
					SetQueryParam("digest", blobADigest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", blobALength)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
				location := resp.Header().Get("Location")
				Expect(location).ToNot(BeEmpty())
			})
		})

		g.Context("Blob Upload Monolithic", func() {
			g.Specify("GET nonexistent blob should result in 404 response", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>",
					reggie.WithDigest(dummyDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("POST request with digest and blob should yield a 201", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/").
					SetHeader("Content-Length", configContentLength).
					SetHeader("Content-Type", "application/octet-stream").
					SetQueryParam("digest", blobDigest).
					SetBody(configContent)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				location := resp.Header().Get("Location")
				Expect(location).ToNot(BeEmpty())
				Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
			})

			g.Specify("GET request to blob URL from prior request should yield 200", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>", reggie.WithDigest(blobDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})

			g.Specify("POST request should yield a session ID", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusAccepted))
				lastResponse = resp
			})

			g.Specify("PUT upload of a blob should yield a 201 Response", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.PUT, lastResponse.GetRelativeLocation()).
					SetHeader("Content-Length", configContentLength).
					SetHeader("Content-Type", "application/octet-stream").
					SetQueryParam("digest", blobDigest).
					SetBody(configContent)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				location := resp.Header().Get("Location")
				Expect(location).ToNot(BeEmpty())
				Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
			})

			g.Specify("GET request to existing blob should yield 200 response", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>", reggie.WithDigest(blobDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})

		g.Context("Blob Upload Chunked", func() {
			g.Specify("Out-of-order blob upload should return 416", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/").
					SetHeader("Content-Length", "0")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				location := resp.Header().Get("Location")
				Expect(location).ToNot(BeEmpty())

				req = client.NewRequest(reggie.PATCH, resp.GetRelativeLocation()).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", blobBChunk2Length).
					SetHeader("Content-Range", blobBChunk2Range).
					SetBody(blobBChunk2)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusRequestedRangeNotSatisfiable))
			})

			g.Specify("PATCH request with first chunk should return 202", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/").
					SetHeader("Content-Length", "0")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				location := resp.Header().Get("Location")
				Expect(location).ToNot(BeEmpty())

				req = client.NewRequest(reggie.PATCH, resp.GetRelativeLocation()).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", blobBChunk1Length).
					SetHeader("Content-Range", blobBChunk1Range).
					SetBody(blobBChunk1)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusAccepted))
				lastResponse = resp
			})

			g.Specify("PUT request with final chunk should return 201", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.PUT, lastResponse.GetRelativeLocation()).
					SetHeader("Content-Length", blobBChunk2Length).
					SetHeader("Content-Range", blobBChunk2Range).
					SetHeader("Content-Type", "application/octet-stream").
					SetQueryParam("digest", blobBDigest).
					SetBody(blobBChunk2)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				location := resp.Header().Get("Location")
				Expect(location).ToNot(BeEmpty())
				Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
			})
		})

		g.Context("Manifest Upload", func() {
			g.Specify("GET nonexistent manifest should return 404", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(nonexistentManifest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("PUT should accept a manifest upload", func() {
				SkipIfDisabled(push)
				for i := 0; i < 4; i++ {
					tag := fmt.Sprintf("test%d", i)
					req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
						reggie.WithReference(tag)).
						SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
						SetBody(manifestContent)
					resp, err := client.Do(req)
					Expect(err).To(BeNil())
					location := resp.Header().Get("Location")
					Expect(location).ToNot(BeEmpty())
					Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
				}
			})

			g.Specify("GET request to manifest URL (digest) should yield 200 response", func() {
				SkipIfDisabled(push)
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifestDigest)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})
	})
}
