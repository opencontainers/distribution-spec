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
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	mathrand "math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/bloodorangeio/reggie"
	"github.com/google/uuid"
	g "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/formatter"
	. "github.com/onsi/gomega"
	godigest "github.com/opencontainers/go-digest"
)

type (
	TagList struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	TestBlob struct {
		Content       []byte
		ContentLength string
		Digest        string
	}
)

const (
	pull = 1 << iota
	push
	contentDiscovery
	contentManagement
	numWorkflows

	BLOB_UNKNOWN = iota
	BLOB_UPLOAD_INVALID
	BLOB_UPLOAD_UNKNOWN
	DIGEST_INVALID
	MANIFEST_BLOB_UNKNOWN
	MANIFEST_INVALID
	MANIFEST_UNKNOWN
	MANIFEST_UNVERIFIED
	NAME_INVALID
	NAME_UNKNOWN
	SIZE_INVALID
	TAG_INVALID
	UNAUTHORIZED
	DENIED
	UNSUPPORTED

	envVarRootURL                   = "OCI_ROOT_URL"
	envVarNamespace                 = "OCI_NAMESPACE"
	envVarUsername                  = "OCI_USERNAME"
	envVarPassword                  = "OCI_PASSWORD"
	envVarDebug                     = "OCI_DEBUG"
	envVarPull                      = "OCI_TEST_PULL"
	envVarPush                      = "OCI_TEST_PUSH"
	envVarContentDiscovery          = "OCI_TEST_CONTENT_DISCOVERY"
	envVarContentManagement         = "OCI_TEST_CONTENT_MANAGEMENT"
	envVarPushEmptyLayer            = "OCI_SKIP_EMPTY_LAYER_PUSH_TEST"
	envVarBlobDigest                = "OCI_BLOB_DIGEST"
	envVarManifestDigest            = "OCI_MANIFEST_DIGEST"
	envVarTagName                   = "OCI_TAG_NAME"
	envVarTagList                   = "OCI_TAG_LIST"
	envVarHideSkippedWorkflows      = "OCI_HIDE_SKIPPED_WORKFLOWS"
	envVarAuthScope                 = "OCI_AUTH_SCOPE"
	envVarDeleteManifestBeforeBlobs = "OCI_DELETE_MANIFEST_BEFORE_BLOBS"
	envVarCrossmountNamespace       = "OCI_CROSSMOUNT_NAMESPACE"
	envVarAutomaticCrossmount       = "OCI_AUTOMATIC_CROSSMOUNT"
	envVarReportDir                 = "OCI_REPORT_DIR"

	emptyLayerTestTag = "emptylayer"
	testTagName       = "tagtest0"

	titlePull              = "Pull"
	titlePush              = "Push"
	titleContentDiscovery  = "Content Discovery"
	titleContentManagement = "Content Management"

	//	layerBase64String is a base64 encoding of a simple tarball, obtained like this:
	//		$ echo 'you bothered to find out what was in here. Congratulations!' > test.txt
	//		$ tar czvf test.tar.gz test.txt
	//		$ cat test.tar.gz | base64
	layerBase64String = "H4sIAAAAAAAAA+3OQQrCMBCF4a49xXgBSUnaHMCTRBptQRNpp6i3t0UEV7oqIv7fYgbmzeJpHHSjVy0" +
		"WZCa1c/MufWVe94N3RWlrZ72x3k/30nhbFWKWLPU0Dhp6keJ8im//PuU/6pZH2WVtYx8b0Sz7LjWSR5VLG6YRBumSzOlGtjkd+qD" +
		"jMWiX07Befbs7AAAAAAAAAAAAAAAAAPyzO34MnqoAKAAA"

	// filter types
	artifactTypeFilter = "artifactType"
)

