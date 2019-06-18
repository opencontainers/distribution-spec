# OCI Artifacts

OCI distribution extends support from OCI images to additional artifacts. An artifact is a generic reference to an entity stored within a registry. An OCI or Docker image are specific types of artifacts. By extending OCI Distribution to support additional artifacts, developers and operational teams can leverage the investments in storing images in secured, compliant, managed registries. 

Developers can extend the pattern of building and pushing images to building and pushing additional artifact types. Operations can leverage the security and reliability of their registry for additional artifacts.

## Table of Contents

- [Introduction](#oci-artifacts)
- [Storing Artifacts](#storing-artifacts)
- [Defining Artifact Types](#defining-artifact-types)
- [Config Schema Validation](#config-schema-validation)
- [Artifact Layers](#artifact-layers)
  - [Artifact Layer Types](#artifact-layer-types)
  - [Example Layer Types](#example-layer-types)
- [Artifact Mapping Schema](#artifact-mapping-schema)
  - [Artifact Mapping Example](#artifact-mapping-example)
  - [Registering Artifact Types](#registering-artifact-types)
  - [Sample Artifacts](#sample-artifacts)
- [Authoring New Artifacts](#authoring-new-artifacts)

## Storing Artifacts

OCI artifacts are stored with the same infrastructure used to store OCI images. Manifests are used to define the artifact, with layers representing the files used to store and instantiate the artifact. Registries will store and serve artifacts, with the same semantics as they store and serve images. 

Compatibility between artifact types is a non-goal. Storage solution store .gif, .jpg, .mpg and hundreds of other file types. The underlying storage solution can store and serve files, with no expectation of the different client tools being able to read or write other file types. The compatibility is a decision of each artifact author.

## Defining Artifact Types

Registries and registry tools must understand the types of artifacts they represent. Registry scanning tools may only support a subset of artifact types, or they may need to apply different scanning methods based on the artifact type. 

If a security scanning solution were to scan all types, it would fail when it encounters types it doesn't understand, representing false negatives.  By differentiating types, a registry scanning solution can ignore unknown types, representing a known state. As new artifact types become well known, scanners can expand the types they offer, providing a more complete known state. 

Artifacts are defined by setting the `manifest.config.mediaType` to a globally unique value. The following format is used to differentiate the type of artifact:

`application/vnd.`[org|company]`.`[objectType]`.`[optionalSubType]`.config.`[version]`+json`

## Config Schema Validation

While the value of `manifest.config.mediaType` is used to determine the artifact type, the persistance of a config.json file is OPTIONAL. `manifest.config` schema verification is OPTIONAL. The [artifactMapping.json](#artifact-mapping-schema) specifies whether an artifact defines a config schema. 

Distribution instances MAY:

- parse and process the contents of  `manifest.config`, based on the provided schema of `manifest.config.mediaType`, offering additional information or actions. 
- ignore the contents and validation of the config.json file

## Artifact Layers

Artifacts are intended to have content. The content of an artifact is represented through one ore more [OCI Layers](https://github.com/opencontainers/image-spec/blob/master/layer.md). How the layers are constructed, and whether the layers are ordinal is a decision of the artifact author. 

As an example, OCI images are represented through an ordinal collection of compressed files. Each layer overlays the previous layer. 
Other artifacts may be represented by a single file, such as a markdown document, or a config file. Other artifacts may include single config file as one layer, and a collection of binaries, compressed as another layer. By separating the layers, the artifact author can benefit from layer de-duplication and concurrent downloading of the blobs which represent the layer. 

### Artifact Layer Types

Artifact layers are expected to be unique, per artifact author. A layer that represents an OCI Image is not expected to be shared with an Helm chart. To differentiate layers, artifact authors MAY create unique artifact layer types. 

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

## Artifact Mapping Schema

For a registry to understand specific artifacts, and optionally present information to users of the registry, an `artifactMapping.json` file is defined using the following format.

- **`mediaType`** *string*

  This REQUIRED property uniquely identifies the artifact for computer consumption. It may be owned by an org or a company and must be versioned.
  
  The format of `mediaType` uses the following format: `application/vnd.`[org|company]`.`[objectType]`.`[optionalSubType]`.config.`[version]`+json`

- **`title`** *string*

  This REQUIRED property represents the name of the type displayed for human consumption. The title may be displayed in a repository listing, or registry tooling.  The max length MUST not exceed 30 characters.

- **`description`** *string*

  This REQUIRED property provides a short description of the type for human consumption. The description may be displayed in repository listings or registry tooling. The max length MUST not exceed 255 characters and MUST not encode any formatting characters.

- **`url`** *url*

  This OPTIONAL property provides additional details, intended for consumers of the type. This is most often a marketing awareness overview page. 

- **`spec`** *url*

  This OPTIONAL property references a spec, providing additional details on the type. 

- **`tools`** *string-string map*

  This OPTIONAL property provides a collection of tools that may be used with artifact type. The property is intended for end users to find more info on how to find and install related tools. Registry operators MAY provide links to the tools in their repository listing. 

- **`tools-properties`** 

  - **`title`** *string*

    This REQUIRED property represents the name of the tool, displayed for human consumption. The title may be displayed in a repository listing, or registry tooling.  The max length MUST not exceed 30 characters.    

  - **`url`** *url*

    This REQUIRED property links to a page where users can download the tool. The URL MAY be a direct link to a download URL, or a link to documentation for how to download the tool.

- **`configSchemaReference`** *url*

  This OPTIONAL property provides a schema reference for the artifact config object. The schema is provided for registry operators and tools to optionally validate and process information within the config. A registry operator MAY wish to present information, such as the OCI image architecture type.

- **`layerMediaTypes`** string-string map

  This REQUIRED property represents one or more layer `mediaTypes`, defined by the artifact. Layer mediaTypes SHOULD be unique to the specific artifact. Layer mediaTypes are NOT REQUIRED to be unique across different artifact types when artifacts intend to share layers across different artifact tooling. 

  `layerMediaTypes` use the following format: `application/vnd.`[org|company]`.`[objectType]`.`[optionalSubType]`.layer.`[version]`+`[fileFormat]. 
  Layers are NOT required to use a specific format. The artifact author MAY choose to compress the file(s) within the layer, or store them natively. 

  - **`mediaType`** *string*

    This REQUIRED property represents a valid layer `mediaTypes`for the artifact.

### Artifact Mapping Example

The following is an example of the Helm artifact type.

```json
{
  "mediaType": "application/vnd.cncf.helm.config.v1+json",
  "title": "Helm Chart",
  "description": "The package manager for Kubernetes",
  "spec": "https://github.com/helm/helm/blob/master/oci-artifact-mediatypes",
  "url": "https://helm.sh/",
  "tools": [
    {
      "title": "Helm",
      "url": "https://github.com/helm/helm#install"
    }
  ],
  "configSchemaReference": "",
  "layerMediaTypes": [
    {
      "mediaType": "application/vnd.cncf.helm.chart.layer.v1.tar"
    }
  ]
}
```

### Registering Artifact Types

It is NOT a requirement for distribution implementations to support all artifact types. To ease in the awareness of new artifact types, artifact authors MAY publish their types by providing the following files under the [./artifactTypes](./artifactTypes) folder.

- Each new artifact and versioned definition is placed in a separate folder. 
- Each folder name MUST use the following convention: `vnd.`[org|company]`.`[objectType]`.`[optionalSubType]`.`[version]


The `artifactTypes` folder contains the following files:
|File|Content|
|-|-|
|`artifactMapping.json`|REQUIRED: Information that enables registries and artifact specific tooling to represent the artifact|
|`artifactConfigSchema.json`|OPTIONAL: Schema validation for the optional configuration. If a file is present, the config has schema. A missing `artifactConfigSchema.json` file states the artifact does not support configuration information.
|`logo.svg`|OPTIONAL: The artifact logo, in svg format, enabling distributions and tooling to associate the artifact logo. 

### Sample Artifacts

|Artifact|mediaType|Folder Name|
|-|-|-|
|[OCI Image](https://github.com/opencontainers/image-spec/)|`application/vnd.oci.image.config.v1+json`|[vnd.oci.image](./artifactTypes/vnd.oci.image.1/)
|[Helm Chart](https://helm.sh)|`application/vnd.cncf.helm.chart.config.v1+json`|[vnd.cncf.helm.chart](./artifactTypes/vnd.cncf.helm.chart.1/)|
|[Singularity Images](https://www.sylabs.io/singularity/), by Sylabs|`application/vnd.sylabs.sif.config.v1+json`|[vnd.sylabs.sif](./artifactTypes/vnd.sylabs.sif.1/)

## Authoring New Artifacts

Authoring a new artifact involves the following steps:

1. Define the [artifact type](#defining-artifact-types), by identifying a unique `manifest.config.mediaType`
1. Define the [layer types](#artifact-layers)
1. OPTIONAL: [define a config schema](#config-schema-validation)
1. OPTIONAL: [register the artifact information](#registering-artifact-types)