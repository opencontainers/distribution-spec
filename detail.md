## Detail

> **Note**: This section is still under construction.
> For the purposes of implementation, if any details below differ from the described request flows above, the section below SHOULD be corrected.
> When they match, this note SHOULD be removed.

The behavior of the endpoints are covered in detail in this section, organized by the API protocol flows (route)
consisting of requests and responses between the entities implementing the API. All aspects of the request and responses
are covered, including headers, parameters and body formats. Examples of requests and their corresponding responses, with
success and failure, are enumerated.

> **Note**: The sections on endpoint detail are arranged with an example request, a description of the request, followed by information about that request.

A list of methods and URIs are covered in the table below:

| Method | Path                               | Entity               | Description                                                                                                                                                                                                                              |
|--------|------------------------------------|----------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| GET    | `/v2/`                             | Base                 | Check that the endpoint implements distribution API.                                                                                                                                                                               |
| GET    | `/v2/<name>/tags/list`             | Tags                 | Fetch the tags under the repository identified by `name`.                                                                                                                                                                                |
| GET    | `/v2/<name>/manifests/<reference>` | Manifest             | Fetch the manifest identified by `name` and `reference` where `reference` can be a tag or digest. A `HEAD` request can also be issued to this endpoint to obtain resource information without receiving all data.                        |
| PUT    | `/v2/<name>/manifests/<reference>` | Manifest             | Put the manifest identified by `name` and `reference` where `reference` can be a tag or digest.                                                                                                                                          |
| DELETE | `/v2/<name>/manifests/<reference>` | Manifest             | Delete the manifest identified by `name` and `reference`. Note that a manifest can _only_ be deleted by `digest`.                                                                                                                        |
| GET    | `/v2/<name>/blobs/<digest>`        | Blob                 | Retrieve the blob from the registry identified by `digest`. A `HEAD` request can also be issued to this endpoint to obtain resource information without receiving all data.                                                              |
| DELETE | `/v2/<name>/blobs/<digest>`        | Blob                 | Delete the blob identified by `name` and `digest`                                                                                                                                                                                        |
| POST   | `/v2/<name>/blobs/uploads/`        | Initiate Blob Upload | Initiate a resumable blob upload. If successful, an upload location will be provided to complete the upload. Optionally, if the `digest` parameter is present, the request body will be used to complete the upload in a single request. |
| GET    | `/v2/<name>/blobs/uploads/<session_id>`  | Blob Upload          | Retrieve status of upload identified by `session_id`. The primary purpose of this endpoint is to resolve the current status of a resumable upload.                                                                                             |
| PATCH  | `/v2/<name>/blobs/uploads/<session_id>`  | Blob Upload          | Upload a chunk of data for the specified upload.                                                                                                                                                                                         |
| PUT    | `/v2/<name>/blobs/uploads/<session_id>`  | Blob Upload          | Complete the upload specified by `session_id`, optionally appending the body as the final chunk.                                                                                                                                               |
| DELETE | `/v2/<name>/blobs/uploads/<session_id>`  | Blob Upload          | Cancel outstanding upload processes, releasing associated resources. If this is not called, the unfinished uploads will eventually timeout.                                                                                              |
The detail for each endpoint is covered in the following sections.

### Errors

The error codes encountered via the API are enumerated in the following table:

| Code                    | Message                                        | Description                                                                                                                                                                                                                                                                                         |
|-------------------------|------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `BLOB_UNKNOWN`          | blob unknown to registry                       | This error MAY be returned when a blob is unknown to the registry in a specified repository. This can be returned with a standard get or if a manifest references an unknown layer during upload.                                                                                                   |
| `BLOB_UPLOAD_INVALID`   | blob upload invalid                            | The blob upload encountered an error and can no longer proceed.                                                                                                                                                                                                                                     |
| `BLOB_UPLOAD_UNKNOWN`   | blob upload unknown to registry                | If a blob upload has been cancelled or was never started, this error code MAY be returned.                                                                                                                                                                                                          |
| `DIGEST_INVALID`        | provided digest did not match uploaded content | When a blob is uploaded, the registry will check that the content matches the digest provided by the client. The error MAY include a detail structure with the key "digest", including the invalid digest string. This error MAY also be returned when a manifest includes an invalid layer digest. |
| `MANIFEST_BLOB_UNKNOWN` | blob unknown to registry                       | This error MAY be returned when a manifest blob is  unknown to the registry.                                                                                                                                                                                                                        |
| `MANIFEST_INVALID`      | manifest invalid                               | During upload, manifests undergo several checks ensuring validity. If those checks fail, this error MAY be returned, unless a more specific error is included. The detail will contain information the failed validation.                                                                           |
| `MANIFEST_UNKNOWN`      | manifest unknown                               | This error is returned when the manifest, identified by name and tag is unknown to the repository.                                                                                                                                                                                                  |
| `MANIFEST_UNVERIFIED`   | manifest failed signature verification         | During manifest upload, if the manifest fails signature verification, this error will be returned.                                                                                                                                                                                                  |
| `NAME_INVALID`          | invalid repository name                        | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                                                                                                         |
| `NAME_UNKNOWN`          | repository name not known to registry          | This is returned if the name used during an operation is unknown to the registry.                                                                                                                                                                                                                   |
| `SIZE_INVALID`          | provided length did not match content length   | When a layer is uploaded, the provided size will be checked against the uploaded content. If they do not match, this error will be returned.                                                                                                                                                        |
| `TAG_INVALID`           | manifest tag did not match URI                 | During a manifest upload, if the tag in the manifest does not match the uri tag, this error will be returned.                                                                                                                                                                                       |
| `UNAUTHORIZED`          | authentication required                        | The access controller was unable to authenticate the client. Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate.                                                                                                                              |
| `DENIED`                | requested access to the resource is denied     | The access controller denied access for the operation on a resource.                                                                                                                                                                                                                                |
| `UNSUPPORTED`           | The operation is unsupported.                  | The operation was unsupported due to a missing implementation or invalid set of parameters.                                                                                                                                                                                                         |

### Base

Base V2 API route.
Typically, this can be used for lightweight version checks and to validate registry authentication.

#### GET Base

Check that the endpoint implements the distribution API.

```HTTP
GET /v2/
Host: <registry host>
Authorization: <scheme> <token>
```

The following parameters SHOULD be specified on the request:

