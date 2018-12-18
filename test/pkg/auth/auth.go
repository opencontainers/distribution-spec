// Copyright © 2018 ocicert authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	distp "github.com/opencontainers/distribution-spec/test/pkg/distp"
)

var (
	defaultScopeAccess string = "pull"
	DefaultRepoPrefix  string = "library/"

	defaultRegURL        string = "docker.io/busybox:latest"
	DefaultIndexURLPlain string = "registry-1.docker.io"
	dockerHostname       string = "docker.io"
	dockerV1Hostname     string = "index.docker.io"
)

type AuthScope struct {
	RemoteName string
	Actions    string
}

type RegAuthContext struct {
	Hclient    *http.Client
	RegURL     string
	ReqHost    string
	AuthTokens map[string]string

	Realm   string
	Service string
	Scope   AuthScope
}

type TokenStruct struct {
	Token string `json:"token"`
}

func init() {
	reg := os.Getenv("OCICERT_REGISTRY")
	if reg != "" {
		defaultRegURL = reg
	}
}

func NewRegAuthContext() RegAuthContext {
	return RegAuthContext{
		Hclient:    newHTTPClient(),
		RegURL:     defaultRegURL,
		AuthTokens: make(map[string]string),
		Scope: AuthScope{
			RemoteName: "",
			Actions:    "*",
		},
	}
}

// Get challenges from the index server, to be able to get necessary
// info like bearer realm, service, and scope, by parsing the www-authenticate
// header in the response.
func (sc *RegAuthContext) PrepareAuth(indexServer string) error {
	inputURL := "https://" + indexServer + "/v2/"

	req, res, err := sc.SendRequestWithToken(inputURL, "GET", nil)
	if err != nil {
		return fmt.Errorf("failed to send request to %s: %v", inputURL, err)
	}

	sc.ReqHost = req.URL.Host

	wwwAuthHdr := res.Header.Get("www-authenticate")
	if res.StatusCode != http.StatusUnauthorized || wwwAuthHdr == "" {
		return fmt.Errorf("received invalid result: %v", res)
	}

	tokens := strings.Split(wwwAuthHdr, ",")

	for _, token := range tokens {
		if strings.HasPrefix(strings.ToLower(token), "bearer realm") {
			sc.Realm = strings.Trim(token[len("bearer realm="):], "\"")
		}
		if strings.HasPrefix(token, "service") {
			sc.Service = strings.Trim(token[len("service="):], "\"")
		}
		if strings.HasPrefix(token, "scope") {
			sc.Scope = parseScope(strings.Trim(token[len("scope="):], "\""))
		}
	}

	if sc.Realm == "" {
		return fmt.Errorf("missing realm in bearer with challenge")
	}

	if sc.Service == "" {
		return fmt.Errorf("missing service in bearer with challenge")
	}

	return sc.getAuthToken(inputURL)
}

// Get auth token from the token server.
// For example it's equivalent to:
//
// $ curl "https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/busybox:pull"
//
func (sc *RegAuthContext) getAuthToken(inputURL string) error {
	authReq, err := http.NewRequest("GET", sc.Realm, nil)
	if err != nil {
		return fmt.Errorf("cannot send HTTP request to %s: %v", sc.Realm, err)
	}

	getParams := authReq.URL.Query()
	getParams.Add("service", sc.Service)
	if sc.Scope.Actions != "" {
		getParams.Add("scope", fmt.Sprintf("repository:%s:%s", sc.Scope.RemoteName, sc.Scope.Actions))
	}
	authReq.URL.RawQuery = getParams.Encode()

	res, err := sc.Hclient.Do(authReq)
	if err != nil {
		return fmt.Errorf("failed to send auth request: %v", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("unable to retrieve auth token: 401 unauthorized")
	case http.StatusOK:
		break
	default:
		return fmt.Errorf("statusCode = %v, request URL = %v", res.StatusCode, authReq.URL)
	}

	tokenBlob, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read token from body: %v", err)
	}

	var tokenStruct TokenStruct
	if err := json.Unmarshal(tokenBlob, &tokenStruct); err != nil {
		return fmt.Errorf("failed to unmarshal json for token: %v", err)
	}

	sc.AuthTokens[sc.ReqHost] = tokenStruct.Token

	if _, _, err := sc.SendRequestWithToken(inputURL, "GET", nil); err != nil {
		return fmt.Errorf("failed to send request to %s: %v", inputURL, err)
	}

	return nil
}

