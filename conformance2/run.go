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
	"sort"
	"strings"
	"time"

	digest "github.com/opencontainers/go-digest"
)

const (
	testName  = "OCI Conformance Test"
	dataImage = "01-image"
	dataIndex = "02-index"
)

var errTestStatus = errors.New("API test tracked a failure")

var blobAPIs = []stateAPIType{stateAPIBlobPostPut, stateAPIBlobPostOnly}

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

	err = r.TestPush(r.Results, dataImage)
	if err != nil {
		errs = append(errs, err)
	}
	err = r.TestPush(r.Results, dataIndex)
	if err != nil {
		errs = append(errs, err)
	}
	// TODO: add tests for different types of data

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
		NumTotal:        r.Results.Counts[statusPass] + r.Results.Counts[statusFail] + r.Results.Counts[statusSkip],
		NumPassed:       r.Results.Counts[statusPass],
		NumFailed:       r.Results.Counts[statusFail],
		NumSkipped:      r.Results.Counts[statusSkip],
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
			r.TestSkip(res, err)
			return nil
		}
		if _, err := r.API.TagList(r.Config.schemeReg, r.Config.Repo1, apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, "", stateAPITagList)
			return fmt.Errorf("%.0w%w", errTestStatus, err)
		}
		r.TestPass(res, "", stateAPITagList)
		return nil
	})
}

func (r *runner) TestPush(parent *results, tdName string) error {
	// add more APIs
	return r.ChildRun("push", parent, func(r *runner, res *results) error {
		errs := []error{}
		curAPI := 0
		for dig := range r.State.Data[tdName].blobs {
			curAPI = (curAPI + 1) % len(blobAPIs)
			var err error
			switch blobAPIs[curAPI] {
			case stateAPIBlobPostPut:
				err = r.TestPushBlobPostPut(res, tdName, dig)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to push blob %s%.0w", dig.String(), err))
				}
			case stateAPIBlobPostOnly:
				err = r.TestPushBlobPostOnly(res, tdName, dig)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to push blob %s%.0w", dig.String(), err))
				}
			}
			// TODO: fallback to any blob push method
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

func (r *runner) TestPushBlobAny(parent *results, tdName string, dig digest.Digest) error {
	// TODO: try each API, preferring untested APIs, then APIs untested with a given digest algo, then default to preferred API
	// generate an ordered list of APIs to test
	// return on the first successful API
	// else return a join of all errors
	return fmt.Errorf("not implemented")
}

func (r *runner) TestPushBlobPostPut(parent *results, tdName string, dig digest.Digest) error {
	return r.ChildRun("blob-post-put", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPostPut); err != nil {
			r.TestSkip(res, err)
			return nil
		}
		if err := r.API.BlobPostPut(r.Config.schemeReg, r.Config.Repo1, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPostPut)
			return fmt.Errorf("%.0w%w", errTestStatus, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPostPut)
		return nil
	})
}

func (r *runner) TestPushBlobPostOnly(parent *results, tdName string, dig digest.Digest) error {
	return r.ChildRun("blob-post-only", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPostOnly); err != nil {
			r.TestSkip(res, err)
			return nil
		}
		if err := r.API.BlobPostOnly(r.Config.schemeReg, r.Config.Repo1, dig, r.State.Data[tdName], apiSaveOutput(res.Output)); err != nil {
			r.TestFail(res, err, tdName, stateAPIBlobPostOnly)
			return fmt.Errorf("%.0w%w", errTestStatus, err)
		}
		r.TestPass(res, tdName, stateAPIBlobPostOnly)
		return nil
	})
}

func (r *runner) TestPushManifest(parent *results, tdName string, dig digest.Digest) error {
	td := r.State.Data[tdName]
	opts := []apiDoOpt{}
	// if the referrers API is being tested, verify OCI-Subject header is returned when appropriate
	if r.Config.APIs.Referrer {
		subj := detectSubject(td.manifests[dig])
		if subj != nil {
			opts = append(opts, apiExpectHeader("OCI-Subject", subj.Digest.String()))
		}
	}
	if td.manOrder[len(td.manOrder)-1] == dig && td.tag != "" {
		// push by tag
		return r.ChildRun("manifest-by-tag", parent, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestPutTag); err != nil {
				r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
				r.TestSkip(res, err)
				return nil
			}
			opts = append(opts, apiSaveOutput(res.Output))
			if err := r.API.ManifestPut(r.Config.schemeReg, r.Config.Repo1, td.tag, dig, td, opts...); err != nil {
				r.TestFail(res, err, tdName, stateAPIManifestPutTag)
				return fmt.Errorf("%.0w%w", errTestStatus, err)
			}
			r.TestPass(res, tdName, stateAPIManifestPutTag)
			return nil
		})
	} else {
		// push by digest
		return r.ChildRun("manifest-by-digest", parent, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestPutDigest); err != nil {
				r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusSkip)
				r.TestSkip(res, err)
				return nil
			}
			opts = append(opts, apiSaveOutput(res.Output))
			if err := r.API.ManifestPut(r.Config.schemeReg, r.Config.Repo1, dig.String(), dig, td, opts...); err != nil {
				r.TestFail(res, err, tdName, stateAPIManifestPutDigest)
				return fmt.Errorf("%.0w%w", errTestStatus, err)
			}
			r.TestPass(res, tdName, stateAPIManifestPutDigest)
			return nil
		})
	}
}

func (r *runner) ChildRun(name string, parent *results, fn func(*runner, *results) error) error {
	res := resultsNew(name, parent)
	if parent != nil {
		parent.Children = append(parent.Children, res)
	}
	err := fn(r, res)
	res.Stop = time.Now()
	if err != nil && !errors.Is(err, errTestStatus) {
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

func (r *runner) TestSkip(res *results, err error) {
	s := statusSkip
	if errors.Is(err, ErrDisabled) {
		s = statusDisabled
	}
	res.Status = res.Status.Set(s)
	res.Counts[s]++
	fmt.Fprintf(res.Output, "%s: skipping test:\n  %s\n", res.Name,
		strings.ReplaceAll(err.Error(), "\n", "\n  "))
	r.Log.Info("skipping test", "name", res.Name, "error", err.Error())
}

func (r *runner) TestFail(res *results, err error, tdName string, apis ...stateAPIType) {
	res.Status = res.Status.Set(statusFail)
	res.Counts[statusFail]++
	res.Errs = append(res.Errs, err)
	if tdName != "" {
		r.State.DataStatus[tdName] = r.State.DataStatus[tdName].Set(statusFail)
	}
	for _, a := range apis {
		r.State.APIStatus[a] = r.State.APIStatus[a].Set(statusFail)
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
			stateAPIBlobPatchChunk, stateAPIBlobPatchStream, stateAPIBlobMountSource, stateAPIBlobMountAnonymous:
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
