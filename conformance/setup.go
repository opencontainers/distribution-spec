package conformance

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/bloodorangeio/reggie"
	"github.com/google/uuid"
	g "github.com/onsi/ginkgo"
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
)

var (
	testMap = map[string]int{
		envVarPull:              pull,
		envVarPush:              push,
		envVarContentDiscovery:  contentDiscovery,
		envVarContentManagement: contentManagement,
	}

	testBlobA                  []byte
	testBlobALength            string
	testBlobADigest            string
	testBlobB                  []byte
	testBlobBDigest            string
	testBlobBChunk1            []byte
	testBlobBChunk1Length      string
	testBlobBChunk2            []byte
	testBlobBChunk2Length      string
	testBlobBChunk1Range       string
	testBlobBChunk2Range       string
	client                     *reggie.Client
	crossmountNamespace        string
	dummyDigest                string
	errorCodes                 []string
	invalidManifestContent     []byte
	layerBlobData              []byte
	layerBlobDigest            string
	layerBlobContentLength     string
	emptyLayerManifestContent  []byte
	nonexistentManifest        string
	reportJUnitFilename        string
	reportHTMLFilename         string
	httpWriter                 *httpDebugWriter
	testsToRun                 int
	suiteDescription           string
	runPullSetup               bool
	runPushSetup               bool
	runContentDiscoverySetup   bool
	runContentManagementSetup  bool
	skipEmptyLayerTest         bool
	deleteManifestBeforeBlobs  bool
	runAutomaticCrossmountTest bool
	automaticCrossmountEnabled bool
	configs                    []TestBlob
	manifests                  []TestBlob
	Version                    = "unknown"
)

func init() {
	var err error

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
		reggie.WithAuthScope(authScope))
	if err != nil {
		panic(err)
	}

	client.SetLogger(logger)
	client.SetCookieJar(nil)

	// create a unique config for each workflow category
	for i := 0; i < 4; i++ {

		// in order to get a unique blob digest, we create a new author
		// field for the config on each run.
		randomAuthor := randomString(16)
		config := Image{
			Architecture: "amd64",
			OS:           "linux",
			RootFS: RootFS{
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

	layers := []Descriptor{{
		MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
		Size:      int64(len(layerBlobData)),
		Digest:    layerBlobDigestRaw,
	}}

	// create a unique manifest for each workflow category
	for i := 0; i < 4; i++ {
		manifest := Manifest{
			Config: Descriptor{
				MediaType:           "application/vnd.oci.image.config.v1+json",
				Digest:              godigest.Digest(configs[i].Digest),
				Size:                int64(len(configs[i].Content)),
				NewUnspecifiedField: configs[i].Content,
			},
			Layers: layers,
		}
		manifest.SchemaVersion = 2

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
	emptyLayerManifest := Manifest{
		Config: Descriptor{
			MediaType:           "application/vnd.oci.image.config.v1+json",
			Digest:              godigest.Digest(configs[1].Digest),
			Size:                int64(len(configs[1].Content)),
			NewUnspecifiedField: configs[1].Content,
		},
		Layers: []Descriptor{},
	}
	emptyLayerManifest.SchemaVersion = 2

	emptyLayerManifestContent, err = json.MarshalIndent(&emptyLayerManifest, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	nonexistentManifest = ".INVALID_MANIFEST_NAME"
	invalidManifestContent = []byte("blablabla")

	testBlobA = []byte("NBA Jam on my NBA toast")
	testBlobALength = strconv.Itoa(len(testBlobA))
	testBlobADigest = godigest.FromBytes(testBlobA).String()

	testBlobB = []byte("Hello, how are you today?")
	testBlobBDigest = godigest.FromBytes(testBlobB).String()
	testBlobBChunk1 = testBlobB[:3]
	testBlobBChunk1Length = strconv.Itoa(len(testBlobBChunk1))
	testBlobBChunk1Range = fmt.Sprintf("0-%d", len(testBlobBChunk1)-1)
	testBlobBChunk2 = testBlobB[3:]
	testBlobBChunk2Length = strconv.Itoa(len(testBlobBChunk2))
	testBlobBChunk2Range = fmt.Sprintf("%d-%d", len(testBlobBChunk1), len(testBlobB)-1)

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
	skipEmptyLayerTest = false
	deleteManifestBeforeBlobs = false

	if os.Getenv(envVarTagName) != "" &&
		os.Getenv(envVarManifestDigest) != "" &&
		os.Getenv(envVarBlobDigest) != "" {
		runPullSetup = false
	}

	if os.Getenv(envVarTagList) != "" {
		runContentDiscoverySetup = false
	}

	skipEmptyLayerTest, _ = strconv.ParseBool(os.Getenv(envVarPushEmptyLayer))
	deleteManifestBeforeBlobs, _ = strconv.ParseBool(os.Getenv(envVarDeleteManifestBeforeBlobs))
	automaticCrossmountVal := ""
	automaticCrossmountVal, runAutomaticCrossmountTest = os.LookupEnv(envVarAutomaticCrossmount)
	automaticCrossmountEnabled, _ = strconv.ParseBool(automaticCrossmountVal)

	reportJUnitFilename = "junit.xml"
	reportHTMLFilename = "report.html"
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

func getTagNameFromResponse(lastResponse *reggie.Response) (tagName string) {
	tl := &TagList{}
	if lastResponse != nil {
		jsonData := lastResponse.Body()
		err := json.Unmarshal(jsonData, tl)
		if err == nil && len(tl.Tags) > 0 {
			tagName = tl.Tags[0]
		}
	}

	return
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
