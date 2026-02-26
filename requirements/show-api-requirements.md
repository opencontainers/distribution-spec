# Show/GetInfo API Requirements

Registries provide the basic means to push, pull, discover and manage content. While push and pull are generally feature-rich and standardized, the discover and manage APIs are extremely limited. Each registry project and operator have had to implement discover (list), show (get-info) and manage APIs, and each of them are unique to that particular registry.

To foster consistency across registries, enabling the community to build tooling that works consistently across all [OCI distribution-spec conformant registries][oci-compliance], we propose a common set of `Show` APIs.

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

1. A user can get info on a specific repository within a given registry, within a specified org and/or namespace.  
    - `oci-reg show acme-rockets.io/hello-world`
2. A user can get info on a specific tag, within a namespace.
    - `oci-reg show acme-rockets.io/hello-world:latest`
3. A user can get info on a specific digest, within a namespace.
    - `oci-reg show acme-rockets.io/org2/team-a/image-a1@sha256:v1qrst...`

## Data Returned

The type of data returned by the `show` API would need to support common elements, consistent across all OCI Conformant registries, and registry specific content. Whether we use annotations and namespaces to provide globally unique names is deferred from the requirements. Rather we wish to capture the types of data returned to understand what types of data structures we must support.

Common elements, likely defined within an `oci.distribution` namespace:
|Property Name|Description|
|-|-|
| repoName | Name of the repository |
| dateCreated | The date the repository was created |
| manifestCount | The total number of manifests currently in the repository |
| tagCount | The total number of tags. This number would `<=` manifestCount as not all manifests are tagged. |

Each registry likely has additional meta-data they wish to surface. Some are captured here for examples of extensibility, or possible consideration to a common set of elements. 

> **Note:** See an early proposal on a an [OCI Registry Meta-data][meta-data-service] enhancement.

|Property Name|Description|
|-|-|
| deleteEnabled | Is delete enabled for the requesting users context |
| listEnabled | Is listing content enabled for the requesting users context |
| readEnabled | Is content read enabled for the requesting users context |
| writeEnabled | Is content write enabled for the requesting users context |
| totalPullCount | The total pull count of content within the repository |

[oci-compliance]:            https://github.com/opencontainers/oci-conformance/tree/master/distribution-spec
[meta-data-service]:         https://github.com/SteveLasker/drafts/blob/meta-data-services/meta-data-services.md