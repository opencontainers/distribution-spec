{
  new(manifests=[]):: {
    schemaVersion: 2,
    mediaType: 'application/vnd.oci.image.index.v1+json',
    manifests: manifests,

    addManifest(
      descriptor=error 'descriptor not defined for Manifest Index object.',
      architecture='amd64',
      os='linux',
      osVersion=null,
      osFeatures=null,
      variant=null,
      features=null,
    ):: self {
      manifests+: [
        descriptor {
          platform: {
            architecture: architecture,
            os: os,
            [if osVersion != null then 'os.version']: osVersion,
            [if osFeatures != null then 'os.features']: osFeatures,
            [if variant != null then 'variant']: variant,
            [if features != null then 'features']: features,
          },
        },
      ],
    },

    addAnnotation(key, value):: self {
      annotations+: { [key]: value },
    },

  },
}
