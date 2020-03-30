package conformance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test03Discovery = func() {
	g.Context("Discovery", func() {

		var numTags = 4
		var tagList []string
		var lastResponse *reggie.Response

		g.Context("Setup", func() {
			g.Specify("Populate registry with test tags", func() {
				RunOnlyIf(runDiscoverySetup)
				SkipIfDisabled(discovery)
				for i := 0; i < numTags; i++ {
					tag := fmt.Sprintf("test%d", i)
					req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
						reggie.WithReference(tag)).
						SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
						SetBody(manifestContent)
					client.Do(req)
				}
			})
		})

		g.Context("Test discovery endpoints", func() {
			g.Specify("GET request to list tags should yield 200 response", func() {
				SkipIfDisabled(discovery)
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				lastResponse = resp
				tagList = getTagList(resp)
				Expect(err).To(BeNil())
				numTags = len(tagList)
			})

			g.Specify("GET request to manifest URL (tag) should yield 200 response", func() {
				SkipIfDisabled(discovery)
				tl := &TagList{}
				jsonData := lastResponse.Body()
				err := json.Unmarshal(jsonData, tl)
				Expect(err).To(BeNil())
				Expect(tl.Tags).ToNot(BeEmpty())
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(tl.Tags[0])).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})

			g.Specify("GET number of tags should be limitable by `n` query parameter", func() {
				SkipIfDisabled(discovery)
				numResults := numTags / 2
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list").
					SetQueryParam("n", strconv.Itoa(numResults))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				tagList = getTagList(resp)
				Expect(err).To(BeNil())
				Expect(len(tagList)).To(Equal(numResults))
			})

			g.Specify("GET start of tag is set by `last` query parameter", func() {
				SkipIfDisabled(discovery)
				numResults := numTags / 2
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list").
					SetQueryParam("n", strconv.Itoa(numResults))
				resp, err := client.Do(req)
				tagList = getTagList(resp)
				req = client.NewRequest(reggie.GET, "/v2/<name>/tags/list").
					SetQueryParam("n", strconv.Itoa(numResults)).
					SetQueryParam("last", tagList[numResults-1])
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				Expect(err).To(BeNil())
				Expect(len(tagList)).To(BeNumerically("<=", numResults))
				Expect(tagList).To(ContainElement(tagList[numResults-1]))
			})
		})

		g.Context("Teardown", func() {
			// TODO: delete tags?
			// No teardown required at this time for discovery tests
		})
	})
}
