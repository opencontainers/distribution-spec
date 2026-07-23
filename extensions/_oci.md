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
    GET /v2/<name>/_oci/ext/discover
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

### Component: `tag-history`

This component is for endpoints relating to tag history operations on a repository.

This endpoint returns the history of a tag, listing each manifest the tag has pointed to over time and each time the tag was deleted, in descending order (newest first).

Tag history MAY be retrieved with a standard `GET` as follows.

```HTTP
GET /v2/<name>/_oci/tag-history/<tag>
```

`<name>` is the namespace of the repository, and `<tag>` is the name of the tag whose history is being queried.

A successful request MUST return a `200 OK` response code.
If the registry does not implement this extension, the registry MUST return a `404 Not Found` response code.
The registry MUST NOT return a `404 Not Found` response code for any other condition.
If the repository does not exist or the tag has no history, the registry MUST return a `200 OK` response code with an empty manifest list.
Tag history SHOULD remain queryable after the tag is deleted.
Deleted manifests SHOULD NOT be removed from tag history.
Tag history MAY be deleted after the repository is deleted.

Upon success, the response body MUST be an OCI Index containing a list of descriptor objects in the following format:

```HTTP
200 OK
Content-Length: <length>
Content-Type: application/vnd.oci.image.index.v1+json

{
    "schemaVersion": 2,
    "mediaType": "application/vnd.oci.image.index.v1+json",
    "manifests": [
        {
            "mediaType": "application/vnd.oci.image.manifest.v1+json",
            "size": 1234,
            "digest": "sha256:a1a1a1...",
            "annotations": {
                "org.opencontainers.distribution.tag.timestamp": "2026-05-04T03:02:01Z",
                "org.opencontainers.distribution.tag.event": "created"
            }
        },
        {
            "mediaType": "application/vnd.oci.image.manifest.v1+json",
            "size": 1234,
            "digest": "sha256:a1a1a1...",
            "annotations": {
                "org.opencontainers.distribution.tag.timestamp": "2026-02-03T03:02:01Z",
                "org.opencontainers.distribution.tag.event": "deleted"
            }
        },
        {
            "mediaType": "application/vnd.oci.image.manifest.v1+json",
            "size": 1234,
            "digest": "sha256:b2b2b2...",
            "annotations": {
                "org.opencontainers.distribution.tag.timestamp": "2025-04-03T02:01:00Z",
                "org.opencontainers.distribution.tag.event": "created"
            }
        }
    ]
}
```

Results MUST be sorted in descending order by the history entries (i.e. the most recent entry appears first).

Each descriptor object in the response MUST be a create or delete entry.
A descriptor for a tag creation event MUST describe the manifest the tag was assigned to.
A descriptor for a tag deletion event MUST describe the manifest the tag pointed to immediately before deletion.

Manifest descriptors MUST include the following properties:

- **`mediaType`** *string*, REQUIRED

  The media type of the manifest this tag pointed to at that point in time.

- **`size`** *integer*, REQUIRED

  The size in bytes of the manifest.

- **`digest`** *string*, REQUIRED

  The digest of the manifest, in the form `<algorithm>:<encoded>`.

- **`annotations`** *map of strings*, REQUIRED

  Annotations associated with the historical tag entry. Registries MAY include annotations from the underlying manifest or index descriptor.
- When doing so, any existing annotations with keys beginning with `org.opencontainers.distribution` MUST be ignored because this prefix is reserved for annotations generated by the registry.

  The annotations MUST include both of the following:

  - **`org.opencontainers.distribution.tag.timestamp`** *string*

    The RFC 3339 timestamp at which the tag history event occurred. Used as the sort key and as the cursor for time-based pagination.

  - **`org.opencontainers.distribution.tag.event`** *string*

    The tag history event type. MUST be either `created` when the tag was assigned to the manifest or `deleted` when the tag was deleted from the manifest.

##### Query Parameters

The following query parameters MAY be provided:

- **`n`** *integer*, OPTIONAL

  Specifies the maximum number of results to return.
  The registry MAY return fewer results than requested if fewer historical entries exist.
  When `n` is zero, this endpoint MUST return an OCI Index with an empty manifest list and MUST NOT include a `Link` header, useful for determining whether the extension is supported (`200 OK`) or not (`404 Not Found`).
  When `n` is not specified, the registry MAY apply a default limit.

- **`before`** *string (RFC 3339 timestamp)*, OPTIONAL

  When provided, only entries whose history entry timestamp is strictly less than (i.e. older than) `before` will be returned.
  This is used for time-based pagination: pass the `org.opencontainers.distribution.tag.timestamp` value of the last entry returned in the previous response to retrieve the next page.

- **`since`** *string (RFC 3339 timestamp)*, OPTIONAL

  When provided, only entries whose history entry timestamp is strictly greater than (i.e. newer than) `since` will be returned.
  The `since` and `before` parameters MAY be combined to return entries within an open time interval.

- **`digest`** *string*, OPTIONAL

  When provided, only history entries for the given digest will be returned.
  If the digest does not exist or was not part of the given tag's history, the registry MUST return a `200 OK` response code with an empty manifest list.

##### Pagination Example

A `Link` header MUST be included in the response when additional history entries are available.
The `Link` header MUST be set according to [RFC 5988](https://www.rfc-editor.org/rfc/rfc5988) with the Relation Type `rel="next"`.
The target URI MAY contain registry-specific query parameters.
Clients MUST use the target URI as provided and MUST NOT modify or interpret registry-specific query parameters.
When a response does not include a `Link` header, the client has reached the end of the history.

To fetch the first page of up to 10 results:

```HTTP
GET /v2/<name>/_oci/tag-history/<tag>?n=10
```

When constructing the next request without using a `Link` header, pass the history entry timestamp of the last result from the prior response as the `before` parameter:

```HTTP
GET /v2/<name>/_oci/tag-history/<tag>?n=10&before=2026-04-03T02%3A01%3A00Z
```

To fetch all history entries after a given timestamp:

```HTTP
GET /v2/<name>/_oci/tag-history/<tag>?since=2026-04-01T00%3A00%3A00Z
```

## Code representations

Golang structures for these JSON structures is available at [`github.com/opencontainers/distribution-spec/specs-go/v1/extensions`](https://github.com/opencontainers/distribution-spec/tree/main/specs-go/v1/extensions/)

## Error Codes

Registry implementations MAY chose to not support any extension and the base extension MAY return the following error message.

| Code          | Message                       | Description                                                                                 |
|---------------|-------------------------------|---------------------------------------------------------------------------------------------|
| `UNSUPPORTED` | The operation is unsupported. | The operation was unsupported due to a missing implementation or invalid set of parameters. |
