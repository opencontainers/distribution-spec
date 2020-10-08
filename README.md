## Open Container Initiative Distribution Specification

The [Open Container Initiative](https://www.opencontainers.org/) develops specifications for standards on Operating System process and application containers.

The specification can be found [here](https://github.com/opencontainers/distribution-spec/blob/master/spec.md).

### Table of Contents

- [Frequently Asked Questions](#frequently-asked-questions)
- [Code of Conduct][code-of-conduct]
- [Releases](RELEASES.md)
- [charter][charter]


### Frequently Asked Questions

Before diving into the [specification](spec.md), you should consider the following frequently asked questions
and ideas about the overall design of your registry.

**Q: How do I design my registry models for Blobs, Image Manifests, and Repositories?**

By way of specifying the name of the repository (`<name>`) for upload endpoints,
this gives you freedom to design your registry in several ways:

 - Having blobs linked 1:1 to a repository, and then referenced in Image Manifests only belonging to that repository (you wouldn't need to worry about blob sharing across repositories, and cleanup would be limited to ensuring the blob only isn't needed by the repository in question)
 - Having blobs linked to Image Manifests, which are linked to Repositories.
 - Having blobs not linked to any Image Manifest or Repository and shared amongst all blobs (and for this implementation idea you would need to ensure that blobs are not deleted that are still being used, and that there aren't any security loopholes with someone uploading a blob that would be used by another repository.

If there is another design idea not listed, please contribute it to the specification. For all of the above designs, you should consider storage (organization), along with permissions and cleanup. Different registries with varying user groups will likely have different use cases that warrant different behavior.

**Q: How do I handle garbage collection?**

Based on the design you choose above, you might want to implement integrity checking. For example, when you upload a manifest all blobs that it references should already exist in your registry.

**Q: How do I store the manifest?**

It's typically easiest to store the manifest as the raw byte stream that was originally provided. If you want to do additional parsing of content (e.g. to derive layers or annotations) that can be done as well.

**Q: What about annotations in manifests?**

While the specification does not say anything about annotations, you might choose
to parse a manifest and represent them in a model, for easy query. It's up to you
how you want to design an Annotation table. For example, since annotations could contain
secrets you might want to keep them within the scope of an image or repository.

**Q: What kind of digest are we talking about?**

For these push endpoints and generally all distribution specification mentions of blobs, we are generally refering to a sha256 digest, however [other types](https://github.com/opencontainers/image-spec/blob/master/descriptor.md#digests) could be supported.

**Q: Can I change the upload location?**

For a chunked upload, you could implement it so that each PATCH request uses the
same `<location>`, but you could also generate new session ids for scoped uploads.
This decision is up to you. Expiring and then generating a new `<location>` could better match specific chunks to upload sessions, but a single `<location>` shared across chunks could better support more parallel operations (if your registry can support this).

**Q: Should I validate the content type of the patch request body?**

The content type for blob uploads isn't meaningful since it's consistently the same (application/octet-stream).
However, you may so choose to check that the content type is consistent for each chunk in the upload.
It would not be logical for it to change part of the way through.

**Q: Is there a suggested chunk size, or number of chunks?**

There is currently no best practice for an upload size or number of chunks.

**Q: What is the order of a push for some client?**

Generally, the data dependency between components drives the order of operations.
For example, you can't upload a mainfest without knowing the config blob and layer digests. You can't know the config blob digest without knowing the layer diffids. You can't know the layer digests until you've gzipped and uploaded them (unless you know them ahead of time). This generally means that we do an upload like:

 - blobs are uploaded first (config and layers)
 - blobs (optionally) are then linked to an image manifest (and the manifest uploaded)
 - the layers and manifest are linked to a repository

Keep in mind that along with image layers, the config is a [separate blob](https://github.com/opencontainers/image-spec/blob/master/config.md#example) that describes the container runtime specification.

**Q: Can the location session for a push expire?**

While there is no detail in the specification about the session expiring, you could
choose to expire it after, for example, a minute or an hour, in the case that you don't want an upload session to be re-used. If you choose to expire sessions, you should test this expiration time to ensure that it works for different kinds of network connectivity for your user base.

**Q: What happens if the `<tagname>` (last) parameter does not exist?**

There is no suggested behavior in the specification for what to do if the tag does not exist. Registries might consider ignoring te parameter, or assuming a non-existing tag is at the start or the end of the sorted list. In the first case, at the start of the list would imply returning the entire set of tags. In the second cast, at the end of the list would imply returningan empty list, as it references the last tag onward (an empty set).

### Use Cases

Following sections give context for aspects of this specification:

### Registry Developers

### Registry Administrators

### Client Side Image Tools

### Container Image Pipeline

## Contributing

See [our contribution documentation](CONTRIBUTING.md).

Development happens on GitHub.
Responsible disclosure for security issues is discussed [here](CONTRIBUTING.md#security-issues).
[Issues][issues] are used for non-security bugs and actionable items; longer discussions can happen on the [mailing list](#mailing-list).

### Mailing list

You can subscribe and browse the mailing list on [Google Groups][mailing-list].

[charter]: https://www.opencontainers.org/about/governance
[code-of-conduct]: https://github.com/opencontainers/org/blob/master/CODE_OF_CONDUCT.md
[issues]: https://github.com/opencontainers/distribution-spec/issues
[mailing-list]: https://groups.google.com/a/opencontainers.org/forum/#!forum/dev
