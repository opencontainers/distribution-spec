local catalog = import 'catalog.libsonnet';
local errors = import 'errors.libsonnet';
local tags = import 'tags.libsonnet';

{
  'catalog.schema.json': catalog.catalog(),
  'errors.schema.json': errors.errors(),
  'tags-list.schema.json': tags.tagsList(),
}