var (
	testMap = map[string]int{
		envVarPull:              pull,
		envVarPush:              push,
		envVarContentDiscovery:  contentDiscovery,
		envVarContentManagement: contentManagement,
	}

	testBlobA                          []byte
	testBlobALength                    string
	testBlobADigest                    string
	testRefBlobA                       []byte
	testRefBlobALength                 string
	testRefBlobADigest                 string
	testRefArtifactTypeA               string
	testRefArtifactTypeB               string
	testRefArtifactTypeIndex           string
	testRefBlobB                       []byte
	testRefBlobBLength                 string
	testRefBlobBDigest                 string
	testBlobB                          []byte
	testBlobBDigest                    string
	testBlobBChunk1                    []byte
	testBlobBChunk1Length              string
	testBlobBChunk2                    []byte
	testBlobBChunk2Length              string
	testBlobBChunk1Range               string
	testBlobBChunk2Range               string
	testAnnotationKey                  string
	testAnnotationValues               map[string]string
	client                             *reggie.Client
	crossmountNamespace                string
	dummyDigest                        string
	errorCodes                         []string
	invalidManifestContent             []byte
	layerBlobData                      []byte
	layerBlobDigest                    string
	layerBlobContentLength             string
	emptyLayerManifestContent          []byte
	emptyLayerManifestDigest           string
	nonexistentManifest                string
	emptyJSONBlob                      []byte
	emptyJSONDescriptor                descriptor
	refsManifestAConfigArtifactContent []byte
	refsManifestAConfigArtifactDigest  string
	refsManifestALayerArtifactContent  []byte
	refsManifestALayerArtifactDigest   string
	refsManifestBConfigArtifactContent []byte
	refsManifestBConfigArtifactDigest  string
	refsManifestBLayerArtifactContent  []byte
	refsManifestBLayerArtifactDigest   string
	refsManifestCLayerArtifactContent  []byte
	refsManifestCLayerArtifactDigest   string
	refsIndexArtifactContent           []byte
	refsIndexArtifactDigest            string
	reportJUnitFilename                string
	reportHTMLFilename                 string
	httpWriter                         *httpDebugWriter
	testsToRun                         int
	suiteDescription                   string
	runPullSetup                       bool
	runPushSetup                       bool
	runContentDiscoverySetup           bool
	runContentManagementSetup          bool
	deleteManifestBeforeBlobs          bool
	runAutomaticCrossmountTest         bool
	automaticCrossmountEnabled         bool
	configs                            []TestBlob
	manifests                          []TestBlob
	seed                               int64
	Version                            = "unknown"
)

