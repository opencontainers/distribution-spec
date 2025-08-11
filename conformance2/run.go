package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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

var blobAPIs = []stateAPIType{stateAPIBlobPostPut, stateAPIBlobPostOnly}

type runner struct {
	config  config
	api     *api
	state   *state
	results *results
}

func runnerNew(c config) (*runner, error) {
	lvl := slog.LevelWarn
	if c.LogLevel != "" {
		err := lvl.UnmarshalText([]byte(c.LogLevel))
		if err != nil {
			return nil, fmt.Errorf("failed to parse logging level %s: %w", c.LogLevel, err)
		}
	}
	r := runner{
		config:  c,
		api:     apiNew(http.DefaultClient),
		state:   stateNew(),
		results: resultsNew(testName, nil),
	}
	return &r, nil
}

func (r *runner) TestAll() error {
	errs := []error{}
	r.results.Start()

	err := r.GenerateData()
	if err != nil {
		return fmt.Errorf("aborting tests, unable to generate data: %w", err)
	}

	err = r.TestEmpty(r.results)
	if err != nil {
		errs = append(errs, err)
	}

	err = r.TestPush(r.results, dataImage)
	if err != nil {
		errs = append(errs, err)
	}
	err = r.TestPush(r.results, dataIndex)
	if err != nil {
		errs = append(errs, err)
	}
	// TODO: add tests for different types of data

	r.results.stop = time.Now()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (r *runner) GenerateData() error {
	// standard image with a layer per blob test
	tdName := dataImage
	r.state.dataStatus[tdName] = statusUnknown
	r.state.data[tdName] = newTestData("OCI Image", "image")
	digCList := []digest.Digest{}
	digUCList := []digest.Digest{}
	for l := 0; l < len(blobAPIs); l++ {
		digC, digUC, _, err := r.state.data[tdName].genLayer(l)
		if err != nil {
			return fmt.Errorf("failed to generate test data layer %d: %w", l, err)
		}
		digCList = append(digCList, digC)
		digUCList = append(digUCList, digUC)
	}
	cDig, _, err := r.state.data[tdName].genConfig(platform{OS: "linux", Architecture: "amd64"}, digUCList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	mDig, _, err := r.state.data[tdName].genManifest(cDig, digCList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	_ = mDig
	// multi-platform index
	tdName = dataIndex
	r.state.dataStatus[tdName] = statusUnknown
	r.state.data[tdName] = newTestData("OCI Index", "index")
	platList := []*platform{
		{OS: "linux", Architecture: "amd64"},
		{OS: "linux", Architecture: "arm64"},
	}
	digImgList := []digest.Digest{}
	for _, p := range platList {
		digCList = []digest.Digest{}
		digUCList = []digest.Digest{}
		for l := 0; l < len(blobAPIs); l++ {
			digC, digUC, _, err := r.state.data[tdName].genLayer(l)
			if err != nil {
				return fmt.Errorf("failed to generate test data layer %d: %w", l, err)
			}
			digCList = append(digCList, digC)
			digUCList = append(digUCList, digUC)
		}
		cDig, _, err := r.state.data[tdName].genConfig(*p, digUCList)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		mDig, _, err := r.state.data[tdName].genManifest(cDig, digCList)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		digImgList = append(digImgList, mDig)
	}
	_, _, err = r.state.data[tdName].genIndex(platList, digImgList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}

	return nil
}

func (r *runner) Report(w io.Writer) {
	fmt.Fprintf(w, "Test results\n")
	r.results.ReportWalkErr(w, "")
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "OCI Conformance Result: %s\n", r.results.status.String())
	padWidth := 30

	statusTotal := 0
	for i := status(1); i < statusMax; i++ {
		pad := ""
		if len(i.String()) < padWidth {
			pad = strings.Repeat(".", padWidth-len(i.String()))
		}
		fmt.Fprintf(w, "  %s%s: %10d\n", i.String(), pad, r.results.counts[i])
		statusTotal += r.results.counts[i]
	}
	pad := strings.Repeat(".", padWidth-len("Total"))
	fmt.Fprintf(w, "  %s%s: %10d\n\n", "Total", pad, statusTotal)

	if len(r.results.errs) > 0 {
		fmt.Fprintf(w, "Errors:\n%s\n\n", errors.Join(r.results.errs...))
	}

	fmt.Fprintf(w, "API conformance:\n")
	for i := stateAPIType(0); i < stateAPIMax; i++ {
		pad := ""
		if len(i.String()) < padWidth {
			pad = strings.Repeat(".", padWidth-len(i.String()))
		}
		fmt.Fprintf(w, "  %s%s: %10s\n", i.String(), pad, r.state.apiStatus[i].String())
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "Data conformance:\n")
	tdNames := []string{}
	for tdName := range r.state.data {
		tdNames = append(tdNames, tdName)
	}
	sort.Strings(tdNames)
	for _, tdName := range tdNames {
		pad := ""
		if len(r.state.data[tdName].name) < padWidth {
			pad = strings.Repeat(".", padWidth-len(r.state.data[tdName].name))
		}
		fmt.Fprintf(w, "  %s%s: %10s\n", r.state.data[tdName].name, pad, r.state.dataStatus[tdName].String())
	}
	fmt.Fprintf(w, "\n")
	// TODO: include config
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
			r.Skip(res, err)
			return nil
		}
		if _, err := r.api.TagList(r.config.schemeReg, r.config.Repo1); err != nil {
			r.APIFail(res, err, stateAPITagList)
		} else {
			r.APIPass(res, stateAPITagList)
		}
		return nil
	})
}

