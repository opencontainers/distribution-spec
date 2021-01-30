# Listing API Requirements

Registries provide the basic means to push, pull, discover and manage content. While push and pull are generally feature-rich and standardized, the discover and manage APIs are extremely limited. Each registry project and operator have had to implement discover (list) and manage APIs, and each of them are unique to that particular registry.

To foster consistency across registries, enabling the community to build tooling that works consistently across all [OCI distribution-spec conformant registries][oci-compliance], we propose a common set of listing APIs.

To facilitate discussions, prioritization and phased implementations, we will capture a list of scenarios and requirements. Once the requirements are felt complete and prioritized, we can transition to a design that implements the agreed prioritized items, while reserving space for the future enhancements.

For the purposes of examples within a scenario, we'll assume an `oci-reg` cli exists. The purpose is not to design the specific CLI, rather provide concrete examples whose specific names and parameters will evolve.

## Scenarios

1. A user can get a list of repositories within a given registry, within a specified org and/or namespace.  
      - `oci-reg list acme-rockets.io/`
      - `oci-reg list acme-rockets.io/org1/`
2. A user can get a list of tags within a given registry/namespace.
3. A user can get a list of manifests, within a given registry/namespace.
4. A user can get a list of artifacts that reference a specified digest. Example: (a list of Notary v2 signatures)
5. A user can get a list of tags for a specified digest.

## Requirements

1. Provide paging of results
2. Provide for registry specific extensions.
3. Provide filtering by `artifactType`
4. Provide filtering by annotation
5. Provide sorting

## Prioritization

To deliver on an incremental phased delivery, the following requirements and scenarios are identified for the following phases

### Phase 1

### Phase 2

### Backlog


[oci-compliance]:            https://github.com/opencontainers/oci-conformance/tree/master/distribution-spec
