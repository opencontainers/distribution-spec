// Copyright © 2018 Atlas Kerr atlaskerr@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


local common = import 'common.libsonnet';

local JSV7Schema = 'http://json-schema.org/draft-07/schema#';
local JSV7 = 'jsonschemaV7';
local OAPI3 = 'openapiV3';

// Descriptions for schema objects. Separated out to make code more readable.
local d = {
  user: 'The username or UID which is a platform-specific structure that allows specific control over which user the process run as. This acts as a default value to use when the value is not specified when creating a container. For Linux based systems, all of the following are valid: `user`, `uid`, `user:group`, `uid:gid`, `uid:group`, `user:gid`. If `group`/`gid` is not specified, the default group and supplementary groups of the given `user`/`uid` in `/etc/passwd` from the container are applied.',
  exposedPorts: 'A set of ports to expose from a container running this image. Its keys can be in the format of `port/tcp`, `port/udp`, or `port` with the default protocol being `tcp` if not specified. These values act as defaults and are merged with any specified when creating a container.',
  env: 'Entries are in the format of `VARNAME=VARVALUE`. These values act as defaults and are merged with any specified when creating a container.',
  entrypoint: 'A list of arguments to use as the command to execute when the container starts. These values act as defaults and may be replaced by an entrypoint specified when creating a container.',
  cmd: 'Default arguments to the entrypoint of the container. These values act as defaults and may be replaced by any specified when creating a container. If an `Entrypoint` value is not specified, then the first entry of the `Cmd` array SHOULD be interpreted as the executable to run.',
  volumes: 'A set of directories describing where the process is likely to write data specific to a container instance.',
  workingDir: 'Sets the current working directory of the entrypoint process in the container. This value acts as a default and may be replaced by a working directory specified when creating a container.',
  labels: 'The field contains arbitrary metadata for the container. This property MUST use the annotation rules.',
  stopSignal: 'The field contains the system call that will be sent to the container to exit. The signal can be a signal name in the format `SIGNAME`, for instance `SIGKILL` or `SIGRTMIN+3`',
  rootfsType: 'MUST be set to `layers`. Implementations MUST generate an error if they encounter an unknown value while verifying or unpacking an image.',
  diffIDs: 'An array of layer content hashes in order from first to last.',
  created: 'A combined date and time at which the layer was created, formatted as defined by RFC 3339, section 5.6.',
  author: 'The author of the build point.',
  imageAuthor: 'Gives the name and/or email address of the person or entity which created and is responsible for maintaining the image.',
  configAuthor: 'Gives the name and/or email address of the person or entity which created and is responsible for maintaining the image.',
  createdBy: 'The command which created the layer.',
  comment: 'A custom message set when creating the layer.',
  emptyLayer: "This field is used to mark if the history item created a filesystem diff. It is set to true if the history item doesn't correspond to an actual layer in the rootfs section.",
  arch: 'The CPU architecture which the binaries in the image are built to run on. Configurations SHOULD use, and implementations SHOULD understand, values listed in the Go Language documentation for `GOARCH`.',
  indexArch: 'This REQUIRED property specifies the CPU architecture. Image indexes SHOULD use, and implementations SHOULD understand, values listed in the Go Language document for `GOARCH`',
  os: 'The name of the operating system which the image is built to run on. Configurations SHOULD use, and implementations SHOULD understand, values listed in the Go Language document for `GOOS`.',
  configConfig: 'The execution parameters which SHOULD be used as a base when running a container using the image. This field can be null, in which case any execution parameters should be specified at creation of the container.',
  rootfs: 'The rootfs key references the layer content addresses used by the image. This makes the image config hash depend on the filesystem hash.',
  history: 'Describes the history of each layer. In order from first to last.',
  annotations: 'This property contains arbitrary metadata. Property MUST use the annotation rules.',
  contentDescriptor: 'A descriptor for referencing a blob.',
  size: 'This REQUIRED property specifies the size, in bytes, of the raw content. This property exists so that a client will have an expected size for the content before processing. If the length of the retrieved content does not match the specified length, the content SHOULD NOT be trusted.',
  digest: 'the cryptographic checksum digest of the object, in the pattern `<algorithm>:<encoded>`',
  urls: 'A list of URLs from which the content may be downloaded.',
  schemaVersion: 'This REQUIRED property specifies the schema version. For this version ofthe specification, this MUST be `2` to ensure backward compatibility with older versions of Docker. The value of this field will not change.',
  layers: 'Each item in the array MUST be a descriptor. The array MUST have the base layer at index 0. Subsequent layers MUST then follow in stack order (i.e. from layers[0] to layers[len(layers)-1]). The final filesystem layout MUST match the result of applying the layers to an empty directory. The ownership, mode, and other attributes of the initial empty directory are unspecified.',
  osVersion: 'This OPTIONAL property specifies the version of the operating system targeted by th referenced blob. Implementations MAY refuse to use manifests where os.version is not known to work with the host OS version. Valid values are implementation-defined. eg. `10.0.14393.1066` on windows.',
  osFeatures: 'This OPTIONAL property specifies an array of strings, each specifying a mandatory OS feature.',
  variant: 'This OPTIONAL property specifies the variant of the CPU.',
  features: 'This property is RESERVED for future versions of the specification.',
  manifests: 'This REQUIRED property contains a list of manifests for specific platforms. While this property MUST be present, the size of the array MAY be zero',
  errors: 'This property contains a list of errors messages returned from the server due to an invalid request.',
  catalog: 'This property contains a list of repositories hosted on the registry.',
  tagsList: 'This property contains a list of all tags in a repository.',
  mediaType: '',

};

