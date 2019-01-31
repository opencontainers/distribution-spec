local JSV7 = 'jsonschemaV7';
local JSV7Schema = 'http://json-schema.org/draft-07/schema#';

// Descriptions for schema objects. Separated out to make code more readable.
local d = {
  tagsList: 'This property contains a list of all tags in a repository.',
};

// Tags List Schema
local tagsList(output=JSV7) = {
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

{
  tagsList: tagsList,
}