func (r *runner) TestPush(parent *results, tdName string) error {
	// add more APIs
	return r.ChildRun("push", parent, func(r *runner, res *results) error {
		errs := []error{}
		curAPI := 0
		for dig := range r.state.data[tdName].blobs {
			curAPI = (curAPI + 1) % len(blobAPIs)
			var err error
			switch blobAPIs[curAPI] {
			case stateAPIBlobPostPut:
				err = r.TestPushBlobPostPut(res, tdName, dig)
				if err != nil {
					errs = append(errs, err)
				}
			case stateAPIBlobPostOnly:
				err = r.TestPushBlobPostOnly(res, tdName, dig)
				if err != nil {
					errs = append(errs, err)
				}
			}
			// TODO: fallback to any blob push method
		}
		for _, dig := range r.state.data[tdName].manOrder {
			err := r.TestPushManifest(res, tdName, dig)
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			r.state.dataStatus[tdName] = r.state.dataStatus[tdName].Set(statusFail)
			return errors.Join(errs...)
		}
		r.state.dataStatus[tdName] = r.state.dataStatus[tdName].Set(statusPass)
		return nil
	})
}

func (r *runner) TestPushBlobPostPut(parent *results, tdName string, dig digest.Digest) error {
	return r.ChildRun("blob-post-put", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPostPut); err != nil {
			r.Skip(res, err)
			return nil
		}
		if err := r.api.BlobPostPut(r.config.schemeReg, r.config.Repo1, dig, r.state.data[tdName]); err != nil {
			r.APIFail(res, err, stateAPIBlobPostPut)
			return nil
		}
		r.APIPass(res, stateAPIBlobPostPut)
		return nil
	})
}

func (r *runner) TestPushBlobPostOnly(parent *results, tdName string, dig digest.Digest) error {
	return r.ChildRun("blob-post-only", parent, func(r *runner, res *results) error {
		if err := r.APIRequire(stateAPIBlobPostOnly); err != nil {
			r.Skip(res, err)
			return nil
		}
		if err := r.api.BlobPostOnly(r.config.schemeReg, r.config.Repo1, dig, r.state.data[tdName]); err != nil {
			r.APIFail(res, err, stateAPIBlobPostOnly)
			return nil
		}
		r.APIPass(res, stateAPIBlobPostOnly)
		return nil
	})
}

