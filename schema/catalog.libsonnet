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