func init() {
	var err error

	seed = g.GinkgoRandomSeed()
	hostname := os.Getenv(envVarRootURL)
	namespace := os.Getenv(envVarNamespace)
	username := os.Getenv(envVarUsername)
	password := os.Getenv(envVarPassword)
	authScope := os.Getenv(envVarAuthScope)
	crossmountNamespace = os.Getenv(envVarCrossmountNamespace)
	if len(crossmountNamespace) == 0 {
		crossmountNamespace = fmt.Sprintf("conformance-%s", uuid.New())
	}

	debug, _ := strconv.ParseBool(os.Getenv(envVarDebug))

	for envVar, enableTest := range testMap {
		if varIsTrue, _ := strconv.ParseBool(os.Getenv(envVar)); varIsTrue {
			testsToRun |= enableTest
		}
	}

	httpWriter = newHTTPDebugWriter(debug)
	logger := newHTTPDebugLogger(httpWriter)
	client, err = reggie.NewClient(hostname,
		reggie.WithDefaultName(namespace),
		reggie.WithUsernamePassword(username, password),
		reggie.WithDebug(true),
		reggie.WithUserAgent("distribution-spec-conformance-tests"),
		reggie.WithAuthScope(authScope),
		reggie.WithInsecureSkipTLSVerify(true))
	if err != nil {
		panic(err)
	}

	client.SetLogger(logger)
	client.SetCookieJar(nil)

	// create a unique config for each workflow category
	for i := 0; i < numWorkflows; i++ {

		// in order to get a unique blob digest, we create a new author
		// field for the config on each run.
		randomAuthor := randomString(16)
		config := image{
			Architecture: "amd64",
			OS:           "linux",
			RootFS: rootFS{
				Type:    "layers",
				DiffIDs: []godigest.Digest{},
			},
			Author: randomAuthor,
		}
		configBlobContent, err := json.MarshalIndent(&config, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		configBlobContentLength := strconv.Itoa(len(configBlobContent))
		configBlobDigestRaw := godigest.FromBytes(configBlobContent)
		configBlobDigest := configBlobDigestRaw.String()
		if v := os.Getenv(envVarBlobDigest); v != "" {
			configBlobDigest = v
		}

		configs = append(configs, TestBlob{
			Content:       configBlobContent,
			ContentLength: configBlobContentLength,
			Digest:        configBlobDigest,
		})
	}

	layerBlobData, err = base64.StdEncoding.DecodeString(layerBase64String)
	if err != nil {
		log.Fatal(err)
	}

	layerBlobDigestRaw := godigest.FromBytes(layerBlobData)
	layerBlobDigest = layerBlobDigestRaw.String()
	layerBlobContentLength = fmt.Sprintf("%d", len(layerBlobData))

	layers := []descriptor{{
		MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
		Size:      int64(len(layerBlobData)),
		Digest:    layerBlobDigestRaw,
	}}

	// create a unique manifest for each workflow category
	for i := 0; i < numWorkflows; i++ {
		manifest := manifest{
			SchemaVersion: 2,
			MediaType:     "application/vnd.oci.image.manifest.v1+json",
			Config: descriptor{
				MediaType:           "application/vnd.oci.image.config.v1+json",
				Digest:              godigest.Digest(configs[i].Digest),
				Size:                int64(len(configs[i].Content)),
				Data:                configs[i].Content,    // must be the config content.
				NewUnspecifiedField: []byte("hello world"), // content doesn't matter.
			},
			Layers: layers,
		}

		manifestContent, err := json.MarshalIndent(&manifest, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		manifestContentLength := strconv.Itoa(len(manifestContent))
		manifestDigest := godigest.FromBytes(manifestContent).String()
		if v := os.Getenv(envVarManifestDigest); v != "" {
			manifestDigest = v
		}

		manifests = append(manifests, TestBlob{
			Content:       manifestContent,
			ContentLength: manifestContentLength,
			Digest:        manifestDigest,
		})
	}

	// used in push test
	emptyLayerManifest := manifest{
		SchemaVersion: 2,
		Config: descriptor{
			MediaType:           "application/vnd.oci.image.config.v1+json",
			Digest:              godigest.Digest(configs[1].Digest),
			Size:                int64(len(configs[1].Content)),
			Data:                configs[1].Content,    // must be the config content.
			NewUnspecifiedField: []byte("hello world"), // content doesn't matter.
		},
		Layers: []descriptor{},
	}

	emptyLayerManifestContent, err = json.MarshalIndent(&emptyLayerManifest, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	emptyLayerManifestDigest = string(godigest.FromBytes(emptyLayerManifestContent))

	nonexistentManifest = ".INVALID_MANIFEST_NAME"
	invalidManifestContent = []byte("blablabla")

	dig, blob := randomBlob(42, seed+1)
	testBlobA = blob
	testBlobALength = strconv.Itoa(len(testBlobA))
	testBlobADigest = dig.String()

	setupChunkedBlob(42)

	// used in referrers test (artifacts with Subject field set)
	emptyJSONBlob = []byte("{}")
	emptyJSONDescriptor = descriptor{
		MediaType: "application/vnd.oci.empty.v1+json",
		Size:      int64(len(emptyJSONBlob)),
		Digest:    godigest.FromBytes(emptyJSONBlob),
	}

	testRefBlobA = []byte("NHL Peanut Butter on my NHL bagel")
	testRefBlobALength = strconv.Itoa(len(testRefBlobA))
	testRefBlobADigest = godigest.FromBytes(testRefBlobA).String()

	testRefArtifactTypeA = "application/vnd.nhl.peanut.butter.bagel"

	testRefBlobB = []byte("NBA Strawberry Jam on my NBA croissant")
	testRefBlobBLength = strconv.Itoa(len(testRefBlobB))
	testRefBlobBDigest = godigest.FromBytes(testRefBlobB).String()

	testRefArtifactTypeB = "application/vnd.nba.strawberry.jam.croissant"

	testAnnotationKey = "org.opencontainers.conformance.test"
	testAnnotationValues = map[string]string{}

	// artifact with Subject ref using config.MediaType = artifactType
	refsManifestAConfigArtifact := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType: testRefArtifactTypeA,
			Size:      int64(len(testRefBlobA)),
			Digest:    godigest.FromBytes(testRefBlobA),
		},
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[4].Content)),
			Digest:    godigest.FromBytes(manifests[4].Content),
		},
		Layers: []descriptor{
			emptyJSONDescriptor,
		},
		Annotations: map[string]string{
			testAnnotationKey: "test config a",
		},
	}

	refsManifestAConfigArtifactContent, err = json.MarshalIndent(&refsManifestAConfigArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	refsManifestAConfigArtifactDigest = godigest.FromBytes(refsManifestAConfigArtifactContent).String()
	testAnnotationValues[refsManifestAConfigArtifactDigest] = refsManifestAConfigArtifact.Annotations[testAnnotationKey]

	refsManifestBConfigArtifact := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType: testRefArtifactTypeB,
			Size:      int64(len(testRefBlobB)),
			Digest:    godigest.FromBytes(testRefBlobB),
		},
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[4].Content)),
			Digest:    godigest.FromBytes(manifests[4].Content),
		},
		Layers: []descriptor{
			emptyJSONDescriptor,
		},
		Annotations: map[string]string{
			testAnnotationKey: "test config b",
		},
	}

	refsManifestBConfigArtifactContent, err = json.MarshalIndent(&refsManifestBConfigArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	refsManifestBConfigArtifactDigest = godigest.FromBytes(refsManifestBConfigArtifactContent).String()
	testAnnotationValues[refsManifestBConfigArtifactDigest] = refsManifestBConfigArtifact.Annotations[testAnnotationKey]

	// artifact with Subject ref using ArtifactType, config.MediaType = emptyJSON
	refsManifestALayerArtifact := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		ArtifactType:  testRefArtifactTypeA,
		Config:        emptyJSONDescriptor,
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[4].Content)),
			Digest:    godigest.FromBytes(manifests[4].Content),
		},
		Layers: []descriptor{
			{
				MediaType: testRefArtifactTypeA,
				Size:      int64(len(testRefBlobA)),
				Digest:    godigest.FromBytes(testRefBlobA),
			},
		},
		Annotations: map[string]string{
			testAnnotationKey: "test layer a",
		},
	}

	refsManifestALayerArtifactContent, err = json.MarshalIndent(&refsManifestALayerArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	refsManifestALayerArtifactDigest = godigest.FromBytes(refsManifestALayerArtifactContent).String()
	testAnnotationValues[refsManifestALayerArtifactDigest] = refsManifestALayerArtifact.Annotations[testAnnotationKey]

	refsManifestBLayerArtifact := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		ArtifactType:  testRefArtifactTypeB,
		Config:        emptyJSONDescriptor,
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[4].Content)),
			Digest:    godigest.FromBytes(manifests[4].Content),
		},
		Layers: []descriptor{
			{
				MediaType: testRefArtifactTypeB,
				Size:      int64(len(testRefBlobB)),
				Digest:    godigest.FromBytes(testRefBlobB),
			},
		},
		Annotations: map[string]string{
			testAnnotationKey: "test layer b",
		},
	}

	refsManifestBLayerArtifactContent, err = json.MarshalIndent(&refsManifestBLayerArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	refsManifestBLayerArtifactDigest = godigest.FromBytes(refsManifestBLayerArtifactContent).String()
	testAnnotationValues[refsManifestBLayerArtifactDigest] = refsManifestBLayerArtifact.Annotations[testAnnotationKey]

	// ManifestCLayerArtifact is the same as B but based on a subject that has not been pushed
	refsManifestCLayerArtifact := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		ArtifactType:  testRefArtifactTypeB,
		Config:        emptyJSONDescriptor,
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[3].Content)),
			Digest:    godigest.FromBytes(manifests[3].Content),
		},
		Layers: []descriptor{
			{
				MediaType: testRefArtifactTypeB,
				Size:      int64(len(testRefBlobB)),
				Digest:    godigest.FromBytes(testRefBlobB),
			},
		},
	}

	refsManifestCLayerArtifactContent, err = json.MarshalIndent(&refsManifestCLayerArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	refsManifestCLayerArtifactDigest = godigest.FromBytes(refsManifestCLayerArtifactContent).String()

	testRefArtifactTypeIndex = "application/vnd.food.stand"
	refsIndexArtifact := index{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		ArtifactType:  testRefArtifactTypeIndex,
		Manifests: []descriptor{
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Size:      int64(len(refsManifestAConfigArtifactContent)),
				Digest:    godigest.FromBytes(refsManifestAConfigArtifactContent),
			},
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Size:      int64(len(refsManifestALayerArtifactContent)),
				Digest:    godigest.FromBytes(refsManifestALayerArtifactContent),
			},
		},
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[4].Content)),
			Digest:    godigest.FromBytes(manifests[4].Content),
		},
		Annotations: map[string]string{
			testAnnotationKey: "test index",
		},
	}
	refsIndexArtifactContent, err = json.MarshalIndent(&refsIndexArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	refsIndexArtifactDigest = godigest.FromBytes(refsIndexArtifactContent).String()
	testAnnotationValues[refsIndexArtifactDigest] = refsIndexArtifact.Annotations[testAnnotationKey]

	dummyDigest = godigest.FromString("hello world").String()

	errorCodes = []string{
		BLOB_UNKNOWN:          "BLOB_UNKNOWN",
		BLOB_UPLOAD_INVALID:   "BLOB_UPLOAD_INVALID",
		BLOB_UPLOAD_UNKNOWN:   "BLOB_UPLOAD_UNKNOWN",
		DIGEST_INVALID:        "DIGEST_INVALID",
		MANIFEST_BLOB_UNKNOWN: "MANIFEST_BLOB_UNKNOWN",
		MANIFEST_INVALID:      "MANIFEST_INVALID",
		MANIFEST_UNKNOWN:      "MANIFEST_UNKNOWN",
		MANIFEST_UNVERIFIED:   "MANIFEST_UNVERIFIED",
		NAME_INVALID:          "NAME_INVALID",
		NAME_UNKNOWN:          "NAME_UNKNOWN",
		SIZE_INVALID:          "SIZE_INVALID",
		TAG_INVALID:           "TAG_INVALID",
		UNAUTHORIZED:          "UNAUTHORIZED",
		DENIED:                "DENIED",
		UNSUPPORTED:           "UNSUPPORTED",
	}

	runPullSetup = true
	runPushSetup = true
	runContentDiscoverySetup = true
	runContentManagementSetup = true
	deleteManifestBeforeBlobs = true

	if os.Getenv(envVarTagName) != "" &&
		os.Getenv(envVarManifestDigest) != "" &&
		os.Getenv(envVarBlobDigest) != "" {
		runPullSetup = false
	}

	if os.Getenv(envVarTagList) != "" {
		runContentDiscoverySetup = false
	}

	if v, ok := os.LookupEnv(envVarDeleteManifestBeforeBlobs); ok {
		deleteManifestBeforeBlobs, _ = strconv.ParseBool(v)
	}
	automaticCrossmountVal := ""
	automaticCrossmountVal, runAutomaticCrossmountTest = os.LookupEnv(envVarAutomaticCrossmount)
	automaticCrossmountEnabled, _ = strconv.ParseBool(automaticCrossmountVal)

	if dir := os.Getenv(envVarReportDir); dir != "none" {
		reportJUnitFilename = filepath.Join(dir, "junit.xml")
		reportHTMLFilename = filepath.Join(dir, "report.html")
	}
	suiteDescription = "OCI Distribution Conformance Tests"
}

