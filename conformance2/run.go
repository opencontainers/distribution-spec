package main

import (
	"crypto/rand"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	digest "github.com/opencontainers/go-digest"
	image "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	testName = "OCI Conformance Test"
)

var (
	dataTests             = []string{}
	dataFailManifestTests = []struct {
		tdName string
		opts   []apiDoOpt
	}{}
)

type runner struct {
	Config  config
	API     *api
	State   *state
	Results *results
	Log     *slog.Logger
}

func runnerNew(c config) (*runner, error) {
	lvl := slog.LevelWarn
	if c.LogLevel != "" {
		err := lvl.UnmarshalText([]byte(c.LogLevel))
		if err != nil {
			return nil, fmt.Errorf("failed to parse logging level %s: %w", c.LogLevel, err)
		}
	}
	if c.LogWriter == nil {
		c.LogWriter = os.Stderr
	}
	apiOpts := []apiOpt{}
	if c.LoginUser != "" && c.LoginPass != "" {
		apiOpts = append(apiOpts, apiWithAuth(c.LoginUser, c.LoginPass))
	}
	r := runner{
		Config:  c,
		API:     apiNew(http.DefaultClient, apiOpts...),
		State:   stateNew(),
		Results: resultsNew(testName, nil),
		Log:     slog.New(slog.NewTextHandler(c.LogWriter, &slog.HandlerOptions{Level: lvl})),
	}
	return &r, nil
}

