// PASS: All fields defined.
local golden = {
  digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
  mediaType: 'application/vnd.oci.image.manifest.v1+json',
  size: 642,
};

// PASS: sha256 digest
local digestSHA256 = {
  digest: 'sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
  mediaType: 'application/vnd.oci.image.manifest.v1+json',
  size: 642,
};

// FAIL: invalid sha256 digest
local digestSHA256Invalid = {
  digest: 'sha256:e3sjfafgjalj161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
  mediaType: 'application/vnd.oci.image.manifest.v1+json',
  size: 642,
};

// PASS: sha512 digest
local digestSHA512 = {
  digest: 'sha512:9e8cdc1ea0d5e4b31626d28d1753954ad13915677ef6aebb6b0cd360162e996636949c1bc542d5c89b1efa7dd2594ef90188a06b62e22448a3ff443e3865e812',
  mediaType: 'application/vnd.oci.image.manifest.v1+json',
  size: 642,
};

// FAIL: invalid sha512 digest
local digestSHA512Invalid = {
  digest: 'sha512:e3sjfafgjalj161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
  mediaType: 'application/vnd.oci.image.manifest.v1+json',
  size: 642,
};

{
  'golden.json': golden,
  'digest-sha256.json': digestSHA256,
  'digest-sha256-invalid.json': digestSHA256Invalid,
  'digest-sha512.json': digestSHA512,
  'digest-sha512-invalid.json': digestSHA512Invalid,
}
