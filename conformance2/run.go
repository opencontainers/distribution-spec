package main

import (
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
)

const (
	testName          = "OCI Conformance Test"
	dataImage         = "01-image"
	dataIndex         = "02-index"
	dataArtifact      = "03-artifact"
	dataArtifactIndex = "04-artifact-index"
)

var errTestAPIFail = errors.New("API test with a known failure")

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
	// standard image with a layer per blob test
	tdName := dataImage
	r.State.DataStatus[tdName] = statusUnknown
	r.State.Data[tdName] = newTestData("OCI Image")
	digCList := []digest.Digest{}
	digUCList := []digest.Digest{}
	for l := range blobAPIs {
		digC, digUC, _, err := r.State.Data[tdName].genLayer(l)
		if err != nil {
			return fmt.Errorf("failed to generate test data layer %d: %w", l, err)
		}
		digCList = append(digCList, digC)
		digUCList = append(digUCList, digUC)
	}
	cDig, _, err := r.State.Data[tdName].genConfig(platform{OS: "linux", Architecture: "amd64"}, digUCList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	mDig, _, err := r.State.Data[tdName].genManifest(cDig, digCList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	r.State.Data[tdName].tags["image"] = mDig
	// multi-platform index
	tdName = dataIndex
	r.State.DataStatus[tdName] = statusUnknown
	r.State.Data[tdName] = newTestData("OCI Index")
	platList := []*platform{
		{OS: "linux", Architecture: "amd64"},
		{OS: "linux", Architecture: "arm64"},
	}
	digImgList := []digest.Digest{}
	for _, p := range platList {
		digCList = []digest.Digest{}
		digUCList = []digest.Digest{}
		for l := range blobAPIs {
			digC, digUC, _, err := r.State.Data[tdName].genLayer(l)
			if err != nil {
				return fmt.Errorf("failed to generate test data layer %d: %w", l, err)
			}
			digCList = append(digCList, digC)
			digUCList = append(digUCList, digUC)
		}
		cDig, _, err := r.State.Data[tdName].genConfig(*p, digUCList)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		mDig, _, err := r.State.Data[tdName].genManifest(cDig, digCList)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		digImgList = append(digImgList, mDig)
	}
	iDig, _, err := r.State.Data[tdName].genIndex(platList, digImgList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	r.State.Data[tdName].tags["index"] = iDig
	// two artifacts with subject image
	tdName = dataArtifact
	r.State.DataStatus[tdName] = statusUnknown
	r.State.Data[tdName] = newTestData("OCI Artifact")
	digCList = []digest.Digest{}
	digUCList = []digest.Digest{}
	for l := range blobAPIs {
		digC, digUC, _, err := r.State.Data[tdName].genLayer(l)
		if err != nil {
			return fmt.Errorf("failed to generate test data layer %d: %w", l, err)
		}
		digCList = append(digCList, digC)
		digUCList = append(digUCList, digUC)
	}
	cDig, _, err = r.State.Data[tdName].genConfig(platform{OS: "linux", Architecture: "amd64"}, digUCList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	subjDig, _, err := r.State.Data[tdName].genManifest(cDig, digCList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	r.State.Data[tdName].tags["subject"] = subjDig
	abDig, _, err := r.State.Data[tdName].genBlob(2048)
	if err != nil {
		return fmt.Errorf("failed to generate test artifact blob: %w", err)
	}
	emptyConfDig, err := r.State.Data[tdName].addBlob([]byte("{}"), genWithMediaType("application/vnd.oci.empty.v1+json"), genSetData())
	if err != nil {
		return fmt.Errorf("failed to add test artifact config: %w", err)
	}
	aDig1, _, err := r.State.Data[tdName].genManifest(emptyConfDig, []digest.Digest{abDig},
		genWithArtifactType("application/vnd.example.oci.conformance"),
		genWithSubject(*r.State.Data[tdName].desc[subjDig]))
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	r.State.Data[tdName].tags["artifact1"] = aDig1
	abDig, _, err = r.State.Data[tdName].genBlob(2048)
	if err != nil {
		return fmt.Errorf("failed to generate test artifact blob: %w", err)
	}
	aDig2, _, err := r.State.Data[tdName].genManifest(emptyConfDig, []digest.Digest{abDig},
		genWithArtifactType("application/vnd.example.oci.conformance"),
		genWithSubject(*r.State.Data[tdName].desc[subjDig]))
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	r.State.Data[tdName].tags["artifact2"] = aDig2
	// artifact packaged as an index with subject image
	tdName = dataArtifactIndex
	r.State.DataStatus[tdName] = statusUnknown
	r.State.Data[tdName] = newTestData("OCI Artifact Index")
	digCList = []digest.Digest{}
	digUCList = []digest.Digest{}
	for l := range blobAPIs {
		digC, digUC, _, err := r.State.Data[tdName].genLayer(l)
		if err != nil {
			return fmt.Errorf("failed to generate test data layer %d: %w", l, err)
		}
		digCList = append(digCList, digC)
		digUCList = append(digUCList, digUC)
	}
	cDig, _, err = r.State.Data[tdName].genConfig(platform{OS: "linux", Architecture: "amd64"}, digUCList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	subjDig, _, err = r.State.Data[tdName].genManifest(cDig, digCList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	r.State.Data[tdName].tags["subject"] = subjDig
	platList = []*platform{
		{OS: "linux", Architecture: "amd64"},
		{OS: "linux", Architecture: "arm64"},
	}
	digImgList = []digest.Digest{}
	emptyConfDig, err = r.State.Data[tdName].addBlob([]byte("{}"), genWithMediaType("application/vnd.oci.empty.v1+json"), genSetData())
	if err != nil {
		return fmt.Errorf("failed to add test artifact config: %w", err)
	}
	for range platList {
		abDig, _, err := r.State.Data[tdName].genBlob(2048)
		if err != nil {
			return fmt.Errorf("failed to generate test artifact blob: %w", err)
		}
		aDig, _, err := r.State.Data[tdName].genManifest(emptyConfDig, []digest.Digest{abDig},
			genWithArtifactType("application/vnd.example.oci.conformance.image"),
		)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		digImgList = append(digImgList, aDig)
	}
	aDig, _, err := r.State.Data[tdName].genIndex(platList, digImgList,
		genWithArtifactType("application/vnd.example.oci.conformance.index"),
		genWithSubject(*r.State.Data[tdName].desc[subjDig]))
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	r.State.Data[tdName].tags["artifact"] = aDig
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
		NumTotal:        r.Results.Counts[statusPass] + r.Results.Counts[statusFail] + r.Results.Counts[statusSkip] + r.Results.Counts[statusDisabled],
		NumPassed:       r.Results.Counts[statusPass],
		NumFailed:       r.Results.Counts[statusFail],
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

	err := r.GenerateData()
	if err != nil {
		return fmt.Errorf("aborting tests, unable to generate data: %w", err)
	}

	err = r.TestEmpty(r.Results, repo)
	if err != nil {
		errs = append(errs, err)
	}

	for _, algo := range []digest.Algorithm{digest.SHA256, digest.SHA512} {
		err = r.TestBlobAPIs(r.Results, "blobs-"+algo.String(), algo, repo)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// loop over different types of data
	for _, tdName := range []string{dataImage, dataIndex, dataArtifact, dataArtifactIndex} {
		err = r.ChildRun(tdName, r.Results, func(r *runner, res *results) error {
			errs := []error{}
			// push content
			err := r.TestPush(res, tdName, repo)
			if err != nil {
				errs = append(errs, err)
			}
			// TODO: add head requests
			err = r.TestPull(res, tdName, repo)
			if err != nil {
				errs = append(errs, err)
			}
			// TODO: add APIs to list/discover content
			err = r.TestReferrers(res, tdName, repo)
			if err != nil {
				errs = append(errs, err)
			}

			// cleanup
			err = r.TestDelete(res, tdName, repo)
			if err != nil {
				errs = append(errs, err)
			}

			// TODO: verify tag listing show tag was deleted if delete API enabled

			return errors.Join(errs...)
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	r.Results.Stop = time.Now()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (r *runner) TestDelete(parent *results, tdName string, repo string) error {
	return r.ChildRun("delete", parent, func(r *runner, res *results) error {
		errs := []error{}
		delOrder := slices.Clone(r.State.Data[tdName].manOrder)
		slices.Reverse(delOrder)
		for tag, dig := range r.State.Data[tdName].tags {
			err := r.TestDeleteManifestTag(res, tdName, repo, tag, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to delete manifest tag %s%.0w", tag, err))
			}
		}
		for i, dig := range delOrder {
			err := r.TestDeleteManifestDigest(res, tdName, repo, dig)
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

func (r *runner) TestDeleteManifestDigest(parent *results, tdName string, repo string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	return r.ChildRun("manifest-by-digest", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestDeleteDigest); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestDeleteDigest)
			return nil
		}
		if err := r.API.ManifestDelete(r.Config.schemeReg, repo, dig.String(), dig, td, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIManifestDeleteDigest)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIManifestDeleteDigest)
		return nil
	})
}

func (r *runner) TestDeleteManifestTag(parent *results, tdName string, repo string, tag string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	return r.ChildRun("manifest-by-tag", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestDeleteTag); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestDeleteTag)
			return nil
		}
		if err := r.API.ManifestDelete(r.Config.schemeReg, repo, tag, dig, td, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIManifestDeleteTag)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIManifestDeleteTag)
		return nil
	})
}

func (r *runner) TestDeleteBlob(parent *results, tdName string, repo string, dig digest.Digest) error {
	return r.ChildRun("blob-delete", parent, func(r *runner, res *results) error {
		td := r.State.Data[tdName]
		if err := r.APIRequire(stateAPIBlobDelete); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobDelete)
			return nil
		}
		if err := r.API.BlobDelete(r.Config.schemeReg, repo, dig, td, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobDelete)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobDelete, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestEmpty(parent *results, repo string) error {
	return r.ChildRun("empty", parent, func(r *runner, res *results) error {
		errs := []error{}
		if err := r.TestEmptyTagList(res, repo); err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		return nil
	})
}

func (r *runner) TestEmptyTagList(parent *results, repo string) error {
	return r.ChildRun("tag list", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPITagList); err != nil {
			r.TestSkip(res, err, "", stateAPITagList)
			return nil
		}
		if _, err := r.API.TagList(r.Config.schemeReg, repo, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, "", stateAPITagList)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, "", stateAPITagList)
		return nil
	})
}

func (r *runner) TestBlobAPIs(parent *results, tdName string, algo digest.Algorithm, repo string) error {
	return r.ChildRun(algo.String()+" blobs", parent, func(r *runner, res *results) error {
		errs := []error{}
		// setup testdata
		r.State.DataStatus[tdName] = statusUnknown
		r.State.Data[tdName] = newTestData(tdName)
		digests := map[string]digest.Digest{}
		testBlobs := map[string][]byte{
			"empty":     []byte(""),
			"emptyJSON": []byte("{}"),
		}
		dataTests := []string{"empty", "emptyJSON"}
		for name, val := range testBlobs {
			dig := algo.FromBytes(val)
			digests[name] = dig
			r.State.Data[tdName].blobs[dig] = val
		}
		// test the various blob push APIs
		if _, ok := blobAPIsTestedByAlgo[algo]; !ok {
			blobAPIsTestedByAlgo[algo] = &[stateAPIMax]bool{}
		}
		apiTests := []string{"post only", "post+put", "chunked single", "stream"}
		for _, name := range apiTests {
			dig, _, err := r.State.Data[tdName].genBlob(512, genWithAlgo(algo))
			if err != nil {
				return fmt.Errorf("failed to generate blob: %w", err)
			}
			digests[name] = dig
		}
		apiTests = append(apiTests, "chunked multi")
		minChunkSize := int64(chunkMin)
		minHeader := ""
		for _, testName := range apiTests {
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
					dig, _, err = r.State.Data[tdName].genBlob(minChunkSize*3-5, genWithAlgo(algo))
					if err != nil {
						return fmt.Errorf("failed to generate chunked blob of size %d: %w", minChunkSize*3-5, err)
					}
					digests[testName] = dig
					err = r.TestPushBlobPatchChunked(res, tdName, repo, dig)
					if err != nil {
						errs = append(errs, err)
					}
				case "stream":
					api = stateAPIBlobPatchStream
					err = r.TestPushBlobPatchStream(res, tdName, repo, dig)
					if err != nil {
						errs = append(errs, err)
					}
				default:
					return fmt.Errorf("unknown api test %s", testName)
				}
				// track the used APIs so TestPushBlobAny doesn't rerun tests
				blobAPIsTested[api] = true
				blobAPIsTestedByAlgo[dig.Algorithm()][api] = true
				if err == nil {
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
				return errors.Join(errs...)
			})
			if err != nil {
				errs = append(errs, err)
			}
		}
		// test various well known blob contents
		for _, name := range dataTests {
			err := r.ChildRun(name, res, func(r *runner, res *results) error {
				dig := digests[name]
				err := r.TestPushBlobAny(res, tdName, repo, dig)
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

		// TODO
		// cross repository blob mount
		// anonymous blob mount
		// test pull command on pushed blobs
		// cleanup

		return errors.Join(errs...)
	})
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
			return nil
		}
		if err := r.API.BlobGetFull(r.Config.schemeReg, repo, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobGetFull)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
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
			return nil
		}
		apis = append(apis, stateAPIManifestGetDigest)
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.ManifestGet(r.Config.schemeReg, repo, dig.String(), dig, td, opts...); err != nil {
			r.TestFail(res, err, tdName, apis...)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
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
			return nil
		}
		apis = append(apis, stateAPIManifestGetTag)
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.ManifestGet(r.Config.schemeReg, repo, tag, dig, td, opts...); err != nil {
			r.TestFail(res, err, tdName, apis...)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
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

func (r *runner) TestPushBlobAny(parent *results, tdName string, repo string, dig digest.Digest) error {
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
		err := errors.New("not implemented")
		switch api {
		case stateAPIBlobPostPut:
			err = r.TestPushBlobPostPut(parent, tdName, repo, dig)
		case stateAPIBlobPostOnly:
			err = r.TestPushBlobPostOnly(parent, tdName, repo, dig)
		case stateAPIBlobPatchStream:
			err = r.TestPushBlobPatchStream(parent, tdName, repo, dig)
		case stateAPIBlobPatchChunked:
			err = r.TestPushBlobPatchChunked(parent, tdName, repo, dig)
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

func (r *runner) TestPushBlobPostPut(parent *results, tdName string, repo string, dig digest.Digest) error {
	return r.ChildRun("blob-post-put", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPostPut); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPostPut)
			return nil
		}
		if err := r.API.BlobPostPut(r.Config.schemeReg, repo, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPostPut)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPostPut, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushBlobPostOnly(parent *results, tdName string, repo string, dig digest.Digest) error {
	return r.ChildRun("blob-post-only", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPostOnly); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPostOnly)
			return nil
		}
		if err := r.API.BlobPostOnly(r.Config.schemeReg, repo, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPostOnly)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPostOnly, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushBlobPatchChunked(parent *results, tdName string, repo string, dig digest.Digest, opts ...apiDoOpt) error {
	return r.ChildRun("blob-patch-chunked", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPatchChunked); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPatchChunked)
			return nil
		}
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.BlobPatchChunked(r.Config.schemeReg, repo, dig, r.State.Data[tdName], opts...); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPatchChunked)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPatchChunked, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushBlobPatchStream(parent *results, tdName string, repo string, dig digest.Digest) error {
	return r.ChildRun("blob-patch-stream", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPatchStream); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPatchStream)
			return nil
		}
		if err := r.API.BlobPatchStream(r.Config.schemeReg, repo, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPatchStream)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPatchStream, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushManifestDigest(parent *results, tdName string, repo string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	opts := []apiDoOpt{}
	apis := []stateAPIType{}
	// if the referrers API is being tested, verify OCI-Subject header is returned when appropriate
	subj := detectSubject(td.manifests[dig])
	if subj != nil {
		apis = append(apis, stateAPIManifestPutSubject)
		if r.Config.APIs.Referrer {
			opts = append(opts, apiExpectHeader("OCI-Subject", subj.Digest.String()))
		}
	}
	return r.ChildRun("manifest-by-digest", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestPutDigest); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestPutDigest)
			return nil
		}
		apis = append(apis, stateAPIManifestPutDigest)
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.ManifestPut(r.Config.schemeReg, repo, dig.String(), dig, td, opts...); err != nil {
			r.TestFail(res, err, tdName, apis...)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, apis...)
		return nil
	})
}