func (r *runner) GenerateData() error {
	var tdName string
	if !r.Config.Data.Image {
		// all data tests require the image manifest
		return nil
	}
	// include empty tests for user provided read-only data, no validation is done on the content of the response since we don't know it
	if len(r.Config.ROData.Tags) > 0 || len(r.Config.ROData.Manifests) > 0 || len(r.Config.ROData.Blobs) > 0 || len(r.Config.ROData.Referrers) > 0 {
		tdName = "read-only"
		r.State.Data[tdName] = newTestData("Read Only Inputs")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		for _, tag := range r.Config.ROData.Tags {
			r.State.Data[tdName].tags[tag] = ""
		}
		for _, manifest := range r.Config.ROData.Manifests {
			dig, err := digest.Parse(manifest)
			if err != nil {
				return fmt.Errorf("failed to parse manifest digest %s: %w", manifest, err)
			}
			r.State.Data[tdName].manifests[dig] = []byte{}
			r.State.Data[tdName].manOrder = append(r.State.Data[tdName].manOrder, dig)
		}
		for _, blob := range r.Config.ROData.Blobs {
			dig, err := digest.Parse(blob)
			if err != nil {
				return fmt.Errorf("failed to parse blob digest %s: %w", blob, err)
			}
			r.State.Data[tdName].blobs[dig] = []byte{}
		}
		for _, subject := range r.Config.ROData.Referrers {
			dig, err := digest.Parse(subject)
			if err != nil {
				return fmt.Errorf("failed to parse subject digest %s: %w", subject, err)
			}
			r.State.Data[tdName].referrers[dig] = []*image.Descriptor{}
		}
	}
	if !r.Config.APIs.Push {
		// do not generate random data if push is disabled
		return nil
	}
	// standard image with a layer per blob test
	tdName = "image"
	r.State.Data[tdName] = newTestData("Image")
	r.State.DataStatus[tdName] = statusUnknown
	dataTests = append(dataTests, tdName)
	_, err := r.State.Data[tdName].genManifestFull(
		genWithTag("image"),
	)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	tdName = "image-uncompressed"
	r.State.Data[tdName] = newTestData("Image Uncompressed")
	r.State.DataStatus[tdName] = statusUnknown
	dataTests = append(dataTests, tdName)
	_, err = r.State.Data[tdName].genManifestFull(
		genWithTag("image-uncompressed"),
		genWithCompress(genCompUncomp),
	)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	// multi-platform index
	if r.Config.Data.Index {
		tdName = "index"
		r.State.Data[tdName] = newTestData("Index")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		_, err = r.State.Data[tdName].genIndexFull(
			genWithTag("index"),
			genWithPlatforms([]*image.Platform{
				{OS: "linux", Architecture: "amd64"},
				{OS: "linux", Architecture: "arm64"},
			}),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// index containing an index
	if r.Config.Data.Index && r.Config.Data.IndexList {
		tdName = "nested-index"
		r.State.Data[tdName] = newTestData("Nested Index")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		dig1, err := r.State.Data[tdName].genIndexFull(
			genWithPlatforms([]*image.Platform{
				{OS: "linux", Architecture: "amd64"},
				{OS: "linux", Architecture: "arm64"},
			}),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		dig2, err := r.State.Data[tdName].genIndexFull(
			genWithPlatforms([]*image.Platform{
				{OS: "linux", Architecture: "amd64"},
				{OS: "linux", Architecture: "arm64"},
			}),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		_, _, err = r.State.Data[tdName].genIndex([]*image.Platform{nil, nil}, []digest.Digest{dig1, dig2},
			genWithTag("index-of-index"),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// empty index
	if r.Config.Data.Index {
		tdName = "empty-index"
		r.State.Data[tdName] = newTestData("Empty Index")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		_, err = r.State.Data[tdName].genIndexFull(
			genWithTag("index"),
			genWithPlatforms([]*image.Platform{}),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// artifact manifest
	if r.Config.Data.Artifact {
		tdName = "artifact"
		r.State.Data[tdName] = newTestData("Artifact")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		_, err = r.State.Data[tdName].genManifestFull(
			genWithTag("artifact"),
			genWithArtifactType(mtExampleConf),
			genWithConfigMediaType(mtOCIEmptyJSON),
			genWithConfigBytes([]byte("{}")),
			genWithLayerCount(1),
			genWithLayerMediaType(mtExampleConf),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// artifact index
	if r.Config.Data.ArtifactList {
		tdName = "artifact-index"
		r.State.Data[tdName] = newTestData("Artifact Index")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		_, err = r.State.Data[tdName].genIndexFull(
			genWithTag("artifact-index"),
			genWithPlatforms([]*image.Platform{
				{OS: "linux", Architecture: "amd64"},
				{OS: "linux", Architecture: "arm64"},
			}),
			genWithArtifactType(mtExampleConf),
			genWithConfigMediaType(mtOCIEmptyJSON),
			genWithConfigBytes([]byte("{}")),
			genWithLayerCount(1),
			genWithLayerMediaType(mtExampleConf),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// artifact without layers
	if r.Config.Data.Artifact {
		tdName = "artifact-without-layers"
		r.State.Data[tdName] = newTestData("Artifact without Layers")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		_, err = r.State.Data[tdName].genManifestFull(
			genWithTag("artifact-without-layers"),
			genWithArtifactType(mtExampleConf),
			genWithConfigMediaType(mtOCIEmptyJSON),
			genWithConfigBytes([]byte("{}")),
			genWithLayerCount(1),
			genWithLayerBytes([]byte("{}")),
			genWithLayerMediaType(mtOCIEmptyJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// image and two referrers
	if r.Config.Data.Subject {
		tdName = "artifacts-with-subject"
		r.State.Data[tdName] = newTestData("Artifacts with Subject")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		subjDig, err := r.State.Data[tdName].genManifestFull(
			genWithTag("image"),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		subjDesc := *r.State.Data[tdName].desc[subjDig]
		_, err = r.State.Data[tdName].genManifestFull(
			genWithSubject(subjDesc),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		_, err = r.State.Data[tdName].genManifestFull(
			genWithArtifactType(mtExampleConf),
			genWithAnnotations(map[string]string{
				"org.opencontainers.conformance": "hello conformance test",
			}),
			genWithAnnotationUniq(),
			genWithConfigMediaType(mtOCIEmptyJSON),
			genWithConfigBytes([]byte("{}")),
			genWithLayerCount(1),
			genWithLayerMediaType(mtExampleConf),
			genWithSubject(subjDesc),
			genWithTag("tagged-artifact"),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// index and artifact-index with a subject
	if r.Config.Data.SubjectList {
		tdName = "index-with-subject"
		r.State.Data[tdName] = newTestData("Index with Subject")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		subjDig, err := r.State.Data[tdName].genIndexFull(
			genWithTag("index"),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		subjDesc := *r.State.Data[tdName].desc[subjDig]
		_, err = r.State.Data[tdName].genIndexFull(
			genWithArtifactType(mtExampleConf),
			genWithAnnotations(map[string]string{
				"org.opencontainers.conformance": "hello conformance test",
			}),
			genWithAnnotationUniq(),
			genWithConfigMediaType(mtOCIEmptyJSON),
			genWithConfigBytes([]byte("{}")),
			genWithLayerCount(1),
			genWithLayerMediaType(mtExampleConf),
			genWithSubject(subjDesc),
			genWithTag("tagged-artifact"),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// artifact with missing subject
	if r.Config.Data.SubjectMissing {
		tdName = "missing-subject"
		r.State.Data[tdName] = newTestData("Missing Subject")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		subjDesc := image.Descriptor{
			MediaType: mtOCIImage,
			Size:      123,
			Digest:    digest.FromString("missing content"),
		}
		_, err = r.State.Data[tdName].genManifestFull(
			genWithArtifactType(mtExampleConf),
			genWithAnnotations(map[string]string{
				"org.opencontainers.conformance": "hello conformance test",
			}),
			genWithAnnotationUniq(),
			genWithConfigMediaType(mtOCIEmptyJSON),
			genWithConfigBytes([]byte("{}")),
			genWithLayerCount(1),
			genWithLayerMediaType(mtExampleConf),
			genWithSubject(subjDesc),
			genWithTag("tagged-artifact"),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// data field in descriptor
	if r.Config.Data.DataField {
		tdName = "data-field"
		r.State.Data[tdName] = newTestData("Data Field")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		_, err := r.State.Data[tdName].genManifestFull(
			genWithTag("data-field"),
			genWithDescriptorData(),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// image with non-distributable layers
	if r.Config.Data.Nondistributable {
		tdName = "non-distributable-layers"
		r.State.Data[tdName] = newTestData("Non-distributable Layers")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)

		b := make([]byte, 256)
		layers := make([]image.Descriptor, 3)
		confDig := make([]digest.Digest, 3)
		// first layer is compressed + non-distributable
		_, err := rand.Read(b)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		confDig[0] = digest.Canonical.FromBytes(b)
		_, err = rand.Read(b)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		dig := digest.Canonical.FromBytes(b)
		layers[0] = image.Descriptor{
			MediaType: mtOCILayerNdGz,
			Digest:    dig,
			Size:      123456,
			URLs:      []string{"https://store.example.com/blobs/sha256/" + dig.Encoded()},
		}
		// second layer is uncompressed + non-distributable
		_, err = rand.Read(b)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		dig = digest.Canonical.FromBytes(b)
		confDig[1] = dig
		layers[1] = image.Descriptor{
			MediaType: mtOCILayerNd,
			Digest:    dig,
			Size:      12345,
			URLs:      []string{"https://store.example.com/blobs/sha256/" + dig.Encoded()},
		}
		// third layer is normal
		cDig, ucDig, _, err := r.State.Data[tdName].genLayer(1)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		confDig[2] = ucDig
		layers[2] = *r.State.Data[tdName].desc[cDig]
		// generate the config
		cDig, _, err = r.State.Data[tdName].genConfig(image.Platform{OS: "linux", Architecture: "amd64"}, confDig)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		// generate the manifest
		_, _, err = r.State.Data[tdName].genManifest(*r.State.Data[tdName].desc[cDig], layers,
			genWithTag("non-distributable-image"),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// add a randomized unknown field to manifests and config
	if r.Config.Data.CustomFields {
		tdName = "custom-fields"
		r.State.Data[tdName] = newTestData("Custom Fields")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		_, err = r.State.Data[tdName].genIndexFull(
			genWithTag("custom-fields"),
			genWithPlatforms([]*image.Platform{
				{OS: "linux", Architecture: "amd64"},
				{OS: "linux", Architecture: "arm64"},
			}),
			genWithExtraField(),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	// sparse manifests missing layers/platforms
	if r.Config.Data.Sparse {
		tdName = "sparse"
		r.State.Data[tdName] = newTestData("Sparse Manifests")
		r.State.DataStatus[tdName] = statusUnknown
		dataTests = append(dataTests, tdName)
		_, err := r.State.Data[tdName].genManifestFull(
			genWithTag("sparse-image"),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		for dig := range r.State.Data[tdName].blobs {
			if strings.HasPrefix(r.State.Data[tdName].desc[dig].MediaType, mtOCILayerPre) {
				// remove the first layer we find
				delete(r.State.Data[tdName].desc, dig)
				delete(r.State.Data[tdName].blobs, dig)
				break
			}
		}
		// for the index, make an image and a random digest/descriptor, add both to an index
		imagePlat := image.Platform{
			OS:           "linux",
			Architecture: "amd64",
		}
		imageDig, err := r.State.Data[tdName].genManifestFull(
			genWithPlatform(imagePlat),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		randPlat := image.Platform{
			OS:           "linux",
			Architecture: "arm64",
		}
		b := make([]byte, 1024)
		_, err = rand.Read(b)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		randDig := digest.Canonical.FromBytes(b)
		r.State.Data[tdName].desc[randDig] = &image.Descriptor{
			MediaType: mtOCIImage,
			Digest:    randDig,
			Size:      1024,
		}
		_, _, err = r.State.Data[tdName].genIndex(
			[]*image.Platform{&imagePlat, &randPlat},
			[]digest.Digest{imageDig, randDig},
			genWithTag("sparse-index"),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
	}
	tdName = "bad-digest-image"
	r.State.Data[tdName] = newTestData("Bad Digest Image")
	r.State.DataStatus[tdName] = statusUnknown
	dataFailManifestTests = append(dataFailManifestTests, struct {
		tdName string
		opts   []apiDoOpt
	}{tdName: tdName, opts: []apiDoOpt{apiWithFlag("ExpectBadDigest")}})
	dig, err := r.State.Data[tdName].genManifestFull()
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	// add some whitespace to make the digest mismatch
	r.State.Data[tdName].manifests[dig] = append(r.State.Data[tdName].manifests[dig], []byte("  ")...)

	// TODO: sha512 digest

	return nil
}

func (r *runner) Report(w io.Writer) {
	fmt.Fprintf(w, "Test results\n")
	r.Results.ReportWalkErr(w, "")
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "OCI Conformance Result: %s\n", r.Results.Status.String())
	padWidth := 30

	statusTotal := 0
	for i := status(1); i < statusMax; i++ {
		pad := ""
		if len(i.String()) < padWidth {
			pad = strings.Repeat(".", padWidth-len(i.String()))
		}
		fmt.Fprintf(w, "  %s%s: %10d\n", i.String(), pad, r.Results.Counts[i])
		statusTotal += r.Results.Counts[i]
	}
	pad := strings.Repeat(".", padWidth-len("Total"))
	fmt.Fprintf(w, "  %s%s: %10d\n\n", "Total", pad, statusTotal)

	if len(r.Results.Errs) > 0 {
		fmt.Fprintf(w, "Errors:\n%s\n\n", errors.Join(r.Results.Errs...))
	}

	fmt.Fprintf(w, "API conformance:\n")
	for i := range stateAPIMax {
		pad := ""
		if len(i.String()) < padWidth {
			pad = strings.Repeat(".", padWidth-len(i.String()))
		}
		fmt.Fprintf(w, "  %s%s: %10s\n", i.String(), pad, r.State.APIStatus[i].String())
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "Data conformance:\n")
	tdNames := []string{}
	for tdName := range r.State.Data {
		tdNames = append(tdNames, tdName)
	}
	sort.Strings(tdNames)
	for _, tdName := range tdNames {
		pad := ""
		if len(r.State.Data[tdName].name) < padWidth {
			pad = strings.Repeat(".", padWidth-len(r.State.Data[tdName].name))
		}
		fmt.Fprintf(w, "  %s%s: %10s\n", r.State.Data[tdName].name, pad, r.State.DataStatus[tdName].String())
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "Configuration:\n")
	fmt.Fprintf(w, "  %s", strings.ReplaceAll(r.Config.Report(), "\n", "\n  "))
	fmt.Fprintf(w, "\n")
}

func (r *runner) ReportJunit(w io.Writer) error {
	ju := r.toJunit()
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(ju)
}

func (r *runner) toJunit() *junitTestSuites {
	statusTotal := 0
	for i := status(1); i < statusMax; i++ {
		statusTotal += r.Results.Counts[i]
	}
	tSec := fmt.Sprintf("%f", r.Results.Stop.Sub(r.Results.Start).Seconds())
	jTSuites := junitTestSuites{
		Tests:    statusTotal,
		Errors:   r.Results.Counts[statusError],
		Failures: r.Results.Counts[statusFail],
		Skipped:  r.Results.Counts[statusSkip],
		Disabled: r.Results.Counts[statusDisabled],
		Time:     tSec,
	}
	jTSuite := junitTestSuite{
		Name:      r.Results.Name,
		Tests:     statusTotal,
		Errors:    r.Results.Counts[statusError],
		Failures:  r.Results.Counts[statusFail],
		Skipped:   r.Results.Counts[statusSkip],
		Disabled:  r.Results.Counts[statusDisabled],
		Time:      tSec,
		Testcases: r.Results.ToJunitTestCases(),
	}
	jTSuite.Properties = []junitProperty{{Name: "Config", Value: r.Config.Report()}}
	jTSuites.Suites = []junitTestSuite{jTSuite}
	return &jTSuites
}

type reportData struct {
	Config          config
	Results         *results
	NumTotal        int
	NumPassed       int
	NumFailed       int
	NumSkipped      int
	PercentPassed   int
	PercentFailed   int
	PercentSkipped  int
	StartTimeString string
	EndTimeString   string
	RunTime         string
	AllPassed       bool
	AllFailed       bool
	AllSkipped      bool
	Version         string
}

func (r *runner) ReportHTML(w io.Writer) error {
	data := reportData{
		Config:          r.Config,
		Results:         r.Results,
		NumTotal:        r.Results.Counts[statusPass] + r.Results.Counts[statusFail] + r.Results.Counts[statusError] + r.Results.Counts[statusSkip] + r.Results.Counts[statusDisabled],
		NumPassed:       r.Results.Counts[statusPass],
		NumFailed:       r.Results.Counts[statusFail] + r.Results.Counts[statusError],
		NumSkipped:      r.Results.Counts[statusSkip] + r.Results.Counts[statusDisabled],
		StartTimeString: r.Results.Start.Format("Jan 2 15:04:05.000 -0700 MST"),
		EndTimeString:   r.Results.Stop.Format("Jan 2 15:04:05.000 -0700 MST"),
		RunTime:         r.Results.Stop.Sub(r.Results.Start).String(),
	}
	data.PercentPassed = int(math.Round(float64(data.NumPassed) / float64(data.NumTotal) * 100))
	data.PercentFailed = int(math.Round(float64(data.NumFailed) / float64(data.NumTotal) * 100))
	data.PercentSkipped = int(math.Round(float64(data.NumSkipped) / float64(data.NumTotal) * 100))
	data.AllPassed = data.NumPassed == data.NumTotal
	data.AllFailed = data.NumFailed == data.NumTotal
	data.AllSkipped = data.NumSkipped == data.NumTotal
	data.Version = r.Config.Version
	// load all templates
	t := template.New("report")
	for name, value := range confHTMLTemplates {
		tAdd, err := template.New(name).Parse(value)
		if err != nil {
			return fmt.Errorf("cannot parse report template %s: %v", name, err)
		}
		t, err = t.AddParseTree(name, tAdd.Tree)
		if err != nil {
			return fmt.Errorf("cannot add report template %s to tree: %v", name, err)
		}
	}
	// execute the top level report template
	return t.ExecuteTemplate(w, "report", data)
}

func (r *runner) TestAll() error {
	errs := []error{}
	r.Results.Start = time.Now()
	repo := r.Config.Repo1
	repo2 := r.Config.Repo2

	err := r.GenerateData()
	if err != nil {
		return fmt.Errorf("aborting tests, unable to generate data: %w", err)
	}

	err = r.TestEmpty(r.Results, repo)
	if err != nil {
		errs = append(errs, err)
	}

	algos := []digest.Algorithm{digest.SHA256}
	if r.Config.Data.Sha512 {
		algos = append(algos, digest.SHA512)
	}
	for _, algo := range algos {
		err = r.TestBlobAPIs(r.Results, "blobs-"+algo.String(), "Blobs "+algo.String(), algo, repo, repo2)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// loop over different types of data
	for _, tdName := range dataTests {
		err = r.ChildRun(tdName, r.Results, func(r *runner, res *results) error {
			errs := []error{}
			// push
			err := r.TestPush(res, tdName, repo)
			if err != nil {
				errs = append(errs, err)
			}
			// list, pull, and query
			err = r.TestList(res, tdName, repo)
			if err != nil {
				errs = append(errs, err)
			}
			err = r.TestHead(res, tdName, repo)
			if err != nil {
				errs = append(errs, err)
			}
			err = r.TestPull(res, tdName, repo)
			if err != nil {
				errs = append(errs, err)
			}
			err = r.TestReferrers(res, tdName, repo)
			if err != nil {
				errs = append(errs, err)
			}
			// delete
			err = r.TestDelete(res, tdName, repo)
			if err != nil {
				errs = append(errs, err)
			}
			return errors.Join(errs...)
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	// other tests with expected failures to push the manifest
	for _, failTest := range dataFailManifestTests {
		tdName := failTest.tdName
		err = r.ChildRun(tdName, r.Results, func(r *runner, res *results) error {
			errs := []error{}
			for dig := range r.State.Data[tdName].blobs {
				err := r.TestPushBlobAny(res, tdName, repo, dig)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to push blob %s%.0w", dig.String(), err))
				}
			}
			for i, dig := range r.State.Data[tdName].manOrder {
				err := r.TestPushManifestDigest(res, tdName, repo, dig, failTest.opts...)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to push manifest %d, digest %s%.0w", i, dig.String(), err))
				}
			}
			if len(errs) > 0 {
				return errors.Join(errs...)
			}
			return nil
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	// various manifest error conditions
	err = r.TestManifestErrors(r.Results, repo)
	if err != nil {
		errs = append(errs, err)
	}

	r.Results.Stop = time.Now()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (r *runner) TestBlobAPIs(parent *results, tdName, tdDesc string, algo digest.Algorithm, repo, repo2 string) error {
	return r.ChildRun(algo.String()+" blobs", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPush); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPush)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		errs := []error{}
		r.State.Data[tdName] = newTestData(tdDesc)
		r.State.DataStatus[tdName] = statusUnknown
		digests := map[string]digest.Digest{}
		if _, ok := blobAPIsTestedByAlgo[algo]; !ok {
			blobAPIsTestedByAlgo[algo] = &[stateAPIMax]bool{}
		}
		blobAPITests := []string{"post only", "post+put", "chunked single", "stream", "mount", "mount anonymous", "mount missing", "post cancel"}
		for _, name := range blobAPITests {
			dig, _, err := r.State.Data[tdName].genBlob(genWithBlobSize(512), genWithAlgo(algo))
			if err != nil {
				return fmt.Errorf("failed to generate blob: %w", err)
			}
			digests[name] = dig
		}
		// try pulling a blob that has not been pushed
		err := r.ChildRun("get-missing", res, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIBlobGetFull); err != nil {
				r.TestSkip(res, err, tdName, stateAPIBlobGetFull)
				return fmt.Errorf("%.0w%w", errAPITestSkip, err)
			}
			if err := r.API.BlobGetReq(r.Config.schemeReg, repo, digests["post cancel"], r.State.Data[tdName], apiExpectStatus(http.StatusNotFound), apiSaveOutput(res.Output)); err != nil {
				r.TestFail(res, err, tdName, stateAPIBlobGetFull)
				return fmt.Errorf("%.0w%w", errAPITestFail, err)
			}
			r.TestPass(res, tdName, stateAPIBlobGetFull)
			return nil
		})
		if err != nil {
			errs = append(errs, err)
		}
		blobAPITests = append(blobAPITests, "chunked multi", "chunked multi and put chunk", "chunked out-of-order", "chunked out-of-order and put chunk")
		minChunkSize := int64(chunkMin)
		minHeader := ""
		// test the various blob push APIs
		for _, testName := range blobAPITests {
			err := r.ChildRun(testName, res, func(r *runner, res *results) error {
				var err error
				errs := []error{}
				dig := digests[testName]
				var api stateAPIType
				switch testName {
				case "post only":
					api = stateAPIBlobPostOnly
					err = r.TestPushBlobPostOnly(res, tdName, repo, dig)
					if err != nil {
						errs = append(errs, err)
					}
				case "post+put":
					api = stateAPIBlobPostPut
					err = r.TestPushBlobPostPut(res, tdName, repo, dig)
					if err != nil {
						errs = append(errs, err)
					}
				case "chunked single":
					api = stateAPIBlobPatchChunked
					// extract the min chunk length from a chunked push with a single chunk
					err = r.TestPushBlobPatchChunked(res, tdName, repo, dig, apiReturnHeader("OCI-Chunk-Min-Length", &minHeader))
					if err != nil {
						errs = append(errs, err)
					}
					if minHeader != "" {
						minParse, err := strconv.Atoi(minHeader)
						if err == nil && int64(minParse) > minChunkSize {
							minChunkSize = int64(minParse)
						}
					}
				case "chunked multi":
					api = stateAPIBlobPatchChunked
					// generate a blob large enough to span three chunks
					dig, _, err = r.State.Data[tdName].genBlob(genWithBlobSize(minChunkSize*3-5), genWithAlgo(algo))
					if err != nil {
						return fmt.Errorf("failed to generate chunked blob of size %d: %w", minChunkSize*3-5, err)
					}
					digests[testName] = dig
					err = r.TestPushBlobPatchChunked(res, tdName, repo, dig)
					if err != nil {
						errs = append(errs, err)
					}
				case "chunked multi and put chunk":
					api = stateAPIBlobPatchChunked
					// generate a blob large enough to span three chunks
					dig, _, err = r.State.Data[tdName].genBlob(genWithBlobSize(minChunkSize*3-5), genWithAlgo(algo))
					if err != nil {
						return fmt.Errorf("failed to generate chunked blob of size %d: %w", minChunkSize*3-5, err)
					}
					digests[testName] = dig
					err = r.TestPushBlobPatchChunked(res, tdName, repo, dig, apiWithFlag("PutLastChunk"))
					if err != nil {
						errs = append(errs, err)
					}
				case "chunked out-of-order":
					api = stateAPIBlobPatchChunked
					// generate a blob large enough to span three chunks
					dig, _, err = r.State.Data[tdName].genBlob(genWithBlobSize(minChunkSize*3-5), genWithAlgo(algo))
					if err != nil {
						return fmt.Errorf("failed to generate chunked blob of size %d: %w", minChunkSize*3-5, err)
					}
					digests[testName] = dig
					err = r.TestPushBlobPatchChunked(res, tdName, repo, dig, apiWithFlag("OutOfOrderChunks"))
					if err != nil {
						errs = append(errs, err)
					}
				case "chunked out-of-order and put chunk":
					api = stateAPIBlobPatchChunked
					// generate a blob large enough to span three chunks
					dig, _, err = r.State.Data[tdName].genBlob(genWithBlobSize(minChunkSize*3-5), genWithAlgo(algo))
					if err != nil {
						return fmt.Errorf("failed to generate chunked blob of size %d: %w", minChunkSize*3-5, err)
					}
					digests[testName] = dig
					err = r.TestPushBlobPatchChunked(res, tdName, repo, dig, apiWithFlag("PutLastChunk"), apiWithFlag("OutOfOrderChunks"))
					if err != nil {
						errs = append(errs, err)
					}
				case "stream":
					api = stateAPIBlobPatchStream
					err = r.TestPushBlobPatchStream(res, tdName, repo, dig)
					if err != nil {
						errs = append(errs, err)
					}
				case "post cancel":
					api = stateAPIBlobPostPut
					err = r.TestPushBlobPostCancel(res, tdName, repo, dig)
					if err != nil {
						errs = append(errs, err)
					}
				case "mount":
					api = stateAPIBlobMountSource
					// first push to repo2
					err = r.TestPushBlobAny(res, tdName, repo2, dig)
					if err != nil {
						errs = append(errs, err)
					}
					// then mount repo2 to repo
					err = r.TestPushBlobMount(res, tdName, repo, repo2, dig)
					if err != nil {
						errs = append(errs, err)
					}
				case "mount anonymous":
					api = stateAPIBlobMountAnonymous
					// first push to repo2
					err = r.TestPushBlobAny(res, tdName, repo2, dig)
					if err != nil {
						errs = append(errs, err)
					}
					// then mount repo2 to repo
					err = r.TestPushBlobMountAnonymous(res, tdName, repo, dig)
					if err != nil {
						errs = append(errs, err)
					}
				case "mount missing":
					// mount repo2 to repo without first pushing there
					err = r.TestPushBlobMountMissing(res, tdName, repo, repo2, dig)
					if err != nil {
						errs = append(errs, err)
					}
				default:
					return fmt.Errorf("unknown api test %s", testName)
				}
				// track the used APIs so TestPushBlobAny doesn't rerun tests
				blobAPIsTested[api] = true
				blobAPIsTestedByAlgo[dig.Algorithm()][api] = true
				if err == nil && testName != "post cancel" {
					// head request
					err = r.TestHeadBlob(res, tdName, repo, dig)
					if err != nil {
						errs = append(errs, err)
					}
					// pull each blob
					err = r.TestPullBlob(res, tdName, repo, dig)
					if err != nil {
						errs = append(errs, err)
					}
				}
				// cleanup
				err = r.TestDeleteBlob(res, tdName, repo, dig)
				if err != nil {
					errs = append(errs, err)
				}
				if testName == "mount" || testName == "mount anonymous" {
					err = r.TestDeleteBlob(res, tdName, repo2, dig)
					if err != nil {
						errs = append(errs, err)
					}
				}
				return errors.Join(errs...)
			})
			if err != nil {
				errs = append(errs, err)
			}
		}
		// verify support for range requests
		err = r.ChildRun("range requests", res, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIBlobGetRange, stateAPIBlobPush); err != nil {
				r.TestSkip(res, err, tdName, stateAPIBlobGetRange)
				return fmt.Errorf("%.0w%w", errAPITestSkip, err)
			}
			// setup by pushing a blob, any failures will return immediately
			blobLen := int64(2048)
			blobLenStr := fmt.Sprintf("%d", blobLen)
			dig, blobBody, err := r.State.Data[tdName].genBlob(genWithBlobSize(blobLen), genWithAlgo(algo))
			if err != nil {
				return err
			}
			if err := r.TestPushBlobAny(res, tdName, repo, dig); err != nil {
				r.TestSkip(res, err, tdName, stateAPIBlobGetRange)
				return err
			}
			errs := []error{}
			rangeTests := []struct {
				name     string
				reqOpts  []apiDoOpt
				respOpts []apiDoOpt // response opts are separated to run them conditionally with a fallback
			}{
				{
					name: "range 500-1499",
					reqOpts: []apiDoOpt{
						apiWithHeaderAdd("Range", "bytes=500-1499"),
					},
					respOpts: []apiDoOpt{
						apiExpectBody(blobBody[500:1500]),
						apiExpectStatus(http.StatusPartialContent),
						apiExpectHeader("Content-Length", "1000"),
						apiWithOr(
							[]apiDoOpt{apiExpectHeader("Content-Range", "bytes 500-1499/"+blobLenStr)},
							[]apiDoOpt{apiExpectHeader("Content-Range", "bytes 500-1499/*")},
						),
					},
				},
				{
					name: "range 500-",
					reqOpts: []apiDoOpt{
						apiWithHeaderAdd("Range", "bytes=500-"),
					},
					respOpts: []apiDoOpt{
						apiExpectBody(blobBody[500:]),
						apiExpectStatus(http.StatusPartialContent),
						apiExpectHeader("Content-Length", fmt.Sprintf("%d", blobLen-500)),
						apiWithOr(
							[]apiDoOpt{apiExpectHeader("Content-Range", fmt.Sprintf("bytes 500-%d/%d", blobLen-1, blobLen))},
							[]apiDoOpt{apiExpectHeader("Content-Range", fmt.Sprintf("bytes 500-%d/*", blobLen-1))},
						),
					},
				},
				{
					name: "range -500",
					reqOpts: []apiDoOpt{
						apiWithHeaderAdd("Range", "bytes=-500"),
					},
					respOpts: []apiDoOpt{
						apiExpectBody(blobBody[blobLen-500:]),
						apiExpectStatus(http.StatusPartialContent),
						apiExpectHeader("Content-Length", "500"),
						apiWithOr(
							[]apiDoOpt{apiExpectHeader("Content-Range", fmt.Sprintf("bytes %d-%d/%d", blobLen-500, blobLen-1, blobLen))},
							[]apiDoOpt{apiExpectHeader("Content-Range", fmt.Sprintf("bytes %d-%d/*", blobLen-500, blobLen-1))},
						),
					},
				},
				{
					name: "range 2000-5000",
					reqOpts: []apiDoOpt{
						apiWithHeaderAdd("Range", "bytes=2000-5000"),
					},
					respOpts: []apiDoOpt{
						apiExpectBody(blobBody[2000:]),
						apiExpectStatus(http.StatusPartialContent),
						apiExpectHeader("Content-Length", fmt.Sprintf("%d", blobLen-2000)),
						apiWithOr(
							[]apiDoOpt{apiExpectHeader("Content-Range", fmt.Sprintf("bytes %d-%d/%d", 2000, blobLen-1, blobLen))},
							[]apiDoOpt{apiExpectHeader("Content-Range", fmt.Sprintf("bytes %d-%d/*", 2000, blobLen-1))},
						),
					},
				},
				{
					name: "range 500-0",
					reqOpts: []apiDoOpt{
						apiWithHeaderAdd("Range", "bytes=500-0"),
					},
					respOpts: []apiDoOpt{
						apiExpectStatus(http.StatusRequestedRangeNotSatisfiable),
					},
				},
				{
					name: "range 5000-10000",
					reqOpts: []apiDoOpt{
						apiWithHeaderAdd("Range", "bytes=5000-10000"),
					},
					respOpts: []apiDoOpt{
						apiExpectStatus(http.StatusRequestedRangeNotSatisfiable),
					},
				},
			}
			for _, rt := range rangeTests {
				err := r.ChildRun(rt.name, res, func(r *runner, res *results) error {
					var status int
					rangeOpts := []apiDoOpt{
						apiSaveOutput(res.Output),
						apiWithAnd(rt.reqOpts),
						apiWithOr(rt.respOpts,
							[]apiDoOpt{ // if rt.opts fails, it may fall back to a standard blob pull
								apiExpectStatus(http.StatusOK),
								apiExpectHeader("Content-Length", blobLenStr),
								apiExpectBody(blobBody),
								apiReturnStatus(&status),
							}),
					}
					if err := r.API.BlobGetReq(r.Config.schemeReg, repo, dig, r.State.Data[tdName], rangeOpts...); err != nil {
						r.TestFail(res, err, tdName, stateAPIBlobGetRange)
						return fmt.Errorf("%.0w%w", errAPITestFail, err)
					}
					// detect a fallback
					if status == http.StatusOK {
						err := fmt.Errorf("range request unsupported, full blob returned%.0w", errRegUnsupported)
						r.TestFail(res, err, tdName, stateAPIBlobGetRange)
						return fmt.Errorf("%.0w%w", errAPITestFail, err)
					}
					r.TestPass(res, tdName, stateAPIBlobGetRange)
					return nil
				})
				if err != nil {
					errs = append(errs, err)
				}
			}
			if err := r.TestDeleteBlob(res, tdName, repo, dig); err != nil {
				errs = append(errs, err)
			}
			return errors.Join(errs...)
		})
		if err != nil {
			errs = append(errs, err)
		}
		// test various well known blob contents
		blobDataTests := map[string][]byte{}
		if r.Config.Data.EmptyBlob {
			blobDataTests["empty"] = []byte("")
		}
		blobDataTests["emptyJSON"] = []byte("{}")
		for name, val := range blobDataTests {
			dig := algo.FromBytes(val)
			digests[name] = dig
			r.State.Data[tdName].blobs[dig] = val
		}
		for name := range blobDataTests {
			err := r.ChildRun(name, res, func(r *runner, res *results) error {
				dig := digests[name]
				err := r.TestPushBlobAny(res, tdName, repo, dig)
				if err != nil {
					errs = append(errs, err)
				}
				err = r.TestHeadBlob(res, tdName, repo, dig)
				if err != nil {
					errs = append(errs, err)
				}
				err = r.TestPullBlob(res, tdName, repo, dig)
				if err != nil {
					errs = append(errs, err)
				}
				err = r.TestDeleteBlob(res, tdName, repo, dig)
				if err != nil {
					errs = append(errs, err)
				}
				return errors.Join(errs...)
			})
			if err != nil {
				errs = append(errs, err)
			}
		}
		// test the various blob push APIs with a bad digest
		blobAPIBadDigTests := []string{"bad digest post only", "bad digest post+put", "bad digest chunked", "bad digest chunked and put chunk", "bad digest stream"}
		for _, name := range blobAPIBadDigTests {
			dig, _, err := r.State.Data[tdName].genBlob(genWithBlobSize(minChunkSize*3-5), genWithAlgo(algo))
			if err != nil {
				return fmt.Errorf("failed to generate blob: %w", err)
			}
			// corrupt the blob bytes
			r.State.Data[tdName].blobs[dig] = append(r.State.Data[tdName].blobs[dig], []byte("oh no")...)
			digests[name] = dig
		}
		optBadDig := apiWithFlag("ExpectBadDigest")
		for _, testName := range blobAPIBadDigTests {
			err := r.ChildRun(testName, res, func(r *runner, res *results) error {
				dig := digests[testName]
				switch testName {
				case "bad digest post only":
					return r.TestPushBlobPostOnly(res, tdName, repo, dig, optBadDig)
				case "bad digest post+put":
					return r.TestPushBlobPostPut(res, tdName, repo, dig, optBadDig)
				case "bad digest chunked":
					return r.TestPushBlobPatchChunked(res, tdName, repo, dig, optBadDig)
				case "bad digest chunked and put chunk":
					return r.TestPushBlobPatchChunked(res, tdName, repo, dig, optBadDig)
				case "bad digest stream":
					return r.TestPushBlobPatchStream(res, tdName, repo, dig, optBadDig)
				default:
					return fmt.Errorf("unknown api test %s", testName)
				}
			})
			if err != nil {
				errs = append(errs, err)
			}
		}

		return errors.Join(errs...)
	})
}

func (r *runner) TestDelete(parent *results, tdName string, repo string) error {
	return r.ChildRun("delete", parent, func(r *runner, res *results) error {
		errs := []error{}
		delOrder := slices.Clone(r.State.Data[tdName].manOrder)
		slices.Reverse(delOrder)
		for tag, dig := range r.State.Data[tdName].tags {
			err := r.TestDeleteTag(res, tdName, repo, tag, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to delete manifest tag %s%.0w", tag, err))
			}
		}
		for i, dig := range delOrder {
			err := r.TestDeleteManifest(res, tdName, repo, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to delete manifest %d, digest %s%.0w", i, dig.String(), err))
			}
		}
		for dig := range r.State.Data[tdName].blobs {
			err := r.TestDeleteBlob(res, tdName, repo, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to delete blob %s%.0w", dig.String(), err))
			}
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		return nil
	})
}

func (r *runner) TestDeleteTag(parent *results, tdName string, repo string, tag string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	return r.ChildRun("tag-delete", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPITagDelete); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPITagDelete, stateAPITagDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if err := r.API.ManifestDelete(r.Config.schemeReg, repo, tag, dig, td, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPITagDelete)
			r.TestSkip(res, err, tdName, stateAPITagDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPITagDelete)
		// verify tag delete finished immediately
		if err := r.APIRequire(stateAPITagDeleteAtomic); err != nil {
			r.TestSkip(res, err, tdName, stateAPITagDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if err := r.API.ManifestHeadReq(r.Config.schemeReg, repo, tag, dig, td, apiSaveOutput(res.Output), apiExpectStatus(http.StatusNotFound)); err != nil {
			r.TestFail(res, err, tdName, stateAPITagDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPITagDeleteAtomic)
		return nil
	})
}

func (r *runner) TestDeleteManifest(parent *results, tdName string, repo string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	return r.ChildRun("manifest-delete", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestDelete); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestDelete, stateAPIManifestDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if err := r.API.ManifestDelete(r.Config.schemeReg, repo, dig.String(), dig, td, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIManifestDelete)
			r.TestSkip(res, err, tdName, stateAPIManifestDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIManifestDelete)
		// verify manifest delete finished immediately
		if err := r.APIRequire(stateAPIManifestDeleteAtomic); err != nil {
			r.TestSkip(res, err, tdName, stateAPIManifestDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if err := r.API.ManifestHeadReq(r.Config.schemeReg, repo, dig.String(), dig, td, apiSaveOutput(res.Output), apiExpectStatus(http.StatusNotFound)); err != nil {
			r.TestFail(res, err, tdName, stateAPIManifestDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIManifestDeleteAtomic)
		return nil
	})
}

func (r *runner) TestDeleteBlob(parent *results, tdName string, repo string, dig digest.Digest) error {
	return r.ChildRun("blob-delete", parent, func(r *runner, res *results) error {
		td := r.State.Data[tdName]
		if err := r.APIRequire(stateAPIBlobDelete); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobDelete, stateAPIBlobDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if err := r.API.BlobDelete(r.Config.schemeReg, repo, dig, td, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobDelete)
			r.TestSkip(res, err, tdName, stateAPIBlobDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobDelete)
		// verify blob delete finished immediately
		if err := r.APIRequire(stateAPIBlobDeleteAtomic); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if err := r.API.BlobHeadReq(r.Config.schemeReg, repo, dig, td, apiSaveOutput(res.Output), apiExpectStatus(http.StatusNotFound)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobDeleteAtomic)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobDeleteAtomic)
		return nil
	})
}

func (r *runner) TestEmpty(parent *results, repo string) error {
	return r.ChildRun("empty", parent, func(r *runner, res *results) error {
		errs := []error{}
		if err := r.TestEmptyTagList(res, repo); err != nil {
			errs = append(errs, err)
		}
		if err := r.TestEmptyReferrers(res, repo); err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		return nil
	})
}

func (r *runner) TestEmptyReferrers(parent *results, repo string) error {
	return r.ChildRun("referrers", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIReferrers); err != nil {
			r.TestSkip(res, err, "", stateAPIReferrers)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		subj := digest.Canonical.FromString(rand.Text())
		_, err := r.API.ReferrersList(r.Config.schemeReg, repo, subj, apiSaveOutput(res.Output))
		if err != nil {
			r.TestFail(res, err, "", stateAPIReferrers)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, "", stateAPIReferrers)
		return nil
	})
}

func (r *runner) TestEmptyTagList(parent *results, repo string) error {
	return r.ChildRun("tag list", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPITagList); err != nil {
			r.TestSkip(res, err, "", stateAPITagList)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if _, err := r.API.TagList(r.Config.schemeReg, repo, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, "", stateAPITagList)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, "", stateAPITagList)
		return nil
	})
}

func (r *runner) TestHead(parent *results, tdName string, repo string) error {
	return r.ChildRun("head", parent, func(r *runner, res *results) error {
		errs := []error{}
		for tag, dig := range r.State.Data[tdName].tags {
			err := r.TestHeadManifestTag(res, tdName, repo, tag, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to send head request for manifest by tag %s%.0w", tag, err))
			}
		}
		for i, dig := range r.State.Data[tdName].manOrder {
			err := r.TestHeadManifestDigest(res, tdName, repo, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to send head request for manifest %d, digest %s%.0w", i, dig.String(), err))
			}
		}
		for dig := range r.State.Data[tdName].blobs {
			err := r.TestHeadBlob(res, tdName, repo, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to send head request for blob %s%.0w", dig.String(), err))
			}
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		return nil
	})
}

func (r *runner) TestHeadBlob(parent *results, tdName string, repo string, dig digest.Digest) error {
	return r.ChildRun("blob-head", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobHead); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobHead)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if err := r.API.BlobHeadExists(r.Config.schemeReg, repo, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobHead)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobHead)
		return nil
	})
}

func (r *runner) TestHeadManifestDigest(parent *results, tdName string, repo string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	opts := []apiDoOpt{}
	apis := []stateAPIType{}
	return r.ChildRun("manifest-head-by-digest", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestHeadDigest); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestHeadDigest)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		apis = append(apis, stateAPIManifestHeadDigest)
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.ManifestHeadExists(r.Config.schemeReg, repo, dig.String(), dig, td, opts...); err != nil {
			r.TestFail(res, err, tdName, apis...)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, apis...)
		return nil
	})
}

func (r *runner) TestHeadManifestTag(parent *results, tdName string, repo string, tag string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	opts := []apiDoOpt{}
	apis := []stateAPIType{}
	return r.ChildRun("manifest-head-by-tag", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestHeadTag); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestHeadTag)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		apis = append(apis, stateAPIManifestHeadTag)
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.ManifestHeadExists(r.Config.schemeReg, repo, tag, dig, td, opts...); err != nil {
			r.TestFail(res, err, tdName, apis...)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, apis...)
		return nil
	})
}

func (r *runner) TestList(parent *results, tdName string, repo string) error {
	return r.ChildRun("tag-list", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPITagList); err != nil {
			r.TestSkip(res, err, tdName, stateAPITagList)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		tagList, err := r.API.TagList(r.Config.schemeReg, repo, apiSaveOutput(res.Output))
		if err != nil {
			r.TestFail(res, err, tdName, stateAPITagList)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		errs := []error{}
		for tag := range r.State.Data[tdName].tags {
			if !slices.Contains(tagList.Tags, tag) {
				errs = append(errs, fmt.Errorf("missing tag %q from listing%.0w", tag, errAPITestFail))
			}
		}
		if len(errs) > 0 {
			r.TestFail(res, errors.Join(errs...), tdName, stateAPITagList)
			return errors.Join(errs...)
		}
		r.TestPass(res, tdName, stateAPITagList)
		return nil
	})
}

func (r *runner) TestManifestErrors(parent *results, repo string) error {
	errs := []error{}
	err := r.ChildRun("missing-manifest", parent, func(r *runner, res *results) error {
		errs := []error{}
		err := r.ChildRun("by-digest", res, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestGetDigest); err != nil {
				r.TestSkip(res, err, "", stateAPIManifestGetDigest)
				return fmt.Errorf("%.0w%w", errAPITestSkip, err)
			}
			b := make([]byte, 1024)
			_, err := rand.Read(b)
			if err != nil {
				return err
			}
			dig := digest.Canonical.FromBytes(b)
			if err := r.API.ManifestGetReq(r.Config.schemeReg, repo, dig.String(), dig, nil,
				apiExpectStatus(http.StatusNotFound), apiSaveOutput(res.Output)); err != nil {
				r.TestFail(res, err, "", stateAPIManifestGetDigest)
				return fmt.Errorf("%.0w%w", errAPITestFail, err)
			}
			r.TestPass(res, "", stateAPIManifestGetDigest)
			return nil
		})
		if err != nil {
			errs = append(errs, err)
		}
		err = r.ChildRun("by-tag", res, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestGetTag); err != nil {
				r.TestSkip(res, err, "", stateAPIManifestGetTag)
				return fmt.Errorf("%.0w%w", errAPITestSkip, err)
			}
			rnd := rand.Text()
			tag := fmt.Sprintf("missing-%.20s", strings.ToLower(rnd))
			if err := r.API.ManifestGetReq(r.Config.schemeReg, repo, tag, digest.Digest(""), nil,
				apiExpectStatus(http.StatusNotFound), apiSaveOutput(res.Output)); err != nil {
				r.TestFail(res, err, "", stateAPIManifestGetTag)
				return fmt.Errorf("%.0w%w", errAPITestFail, err)
			}
			r.TestPass(res, "", stateAPIManifestGetTag)
			return nil
		})
		if err != nil {
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	})
	if err != nil {
		errs = append(errs, err)
	}
	err = r.ChildRun("invalid-digest-format", parent, func(r *runner, res *results) error {
		errs := []error{}
		tdName := "invalid-manifest-digest"
		r.State.Data[tdName] = newTestData("invalid manifest digest")
		manDig, err := r.State.Data[tdName].genManifestFull(genWithLayerCount(1))
		if err != nil {
			return err
		}
		for dig := range r.State.Data[tdName].blobs {
			err := r.TestPushBlobAny(res, tdName, repo, dig)
			if err != nil {
				errs = append(errs, err)
			}
		}
		// push digest "sha256:baddigeststring"
		err = r.ChildRun("manifest-put", res, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestPutDigest); err != nil {
				r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
				r.TestSkip(res, err, tdName, stateAPIManifestPutDigest)
				return fmt.Errorf("%.0w%w", errAPITestSkip, err)
			}
			if err := r.API.ManifestPut(r.Config.schemeReg, repo, "sha256:baddigeststring", manDig, r.State.Data[tdName], r.Config.APIs.Referrer,
				apiWithFlag("ExpectBadDigest"), apiSaveOutput(res.Output)); err != nil {
				r.TestFail(res, err, tdName, stateAPIManifestPutDigest)
				return fmt.Errorf("%.0w%w", errAPITestFail, err)
			}
			r.TestPass(res, tdName, stateAPIManifestPutDigest)
			return nil
		})
		if err != nil {
			errs = append(errs, err)
		}
		// pull digest "sha256:baddigeststring"
		err = r.ChildRun("manifest-get", res, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestGetDigest); err != nil {
				r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
				r.TestSkip(res, err, tdName, stateAPIManifestGetDigest)
				return fmt.Errorf("%.0w%w", errAPITestSkip, err)
			}
			if err := r.API.ManifestGetReq(r.Config.schemeReg, repo, "sha256:baddigeststring", manDig, r.State.Data[tdName],
				apiExpectStatus(http.StatusNotFound, http.StatusBadRequest), apiSaveOutput(res.Output)); err != nil {
				r.TestFail(res, err, tdName, stateAPIManifestGetDigest)
				return fmt.Errorf("%.0w%w", errAPITestFail, err)
			}
			r.TestPass(res, tdName, stateAPIManifestGetDigest)
			return nil
		})
		if err != nil {
			errs = append(errs, err)
		}
		// cleanup
		err = r.TestDelete(res, tdName, repo)
		if err != nil {
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	})
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (r *runner) TestPull(parent *results, tdName string, repo string) error {
	return r.ChildRun("pull", parent, func(r *runner, res *results) error {
		errs := []error{}
		for tag, dig := range r.State.Data[tdName].tags {
			err := r.TestPullManifestTag(res, tdName, repo, tag, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to pull manifest by tag %s%.0w", tag, err))
			}
		}
		for i, dig := range r.State.Data[tdName].manOrder {
			err := r.TestPullManifestDigest(res, tdName, repo, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to pull manifest %d, digest %s%.0w", i, dig.String(), err))
			}
		}
		for dig := range r.State.Data[tdName].blobs {
			err := r.TestPullBlob(res, tdName, repo, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to pull blob %s%.0w", dig.String(), err))
			}
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		return nil
	})
}

func (r *runner) TestPullBlob(parent *results, tdName string, repo string, dig digest.Digest) error {
	return r.ChildRun("blob-get", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobGetFull); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobGetFull)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if err := r.API.BlobGetExistsFull(r.Config.schemeReg, repo, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobGetFull)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobGetFull)
		return nil
	})
}

func (r *runner) TestPullManifestDigest(parent *results, tdName string, repo string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	opts := []apiDoOpt{}
	apis := []stateAPIType{}
	return r.ChildRun("manifest-by-digest", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestGetDigest); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestGetDigest)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		apis = append(apis, stateAPIManifestGetDigest)
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.ManifestGetExists(r.Config.schemeReg, repo, dig.String(), dig, td, opts...); err != nil {
			r.TestFail(res, err, tdName, apis...)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, apis...)
		return nil
	})
}

func (r *runner) TestPullManifestTag(parent *results, tdName string, repo string, tag string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	opts := []apiDoOpt{}
	apis := []stateAPIType{}
	return r.ChildRun("manifest-by-tag", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestGetTag); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestGetTag)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		apis = append(apis, stateAPIManifestGetTag)
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.ManifestGetExists(r.Config.schemeReg, repo, tag, dig, td, opts...); err != nil {
			r.TestFail(res, err, tdName, apis...)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, apis...)
		return nil
	})
}

func (r *runner) TestPush(parent *results, tdName string, repo string) error {
	return r.ChildRun("push", parent, func(r *runner, res *results) error {
		errs := []error{}
		for dig := range r.State.Data[tdName].blobs {
			err := r.TestPushBlobAny(res, tdName, repo, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to push blob %s%.0w", dig.String(), err))
			}
		}
		for i, dig := range r.State.Data[tdName].manOrder {
			err := r.TestPushManifestDigest(res, tdName, repo, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to push manifest %d, digest %s%.0w", i, dig.String(), err))
			}
		}
		for tag, dig := range r.State.Data[tdName].tags {
			err := r.TestPushManifestTag(res, tdName, repo, tag, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to push manifest tag %s%.0w", tag, err))
			}
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		return nil
	})
}

var (
	blobAPIs             = []stateAPIType{stateAPIBlobPostPut, stateAPIBlobPostOnly, stateAPIBlobPatchStream, stateAPIBlobPatchChunked}
	blobAPIsTested       = [stateAPIMax]bool{}
	blobAPIsTestedByAlgo = map[digest.Algorithm]*[stateAPIMax]bool{}
)

func (r *runner) TestPushBlobAny(parent *results, tdName string, repo string, dig digest.Digest, opts ...apiDoOpt) error {
	if err := r.APIRequire(stateAPIBlobPush); err != nil {
		return fmt.Errorf("%.0w%w", errAPITestSkip, err)
	}
	apis := []stateAPIType{}
	if _, ok := blobAPIsTestedByAlgo[dig.Algorithm()]; !ok {
		blobAPIsTestedByAlgo[dig.Algorithm()] = &[stateAPIMax]bool{}
	}
	// first try untested APIs
	for _, api := range blobAPIs {
		if !blobAPIsTested[api] {
			apis = append(apis, api)
		}
	}
	// then untested with a given algorithm
	for _, api := range blobAPIs {
		if !blobAPIsTestedByAlgo[dig.Algorithm()][api] && !slices.Contains(apis, api) {
			apis = append(apis, api)
		}
	}
	// next use APIs that are known successful
	for _, api := range blobAPIs {
		if r.State.APIStatus[api] == statusPass && !slices.Contains(apis, api) {
			apis = append(apis, api)
		}
	}
	// lastly use APIs in preferred order
	for _, api := range blobAPIs {
		if !slices.Contains(apis, api) {
			apis = append(apis, api)
		}
	}
	// return on the first successful API
	errs := []error{}
	for _, api := range apis {
		var err error
		switch api {
		case stateAPIBlobPostPut:
			err = r.TestPushBlobPostPut(parent, tdName, repo, dig, opts...)
		case stateAPIBlobPostOnly:
			err = r.TestPushBlobPostOnly(parent, tdName, repo, dig, opts...)
		case stateAPIBlobPatchStream:
			err = r.TestPushBlobPatchStream(parent, tdName, repo, dig, opts...)
		case stateAPIBlobPatchChunked:
			err = r.TestPushBlobPatchChunked(parent, tdName, repo, dig, opts...)
		default:
			err = fmt.Errorf("blob API %s is not handled by TestPushBlobAny", api.String())
		}
		blobAPIsTested[api] = true
		blobAPIsTestedByAlgo[dig.Algorithm()][api] = true
		if err == nil {
			return nil
		}
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (r *runner) TestPushBlobMount(parent *results, tdName string, repo, repo2 string, dig digest.Digest) error {
	return r.ChildRun("blob-mount", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobMountSource); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobMountSource)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if err := r.API.BlobMount(r.Config.schemeReg, repo, repo2, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobMountSource)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobMountSource)
		return nil
	})
}

func (r *runner) TestPushBlobMountAnonymous(parent *results, tdName string, repo string, dig digest.Digest) error {
	return r.ChildRun("blob-mount-anonymous", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobMountAnonymous); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobMountAnonymous)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if err := r.API.BlobMount(r.Config.schemeReg, repo, "", dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobMountAnonymous)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobMountAnonymous)
		return nil
	})
}

func (r *runner) TestPushBlobMountMissing(parent *results, tdName string, repo, repo2 string, dig digest.Digest) error {
	return r.ChildRun("blob-mount", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobMountSource); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobMountSource)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		if err := r.API.BlobMount(r.Config.schemeReg, repo, repo2, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); !errors.Is(err, errRegUnsupported) {
			if err == nil {
				err = fmt.Errorf("blob mount of missing blob incorrectly succeeded")
			}
			r.TestFail(res, err, tdName, stateAPIBlobMountSource)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobMountSource)
		return nil
	})
}

