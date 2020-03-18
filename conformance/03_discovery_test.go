package conformance

import (
	"encoding/json"
	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"strconv"
)

var discoveryTest = func() {
	g.Context("Discovery", func() {
		g.Context("Setup", func() {
			g.Specify("Push", func() {

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
				tagList := &TagList{}
				jsonData := []byte(resp.String())
				err = json.Unmarshal(jsonData, tagList)
				Expect(err).To(BeNil())
				numTags = len(tagList.Tags)
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
				jsonData := resp.Body()
				tagList := &TagList{}
				err = json.Unmarshal(jsonData, tagList)
				Expect(err).To(BeNil())
				Expect(len(tagList.Tags)).To(Equal(numResults))
				lastTagList = *tagList
			})

			g.Specify("GET start of tag is set by `last` query parameter", func() {
				SkipIfDisabled(discovery)
				numResults := numTags / 2
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list").
					SetQueryParam("n", strconv.Itoa(numResults)).
					SetQueryParam("last", lastTagList.Tags[numResults-1])
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				jsonData := resp.Body()
				tagList := &TagList{}
				err = json.Unmarshal(jsonData, tagList)
				Expect(err).To(BeNil())
				Expect(tagList.Tags).To(ContainElement("test3"))
			})
		})
	})
}
