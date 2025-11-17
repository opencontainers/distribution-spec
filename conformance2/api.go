package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	specs "github.com/opencontainers/distribution-spec/specs-go/v1"
	digest "github.com/opencontainers/go-digest"
)

type api struct {
	client     *http.Client
	user, pass string
}

type apiOpt func(*api)

func apiNew(client *http.Client, opts ...apiOpt) *api {
	a := &api{
		client: client,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func apiWithAuth(user, pass string) apiOpt {
	return func(a *api) {
		a.user = user
		a.pass = pass
	}
}

type apiDoOpt struct {
	reqFn  func(*http.Request) error
	respFn func(*http.Response) error
	out    io.Writer
}

func (a *api) Do(opts ...apiDoOpt) error {
	errs := []error{}
	reqFns := []func(*http.Request) error{}
	respFns := []func(*http.Response) error{}
	var out io.Writer
	for _, opt := range opts {
		if opt.reqFn != nil {
			reqFns = append(reqFns, opt.reqFn)
		}
		if opt.respFn != nil {
			respFns = append(respFns, opt.respFn)
		}
		if opt.out != nil {
			out = opt.out
		}
	}
	req, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		return err
	}
	for _, reqFn := range reqFns {
		err := reqFn(req)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return errors.Join(errs...)
	}
	if out != nil {
		out = redactWriter{w: out}
	}
	wt := &wrapTransport{out: out, orig: a.client.Transport}
	if a.client.Transport == nil {
		wt.orig = http.DefaultTransport
	}
	c := *a.client
	c.Transport = wt
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	// on auth failures, generate the auth header and retry
	if resp.StatusCode == http.StatusUnauthorized {
		auth, err := a.getAuthHeader(c, resp)
		if err != nil {
			errs = append(errs, err)
		}
		if err == nil && auth != "" {
			req.Header.Set("Authorization", auth)
			if req.GetBody != nil {
				req.Body, err = req.GetBody()
				if err != nil {
					return fmt.Errorf("failed to reset body after auth request: %w", err)
				}
			}
			resp, err = c.Do(req)
			if err != nil {
				return err
			}
		}
	}
	for _, respFn := range respFns {
		err := respFn(resp)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return errors.Join(errs...)
	}
	return nil
}

func (a *api) BlobDelete(registry, repo string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	u, err := url.Parse(registry + "/v2/" + repo + "/blobs/" + dig.String())
	if err != nil {
		return err
	}
	var status int
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("DELETE"),
		apiWithURL(u),
		apiExpectStatus(http.StatusAccepted, http.StatusNotFound, http.StatusBadRequest, http.StatusMethodNotAllowed),
		apiReturnStatus(&status),
	)
	if err != nil {
		return fmt.Errorf("blob delete failed: %v", err)
	}
	if status == http.StatusBadRequest || status == http.StatusMethodNotAllowed {
		return fmt.Errorf("registry returned status %d%.0w", status, ErrRegUnsupported)
	}
	return nil
}

func (a *api) BlobGetFull(registry, repo string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	u, err := url.Parse(registry + "/v2/" + repo + "/blobs/" + dig.String())
	if err != nil {
		return err
	}
	if val, ok := td.blobs[dig]; ok {
		opts = append(opts, apiExpectBody(val), apiExpectHeader("Content-Length", fmt.Sprintf("%d", len(val))))
	}
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("GET"),
		apiWithURL(u),
		apiExpectStatus(http.StatusOK),
	)
	if err != nil {
		return fmt.Errorf("blob get failed: %v", err)
	}
	return nil
}

func (a *api) BlobHead(registry, repo string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	u, err := url.Parse(registry + "/v2/" + repo + "/blobs/" + dig.String())
	if err != nil {
		return err
	}
	if val, ok := td.blobs[dig]; ok {
		opts = append(opts, apiExpectHeader("Content-Length", fmt.Sprintf("%d", len(val))))
	}
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("HEAD"),
		apiWithURL(u),
		apiExpectStatus(http.StatusOK),
		apiExpectBody([]byte{}),
	)
	if err != nil {
		return fmt.Errorf("blob head failed: %v", err)
	}
	return nil
}