| Name            | Kind   | Description                                                    |
|-----------------|--------|----------------------------------------------------------------|
| `Host`          | header | Standard HTTP Host Header. SHOULD be set to the registry host. |
| `Authorization` | header | An RFC7235 compliant authorization header.                     |

###### On Success: OK

```HTTP
200 OK
```

The API implements V2 protocol and is accessible.

###### On Failure: Not Found

```HTTP
404 Not Found
```

The registry does not implement the V2 API.

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                            |
|----------------|-------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client. Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

### Tags

Retrieve information about tags.

#### GET Tags

Fetch the tags under the repository identified by `name`.

##### Tags

```HTTP
GET /v2/<name>/tags/list
Host: <registry host>
Authorization: <scheme> <token>
```

Return all tags for the repository.
The implementation MAY impose a maximum limit and return a partial set with pagination links.

The following parameters SHOULD be specified on the request:

| Name            | Kind   | Description                                                    |
|-----------------|--------|----------------------------------------------------------------|
| `Host`          | header | Standard HTTP Host Header. SHOULD be set to the registry host. |
| `Authorization` | header | An RFC7235 compliant authorization header.                     |
| `name`          | path   | Name of the target repository.                                 |

###### On Success: OK

```HTTP
200 OK
Content-Length: <length>
Content-Type: application/json

{
    "name": "<name>",
    "tags": [
        "<tag>",
        ...
    ]
}
```

A list of tags for the named repository.

The following headers will be returned with the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                            |
|----------------|-------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client. Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

##### Tags Paginated

```HTTP
GET /v2/<name>/tags/list?n=<integer>&last=<integer>
```

Return a portion of the tags for the specified repository.
The implementation MAY impose a maximum limit and return a partial set with pagination links.

The following parameters SHOULD be specified on the request:

| Name   | Kind  | Description                                                                                 |
|--------|-------|---------------------------------------------------------------------------------------------|
| `name` | path  | Name of the target repository.                                                              |
| `n`    | query | Limit the number of entries in each response. If not present, all entries will be returned. |
| `last` | query | Result set will include values lexically after last.                                        |

###### On Success: OK

```HTTP
200 OK
Content-Length: <length>
Link: <<url>?n=<last n value>&last=<last entry from response>>; rel="next"
Content-Type: application/json

{
    "name": "<name>",
    "tags": [
        "<tag>",
        ...
    ],
}
```

A list of tags for the named repository.

The following headers will be returned with the response:

| Name             | Description                                                            |
|------------------|------------------------------------------------------------------------|
| `Content-Length` | Length of the JSON response body.                                      |
| `Link`           | RFC5988 compliant rel='next' with URL to next result set, if available |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                            |
|----------------|-------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client. Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

### Manifest

Create, update, delete and retrieve manifests.

#### GET Manifest

Fetch the manifest identified by `name` and `reference` where `reference` can be a tag or digest.
A `HEAD` request can also be issued to this endpoint to obtain resource information without receiving all data.

```HTTP
GET /v2/<name>/manifests/<reference>
Host: <registry host>
Authorization: <scheme> <token>
```

The following parameters SHOULD be specified on the request:

| Name            | Kind   | Description                                                    |
|-----------------|--------|----------------------------------------------------------------|
| `Host`          | header | Standard HTTP Host Header. SHOULD be set to the registry host. |
| `Authorization` | header | An RFC7235 compliant authorization header.                     |
| `name`          | path   | Name of the target repository.                                 |
| `reference`     | path   | Tag or digest of the target manifest.                          |

###### On Success: OK

```HTTP
200 OK
Docker-Content-Digest: <digest>
Content-Type: <media type of manifest>

{
   "annotations": {
      "com.example.key1": "value1",
      "com.example.key2": "value2"
   },
   "config": {
      "digest": "sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401",
      "mediaType": "application/vnd.oci.image.config.v1+json",
      "size": 452
   },
   "layers": [
      {
         "digest": "sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401",
         "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
         "size": 78343
      }
   ],
   "schemaVersion": 2
}
```

The manifest identified by `name` and `reference`.
The contents can be used to identify and resolve resources required to run the specified image.

The following headers will be returned with the response:

| Name                    | Description                                     |
|-------------------------|-------------------------------------------------|
| `Docker-Content-Digest` | Digest of the targeted content for the request. |

###### On Failure: Bad Request

```HTTP
400 Bad Request
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The name or reference was invalid.

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                        | Description                                                                                                   |
|----------------|--------------------------------|---------------------------------------------------------------------------------------------------------------|
| `NAME_INVALID` | invalid repository name        | Invalid repository name encountered either during manifest validation or any API operation.                   |
| `TAG_INVALID`  | manifest tag did not match URI | During a manifest upload, if the tag in the manifest does not match the uri tag, this error will be returned. |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                            |
|----------------|-------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client. Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

#### PUT Manifest

Put the manifest identified by `name` and `reference` where `reference` can be a tag or digest.

```HTTP
PUT /v2/<name>/manifests/<reference>
Host: <registry host>
Authorization: <scheme> <token>
Content-Type: <media type of manifest>

{
   "annotations": {
      "com.example.key1": "value1",
      "com.example.key2": "value2"
   },
   "config": {
      "digest": "sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401",
      "mediaType": "application/vnd.oci.image.config.v1+json",
      "size": 452
   },
   "layers": [
      {
         "digest": "sha256:6f4e69a5ff18d92e7315e3ee31c62165ebf25bfa05cad05c0d09d8f412dae401",
         "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
         "size": 78343
      }
   ],
   "schemaVersion": 2
}
```

The following parameters SHOULD be specified on the request:

| Name            | Kind   | Description                                                   |
|-----------------|--------|---------------------------------------------------------------|
| `Host`          | header | Standard HTTP Host Header.SHOULD be set to the registry host. |
| `Authorization` | header | An RFC7235 compliant authorization header.                    |
| `name`          | path   | Name of the target repository.                                |
| `reference`     | path   | Tag or digest of the target manifest.                         |

###### On Success: Created

```HTTP
201 Created
Location: <url>
Content-Length: 0
Docker-Content-Digest: <digest>
```

The manifest has been accepted by the registry and is stored under the specified `name` and `tag`.

The following headers will be returned with the response:

| Name                    | Description                                                          |
|-------------------------|----------------------------------------------------------------------|
| `Location`              | The canonical location url of the uploaded manifest.                 |
| `Content-Length`        | The `Content-Length` header MUST be zero and the body MUST be empty. |
| `Docker-Content-Digest` | Digest of the targeted content for the request.                      |

###### On Failure: Invalid Manifest

```HTTP
400 Bad Request
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The received manifest was invalid in some way, as described by the error codes.
The client SHOULD resolve the issue and retry the request.

