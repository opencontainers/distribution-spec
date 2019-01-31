//   Copyright 2018 The Linux Foundation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

local JSV7 = 'jsonschemaV7';
local JSV7Schema = 'http://json-schema.org/draft-07/schema#';

// Descriptions for schema objects. Separated out to make code more readable.
local d = {
  errors: 'This property contains a list of errors messages returned from the server due to an invalid request.',
};

// Error Schema
local errors(output=JSV7) = {

  local code = {
    type: 'string',
    enum: [
      'BLOB_UNKNOWN',
      'BLOB_UPLOAD_INVALID',
      'BLOB_UPLOAD_UNKNOWN',
      'DIGEST_INVALID',
      'MANIFEST_BLOB_UNKNOWN',
      'MANIFEST_INVALID',
      'MANIFEST_UNKNOWN',
      'MANIFEST_UNVERIFED',
      'NAME_INVALID',
      'NAME_UNKNOWN',
      'SIZE_INVALID',
      'TAG_INVALID',
      'UNAUTHORIZED',
      'DENIED',
      'UNSUPPORTED',
    ],
  },
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

  [if output == JSV7 then '$id']: 'http://opencontainers.org/schema/errors',
  [if output == JSV7 then '$schema']: JSV7Schema,
  title: 'OCI Error Response',
  type: 'object',
  description: d.errors,
  properties: {
    errors: errArray,
  },
};

{
  errors: errors,
}