func (a *api) BlobMount(registry, repo, source string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	bodyBytes, ok := td.blobs[dig]
	if !ok {
		return fmt.Errorf("BlobPostPut missing expected digest to send: %s", dig.String())
	}
	u, err := url.Parse(registry + "/v2/" + repo + "/blobs/uploads/")
	if err != nil {
		return err
	}
	qa := u.Query()
	qa.Set("mount", dig.String())
	if source != "" {
		qa.Set("from", source)
	}
	u.RawQuery = qa.Encode()
	// TODO: add digest algorithm if not sha256
	loc := ""
	status := 0
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("POST"),
		apiWithURL(u),
		apiExpectStatus(http.StatusCreated, http.StatusAccepted),
		apiReturnHeader("Location", &loc),
		apiReturnStatus(&status),
	)
	if err != nil {
		return fmt.Errorf("blob post failed: %v", err)
	}
	if loc == "" {
		return fmt.Errorf("blob post did not return a location")
	}
	if status == http.StatusCreated {
		// successful mount
		return nil
	}
	// fallback to post+put
	u, err = u.Parse(loc)
	if err != nil {
		return fmt.Errorf("blob post could not parse location header: %v", err)
	}
	qa = u.Query()
	qa.Set("digest", dig.String())
	u.RawQuery = qa.Encode()
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("PUT"),
		apiWithURL(u),
		apiWithHeaderAdd("Content-Length", fmt.Sprintf("%d", len(bodyBytes))),
		apiWithHeaderAdd("Content-Type", "application/octet-stream"),
		apiWithBody(bodyBytes),
		apiExpectStatus(http.StatusCreated),
		apiExpectHeader("Location", ""),
	)
	if err != nil {
		return fmt.Errorf("blob put failed: %v", err)
	}
	return fmt.Errorf("registry returned status %d, fell back to blob POST+PUT%.0w", status, ErrRegUnsupported)
}

func (a *api) BlobPatchChunked(registry, repo string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	bodyBytes, ok := td.blobs[dig]
	if !ok {
		return fmt.Errorf("BlobPatchChunked missing expected digest to send: %s", dig.String())
	}
	u, err := url.Parse(registry + "/v2/" + repo + "/blobs/uploads/")
	if err != nil {
		return err
	}
	// TODO: add digest algorithm if not sha256
	minStr := ""
	loc := ""
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("POST"),
		apiWithURL(u),
		apiExpectStatus(http.StatusAccepted),
		apiReturnHeader("OCI-Chunk-Min-Length", &minStr),
		apiReturnHeader("Location", &loc),
	)
	if err != nil {
		return fmt.Errorf("blob post failed: %v", err)
	}
	// calc chunk size to make 3 chunks, adjust to min chunk size if specified
	chunkSize := len(bodyBytes)/3 + 1
	if minStr != "" {
		min, err := strconv.Atoi(minStr)
		if err != nil {
			return fmt.Errorf("parsing OCI-Chunk-Min-Length size %q failed: %v", minStr, err)
		}
		if min > chunkSize {
			chunkSize = min
		}
	}
	if chunkSize < chunkMin {
		chunkSize = chunkMin
	}
	if chunkSize > len(bodyBytes) {
		chunkSize = len(bodyBytes)
	}
	lastByte := -1
	// loop over the number of chunks
	for lastByte < len(bodyBytes)-1 {
		if loc == "" {
			return fmt.Errorf("blob request did not return a location")
		}
		u, err = u.Parse(loc)
		if err != nil {
			return fmt.Errorf("blob request could not parse location header: %v", err)
		}
		start := lastByte + 1
		lastByte = min(start+chunkSize-1, len(bodyBytes)-1)
		err = a.Do(apiWithAnd(opts),
			apiWithMethod("PATCH"),
			apiWithURL(u),
			apiWithHeaderAdd("Content-Type", "application/octet-stream"),
			apiWithHeaderAdd("Content-Length", fmt.Sprintf("%d", lastByte-start+1)),
			apiWithHeaderAdd("Content-Range", fmt.Sprintf("%d-%d", start, lastByte)),
			apiWithBody(bodyBytes[start:lastByte+1]),
			apiExpectStatus(http.StatusAccepted),
			apiReturnHeader("Location", &loc),
		)
		if err != nil {
			return fmt.Errorf("blob patch failed: %v", err)
		}
	}
	if loc == "" {
		return fmt.Errorf("blob patch did not return a location")
	}
	u, err = u.Parse(loc)
	if err != nil {
		return fmt.Errorf("blob patch could not parse location header: %v", err)
	}
	qa := u.Query()
	qa.Set("digest", dig.String())
	u.RawQuery = qa.Encode()
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("PUT"),
		apiWithURL(u),
		apiWithHeaderAdd("Content-Type", "application/octet-stream"),
		apiWithHeaderAdd("Content-Length", "0"),
		apiExpectStatus(http.StatusCreated),
		apiExpectHeader("Location", ""),
	)
	if err != nil {
		return fmt.Errorf("blob put failed: %v", err)
	}
	return nil
}