The error codes that MAY be included in the response body are enumerated below:

| Code                  | Message                                | Description                                                                                                                                                                                                              |
|-----------------------|----------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `NAME_INVALID`        | invalid repository name                | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                              |
| `TAG_INVALID`         | manifest tag did not match URI         | During a manifest upload, if the tag in the manifest does not match the uri tag, this error will be returned.                                                                                                            |
| `MANIFEST_INVALID`    | manifest invalid                       | During upload, manifests undergo several checks ensuring validity. If those checks fail, this error MAY be returned, unless a more specific error is included.The detail will contain information the failed validation. |
| `MANIFEST_UNVERIFIED` | manifest failed signature verification | During manifest upload, if the manifest fails signature verification, this error will be returned.                                                                                                                       |
| `BLOB_UNKNOWN`        | blob unknown to registry               | This error MAY be returned when a blob is unknown to the registry in a specified repository.This can be returned with a standard get or if a manifest references an unknown layer during upload.                         |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

###### On Failure: Missing Layer(s)

```HTTP
400 Bad Request
Content-Type: application/json

{
    "errors": [
        {
            "code": "BLOB_UNKNOWN",
            "message": "blob unknown to registry",
            "detail": {
                "digest": "<digest>"
            }
        },
        ...
    ]
}
```

One or more layers MAY be missing during a manifest upload.
If so, the missing layers will be enumerated in the error response.

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                  | Description                                                                                                                                                                                      |
|----------------|--------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `BLOB_UNKNOWN` | blob unknown to registry | This error MAY be returned when a blob is unknown to the registry in a specified repository.This can be returned with a standard get or if a manifest references an unknown layer during upload. |

###### On Failure: Not allowed

```HTTP
405 Method Not Allowed
```

Manifest put is not allowed because the registry is configured as a pull-through cache or for some other reason

The error codes that MAY be included in the response body are enumerated below:

| Code          | Message                       | Description                                                                                 |
|---------------|-------------------------------|---------------------------------------------------------------------------------------------|
| `UNSUPPORTED` | The operation is unsupported. | The operation was unsupported due to a missing implementation or invalid set of parameters. |

#### DELETE Manifest

Delete the manifest identified by `name` and `reference`.
Note that a manifest can _only_ be deleted by `digest`.

```HTTP
DELETE /v2/<name>/manifests/<reference>
Host: <registry host>
Authorization: <scheme> <token>
```

The following parameters SHOULD be specified on the request:

| Name            | Kind   | Description                                                    |
|-----------------|--------|----------------------------------------------------------------|
| `Host`          | header | Standard HTTP Host Header. SHOULD be set to the registry host. |
| `Authorization` | header | An RFC7235 compliant authorization header.                     |
| `name`          | path   | Name of the target repository.                                 |
| `reference`     | path   | Tag or digest of the target manifest.                          |

###### On Success: Accepted

```HTTP
202 Accepted
```

###### On Failure: Invalid Name or Reference

```HTTP
400 Bad Request
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The specified `name` or `reference` were invalid and the delete was unable to proceed.

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                        | Description                                                                                                   |
|----------------|--------------------------------|---------------------------------------------------------------------------------------------------------------|
| `NAME_INVALID` | invalid repository name        | Invalid repository name encountered either during manifest validation or any API operation.                   |
| `TAG_INVALID`  | manifest tag did not match URI | During a manifest upload, if the tag in the manifest does not match the uri tag, this error will be returned. |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                          |
|-------------------|-------------------|----------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times. |

###### On Failure: Unknown Manifest

```HTTP
404 Not Found
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The specified `name` or `reference` are unknown to the registry and the delete was unable to proceed.
Clients can assume the manifest was already deleted if this response is returned.

The error codes that MAY be included in the response body are enumerated below:

| Code               | Message                               | Description                                                                                        |
|--------------------|---------------------------------------|----------------------------------------------------------------------------------------------------|
| `NAME_UNKNOWN`     | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry.                  |
| `MANIFEST_UNKNOWN` | manifest unknown                      | This error is returned when the manifest, identified by name and tag is unknown to the repository. |

###### On Failure: Not allowed

```HTTP
405 Method Not Allowed
```

Manifest delete is not allowed because the registry is configured as a pull-through cache or `delete` has been disabled.

The error codes that MAY be included in the response body are enumerated below:

| Code          | Message                       | Description                                                                                 |
|---------------|-------------------------------|---------------------------------------------------------------------------------------------|
| `UNSUPPORTED` | The operation is unsupported. | The operation was unsupported due to a missing implementation or invalid set of parameters. |

### Blob

Operations on blobs identified by `name` and `digest`.
Used to fetch or delete layers by digest.

#### GET Blob

Retrieve the blob from the registry identified by `digest`.
A `HEAD` request can also be issued to this endpoint to obtain resource information without receiving all data.

##### Fetch Blob

```HTTP
GET /v2/<name>/blobs/<digest>
Host: <registry host>
Authorization: <scheme> <token>
```

The following parameters SHOULD be specified on the request:

| Name            | Kind   | Description                                                   |
|-----------------|--------|---------------------------------------------------------------|
| `Host`          | header | Standard HTTP Host Header.SHOULD be set to the registry host. |
| `Authorization` | header | An RFC7235 compliant authorization header.                    |
| `name`          | path   | Name of the target repository.                                |
| `digest`        | path   | Digest of desired blob.                                       |

###### On Success: OK

```HTTP
200 OK
Content-Length: <length>
Docker-Content-Digest: <digest>
Content-Type: application/octet-stream

<blob binary data>
```

The blob identified by `digest` is available.
The blob content will be present in the body of the request.

The following headers will be returned with the response:

| Name                    | Description                                     |
|-------------------------|-------------------------------------------------|
| `Content-Length`        | The length of the requested blob content.       |
| `Docker-Content-Digest` | Digest of the targeted content for the request. |

