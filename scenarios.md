## Use Cases

### Image Verification

A container engine would like to run verified image named "library/ubuntu", with the tag "latest".
The engine contacts the registry, requesting the manifest for "library/ubuntu:latest".
An untrusted registry returns a manifest.
Before proceeding to download the individual layers, the engine verifies the manifest's signature, ensuring that the content was produced from a trusted source and no tampering has occurred.
After each layer is downloaded, the engine verifies the digest of the layer, ensuring that the content matches that specified by the manifest.

### Resumable Push

Company X's build servers lose connectivity to a distribution endpoint before completing an image layer transfer.
After connectivity returns, the build server attempts to re-upload the image.
The registry notifies the build server that the upload has already been partially attempted.
The build server responds by only sending the remaining data to complete the image file.

### Resumable Pull

Company X is having more connectivity problems but this time in their deployment datacenter.
When downloading an image, the connection is interrupted before completion.
The client keeps the partial data and uses http `Range` requests to avoid downloading repeated data.

### Layer Upload De-duplication

Company Y's build system creates two identical layers from build processes A and B.
Build process A completes uploading the layer before B.
When process B attempts to upload the layer, the registry indicates that its not necessary because the layer is already known.

If process A and B upload the same layer at the same time, both operations will proceed and the first to complete will be stored in the registry (Note: we MAY modify this to prevent dogpile with some locking mechanism).
