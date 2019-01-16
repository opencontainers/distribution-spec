local errDescriptor(param) =
  error std.format('%s not defined for Descriptor object.', param);

{
  new(
    mediaType=errDescriptor('mediaType'),
    digest=errDescriptor('digest'),
    size=errDescriptor('size'),
    urls=null,
    annotations=null,
    data=null,
  ):: {
    mediaType: mediaType,
    digest: digest,
    size: size,

    [if urls != null then 'urls']: urls,
    [if annotations != null then 'annotations']: annotations,
    [if data != null then 'data']: data,
  },
}
