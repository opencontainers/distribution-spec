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
  catalog: 'This property contains a list of repositories hosted on the registry.',
};

// Catalog Schema
local catalog(output=JSV7) = {
  local repositories = {
    type: 'array',
    items: { type: 'string' },
  },

  [if output == JSV7 then '$id']: 'http://opencontainers.org/schema/catalog',
  [if output == JSV7 then '$schema']: JSV7Schema,
  title: 'Repository Catalog',
  type: 'object',
  description: d.catalog,
  properties: {
    repositories: repositories,
  },
};

{
  catalog: catalog,
}
