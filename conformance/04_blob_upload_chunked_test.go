package conformance

import (
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test04BlobUploadChunked = func() {
	g.Context("Blob Upload Chunked", func() {
		g.Specify("PATCH request with first chunk should return 202", func() {
			req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/").
				SetHeader("Content-Length", "0")
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			lastResponse = resp

			req = client.NewRequest(reggie.PATCH, lastResponse.GetRelativeLocation()).
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
			req := client.NewRequest(reggie.PUT, lastResponse.GetRelativeLocation()).
				SetHeader("Content-Length", blobBChunk2Length).
				SetHeader("Content-Range", blobBChunk2Range).
				SetHeader("Content-Type", "application/octet-stream").
				SetQueryParam("digest", blobBDigest).
				SetBody(blobBChunk2)
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
		})
	})
}
