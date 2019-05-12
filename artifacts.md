# OCI Artifacts

In the same way distribution can support OCI Images, distribution has been extended to support additional artifacts.

## Artifact Types

In the same way a file system understands the types of files it stores, registries and registry tools must understand the types of artifacts they may interact with. For example, registry scanning tools may only support a subset of artifact types, or they may need to apply different scanning methods based on the artifact type. 

Artifacts are defined by setting the `manifest.config.mediaType` to a globally unique value, using hte following format:

`application/vnd.`[org|company]`.`[objectType]`.`[optional sub type]`.config.`[version]`+json`

|Artifact|mediaType|
|-|-|
|OCI Image|`application/vnd.oci.image.config.v1+json`|
|[Helm Chart](https://helm.sh|`application/vnd.cncf.helm.chart.config.v1+json`|
|[Singularity Images](https://www.sylabs.io/singularity/), by Sylabs|`application/vnd.sylabs.sif.config.v1+json`|

## Registering Artifact Types

It is NOT a requirement for distribution implementations to support all artifact types. To ease in the awareness of new artifact types, artifact authors MAY define their types by contributing to [artifact-media-types.json](./artifacts/artifact-media-types.json).

## Config Schema Validation

While `manifest.config.mediaType` is used to determine the artifact type, the persistance of a config.json file is OPTIONAL. `manifest.config` schema verification is OPTIONAL. The [artifact-media-types.json](./artifacts/artifact-media-types.json) specifies whether an artifact defines a config schema. 