func (a *api) BlobPatchStream(registry, repo string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	bodyBytes, ok := td.blobs[dig]
	if !ok {
		return fmt.Errorf("BlobPatchStream missing expected digest to send: %s", dig.String())
	}
	u, err := url.Parse(registry + "/v2/" + repo + "/blobs/uploads/")
	if err != nil {
		return err
	}
	// TODO: add digest algorithm if not sha256
	loc := ""
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("POST"),
		apiWithURL(u),
		apiExpectStatus(http.StatusAccepted),
		apiReturnHeader("Location", &loc),
	)
	if err != nil {
		return fmt.Errorf("blob post failed: %v", err)
	}
	if loc == "" {
		return fmt.Errorf("blob post did not return a location")
	}
	u, err = u.Parse(loc)
	if err != nil {
		return fmt.Errorf("blob post could not parse location header: %v", err)
	}
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("PATCH"),
		apiWithURL(u),
		apiWithHeaderAdd("Content-Type", "application/octet-stream"),
		apiWithBody(bodyBytes),
		apiExpectStatus(http.StatusAccepted),
		apiReturnHeader("Location", &loc),
	)
	if err != nil {
		return fmt.Errorf("blob patch failed: %v", err)
	}
	if loc == "" {
		return fmt.Errorf("blob patch did not return a location")
	}
	u, err = u.Parse(loc)
	if err != nil {
		return fmt.Errorf("blob patch could not parse location header: %v", err)
	}
	qa := u.Query()
	qa.Set("digest", dig.String())
	u.RawQuery = qa.Encode()
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("PUT"),
		apiWithURL(u),
		apiWithHeaderAdd("Content-Type", "application/octet-stream"),
		apiExpectStatus(http.StatusCreated),
		apiExpectHeader("Location", ""),
	)
	if err != nil {
		return fmt.Errorf("blob put failed: %v", err)
	}
	return nil
}

func (a *api) BlobPostOnly(registry, repo string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	bodyBytes, ok := td.blobs[dig]
	if !ok {
		return fmt.Errorf("BlobPostOnly missing expected digest to send: %s", dig.String())
	}
	u, err := url.Parse(registry + "/v2/" + repo + "/blobs/uploads/")
	if err != nil {
		return err
	}
	qa := u.Query()
	qa.Set("digest", dig.String())
	u.RawQuery = qa.Encode()
	var status int
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("POST"),
		apiWithURL(u),
		apiWithHeaderAdd("Content-Length", fmt.Sprintf("%d", len(bodyBytes))),
		apiWithHeaderAdd("Content-Type", "application/octet-stream"),
		apiWithBody(bodyBytes),
		apiExpectStatus(http.StatusCreated, http.StatusAccepted),
		apiExpectHeader("Location", ""),
		apiReturnStatus(&status),
	)
	if err != nil {
		return fmt.Errorf("blob post failed: %v", err)
	}
	if status != http.StatusCreated {
		return fmt.Errorf("registry returned status %d%.0w", status, ErrRegUnsupported)
	}
	return nil
}