###### On Success: Temporary Redirect

```HTTP
307 Temporary Redirect
Location: <blob location>
Docker-Content-Digest: <digest>
```

The blob identified by `digest` is available at the provided location.

The following headers will be returned with the response:

| Name                    | Description                                        |
|-------------------------|----------------------------------------------------|
| `Location`              | The location where the layer SHOULD be accessible. |
| `Docker-Content-Digest` | Digest of the targeted content for the request.    |

###### On Failure: Bad Request

```HTTP
400 Bad Request
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

There was a problem with the request that needs to be addressed by the client, such as an invalid `name` or `tag`.

The error codes that MAY be included in the response body are enumerated below:

| Code             | Message                                        | Description                                                                                                                                                                                                                                                                                       |
|------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `NAME_INVALID`   | invalid repository name                        | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                                                                                                       |
| `DIGEST_INVALID` | provided digest did not match uploaded content | When a blob is uploaded, the registry will check that the content matches the digest provided by the client.The error MAY include a detail structure with the key "digest", including the invalid digest string.This error MAY also be returned when a manifest includes an invalid layer digest. |

###### On Failure: Not Found

```HTTP
404 Not Found
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The blob, identified by `name` and `digest`, is unknown to the registry.

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                                                                                                                                      |
|----------------|---------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry.                                                                                                                |
| `BLOB_UNKNOWN` | blob unknown to registry              | This error MAY be returned when a blob is unknown to the registry in a specified repository.This can be returned with a standard get or if a manifest references an unknown layer during upload. |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |



The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

##### Fetch Blob Part

```HTTP
GET /v2/<name>/blobs/<digest>
Host: <registry host>
Authorization: <scheme> <token>
Range: bytes=<start>-<end>
```

This endpoint MAY also support RFC7233 compliant range requests.
Support can be detected by issuing a HEAD request.
If the header `Accept-Range: bytes` is returned, range requests can be used to fetch partial content.

The following parameters SHOULD be specified on the request:

| Name            | Kind   | Description                                                   |
|-----------------|--------|---------------------------------------------------------------|
| `Host`          | header | Standard HTTP Host Header.SHOULD be set to the registry host. |
| `Authorization` | header | An RFC7235 compliant authorization header.                    |
| `Range`         | header | HTTP Range header specifying blob chunk.                      |
| `name`          | path   | Name of the target repository.                                |
| `digest`        | path   | Digest of desired blob.                                       |

###### On Success: Partial Content

```HTTP
206 Partial Content
Content-Length: <length>
Content-Range: bytes <start>-<end>/<size>
Content-Type: application/octet-stream

<blob binary data>
```

The blob identified by `digest` is available.
The specified chunk of blob content will be present in the body of the request.

The following headers will be returned with the response:

| Name             | Description                             |
|------------------|-----------------------------------------|
| `Content-Length` | The length of the requested blob chunk. |
| `Content-Range`  | Content range of blob chunk.            |

###### On Failure: Bad Request

```HTTP
400 Bad Request
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

There was a problem with the request that needs to be addressed by the client, such as an invalid `name` or `tag`.

The error codes that MAY be included in the response body are enumerated below:

| Code             | Message                                        | Description                                                                                                                                                                                                                                                                                       |
|------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `NAME_INVALID`   | invalid repository name                        | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                                                                                                       |
| `DIGEST_INVALID` | provided digest did not match uploaded content | When a blob is uploaded, the registry will check that the content matches the digest provided by the client.The error MAY include a detail structure with the key "digest", including the invalid digest string.This error MAY also be returned when a manifest includes an invalid layer digest. |

###### On Failure: Not Found

```HTTP
404 Not Found
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                                                                                                                                      |
|----------------|---------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry.                                                                                                                |
| `BLOB_UNKNOWN` | blob unknown to registry              | This error MAY be returned when a blob is unknown to the registry in a specified repository.This can be returned with a standard get or if a manifest references an unknown layer during upload. |

###### On Failure: Requested Range Not Satisfiable

```HTTP
416 Requested Range Not Satisfiable
```

The range specification cannot be satisfied for the requested content.
This can happen when the range is not formatted correctly or if the range is outside of the valid size of the content.

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

#### DELETE Blob

Delete the blob identified by `name` and `digest`

```HTTP
DELETE /v2/<name>/blobs/<digest>
Host: <registry host>
Authorization: <scheme> <token>
```

The following parameters SHOULD be specified on the request:

| Name            | Kind   | Description                                                   |
|-----------------|--------|---------------------------------------------------------------|
| `Host`          | header | Standard HTTP Host Header.SHOULD be set to the registry host. |
| `Authorization` | header | An RFC7235 compliant authorization header.                    |
| `name`          | path   | Name of the target repository.                                |
| `digest`        | path   | Digest of desired blob.                                       |

###### On Success: Accepted

```HTTP
202 Accepted
Content-Length: 0
Docker-Content-Digest: <digest>
```

The following headers will be returned with the response:

| Name                    | Description                                     |
|-------------------------|-------------------------------------------------|
| `Content-Length`        | 0                                               |
| `Docker-Content-Digest` | Digest of the targeted content for the request. |

###### On Failure: Invalid Name or Digest

```HTTP
400 Bad Request
```

The error codes that MAY be included in the response body are enumerated below:

| Code             | Message                                        | Description                                                                                                                                                                                                                                                                                       |
|------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `DIGEST_INVALID` | provided digest did not match uploaded content | When a blob is uploaded, the registry will check that the content matches the digest provided by the client.The error MAY include a detail structure with the key "digest", including the invalid digest string.This error MAY also be returned when a manifest includes an invalid layer digest. |
| `NAME_INVALID`   | invalid repository name                        | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                                                                                                       |

###### On Failure: Not Found

```HTTP
404 Not Found
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The blob, identified by `name` and `digest`, is unknown to the registry.

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                                                                                                                                      |
|----------------|---------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry.                                                                                                                |
| `BLOB_UNKNOWN` | blob unknown to registry              | This error MAY be returned when a blob is unknown to the registry in a specified repository.This can be returned with a standard get or if a manifest references an unknown layer during upload. |

###### On Failure: Method Not Allowed

```HTTP
405 Method Not Allowed
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

Blob delete is not allowed because the registry is configured as a pull-through cache or `delete` has been disabled