func SkipIfDisabled(test int) {
	if userDisabled(test) {
		report := generateSkipReport()
		g.Skip(report)
	}
}

func RunOnlyIf(v bool) {
	if !v {
		g.Skip("you have skipped this test.")
	}
}

func RunOnlyIfNot(v bool) {
	if v {
		g.Skip("you have skipped this test.")
	}
}

func Warn(message string) {
	// print message
	fmt.Fprint(os.Stderr, formatter.Fi(2, "\n{{magenta}}WARNING: %s\n{{/}}", message))
	// print file:line
	_, file, line, _ := runtime.Caller(1)
	fmt.Fprint(os.Stderr, formatter.Fi(2, "\n%s:%d\n", file, line))
}

func generateSkipReport() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "you have skipped this test; if this is an error, check your environment variable settings:\n")
	for k := range testMap {
		fmt.Fprintf(buf, "\t%s=%s\n", k, os.Getenv(k))
	}
	return buf.String()
}

func userDisabled(test int) bool {
	return !(test&testsToRun > 0)
}

func getTagList(resp *reggie.Response) []string {
	jsonData := resp.Body()
	tagList := &TagList{}
	err := json.Unmarshal(jsonData, tagList)
	if err != nil {
		return []string{}
	}

	return tagList.Tags
}

