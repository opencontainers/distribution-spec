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
	testName  = "OCI Conformance Test"
	dataImage = "01-image"
	dataIndex = "02-index"
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

func (r *runner) TestAll() error {
	errs := []error{}
	r.Results.Start = time.Now()

	err := r.GenerateData()
	if err != nil {
		return fmt.Errorf("aborting tests, unable to generate data: %w", err)
	}

	err = r.TestEmpty(r.Results)
	if err != nil {
		errs = append(errs, err)
	}

	for _, algo := range []digest.Algorithm{digest.SHA256, digest.SHA512} {
		err = r.TestBlobAPIs(r.Results, "blobs-"+algo.String(), algo)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// loop over different types of data
	for _, tdName := range []string{dataImage, dataIndex} {
		err = r.ChildRun(tdName, r.Results, func(r *runner, res *results) error {
			errs := []error{}
			// push content
			err := r.TestPush(res, tdName)
			if err != nil {
				errs = append(errs, err)
			}
			err = r.TestPull(res, tdName)
			if err != nil {
				errs = append(errs, err)
			}
			// TODO: add APIs to list/discover content

			// cleanup
			err = r.TestDelete(res, tdName)
			if err != nil {
				errs = append(errs, err)
			}
			r.State.Data[tdName].repo = ""

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

func (r *runner) GenerateData() error {
	// standard image with a layer per blob test
	tdName := dataImage
	r.State.DataStatus[tdName] = statusUnknown
	r.State.Data[tdName] = newTestData("OCI Image", "image")
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
	_ = mDig
	// multi-platform index
	tdName = dataIndex
	r.State.DataStatus[tdName] = statusUnknown
	r.State.Data[tdName] = newTestData("OCI Index", "index")
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
	_, _, err = r.State.Data[tdName].genIndex(platList, digImgList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}

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

func (r *runner) TestEmpty(parent *results) error {
	return r.ChildRun("empty", parent, func(r *runner, res *results) error {
		errs := []error{}
		if err := r.TestEmptyTagList(res); err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		return nil
	})
}

func (r *runner) TestEmptyTagList(parent *results) error {
	return r.ChildRun("tag list", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPITagList); err != nil {
			r.TestSkip(res, err, "", stateAPITagList)
			return nil
		}
		if _, err := r.API.TagList(r.Config.schemeReg, r.Config.Repo1, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, "", stateAPITagList)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, "", stateAPITagList)
		return nil
	})
}

func (r *runner) TestBlobAPIs(parent *results, tdName string, algo digest.Algorithm) error {
	return r.ChildRun(algo.String()+" blobs", parent, func(r *runner, res *results) error {
		errs := []error{}
		// setup testdata
		r.State.DataStatus[tdName] = statusUnknown
		r.State.Data[tdName] = newTestData(tdName, "")
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
			dig, _, err := r.State.Data[tdName].genBlob(algo, 512)
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
					err = r.TestPushBlobPostOnly(res, tdName, dig)
					if err != nil {
						errs = append(errs, err)
					}
				case "post+put":
					api = stateAPIBlobPostPut
					err = r.TestPushBlobPostPut(res, tdName, dig)
					if err != nil {
						errs = append(errs, err)
					}
				case "chunked single":
					api = stateAPIBlobPatchChunked
					// extract the min chunk length from a chunked push with a single chunk
					err = r.TestPushBlobPatchChunked(res, tdName, dig, apiReturnHeader("OCI-Chunk-Min-Length", &minHeader))
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
					dig, _, err = r.State.Data[tdName].genBlob(algo, minChunkSize*3-5)
					if err != nil {
						return fmt.Errorf("failed to generate chunked blob of size %d: %w", minChunkSize*3-5, err)
					}
					digests[testName] = dig
					err = r.TestPushBlobPatchChunked(res, tdName, dig)
					if err != nil {
						errs = append(errs, err)
					}
				case "stream":
					api = stateAPIBlobPatchStream
					err = r.TestPushBlobPatchStream(res, tdName, dig)
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
					err = r.TestPullBlob(res, tdName, dig)
					if err != nil {
						errs = append(errs, err)
					}
				}
				// cleanup
				err = r.TestDeleteBlob(res, tdName, dig)
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
				err := r.TestPushBlobAny(res, tdName, dig)
				if err != nil {
					errs = append(errs, err)
				}
				err = r.TestPullBlob(res, tdName, dig)
				if err != nil {
					errs = append(errs, err)
				}
				err = r.TestDeleteBlob(res, tdName, dig)
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

func (r *runner) TestPull(parent *results, tdName string) error {
	return r.ChildRun("pull", parent, func(r *runner, res *results) error {
		errs := []error{}
		for i, dig := range r.State.Data[tdName].manOrder {
			err := r.TestPullManifest(res, tdName, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to pull manifest %d, digest %s%.0w", i, dig.String(), err))
			}
		}
		for dig := range r.State.Data[tdName].blobs {
			err := r.TestPullBlob(res, tdName, dig)
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

func (r *runner) TestPullBlob(parent *results, tdName string, dig digest.Digest) error {
	return r.ChildRun("blob-get", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobGetFull); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobGetFull)
			return nil
		}
		if err := r.API.BlobGetFull(r.Config.schemeReg, r.Config.Repo1, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobGetFull)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobGetFull)
		return nil
	})
}

func (r *runner) TestPullManifest(parent *results, tdName string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	opts := []apiDoOpt{}
	apis := []stateAPIType{}
	errs := []error{}
	if td.manOrder[len(td.manOrder)-1] == dig && td.tag != "" {
		// pull by tag
		err := r.ChildRun("manifest-by-tag", parent, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestGetTag); err != nil {
				r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
				r.TestSkip(res, err, tdName, stateAPIManifestGetTag)
				return nil
			}
			apis = append(apis, stateAPIManifestGetTag)
			opts = append(opts, apiSaveOutput(res.Output))
			if err := r.API.ManifestGet(r.Config.schemeReg, r.Config.Repo1, td.tag, dig, td, opts...); err != nil {
				r.TestFail(res, err, tdName, apis...)
				return fmt.Errorf("%.0w%w", errTestAPIFail, err)
			}
			r.TestPass(res, tdName, apis...)
			return nil
		})
		if err != nil {
			errs = append(errs, err)
		}
	}
	// push by digest
	err := r.ChildRun("manifest-by-digest", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestGetDigest); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestGetDigest)
			return nil
		}
		apis = append(apis, stateAPIManifestGetDigest)
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.ManifestGet(r.Config.schemeReg, r.Config.Repo1, dig.String(), dig, td, opts...); err != nil {
			r.TestFail(res, err, tdName, apis...)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, apis...)
		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (r *runner) TestPush(parent *results, tdName string) error {
	return r.ChildRun("push", parent, func(r *runner, res *results) error {
		errs := []error{}
		for dig := range r.State.Data[tdName].blobs {
			err := r.TestPushBlobAny(res, tdName, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to push blob %s%.0w", dig.String(), err))
			}
		}
		for i, dig := range r.State.Data[tdName].manOrder {
			err := r.TestPushManifest(res, tdName, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to push manifest %d, digest %s%.0w", i, dig.String(), err))
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

func (r *runner) TestPushBlobAny(parent *results, tdName string, dig digest.Digest) error {
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
			err = r.TestPushBlobPostPut(parent, tdName, dig)
		case stateAPIBlobPostOnly:
			err = r.TestPushBlobPostOnly(parent, tdName, dig)
		case stateAPIBlobPatchStream:
			err = r.TestPushBlobPatchStream(parent, tdName, dig)
		case stateAPIBlobPatchChunked:
			err = r.TestPushBlobPatchChunked(parent, tdName, dig)
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

func (r *runner) TestPushBlobPostPut(parent *results, tdName string, dig digest.Digest) error {
	return r.ChildRun("blob-post-put", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPostPut); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPostPut)
			return nil
		}
		if err := r.API.BlobPostPut(r.Config.schemeReg, r.Config.Repo1, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPostPut)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPostPut, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushBlobPostOnly(parent *results, tdName string, dig digest.Digest) error {
	return r.ChildRun("blob-post-only", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPostOnly); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPostOnly)
			return nil
		}
		if err := r.API.BlobPostOnly(r.Config.schemeReg, r.Config.Repo1, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPostOnly)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPostOnly, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushBlobPatchChunked(parent *results, tdName string, dig digest.Digest, opts ...apiDoOpt) error {
	return r.ChildRun("blob-patch-chunked", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPatchChunked); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPatchChunked)
			return nil
		}
		opts = append(opts, apiSaveOutput(res.Output))
		if err := r.API.BlobPatchChunked(r.Config.schemeReg, r.Config.Repo1, dig, r.State.Data[tdName], opts...); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPatchChunked)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPatchChunked, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushBlobPatchStream(parent *results, tdName string, dig digest.Digest) error {
	return r.ChildRun("blob-patch-stream", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPatchStream); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobPatchStream)
			return nil
		}
		if err := r.API.BlobPatchStream(r.Config.schemeReg, r.Config.Repo1, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPatchStream)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPatchStream, stateAPIBlobPush)
		return nil
	})
}

