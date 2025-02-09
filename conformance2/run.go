package main

import (
	"bytes"
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
	dataImage = "01-image"
	dataIndex = "02-index"
)

var blobAPIs = []apiType{apiBlobPostPut, apiBlobPostOnly}

type runner struct {
	common   *runnerCommon
	name     string // name of current runner step, concatenated onto the parent's name
	results  results
	children []*runner
}

type runnerCommon struct {
	config     config
	api        *api
	apiStatus  map[apiType]status
	data       map[string]*testData
	dataStatus map[string]status
}

func runnerNew(c config) (*runner, error) {
	lvl := slog.LevelWarn
	if c.LogLevel != "" {
		err := lvl.UnmarshalText([]byte(c.LogLevel))
		if err != nil {
			return nil, fmt.Errorf("failed to parse logging level %s: %w", c.LogLevel, err)
		}
	}
	ret := runner{
		name: "OCI Conformance Test",
		results: results{
			output: &bytes.Buffer{},
			start:  time.Now(),
		},
		common: &runnerCommon{
			config:     c,
			api:        apiNew(http.DefaultClient),
			apiStatus:  map[apiType]status{},
			data:       map[string]*testData{},
			dataStatus: map[string]status{},
		},
	}
	return &ret, nil
}

func (r *runner) TestAll() error {
	errs := []error{}
	r.results.start = time.Now()

	err := r.GenerateData()
	if err != nil {
		return fmt.Errorf("aborting tests, unable to generate data: %w", err)
	}

	err = r.TestEmpty()
	if err != nil {
		errs = append(errs, err)
	}

	err = r.TestPush(dataImage)
	if err != nil {
		errs = append(errs, err)
	}
	err = r.TestPush(dataIndex)
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
	r.common.dataStatus[tdName] = statusUnknown
	r.common.data[tdName] = newTestData("OCI Image", "image")
	digCList := []digest.Digest{}
	digUCList := []digest.Digest{}
	for l := 0; l < len(blobAPIs); l++ {
		digC, digUC, _, err := r.common.data[tdName].genLayer(l)
		if err != nil {
			return fmt.Errorf("failed to generate test data layer %d: %w", l, err)
		}
		digCList = append(digCList, digC)
		digUCList = append(digUCList, digUC)
	}
	cDig, _, err := r.common.data[tdName].genConfig(platform{OS: "linux", Architecture: "amd64"}, digUCList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	mDig, _, err := r.common.data[tdName].genManifest(cDig, digCList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}
	_ = mDig
	// multi-platform index
	tdName = dataIndex
	r.common.dataStatus[tdName] = statusUnknown
	r.common.data[tdName] = newTestData("OCI Index", "index")
	platList := []*platform{
		{OS: "linux", Architecture: "amd64"},
		{OS: "linux", Architecture: "arm64"},
	}
	digImgList := []digest.Digest{}
	for _, p := range platList {
		digCList = []digest.Digest{}
		digUCList = []digest.Digest{}
		for l := 0; l < len(blobAPIs); l++ {
			digC, digUC, _, err := r.common.data[tdName].genLayer(l)
			if err != nil {
				return fmt.Errorf("failed to generate test data layer %d: %w", l, err)
			}
			digCList = append(digCList, digC)
			digUCList = append(digUCList, digUC)
		}
		cDig, _, err := r.common.data[tdName].genConfig(*p, digUCList)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		mDig, _, err := r.common.data[tdName].genManifest(cDig, digCList)
		if err != nil {
			return fmt.Errorf("failed to generate test data: %w", err)
		}
		digImgList = append(digImgList, mDig)
	}
	_, _, err = r.common.data[tdName].genIndex(platList, digImgList)
	if err != nil {
		return fmt.Errorf("failed to generate test data: %w", err)
	}

	return nil
}

func (r *runner) Report(w io.Writer) {
	fmt.Fprintf(w, "Test results\n")
	r.ReportWalkErr(w, "")
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
	for i := apiType(0); i < apiMax; i++ {
		pad := ""
		if len(i.String()) < padWidth {
			pad = strings.Repeat(".", padWidth-len(i.String()))
		}
		fmt.Fprintf(w, "  %s%s: %10s\n", i.String(), pad, r.common.apiStatus[i].String())
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "Data conformance:\n")
	tdNames := []string{}
	for tdName := range r.common.data {
		tdNames = append(tdNames, tdName)
	}
	sort.Strings(tdNames)
	for _, tdName := range tdNames {
		pad := ""
		if len(r.common.data[tdName].name) < padWidth {
			pad = strings.Repeat(".", padWidth-len(r.common.data[tdName].name))
		}
		fmt.Fprintf(w, "  %s%s: %10s\n", r.common.data[tdName].name, pad, r.common.dataStatus[tdName].String())
	}
	fmt.Fprintf(w, "\n")
	// TODO: include config
}

func (r *runner) ReportWalkErr(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s%s: %s\n", prefix, r.name, r.results.status)
	if len(r.children) == 0 && len(r.results.errs) > 0 {
		// show errors from leaf nodes
		for _, err := range r.results.errs {
			fmt.Fprintf(w, "%s - %s\n", prefix, err.Error())
		}
	}
	if len(r.children) > 0 {
		for _, child := range r.children {
			child.ReportWalkErr(w, prefix+"  ")
		}
	}
}

func (r *runner) TestEmpty() error {
	return r.Child("empty", func(r *runner) error {
		errs := []error{}
		if err := r.TestEmptyTagList(); err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		return nil
	})
}

func (r *runner) TestEmptyTagList() error {
	return r.Child("tag list", func(r *runner) error {
		errs := []error{}
		if err := r.APIRequire(apiTagList); err != nil {
			r.Skip(err)
			return nil
		}
		if _, err := r.common.api.TagList(r.common.config.schemeReg, r.common.config.Repo1); err != nil {
			r.APIFail(apiTagList)
			errs = append(errs, err)
		} else {
			r.APIPass(apiTagList)
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		return nil
	})
}

func (r *runner) TestPush(tdName string) error {
	// add more APIs
	return r.Child("push", func(r *runner) error {
		errs := []error{}
		curAPI := 0
		for dig := range r.common.data[tdName].blobs {
			curAPI = (curAPI + 1) % len(blobAPIs)
			var err error
			switch blobAPIs[curAPI] {
			case apiBlobPostPut:
				err = r.TestPushBlobPostPut(tdName, dig)
				if err != nil {
					errs = append(errs, err)
				}
			case apiBlobPostOnly:
				err = r.TestPushBlobPostOnly(tdName, dig)
				if err != nil {
					errs = append(errs, err)
				}
			}
			// TODO: fallback to any blob push method
		}
		for _, dig := range r.common.data[tdName].manOrder {
			err := r.TestPushManifest(tdName, dig)
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			r.common.dataStatus[tdName] = r.common.dataStatus[tdName].Set(statusFail)
			return errors.Join(errs...)
		}
		r.common.dataStatus[tdName] = r.common.dataStatus[tdName].Set(statusPass)
		return nil
	})
}

func (r *runner) TestPushBlobPostPut(tdName string, dig digest.Digest) error {
	return r.Child("blob-post-put", func(r *runner) error {
		if err := r.APIRequire(apiBlobPostPut); err != nil {
			r.Skip(err)
			return nil
		}
		if err := r.common.api.BlobPostPut(r.common.config.schemeReg, r.common.config.Repo1, dig, r.common.data[tdName]); err != nil {
			r.APIFail(apiBlobPostPut)
			return err
		}
		r.APIPass(apiBlobPostPut)
		return nil
	})
}

func (r *runner) TestPushBlobPostOnly(tdName string, dig digest.Digest) error {
	return r.Child("blob-post-only", func(r *runner) error {
		if err := r.APIRequire(apiBlobPostOnly); err != nil {
			r.Skip(err)
			return nil
		}
		if err := r.common.api.BlobPostOnly(r.common.config.schemeReg, r.common.config.Repo1, dig, r.common.data[tdName]); err != nil {
			r.APIFail(apiBlobPostOnly)
			return err
		}
		r.APIPass(apiBlobPostOnly)
		return nil
	})
}

func (r *runner) TestPushManifest(tdName string, dig digest.Digest) error {
	td := r.common.data[tdName]
	if td.manOrder[len(td.manOrder)-1] == dig && td.tag != "" {
		// push by tag
		return r.Child("manifest-by-tag", func(r *runner) error {
			if err := r.APIRequire(apiManifestPutTag); err != nil {
				r.common.dataStatus[tdName] = r.common.dataStatus[tdName].Set(statusSkip)
				r.Skip(err)
				return nil
			}
			if err := r.common.api.ManifestPut(r.common.config.schemeReg, r.common.config.Repo1, td.tag, dig, td); err != nil {
				r.APIFail(apiManifestPutTag)
				return err
			}
			r.APIPass(apiManifestPutTag)
			return nil
		})
	} else {
		// push by digest
		return r.Child("manifest-by-digest", func(r *runner) error {
			if err := r.APIRequire(apiManifestPutDigest); err != nil {
				r.common.dataStatus[tdName] = r.common.dataStatus[tdName].Set(statusSkip)
				r.Skip(err)
				return nil
			}
			if err := r.common.api.ManifestPut(r.common.config.schemeReg, r.common.config.Repo1, dig.String(), dig, td); err != nil {
				r.APIFail(apiManifestPutDigest)
				return err
			}
			r.APIPass(apiManifestPutDigest)
			return nil
		})
	}
}

func (r *runner) Child(name string, fn func(*runner) error) error {
	rChild := runner{
		results: results{
			output: &bytes.Buffer{},
		},
		common: r.common,
	}
	if r.name != "" {
		rChild.name = fmt.Sprintf("%s/%s", r.name, name)
	} else {
		rChild.name = name
	}
	r.children = append(r.children, &rChild)
	rChild.results.start = time.Now()
	err := fn(&rChild)
	rChild.results.stop = time.Now()
	if err != nil {
		rChild.results.errs = append(rChild.results.errs, err)
		rChild.results.status = rChild.results.status.Set(statusError)
		rChild.results.counts[statusError]++
	}
	for i := statusUnknown; i < statusMax; i++ {
		r.results.counts[i] += rChild.results.counts[i]
	}
	r.results.status = r.results.status.Set(rChild.results.status)
	return err
}

func (r *runner) Skip(err error) {
	r.results.status = statusSkip
	r.results.counts[statusSkip]++
	fmt.Fprintf(r.results.output, "%s: skipping test:\n  %s\n", r.name,
		strings.ReplaceAll(err.Error(), "\n", "\n  "))
}

func (r *runner) APIFail(apis ...apiType) {
	r.results.status = r.results.status.Set(statusFail)
	r.results.counts[statusFail]++
	for _, a := range apis {
		r.common.apiStatus[a] = r.common.apiStatus[a].Set(statusFail)
	}
}

func (r *runner) APIPass(apis ...apiType) {
	r.results.status = r.results.status.Set(statusPass)
	r.results.counts[statusPass]++
	for _, a := range apis {
		r.common.apiStatus[a] = r.common.apiStatus[a].Set(statusPass)
	}
}

func (r *runner) APIRequire(apis ...apiType) error {
	errs := []error{}
	for _, a := range apis {
		aText, err := a.MarshalText()
		if err != nil {
			errs = append(errs, fmt.Errorf("unknown api %d", a))
			continue
		}
		// check the configuration disables the api
		switch a {
		case apiTagList:
			if !r.common.config.APIs.Tags {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration", aText))
			}
		case apiManifestPutTag, apiManifestPutDigest, apiManifestPutSubject,
			apiBlobPush, apiBlobPostOnly, apiBlobPostPut,
			apiBlobPatchChunk, apiBlobPatchStream, apiBlobMountSource, apiBlobMountAnonymous:
			if !r.common.config.APIs.Push {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration", aText))
			}
		case apiManifestGetTag, apiManifestGetDigest, apiBlobGetFull, apiBlobGetRange:
			if !r.common.config.APIs.Pull {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration", aText))
			}
		case apiBlobDelete:
			if !r.common.config.APIs.Delete.Blob {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration", aText))
			}
		case apiManifestDeleteTag:
			if !r.common.config.APIs.Delete.Tag {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration", aText))
			}
		case apiManifestDeleteDigest:
			if !r.common.config.APIs.Delete.Manifest {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration", aText))
			}
		case apiReferrers:
			if !r.common.config.APIs.Referrer {
				errs = append(errs, fmt.Errorf("api %s is disabled in the configuration", aText))
			}
		}
		// do not check the [r.global.apiState] since tests may pass or fail based on different input data
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
