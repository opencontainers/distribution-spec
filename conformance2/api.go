package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
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
	out    *bytes.Buffer
}

func (a *api) Do(opts ...apiDoOpt) error {
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
	// on auth failures, generate the auth header and retry
	if resp.StatusCode == http.StatusUnauthorized {
		auth, err := a.getAuthHeader(c, resp)
		if err == nil && auth != "" {
			req.Header.Set("Authorization", auth)
			resp, err = c.Do(req)
			if err != nil {
				return err
			}
		}
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
			[]apiDoOpt{
				apiExpectStatus(http.StatusOK),
				apiExpectJSONBody(&tl),
			},
			[]apiDoOpt{
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
func apiWithOr(optLists ...[]apiDoOpt) apiDoOpt {
	return apiDoOpt{
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

func apiWithBody(body io.ReadCloser) apiDoOpt {
	return apiDoOpt{
		reqFn: func(req *http.Request) error {
			req.Body = body
			return nil
		},
	}
}

func apiExpectHeaders(h *http.Header) apiDoOpt {
	return apiDoOpt{
		respFn: func(resp *http.Response) error {
			*h = resp.Header
			return nil
		},
	}
}

func apiExpectStatus(statusCodes ...int) apiDoOpt {
	return apiDoOpt{
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

func apiExpectJSONBody(data any) apiDoOpt {
	return apiDoOpt{
		respFn: func(resp *http.Response) error {
			return json.NewDecoder(resp.Body).Decode(data)
		},
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
	return "", fmt.Errorf("failed to parse auth header: %s", header)
}

var (
	authHeaderMatcher = regexp.MustCompile("(?i).*(bearer|basic).*")
	authParamsMatcher = regexp.MustCompile(`([a-zA-z]+)="(.+?)"`)
)

func parseAuthHeader(header string) (authHeader, error) {
	// TODO: replace with a better parser, quotes should be optional, get character set from upstream http rfc
	var parsed authHeader
	parsed.Type = authHeaderMatcher.ReplaceAllString(header, "$1")
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
