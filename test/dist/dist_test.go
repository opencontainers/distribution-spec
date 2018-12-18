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

package dist

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/opencontainers/distribution-spec/test/pkg/auth"
	distp "github.com/opencontainers/distribution-spec/test/pkg/distp"
)

var (
	homeDir    string
	regAuthCtx auth.RegAuthContext

	testImageName string = "busybox"
	testRefName   string = "latest"
	regURL        string
)

func init() {
	homeDir = os.Getenv("HOME")

	regURL = regAuthCtx.RegURL
}

func TestCheckAPIVersion(t *testing.T) {
	reqPath := ""

	regAuthCtx := auth.NewRegAuthContext()
	regAuthCtx.Scope.RemoteName = reqPath
	regAuthCtx.Scope.Actions = "pull"

	indexServer := auth.GetIndexServer(regURL)

	if err := regAuthCtx.PrepareAuth(indexServer); err != nil {
		t.Fatalf("failed to prepare auth to %s for %s: %v", indexServer, reqPath, err)
	}

	inputURL := "https://" + indexServer + "/v2/" + reqPath

	res, err := regAuthCtx.GetResponse(inputURL, "GET", nil, []int{http.StatusOK})
	if err != nil {
		t.Fatalf("got an unexpected reply: %v", err)
	}

	if vers := res.Header.Get(distp.DistAPIVersionKey); vers != distp.DistAPIVersionValue {
		t.Fatalf("got an unexpected API version %v", vers)
	}
}

func TestPullManifest(t *testing.T) {
	indexServer := auth.GetIndexServer(regURL)

	remoteName := filepath.Join(auth.DefaultRepoPrefix, testImageName)
	reqPath := filepath.Join(remoteName, "manifests", testRefName)

	regAuthCtx := auth.NewRegAuthContext()
	regAuthCtx.Scope.RemoteName = remoteName
	regAuthCtx.Scope.Actions = "pull"

	if err := regAuthCtx.PrepareAuth(indexServer); err != nil {
		t.Fatalf("failed to prepare auth to %s for %s: %v", indexServer, reqPath, err)
	}

	inputURL := "https://" + indexServer + "/v2/" + reqPath

	if _, err := regAuthCtx.GetResponse(inputURL, "GET", nil, []int{http.StatusOK}); err != nil {
		t.Fatalf("got an unexpected reply: %v", err)
	}
}

func TestPushManifest(t *testing.T) {
	indexServer := auth.GetIndexServer(regURL)

	remoteName := filepath.Join(auth.DefaultRepoPrefix, testImageName)
	reqPath := filepath.Join(remoteName, "manifests", testRefName)

	regAuthCtx := auth.NewRegAuthContext()
	regAuthCtx.Scope.RemoteName = remoteName
	regAuthCtx.Scope.Actions = "push"

	if err := regAuthCtx.PrepareAuth(indexServer); err != nil {
		t.Fatalf("failed to prepare auth to %s for %s: %v", indexServer, reqPath, err)
	}

	inputURL := "https://" + indexServer + "/v2/" + reqPath

	if _, err := regAuthCtx.GetResponse(inputURL, "PUT", nil, []int{http.StatusOK}); err != nil {
		t.Fatalf("got an unexpected reply: %v", err)
	}
}