func (a *api) BlobPostPut(registry, repo string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	bodyBytes, ok := td.blobs[dig]
	if !ok {
		return fmt.Errorf("BlobPostPut missing expected digest to send: %s", dig.String())
	}
	u, err := url.Parse(registry + "/v2/" + repo + "/blobs/uploads/")
	if err != nil {
		return err
	}
	// TODO: add digest algorithm if not sha256
	loc := ""
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("POST"),
		apiWithURL(u),
		apiExpectStatus(http.StatusAccepted),
		apiReturnHeader("Location", &loc),
	)
	if err != nil {
		return fmt.Errorf("blob post failed: %v", err)
	}
	if loc == "" {
		return fmt.Errorf("blob post did not return a location")
	}
	u, err = u.Parse(loc)
	if err != nil {
		return fmt.Errorf("blob post could not parse location header: %v", err)
	}
	qa := u.Query()
	qa.Set("digest", dig.String())
	u.RawQuery = qa.Encode()
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("PUT"),
		apiWithURL(u),
		apiWithHeaderAdd("Content-Length", fmt.Sprintf("%d", len(bodyBytes))),
		apiWithHeaderAdd("Content-Type", "application/octet-stream"),
		apiWithBody(bodyBytes),
		apiExpectStatus(http.StatusCreated),
		apiExpectHeader("Location", ""),
	)
	if err != nil {
		return fmt.Errorf("blob put failed: %v", err)
	}
	return nil
}

func (a *api) ManifestDelete(registry, repo, ref string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	u, err := url.Parse(registry + "/v2/" + repo + "/manifests/" + ref)
	if err != nil {
		return err
	}
	var status int
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("DELETE"),
		apiWithURL(u),
		apiExpectStatus(http.StatusAccepted, http.StatusNotFound),
		apiReturnStatus(&status),
	)
	if err != nil {
		return fmt.Errorf("manifest delete failed: %v", err)
	}
	if status == http.StatusBadRequest || status == http.StatusMethodNotAllowed {
		return fmt.Errorf("registry returned status %d%.0w", status, ErrRegUnsupported)
	}
	return nil
}

func (a *api) ManifestGet(registry, repo, ref string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	u, err := url.Parse(registry + "/v2/" + repo + "/manifests/" + ref)
	if err != nil {
		return err
	}
	if val, ok := td.manifests[dig]; ok {
		mediaType := detectMediaType(val)
		opts = append(opts,
			apiExpectBody(val),
			apiExpectHeader("Content-Type", mediaType),
			apiExpectHeader("Content-Length", fmt.Sprintf("%d", len(val))),
		)
	}
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("GET"),
		apiWithURL(u),
		apiExpectStatus(http.StatusOK),
		apiWithHeaderAdd("Accept", "application/vnd.oci.image.index.v1+json"),
		apiWithHeaderAdd("Accept", "application/vnd.oci.image.manifest.v1+json"),
	)
	if err != nil {
		return fmt.Errorf("manifest get failed: %v", err)
	}
	return nil
}

