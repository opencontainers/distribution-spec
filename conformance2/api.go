package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	specs "github.com/opencontainers/distribution-spec/specs-go/v1"
	digest "github.com/opencontainers/go-digest"
)

type api struct {
	client *http.Client
	// TODO: include auth
}

func apiNew(client *http.Client) *api {
	return &api{
		client: client,
	}
}

type apiOpt struct {
	reqFn  func(*http.Request) error
	respFn func(*http.Response) error
	out    *bytes.Buffer
}

func (a *api) Do(opts ...apiOpt) error {
	reqFns := []func(*http.Request) error{}
	respFns := []func(*http.Response) error{}
	var out *bytes.Buffer
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
	req := &http.Request{}
	for _, reqFn := range reqFns {
		err := reqFn(req)
		if err != nil {
			return err
		}
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
	errs := []error{}
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

func (a *api) TagList(registry, repo string) (specs.TagList, error) {
	tl := specs.TagList{}
	u, err := url.Parse(registry + "/v2/" + repo + "/tags/list")
	if err != nil {
		return tl, err
	}
	err = a.Do(
		apiWithURL(u),
		apiWithOr(
			[]apiOpt{
				apiExpectStatus(http.StatusOK),
				apiExpectJSONBody(&tl),
			},
			[]apiOpt{
				apiExpectStatus(http.StatusNotFound),
			},
		),
	)
	return tl, err
}

func (a *api) BlobPostOnly(registry, repo string, dig digest.Digest, td *testData) error {
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
	headers := http.Header{}
	err = a.Do(
		apiWithMethod("POST"),
		apiWithURL(u),
		apiWithHeaderAdd("Content-Length", fmt.Sprintf("%d", len(bodyBytes))),
		apiWithHeaderAdd("Content-Type", "application/octet-stream"),
		apiWithBody(io.NopCloser(bytes.NewReader(bodyBytes))),
		apiExpectStatus(http.StatusCreated),
		apiExpectHeaders(&headers),
	)
	if err != nil {
		return fmt.Errorf("blob post failed: %v", err)
	}
	l := headers.Get("Location")
	if l == "" {
		return fmt.Errorf("blob post did not return a location")
	}
	td.repo = repo
	return nil
}

func (a *api) BlobPostPut(registry, repo string, dig digest.Digest, td *testData) error {
	bodyBytes, ok := td.blobs[dig]
	if !ok {
		return fmt.Errorf("BlobPostPut missing expected digest to send: %s", dig.String())
	}
	u, err := url.Parse(registry + "/v2/" + repo + "/blobs/uploads/")
	if err != nil {
		return err
	}
	headers := http.Header{}
	err = a.Do(
		apiWithMethod("POST"),
		apiWithURL(u),
		apiExpectStatus(http.StatusAccepted),
		apiExpectHeaders(&headers),
	)
	if err != nil {
		return fmt.Errorf("blob post failed: %v", err)
	}
	l := headers.Get("Location")
	if l == "" {
		return fmt.Errorf("blob post did not return a location")
	}
	uPut, err := u.Parse(l)
	if err != nil {
		return fmt.Errorf("blob post could not parse location header: %v", err)
	}
	qa := uPut.Query()
	qa.Set("digest", dig.String())
	uPut.RawQuery = qa.Encode()
	err = a.Do(
		apiWithMethod("PUT"),
		apiWithURL(uPut),
		apiWithHeaderAdd("Content-Length", fmt.Sprintf("%d", len(bodyBytes))),
		apiWithHeaderAdd("Content-Type", "application/octet-stream"),
		apiWithBody(io.NopCloser(bytes.NewReader(bodyBytes))),
		apiExpectStatus(http.StatusCreated),
		apiExpectHeaders(&headers),
	)
	if err != nil {
		return fmt.Errorf("blob put failed: %v", err)
	}
	l = headers.Get("location")
	if l == "" {
		return fmt.Errorf("blob put did not return a location header")
	}
	td.repo = repo
	return nil
}

func (a *api) ManifestPut(registry, repo, ref string, dig digest.Digest, td *testData) error {
	bodyBytes, ok := td.manifests[dig]
	if !ok {
		return fmt.Errorf("ManifestPut missing expected digest to send: %s", dig.String())
	}
	u, err := url.Parse(registry + "/v2/" + repo + "/manifests/" + ref)
	if err != nil {
		return err
	}
	mediaType := getMediaType(bodyBytes)
	headers := http.Header{}
	err = a.Do(
		apiWithMethod("PUT"),
		apiWithURL(u),
		apiWithBody(io.NopCloser(bytes.NewReader(bodyBytes))),
		apiWithHeaderAdd("Content-Type", mediaType),
		apiExpectStatus(http.StatusCreated),
		apiExpectHeaders(&headers),
	)
	if err != nil {
		return fmt.Errorf("manifest put failed: %v", err)
	}
	// TODO: validate headers: location, docker-content-digest (optional), oci-subject (depending on option)
	td.repo = repo
	return nil
}

// apiWithOr succeeds with any of the lists of respFn's are all successful.
// Note that reqFn entries are ignored.
func apiWithOr(optLists ...[]apiOpt) apiOpt {
	return apiOpt{
		respFn: func(resp *http.Response) error {
			var err error
			for _, opts := range optLists {
				err = nil
				for _, opt := range opts {
					if opt.respFn != nil {
						err = opt.respFn(resp)
						if err != nil {
							break
						}
					}
				}
				if err == nil {
					return nil
				}
			}
			return err
		},
	}
}

func apiWithMethod(method string) apiOpt {
	return apiOpt{
		reqFn: func(req *http.Request) error {
			req.Method = method
			return nil
		},
	}
}

func apiWithURL(u *url.URL) apiOpt {
	return apiOpt{
		reqFn: func(req *http.Request) error {
			req.URL = u
			return nil
		},
	}
}

func apiWithHeaderAdd(key, value string) apiOpt {
	return apiOpt{
		reqFn: func(req *http.Request) error {
			if req.Header == nil {
				req.Header = http.Header{}
			}
			req.Header.Add(key, value)
			return nil
		},
	}
}

func apiWithBody(body io.ReadCloser) apiOpt {
	return apiOpt{
		reqFn: func(req *http.Request) error {
			req.Body = body
			return nil
		},
	}
}

func apiExpectHeaders(h *http.Header) apiOpt {
	return apiOpt{
		respFn: func(resp *http.Response) error {
			*h = resp.Header
			return nil
		},
	}
}

func apiExpectStatus(statusCodes ...int) apiOpt {
	return apiOpt{
		respFn: func(resp *http.Response) error {
			for _, c := range statusCodes {
				if resp.StatusCode == c {
					return nil
				}
			}
			return fmt.Errorf("unexpected status code %d", resp.StatusCode)
		},
	}
}

func apiExpectJSONBody(data any) apiOpt {
	return apiOpt{
		respFn: func(resp *http.Response) error {
			return json.NewDecoder(resp.Body).Decode(data)
		},
	}
}

type wrapTransport struct {
	out  *bytes.Buffer
	orig http.RoundTripper
}

func (wt *wrapTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := wt.orig.RoundTrip(req)
	if wt.out != nil {
		// copy headers to censor auth field
		reqHead := req.Header.Clone()
		if reqHead.Get("Authorization") != "" {
			reqHead.Set("Authorization", "[censored]")
		}
		reqCensored := req
		reqCensored.Header = reqHead
		fmt.Fprintf(wt.out, "%s\n~~~ REQUEST ~~~\n", strings.Repeat("=", 80))
		// TODO: switch to output individual fields
		reqCensored.Write(wt.out)
		if err == nil {
			fmt.Fprintf(wt.out, "%s\n~~~ RESPONSE ~~~\n", strings.Repeat("-", 80))
			// TODO: switch to ouput individual fields, do not output body
			resp.Write(wt.out)
		}
		if err != nil {
			fmt.Fprintf(wt.out, "%s\n~~~ Error ~~~\n%s\n", strings.Repeat("-", 80), err.Error())
		}
		fmt.Fprintf(wt.out, "%s\n", strings.Repeat("=", 80))
	}
	return resp, err
}

type detectMT struct {
	MediaType string `json:"mediaType"`
}

func getMediaType(body []byte) string {
	dmt := detectMT{
		MediaType: "application/vnd.oci.image.manifest.v1+json",
	}
	_ = json.Unmarshal(body, &dmt)
	return dmt.MediaType
}
