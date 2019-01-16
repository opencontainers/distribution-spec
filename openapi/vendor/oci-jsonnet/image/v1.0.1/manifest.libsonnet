local err(param) =
  error std.format('%s not defined for Image Manifest object.', param);

{
  new(
    mediaType=null,
    config=err('config'),
    layers=[],
  ):: {
    schemaVersion: 2,
    [if mediaType != null then 'mediaType']: mediaType,
    config: config,
    layers: layers,

    addLayer(descriptor):: self {
      layers+: [descriptor],
    },

    addAnnotation(key, value):: self {
      annotations+: { [key]: value },
    },

  },
}