func (r *runner) TestPushManifestTag(parent *results, tdName string, repo string, tag string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	opts := []apiDoOpt{}
	apis := []stateAPIType{}
	// if the referrers API is being tested, verify OCI-Subject header is returned when appropriate
	subj := detectSubject(td.manifests[dig])
	if subj != nil {
		apis = append(apis, stateAPIManifestPutSubject)
		if r.Config.APIs.Referrer {
			opts = append(opts, apiExpectHeader("OCI-Subject", subj.Digest.String()))
		}
	}
	return r.ChildRun("manifest-by-tag", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestPutTag); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestPutTag)
			return nil
		}
		apis = append(apis, stateAPIManifestPutTag)
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.ManifestPut(r.Config.schemeReg, repo, tag, dig, td, opts...); err != nil {
			r.TestFail(res, err, tdName, apis...)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
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
		for subj, referrerList := range r.State.Data[tdName].referrers {
			if err := r.APIRequire(stateAPIReferrers); err != nil {
				r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
				r.TestSkip(res, err, tdName, stateAPIReferrers)
				return nil
			}
			referrerResp, err := r.API.ReferrersList(r.Config.schemeReg, repo, subj, apiSaveOutput(res.Output))
			if err != nil {
				errs = append(errs, err)
			}
			if err == nil {
				for _, dig := range referrerList {
					if !slices.ContainsFunc(referrerResp.Manifests, func(desc descriptor) bool { return desc.Digest == dig }) {
						errs = append(errs, fmt.Errorf("entry missing from referrers list, subject %s, referrer %s", subj, dig))
					}
				}
			}
		}
		if len(errs) > 0 {
			r.TestFail(res, errors.Join(errs...), tdName, stateAPIReferrers)
			return fmt.Errorf("%.0w%w", errTestAPIFail, errors.Join(errs...))
		}
		r.TestPass(res, tdName, stateAPIReferrers)
		return nil
	})
}