func (r *runner) TestPushManifest(parent *results, tdName string, dig digest.Digest) error {
	td := r.state.data[tdName]
	if td.manOrder[len(td.manOrder)-1] == dig && td.tag != "" {
		// push by tag
		return r.ChildRun("manifest-by-tag", parent, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestPutTag); err != nil {
				r.state.dataStatus[tdName] = r.state.dataStatus[tdName].Set(statusSkip)
				r.Skip(res, err)
				return nil
			}
			if err := r.api.ManifestPut(r.config.schemeReg, r.config.Repo1, td.tag, dig, td); err != nil {
				r.APIFail(res, err, stateAPIManifestPutTag)
				return nil
			}
			r.APIPass(res, stateAPIManifestPutTag)
			return nil
		})
	} else {
		// push by digest
		return r.ChildRun("manifest-by-digest", parent, func(r *runner, res *results) error {
			if err := r.APIRequire(stateAPIManifestPutDigest); err != nil {
				r.state.dataStatus[tdName] = r.state.dataStatus[tdName].Set(statusSkip)
				r.Skip(res, err)
				return nil
			}
			if err := r.api.ManifestPut(r.config.schemeReg, r.config.Repo1, dig.String(), dig, td); err != nil {
				r.APIFail(res, err, stateAPIManifestPutDigest)
				return nil
			}
			r.APIPass(res, stateAPIManifestPutDigest)
			return nil
		})
	}
}

func (r *runner) ChildRun(name string, parent *results, fn func(*runner, *results) error) error {
	res := resultsNew(name, parent)
	if parent != nil {
		parent.children = append(parent.children, res)
	}
	err := fn(r, res)
	res.stop = time.Now()
	if err != nil {
		res.errs = append(res.errs, err)
		res.status = res.status.Set(statusError)
		res.counts[statusError]++
	}
	if parent != nil {
		for i := statusUnknown; i < statusMax; i++ {
			parent.counts[i] += res.counts[i]
		}
		parent.status = parent.status.Set(res.status)
	}
	return err
}

func (r *runner) Skip(res *results, err error) {
	s := statusSkip
	if errors.Is(err, ErrDisabled) {
		s = statusDisabled
	}
	res.status = res.status.Set(s)
	res.counts[s]++
	fmt.Fprintf(res.output, "%s: skipping test:\n  %s\n", res.name,
		strings.ReplaceAll(err.Error(), "\n", "\n  "))
}

func (r *runner) APIFail(res *results, err error, apis ...stateAPIType) {
	res.status = res.status.Set(statusFail)
	res.counts[statusFail]++
	res.errs = append(res.errs, err)
	for _, a := range apis {
		r.state.apiStatus[a] = r.state.apiStatus[a].Set(statusFail)
	}
}

func (r *runner) APIPass(res *results, apis ...stateAPIType) {
	res.status = res.status.Set(statusPass)
	r.results.counts[statusPass]++
	for _, a := range apis {
		r.state.apiStatus[a] = r.state.apiStatus[a].Set(statusPass)
	}
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
			if !r.config.APIs.Tags {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIManifestPutTag, stateAPIManifestPutDigest, stateAPIManifestPutSubject,
			stateAPIBlobPush, stateAPIBlobPostOnly, stateAPIBlobPostPut,
			stateAPIBlobPatchChunk, stateAPIBlobPatchStream, stateAPIBlobMountSource, stateAPIBlobMountAnonymous:
			if !r.config.APIs.Push {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIManifestGetTag, stateAPIManifestGetDigest, stateAPIBlobGetFull, stateAPIBlobGetRange:
			if !r.config.APIs.Pull {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIBlobDelete:
			if !r.config.APIs.Delete.Blob {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIManifestDeleteTag:
			if !r.config.APIs.Delete.Tag {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIManifestDeleteDigest:
			if !r.config.APIs.Delete.Manifest {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration%.0w", aText, ErrDisabled))
			}
		case stateAPIReferrers:
			if !r.config.APIs.Referrer {
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