The error codes that MAY be included in the response body are enumerated below:

| Code          | Message                       | Description                                                                                 |
|---------------|-------------------------------|---------------------------------------------------------------------------------------------|
| `UNSUPPORTED` | The operation is unsupported. | The operation was unsupported due to a missing implementation or invalid set of parameters. |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

### Initiate Blob Upload

Initiate a blob upload.
This endpoint can be used to create resumable uploads or monolithic uploads.

#### POST Initiate Blob Upload

Initiate a resumable blob upload.
If successful, an upload location will be provided to complete the upload.
Optionally, if the `digest` parameter is present, the request body will be used to complete the upload in a single request.

##### Initiate Monolithic Blob Upload

```
POST /v2/<name>/blobs/uploads/?digest=<digest>
Host: <registry host>
Authorization: <scheme> <token>
Content-Length: <length of blob>
Content-Type: application/octet-stream

<binary data>
```

Upload a blob identified by the `digest` parameter in single request.
This upload will not be resumable unless a recoverable error is returned.

The following parameters SHOULD be specified on the request:

| Name             | Kind   | Description                                                                                                                                     |
|------------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| `Host`           | header | Standard HTTP Host Header.SHOULD be set to the registry host.                                                                                   |
| `Authorization`  | header | An RFC7235 compliant authorization header.                                                                                                      |
| `Content-Length` | header |                                                                                                                                                 |
| `name`           | path   | Name of the target repository.                                                                                                                  |
| `digest`         | query  | Digest of uploaded blob.If present, the upload will be completed, in a single request, with contents of the request body as the resulting blob. |

###### On Success: Created

```HTTP
201 Created
Location: <blob location>
Content-Length: 0
Blob-Upload-Session-ID: <session_id>
```

The blob has been created in the registry and is available at the provided location.

The following headers will be returned with the response:

| Name                 | Description                                                          |
|----------------------|----------------------------------------------------------------------|
| `Location`           |                                                                      |
| `Content-Length`     | The `Content-Length` header MUST be zero and the body MUST be empty. |
| `Blob-Upload-Session-ID` | Identifies the upload session_id for the current request.           |

###### On Failure: Invalid Name or Digest

```HTTP
400 Bad Request
```

The error codes that MAY be included in the response body are enumerated below:

| Code             | Message                                        | Description                                                                                                                                                                                                                                                                                       |
|------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `DIGEST_INVALID` | provided digest did not match uploaded content | When a blob is uploaded, the registry will check that the content matches the digest provided by the client.The error MAY include a detail structure with the key "digest", including the invalid digest string.This error MAY also be returned when a manifest includes an invalid layer digest. |
| `NAME_INVALID`   | invalid repository name                        | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                                                                                                       |

###### On Failure: Not allowed

```HTTP
405 Method Not Allowed
```

Blob upload is not allowed because the registry is configured as a pull-through cache or for some other reason

The error codes that MAY be included in the response body are enumerated below:

| Code          | Message                       | Description                                                                                 |
|---------------|-------------------------------|---------------------------------------------------------------------------------------------|
| `UNSUPPORTED` | The operation is unsupported. | The operation was unsupported due to a missing implementation or invalid set of parameters. |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
             "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
             "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

##### Initiate Resumable Blob Upload

```HTTP
POST /v2/<name>/blobs/uploads/
Host: <registry host>
Authorization: <scheme> <token>
Content-Length: 0
```

Initiate a resumable blob upload with an empty request body.

The following parameters SHOULD be specified on the request:

| Name             | Kind   | Description                                                          |
|------------------|--------|----------------------------------------------------------------------|
| `Host`           | header | Standard HTTP Host Header.SHOULD be set to the registry host.        |
| `Authorization`  | header | An RFC7235 compliant authorization header.                           |
| `Content-Length` | header | The `Content-Length` header MUST be zero and the body MUST be empty. |
| `name`           | path   | Name of the target repository.                                       |

###### On Success: Accepted

```HTTP
202 Accepted
Content-Length: 0
Location: /v2/<name>/blobs/uploads/<session_id>
Range: 0-0
Blob-Upload-Session-ID: <session_id>
```

The upload has been created.
The `Location` header MUST be used to complete the upload.
The response SHOULD be identical to a `GET` request on the contents of the returned `Location` header.

The following headers will be returned with the response:

| Name                 | Description                                                                                                                                    |
|----------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| `Content-Length`     | The `Content-Length` header MUST be zero and the body MUST be empty.                                                                           |
| `Location`           | The location of the created upload.Clients SHOULD use the contents verbatim to complete the upload, adding parameters where required.          |
| `Range`              | Range header indicating the progress of the upload.When starting an upload, it will return an empty range, since no content has been received. |
| `Blob-Upload-Session-ID` | Identifies the upload session_id for the current request.                                                                                     |

###### On Failure: Invalid Name or Digest

```HTTP
400 Bad Request
```

The error codes that MAY be included in the response body are enumerated below:

| Code             | Message                                        | Description                                                                                                                                                                                                                                                                                       |
|------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `DIGEST_INVALID` | provided digest did not match uploaded content | When a blob is uploaded, the registry will check that the content matches the digest provided by the client.The error MAY include a detail structure with the key "digest", including the invalid digest string.This error MAY also be returned when a manifest includes an invalid layer digest. |
| `NAME_INVALID`   | invalid repository name                        | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                                                                                                       |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
             "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
             "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

##### Mount Blob

```HTTP
POST /v2/<name>/blobs/uploads/?mount=<digest>&from=<repository name>
Host: <registry host>
Authorization: <scheme> <token>
Content-Length: 0
```

Mount a blob identified by the `mount` parameter from another repository.

The following parameters SHOULD be specified on the request:

| Name             | Kind   | Description                                                          |
|------------------|--------|----------------------------------------------------------------------|
| `Host`           | header | Standard HTTP Host Header.SHOULD be set to the registry host.        |
| `Authorization`  | header | An RFC7235 compliant authorization header.                           |
| `Content-Length` | header | The `Content-Length` header MUST be zero and the body MUST be empty. |
| `name`           | path   | Name of the target repository.                                       |
| `mount`          | query  | Digest of blob to mount from the source repository.                  |
| `from`           | query  | Name of the source repository.                                       |

###### On Success: Created

```HTTP
201 Created
Location: <blob location>
Content-Length: 0
Blob-Upload-Session-ID: <session_id>
```