// Adapted from https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func randomString(n int) string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret)
}

// randomBlob outputs a reproducible random blob (based on the seed) for testing
func randomBlob(size int, seed int64) (godigest.Digest, []byte) {
	r := mathrand.New(mathrand.NewSource(seed))
	b := make([]byte, size)
	if n, err := r.Read(b); err != nil {
		panic(err)
	} else if n != size {
		panic("unable to read enough bytes")
	}
	return godigest.FromBytes(b), b
}

func setupChunkedBlob(size int) {
	dig, blob := randomBlob(size, seed+2)
	testBlobB = blob
	testBlobBDigest = dig.String()
	testBlobBChunk1 = testBlobB[:size/2+1]
	testBlobBChunk1Length = strconv.Itoa(len(testBlobBChunk1))
	testBlobBChunk1Range = fmt.Sprintf("0-%d", len(testBlobBChunk1)-1)
	testBlobBChunk2 = testBlobB[size/2+1:]
	testBlobBChunk2Length = strconv.Itoa(len(testBlobBChunk2))
	testBlobBChunk2Range = fmt.Sprintf("%d-%d", len(testBlobBChunk1), len(testBlobB)-1)
}

type ManifestInfo struct {
	Index   bool
	Tag     string
	Digest  string
	Content []byte
	Subject string
}

