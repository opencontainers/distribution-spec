package conformance

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test06TagsList = func() {
	g.Context("Tags List", func() {
		g.Specify("GET request to list tags should yield 200 response", func() {
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
			lastTagList = &TagList{
				Name: tagList.Name,
				Tags: tagList.Tags,
			}
		})

		g.Specify("GET start of tag is set by `last` query parameter", func() {
			numResults := numTags / 2
			Expect(len(lastTagList.Tags)).To(BeNumerically(">=", 0))
			lastTag := lastTagList.Tags[numResults-1]
			req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list").
				SetQueryParam("n", strconv.Itoa(numResults)).
				SetQueryParam("last", lastTag)
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
}
