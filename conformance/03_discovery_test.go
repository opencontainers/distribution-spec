package conformance

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test03ContentDiscovery = func() {
	g.Context(titleContentDiscovery, func() {

		var numTags = 4
		var tagList []string

		g.Context("Setup", func() {
			g.Specify("Populate registry with test blobs", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				for i := 1; i <= 2; i++ {
					req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
					resp, err := client.Do(req)
					Expect(err).To(BeNil())
					req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
						SetQueryParam("digest", configs[i].Digest).
						SetHeader("Content-Type", "application/octet-stream").
						SetHeader("Content-Length", configs[i].ContentLength).
						SetBody(configs[i].Content)
					resp, err = client.Do(req)
					Expect(err).To(BeNil())
					Expect(resp.StatusCode()).To(SatisfyAll(
						BeNumerically(">=", 200),
						BeNumerically("<", 300)))
				}
			})

			g.Specify("Populate registry with test layer", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", layerBlobDigest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", layerBlobContentLength).
					SetBody(layerBlobData)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Populate registry with referred-to manifest", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference("test-something-points-to-me")).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(manifests[1].Content)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Populate registry with test tags", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				for i := 0; i < numTags; i++ {
					tag := fmt.Sprintf("test%d", i)
					tagList = append(tagList, tag)
					req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
						reggie.WithReference(tag)).
						SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
						SetBody(manifests[2].Content)
					resp, err := client.Do(req)
					Expect(err).To(BeNil())
					Expect(resp.StatusCode()).To(SatisfyAll(
						BeNumerically(">=", 200),
						BeNumerically("<", 300)))
				}
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list")
				resp, err := client.Do(req)
				tagList = getTagList(resp)
				_ = err
			})

			g.Specify("Populate registry with test tags (no push)", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIfNot(runContentDiscoverySetup)
				tagList = strings.Split(os.Getenv(envVarTagList), ",")
			})
		})

		g.Context("Test content discovery endpoints", func() {
			g.Specify("GET request to list tags should yield 200 response", func() {
				SkipIfDisabled(contentDiscovery)
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				numTags = len(tagList)
			})

			g.Specify("GET number of tags should be limitable by `n` query parameter", func() {
				SkipIfDisabled(contentDiscovery)
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
				SkipIfDisabled(contentDiscovery)
				numResults := numTags / 2
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list").
					SetQueryParam("n", strconv.Itoa(numResults))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
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

			g.Specify("GET request to list of referrers should yield 200 response", func() {
				SkipIfDisabled(contentDiscovery)
				// TODO: should move to this form per the spec,
				// the endpoint used is supported currently by oci-playground/distribution
				// req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>", reggie.WithDigest(manifests[2].Digest))
				req := client.NewRequest(reggie.GET, "/v2/<name>/_oci/artifacts/referrers")
				// set the digest to the one being pointed to by manifests[2]
				req.QueryParam.Add("digest", manifests[1].Digest)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})

		g.Context("Teardown", func() {
			if deleteManifestBeforeBlobs {
				g.Specify("Delete created manifest & associated tags", func() {
					SkipIfDisabled(contentDiscovery)
					RunOnlyIf(runContentDiscoverySetup)
					req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[2].Digest))
					resp, err := client.Do(req)
					Expect(err).To(BeNil())
					Expect(resp.StatusCode()).To(SatisfyAny(
						SatisfyAll(
							BeNumerically(">=", 200),
							BeNumerically("<", 300),
						),
						Equal(http.StatusMethodNotAllowed),
					))
				})
			}

			g.Specify("Delete config blob created in tests", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(configs[2].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					SatisfyAll(
						BeNumerically(">=", 200),
						BeNumerically("<", 300),
					),
					Equal(http.StatusMethodNotAllowed),
				))
			})

			g.Specify("Delete layer blob created in setup", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(layerBlobDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					SatisfyAll(
						BeNumerically(">=", 200),
						BeNumerically("<", 300),
					),
					Equal(http.StatusMethodNotAllowed),
				))
			})

			if !deleteManifestBeforeBlobs {
				g.Specify("Delete created manifest & associated tags", func() {
					SkipIfDisabled(contentDiscovery)
					RunOnlyIf(runContentDiscoverySetup)
					req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[2].Digest))
					resp, err := client.Do(req)
					Expect(err).To(BeNil())
					Expect(resp.StatusCode()).To(SatisfyAny(
						SatisfyAll(
							BeNumerically(">=", 200),
							BeNumerically("<", 300),
						),
						Equal(http.StatusMethodNotAllowed),
					))
				})
			}
		})
	})
}
