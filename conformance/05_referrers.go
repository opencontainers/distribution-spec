package conformance

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	godigest "github.com/opencontainers/go-digest"
)

var test05Referrers = func() {
	g.Context(titleReferrers, func() {

		g.Context("Setup", func() {
			g.Specify("Populate registry with empty JSON blob", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				// validate expected empty JSON blob digest
				Expect(emptyJSONDescriptor.Digest).To(Equal(godigest.Digest("sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a")))
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", emptyJSONDescriptor.Digest.String()).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", fmt.Sprintf("%d", emptyJSONDescriptor.Size)).
					SetBody(emptyJSONBlob)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Populate registry with reference blob before the image manifest is pushed", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", testRefBlobADigest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", testRefBlobALength).
					SetBody(testRefBlobA)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Populate registry with test references manifest (config.MediaType = artifactType)", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(refsManifestAConfigArtifactDigest)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(refsManifestAConfigArtifactContent)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
				Expect(resp.Header().Get("OCI-Subject")).To(Equal(manifests[4].Digest))
			})

			g.Specify("Populate registry with test references manifest (ArtifactType, config.MediaType = emptyJSON)", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(refsManifestALayerArtifactDigest)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(refsManifestALayerArtifactContent)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
				Expect(resp.Header().Get("OCI-Subject")).To(Equal(manifests[4].Digest))
			})

			g.Specify("Populate registry with test blob", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", configs[4].Digest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", configs[4].ContentLength).
					SetBody(configs[4].Content)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Populate registry with test layer", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
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

			g.Specify("Populate registry with test manifest", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				tag := testTagName
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(tag)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(manifests[4].Content)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Populate registry with reference blob after the image manifest is pushed", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", testRefBlobBDigest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", testRefBlobBLength).
					SetBody(testRefBlobB)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Populate registry with test references manifest (config.MediaType = artifactType)", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(refsManifestBConfigArtifactDigest)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(refsManifestBConfigArtifactContent)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
				Expect(resp.Header().Get("OCI-Subject")).To(Equal(manifests[4].Digest))
			})

			g.Specify("Populate registry with test references manifest (ArtifactType, config.MediaType = emptyJSON)", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(refsManifestBLayerArtifactDigest)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(refsManifestBLayerArtifactContent)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
				Expect(resp.Header().Get("OCI-Subject")).To(Equal(manifests[4].Digest))
			})
		})

		g.Context("Get references", func() {
			g.Specify("GET request to nonexistent blob should result in empty 200 response", func() {
				SkipIfDisabled(referrers)
				req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>",
					reggie.WithDigest(dummyDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))

				var index Index
				err = json.Unmarshal(resp.Body(), &index)
				Expect(err).To(BeNil())
				Expect(len(index.Manifests)).To(BeZero())
			})

			g.Specify("GET request to existing blob should yield 200", func() {
				SkipIfDisabled(referrers)
				req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>",
					reggie.WithDigest(manifests[4].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				if h := resp.Header().Get("Docker-Content-Digest"); h != "" {
					Expect(h).To(Equal(configs[4].Digest))
				}

				var index Index
				err = json.Unmarshal(resp.Body(), &index)
				Expect(err).To(BeNil())
				Expect(len(index.Manifests)).To(Equal(4))
				Expect(index.Manifests[0].Digest).ToNot(Equal(index.Manifests[1].Digest))
			})

			g.Specify("GET request to existing blob with filter should yield 200", func() {
				SkipIfDisabled(referrers)
				req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>",
					reggie.WithDigest(manifests[4].Digest)).
					SetQueryParam("artifactType", testRefArtifactTypeA)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				if h := resp.Header().Get("Docker-Content-Digest"); h != "" {
					Expect(h).To(Equal(configs[4].Digest))
				}

				var index Index
				err = json.Unmarshal(resp.Body(), &index)
				Expect(err).To(BeNil())

				// also check resp header "OCI-Filters-Applied: artifactType" denoting that an artifactType filter was applied
				if resp.Header().Get("OCI-Filters-Applied") != "" {
					Expect(len(index.Manifests)).To(Equal(2))
					Expect(resp.Header().Get("OCI-Filters-Applied")).To(Equal(testRefArtifactTypeA))
				} else {
					Expect(len(index.Manifests)).To(Equal(4))
					Warn("filtering by artifact-type is not implemented")
				}
			})
		})

		g.Context("Teardown", func() {
			deleteReq := func(req *reggie.Request) {
				SkipIfDisabled(push)
				RunOnlyIf(runReferencesSetup)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					SatisfyAll(
						BeNumerically(">=", 200),
						BeNumerically("<", 300),
					),
					Equal(http.StatusMethodNotAllowed),
				))
			}

			if deleteManifestBeforeBlobs {
				g.Specify("Delete manifest created in setup", func() {
					SkipIfDisabled(referrers)
					RunOnlyIf(runReferencesSetup)
					req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
						reggie.WithReference(refsManifestAConfigArtifactDigest))
					deleteReq(req)
					req = client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
						reggie.WithReference(refsManifestALayerArtifactDigest))
					deleteReq(req)
					req = client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[4].Digest))
					deleteReq(req)
					req = client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
						reggie.WithReference(refsManifestBConfigArtifactDigest))
					deleteReq(req)
					req = client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
						reggie.WithReference(refsManifestBLayerArtifactDigest))
					deleteReq(req)
				})
			}

			g.Specify("Delete config blob created in setup", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(configs[4].Digest))
				deleteReq(req)
			})

			g.Specify("Delete layer blob created in setup", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(layerBlobDigest))
				deleteReq(req)
			})

			g.Specify("Delete reference blob created in setup", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(testRefBlobADigest))
				deleteReq(req)
				req = client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(testRefBlobBDigest))
				deleteReq(req)
			})

			g.Specify("Delete empty JSON blob created in setup", func() {
				SkipIfDisabled(referrers)
				RunOnlyIf(runReferencesSetup)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(emptyJSONDescriptor.Digest.String()))
				deleteReq(req)
			})

			if !deleteManifestBeforeBlobs {
				g.Specify("Delete manifest created in setup", func() {
					SkipIfDisabled(referrers)
					RunOnlyIf(runReferencesSetup)
					req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
						reggie.WithReference(refsManifestAConfigArtifactDigest))
					deleteReq(req)
					req = client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
						reggie.WithReference(refsManifestALayerArtifactDigest))
					deleteReq(req)
					req = client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[4].Digest))
					deleteReq(req)
					req = client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
						reggie.WithReference(refsManifestBConfigArtifactDigest))
					deleteReq(req)
					req = client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
						reggie.WithReference(refsManifestBLayerArtifactDigest))
					deleteReq(req)
				})
			}
		})
	})
}