// Catalog Schema
local catalog(output=JSV7) = {
  local repositories = {
    type: 'array',
    items: { type: 'string' },
  },

  [if output == JSV7 then '$id']: 'http://opencontainers.org/schema/catalog',
  [if output == JSV7 then '$schema']: JSV7Schema,
  title: 'OCI Catalog List',
  type: 'object',
  description: d.catalog,
  properties: {
    repositories: repositories,
  },
};

// Tags List Schema
local tagsList(output='jsonschemaV7') = {
  local name = { type: 'string' },
  local tags = {
    type: 'array',
    items: { type: 'string' },
  },

  [if output == JSV7 then '$id']: 'http://opencontainers.org/schema/tags/list',
  [if output == JSV7 then '$schema']: JSV7Schema,
  title: 'OCI Tags List',
  type: 'object',
  description: d.tagsList,
  properties: {
    name: name,
    tags: tags,
  },
};

// Error Schema
local errors(output='jsonschemaV7') = {

  local code = { type: 'string' },
  local message = { type: 'string' },
  local detail = { type: 'string' },

  local errObj = {
    type: 'object',
    properties: {
      code: code,
      message: message,
      detail: detail,
    },

  },

  local errArray = {
    type: 'array',
    items: errObj,
  },

  [if output == JSV7 then '$id']: 'http://opencontainers.org/schema/error',
  [if output == JSV7 then '$schema']: JSV7Schema,
  title: 'OCI Error Response',
  type: 'object',
  description: d.errors,
  properties: {
    errors: errArray,
  },
};

// Media Type Schema
local mediaType = {
  type: 'string',
  description: d.mediaType,
  pattern: '^[A-Za-z0-9][A-Za-z0-9!#$&-^_.+]{0,126}/[A-Za-z0-9][A-Za-z0-9!#$&-^_.+]{0,126}$',
};

// Annotations Schema
local annotations(output='jsonschemaV7') = {
  description: d.annotations,
  type: 'object',
  [if output == JSV7 then 'patternProperties' else 'x-patternProperties']: {
    '.{1,}': {
      type: 'string',
    },
  },
};

// Content Descriptor Schema
local contentDescriptor(output=JSV7) = {

  local size = {
    description: d.size,
    type: 'integer',
    minimum: -9223372036854776000,
    maximum: 9223372036854776000,
  },

  local digestSHA256 = {
    properties: {
      digest: {
        pattern: '^sha256:[A-Fa-f0-9]{64}$',
      },
    },
  },

  local digestSHA512 = {
    properties: {
      digest: {
        pattern: '^sha512:[A-Fa-f0-9]{128}$',
      },
    },
  },

  local digest = {
    type: 'string',
    description: d.digest,
    pattern: '^[a-z0-9]+(?:[+._-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$',
  },

  local urls = {
    type: 'array',
    description: d.urls,
    items: {
      type: 'string',
      format: 'uri',
    },
  },

  [if output == JSV7 then '$id']: 'http://opencontainers.org/image/descriptor',
  [if output == JSV7 then '$schema']: JSV7Schema,
  title: 'OCI Content Descriptor',
  description: d.contentDescriptor,
  type: 'object',
  properties: {
    mediaType: mediaType,
    size: size,
    digest: digest,
    urls: urls,
    annotations: annotations(output),
  },
  required: [
    'size',
    'mediaType',
    'digest',
  ],
  anyOf: [
    digestSHA256,
    digestSHA512,
  ],
};

