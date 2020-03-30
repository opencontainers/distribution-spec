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

The tests are broken down into 4 major categories:

1. Pull
2. Push
3. Discovery
4. Management

In addition, each category has its own setup and teardown processes where appropriate.

##### Pull

The Pull tests validate that content can be retrieved from a registry.

These tests are *always* run, as this is the baseline for registry conformance.

Regardless of whether the Push tests are enabled, as part of setup for the Pull tests,
content will be uploaded to the registry.
If you wish to prevent this, you can set the following environment variables pointing
to content already present in the registry:

```
# Optional: set to prevent automatic setup
OCI_MANIFEST_DIGEST=<digest>
OCI_TAG_NAME=<tag>
OCI_BLOB_DIGEST=<digest>
```

##### Push

The Push tests validate that content can be uploaded to a registry.

To enable the Push tests, you must explicitly set the following in the environment:

```
# Required to enable
OCI_TEST_PUSH=1
```

##### Discovery

The Discovery tests validate that the contents of a registry can be discovered.

To enable the Discovery tests, you must explicitly set the following in the environment:

```
# Required to enable
OCI_TEST_DISCOVERY=1
```

As part of setup of these tests, a manifest and associated tags will be pushed to the registry.
If you wish to prevent this, you can set the following environment variable pointing
to list of tags to be returned from `GET /v2/<name>/tags/list`:

```
# Optional: set to prevent automatic setup
OCI_TAG_LIST=<tag1>,<tag2>,<tag3>,<tag4>
```

##### Management

The Management tests validate that the contents of a registry can be deleted or otherwise modified.

To enable the Management tests, you must explicitly set the following in the environment:

```
# Required to enable
OCI_TEST_MANAGEMENT=1
```

Note: The Management tests explicitly depend upon the Push and Discovery tests, as there is no
way to test content management without also supporting push and discovery.

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
