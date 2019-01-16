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
local param = oapi.parameter;
local schemas = import '../schema/schema.libsonnet';
local headers = import 'headers.libsonnet';

local string = { type: 'string' };


local namespace = param.new(
  name='namespace',
  paramIn='path',
  description='A logical grouping of repositories.',
  schema=string,
);

local repository = param.new(
  name='repository',
  paramIn='path',
  description='A place to store an image.',
  schema=string,
);

local reference = param.new(
  name='reference',
  paramIn='path',
  description='Can either be a tag name or digest',
  schema=string,
);

local digest = param.new(
  name='digest',
  paramIn='path',
  description='A content addressable identifier.',
  schema=string,
);

local digestQuery = param.new(
  name='digest',
  paramIn='query',
  style='form',
  description='A content addressable identifier.',
  schema=string,
);

local mount = param.new(
  name='mount',
  paramIn='query',
  style='form',
  description='The digest of a blob in the source repository.',
  schema=string,
);

local mountFrom = param.new(
  name='from',
  paramIn='query',
  style='form',
  description='The source repository the referenced blob.',
  schema=string,
);

local uuid = param.new(
  name='uuid',
  paramIn='path',
  description='A universally unique identifier.',
  schema=string,
);

local n = param.new(
  name='n',
  paramIn='query',
  style='form',
  description='',
  schema=string,
);

local last = param.new(
  name='last',
  paramIn='query',
  style='form',
  description='',
  schema=string,
);

local host = param.new(
  name='Host',
  paramIn='header',
  style='simple',
  description='Standard HTTP Host header.',
  schema=string,
);

local range = param.new(
  name='Range',
  paramIn='header',
  style='simple',
  description='HTTP Range header specifying a blob chunk.',
  schema=string,
);

local contentLength = param.new(
  name='Content-Length',
  paramIn='header',
  style='simple',
  description='The byte length of the data returned.',
  schema=string,
);

{
  namespace: namespace,
  repository: repository,
  reference: reference,
  digest: digest,
  digestQuery: digestQuery,
  mount: mount,
  mountFrom: mountFrom,
  uuid: uuid,
  n: n,
  last: last,
  host: host,
  range: range,
  contentLength: contentLength,
}