func (a *api) ManifestHead(registry, repo, ref string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	u, err := url.Parse(registry + "/v2/" + repo + "/manifests/" + ref)
	if err != nil {
		return err
	}
	if val, ok := td.manifests[dig]; ok {
		mediaType := detectMediaType(val)
		opts = append(opts,
			apiExpectHeader("Content-Type", mediaType),
			apiExpectHeader("Content-Length", fmt.Sprintf("%d", len(val))),
		)
	}
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("HEAD"),
		apiWithURL(u),
		apiExpectStatus(http.StatusOK),
		apiWithHeaderAdd("Accept", "application/vnd.oci.image.index.v1+json"),
		apiWithHeaderAdd("Accept", "application/vnd.oci.image.manifest.v1+json"),
		apiExpectBody([]byte{}),
	)
	if err != nil {
		return fmt.Errorf("manifest head failed: %v", err)
	}
	return nil
}

func (a *api) ManifestPut(registry, repo, ref string, dig digest.Digest, td *testData, opts ...apiDoOpt) error {
	bodyBytes, ok := td.manifests[dig]
	if !ok {
		return fmt.Errorf("ManifestPut missing expected digest to send: %s", dig.String())
	}
	u, err := url.Parse(registry + "/v2/" + repo + "/manifests/" + ref)
	if err != nil {
		return err
	}
	mediaType := detectMediaType(bodyBytes)
	resp := http.Response{}
	err = a.Do(apiWithAnd(opts),
		apiWithMethod("PUT"),
		apiWithURL(u),
		apiWithBody(bodyBytes),
		apiWithHeaderAdd("Content-Type", mediaType),
		apiExpectStatus(http.StatusCreated),
		apiExpectHeader("Location", ""),
		apiReturnResponse(&resp),
	)
	if err != nil {
		return fmt.Errorf("manifest put failed: %v", err)
	}
	digHeader := resp.Header.Get("Docker-Content-Digest")
	if digHeader == "" {
		return fmt.Errorf("warning: registry does not return a Docker-Content-Digest header")
	}
	if digHeader != "" && digHeader != dig.String() {
		return fmt.Errorf("Docker-Content-Digest header value expected %q, received %q", dig.String(), digHeader)
	}
	return nil
}

func (a *api) ReferrersList(registry, repo string, dig digest.Digest, opts ...apiDoOpt) (index, error) {
	rl := index{}
	u, err := url.Parse(registry + "/v2/" + repo + "/referrers/" + dig.String())
	if err != nil {
		return rl, err
	}
	err = a.Do(apiWithAnd(opts),
		apiWithURL(u),
		apiExpectHeader("Content-Type", "application/vnd.oci.image.index.v1+json"),
		apiExpectStatus(http.StatusOK),
		apiReturnJSONBody(&rl),
	)
	return rl, err
}

func (a *api) TagList(registry, repo string, opts ...apiDoOpt) (specs.TagList, error) {
	tl := specs.TagList{}
	u, err := url.Parse(registry + "/v2/" + repo + "/tags/list")
	if err != nil {
		return tl, err
	}
	err = a.Do(apiWithAnd(opts),
		apiWithURL(u),
		apiWithOr(
			[]apiDoOpt{
				apiExpectStatus(http.StatusOK),
				apiReturnJSONBody(&tl),
			},
			[]apiDoOpt{
				apiExpectStatus(http.StatusNotFound),
			},
		),
	)
	return tl, err
}

func apiWithAnd(opts []apiDoOpt) apiDoOpt {
	ret := apiDoOpt{}
	reqFns := [](func(*http.Request) error){}
	respFns := [](func(*http.Response) error){}
	for _, opt := range opts {
		if opt.reqFn != nil {
			reqFns = append(reqFns, opt.reqFn)
		}
		if opt.respFn != nil {
			respFns = append(respFns, opt.respFn)
		}
		if opt.out != nil {
			ret.out = opt.out
		}
	}
	if len(reqFns) == 1 {
		ret.reqFn = reqFns[0]
	} else if len(reqFns) > 0 {
		ret.reqFn = func(r *http.Request) error {
			errs := []error{}
			for _, fn := range reqFns {
				err := fn(r)
				if err != nil {
					errs = append(errs, err)
				}
			}
			if len(errs) == 1 {
				return errs[0]
			}
			return errors.Join(errs...)
		}
	}
	if len(respFns) == 1 {
		ret.respFn = respFns[0]
	} else if len(respFns) > 0 {
		ret.respFn = func(r *http.Response) error {
			errs := []error{}
			for _, fn := range respFns {
				err := fn(r)
				if err != nil {
					errs = append(errs, err)
				}
			}
			if len(errs) == 1 {
				return errs[0]
			}
			return errors.Join(errs...)
		}
	}
	return ret
}

