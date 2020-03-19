## Conformance Tests

### How to Run

#### Binary

Requires Go 1.13+.

In this directory, build the test binary:
```
go test -c
```

This will produce an executable at `conformance.test`.

Next, set environment variables with your registry details:
```
export OCI_ROOT_URL="https://r.myreg.io"
export OCI_NAMESPACE="myorg/myrepo"
export OCI_USERNAME="myuser"
export OCI_PASSWORD="mypass"
export OCI_DEBUG="true"
```

Lastly, run the tests:
```
./conformance.test
```

This will produce `junit.xml` and `report.html` with the results.

Note: for some registries, you may need to create `OCI_NAMESPACE` ahead of time.

#### Testing registry workflows
##### Setup and teardown
Setup and teardown will be performed for each workflow that is tested. If actions taken during
test setup fail, the workflow tests will likely fail as well since they depend on a successful
setup. However, the setup may require actions outside of the scope of a given workflow. For example,
the setup for the pull workflow tests will attempt to use push actions, which the workflow tests will
later attempt to pull. If push is not supported by a registry, pull workflow failures can be avoided
by providing more information via environment variables. See the sections below for a list of environment
variables required for each workflow when other workflows are disabled.

##### Pull
With no additional environment variables set, the OCI tests will check for
conformance with the OCI spec on pull actions only. For convenience, the test setup will
attempt to push blobs and use those blobs for the pull actions. The teardown will
remove any blobs created by the setup. The setup and teardown are optional and 
do not directly affect test results. However, if the setup fails, subsequent pull
tests may fail since they will be looking for specific blobs and manifests. To 
prevent these cascading variables, a registry can supply a blob digest, manifest
digest, and manifest tag relating to blobs and manifests that are known to exist in the
repository. This information is supplied via the following environment variables:
```
with push disabled:
OCI_MANIFEST_DIGEST=
OCI_TAG_NAME=
OCI_BLOB_DIGEST=
```

##### Push
To enable testing for the push workflow, set `OCI_TEST_PUSH=1` prior to running the
tests. All tests pertaining to the push workflow will be disabled unless this variable
is set to specifically `1`. Aside from this, no additional environment variables are required
for the push workflow.

##### Discovery
Enable by setting `OCI_TEST_DISCOVERY=1`
The discovery workflow will test the `/v2/<name>/tags/list` endpoint, with different combinations
of query parameters to test content discovery. For more information on what should be supported for
the discovery flow, refer to the OCI distribution specification. The setup for these tests will
push a manifest with several associated tags. If push is disabled, supply a list of tags via the
`OCI_TAG_LIST` environment variable, with the tags separated by commas. For example,
`OCI_TAG_LIST=test0,test1,test2,test3`. It is expected that this is the exact list of tags that will
be found via a request to `/v2/<name>/tags/list`

```
with push disabled:
OCI_TAG_LIST=
```

##### Content Management (delete)
Enable by setting `OCI_TEST_CONTENT_MANAGEMENT=1`
The content management workflow explicitly depends upon the push and discovery workflows, and there is no
way to test content management without also supporting push and discovery. No addional variables are required
to test the content management workflow

#### Container Image

You may use the [Dockerfile](./Dockerfile) located in this directory
to build a container image that contains the test binary.

Example (using `docker`):
```
# build the image, using git SHA as the version
docker build -t conformance:latest \
    --build-arg VERSION=$(git log --format="%H" -n 1) .

# run the image
docker run --rm \
  -v $(pwd)/results:/results \
  -w /results \
  -e OCI_ROOT_URL="https://r.myreg.io" \
  -e OCI_NAMESPACE="myorg/myrepo" \
  -e OCI_USERNAME="myuser" \
  -e OCI_PASSWORD="mypass" \
  -e OCI_DEBUG="true" \
  conformance:latest
```

This will create a local `results/` directory containing all of the test report files.

#### GitHub Action

A GitHub Action is provided by this repo which you can use
as part of a GitHub-based CI pipeline.

The following example will build the binary off of the master branch,
run the tests, and upload `junit.xml` and `report.html` as build artifacts:

```yaml
# Place in repo at .github/workflows/oci-distribution-conformance.yml
name: oci-distribution-conformance
on: push
jobs:
  run:
    runs-on: ubuntu-latest
    steps:
      - name: Run OCI Distribution Spec conformance tests
        uses: opencontainers/distribution-spec@master
        env:
          OCI_ROOT_URL: https://myreg.io
          OCI_NAMESPACE: mytestorg/mytestrepo
          OCI_USERNAME: ${{ secrets.MY_REGISTRY_USERNAME }}
          OCI_PASSWORD: ${{ secrets.MY_REGISTRY_PASSWORD }}
      - run: mkdir -p .out/ && mv {report.html,junit.xml} .out/
        if: always()
      - name: Upload test results zip as build artifact
        uses: actions/upload-artifact@v1
        with:
          name: oci-test-results-${{ github.sha }}
          path: .out/
        if: always()
```

You can also add a badge pointing to list of runs for this action using the following markdown:

```
[![](https://github.com/<org>/<repo>/workflows/oci-distribution-conformance/badge.svg)](https://github.com/<org>/<repo>/actions?query=workflow%3Aoci-distribution-conformance)
```

(replacing `<org>` and `<repo>` with your GitHub repo details).
