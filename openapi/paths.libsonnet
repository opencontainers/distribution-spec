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
local ops = import 'operations.libsonnet';
local params = import 'parameters.libsonnet';
local pathItem = oapi.pathItem;

local endpoints = {
  oci:: {
    base:: '/v2',
    catalog:: '/v2/_catalog',
    tags:: '/v2/{namespace}/{repository}/tags/list',
    manifest:: '/v2/{namespace}/{repository}/manifests/{reference}',
    blob:: '/v2/{namespace}/{repository}/blobs/{digest}',
    upload:: '/v2/{namespace}/{repository}/blobs/uploads',
    uploadId:: '/v2/{namespace}/{repository}/blobs/uploads/{uuid}',
  },
};

local items = {
  oci:: {
    local base =
      pathItem.new()
      .addOperation('GET', ops.oci.verify),

    local catalog =
      pathItem.new()
      .addParameter(params.n)
      .addParameter(params.last)
      .addOperation('GET', ops.oci.catalogList),

    local tags =
      pathItem.new()
      .addParameter(params.namespace)
      .addParameter(params.repository)
      .addParameter(params.n)
      .addParameter(params.last)
      .addOperation('GET', ops.oci.tagsList),

    local manifest =
      pathItem.new()
      .addParameter(params.namespace)
      .addParameter(params.repository)
      .addParameter(params.reference)
      .addOperation('GET', ops.oci.manifestGet)
      .addOperation('HEAD', ops.oci.manifestExists)
      .addOperation('PUT', ops.oci.manifestCreate)
      .addOperation('DELETE', ops.oci.manifestDelete),

    local blob =
      pathItem.new()
      .addParameter(params.namespace)
      .addParameter(params.repository)
      .addParameter(params.digest)
      .addOperation('GET', ops.oci.blobDownload)
      .addOperation('HEAD', ops.oci.blobExists)
      .addOperation('DELETE', ops.oci.blobDelete),

    local upload =
      pathItem.new()
      .addParameter(params.namespace)
      .addParameter(params.repository)
      .addParameter(params.mount)
      .addParameter(params.mountFrom)
      .addOperation('POST', ops.oci.blobUploadInit),

    local uploadId =
      pathItem.new()
      .addParameter(params.namespace)
      .addParameter(params.repository)
      .addParameter(params.uuid)
      .addOperation('GET', ops.oci.blobUploadStatus)
      .addOperation('PATCH', ops.oci.blobUploadChunk)
      .addOperation('PUT', ops.oci.blobUploadComplete)
      .addOperation('DELETE', ops.oci.blobUploadCancel),

    base:: base,
    catalog:: catalog,
    tags:: tags,
    manifest:: manifest,
    blob:: blob,
    upload:: upload,
    uploadId:: uploadId,
  },
};

local ociPaths =
  oapi.paths.new()
  .addPath(endpoints.oci.base, items.oci.base)
  .addPath(endpoints.oci.catalog, items.oci.catalog)
  .addPath(endpoints.oci.tags, items.oci.tags)
  .addPath(endpoints.oci.manifest, items.oci.manifest)
  .addPath(endpoints.oci.blob, items.oci.blob)
  .addPath(endpoints.oci.upload, items.oci.upload)
  .addPath(endpoints.oci.uploadId, items.oci.uploadId)
;

ociPaths
