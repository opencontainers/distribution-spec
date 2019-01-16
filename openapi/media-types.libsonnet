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

local oapi = import 'openapi-jsonnet/v3.0.0/openapi.libsonnet';
local mt = oapi.mediaType;
local ex = oapi.example;
local liboci = import 'oci-jsonnet/image/v1.0.1/spec.libsonnet';
local ociExamples = liboci.examples;
local schema = import 'oci-schemas/oci.libsonnet';

local json = 'application/json';

local err = {
  [json]: mt.new(schema.errors('openapi')),
};


local binary = { type: 'string', format: 'binary' };
local common = {

  octetStream: {
    'application/octet-stream':
      mt.new(binary),
  },

};

local oci = {
  v1:: {

    imageManifest:: {
      'application/vnd.oci.image.manifest.v1+json':
        mt.new(schema.imageManifest('openapi'))
        .addExample(ociExamples.manifest),
    },

    index:: {
      'application/vnd.oci.index.v1+json':
        mt.new(schema.imageIndex('openapi')),
    },

    config:: {
      'application/vnd.oci.config.v1+json':
        mt.new(schema.config('openapi')),
    },

    descriptor:: 'application/vnd.oci.descriptor.v1+json',
    layout:: 'application/vnd.oci.layout.header.v1+json',

    layerTar:: {
      'application/vnd.oci.layer.v1.tar':
        mt.new(binary),
    },

    layerTarGz:: {
      'application/vnd.oci.layer.v1.tar+gzip':
        mt.new(binary),
    },

    layerNoDistTar:: {
      'application/vnd.oci.layer.nondistributable.v1.tar':
        mt.new(binary),
    },

    layerNoDistTarGz:: {
      'application/vnd.oci.layer.nondistributable.v1.tar+gzip':
        mt.new(binary),
    },

    catalog:: {
      [json]: mt.new(schema.catalog('openapi')),
    },

    tagsList:: {
      [json]: mt.new(schema.tagsList('openapi')),
    },
  },


};


{
  err: err,
  common: common,
  oci: oci,
}
