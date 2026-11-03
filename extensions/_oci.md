# _oci Extension Endpoints

## Summary

This base extension namespace for OCI namespaced extension endpoints.

## Reference Explanation

### Component: `ext`

This component is for endpoints relating to extensions on the registry API being queried.

#### Module: `discover`

This endpoint may be queried to discover extensions available on this registry API.

Registry-level extensions may be discovered with a standard `GET` as follows.

```HTTP
GET /v2/_oci/ext/discover
```

Repository-level extensions may be discovered with a standard GET as follows.

```HTTP
    GET /v2/{name}/_oci/ext/discover
```

The base extension returns an array of supported extensions with details of the endpoints as shown below.

```HTTP
200 OK
Content-Length: <length>
Content-Type: application/json

{
    "extensions": [
        {
            "name": "_<extension>",
            "description": "",
            "url": "..."
        }
    ]    
}
```

### *Extensions* Property Descriptions

- **`extensions`** *array of extension objects*, REQUIRED

  This property contains a list of supported extension endpoints.

  - **`name`** *string*, REQUIRED

    The name of the extension (i.e. '`_oci`') as it will appear in the URL path.

  - **`url`** *string*, REQUIRED

    URL link to the documentation defining this extension and its endpoints.

  - **`description`** *string*, OPTIONAL

    Human readable description of this endpoint.

  - **`endpoints`** *array of strings*, REQUIRED

    Enumeration of the endpoints provided on this registry (as not all "OPTIONAL" endpoints may be present in all registries)

### Component: `copy`

This component is for server-side copying of content between repositories within the same registry.
It allows a client to request a copy without fetching manifests, traversing indices, mounting individual blobs, or pushing manifests and tags separately.
The registry MAY share, link, mount, or physically copy stored content; the extension does not prescribe a storage implementation or guarantee a particular latency.

#### Module: `manifest`

This OPTIONAL endpoint copies a root manifest and its required dependencies into a destination repository, then applies destination tags.
The root may be an image manifest, an image index, or an artifact packaged as an OCI image manifest.

```HTTP
POST /v2/<name>/_oci/copy/manifest?from=<source>&reference=<tag-or-digest>&tag=<tag1>&tag=<tag2>
Accept: application/vnd.oci.descriptor.v1+json
Content-Length: 0
```

