// PASS: All fields defined.
local golden = {
  schemaVersion: 2,
  config: {
    mediaType: 'application/vnd.oci.image.config.v1+json',
    digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
    size: 452,
  },
  layers: [
    {
      mediaType: 'application/vnd.oci.image.layer.v1.tar+gzip',
      digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
      size: 78343,
    },
  ],
  annotations: {
    'com.example.key1': 'value1',
    'com.example.key2': 'value2',
  },
};

// PASS: Requried fields only.
local goldenMinimal = {
  schemaVersion: 2,
  config: {
    mediaType: 'application/vnd.oci.image.config.v1+json',
    digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
    size: 452,
  },
  layers: [
    {
      mediaType: 'application/vnd.oci.image.layer.v1.tar+gzip',
      digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
      size: 78343,
    },
  ],
};

// FAIL: Schema version too low.
local schemaVersionTooLow = {
  schemaVersion: 1,
  config: {
    mediaType: 'application/vnd.oci.image.config.v1+json',
    digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
    size: 452,
  },
  layers: [
    {
      mediaType: 'application/vnd.oci.image.layer.v1.tar+gzip',
      digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
      size: 78343,
    },
  ],
};

// FAIL: Schema version too high.
local schemaVersionTooHigh = {
  schemaVersion: 3,
  config: {
    mediaType: 'application/vnd.oci.image.config.v1+json',
    digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
    size: 452,
  },
  layers: [
    {
      mediaType: 'application/vnd.oci.image.layer.v1.tar+gzip',
      digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
      size: 78343,
    },
  ],
};

// FAIL: Schema version missing.
local schemaVersionMissing = {
  config: {
    mediaType: 'application/vnd.oci.image.config.v1+json',
    digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
    size: 452,
  },
  layers: [
    {
      mediaType: 'application/vnd.oci.image.layer.v1.tar+gzip',
      digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
      size: 78343,
    },
  ],
};

// FAIL: Config mediatype invalid.
local configMediaTypeInvalid = {
  schemaVersion: 2,
  config: {
    mediaType: 'invalid',
    digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
    size: 452,
  },
  layers: [
    {
      mediaType: 'application/vnd.oci.image.layer.v1.tar+gzip',
      digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
      size: 78343,
    },
  ],
};

// FAIL: Layers mediatype invalid.
local layerMediaTypeInvalid = {
  schemaVersion: 2,
  config: {
    mediaType: 'application/vnd.oci.image.config.v1+json',
    digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
    size: 452,
  },
  layers: [
    {
      mediaType: 'invalid',
      digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
      size: 78343,
    },
  ],
};

// FAIL: config size not number.
local configSizeNotNumber = {
  schemaVersion: 2,
  config: {
    mediaType: 'application/vnd.oci.image.config.v1+json',
    digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
    size: '452',
  },
  layers: [
    {
      mediaType: 'application/vnd.oci.image.layer.v1.tar+gzip',
      digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
      size: 78343,
    },
  ],
};

// FAIL: Layer size not number.
local layerSizeNotNumber = {
  schemaVersion: 2,
  config: {
    mediaType: 'application/vnd.oci.image.config.v1+json',
    digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
    size: 452,
  },
  layers: [
    {
      mediaType: 'application/vnd.oci.image.layer.v1.tar+gzip',
      digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
      size: '78343',
    },
  ],
};

// FAIL: At least one layer descriptor should be defined.
local layersEmptyArray = {
  schemaVersion: 2,
  config: {
    mediaType: 'application/vnd.oci.image.config.v1+json',
    digest: 'sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401',
    size: 452,
  },
  layers: [],
};


{
  'golden.json': golden,
  'golden-minimal.json': goldenMinimal,
  'schema-version-too-low.json': schemaVersionTooLow,
  'schema-version-too-high.json': schemaVersionTooHigh,
  'schema-version-missing.json': schemaVersionMissing,
  'config-media-type-invalid.json': configMediaTypeInvalid,
  'layer-media-type-invalid.json': layerMediaTypeInvalid,
  'config-size-not-number.json': configSizeNotNumber,
  'layer-size-not-number.json': layerSizeNotNumber,
  'layers-empty-array.json': layersEmptyArray,
}
