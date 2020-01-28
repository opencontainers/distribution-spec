package conformance

import (
	"fmt"
	"os"
	"strconv"

	"github.com/bloodorangeio/reggie"
	godigest "github.com/opencontainers/go-digest"
)

// TODO: import from opencontainers/distribution-spec
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
	configDigest           string
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
	suiteDescription       string
	Version                = "unknown"
)

func init() {
	hostname := os.Getenv("OCI_ROOT_URL")
	namespace := os.Getenv("OCI_NAMESPACE")
	username := os.Getenv("OCI_USERNAME")
	password := os.Getenv("OCI_PASSWORD")
	debug := os.Getenv("OCI_DEBUG") == "true"

	var err error

	httpWriter = newHTTPDebugWriter(debug)
	logger := newHTTPDebugLogger(httpWriter)
	client, err = reggie.NewClient(hostname,
		reggie.WithDefaultName(namespace),
		reggie.WithUsernamePassword(username, password),
		reggie.WithDebug(true),
		reggie.WithUserAgent("distribution-spec-conformance-tests"))
	client.SetLogger(logger)
	if err != nil {
		panic(err)
	}

	configContent = []byte("{}\n")
	configContentLength = strconv.Itoa(len(configContent))
	configDigest = godigest.FromBytes(configContent).String()

	manifestContent = []byte(fmt.Sprintf(
		"{ \"mediaType\": \"application/vnd.oci.image.manifest.v1+json\", \"config\":  { \"digest\": \"%s\", "+
			"\"mediaType\": \"application/vnd.oci.image.config.v1+json\","+" \"size\": %s }, \"layers\": [], "+
			"\"schemaVersion\": 2 }",
		configDigest, configContentLength))
	manifestDigest = godigest.FromBytes(manifestContent).String()
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
