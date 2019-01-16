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
local resp = import 'responses.libsonnet';
local op = oapi.operation;
local params = import 'parameters.libsonnet';
local mt = import 'media-types.libsonnet';
local req = import 'request-bodies.libsonnet';

// OCI Operations.
local oci = {

  local tags = ['OCI'],

  local verify =
    op.new(
      operationId='oci-verify',
      tags=tags,
      summary='Check that the endpoint implements distribution API.',
      description='This minimal endpoint is used to verify that the registry implements the OCI Distribution Specification.',
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addResponse(resp.oci.verify)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.notFound)
    .addResponse(resp.err.tooManyRequests)
  ,

  local catalogList =
    op.new(
      operationId='oci-catalog-list',
      tags=tags,
      summary='List a set of available repositories in the local registry cluster.',
      description='List a set of available repositories in the local registry cluster. Does not provide any indication of what may be available upstream. Applications can only determine if a repository is available but not if it is not available.',
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addResponse(resp.oci.catalogList)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.tooManyRequests)
  ,

  local tagsList =
    op.new(
      operationId='oci-tags-list',
      tags=tags,
      summary='Get all tags in a repository.',
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addResponse(resp.oci.tagsList)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.tooManyRequests)
  ,

  local manifestGet =
    op.new(
      operationId='oci-manifest-get',
      summary='Get a manifest by name and reference.',
      tags=tags,
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addResponse(resp.oci.manifestGet)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.notFound)
    .addResponse(resp.err.tooManyRequests)
  ,

  local manifestExists =
    op.new(
      operationId='oci-manifest-exists',
      summary='Check for the existence of a manifest by name and reference.',
      tags=tags,
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addResponse(resp.oci.manifestExists)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.notFound)
    .addResponse(resp.err.tooManyRequests)
  ,

  local manifestCreate =
    op.new(
      operationId='oci-manifest-create',
      summary='Add a manifest to a repository.',
      tags=tags,
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)

    .addRequestBody(req.oci.v1.manifestCreate)

    .addResponse(resp.oci.manifestCreate)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.tooManyRequests)
  ,

  local manifestDelete =
    op.new(
      operationId='oci-manifest-delete',
      summary='Delete a manifest from the repository.',
      tags=tags,
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addResponse(resp.oci.manifestDelete)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.notFound)
    .addResponse(resp.err.tooManyRequests)
  ,

  local blobDownload =
    op.new(
      operationId='oci-blob-download',
      summary='Download a blob by digest.',
      tags=tags,
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addParameter(params.range)
    .addResponse(resp.oci.blobDownload)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.notFound)
    .addResponse(resp.err.tooManyRequests)
  ,

  local blobExists =
    op.new(
      operationId='oci-blob-exists',
      tags=tags,
      summary='Check for the existence of a manifest by name and reference.',
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addResponse(resp.oci.blobExists)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.notFound)
    .addResponse(resp.err.tooManyRequests)
  ,

  local blobDelete =
    op.new(
      operationId='oci-blob-delete',
      tags=tags,
      summary='Delete a blob by digest.',
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addResponse(resp.oci.blobDelete)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.notFound)
    .addResponse(resp.err.tooManyRequests)
  ,

  local blobUploadInit =
    op.new(
      operationId='oci-blob-upload-init',
      tags=tags,
      summary='Initiate a blob upload or mount a blob from another respository.',
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addParameter(params.digestQuery)
    .addResponse(resp.oci.blobUploadInit)
    .addResponse(resp.oci.blobMount)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.tooManyRequests)
  ,

  local blobUploadStatus =
    op.new(
      operationId='oci-blob-upload-status',
      tags=tags,
      summary="Check a blob's upload status.",
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addResponse(resp.oci.blobUploadStatus)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.tooManyRequests)
  ,

  local blobUploadChunk =
    op.new(
      operationId='oci-blob-upload-chunk',
      tags=tags,
      summary='Upload a blob chunk to the repository.',
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addResponse(resp.oci.blobUploadChunk)
    .addResponse(resp.err.badRequest)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.notFound)
    .addResponse(resp.err.tooManyRequests)
    .addResponse(resp.err.rangeNotSatisfiable)
  ,

  local blobUploadComplete =
    op.new(
      operationId='oci-blob-upload-complete',
      tags=tags,
      summary='Notify registry that the chunked blob upload is complete.',
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addParameter(params.digestQuery)
    .addResponse(resp.oci.blobUploadComplete)
    .addResponse(resp.err.badRequest)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.notFound)
    .addResponse(resp.err.tooManyRequests)
  ,

  local blobUploadCancel =
    op.new(
      operationId='oci-blob-upload-cancel',
      tags=tags,
      summary='Cancel the chunked blob upload.',
    )
    .addParameter(params.contentLength)
    .addParameter(params.host)
    .addResponse(resp.oci.blobUploadCancel)
    .addResponse(resp.err.badRequest)
    .addResponse(resp.err.unauthorized)
    .addResponse(resp.err.forbidden)
    .addResponse(resp.err.notFound)
    .addResponse(resp.err.tooManyRequests)
  ,

  verify:: verify,
  catalogList:: catalogList,
  tagsList:: tagsList,
  manifestGet:: manifestGet,
  manifestExists:: manifestExists,
  manifestCreate:: manifestCreate,
  manifestDelete:: manifestDelete,
  blobDownload:: blobDownload,
  blobExists:: blobExists,
  blobDelete:: blobDelete,
  blobUploadInit:: blobUploadInit,
  blobUploadStatus:: blobUploadStatus,
  blobUploadChunk:: blobUploadChunk,
  blobUploadComplete:: blobUploadComplete,
  blobUploadCancel:: blobUploadCancel,
};


{
  oci:: oci,
}