func pushManifest(
	manifestInfo *ManifestInfo,
	manifestRefs []string,
	t g.GinkgoTInterface,
) []string {
	t.Helper()
	// if tag passed, use it as reference, otherwise use digest
	reference := manifestInfo.Tag
	if reference == "" {
		reference = manifestInfo.Digest
	}
	// if index, use corresponding mediaType
	mediaType := "application/vnd.oci.image.manifest.v1+json"
	if manifestInfo.Index {
		mediaType = "application/vnd.oci.image.index.v1+json"
	}
	// push manifest
	req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
		reggie.WithReference(reference)).
		SetHeader("Content-Type", mediaType).
		SetBody(manifestInfo.Content)
	resp, err := client.Do(req)
	// check all expectations
	Expect(err).To(BeNil())
	Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
	Expect(resp.Header().Get("Location")).ToNot(BeEmpty())
	// OCI-Subject is empty if referrers API unsupported
	Expect(resp.Header().Get("OCI-Subject")).To(SatisfyAny(
		Equal(manifestInfo.Subject),
		Equal("")))
	// keep track of all reference(s) to manifest
	manifestRefs = append(manifestRefs, reference)
	if manifestInfo.Tag != "" {
		manifestRefs = append(manifestRefs, manifestInfo.Digest)
	}
	return manifestRefs
}