func (r *runner) TestPushBlobPostCancel(parent *results, tdName string, repo string, dig digest.Digest, opts ...apiDoOpt) error {
	return r.ChildRun("blob-post-cancel", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPush, stateAPIBlobCancel); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPush, stateAPIBlobCancel)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.BlobPostCancel(r.Config.schemeReg, repo, dig, r.State.Data[tdName], opts...); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPush, stateAPIBlobCancel)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPush, stateAPIBlobCancel)
		return nil
	})
}

func (r *runner) TestPushBlobPostPut(parent *results, tdName string, repo string, dig digest.Digest, opts ...apiDoOpt) error {
	return r.ChildRun("blob-post-put", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPostPut); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPostPut)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.BlobPostPut(r.Config.schemeReg, repo, dig, r.State.Data[tdName], opts...); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPostPut)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPostPut, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushBlobPostOnly(parent *results, tdName string, repo string, dig digest.Digest, opts ...apiDoOpt) error {
	return r.ChildRun("blob-post-only", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPostOnly); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPostOnly)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.BlobPostOnly(r.Config.schemeReg, repo, dig, r.State.Data[tdName], opts...); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPostOnly)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPostOnly, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushBlobPatchChunked(parent *results, tdName string, repo string, dig digest.Digest, opts ...apiDoOpt) error {
	return r.ChildRun("blob-patch-chunked", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPatchChunked); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPatchChunked)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.BlobPatchChunked(r.Config.schemeReg, repo, dig, r.State.Data[tdName], opts...); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPatchChunked)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPatchChunked, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushBlobPatchStream(parent *results, tdName string, repo string, dig digest.Digest, opts ...apiDoOpt) error {
	return r.ChildRun("blob-patch-stream", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPatchStream); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPatchStream)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.BlobPatchStream(r.Config.schemeReg, repo, dig, r.State.Data[tdName], opts...); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPatchStream)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPatchStream, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushManifestDigest(parent *results, tdName string, repo string, dig digest.Digest, opts ...apiDoOpt) error {
	td := r.State.Data[tdName]
	apis := []stateAPIType{}
	subj := detectSubject(td.manifests[dig])
	if subj != nil {
		apis = append(apis, stateAPIManifestPutSubject)
	}
	return r.ChildRun("manifest-by-digest", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestPutDigest); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestPutDigest)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		apis = append(apis, stateAPIManifestPutDigest)
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.ManifestPut(r.Config.schemeReg, repo, dig.String(), dig, td, r.Config.APIs.Referrer, opts...); err != nil {
			r.TestFail(res, err, tdName, apis...)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, apis...)
		return nil
	})
}

