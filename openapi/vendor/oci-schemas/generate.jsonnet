local oci = import 'oci.libsonnet';

{
  'image-config.schema.json': oci.imageConfig(),
  'image-manifest.schema.json': oci.imageManifest(),
  'image-index.schema.json': oci.imageIndex(),
  'content-descriptor.schema.json': oci.contentDescriptor(),
  'errors.schema.json': oci.errors(),
  'catalog.schema.json': oci.catalog(),
  'tag-list.schema.json': oci.tagsList(),
}
