// PASS: All fields defined.
local golden = {
  schemaVersion: 2,
  mediaType: 'application/vnd.oci.image.index.v1+json',
  manifests: [
    {
      digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
      mediaType: 'application/vnd.oci.image.manifest.v1+json',
      platform: {
        architecture: 'amd64',
        os: 'linux',
      },
      size: 642,
      urls: ['http://example.com'],
    },

  ],
  annotations: {
    'com.example.key1': 'value1',
    'com.example.key2': 'value2',
  },
};

// PASS: Required fields only.
local goldenMinimal = {
  schemaVersion: 2,
  manifests: [
    {
      digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
      mediaType: 'application/vnd.oci.image.manifest.v1+json',
      size: 642,
    },
  ],
};

// PASS: Custom referenced manifest media type.
local customManifestMediaType = {
  schemaVersion: 2,
  manifests: [
    {
      digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
      mediaType: 'application/customized.image.manifest+json',
      size: 642,
    },
  ],
};

// FAIL: schema version too low
local schemaVersionTooLow = {
  schemaVersion: 1,
  manifests: [
    {
      digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
      mediaType: 'application/vnd.oci.image.manifest.v1+json',
      size: 642,
    },
  ],
};

// FAIL: schema version too high
local schemaVersionTooHigh = {
  schemaVersion: 3,
  manifests: [
    {
      digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
      mediaType: 'application/vnd.oci.image.manifest.v1+json',
      size: 642,
    },
  ],
};

// FAIL: Schema version missing.
local schemaVersionMissing = {
  manifests: [
    {
      digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
      mediaType: 'application/vnd.oci.image.manifest.v1+json',
      size: 642,
    },
  ],
};

// FAIL: Manifest is not an array.
local manifestNotArray = {
  schemaVersion: 2,
  manifests: {
    digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
    mediaType: 'application/vnd.oci.image.manifest.v1+json',
    size: 642,
  },
};

// FAIL: Platform OS and Architecture combination invalid.
local platformCombinationInvalid = {
  schemaVersion: 2,
  manifests: [
    {
      digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
      mediaType: 'application/vnd.oci.image.manifest.v1+json',
      size: 642,
      platform: {
        architecture: 'amd64',
        os: 'android',
      },
    },
  ],
};

// FAIL: Platform OS invalid.
local platformOSInvalid = {
  schemaVersion: 2,
  manifests: [
    {
      digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
      mediaType: 'application/vnd.oci.image.manifest.v1+json',
      size: 642,
      platform: {
        os: 'invalid',
        architecture: 'amd64',
      },
    },
  ],
};

// FAIL: Platform Architecture invalid.
local platformArchInvalid = {
  schemaVersion: 2,
  manifests: [
    {
      digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
      mediaType: 'application/vnd.oci.image.manifest.v1+json',
      size: 642,
      platform: {
        os: 'linux',
        architecture: 'invalid',
      },
    },
  ],
};

// FAIL: Digest invalid.
local digestInvalid = {
  schemaVersion: 2,
  manifests: [
    {
      digest: 'sha256:easlkagalskjfalsjfaljasjagijdsogijgakjj3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
      mediaType: 'application/vnd.oci.image.manifest.v1+json',
      size: 642,
    },
  ],
};


// Generate all possible architecture combinations for a given OS.
local genOSCombos(os) = {
  ['platform-' + os + '-' + arch + '.json']: {
    schemaVersion: 2,
    manifests: [
      {
        digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
        mediaType: 'application/vnd.oci.image.manifest.v1+json',
        size: 642,
        platform: {
          os: os,
          architecture: arch,
        },
      },
    ],
  }
  for arch in [
    '386',
    'amd64',
    'arm',
    'arm64',
    'ppc64',
    'ppc64le',
    'mips64',
    'mips64le',
    's390x',
  ]
};

local platformCombos =
  genOSCombos('android')
  + genOSCombos('darwin')
  + genOSCombos('dragonfly')
  + genOSCombos('freebsd')
  + genOSCombos('linux')
  + genOSCombos('netbsd')
  + genOSCombos('openbsd')
  + genOSCombos('plan9')
  + genOSCombos('solaris')
  + genOSCombos('windows')
;


platformCombos {
  'golden.json': golden,
  'golden-minimal.json': goldenMinimal,
  'custom-manifest-media-type.json': customManifestMediaType,
  'schema-version-too-low.json': schemaVersionTooLow,
  'schema-version-too-high.json': schemaVersionTooHigh,
  'schema-version-missing.json': schemaVersionMissing,
  'manifest-not-array.json': manifestNotArray,
  'platform-combination-invalid.json': platformCombinationInvalid,
  'platform-os-invalid.json': platformOSInvalid,
  'platform-arch-invalid.json': platformArchInvalid,
}
