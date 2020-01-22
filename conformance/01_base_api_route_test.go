package conformance

import (
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test01BaseAPIRoute = func() {
	g.Context("Base API Route", func() {
		g.Specify("GET request to base API route must return 200 response", func() {
			req := client.NewRequest(reggie.GET, "/v2/")
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode()).To(Equal(http.StatusOK))
		})
	})
}