The blob has been mounted in the repository and is available at the provided location.

The following headers will be returned with the response:

| Name                 | Description                                                          |
|----------------------|----------------------------------------------------------------------|
| `Location`           |                                                                      |
| `Content-Length`     | The `Content-Length` header MUST be zero and the body MUST be empty. |
| `Blob-Upload-Session-ID` | Identifies the upload session_id for the current request.           |

###### On Failure: Invalid Name or Digest

```HTTP
400 Bad Request
```

The error codes that MAY be included in the response body are enumerated below:

| Code             | Message                                        | Description                                                                                                                                                                                                                                                                                       |
|------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `DIGEST_INVALID` | provided digest did not match uploaded content | When a blob is uploaded, the registry will check that the content matches the digest provided by the client.The error MAY include a detail structure with the key "digest", including the invalid digest string.This error MAY also be returned when a manifest includes an invalid layer digest. |
| `NAME_INVALID`   | invalid repository name                        | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                                                                                                       |

###### On Failure: Not allowed

```HTTP
405 Method Not Allowed
```

Blob mount is not allowed because the registry is configured as a pull-through cache or for some other reason

The error codes that MAY be included in the response body are enumerated below:

| Code          | Message                       | Description                                                                                 |
|---------------|-------------------------------|---------------------------------------------------------------------------------------------|
| `UNSUPPORTED` | The operation is unsupported. | The operation was unsupported due to a missing implementation or invalid set of parameters. |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
             "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

### Blob Upload

Interact with blob uploads.
Clients SHOULD never assemble URLs for this endpoint and SHOULD only take it through the `Location` header on related API requests.
The `Location` header and its parameters SHOULD be preserved by clients, using the latest value returned via upload related API calls.

#### GET Blob Upload

Retrieve status of upload identified by `session_id`.
The primary purpose of this endpoint is to resolve the current status of a resumable upload.

```HTTP
GET /v2/<name>/blobs/uploads/<session_id>
Host: <registry host>
Authorization: <scheme> <token>
```

Retrieve the progress of the current upload, as reported by the `Range` header.

The following parameters SHOULD be specified on the request:

| Name            | Kind   | Description                                                                                   |
|-----------------|--------|-----------------------------------------------------------------------------------------------|
| `Host`          | header | Standard HTTP Host Header.SHOULD be set to the registry host.                                 |
| `Authorization` | header | An RFC7235 compliant authorization header.                                                    |
| `name`          | path   | Name of the target repository.                                                                |
| `session_id`          | path   | A unique string identifying session of the particular upload.This field can accept characters that match `[a-zA-Z0-9-_.=]+`. |

###### On Success: Upload Progress

```HTTP
204 No Content
Range: 0-<offset>
Content-Length: 0
Blob-Upload-Session-ID: <session_id>
```

The upload is known and in progress.
The last received offset is available in the `Range` header.

The following headers will be returned with the response:

| Name                 | Description                                                          |
|----------------------|----------------------------------------------------------------------|
| `Range`              | Range indicating the current progress of the upload.                 |
| `Content-Length`     | The `Content-Length` header MUST be zero and the body MUST be empty. |
| `Blob-Upload-Session-ID` | Identifies the upload session ID for the current request.           |

###### On Failure: Bad Request

```HTTP
400 Bad Request
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

There was an error processing the upload and it MUST be restarted.

The error codes that MAY be included in the response body are enumerated below:

| Code                  | Message                                        | Description                                                                                                                                                                                                                                                                                       |
|-----------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `DIGEST_INVALID`      | provided digest did not match uploaded content | When a blob is uploaded, the registry will check that the content matches the digest provided by the client.The error MAY include a detail structure with the key "digest", including the invalid digest string.This error MAY also be returned when a manifest includes an invalid layer digest. |
| `NAME_INVALID`        | invalid repository name                        | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                                                                                                       |
| `BLOB_UPLOAD_INVALID` | blob upload invalid                            | The blob upload encountered an error and can no longer proceed.                                                                                                                                                                                                                                   |

###### On Failure: Not Found

```HTTP
404 Not Found
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The upload is unknown to the registry.
The upload MUST be restarted.

The error codes that MAY be included in the response body are enumerated below:

| Code                  | Message                         | Description                                                                                |
|-----------------------|---------------------------------|--------------------------------------------------------------------------------------------|
| `BLOB_UPLOAD_UNKNOWN` | blob upload unknown to registry | If a blob upload has been cancelled or was never started, this error code MAY be returned. |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

#### PATCH Blob Upload

Upload a chunk of data for the specified upload.

##### Stream upload

```HTTP
PATCH /v2/<name>/blobs/uploads/<session_id>
Host: <registry host>
Authorization: <scheme> <token>
Content-Type: application/octet-stream

<binary data>
```

Upload a stream of data to upload without completing the upload.

The following parameters SHOULD be specified on the request:

| Name            | Kind   | Description                                                                                   |
|-----------------|--------|-----------------------------------------------------------------------------------------------|
| `Host`          | header | Standard HTTP Host Header.SHOULD be set to the registry host.                                 |
| `Authorization` | header | An RFC7235 compliant authorization header.                                                    |
| `name`          | path   | Name of the target repository.                                                                |
| `session_id`          | path   | A unique string identifying the upload.This field can accept characters that match `[a-zA-Z0-9-_.=]+`. |

###### On Success: Data Accepted

```HTTP
202 Accepted
Location: /v2/<name>/blobs/uploads/<session_id>
Range: 0-<offset>
Content-Length: 0
Blob-Upload-Session-ID: <session_id>
```

The stream of data has been accepted and the current progress is available in the range header.
The updated upload location is available in the `Location` header.

The following headers will be returned with the response:

| Name                 | Description                                                                                                                                                                         |
|----------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Location`           | The location of the upload.Clients SHOULD assume this changes after each request.Clients SHOULD use the contents verbatim to complete the upload, adding parameters where required. |
| `Range`              | Range indicating the current progress of the upload.                                                                                                                                |
| `Content-Length`     | The `Content-Length` header MUST be zero and the body MUST be empty.                                                                                                                |
| `Blob-Upload-Session-ID` | Identifies the upload session ID for the current request.                                                                                                                          |

###### On Failure: Bad Request

```HTTP
400 Bad Request
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