// apiWithOr succeeds with any of the lists of respFn's are all successful.
// Note that reqFn entries are ignored.
func apiWithOr(optLists ...[]apiDoOpt) apiDoOpt {
	return apiDoOpt{
		respFn: func(resp *http.Response) error {
			errsOr := []error{}
			for _, opts := range optLists {
				errsResp := []error{}
				for _, opt := range opts {
					if opt.respFn != nil {
						err := opt.respFn(resp)
						if err != nil {
							errsResp = append(errsResp, err)
						}
					}
				}
				if len(errsResp) == 0 {
					return nil
				}
				errsOr = append(errsOr, errors.Join(errsResp...))
			}
			return fmt.Errorf("response did not match any condition: %w", errors.Join(errsOr...))
		},
	}
}

func apiWithMethod(method string) apiDoOpt {
	return apiDoOpt{
		reqFn: func(req *http.Request) error {
			req.Method = method
			return nil
		},
	}
}

func apiWithURL(u *url.URL) apiDoOpt {
	return apiDoOpt{
		reqFn: func(req *http.Request) error {
			req.URL = u
			return nil
		},
	}
}

func apiWithHeaderAdd(key, value string) apiDoOpt {
	return apiDoOpt{
		reqFn: func(req *http.Request) error {
			if req.Header == nil {
				req.Header = http.Header{}
			}
			req.Header.Add(key, value)
			return nil
		},
	}
}

func apiWithBody(body []byte) apiDoOpt {
	return apiDoOpt{
		reqFn: func(req *http.Request) error {
			req.Body = io.NopCloser(bytes.NewReader(body))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(body)), nil
			}
			return nil
		},
	}
}

func apiExpectBody(bodyExpect []byte) apiDoOpt {
	return apiDoOpt{
		respFn: func(resp *http.Response) error {
			// read body and replace with a buf reader
			bodyReceived, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read body: %w", err)
			}
			resp.Body = io.NopCloser(bytes.NewReader(bodyReceived))
			if bytes.Equal(bodyExpect, bodyReceived) {
				return nil
			}
			var bufExpect, bufReceived bytes.Buffer
			err = printBody(bodyReceived, &bufReceived)
			if err != nil {
				return fmt.Errorf("failed to print received body: %w", err)
			}
			err = printBody(bodyExpect, &bufExpect)
			if err != nil {
				return fmt.Errorf("failed to print expected body: %w", err)
			}
			return fmt.Errorf("body contents mismatch, expected %s, received %s", bufExpect.String(), bufReceived.String())
		},
	}
}

func apiExpectHeader(key, val string) apiDoOpt {
	return apiDoOpt{
		respFn: func(resp *http.Response) error {
			cur := resp.Header.Get(key)
			if val == "" {
				if cur == "" {
					return fmt.Errorf("missing header %q", key)
				}
			} else {
				if cur != val {
					return fmt.Errorf("header value mismatch for %q, expected %q, received %q", key, val, cur)
				}
			}
			return nil
		},
	}
}

func apiExpectStatus(statusCodes ...int) apiDoOpt {
	return apiDoOpt{
		respFn: func(resp *http.Response) error {
			if slices.Contains(statusCodes, resp.StatusCode) {
				return nil
			}
			return fmt.Errorf("unexpected status code %d", resp.StatusCode)
		},
	}
}

func apiReturnHeader(key string, val *string) apiDoOpt {
	return apiDoOpt{
		respFn: func(resp *http.Response) error {
			cur := resp.Header.Get(key)
			if cur != "" {
				*val = cur
			}
			return nil
		},
	}
}

