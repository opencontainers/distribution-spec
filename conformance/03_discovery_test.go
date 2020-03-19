package conformance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	tagList []string
)

var test03Discovery = func() {
	g.Context("Discovery", func() {
		g.Context("Setup", func() {
			g.Specify("Push tags to repository", func() {
				SkipIfDisabled(push)
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
				tagList = getTagList(resp)
				Expect(err).To(BeNil())
				Expect(len(tagList)).To(BeNumerically("<=", numResults))
				Expect(tagList).To(ContainElement(tagList[numResults]))
			})
		})
	})
}

func getTagList(resp *reggie.Response) []string {
	if userDisabled(push) {
		return strings.Split(os.Getenv(envVarTagList), ",")
	}

	jsonData := resp.Body()
	tagList := &TagList{}
	err := json.Unmarshal(jsonData, tagList)
	if err != nil {
		return []string{}
	}

	return tagList.Tags
}
