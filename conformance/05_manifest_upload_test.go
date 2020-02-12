package conformance

import (
	"fmt"
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test05ManifestUpload = func() {
	g.Context("Manifest Upload", func() {
		g.Specify("GET nonexistent manifest should return 404", func() {
			req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>",
				reggie.WithReference(nonexistentManifest))
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
		})

		g.Specify("PUT should accept a manifest upload", func() {
			for i := 0; i < 4; i++ {
				tag := fmt.Sprintf("test%d", i)
				if i == 0 {
					firstTag = tag
				}
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
			req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifestDigest)).
				SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusOK))
		})
	})
}