func apiReturnJSONBody(data any) apiDoOpt {
	return apiDoOpt{
		respFn: func(resp *http.Response) error {
			return json.NewDecoder(resp.Body).Decode(data)
		},
	}
}

func apiReturnResponse(ret *http.Response) apiDoOpt {
	return apiDoOpt{
		respFn: func(r *http.Response) error {
			*ret = *r
			return nil
		},
	}
}

func apiReturnStatus(status *int) apiDoOpt {
	return apiDoOpt{
		respFn: func(resp *http.Response) error {
			*status = resp.StatusCode
			return nil
		},
	}
}

func apiSaveOutput(out io.Writer) apiDoOpt {
	return apiDoOpt{
		out: out,
	}
}

type authHeader struct {
	Type    string
	Realm   string
	Service string
	Scope   string
}

type authInfo struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
}

func (a *api) getAuthHeader(client http.Client, resp *http.Response) (string, error) {
	header := resp.Header.Get("WWW-Authenticate")
	if resp.StatusCode != http.StatusUnauthorized || header == "" {
		return "", fmt.Errorf("status code or header invalid for adding auth, status %d, header %s", resp.StatusCode, header)
	}
	parsed, err := parseAuthHeader(header)
	if err != nil {
		return "", err
	}
	if parsed.Type == "basic" {
		return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(a.user+":"+a.pass))), nil
	}
	if parsed.Type == "bearer" {
		u, err := resp.Request.URL.Parse(parsed.Realm)
		if err != nil {
			return "", fmt.Errorf("failed to parse realm url: %w", err)
		}
		param := url.Values{}
		param.Set("service", parsed.Service)
		if parsed.Scope != "" {
			param.Set("scope", parsed.Scope)
		}
		u.RawQuery = param.Encode()
		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return "", fmt.Errorf("failed to created request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.SetBasicAuth(a.user, a.pass)
		authResp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to send auth request: %w", err)
		}
		if authResp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("invalid status on auth request: %d", authResp.StatusCode)
		}
		ai := authInfo{}
		if err := json.NewDecoder(authResp.Body).Decode(&ai); err != nil {
			return "", fmt.Errorf("failed to parse auth response: %w", err)
		}
		if ai.AccessToken != "" {
			ai.Token = ai.AccessToken
		}
		return fmt.Sprintf("Bearer %s", ai.Token), nil
	}
	return "", fmt.Errorf("failed to parse auth header, type=%s: %s", parsed.Type, header)
}

var (
	authHeaderMatcher = regexp.MustCompile("(?i).*(bearer|basic).*")
	authParamsMatcher = regexp.MustCompile(`([a-zA-z]+)="(.+?)"`)
)

func parseAuthHeader(header string) (authHeader, error) {
	// TODO: replace with a better parser, quotes should be optional, get character set from upstream http rfc
	var parsed authHeader
	parsed.Type = strings.ToLower(authHeaderMatcher.ReplaceAllString(header, "$1"))
	if parsed.Type == "bearer" {
		matches := authParamsMatcher.FindAllStringSubmatch(header, -1)
		for _, match := range matches {
			switch strings.ToLower(match[1]) {
			case "realm":
				parsed.Realm = match[2]
			case "service":
				parsed.Service = match[2]
			case "scope":
				parsed.Scope = match[2]
			}
		}
	}
	return parsed, nil
}

type wrapTransport struct {
	out  io.Writer
	orig http.RoundTripper
}

func (wt *wrapTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if wt.out != nil {
		_ = printRequest(req, wt.out)
	}
	resp, err := wt.orig.RoundTrip(req)
	if wt.out != nil {
		if err == nil {
			_ = printResponse(resp, wt.out)
		}
		if err != nil {
			fmt.Fprintf(wt.out, "%s\n~~~ Error ~~~\n%s\n", strings.Repeat("-", 80), err.Error())
		}
		fmt.Fprintf(wt.out, "%s\n", strings.Repeat("=", 80))
	}
	return resp, err
}