There was an error processing the upload and it MUST be restarted.

The error codes that MAY be included in the response body are enumerated below:

| Code                  | Message                                        | Description                                                                                                                                                                                                                                                                                        |
|-----------------------|------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `DIGEST_INVALID`      | provided digest did not match uploaded content | When a blob is uploaded, the registry will check that the content matches the digest provided by the client. The error MAY include a detail structure with the key "digest", including the invalid digest string.This error MAY also be returned when a manifest includes an invalid layer digest. |
| `NAME_INVALID`        | invalid repository name                        | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                                                                                                        |
| `BLOB_UPLOAD_INVALID` | blob upload invalid                            | The blob upload encountered an error and can no longer proceed.                                                                                                                                                                                                                                    |

###### On Failure: Not Found

```HTTP
404 Not Found
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The upload is unknown to the registry.
The upload MUST be restarted.

The error codes that MAY be included in the response body are enumerated below:

| Code                  | Message                         | Description                                                                                |
|-----------------------|---------------------------------|--------------------------------------------------------------------------------------------|
| `BLOB_UPLOAD_UNKNOWN` | blob upload unknown to registry | If a blob upload has been cancelled or was never started, this error code MAY be returned. |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

##### Chunked upload

```HTTP
PATCH /v2/<name>/blobs/uploads/<session_id>
Host: <registry host>
Authorization: <scheme> <token>
Content-Range: <start of range>-<end of range, inclusive>
Content-Length: <length of chunk>
Content-Type: application/octet-stream

<binary chunk>
```

Upload a chunk of data to specified upload without completing the upload.
The data will be uploaded to the specified Content Range.

The following parameters SHOULD be specified on the request:

| Name             | Kind   | Description                                                                                                                                                                                                       |
|------------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Host`           | header | Standard HTTP Host Header.SHOULD be set to the registry host.                                                                                                                                                     |
| `Authorization`  | header | An RFC7235 compliant authorization header.                                                                                                                                                                        |
| `Content-Range`  | header | Range of bytes identifying the desired block of content represented by the body.Start MUST the end offset retrieved via status check plus one.Note that this is a non-standard use of the `Content-Range` header. |
| `Content-Length` | header | Length of the chunk being uploaded, corresponding the length of the request body.                                                                                                                                 |
| `name`           | path   | Name of the target repository.                                                                                                                                                                                    |
| `session_id`           | path   | A unique string identifying session of the particular upload.This field can accept characters that match `[a-zA-Z0-9-_.=]+`.                                                                                                                     |

###### On Success: Chunk Accepted

```HTTP
202 Accepted
Location: /v2/<name>/blobs/uploads/<session_id>
Range: 0-<offset>
Content-Length: 0
Blob-Upload-Session-ID: <session_id>
```

The chunk of data has been accepted and the current progress is available in the range header.
The updated upload location is available in the `Location` header.

The following headers will be returned with the response:

| Name                 | Description                                                                                                                                                                         |
|----------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Location`           | The location of the upload.Clients SHOULD assume this changes after each request.Clients SHOULD use the contents verbatim to complete the upload, adding parameters where required. |
| `Range`              | Range indicating the current progress of the upload.                                                                                                                                |
| `Content-Length`     | The `Content-Length` header MUST be zero and the body MUST be empty.                                                                                                                |
| `Blob-Upload-Session-ID` | Identifies the upload session ID for the current request.                                                                                                                          |

###### On Failure: Bad Request

```HTTP
400 Bad Request
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

There was an error processing the upload and it MUST be restarted.

The error codes that MAY be included in the response body are enumerated below:

| Code                  | Message                                        | Description                                                                                                                                                                                                                                                                                       |
|-----------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `DIGEST_INVALID`      | provided digest did not match uploaded content | When a blob is uploaded, the registry will check that the content matches the digest provided by the client.The error MAY include a detail structure with the key "digest", including the invalid digest string.This error MAY also be returned when a manifest includes an invalid layer digest. |
| `NAME_INVALID`        | invalid repository name                        | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                                                                                                       |
| `BLOB_UPLOAD_INVALID` | blob upload invalid                            | The blob upload encountered an error and can no longer proceed.                                                                                                                                                                                                                                   |

###### On Failure: Not Found

```HTTP
404 Not Found
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The upload is unknown to the registry.
The upload MUST be restarted.

The error codes that MAY be included in the response body are enumerated below:

| Code                  | Message                         | Description                                                                                |
|-----------------------|---------------------------------|--------------------------------------------------------------------------------------------|
| `BLOB_UPLOAD_UNKNOWN` | blob upload unknown to registry | If a blob upload has been cancelled or was never started, this error code MAY be returned. |

###### On Failure: Requested Range Not Satisfiable

```HTTP
416 Requested Range Not Satisfiable
```

The `Content-Range` specification cannot be accepted, either because it does not overlap with the current progress or it is invalid.

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

|Code|Message|Description|
|----|-------|-----------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

#### PUT Blob Upload

Complete the upload specified by `session_id`, optionally appending the body as the final chunk.

```HTTP
PUT /v2/<name>/blobs/uploads/<session_id>?digest=<digest>
Host: <registry host>
Authorization: <scheme> <token>
Content-Length: <length of data>
Content-Type: application/octet-stream

<binary data>
```

Complete the upload, providing all the data in the body, if necessary.
A request without a body will just complete the upload with previously uploaded content.

The following parameters SHOULD be specified on the request:

| Name             | Kind   | Description                                                                                                            |
|------------------|--------|------------------------------------------------------------------------------------------------------------------------|
| `Host`           | header | Standard HTTP Host Header.SHOULD be set to the registry host.                                                          |
| `Authorization`  | header | An RFC7235 compliant authorization header.                                                                             |
| `Content-Length` | header | Length of the data being uploaded, corresponding to the length of the request body.May be zero if no data is provided. |
| `name`           | path   | Name of the target repository.                                                                                         |
| `session_id`           | path   | A unique string identifying session of the particular upload.This field can accept characters that match `[a-zA-Z0-9-_.=]+`.                          |
| `digest`         | query  | Digest of uploaded blob.                                                                                               |

###### On Success: Upload Complete

```HTTP
204 No Content
Location: <blob location>
Content-Range: <start of range>-<end of range, inclusive>
Content-Length: 0
Docker-Content-Digest: <digest>
```

The upload has been completed and accepted by the registry.
The canonical location will be available in the `Location` header.

The following headers will be returned with the response:

| Name                    | Description                                                                                                                                                                                                       |
|-------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Location`              | The canonical location of the blob for retrieval                                                                                                                                                                  |
| `Content-Range`         | Range of bytes identifying the desired block of content represented by the body.Start MUST match the end of offset retrieved via status check.Note that this is a non-standard use of the `Content-Range` header. |
| `Content-Length`        | The `Content-Length` header MUST be zero and the body MUST be empty.                                                                                                                                              |
| `Docker-Content-Digest` | Digest of the targeted content for the request.                                                                                                                                                                   |

