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
	"net/http"
	"os"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var test01Pull = func() {
	g.Context(titlePull, func() {

		var tag string
		var blobRefs []string
		var manifestRefs []string

		g.Context("Setup", func() {
			g.Specify("Populate registry with test blob", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				blobRefs = pushBlob(
					&BlobInfo{
						Digest:  configs[0].Digest,
						Content: configs[0].Content,
						Length:  configs[0].ContentLength,
					},
					blobRefs, g.GinkgoT(),
				)
			})

			g.Specify("Populate registry with test blob", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				blobRefs = pushBlob(
					&BlobInfo{
						Digest:  configs[1].Digest,
						Content: configs[1].Content,
						Length:  configs[1].ContentLength,
					},
					blobRefs, g.GinkgoT(),
				)
			})

			g.Specify("Populate registry with test layer", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				blobRefs = pushBlob(
					&BlobInfo{
						Digest:  layerBlobDigest,
						Content: layerBlobData,
						Length:  layerBlobContentLength,
					},
					blobRefs, g.GinkgoT(),
				)
			})

			g.Specify("Populate registry with test manifest", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				tag = testTagName
				manifestRefs = pushManifest(
					&ManifestInfo{
						Tag:     tag,
						Digest:  manifests[0].Digest,
						Content: manifests[0].Content},
					manifestRefs, g.GinkgoT(),
				)
			})

			g.Specify("Populate registry with test manifest", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				manifestRefs = pushManifest(
					&ManifestInfo{
						Digest:  manifests[1].Digest,
						Content: manifests[1].Content,
					},
					manifestRefs, g.GinkgoT(),
				)
			})

			g.Specify("Get tag name from environment", func() {
				SkipIfDisabled(pull)
				RunOnlyIfNot(runPullSetup)
				tmp := os.Getenv(envVarTagName)
				if tmp != "" {
					tag = tmp
				}
			})
		})

		g.Context("Pull blobs", func() {
			g.Specify("HEAD request to nonexistent blob should result in 404 response", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.HEAD, "/v2/<name>/blobs/<digest>",
					reggie.WithDigest(dummyDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("HEAD request to existing blob should yield 200", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.HEAD, "/v2/<name>/blobs/<digest>",
					reggie.WithDigest(configs[0].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				if h := resp.Header().Get("Docker-Content-Digest"); h != "" {
					Expect(h).To(Equal(configs[0].Digest))
				}
			})

			g.Specify("GET nonexistent blob should result in 404 response", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>",
					reggie.WithDigest(dummyDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("GET request to existing blob URL should yield 200", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>", reggie.WithDigest(configs[0].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})

		g.Context("Pull manifests", func() {
			g.Specify("HEAD request to nonexistent manifest should return 404", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.HEAD, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(nonexistentManifest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("HEAD request to manifest[0] path (digest) should yield 200 response", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.HEAD, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[0].Digest)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				if h := resp.Header().Get("Docker-Content-Digest"); h != "" {
					Expect(h).To(Equal(manifests[0].Digest))
				}
			})

			g.Specify("HEAD request to manifest[1] path (digest) should yield 200 response", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.HEAD, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[1].Digest)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				if h := resp.Header().Get("Docker-Content-Digest"); h != "" {
					Expect(h).To(Equal(manifests[1].Digest))
				}
			})

			g.Specify("HEAD request to manifest path (tag) should yield 200 response", func() {
				SkipIfDisabled(pull)
				Expect(tag).ToNot(BeEmpty())
				req := client.NewRequest(reggie.HEAD, "/v2/<name>/manifests/<reference>", reggie.WithReference(tag)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				if h := resp.Header().Get("Docker-Content-Digest"); h != "" {
					Expect(h).To(Equal(manifests[0].Digest))
				}
			})

			g.Specify("GET nonexistent manifest should return 404", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(nonexistentManifest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("GET request to manifest[0] path (digest) should yield 200 response", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[0].Digest)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})

			g.Specify("GET request to manifest[1] path (digest) should yield 200 response", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[1].Digest)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})

			g.Specify("GET request to manifest path (tag) should yield 200 response", func() {
				SkipIfDisabled(pull)
				Expect(tag).ToNot(BeEmpty())
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>", reggie.WithReference(tag)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})

		g.Context("Error codes", func() {
			g.Specify("400 response body should contain OCI-conforming JSON message", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>",
					reggie.WithReference("sha256:totallywrong")).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(invalidManifestContent)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					Equal(http.StatusBadRequest),
					Equal(http.StatusNotFound)))
				if resp.StatusCode() == http.StatusBadRequest {
					errorResponses, err := resp.Errors()
					Expect(err).To(BeNil())

					Expect(errorResponses).ToNot(BeEmpty())
					Expect(errorCodes).To(ContainElement(errorResponses[0].Code))
				}
			})
		})

		g.Context("Teardown", func() {
			if deleteManifestBeforeBlobs {
				g.Specify("Delete manifests created in setup", func() {
					SkipIfDisabled(pull)
					RunOnlyIf(runPullSetup)
					deleteManifests(manifestRefs, g.GinkgoT())
				})
			}

			g.Specify("Delete blobs created in setup", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				deleteBlobs(blobRefs, g.GinkgoT())
			})
			if !deleteManifestBeforeBlobs {
				g.Specify("Delete manifests created in setup", func() {
					SkipIfDisabled(pull)
					RunOnlyIf(runPullSetup)
					deleteManifests(manifestRefs, g.GinkgoT())
				})
			}
		})
	})
}