// Send an actual request with the auth token obtained in the previous step.
// For example it's equivalent to:
//
// $ curl -H "Authorization: Bearer TOKEN_STRING" https://index.docker.io/v2/library/busybox/manifests/latest
//
func (sc *RegAuthContext) SendRequestWithToken(inputURL, method string, body io.Reader) (*http.Request, *http.Response, error) {
	setBearerHeader := false

	req, err := http.NewRequest(method, inputURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request to %s: %v", inputURL, err)
	}

	authToken, ok := sc.AuthTokens[req.URL.Host]
	if ok {
		req.Header.Set("Authorization", "Bearer "+authToken)
		setBearerHeader = true
	}

	req.Header.Set(distp.DistAPIVersionKey, "registry/2.0")

	res, err := sc.Hclient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send auth request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized && setBearerHeader {
		return nil, nil, fmt.Errorf("received invalid result: %v", res)
	}

	return req, res, nil
}

// GetResponse sends an HTTP request with the given method, URL, and blob (optional).
// The parammeter acceptedStatus is a list of valid HTTP status codes expected,
// e.g., http.StatusOK. If the actual response does not have any status code in
// acceptedStatus, then it returns error.
func (sc *RegAuthContext) GetResponse(inputURL, method string, blobStream io.Reader, acceptedStatus []int) (*http.Response, error) {
	_, res, err := sc.SendRequestWithToken(inputURL, method, blobStream)
	if err != nil {
		return nil, fmt.Errorf("failed to send request with token to %s: %v", inputURL, err)
	}

	isInAccepted := func(keyStatus int, accptedStatus []int) bool {
		for _, a := range acceptedStatus {
			if a == keyStatus {
				return true
			}
		}
		return false
	}

	if !isInAccepted(res.StatusCode, acceptedStatus) {
		return nil, fmt.Errorf("got an unexpected reply from %s: %v", inputURL, res.StatusCode)
	}

	return res, nil
}

func newHTTPClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	tr := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                dialer.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	return &http.Client{
		Transport: tr,
	}
}

func parseScope(inputScope string) AuthScope {
	outScope := AuthScope{}
	scopeList := strings.Split(inputScope, ":")
	if len(scopeList) >= 3 {
		outScope.RemoteName = scopeList[1]
		outScope.Actions = scopeList[2]
	}

	return outScope
}

// GetIndexName returns the index server from a registry URL.
func GetIndexName(regURL string) string {
	index, _ := SplitReposName(regURL)
	return index
}

// SplitReposName breaks a repo name into an index name and remote name.
func SplitReposName(name string) (indexName, remoteName string) {
	i := strings.IndexRune(name, '/')
	if i == -1 || (!strings.ContainsAny(name[:i], ".:") && name[:i] != "localhost") {
		indexName, remoteName = DefaultIndexURLPlain, name
	} else {
		indexName, remoteName = name[:i], name[i+1:]
	}
	if indexName == DefaultIndexURLPlain && !strings.ContainsRune(remoteName, '/') {
		remoteName = DefaultRepoPrefix + remoteName
	}
	return
}

func GetIndexServer(inputURL string) string {
	indexServer := GetIndexName(inputURL)
	// we should use v1 hostname index.docker.io, because docker.io disabled
	// numerous endpoints.
	if indexServer == dockerHostname {
		indexServer = dockerV1Hostname
	}

	return indexServer
}
