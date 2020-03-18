package conformance

import (
	"bytes"
	"fmt"
	g "github.com/onsi/ginkgo"
	"log"
	"os"
	"strconv"

	"github.com/bloodorangeio/reggie"
	godigest "github.com/opencontainers/go-digest"
)

type (
	TagList struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
)

const (
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
)

const (
	envTrue                 = "1"
	envVarPush              = "OCI_TEST_PUSH"
	envVarDiscovery         = "OCI_TEST_DISCOVERY"
	envVarContentManagement = "OCI_TEST_CONTENT_MANAGEMENT"
	envVarBlobDigest        = "OCI_BLOB_DIGEST"
	envVarManifestDigest    = "OCI_MANIFEST_DIGEST"
	envVarTagName           = "OCI_TAG_NAME"
	push                    = 1 << iota
	discovery
	contentManagement
)

var (
	testMap = map[string]int{
		envVarPush:              push,
		envVarDiscovery:         discovery,
		envVarContentManagement: contentManagement,
	}

	reqVarsForDisabledFlow = map[string][]string{
		envVarPush: {envVarBlobDigest, envVarManifestDigest, envVarTagName},
		envVarDiscovery: {envVarTagName},
		envVarContentManagement: {envVarManifestDigest, envVarTagName, envVarBlobDigest},
	}
)

var (
	blobA                  []byte
	blobALength            string
	blobADigest            string
	blobB                  []byte
	blobBDigest            string
	blobBChunk1            []byte
	blobBChunk1Length      string
	blobBChunk2            []byte
	blobBChunk2Length      string
	blobBChunk1Range       string
	blobBChunk2Range       string
	client                 *reggie.Client
	configContent          []byte
	configContentLength    string
	blobDigest             string
	dummyDigest            string
	errorCodes             []string
	firstTag               string
	lastResponse           *reggie.Response
	lastTagList            TagList
	manifestContent        []byte
	invalidManifestContent []byte
	manifestDigest         string
	nonexistentManifest    string
	numTags                int
	reportJUnitFilename    string
	reportHTMLFilename     string
	httpWriter             *httpDebugWriter
	testsToRun             int
	suiteDescription       string
	Version                = "unknown"
)

func init() {
	hostname := os.Getenv("OCI_ROOT_URL")
	namespace := os.Getenv("OCI_NAMESPACE")
	username := os.Getenv("OCI_USERNAME")
	password := os.Getenv("OCI_PASSWORD")
	debug := os.Getenv("OCI_DEBUG") == "true"

	for envVar, enableTest := range testMap {
		if os.Getenv(envVar) == envTrue {
			testsToRun |= enableTest
		}
	}
	ValidateRequiredEnvVars()

	var err error

	httpWriter = newHTTPDebugWriter(debug)
	logger := newHTTPDebugLogger(httpWriter)
	client, err = reggie.NewClient(hostname,
		reggie.WithDefaultName(namespace),
		reggie.WithUsernamePassword(username, password),
		reggie.WithDebug(true),
		reggie.WithUserAgent("distribution-spec-conformance-tests"))
	if err != nil {
		panic(err)
	}

	client.SetLogger(logger)

	configContent = []byte("{}\n")
	configContentLength = strconv.Itoa(len(configContent))
	blobDigest = godigest.FromBytes(configContent).String()
	if v := os.Getenv(envVarBlobDigest); v != "" {
		blobDigest = v
	}

	manifestContent = []byte(fmt.Sprintf(
		"{ \"mediaType\": \"application/vnd.oci.image.manifest.v1+json\", \"config\":  { \"digest\": \"%s\", "+
			"\"mediaType\": \"application/vnd.oci.image.config.v1+json\","+" \"size\": %s }, \"layers\": [], "+
			"\"schemaVersion\": 2 }",
		blobDigest, configContentLength))
	manifestDigest = godigest.FromBytes(manifestContent).String()
	if v := os.Getenv(envVarManifestDigest); v != "" {
		manifestDigest = v
	}
	nonexistentManifest = ".INVALID_MANIFEST_NAME"
	invalidManifestContent = []byte("blablabla")

	blobA = []byte("NBA Jam on my NBA toast")
	blobALength = strconv.Itoa(len(blobA))
	blobADigest = godigest.FromBytes(blobA).String()

	blobB = []byte("Hello, how are you today?")
	blobBDigest = godigest.FromBytes(blobB).String()
	blobBChunk1 = blobB[:3]
	blobBChunk1Length = strconv.Itoa(len(blobBChunk1))
	blobBChunk1Range = fmt.Sprintf("0-%d", len(blobBChunk1)-1)
	blobBChunk2 = blobB[3:]
	blobBChunk2Length = strconv.Itoa(len(blobBChunk2))
	blobBChunk2Range = fmt.Sprintf("%d-%d", len(blobBChunk1), len(blobB)-1)

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

	reportJUnitFilename = "junit.xml"
	reportHTMLFilename = "report.html"
	suiteDescription = "OCI Distribution Conformance Tests"
}

func SkipIfDisabled(test int) {
	report := generateSkipReport()
	if userDisabled(test) {
		g.Skip(report)
	}
}

func ValidateRequiredEnvVars() {
	buf := new(bytes.Buffer)
	var validationFailed = false
	for envVar, test := range testMap {
		if userDisabled(test) {
			report, ok := checkRequiredVars(envVar)
			if !ok {
				validationFailed = true
				fmt.Fprintf(buf, report)
			}
		}
	}

	if validationFailed {
		log.Fatal(buf.String())
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
	return !(test & testsToRun > 0)
}

func checkRequiredVars(mainVarToValidate string) (string, bool) {
	buf := new(bytes.Buffer)
	var allSupplied = true
	fmt.Fprintf(buf, "\ndisabling %s requires all of the following environment variables to be set. " +
		"here is your current configuration:\n", mainVarToValidate)
	for _, subVarForDisabledFlow := range reqVarsForDisabledFlow[mainVarToValidate] {
		yesNo := "✓"
		if os.Getenv(subVarForDisabledFlow) == "" {
			allSupplied = false
			yesNo = "✘"
		}
		fmt.Fprintf(buf, "\t%s %s=%s\n", yesNo, subVarForDisabledFlow, os.Getenv(subVarForDisabledFlow))
	}
	return buf.String(), allSupplied
}
