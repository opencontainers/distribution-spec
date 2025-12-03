// Copyright contributors to the Open Containers Distribution Specification
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conformance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	godigest "github.com/opencontainers/go-digest"
)

var test03ContentDiscovery = func() {
	g.Context(titleContentDiscovery, func() {

		var numTags = 4
		var tagList []string
		var blobRefs []string
		var manifestRefs []string

		g.Context("Setup", func() {
			g.Specify("Populate registry with test blob", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				blobRefs = pushBlob(
					&BlobInfo{
						Digest:  configs[2].Digest,
						Content: configs[2].Content,
						Length:  configs[2].ContentLength,
					},
					blobRefs, g.GinkgoT(),
				)
			})

			g.Specify("Populate registry with test layer", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				blobRefs = pushBlob(
					&BlobInfo{
						Digest:  layerBlobDigest,
						Content: layerBlobData,
						Length:  layerBlobContentLength,
					},
					blobRefs, g.GinkgoT(),
				)
			})

			g.Specify("Populate registry with test tags", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				for i := 0; i < numTags; i++ {
					for _, tag := range []string{"test" + strconv.Itoa(i), "TEST" + strconv.Itoa(i)} {
						tagList = append(tagList, tag)
						manifestRefs = pushManifest(
							&ManifestInfo{
								Tag:     tag,
								Digest:  manifests[2].Digest,
								Content: manifests[2].Content,
							},
							manifestRefs, g.GinkgoT(),
						)
					}
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

			g.Specify("References setup", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)

				// Populate registry with empty JSON blob
				// validate expected empty JSON blob digest
				Expect(emptyJSONDescriptor.Digest).To(Equal(godigest.Digest("sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a")))
				blobRefs = pushBlob(
					&BlobInfo{
						Digest:  emptyJSONDescriptor.Digest.String(),
						Content: emptyJSONBlob,
						Length:  fmt.Sprintf("%d", emptyJSONDescriptor.Size),
					},
					blobRefs, g.GinkgoT(),
				)

				// Populate registry with reference blob before the image manifest is pushed
				blobRefs = pushBlob(
					&BlobInfo{
						Digest:  testRefBlobADigest,
						Content: testRefBlobA,
						Length:  testRefBlobALength,
					},
					blobRefs, g.GinkgoT(),
				)

				// Populate registry with test references manifest (config.MediaType = artifactType)
				manifestRefs = pushManifest(
					&ManifestInfo{
						Digest:  refsManifestAConfigArtifactDigest,
						Content: refsManifestAConfigArtifactContent,
						Subject: manifests[4].Digest,
					},
					manifestRefs, g.GinkgoT(),
				)

				// Populate registry with test references manifest (ArtifactType, config.MediaType = emptyJSON)
				manifestRefs = pushManifest(
					&ManifestInfo{
						Digest:  refsManifestALayerArtifactDigest,
						Content: refsManifestALayerArtifactContent,
						Subject: manifests[4].Digest,
					},
					manifestRefs, g.GinkgoT(),
				)

				// Populate registry with test index manifest
				manifestRefs = pushManifest(
					&ManifestInfo{
						Index:   true,
						Digest:  refsIndexArtifactDigest,
						Content: refsIndexArtifactContent,
						Subject: manifests[4].Digest,
					},
					manifestRefs, g.GinkgoT(),
				)

				// Populate registry with test blob
				blobRefs = pushBlob(
					&BlobInfo{
						Digest:  configs[4].Digest,
						Content: configs[4].Content,
						Length:  configs[4].ContentLength,
					},
					blobRefs, g.GinkgoT(),
				)

				// Populate registry with test layer
				blobRefs = pushBlob(
					&BlobInfo{
						Digest:  layerBlobDigest,
						Content: layerBlobData,
						Length:  layerBlobContentLength,
					},
					blobRefs, g.GinkgoT(),
				)

				// Populate registry with test manifest
				tag := testTagName
				manifestRefs = pushManifest(
					&ManifestInfo{
						Tag:     tag,
						Digest:  manifests[4].Digest,
						Content: manifests[4].Content,
					},
					manifestRefs, g.GinkgoT(),
				)

				// Populate registry with reference blob after the image manifest is pushed
				blobRefs = pushBlob(
					&BlobInfo{
						Digest:  testRefBlobBDigest,
						Content: testRefBlobB,
						Length:  testRefBlobBLength,
					},
					blobRefs, g.GinkgoT(),
				)

				// Populate registry with test references manifest (config.MediaType = artifactType)
				manifestRefs = pushManifest(
					&ManifestInfo{
						Digest:  refsManifestBConfigArtifactDigest,
						Content: refsManifestBConfigArtifactContent,
						Subject: manifests[4].Digest,
					},
					manifestRefs, g.GinkgoT(),
				)

				// Populate registry with test references manifest (ArtifactType, config.MediaType = emptyJSON)
				manifestRefs = pushManifest(
					&ManifestInfo{
						Digest:  refsManifestBLayerArtifactDigest,
						Content: refsManifestBLayerArtifactContent,
						Subject: manifests[4].Digest,
					},
					manifestRefs, g.GinkgoT(),
				)

				// Populate registry with test references manifest to a non-existent subject
				manifestRefs = pushManifest(
					&ManifestInfo{
						Digest:  refsManifestCLayerArtifactDigest,
						Content: refsManifestCLayerArtifactContent,
						Subject: manifests[3].Digest,
					},
					manifestRefs, g.GinkgoT(),
				)
			})
		})

		g.Context("Test content discovery endpoints (listing tags)", func() {
			g.Specify("GET request to list tags should yield 200 response and be in sorted order", func() {
				SkipIfDisabled(contentDiscovery)
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				tagList = getTagList(resp)
				numTags = len(tagList)
				// If the list is not empty, the tags MUST be in lexical order (i.e. case-insensitive alphanumeric order).
				sortedTagListLexical := append([]string{}, tagList...)
				sort.SliceStable(sortedTagListLexical, func(i, j int) bool {
					return strings.ToLower(sortedTagListLexical[i]) < strings.ToLower(sortedTagListLexical[j])
				})
				// Historically, registries have not been lexical, so allow `sort.Strings` to be valid too.
				sortedTagListAsciibetical := append([]string{}, tagList...)
				sort.Strings(sortedTagListAsciibetical)
				Expect(tagList).To(Or(Equal(sortedTagListLexical), Equal(sortedTagListAsciibetical)))
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
				Expect(len(tagList)).To(Equal(numResults))
			})

			g.Specify("GET start of tag is set by `last` query parameter", func() {
				SkipIfDisabled(contentDiscovery)
				numResults := numTags / 2
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list").
					SetQueryParam("n", strconv.Itoa(numResults))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				tagList = getTagList(resp)
				last := tagList[numResults-1]
				req = client.NewRequest(reggie.GET, "/v2/<name>/tags/list").
					SetQueryParam("n", strconv.Itoa(numResults)).
					SetQueryParam("last", tagList[numResults-1])
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				tagList = getTagList(resp)
				Expect(len(tagList)).To(BeNumerically("<=", numResults))
				Expect(tagList).ToNot(ContainElement(last))
			})
		})

		g.Context("Test content discovery endpoints (listing references)", func() {
			g.Specify("GET request to nonexistent blob should result in empty 200 response", func() {
				SkipIfDisabled(contentDiscovery)
				req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>",
					reggie.WithDigest(dummyDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				Expect(resp.Header().Get("Content-Type")).To(Equal("application/vnd.oci.image.index.v1+json"))

				var index index
				err = json.Unmarshal(resp.Body(), &index)
				Expect(err).To(BeNil())
				Expect(len(index.Manifests)).To(BeZero())
			})

			g.Specify("GET request to existing blob should yield 200", func() {
				SkipIfDisabled(contentDiscovery)
				req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>",
					reggie.WithDigest(manifests[4].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				Expect(resp.Header().Get("Content-Type")).To(Equal("application/vnd.oci.image.index.v1+json"))

				var index index
				err = json.Unmarshal(resp.Body(), &index)
				Expect(err).To(BeNil())
				Expect(len(index.Manifests)).To(Equal(5))
				Expect(index.Manifests[0].Digest).ToNot(Equal(index.Manifests[1].Digest))
				for i := 0; i < len(index.Manifests); i++ {
					Expect(len(index.Manifests[i].Annotations)).To(Equal(1))
					Expect(index.Manifests[i].Annotations[testAnnotationKey]).To(Equal(testAnnotationValues[index.Manifests[i].Digest.String()]))
				}
			})

			g.Specify("GET request to existing blob with filter should yield 200", func() {
				SkipIfDisabled(contentDiscovery)
				req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>",
					reggie.WithDigest(manifests[4].Digest)).
					SetQueryParam("artifactType", testRefArtifactTypeA)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				Expect(resp.Header().Get("Content-Type")).To(Equal("application/vnd.oci.image.index.v1+json"))

				var index index
				err = json.Unmarshal(resp.Body(), &index)
				Expect(err).To(BeNil())

				// also check resp header "OCI-Filters-Applied: artifactType" denoting that an artifactType filter was applied
				if resp.Header().Get("OCI-Filters-Applied") != "" {
					Expect(len(index.Manifests)).To(Equal(2))
					Expect(resp.Header().Get("OCI-Filters-Applied")).To(Equal(artifactTypeFilter))
					for i := 0; i < len(index.Manifests); i++ {
						Expect(len(index.Manifests[i].Annotations)).To(Equal(1))
						Expect(index.Manifests[i].Annotations[testAnnotationKey]).To(Equal(testAnnotationValues[index.Manifests[i].Digest.String()]))
					}
				} else {
					Expect(len(index.Manifests)).To(Equal(5))
					for i := 0; i < len(index.Manifests); i++ {
						Expect(len(index.Manifests[i].Annotations)).To(Equal(1))
						Expect(index.Manifests[i].Annotations[testAnnotationKey]).To(Equal(testAnnotationValues[index.Manifests[i].Digest.String()]))
					}
					Warn("filtering by artifact-type is not implemented")
				}
			})

			g.Specify("GET request to missing manifest should yield 200", func() {
				SkipIfDisabled(contentDiscovery)
				req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>",
					reggie.WithDigest(manifests[3].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				Expect(resp.Header().Get("Content-Type")).To(Equal("application/vnd.oci.image.index.v1+json"))

				var index index
				err = json.Unmarshal(resp.Body(), &index)
				Expect(err).To(BeNil())
				Expect(len(index.Manifests)).To(Equal(1))
				Expect(index.Manifests[0].Digest.String()).To(Equal(refsManifestCLayerArtifactDigest))
			})
		})

		g.Context("Teardown", func() {
			if deleteManifestBeforeBlobs {
				g.Specify("Delete created manifest & associated tags", func() {
					SkipIfDisabled(contentDiscovery)
					RunOnlyIf(runContentDiscoverySetup)
					deleteManifests(manifestRefs, g.GinkgoT())
				})
			}

			g.Specify("Delete blobs created in tests", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				deleteBlobs(blobRefs, g.GinkgoT())
			})

			if !deleteManifestBeforeBlobs {
				g.Specify("Delete created manifest & associated tags", func() {
					SkipIfDisabled(contentDiscovery)
					RunOnlyIf(runContentDiscoverySetup)
					deleteManifests(manifestRefs, g.GinkgoT())
				})
			}
		})
	})
}