// Schema Version Schema
local schemaVersion = {
  type: 'integer',
  description: d.schemaVersion,
  minimum: 2,
  maximum: 2,
};


// Architecture Schema
local arch = {
  type: 'string',
  description: d.arch,
};

// Image Index Schema
local imageIndex(output=JSV7) = {

  local arch = {
    description: d.indexArch,
    type: 'string',
    enum: [
      '386',
      'amd64',
      'arm',
      'arm64',
      'ppc64',
      'ppc64le',
      'mips64',
      'mips64le',
      's390x',
    ],
  },
  local os = {
    description: d.os,
    type: 'string',
    enum: [
      'android',
      'darwin',
      'dragonfly',
      'freebsd',
      'linux',
      'netbsd',
      'openbsd',
      'plan9',
      'solaris',
      'windows',
    ],
  },
  local osVersion = {
    type: 'string',
    description: d.osVersion,
  },
  local osFeatures = {
    description: d.osFeatures,
    type: 'array',
    items: { type: 'string' },
  },
  local variant = {
    type: 'string',
    description: d.variant,
  },
  local features = {
    description: d.features,
    type: 'array',
    items: { type: 'string' },
  },

  local androidCombinations = {
    properties: {
      os: { enum: ['android'] },
      architecture: { enum: ['arm'] },
    },
  },

  local darwinCombinations = {
    properties: {
      os: { enum: ['darwin'] },
      architecture: {
        enum: [
          '386',
          'amd64',
          'arm',
          'arm64',
        ],
      },
    },
  },

  local dragonflyCombinations = {
    properties: {
      os: {
        enum: ['dragonfly'],
      },
      architecture: {
        enum: [
          'amd64',
        ],
      },
    },
  },

  local freebsdCombinations = {
    properties: {
      os: {
        enum: ['freebsd'],
      },
      architecture: {
        enum: [
          '386',
          'amd64',
          'arm',
        ],
      },
    },
  },

  local linuxCombinations = {
    properties: {
      os: {
        enum: ['linux'],
      },
      architecture: {
        enum: [
          '386',
          'amd64',
          'arm',
          'arm64',
          'ppc64',
          'ppc64le',
          'mips64',
          'mips64le',
          's390x',
        ],
      },
    },
  },

  local netbsdCombinations = {
    properties: {
      os: {
        enum: ['netbsd'],
      },
      architecture: {
        enum: [
          '386',
          'amd64',
          'arm',
        ],
      },
    },
  },

  local openbsdCombinations = {
    properties: {
      os: {
        enum: ['openbsd'],
      },
      architecture: {
        enum: [
          '386',
          'amd64',
          'arm',
        ],
      },
    },
  },

  local plan9Combinations = {
    properties: {
      os: {
        enum: ['plan9'],
      },
      architecture: {
        enum: [
          '386',
          'amd64',
        ],
      },
    },
  },

  local solarisCombinations = {
    properties: {
      os: {
        enum: ['solaris'],
      },
      architecture: {
        enum: [
          'amd64',
        ],
      },
    },
  },

  local windowsCombinations = {
    properties: {
      os: {
        enum: ['windows'],
      },
      architecture: {
        enum: [
          '386',
          'amd64',
        ],
      },
    },
  },

  local osArchCombinations = [
    androidCombinations,
    darwinCombinations,
    dragonflyCombinations,
    freebsdCombinations,
    linuxCombinations,
    netbsdCombinations,
    openbsdCombinations,
    plan9Combinations,
    solarisCombinations,
    windowsCombinations,
  ],

  local platform = {
    title: 'OCI Image Index',
    type: 'object',
    properties: {
      architecture: arch,
      os: os,
      'os.version': osVersion,
      'os.features': osFeatures,
      variant: variant,
      features: features,
    },
    required: [
      'architecture',
      'os',
    ],
    anyOf: osArchCombinations,
  },

  local manifests = {
    description: d.manifests,
    type: 'array',
    items: contentDescriptor(output) {
      properties+: {
        platform: platform,
      },
    },
  },


  [if output == JSV7 then '$id']: 'http://opencontainers.org/image/index',
  [if output == JSV7 then '$schema']: JSV7Schema,
  type: 'object',
  properties: {
    schemaVersion: schemaVersion,
    mediaType: mediaType,
    manifests: manifests,
    annotations: annotations(output),
  },
  required: [
    'schemaVersion',
    'manifests',
  ],
};