`<name>` is the destination repository.
The source and destination MUST be on the same registry and MAY be the same repository.
Repository names, tags, and digests MUST follow the syntax defined by [Pulling manifests](../spec.md#pulling-manifests).
The request MUST NOT contain a body.

##### Query Parameters

- **`from`** *string*, REQUIRED

  The source repository name.
  This parameter MUST occur exactly once and MUST NOT contain a URL or registry-qualified image reference.

- **`reference`** *string*, REQUIRED

  The source manifest tag or digest.
  This parameter MUST occur exactly once.

- **`tag`** *string*, OPTIONAL, repeatable

  A tag to create or update in the destination repository, pointing to the copied root manifest.
  If one or more `tag` parameters are supplied, they replace the default destination tag described below; they do not add to it.
  Duplicate tag values MUST be treated as a single tag.
  An empty tag value is invalid and MUST NOT be treated as an omitted parameter.

Registries SHOULD support at least 10 destination tags per request, consistent with [Pushing Manifests](../spec.md#pushing-manifests).
Registries MAY reject an excessively long request URI with `414 Request-URI Too Long` before performing the copy.

##### Destination Tags

The registry MUST select the root manifest's destination tags as follows:

| Source `reference` | Explicit `tag` parameters | Destination tags |
|--------------------|---------------------------|------------------|
| Tag `release` | None | `release` |
| Tag `release` | `stable`, `1.4` | `stable`, `1.4` only |
| Digest | None | None; the manifest is available by digest |
| Digest | `stable`, `1.4` | `stable`, `1.4` |

If `reference` is a tag and no `tag` parameters are supplied, the registry MUST create or update that same tag in the destination repository.
If explicit tags are supplied, the source tag MUST NOT be created or updated unless it is explicitly included.
If `reference` is a digest and no tags are supplied, the registry MUST NOT assign a tag to the copied root manifest.
All selected tags MUST point to the single digest resolved for this operation.
Other source tags, including tags on child manifests, MUST NOT be implicitly copied.
Destination tags not selected by these rules MUST remain unchanged, except for referrers tag schema maintenance described below.

##### Copy Semantics

The registry MUST resolve `reference` to a root manifest digest once and use that digest throughout the operation.
If `reference` is a tag, a concurrent update to the source tag MUST NOT change the graph being copied.

Registries implementing this endpoint MUST support OCI image manifests and image indices, including OCI image manifests used for artifacts.
Registries MAY support additional manifest formats, such as Docker schema 2 manifests and manifest lists.
A registry MUST reject an unsupported manifest format rather than copy it without its required dependencies.

The registry MUST make the root manifest and its required dependencies available through the destination repository's ordinary manifest and blob endpoints:

1. For an image index, recursively copy every entry in `manifests`, including nested indices and all platforms.
1. For an image manifest, copy its `config` and every entry in `layers`.
1. Treat configuration and layer content as opaque bytes. An unfamiliar configuration or layer media type, or an unfamiliar `artifactType`, MUST NOT by itself cause the copy to fail.

The registry MUST preserve the exact bytes and media types of copied manifests and blobs, preserving their digests, annotations, and descriptor fields.
It MUST NOT convert manifest formats, rewrite descriptors, or filter platforms.
The source content MUST remain unchanged.

Required content MUST be available in the source repository or supplied by valid embedded descriptor `data`.
The registry MAY reuse content already available in the destination, subject to the same source authorization and validation requirements.
It MUST NOT fetch content from descriptor `urls` or a remote registry as part of this operation.
Missing required dependencies MUST cause the operation to fail rather than produce an incomplete copy.

The registry MUST make dependencies available before publishing a parent manifest and MUST NOT apply the selected destination tags until the required graph is available.
It MUST protect the copy from concurrent source deletion or garbage collection sufficiently to complete the operation, or fail without publishing an incomplete graph.
After success, deleting the source content MUST NOT invalidate the destination copy; the destination's own deletion, retention, and garbage collection rules still apply.

##### Subject Relationships and Referrers

The registry MUST preserve `subject` descriptors unchanged.
Registries SHOULD NOT traverse subject relationships when determining the content to copy.
Copying the referenced subject or its dependencies is not required, and clients MUST NOT assume that a successful copy makes that content available in the destination repository.
A subject that is independently included through an index's `manifests` entries is still a required dependency.

Discovering and copying referrers, such as separately attached signatures or SBOMs, is outside the scope of this operation.
An artifact included through an index's `manifests` entries MUST still be copied normally.
For every copied manifest with a `subject`, the registry MUST update the destination's referrers API or, if that API is unavailable, maintain the [referrers tag schema](../spec.md#referrers-tag-schema) as described in [Pushing Manifests with Subject](../spec.md#pushing-manifests-with-subject).
Existing unrelated entries in the destination's referrers lists MUST be preserved.

##### Authorization and Validation

Registries MUST enforce the same applicable authorization, validation, and policy checks that would apply if the client performed the equivalent sequence of [Pulling manifests](../spec.md#pulling-manifests), [Mounting a blob from another repository](../spec.md#mounting-a-blob-from-another-repository), and [Pushing Manifests](../spec.md#pushing-manifests) operations, including destination tag updates.
This requirement applies to all content copied by the operation, including any subjects and their dependencies the registry elects to copy.
Server-side copy MUST NOT bypass checks merely because the content already exists in registry storage.
The registry is not required to issue internal HTTP requests to perform these checks.

The caller MUST have permission to read the required source content and to write the destination content and selected tags.
For registries using repository-scoped bearer tokens, this corresponds to `pull` access on the source and `push` access on the destination, with any additional checks required by the equivalent operations.
Knowing a digest or finding content in shared storage MUST NOT substitute for authorization to access it through the source repository.
Existing destination policies, such as tag immutability, quotas, and manifest validation, MUST remain in force.

##### Successful Response

A successful request MUST return `201 Created` only after the graph has been copied and all selected destination tags have been applied.
This also applies when the requested content and tags were already present.
This endpoint is synchronous; it MUST NOT return `202 Accepted` to indicate an unfinished copy or a blob upload session.

The response MUST include:

- `Location`: a pullable URL for the root manifest by digest in the destination repository.
- `Docker-Content-Digest`: the resolved root manifest digest, not the digest of the response body.
- `Content-Type: application/vnd.oci.descriptor.v1+json`.
- `OCI-Tag`: every selected destination tag successfully applied, including an implicitly selected source tag, using the header semantics defined by [Pushing Manifests](../spec.md#pushing-manifests). This header MUST be omitted when no destination tags were selected.

The response body MUST be an [OCI descriptor](https://github.com/opencontainers/image-spec/blob/v1.1.1/descriptor.md) describing the copied root manifest.
Its `digest` MUST match `Docker-Content-Digest`, its `mediaType` MUST identify the root manifest's media type, and its `size` MUST be the size of the root manifest in bytes, not the size of the response body.

For example, this request copies `staging/app:release` into `production/app`, assigning only `1.4` and `stable`:

```HTTP
POST /v2/production/app/_oci/copy/manifest?from=staging%2Fapp&reference=release&tag=1.4&tag=stable
Accept: application/vnd.oci.descriptor.v1+json
Content-Length: 0
```

The response is shown with a placeholder for the resolved digest:

```HTTP
201 Created
Location: /v2/production/app/manifests/<resolved-digest>
Docker-Content-Digest: <resolved-digest>
OCI-Tag: 1.4, stable
Content-Type: application/vnd.oci.descriptor.v1+json

{
    "mediaType": "application/vnd.oci.image.index.v1+json",
    "digest": "<resolved-digest>",
    "size": 1234
}
```

##### Failures and Retries

The operation does not require an atomic transaction across all copied content and tag updates.
Registries SHOULD validate the request and known policy constraints before modifying the destination.
If a copy fails, content and referrers metadata already copied MAY remain in the destination.
Registries MUST NOT modify any selected destination tags if copying the required graph fails.
If a later tag update fails, earlier successful tag updates MAY remain; rollback is not required.
The registry MUST return an error rather than `201 Created` for partial completion.

If any selected destination tags were successfully applied before an error, the error response MUST include `Docker-Content-Digest` with the resolved root digest and `OCI-Tag` listing those tags.
If none were applied, the registry MUST omit `OCI-Tag`.
Tag updates MUST NOT continue after the registry returns an error response.

Errors returned by an implementation of this endpoint MUST use the [Distribution error response format and codes](../spec.md#error-codes).
The following conditions use these status codes:

| Status | Error code | Condition |
|--------|------------|-----------|
| `400 Bad Request` | `NAME_INVALID`, `DIGEST_INVALID`, or `UNSUPPORTED` | Invalid repository name, digest, or request parameters, including missing required parameters or invalid tags. |
| `400 Bad Request` | `MANIFEST_INVALID` | A manifest fails validation. |
| `400 Bad Request` | `MANIFEST_BLOB_UNKNOWN` | A required manifest or blob dependency is missing. |
| `401 Unauthorized` | `UNAUTHORIZED` | Authentication is required. |
| `403 Forbidden` | `DENIED` | Authorization or a destination policy denies the operation. |
| `404 Not Found` | `NAME_UNKNOWN` or `MANIFEST_UNKNOWN` | The source repository or root manifest is not found. |
| `415 Unsupported Media Type` | `UNSUPPORTED` | A manifest format needed for the copy is unsupported. |
| `429 Too Many Requests` | `TOOMANYREQUESTS` | A rate or resource limit prevents the operation. |

Registries MAY enforce resource limits on the operation, but MUST fail rather than report success for an incomplete copy.
A `429 Too Many Requests` response SHOULD include `Retry-After` as described in the Distribution Specification.

A connection failure or lost response does not establish whether the operation completed or which tags were applied.
Repeating a digest-based request selects the same content, but can overwrite intervening destination tag updates.
Repeating a tag-based request may resolve a different source digest.
Clients that require reproducible copies SHOULD resolve the source to a digest before submitting the copy request.
This endpoint does not provide conditional tag updates or compare-and-swap semantics.

##### Discovery and Fallback

Registries implementing this endpoint MUST advertise it in the `endpoints` property of the `_oci` extension returned by repository-level [extension discovery](#module-discover).
Advertising the `_oci` namespace alone does not indicate support for this endpoint.

Clients MAY fall back to ordinary client-side copying when discovery is unavailable or the copy endpoint is unsupported.
Clients MUST NOT treat every failed copy as evidence that the extension is unsupported: a `404 Not Found` may identify missing source content, and authorization or validation errors do not indicate missing support.
Clients MUST NOT assume that a timeout or other ambiguous failure left the destination unchanged.

## Code representations

Golang structures for these JSON structures is available at [`github.com/opencontainers/distribution-spec/specs-go/v1/extensions`](https://github.com/opencontainers/distribution-spec/tree/main/specs-go/v1/extensions/)

## Error Codes

Registry implementations MAY chose to not support any extension and the base extension MAY return the following error message.

| Code          | Message                       | Description                                                                                 |
|---------------|-------------------------------|---------------------------------------------------------------------------------------------|
| `UNSUPPORTED` | The operation is unsupported. | The operation was unsupported due to a missing implementation or invalid set of parameters. |