func (r *runner) TestPushManifestTag(parent *results, tdName string, repo string, tag string, dig digest.Digest, opts ...apiDoOpt) error {
	td := r.State.Data[tdName]
	apis := []stateAPIType{}
	subj := detectSubject(td.manifests[dig])
	if subj != nil {
		apis = append(apis, stateAPIManifestPutSubject)
	}
	return r.ChildRun("manifest-by-tag", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestPutTag); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestPutTag)
			return fmt.Errorf("%.0w%w", errAPITestSkip, err)
		}
		apis = append(apis, stateAPIManifestPutTag)
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.ManifestPut(r.Config.schemeReg, repo, tag, dig, td, r.Config.APIs.Referrer, opts...); err != nil {
			r.TestFail(res, err, tdName, apis...)
			return fmt.Errorf("%.0w%w", errAPITestFail, err)
		}
		r.TestPass(res, tdName, apis...)
		return nil
	})
}

func (r *runner) TestReferrers(parent *results, tdName string, repo string) error {
	if len(r.State.Data[tdName].referrers) == 0 {
		return nil
	}
	return r.ChildRun("referrers", parent, func(r *runner, res *results) error {
		errs := []error{}
		for subj, referrerGoal := range r.State.Data[tdName].referrers {
			if err := r.APIRequire(stateAPIReferrers); err != nil {
				r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
				r.TestSkip(res, err, tdName, stateAPIReferrers)
				return fmt.Errorf("%.0w%w", errAPITestSkip, err)
			}
			referrerResp, err := r.API.ReferrersList(r.Config.schemeReg, repo, subj, apiSaveOutput(res.Output))
			if err != nil {
				errs = append(errs, err)
			}
			if err == nil {
				for _, goal := range referrerGoal {
					if !slices.ContainsFunc(referrerResp.Manifests, func(resp image.Descriptor) bool {
						return resp.Digest == goal.Digest &&
							resp.MediaType == goal.MediaType &&
							resp.Size == goal.Size &&
							resp.ArtifactType == goal.ArtifactType &&
							mapContainsAll(resp.Annotations, goal.Annotations)
					}) {
						errs = append(errs, fmt.Errorf("entry missing from referrers list, subject %s, referrer %+v%.0w", subj, goal, errAPITestFail))
					}
				}
			}
		}
		if len(errs) > 0 {
			r.TestFail(res, errors.Join(errs...), tdName, stateAPIReferrers)
			return fmt.Errorf("%.0w%w", errAPITestFail, errors.Join(errs...))
		}
		r.TestPass(res, tdName, stateAPIReferrers)
		return nil
	})
}