func (r *runner) TestPushManifest(parent *results, tdName string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	opts := []apiDoOpt{}
	apis := []stateAPIType{}
	// if the referrers API is being tested, verify OCI-Subject header is returned when appropriate
	if r.Config.APIs.Referrer {
		subj := detectSubject(td.manifests[dig])
		if subj != nil {
			opts = append(opts, apiExpectHeader("OCI-Subject", subj.Digest.String()))
			apis = append(apis, stateAPIManifestPutSubject)
		}
	}
	if td.manOrder[len(td.manOrder)-1] == dig && td.tag != "" {
		// push by tag
		return r.ChildRun("manifest-by-tag", parent, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestPutTag); err != nil {
				r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
				r.TestSkip(res, err, tdName, stateAPIManifestPutTag)
				return nil
			}
			apis = append(apis, stateAPIManifestPutTag)
			opts = append(opts, apiSaveOutput(res.Output))
			if err := r.API.ManifestPut(r.Config.schemeReg, r.Config.Repo1, td.tag, dig, td, opts...); err != nil {
				r.TestFail(res, err, tdName, apis...)
				return fmt.Errorf("%.0w%w", errTestAPIFail, err)
			}
			r.TestPass(res, tdName, apis...)
			return nil
		})
	} else {
		// push by digest
		return r.ChildRun("manifest-by-digest", parent, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestPutDigest); err != nil {
				r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
				r.TestSkip(res, err, tdName, stateAPIManifestPutDigest)
				return nil
			}
			apis = append(apis, stateAPIManifestPutDigest)
			opts = append(opts, apiSaveOutput(res.Output))
			if err := r.API.ManifestPut(r.Config.schemeReg, r.Config.Repo1, dig.String(), dig, td, opts...); err != nil {
				r.TestFail(res, err, tdName, apis...)
				return fmt.Errorf("%.0w%w", errTestAPIFail, err)
			}
			r.TestPass(res, tdName, apis...)
			return nil
		})
	}
}

