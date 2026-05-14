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

This endpoint returns the history of a tag, listing each manifest the tag has pointed to over time and each time the tag was deleted, in descending order by history entry timestamp (newest first).

Tag history MAY be retrieved with a standard `GET` as follows.

```HTTP
GET /v2/<name>/_oci/tag-history/<tag>
```

`<name>` is the namespace of the repository, and `<tag>` is the name of the tag whose history is being queried.

A successful request MUST return a `200 OK` response code.
If the repository does not exist, the registry MUST return a `404 Not Found` response code.
If the tag has no history, the registry MUST return a `404 Not Found` response code.
Tag history SHOULD remain queryable after the tag is deleted.
Deleted manifests MUST NOT be removed from tag history.
History entries for deleted manifests MUST continue to include the descriptor fields recorded when the tag was assigned to the manifest.
Tag history MAY be deleted after the repository is deleted.

Upon success, the response body MUST be a JSON array of descriptor objects in the following format:

```HTTP
200 OK
Content-Length: <length>
Content-Type: application/json

[
    {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "size": 1234,
        "digest": "sha256:a1a1a1...",
        "annotations": {
            "org.opencontainers.tag.created": "2026-05-04T03:02:01Z"
        }
    },
    {
        "mediaType": "application/vnd.oci.empty.v1+json",
        "digest": "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
        "size": 2,
        "data": "e30=",
        "annotations": {
            "org.opencontainers.tag.deleted": "2026-05-03T03:02:01Z"
        }
    },
    {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "size": 1234,
        "digest": "sha256:b2b2b2...",
        "annotations": {
            "org.opencontainers.tag.created": "2026-04-03T02:01:00Z"
        }
    }
]
```

Results MUST be sorted in descending order by the history entry timestamp, which is either the `org.opencontainers.tag.created` annotation value or the `org.opencontainers.tag.deleted` annotation value (i.e. the most recent entry appears first).

Each descriptor object in the response MUST be either a manifest descriptor or a delete marker.

Manifest descriptors MUST include the following properties:

- **`mediaType`** *string*, REQUIRED

  The media type of the manifest this tag pointed to at that point in time.

- **`size`** *integer*, REQUIRED

  The size in bytes of the manifest.

- **`digest`** *string*, REQUIRED

  The digest of the manifest, in the form `<algorithm>:<encoded>`.

- **`annotations`** *map of strings*, REQUIRED

  Annotations associated with the historical tag entry. MUST include:

  - **`org.opencontainers.tag.created`** *string*, REQUIRED

    The RFC 3339 timestamp at which the tag was assigned to this manifest. Used as the sort key and as the cursor for time-based pagination.

Delete markers MUST be represented by the OCI empty descriptor and MUST include the following properties:

- **`mediaType`** *string*, REQUIRED

  MUST be `application/vnd.oci.empty.v1+json`.

- **`digest`** *string*, REQUIRED

  MUST be `sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a`.

- **`size`** *integer*, REQUIRED

  MUST be `2`.

- **`data`** *string*, REQUIRED

  MUST be `e30=`.

- **`annotations`** *map of strings*, REQUIRED

  Annotations associated with the historical tag entry. MUST include:

  - **`org.opencontainers.tag.deleted`** *string*, REQUIRED

    The RFC 3339 timestamp at which the tag was deleted. Used as the sort key and as the cursor for time-based pagination.

##### Query Parameters

The following query parameters MAY be provided:

- **`n`** *integer*, OPTIONAL

  Specifies the maximum number of results to return.
  The registry MAY return fewer results than requested if fewer historical entries exist.
  When `n` is zero, this endpoint MUST return an empty array.
  When `n` is not specified, the registry MAY apply a default limit.

- **`before`** *string (RFC 3339 timestamp)*, OPTIONAL

  When provided, only entries whose history entry timestamp is strictly less than (i.e. older than) `before` will be returned.
  This is used for time-based pagination: pass the `org.opencontainers.tag.created` or `org.opencontainers.tag.deleted` value of the last entry returned in the previous response to retrieve the next page.

##### Pagination Example

To fetch the first page of up to 10 results:

```HTTP
GET /v2/<name>/_oci/tag-history/<tag>?n=10
```

To fetch the next page, pass the history entry timestamp of the last result from the prior response as the `before` parameter:

```HTTP
GET /v2/<name>/_oci/tag-history/<tag>?n=10&before=2026-04-03T02%3A01%3A00Z
```

When the returned array contains fewer entries than `n` (or is empty), the client has reached the end of the history.

## Code representations

Golang structures for these JSON structures is available at [`github.com/opencontainers/distribution-spec/specs-go/v1/extensions`](https://github.com/opencontainers/distribution-spec/tree/main/specs-go/v1/extensions/)

## Error Codes

Registry implementations MAY chose to not support any extension and the base extension MAY return the following error message.

| Code          | Message                       | Description                                                                                 |
|---------------|-------------------------------|---------------------------------------------------------------------------------------------|
| `UNSUPPORTED` | The operation is unsupported. | The operation was unsupported due to a missing implementation or invalid set of parameters. |
