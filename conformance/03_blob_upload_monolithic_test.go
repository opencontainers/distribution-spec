package conformance

import (
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test03BlobUploadMonolithic = func() {
	g.Context("Blob Upload Monolithic", func() {
		g.Specify("GET nonexistent blob should result in 404 response", func() {
			req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>",
				reggie.WithDigest(dummyDigest))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
		})

		g.Specify("POST request should yield a session ID", func() {
			req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusAccepted))
			lastResponse = resp
		})

		g.Specify("PUT upload of a blob should yield a 201 Response", func() {
			req := client.NewRequest(reggie.PUT, lastResponse.GetRelativeLocation()).
				SetHeader("Content-Length", configContentLength).
				SetHeader("Content-Type", "application/octet-stream").
				SetQueryParam("digest", configDigest).
				SetBody(configContent)
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
		})

		g.Specify("GET request to existing blob should yield 200 response", func() {
			req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>", reggie.WithDigest(configDigest))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusOK))
		})
	})
}