func (r *runner) TestDelete(parent *results, tdName string) error {
	return r.ChildRun("delete", parent, func(r *runner, res *results) error {
		errs := []error{}
		delOrder := slices.Clone(r.State.Data[tdName].manOrder)
		slices.Reverse(delOrder)
		for i, dig := range delOrder {
			err := r.TestDeleteManifest(res, tdName, dig)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to delete manifest %d, digest %s%.0w", i, dig.String(), err))
			}
		}
		for dig := range r.State.Data[tdName].blobs {
			err := r.TestDeleteBlob(res, tdName, dig)
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

func (r *runner) TestDeleteManifest(parent *results, tdName string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	errs := []error{}
	if td.manOrder[len(td.manOrder)-1] == dig && td.tag != "" {
		// delete tag
		err := r.ChildRun("manifest-by-tag", parent, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestDeleteTag); err != nil {
				r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
				r.TestSkip(res, err, tdName, stateAPIManifestDeleteTag)
				return nil
			}
			if err := r.API.ManifestDelete(r.Config.schemeReg, td.repo, td.tag, dig, td, apiSaveOutput(res.Output)); err != nil {
				r.TestFail(res, err, tdName, stateAPIManifestDeleteTag)
				return fmt.Errorf("%.0w%w", errTestAPIFail, err)
			}
			r.TestPass(res, tdName, stateAPIManifestDeleteTag)
			return nil
		})
		if err != nil {
			errs = append(errs, err)
		}
	}
	// delete digest
	err := r.ChildRun("manifest-by-digest", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIManifestDeleteDigest); err != nil {
			r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
			r.TestSkip(res, err, tdName, stateAPIManifestDeleteDigest)
			return nil
		}
		if err := r.API.ManifestDelete(r.Config.schemeReg, td.repo, dig.String(), dig, td, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIManifestDeleteDigest)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIManifestDeleteDigest)
		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (r *runner) TestDeleteBlob(parent *results, tdName string, dig digest.Digest) error {
	return r.ChildRun("blob-delete", parent, func(r *runner, res *results) error {
		td := r.State.Data[tdName]
		if err := r.APIRequire(stateAPIBlobDelete); err != nil {
			r.TestSkip(res, err, tdName, stateAPIBlobDelete)
			return nil
		}
		if err := r.API.BlobDelete(r.Config.schemeReg, td.repo, dig, td, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobDelete)
			return fmt.Errorf("%.0w%w", errTestAPIFail, err)
		}
		r.TestPass(res, tdName, stateAPIBlobDelete, stateAPIBlobPush)
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
	r.Log.Warn("failed test", "name", res.Name, "error", err.Error())
	r.Log.Debug("failed test output", "name", res.Name, "output", res.Output.String())
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