func mapContainsAll[K comparable, V comparable](check, goal map[K]V) bool {
	if len(goal) == 0 {
		return true
	}
	for k, v := range goal {
		if found, ok := check[k]; !ok || found != v {
			return false
		}
	}
	return true
}

func (r *runner) ChildRun(name string, parent *results, fn func(*runner, *results) error) error {
	res := resultsNew(name, parent)
	if parent != nil {
		parent.Children = append(parent.Children, res)
	}
	err := fn(r, res)
	res.Stop = time.Now()
	if err != nil && !errors.Is(err, errAPITestFail) && !errors.Is(err, errAPITestSkip) {
		res.Errs = append(res.Errs, err)
		res.Status = res.Status.Set(statusError)
		res.Counts[statusError]++
	}
	if parent != nil {
		for i := range statusMax {
			parent.Counts[i] += res.Counts[i]
		}
		parent.Status = parent.Status.Set(res.Status)
	}
	return err
}

func (r *runner) TestSkip(res *results, err error, tdName string, apis ...stateAPIType) {
	s := statusSkip
	if errors.Is(err, errAPITestDisabled) {
		s = statusDisabled
	}
	res.Status = res.Status.Set(s)
	res.Counts[s]++
	if tdName != "" {
		r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(s)
	}
	for _, a := range apis {
		r.State.APIStatus[a] = r.State.APIStatus[a].Set(s)
	}
	fmt.Fprintf(res.Output, "%s: skipping test:\n  %s\n", res.Name,
		strings.ReplaceAll(err.Error(), "\n", "\n  "))
	r.Log.Info("skipping test", "name", res.Name, "error", err.Error())
}

