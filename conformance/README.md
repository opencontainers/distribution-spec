## Conformance Tests

### How to Run

#### Binary

Requires Go 1.17+.

In this directory, build the test binary:
```
go test -c
```

This will produce an executable at `conformance.test`.

Next, set environment variables with your registry details:
```
# Registry details
export OCI_ROOT_URL="https://r.myreg.io"
export OCI_NAMESPACE="myorg/myrepo"
export OCI_CROSSMOUNT_NAMESPACE="myorg/other"
export OCI_USERNAME="myuser"
export OCI_PASSWORD="mypass"

# Which workflows to run
export OCI_TEST_PULL=1
export OCI_TEST_PUSH=1
export OCI_TEST_CONTENT_DISCOVERY=1
export OCI_TEST_CONTENT_MANAGEMENT=1

# Extra settings
export OCI_HIDE_SKIPPED_WORKFLOWS=0
export OCI_DEBUG=0
export OCI_DELETE_MANIFEST_BEFORE_BLOBS=0 # defaults to OCI_DELETE_MANIFEST_BEFORE_BLOBS=1 if not set
```

Lastly, run the tests:
```
./conformance.test
```

Note: for some registries, you may need to create `OCI_NAMESPACE` ahead of time.

This will produce `junit.xml` and `report.html` in the current directory with the results. To choose an alternative directory:

```
export OCI_REPORT_DIR=/alternative/directory
```

To disable writing of the result files:

```
export OCI_REPORT_DIR=none
```

#### Testing registry workflows

The tests are broken down into 4 major categories:

1. Pull - Highest priority - All OCI registries MUST support pulling OCI container
images.
2. Push - Registries need a way to get content to be pulled, but clients can/should
be more forgiving here. For example, if needing to fallback after an unsupported endpoint.
3. Content Discovery - Includes tag listing (and possibly search in the future).
4. Content Management - Lowest Priority - Includes tag, blob, and repo deletion.
(Note: Many registries may have other ways to accomplish this than the OCI API.)

In addition, each category has its own setup and teardown processes where appropriate.

##### Pull

The Pull tests validate that content can be retrieved from a registry.

These tests are run when the following is set in the environment:
```
OCI_TEST_PULL=1
```

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

Some registries may require a workaround for Authorization during the push flow. To set your own scope, set the following in the environment:

```
# Set the auth scope
OCI_AUTH_SCOPE="repository:mystuff/myrepo:pull,push"
```

Most registries currently require at least one layer to be uploaded (and referenced in the appropriate section of the manifest)
before a manifest upload will succeed. By default, the push tests will attempt to push two manifests: one with a single layer,
and another with no layers. If the empty-layer test is causing a failure, it can be skipped by setting the following in the
environment:

```
# Enable layer upload
OCI_SKIP_EMPTY_LAYER_PUSH_TEST=1
```

The test suite will need access to a second namespace. This namespace is used to check support for cross-repository mounting
of blobs, and may need to be configured on the server-side in advance. It is specified by setting the following in
the environment:

```
# The destination repository for cross-repository mounting:
OCI_CROSSMOUNT_NAMESPACE="myorg/other"
```

If you want to test the behaviour of automatic content discovery, you should set the `OCI_AUTOMATIC_CROSSMOUNT` variable.

```
# Do not test automatic cross mounting
unset OCI_AUTOMATIC_CROSSMOUNT

# Test that automatic cross mounting is working as expected
OCI_AUTOMATIC_CROSSMOUNT=1

# Test that automatic cross mounting is disabled
OCI_AUTOMATIC_CROSSMOUNT=0
```

##### Content Discovery

The Content Discovery tests validate that the contents of a registry can be discovered.

To enable the Content Discovery tests, you must explicitly set the following in the environment:

```
# Required to enable
OCI_TEST_CONTENT_DISCOVERY=1
```

As part of setup of these tests, a manifest and associated tags will be pushed to the registry.
If you wish to prevent this, you can set the following environment variable pointing
to list of tags to be returned from `GET /v2/<name>/tags/list`:

```
# Optional: set to prevent automatic setup
OCI_TAG_LIST=<tag1>,<tag2>,<tag3>,<tag4>
```

##### Content Management

The Content Management tests validate that the contents of a registry can be deleted or otherwise modified.

To enable the Content Management tests, you must explicitly set the following in the environment:

```
# Required to enable
OCI_TEST_CONTENT_MANAGEMENT=1
```

Note: The Content Management tests explicitly depend upon the Push and Content Discovery tests, as there is no
way to test content management without also supporting push and content discovery.

#### HTML Report
By default, the HTML report will show tests from all workflows. To hide workflows that have been disabled from
the report, you must set the following in the environment:

```
# Required to hide disabled workflows
OCI_HIDE_SKIPPED_WORKFLOWS=1
```

#### Teardown Order

By default, the teardown phase of each test deletes blobs before manifests. Some registries require the opposite order, deleting manifests before blobs. In this case, you must set the following in the environment:

```
# Required to delete manifests before blobs
OCI_DELETE_MANIFEST_BEFORE_BLOBS=1
```

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
  -e OCI_TEST_PULL=1 \
  -e OCI_TEST_PUSH=1 \
  -e OCI_TEST_CONTENT_DISCOVERY=1 \
  -e OCI_TEST_CONTENT_MANAGEMENT=1 \
  -e OCI_HIDE_SKIPPED_WORKFLOWS=0 \
  -e OCI_DEBUG=0 \
  -e OCI_DELETE_MANIFEST_BEFORE_BLOBS=0 \
  conformance:latest
```

This will create a local `results/` directory containing all of the test report files.

#### GitHub Action

A GitHub Action is provided by this repo which you can use
as part of a GitHub-based CI pipeline.

The following example will build the binary off of the main branch,
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
        uses: opencontainers/distribution-spec@main
        # you can also run against a specific tag or commit instead
        # uses: opencontainers/distribution-spec@v1.1.0
        env:
          OCI_ROOT_URL: https://myreg.io
          OCI_NAMESPACE: mytestorg/mytestrepo
          OCI_USERNAME: ${{ secrets.MY_REGISTRY_USERNAME }}
          OCI_PASSWORD: ${{ secrets.MY_REGISTRY_PASSWORD }}
          OCI_TEST_PULL: 1
          OCI_TEST_PUSH: 1
          OCI_TEST_CONTENT_DISCOVERY: 1
          OCI_TEST_CONTENT_MANAGEMENT: 1
          OCI_HIDE_SKIPPED_WORKFLOWS: 0
          OCI_DEBUG: 0
          OCI_DELETE_MANIFEST_BEFORE_BLOBS: 0
```

You can also add a badge pointing to list of runs for this action using the following markdown:

```
[![](https://github.com/<org>/<repo>/workflows/oci-distribution-conformance/badge.svg)](https://github.com/<org>/<repo>/actions?query=workflow%3Aoci-distribution-conformance)
```

(replacing `<org>` and `<repo>` with your GitHub repo details).
