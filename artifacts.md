# OCI Artifacts

OCI distribution extends support from OCI images to additional artifacts. An artifact is a generic reference to an entity stored within a registry. An OCI or Docker image are specific types of artifacts. By extending OCI Distribution to support additional artifacts, developers and operational teams can leverage the investments in storing images in secured, compliant, managed platform registries of cloud providers and on-prem or vendor offerings that include OCI Distribution. 

Developers can extend the familiar pattern of building and pushing images to additional artifact types. Operations can leverage the security put in place to secure connectivity between their registries and production environments for additional artifacts.

## Storing Artifacts

OCI Artifacts are stored with the same infrastructure used to store OCI Images. However, just as different applications use different file formats, OCI artifacts are not intended to be compatible across different tools. The compatibility is a decision of each artifact author.

## Artifact Types

In the same way a file system understands the types of files it stores, registries and registry tools must understand the types of artifacts they may interact with. Registry scanning tools may only support a subset of artifact types, or they may need to apply different scanning methods based on the artifact type. 

Artifacts are defined by setting the `manifest.config.mediaType` to a globally unique value. The following format is used to differentiate the type of artifact:

`application/vnd.`[org|company]`.`[objectType]`.`[optional sub type]`.config.`[version]`+json`

### Sample Artifacts

|Artifact|mediaType|Folder Name|
|-|-|-|
|[OCI Image](https://github.com/opencontainers/image-spec/)|`application/vnd.oci.image.config.v1+json`|[vnd.oci.image](./artifactTypes/vnd.oci.image/)
|[Helm Chart](https://helm.sh)|`application/vnd.cncf.helm.chart.config.v1+json`|[vnd.cncf.helm.chart](./artifactTypes/vnd.cncf.helm.chart/)|
|[Singularity Images](https://www.sylabs.io/singularity/), by Sylabs|`application/vnd.sylabs.sif.config.v1+json`|[vnd.sylabs.sif](./artifactTypes/vnd.sylabs.sif/)

## Registering Artifact Types

It is NOT a requirement for distribution implementations to support all artifact types. To ease in the awareness of new artifact types, artifact authors MAY define their types by providing providing the following files under the [./artifactTypes](./artifactTypes) folder.

- Each new artifact definition is placed in a separate folder. 
- Each folder follows a subset of the `manifest.config.mediaType` definition: `vnd.`[org|company]`.`[objectType]`.`[optional sub type]`.config.`[version]


`artifactTypes` folder contains:
|File|Content|
|-|-|
|`artifactMapping.json`|REQUIRED: Information that enables registries and artifact specific tooling to represent the artifact|
|`artifactConfigSchema.json`|OPTIONAL: Schema validation for the optional configuration. If a file is present, the config has schema. A missing `artifactConfigSchema.json` file states the artifact does not support configuration information.
|`logo.svg`|OPTIONAL: The artifact logo, in svg format, enabling distributions and tooling to associate the artifact logo. 

## Config Schema Validation

While the value of `manifest.config.mediaType` is used to determine the artifact type, the persistance of a config.json file is OPTIONAL. `manifest.config` schema verification is OPTIONAL. The [artifact-media-types.json](./artifacts/artifact-media-types.json) specifies whether an artifact defines a config schema. 

Distribution instances MAY:

- parse and process the contents of  `manifest.config`, based on the provided schema of `manifest.config.mediaType`, offering additional information or actions. 
- ignore the contents and validation of the config.json file

## Artifact Layers

Artifacts are intended to have content. The content of an artifact is represented through one ore more [OCI Layers](https://github.com/opencontainers/image-spec/blob/master/layer.md). How the layers are constructed, and whether the layers are ordinal is a decision of the artifact author. 

OCI Images are represented through an ordinal collection of compressed files. Each layer overlays the previous layer. 
Other artifacts may be represented by a single file, such as a markdown document, or a config file. Other artifacts may include single config file as one layer, and a collection of binaries, compressed as another layer. By separating the layers, the artifact author can benefit from layer de-duplication and concurrent downloading of the blobs which represent the layer. 

## Artifact Layer Types

Artifact layers are expected to be unique, per artifact author. A layer that represents an OCI Image is not expected to be shared with an Helm chart. To differentiate layers, artifact authors MUST create unique artifact layer types. 

Artifact layer types utilize the `layer.mediaType` with the following format:
`application/vnd.`[org|company]`.`[layerType]`.`[layerSubType]`.layer.`[version]`+`[fileFormat]

### Example Layer Types

|Artifact Type|mediaType|
|-|-|
|OCI Image|`application/vnd.oci.image.layer.v1+tar`|
|Docker Image|`application/vnd.docker.image.rootfs.diff.tar.gzip`|
|Helm Chart|`application/vnd.cncf.helm.chart.layer.v1+tar`|
|Helm Chart Meta Layer|`application/vnd.cncf.helm.chart.meta.layer.v1+json`|
|Singularity SIF Layer|`application/vnd.sylabs.sif.layer.v1+tar`|

## Authoring New Artifacts

Authoring a new artifact involves the following steps:

1. Define the [artifact type](#artifact-types), by identifying a unique `manifest.config.mediaType`
1. Define the [layer types](#artifact-layer-types)
1. OPTIONAL: [define a config schema](#config-schema-validation)
1. OPTIONAL: [register the artifact information](#registering-artifact-types)