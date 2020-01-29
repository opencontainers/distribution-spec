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

#### Container Image

You may use [Dockerfile.conformance](./../Dockerfile.conformance) located at the
root of this repo to build a container image that contains the test binary.

Example (using `docker`):
```
# build the image
(cd ../ && docker build -t conformance:latest -f Dockerfile.conformance .)

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