###### On Failure: Bad Request

```HTTP
400 Bad Request
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

There was an error processing the upload and it MUST be restarted.

The error codes that MAY be included in the response body are enumerated below:

| Code                  | Message                                        | Description                                                                                                                                                                                                                                                                                       |
|-----------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `DIGEST_INVALID`      | provided digest did not match uploaded content | When a blob is uploaded, the registry will check that the content matches the digest provided by the client.The error MAY include a detail structure with the key "digest", including the invalid digest string.This error MAY also be returned when a manifest includes an invalid layer digest. |
| `NAME_INVALID`        | invalid repository name                        | Invalid repository name encountered either during manifest validation or any API operation.                                                                                                                                                                                                       |
| `BLOB_UPLOAD_INVALID` | blob upload invalid                            | The blob upload encountered an error and can no longer proceed.                                                                                                                                                                                                                                   |
| `UNSUPPORTED`         | The operation is unsupported.                  | The operation was unsupported due to a missing implementation or invalid set of parameters.                                                                                                                                                                                                       |

###### On Failure: Not Found

```HTTP
404 Not Found
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The upload is unknown to the registry.
The upload MUST be restarted.

The error codes that MAY be included in the response body are enumerated below:

| Code                  | Message                         | Description                                                                                |
|-----------------------|---------------------------------|--------------------------------------------------------------------------------------------|
| `BLOB_UPLOAD_UNKNOWN` | blob upload unknown to registry | If a blob upload has been cancelled or was never started, this error code MAY be returned. |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |

#### DELETE Blob Upload

Cancel outstanding upload processes, releasing associated resources.
If this is not called, the unfinished uploads will eventually timeout.

```HTTP
DELETE /v2/<name>/blobs/uploads/<session_id>
Host: <registry host>
Authorization: <scheme> <token>
Content-Length: 0
```

Cancel the upload specified by `session_id`.

The following parameters SHOULD be specified on the request:

| Name             | Kind   | Description                                                                                   |
|------------------|--------|-----------------------------------------------------------------------------------------------|
| `Host`           | header | Standard HTTP Host Header.SHOULD be set to the registry host.                                 |
| `Authorization`  | header | An RFC7235 compliant authorization header.                                                    |
| `Content-Length` | header | The `Content-Length` header MUST be zero and the body MUST be empty.                          |
| `name`           | path   | Name of the target repository.                                                                |
| `session_id`           | path   | A unique string identifying session of the particular upload.This field can accept characters that match `[a-zA-Z0-9-_.=]+`. |

###### On Success: Upload Deleted

```HTTP
204 No Content
Content-Length: 0
```

The upload has been successfully deleted.

The following headers will be returned with the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

###### On Failure: Bad Request

```HTTP
400 Bad Request
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

An error was encountered processing the delete.
The client MAY ignore this error.

The error codes that MAY be included in the response body are enumerated below:

| Code                  | Message                 | Description                                                                                 |
|-----------------------|-------------------------|---------------------------------------------------------------------------------------------|
| `NAME_INVALID`        | invalid repository name | Invalid repository name encountered either during manifest validation or any API operation. |
| `BLOB_UPLOAD_INVALID` | blob upload invalid     | The blob upload encountered an error and can no longer proceed.                             |

###### On Failure: Not Found

```HTTP
404 Not Found
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The upload is unknown to the registry.
The client MAY ignore this error and assume the upload has been deleted.

The error codes that MAY be included in the response body are enumerated below:

| Code                  | Message                         | Description                                                                                |
|-----------------------|---------------------------------|--------------------------------------------------------------------------------------------|
| `BLOB_UPLOAD_UNKNOWN` | blob upload unknown to registry | If a blob upload has been cancelled or was never started, this error code MAY be returned. |

###### On Failure: Authentication Required

```HTTP
401 Unauthorized
WWW-Authenticate: <scheme> realm="<realm>", ..."
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client is not authenticated.

The following headers will be returned on the response:

| Name               | Description                                           |
|--------------------|-------------------------------------------------------|
| `WWW-Authenticate` | An RFC7235 compliant authentication challenge header. |
| `Content-Length`   | Length of the JSON response body.                     |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                 | Description                                                                                                                                                           |
|----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UNAUTHORIZED` | authentication required | The access controller was unable to authenticate the client.Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate. |

###### On Failure: No Such Repository Error

```HTTP
404 Not Found
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The repository is not known to the registry.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code           | Message                               | Description                                                                       |
|----------------|---------------------------------------|-----------------------------------------------------------------------------------|
| `NAME_UNKNOWN` | repository name not known to registry | This is returned if the name used during an operation is unknown to the registry. |

###### On Failure: Access Denied

```HTTP
403 Forbidden
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client does not have required access to the repository.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code     | Message                                    | Description                                                          |
|----------|--------------------------------------------|----------------------------------------------------------------------|
| `DENIED` | requested access to the resource is denied | The access controller denied access for the operation on a resource. |

###### On Failure: Too Many Requests

```HTTP
429 Too Many Requests
Content-Length: <length>
Content-Type: application/json

{
    "errors": [
        {
            "code": "<error code>",
            "message": "<error message>",
            "detail": ...
        },
        ...
    ]
}
```

The client made too many requests within a time interval.

The following headers will be returned on the response:

| Name             | Description                       |
|------------------|-----------------------------------|
| `Content-Length` | Length of the JSON response body. |

The error codes that MAY be included in the response body are enumerated below:

| Code              | Message           | Description                                                         |
|-------------------|-------------------|---------------------------------------------------------------------|
| `TOOMANYREQUESTS` | too many requests | Returned when a client attempts to contact a service too many times |