type detectManifest struct {
	MediaType string      `json:"mediaType"`
	Subject   *descriptor `json:"subject,omitempty"`
}

func detectMediaType(body []byte) string {
	det := detectManifest{
		MediaType: "application/vnd.oci.image.manifest.v1+json",
	}
	_ = json.Unmarshal(body, &det)
	return det.MediaType
}

func detectSubject(body []byte) *descriptor {
	det := detectManifest{}
	_ = json.Unmarshal(body, &det)
	return det.Subject
}

func cloneBodyReq(req *http.Request) ([]byte, error) {
	if req.GetBody != nil {
		rc, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		out, err := io.ReadAll(rc)
		_ = rc.Close()
		return out, err
	}
	if req.Body == nil {
		return []byte{}, nil
	}
	out, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	_ = req.Body.Close()
	// replace the body with a buffer so it can be reused
	req.Body = io.NopCloser(bytes.NewReader(out))
	return out, err
}

func cloneBodyResp(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return []byte{}, nil
	}
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	// replace the body with a buffer so it can be reused
	resp.Body = io.NopCloser(bytes.NewReader(out))
	return out, err
}

func mediaTypeBase(orig string) string {
	base, _, _ := strings.Cut(orig, ";")
	return strings.TrimSpace(strings.ToLower(base))
}

func printBody(body []byte, w io.Writer) error {
	if len(body) == 0 {
		fmt.Fprintf(w, "--- Empty body ---\n")
		return nil
	}
	ct := http.DetectContentType(body)
	switch mediaTypeBase(ct) {
	case "application/json", "text/plain":
		fmt.Fprintf(w, "%.*s\n", truncateBody, string(body))
		if len(body) > truncateBody {
			fmt.Fprintf(w, "--- Truncated body from %d to %d bytes ---\n", len(body), truncateBody)
		}
	default:
		fmt.Fprintf(w, "--- Output of %s not supported, %d bytes not shown ---\n", ct, len(body))
	}
	return nil
}

func printHeaders(headers http.Header, w io.Writer) error {
	fmt.Fprintf(w, "Headers:\n")
	for _, k := range slices.Sorted(maps.Keys(headers)) {
		fmt.Fprintf(w, "  %25s: %v\n", k, headers[k])
	}
	return nil
}

func printRequest(req *http.Request, w io.Writer) error {
	fmt.Fprintf(w, "%s\n~~~ REQUEST ~~~\n", strings.Repeat("=", 80))
	fmt.Fprintf(w, "Method: %s\nURL: %s\n", req.Method, req.URL.String())
	printHeaders(req.Header, w)
	body, err := cloneBodyReq(req)
	if err != nil {
		return err
	}
	printBody(body, w)

	return nil
}

func printResponse(resp *http.Response, w io.Writer) error {
	fmt.Fprintf(w, "%s\n~~~ RESPONSE ~~~\n", strings.Repeat("-", 80))
	fmt.Fprintf(w, "Status: %d\n", resp.StatusCode)
	printHeaders(resp.Header, w)
	body, err := cloneBodyResp(resp)
	if err != nil {
		return err
	}
	printBody(body, w)

	return nil
}

type redactWriter struct {
	w io.Writer
}

var (
	redactRegexp  = regexp.MustCompile(`(?i)("?\w*(?:authorization|token|state)\w*"?(?:=|:)\s*(?:\[)?\s*"?\s*(?:(?:bearer|basic)? )?)[^\s?&"\]]*`)
	redactReplace = []byte("$1*****")
)

func (rw redactWriter) Write(p []byte) (int, error) {
	pRedact := redactRegexp.ReplaceAll(p, redactReplace)
	n, err := rw.w.Write(pRedact)
	if err != nil || n != len(pRedact) {
		return 0, err
	}
	return len(p), nil
}