func (r *runner) ChildRun(name string, parent *results, fn func(*runner, *results) error) error {
	res := resultsNew(name, parent)
	if parent != nil {
		parent.Children = append(parent.Children, res)
	}
	err := fn(r, res)
	res.Stop = time.Now()
	if err != nil && !errors.Is(err, errTestAPIFail) {
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
	if errors.Is(err, ErrDisabled) {
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
	if errors.Is(err, ErrRegUnsupported) {
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
		switch a {
		case stateAPITagList:
			if !r.Config.APIs.Tags {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIManifestPutTag, stateAPIManifestPutDigest, stateAPIManifestPutSubject,
			stateAPIBlobPush, stateAPIBlobPostOnly, stateAPIBlobPostPut,
			stateAPIBlobPatchChunked, stateAPIBlobPatchStream, stateAPIBlobMountSource, stateAPIBlobMountAnonymous:
			if !r.Config.APIs.Push {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIManifestGetTag, stateAPIManifestGetDigest, stateAPIBlobGetFull, stateAPIBlobGetRange:
			if !r.Config.APIs.Pull {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIBlobDelete:
			if !r.Config.APIs.Delete.Blob {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIManifestDeleteTag:
			if !r.Config.APIs.Delete.Tag {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIManifestDeleteDigest:
			if !r.Config.APIs.Delete.Manifest {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIReferrers:
			if !r.Config.APIs.Referrer {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		}
		// do not check the [r.global.apiState] since tests may pass or fail based on different input data
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