func deleteManifest(
	ref string,
	t g.GinkgoTInterface,
) {
	t.Helper()

	req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>",
		reggie.WithReference(ref))
	resp, err := client.Do(req)
	Expect(err).To(BeNil())
	if resp.StatusCode() == http.StatusOK {
		req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<reference>",
			reggie.WithReference(ref))
		resp, err := client.Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode()).To(SatisfyAny(
			Equal(http.StatusAccepted),
			Equal(http.StatusMethodNotAllowed),
			Equal(http.StatusNotFound),
		))
		req = client.NewRequest(reggie.GET, "/v2/<name>/manifests/<ref>",
			reggie.WithReference(ref))
		resp, err = client.Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
	}
}

func deleteManifests(
	manifestRefs []string,
	t g.GinkgoTInterface,
) {
	t.Helper()

	for len(manifestRefs) > 0 {
		index := len(manifestRefs) - 1
		ref := manifestRefs[index]
		manifestRefs = manifestRefs[:index]
		deleteManifest(ref, t)
	}
}

type BlobInfo struct {
	Digest  string
	Content []byte
	Length  string
	Type    string
}

func pushBlob(
	blobInfo *BlobInfo,
	blobRefs []string,
	t g.GinkgoTInterface,
) []string {
	t.Helper()

	if blobInfo.Type == "" {
		blobInfo.Type = "application/octet-stream"
	}
	req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
	resp, err := client.Do(req)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode()).To(SatisfyAny(
		Equal(http.StatusAccepted),
		Equal(http.StatusCreated),
	))
	req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
		SetQueryParam("digest", blobInfo.Digest).
		SetHeader("Content-Type", blobInfo.Type).
		SetHeader("Content-Length", blobInfo.Length).
		SetBody(blobInfo.Content)
	resp, err = client.Do(req)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
	blobRefs = append(blobRefs, blobInfo.Digest)
	return blobRefs
}

func mountBlob(
	digest string,
	from string,
	blobRefs []string,
	t g.GinkgoTInterface,
) ([]string, *reggie.Response) {
	t.Helper()

	var req *reggie.Request

	if from == "" {
		req = client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/",
			reggie.WithName(crossmountNamespace)).
			SetQueryParam("mount", digest)
	} else {
		req = client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/",
			reggie.WithName(crossmountNamespace)).
			SetQueryParam("mount", digest).
			SetQueryParam("from", from)
	}
	resp, err := client.Do(req)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode()).To(SatisfyAny(
		Equal(http.StatusCreated),
		Equal(http.StatusAccepted),
	))
	Expect(resp.GetAbsoluteLocation()).To(Not(BeEmpty()))

	digest = fmt.Sprintf("%s/%s", crossmountNamespace, digest)
	blobRefs = append(blobRefs, digest)
	return blobRefs, resp
}

func deleteBlob(
	ref string,
	t g.GinkgoTInterface,
) int {
	t.Helper()

	var req *reggie.Request
	separator := "/"
	if strings.Contains(ref, separator) {
		// if blob not in default namespace, get it
		parts := strings.Split(ref, separator)
		namespace := parts[0]
		ref = parts[1]
		req = client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<reference>", reggie.WithName(namespace), reggie.WithReference(ref))
	} else {
		// blob in default namespace
		req = client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<reference>", reggie.WithReference(ref))
	}
	resp, err := client.Do(req)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode()).To(SatisfyAny(
		Equal(http.StatusAccepted),
		Equal(http.StatusNotFound),
		Equal(http.StatusBadRequest),
		Equal(http.StatusMethodNotAllowed),
	))
	return resp.StatusCode()
}

func deleteBlobs(
	blobRefs []string,
	t g.GinkgoTInterface,
) {
	t.Helper()

	for len(blobRefs) > 0 {
		index := len(blobRefs) - 1
		ref := blobRefs[index]
		blobRefs = blobRefs[:index]
		deleteBlob(ref, t)
	}
}
