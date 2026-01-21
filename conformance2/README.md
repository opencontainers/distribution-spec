# OCI Distribution Spec Conformance Test

The distribution-spec conformance test is used to verify the various HTTP endpoints on a registry generate the appropriate responses and handle different types of data.

## Configuration

The test is configured by either a yaml configuration file or environment variables.
When a setting is configured by multiple sources, the precedence from highest to lowest is the environment variable, then yaml configuration file, and lastly any legacy environment variables.

Most registries can be tested by setting the registry, repository, and login credentials.
For APIs with a valid unsupported response code, attempts are made to track the missing feature without needing to manually disable the test.

### Environment Variables

Environment variables can be used to set any configuration setting in the conformance test.
The available variables and their default values are listed here:

```shell
# several variables are used to configure the overall conformance test process
export OCI_CONFIGURATION="oci-conformance.yaml" # see Yaml Configuration File below
export OCI_RESULTS_DIR="./results" # output of the conformance test will be written here, see Results below
export OCI_VERSION="1.1" # distribution-spec version to test against, this adjusts default values for the API tests
export OCI_LOG=warn # adjust logging threshold: debug, info, warn, error (this does not affect the generated reports)

# the registry settings typically need to be configured
export OCI_REGISTRY="localhost:5000"
export OCI_TLS="enabled" # enabled (https), insecure (self signed), or disabled (http)
export OCI_REPO1="conformance/repo1"
export OCI_REPO2="conformance/repo2"
export OCI_USERNAME=
export OCI_PASSWORD=

# API settings with their 1.1 default values can be used to skip specific requests
export OCI_API_PULL=true
export OCI_API_PUSH=true # to disable push requests, see the OCI_RO_DATA variables below
export OCI_API_BLOBS_ATOMIC=true # whether blob delete operations should be immediate
export OCI_API_BLOBS_DELETE=true
export OCI_API_BLOBS_MOUNT_ANONYMOUS=true # attempt to mount a blob without a source repository
export OCI_API_BLOBS_UPLOAD_CANCEL=false # cancel a running upload
export OCI_API_MANIFESTS_ATOMIC=true # whether manifest delete operations should be immediate
export OCI_API_MANIFESTS_DELETE=true
export OCI_API_TAGS_ATOMIC=true # whether tag delete operations should be immediate
export OCI_API_TAGS_DELETE=true
export OCI_API_TAGS_LIST=true
export OCI_API_REFERRER=true

# Data settings are used to generate a variety of OCI content
export OCI_DATA_IMAGE=true # note, this must be left enabled for any tests to run
export OCI_DATA_INDEX=true
export OCI_DATA_INDEX_LIST=true # an index containing a nested index
export OCI_DATA_SPARSE=false # manifest where some descriptors have not been pushed
export OCI_DATA_ARTIFACT=true # an OCI artifact packaged as an image with an artifactType
export OCI_DATA_SUBJECT=true # an OCI image with the subject field defined
export OCI_DATA_SUBJECT_MISSING=true # pushes content with a subject referencing a non-existent digest
export OCI_DATA_ARTIFACT_LIST=true # an OCI index with an artifactType
export OCI_DATA_SUBJECT_LIST=true # an OCI index with the subject field defined
export OCI_DATA_DATA_FIELD=true # descriptors with the data field populated
export OCI_DATA_NONDISTRIBUTABLE=true # an OCI image containing nondistributable layer references that have not been pushed
export OCI_DATA_CUSTOM_FIELDS=true # manifests and config json with additional fields
export OCI_DATA_EMPTY_BLOB=true # zero byte blob
export OCI_DATA_SHA512=true # content pushed using the sha512 digest algorithm

# For testing read-only registries, images must be preloaded.
# OCI_API_PUSH=false must be set, and disabling DELETE APIs is recommended.
# All requests are performed against the OCI_REPO1 repository.
export OCI_RO_DATA_TAGS= # space separated list of tags
export OCI_RO_DATA_MANIFESTS= # space separated list of manifest digests
export OCI_RO_DATA_BLOBS= # space separated list of blob digests
export OCI_RO_DATA_REFERRERS= # space separated list of subject digests for the referrers API

# other settings
export OCI_FILTER_TEST= # used to filter a specific branch of tests in, e.g. "OCI Conformance Test/sha256 blobs"
```

### Yaml Configuration File

The conformance test will load `oci-conformance.yaml` by default, which can be configured with the `OCI_CONFIGURATION` environment variable.

The default yaml configuration is shown below and matches the environment variables described above:

```yaml
resultsDir: ./results
version: "1.1"
registry: localhost:5000
tls: enabled
repo1: conformance/repo1
repo2: conformance/repo2
username: ""
password: ""
logging: warn
filterTest: ""
apis:
  pull: true
  push: true
  blobs:
    atomic: true
    delete: true
    mountAnonymous: true
    uploadCancel: false
  manifests:
    atomic: true
    delete: true
  tags:
    atomic: true
    delete: true
    list: true
  referrer: true
data:
  image: true
  index: true
  indexList: true
  sparse: false
  artifact: true
  subject: true
  subjectMissing: true
  artifactList: true
  subjectList: true
  dataField: true
  nondistributable: true
  customFields: true
  emptyBlob: true
  sha512: true
roData:
  tags: []
  manifests: []
  blobs: []
  referrers: []
```

## Running the Test

The test is available to be run with Go, Docker, or GitHub Actions.

### Go

The tests require Go 1.24 or greater.

They can be run directly with:

```shell
go run -buildvcs=true .
```

Or to compile and run separately:

```shell
go build -o conformance .
./conformance
```

### Docker

First configure the test with environment variables or a configuration file as described above.
Then build and run the conformance test using a command similar to below:

```shell
docker build -t conformance .
docker run -it --rm --net=host \
  -u "$(id -u):$(id -g)" \
  -v "$(pwd)/results:/results" \
  -e OCI_REGISTRY -e OCI_TLS -e OCI_REPO1 -e OCI_REPO2 -e OCI_USERNAME -e OCI_PASSWORD -e OCI_VERSION \
  conformance:latest
```

Additional environment variables can be specified as needed, or the `oci-conformance.yaml` file can be passed as a volume, mounted at `/oci-conformance.yaml` inside the container.

### GitHub Actions

TODO

## Results

A summary of the test is output to the screen along with any logging.
The results directory (`results` by default) is populated with the following files:

- `report.html`: Full report of the test, including censored output of each request and response.
- `junit.xml`: JUnit report.