// Image Manifest Schema
local imageManifest(output=JSV7) = {
  local layers = {
    description: d.layers,
    type: 'array',
    minItems: 1,
    items: contentDescriptor(output),
  },

  [if output == JSV7 then '$id']: 'http://opencontainers.org/image/manifest',
  [if output == JSV7 then '$schema']: JSV7Schema,
  title: 'OCI Image Manifest',
  type: 'object',
  properties: {
    schemaVersion: schemaVersion,
    config: contentDescriptor(output),
    layers: layers,
    annotations: annotations(output),
  },
  required: [
    'schemaVersion',
    'config',
    'layers',
  ],
};

// Image Config Schema
local imageConfig(output=JSV7) = {

  local author = {
    type: 'string',
    description: d.imageAuthor,
  },

  local cmd = {
    type: 'array',
    description: d.cmd,
    items: { type: 'string' },
  },

  local entrypoint = {
    type: 'array',
    description: d.entrypoint,
    items: { type: 'string' },
  },

  local env = {
    type: 'array',
    description: d.env,
    items: { type: 'string' },
  },

  local exposedPorts(output=JSV7) = {
    type: 'object',
    description: d.exposedPorts,
    [if output == JSV7 then 'patternProperties' else 'x-patternProperties']: {
      '.{1,}': {
        type: 'object',
      },
    },
  },

  local labels = {
    type: 'object',
    description: d.labels,
    [if output == JSV7 then 'patternProperties' else 'x-patternProperties']: {
      '.{1,}': {
        type: 'string',
      },
    },
  },

  local stopSignal = {
    type: 'string',
    description: d.stopSignal,
  },

  local user = {
    type: 'string',
    description: d.user,
  },

  local volumes = {
    type: 'object',
    description: d.volumes,
    [if output == JSV7 then 'patternProperties' else 'x-patternProperties']: {
      '.{1,}': {
        type: 'object',
      },
    },
  },

  local workingDir = {
    type: 'string',
    description: d.workingDir,
  },

  local config = {
    type: 'object',
    description: d.configConfig,
    properties: {
      Cmd: cmd,
      Entrypoint: entrypoint,
      Env: env,
      ExposedPorts: exposedPorts(),
      Labels: labels,
      StopSignal: stopSignal,
      User: user,
      Volumes: volumes,
      WorkingDir: workingDir,
    },
  },

  local created = {
    description: d.created,
    type: 'string',
    format: 'date-time',
  },

  local history = {
    local author = {
      description: d.author,
      type: 'string',
    },

    local comment = {
      description: d.comment,
      type: 'string',
    },

    local createdBy = {
      description: d.createdBy,
      type: 'string',
    },

    local emptyLayer = {
      description: d.emptyLayer,
      type: 'boolean',
    },

    description: d.history,
    type: 'array',
    items: {
      type: 'object',
      properties: {
        author: author,
        comment: comment,
        created: created,
        created_by: createdBy,
        empty_layer: emptyLayer,
      },
    },
  },

  local os = {
    description: d.os,
    type: 'string',
  },

  local rootfs = {

    local diffIDs = {
      description: d.diffIDs,
      type: 'array',
      items: {
        type: 'string',
      },
    },

    local type = {
      description: d.rootfsType,
      type: 'string',
      enum: ['layers'],
    },

    description: d.rootfs,
    type: 'object',
    properties: {
      diff_ids: diffIDs,
      type: type,
    },
    required: ['diff_ids', 'type'],
  },

  [if output == JSV7 then '$id']: 'http://opencontainers.org/image/manifest',
  [if output == JSV7 then '$schema']: JSV7Schema,
  title: 'OCI Image Config',
  type: 'object',
  properties: {
    architecture: arch,
    author: author,
    config: config,
    created: created,
    history: history,
    os: os,
    rootfs: rootfs,
  },
  required: [
    'architecture',
    'os',
    'rootfs',
  ],
};


{
  imageConfig:: imageConfig,
  imageManifest:: imageManifest,
  imageIndex:: imageIndex,
  contentDescriptor:: contentDescriptor,
  catalog:: catalog,
  errors:: errors,
  tagsList:: tagsList,
}