func (r *runner) TestFail(res *results, err error, tdName string, apis ...stateAPIType) {
	s := statusFail
	if errors.Is(err, errAPITestError) {
		s = statusError
	} else if errors.Is(err, errRegUnsupported) {
		s = statusDisabled
	}
	res.Status = res.Status.Set(s)
	res.Counts[s]++
	res.Errs = append(res.Errs, err)
	if tdName != "" {
		r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(s)
	}
	for _, a := range apis {
		r.State.APIStatus[a] = r.State.APIStatus[a].Set(s)
	}
	if s == statusFail {
		r.Log.Warn("failed test", "name", res.Name, "error", err.Error())
		r.Log.Debug("failed test output", "name", res.Name, "output", res.Output.String())
	}
}

func (r *runner) TestPass(res *results, tdName string, apis ...stateAPIType) {
	res.Status = res.Status.Set(statusPass)
	res.Counts[statusPass]++
	if tdName != "" {
		r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusPass)
	}
	for _, a := range apis {
		r.State.APIStatus[a] = r.State.APIStatus[a].Set(statusPass)
	}
	r.Log.Info("passing test", "name", res.Name)
	r.Log.Debug("passing test output", "name", res.Name, "output", res.Output.String())
}

func (r *runner) APIRequire(apis ...stateAPIType) error {
	errs := []error{}
	for _, a := range apis {
		aText, err := a.MarshalText()
		if err != nil {
			errs = append(errs, fmt.Errorf("unknown api %d", a))
			continue
		}
		// check the configuration disables the api
		configDisabled := false
		switch a {
		case stateAPITagList:
			if !r.Config.APIs.Tags.List {
				configDisabled = true
			}
		case stateAPIManifestGetTag, stateAPIManifestGetDigest, stateAPIBlobGetFull, stateAPIBlobGetRange:
			if !r.Config.APIs.Pull {
				configDisabled = true
			}
		case stateAPIManifestPutTag, stateAPIManifestPutDigest, stateAPIManifestPutSubject,
			stateAPIBlobPush, stateAPIBlobPostOnly, stateAPIBlobPostPut,
			stateAPIBlobPatchChunked, stateAPIBlobPatchStream, stateAPIBlobMountSource:
			if !r.Config.APIs.Push {
				configDisabled = true
			}
		case stateAPIBlobCancel:
			if !r.Config.APIs.Blobs.UploadCancel {
				configDisabled = true
			}
		case stateAPIBlobMountAnonymous:
			if !r.Config.APIs.Push || !r.Config.APIs.Blobs.MountAnonymous {
				configDisabled = true
			}
		case stateAPITagDelete:
			if !r.Config.APIs.Tags.Delete {
				configDisabled = true
			}
		case stateAPITagDeleteAtomic:
			if !r.Config.APIs.Tags.Delete || !r.Config.APIs.Tags.Atomic {
				configDisabled = true
			}
		case stateAPIManifestDelete:
			if !r.Config.APIs.Manifests.Delete {
				configDisabled = true
			}
		case stateAPIManifestDeleteAtomic:
			if !r.Config.APIs.Manifests.Atomic {
				configDisabled = true
			}
		case stateAPIBlobDelete:
			if !r.Config.APIs.Blobs.Delete {
				configDisabled = true
			}
		case stateAPIBlobDeleteAtomic:
			if !r.Config.APIs.Blobs.Atomic {
				configDisabled = true
			}
		case stateAPIReferrers:
			if !r.Config.APIs.Referrer {
				configDisabled = true
			}
		}
		if configDisabled {
			errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, errAPITestDisabled))
		}
		// do not check the [r.global.apiState] since tests may pass or fail based on different input data
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
