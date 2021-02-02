# Listing API Requirements

Registries provide the basic means to push, pull, discover and manage content. While push and pull are generally feature-rich and standardized, the discover and manage APIs are extremely limited. Each registry project and operator have had to implement discover (list), show (get-info) and manage APIs, and each of them are unique to that particular registry.

To foster consistency across registries, enabling the community to build tooling that works consistently across all [OCI distribution-spec conformant registries][oci-compliance], we propose a common set of listing APIs.

To facilitate discussions, prioritization and phased implementations, we will capture a list of scenarios and requirements. Once the requirements are felt complete and prioritized, we can transition to a design that implements the agreed prioritized items, while reserving space for the future enhancements.

For the purposes of examples within a scenario, we'll assume an `oci-reg` cli exists. The purpose is not to design the specific CLI, rather provide concrete examples whose specific names and parameters will evolve.

## Scenarios

For illustrative purposes, consider a registry with the following content. A repository may have a collection of tags, with associated digests, and untagged content. The untagged content may represent an older version of tagged content, or enhancing content like Notary v2, SBoM, GPL source artifacts.

```bash
acme-rockets.io/
  +-hello-world:
  | |─latest  @sha256:latestaaa...
  | |─v1      @sha256:latestaaa...
  | └───      @sha256:12sab3121...
  +-org1/
  | └─artifact-a:
  |   |─v1    @sha256:v1abcd...
  |   |─v2    @sha256:v2defg...
  |   |───    @sha256:129a52...
  |   └───    @sha256:129a52...
  +-org2/
  | |─team-a/
  | | └─image-a1:
  | | | |─v1    @sha256:v1qrst...
  | | | |─v2    @sha256:v1fedc...
  | | | └───    @sha256:ag2ad2...
  | | └─image-a2:
  | |   └─v1    @sha256:v1s23a...
  | └─team-b/
  |   └─image-b1:
  |     |─v1    @sha256:v1qrst...
  |     |─v2    @sha256:v1fedc...
  |     └───    @sha256:ag2ad2...
  +-org3/
  | └─team-c/
  |   └─image-c1:
  |     |─v1   @sha256:v1f13is...
  |     └───   @sha256:210a113...
```

1. A user can get a list of repositories within a given registry, within a specified org and/or namespace.  Providing a listing under a given namespace enables registries that configure orgs within a namespace and registries that support nested permission of repositories.
    - `oci-reg list acme-rockets.io/`
    - `oci-reg list acme-rockets.io/org1/`
    - `oci-reg list acme-rockets.io/org2/team-a/`
2. A user can get a list of tags within a given registry/namespace.
    - `oci-reg list-tags acme-rockets.io/org2/team-a/image-a1`
3. A user can get a list of manifests, within a given registry/namespace.
    - `oci-reg list-manifestgs acme-rockets.io/org2/team-a/image-a1`
4. A user can get a list of artifacts that reference a specified digest. Example: (a list of Notary v2 signatures)
    - `oci-reg list-references acme-rockets.io/org2/team-a/image-a1@sha256:5c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c82`
5. A user can get a list of tags for a specified digest.
    - `oci-reg list-tags acme-rockets.io/org2/team-a/image-a1@sha256:5c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c82`

The above `oci-reg` cli should not impose a specific design of an api, rather it visualizes the possibility. For example, we may decide the `list-tags` api may return a collection of manifests and tags. What we want to capture here is what types of scenarios and data we want to capture.

```json
oci-reg list-tags acme-rockets.io/org2/team-a/image-a1
[
  {
    "digest": "sha256:4feab8c572ec8e20a3b89cd58d13212d0658d424febfd4b0460bb1a0b0cba452",
    "tags": [
      "ca4qj"
    ],
    "timestamp": "2019-11-02T21:17:26.1501916Z"
  },
  {
    "digest": "sha256:6660dacb106f6a812c662588d3dd6aed6537aa2181f760f4c13e2551a2ff027b",
    "tags": [
      "9-alpine",
      "ca4qm"
    ],
    "timestamp": "2019-11-02T21:21:27.8339591Z"
  }
]
```

## Requirements

1. Provide paging of results
2. Provide for registry specific extensions
3. Provide filtering by `artifactType`
4. Provide filtering by annotation
5. Provide sorting

## Prioritization

To deliver on an incremental phased delivery, the following requirements and scenarios are identified for the following phases

### Phase 1

### Phase 2

### Backlog

[oci-compliance]:            https://github.com/opencontainers/oci-conformance/tree/master/distribution-spec
