---
title: "HTTP API V2"
description: "Specification for the Registry API."
keywords: registry, on-prem, images, tags, repository, distribution, api, advanced
---
# Open Container Initiative

## Table of Contents

- [Overview](#overview)
	- [Introduction](#introduction)
	- [Historical Context](#historical-context)
	- [Definitions](#defintions)
- [Conformance](#conformance)
	- [Notational Conventions](#notational-conventions)
	- [Minimum Requirements](#minimum-requirements)
	- [Official Certification](#official-certification)
	- [Workflow Categories](#workflow-categories)
		1. [Pull](#pull)
		2. [Push](#push)
		3. [Content Discovery](#content-discovery)
		4. [Content Management](#content-management)
- [HTTP API](#http-api)
	- [Endpoints](#endpoints)
	- [Error Codes](#error-codes)

## Overview

### Introduction

The Open Container Initiative Distribution Specification defines an API protocol to facilitate and standardize the distribution of content, especially related to container images and associated artifacts.

### Historical Context

The spec is based on the specification for the [Docker Registry HTTP API V2 protocol](https://github.com/docker/distribution/blob/5cb406d511b7b9163bff9b6439072e4892e5ae3b/docs/spec/api.md).

For relevant details and a history leading up to this specification, please see the following issues:

- [moby/moby#8093](https://github.com/moby/moby/issues/8093)
- [moby/moby#9015](https://github.com/moby/moby/issues/9015)
- [docker/docker-registry#612](https://github.com/docker/docker-registry/issues/612)

### Definitions

Several terms are used frequently in this document and warrant basic definitions:

- **Registry**: a service that handles the required APIs defined in this specification
- **Client**: a tool that communicates with registries
- **Push**: the act of uploading blobs and manifests to a registry
- **Pull**: the act of downloading blobs and manifests from a registry
- **Blob**: the binary form of content that is stored by a registry, addressable by a digest
- **Manifest**: a JSON document which defines an artifact. Manifests are defined under the [OCI Image Spec](https://github.com/opencontainers/image-spec/blob/master/manifest.md)
- **Config**: a section in the manifest (and associated blob) which contains artifact metadata
- **Artifact**: one conceptual piece of content stored as blobs with an accompanying manifest containing a config
- **Digest**: a unique identifier created from a cryptographic hash of a blob's content. Digests are defined under the [OCI Image Spec](https://github.com/opencontainers/image-spec/blob/b6e51fa50549ee0bd5188494912a7f4c382cb0d4/descriptor.md#digests)
- **Tag**: a custom, human-readable manifest identifier

## Conformance

### Notational Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be interpreted as described in [RFC 2119](http://tools.ietf.org/html/rfc2119) (Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", BCP 14, RFC 2119, March 1997).

### Requirements

Registries conforming to this specification MUST handle all APIs required by the following workflow categories:

1. **Pull** - Clients are able to pull from the registry
2. **Push** (OPTIONAL) - Clients are able to push to the registry
3. **Content Discovery** (OPTIONAL) - Clients are able to list or otherwise query the content stored in the registry
4. **Content Management** (OPTIONAL) - Clients are able to control the full life-cycle of the content stored in the registry

In order to test a registry's conformance against these workflow categories, please use the [conformance testing tool](./conformance/).

### Official Certification

Registry providers can self-certify by submitting conformance results to [opencontainers/oci-conformance](https://github.com/opencontainers/oci-conformance).

### Workflow Categories

#### Pull

##### Pulling Blobs

To pull a blob, perform a `GET` request to a url in the following form:
[2a](#Endpoints) `/v2/<name>/blobs/<digest>`

`<name>` is the namespace of the repository, and `<digest>` is the blob's digest.

A GET request to an existing blob URL MUST provide the expected blob, with a reponse code that MUST be `200 OK`.

If the blob is not found in the registry, the response code MUST be `404 Not Found`.

##### Pulling manifests

To pull a manifest, perform a `GET` request to a url in the following form:
[3a](#Endpoints) `/v2/<name>/manifests/<reference>`

`<name>` refers to the namespace of the repository. `<reference>` MUST be either (a) the digest of the manifest or (b) a tag name.

The `<reference>` MUST NOT be in any other format.

A GET request to an existing manifest URL MUST provide the expected manifest, with a response code that MUST be `200 OK`.

If the manifest is not found in the registry, the response code MUST be `404 Not Found`.

#### Push

##### Pushing blobs

There are two ways to push blobs: chunked or monolithic.

##### Pushing a blob monolithically

There are two ways to push a blob monolithically:
1. A single `POST` request
2. A `POST` request followed by a `PUT` request

---

To push a blob monolithically by using a single POST request, perform a `POST` request to a URL in the following form, and with the following headers and body:

[4b](#Endpoints) `/v2/<name>/blobs/uploads/?digest=<digest>`
```
Content-Length: <length>
Content-Type: application/octet-stream
```
```
<upload byte stream>
```

Here, `<name>` is the repository's namespace, `<digest>` is the blob's digest, and `<length>` is the size (in bytes) of the blob.

The `Content-Length` header MUST match the blob's actual content length. Likewise, the `<digest>` MUST match the blob's digest.

Successful completion of the request MUST return a `201 Created`, and MUST include the following header:

```
Location: <blob-location>
```

With `<blob-location>` being a pullable blob URL.

---

To push a blob monolithically by using a POST request followed by a PUT request, there are two steps:
1. Obtain a session id (upload URL)
2. Upload the blob to said URL

To obtain a session ID, perform a `POST` request to a URL in the following format:

[4a](#Endpoints) `/v2/<name>/blobs/uploads/`

Here, `<name>` refers to the namespace of the repository. Upon success, the response MUST have a code of `202 Accepted`, and MUST include the following header:

```
Location: <location>
```

The `<location>` MUST contain a UUID representing a unique session ID for the upload to follow.

Optionally, the location MAY be absolute (containing the protocol and/or hostname), or it MAY be relative (containing just the URL path).

Once the `<location>` has been obtained, perform the upload proper by making a `PUT` request to the following URL path, and with the following headers and body:

[6a](#Endpoints) `<location>?digest=<digest>`
```
Content-Length: <length>
Content-Type: aplication/octet-stream
```
```
<upload byte stream>
```

The `<location>` MAY contain critical query parameters. Additionally, it SHOULD match exactly the `<location>` obtained from the `POST` request. It SHOULD NOT be assembled manually by clients except where absolute/relative conversion is necessary.

Here, `<digest>` is the digest of the blob being uploaded, and `<length>` is its size in bytes.

Upon successful completion of the request, the response MUST have code `201 Created` and MUST have the following header:

```
Location: <blob-location>
```

With `<blob-location>` being a pullable blob URL.

##### Pushing a blob in chunks

A chunked blob upload is accomplished in three phases:
1. Obtain a session ID (upload URL) (`POST`)
2. Upload the chunks (`PATCH`)
3. Close the session (`PUT`)

For information on obtaining a session ID, reference the above section on pushing a blob monolithically via the `POST`/`PUT` method. The process remains unchanged for chunked upload, except that the post request MUST include the following header:

```
Content-Length: 0
```

Please reference the above section for restrictions on the `<location>`.

---
To upload a chunk, issue a `PATCH` request to a URL path in the following format, and with the following headers and body:

URL path: [5a](#Endpoints) `<location>`
```
Content-Type: application/octet-stream
Content-Range: <range>
Content-Length: <length>
```
```
<upload byte stream of chunk>
```

The `<location>` refers to the URL obtained from the preceding `POST` request.

The `<range>` refers to the byte range of the chunk, and MUST be inclusive on both ends.  The first chunk's range MUST begin with `0`. It MUST match the following regular expression:

```regexp
^[0-9]+-[0-9]+$
```

The `<length>` is the content-length, in bytes, of the current chunk.

Each successful chunk upload MUST have a `202 Accepted` response code, and MUST have the following header:

```
Location <location>
```

Each consecutive chunk upload SHOULD use the `<location>` provided in the response to the previous chunk upload.

Chunks MUST be uploaded in order, with the first byte of a chunk being the last chunk's `<end-of-range>` plus one. If a chunk is uploaded out of order, the registry MUST respond with a `416 Requested Range Not Satisfiable` code.

The final chunk MAY be uploaded using a `PATCH` request or it MAY be uploaded in the closing `PUT` request. Regardless of how the final chunk is uploaded, the session MUST be closed with a `PUT` request.

---

To close the session, issue a `PUT` request to a url in the following format, and with the following headers (and optional body, depending on whether or not the final chunk was uploaded already via a `PATCH` request):

`<location>?digest=<digest>`
```
Content-Length: <length of chunk, if present>
Content-Range: <range of chunk, if present>
Content-Type: application/octet-stream <if chunk provided>
```
```
OPTIONAL: <final chunk byte stream>
```

The closing `PUT` request MUST include the `<digest>` of the whole blob (not the final chunk) as a query parameter.

The response to a successful closing of the session MUST be `201 Created`, and MUST contain the following header:
```
Location: <blob-location>
```

Here, `<blob-location>` is a pullable blob URL.


##### Pushing Manifests

To push a manifest, perform a `PUT` request to a path in the following format, and with the following headers
and body:
[7a](#Endpoints) `/v2/<name>/manifests/<reference>`
```
Content-Type: application/vnd.oci.image.manifest.v1+json
```
```
<manifest byte stream>
```

`<name>` is the namespace of the repository, and the `<reference>` MUST be either a) a digest or b) a tag.

The uploaded manifest MUST reference any layers that make up the artifact. However, the layers field MAY
be empty. Upon a successful upload, the registry MUST return response code `201 Created`, and MUST have the
following header:

```
Location: <location>
```

The `<location>` is a pullable manifest URL.

An attempt to pull a nonexistent repository MUST return response code `404 Not Found`

#### Content Discovery

Currently, the only functionality provided by this workflow is the ability to discover tags.

To fetch the list of tags, perform a `GET` request to a path in the following format:
[8a](#Endpoints) `/v2/<name>/tags/list`

`<name>` is the namespace of the repository. Assuming a repository is found, this request MUST return a
`200 OK` response code. The list of tags MAY be empty, if there are no tags on the repository. If the list is not empty,
the tags MUST be in lexical order (i.e. case-insensitive alphanumeric order).

Upon success, the response MUST be a json body in the following format:
```json
{
  "name": "<name>",
  "tags": [
    "<tag1>",
    "<tag2>",
    "<tag3>"
  ]
}
```

`<name>` is the namespace of the repository, and `<tag1>`, `<tag2>`, and `<tag3>` are each tags on the repository.

In addition to fetching the whole list of tags, a subset of the tags can be fetched by providing the `n` query parameter.
In this case, the path will look like the following:
[8b](#Endpoints) `/v2/<name>/tags/list?n=<int>`

`<name>` is the namespace of the repository, and `<int>` is an integer specifying the number of tags requested. The response
to such a request MAY return fewer than `<int>` results, but only when the total number of tags attached to the repository
is less than `<int>`. Otherwise, the response MUST include `<int>` results. Without the `last` query parameter (described
next), the list returned will start at the beginning of the list and include `<int>` results. As above, the tags MUST be
in lexical order.

The `last` query parameter provides further means for limiting the number of tags. It is used exclusively in combination with the
`n` parameter:
[8b](#Endpoints) `/v2/<name>/tags/list?n=<int>&last=<tagname>`

`<name>` is the namespace of the repository, `<int>` is the number of tags requested, and `<tagname>` is the *value* of
the last tag. `<tagname>` MUST NOT be a numerical index, but rather it MUST be a proper tag. A request of this sort will return
up to `<int>` tags, beginning non-inclusively with `<tagname>`. That is to say, `<tagname>` will not be included in the
results, but up to `<int>` tags *after* `<tagname>` will be returned. The tags MUST be in lexical order.

#### Content Management
Content management refers to the deletion of blobs, tags, and manifests. Registries MAY implement deletion or they MAY
disable it. Similarly, a registry MAY implement tag deletion, while others MAY allow deletion only by manifest.

##### Deleting tags
`<name>` is the namespace of the repository, and `<tag>` is the name of the tag to be deleted. Upon success, the registry
MUST respond with a `202 Accepted` code. If tag deletion is disabled, the registry MUST respond with either a
`400 Bad Request` or a `405 Method Not Allowed`.

To delete a tag, perform a `DELETE` request to a path in the following format:
[9a](#Endpoints) `/v2/<name>/manifests/<tag>`

##### Deleting Manifests
To delete a manifest, perform a `DELETE` request to a path in the following format:
[9a](#Endpoints) `/v2/<name>/manifests/<digest>`

`<name>` is the namespace of the repository, and `<digest>` is the digest of the manifest to be deleted. Upon success, the registry
MUST respond with a `202 Accepted` code. If the repository does not exist, the response MUST return `404 Not Found`.

##### Deleting Blobs
To delete a blob, perform a `DELETE` request to a path in the following format:
[10a](#Endpoints) `/v2/<name>/blobs/<digest>`

`<name>` is the namespace of the repository, and `<digest>` is the digest of the blob to be deleted. Upon success, the
registry MUST respond with code `202 Accepted`. If the blob is not found, a `404 Not Found` code MUST be returned.

### API
The API operates over HTTP. Below is a summary of the endpoints used by the API.

#### Determining Support
To check whether or not the registry implements this specification, perform a `GET` request to the following endpoint:
[1a](#Endpoints) `/v2/`.

If the response is `200 OK`, then the registry implements this specification.

This endpoint MAY be used for authentication/authorization purposes, but this is out of the purview
of this specification.

#### Endpoints
| ID | Method | API endpoint | Accepted Successful Response Codes | Accepted Failure Response Codes |
| ---| --- | ---|---|---|
| 1a | `GET` | `/v2/` | `200` | `404`/`401` |
| 2a | `GET` | `/v2/<name>/blobs/<digest>` | `200` | `404` |
| 3a | `GET` | `/v2/<name>/manifests/<reference>` | `200` | `404` |
| 4a | `POST` | `/v2/<name>/blobs/uploads/` | `202` | `404` |
| 4b | `POST` | `/v2/<name>/blobs/uploads/?digest=<digest>` | `201` | `404`/`400` |
| 5a | `PATCH` | `/v2/<name>/blobs/uploads/<reference>` | `202` | `404`/`416` |
| 6a | `PUT` | `/v2/<name>/blobs/uploads/<reference>?digest=<digest>` | `201` | `404`/`400` |
| 7a | `PUT` | `/v2/<name>/manifests/<reference>` | `201` | `404` |
| 8a | `GET` | `/v2/<name>/tags/list` | `200`  | `404` |
| 8b | `GET` | `/v2/<name>/tags/list?n=<integer>&last=<integer>` | `200` | `404` |
| 9a | `DELETE` | `/v2/<name>/manifests/<reference>` | `202` | `404`/`400`/`405` |
| 10a | `DELETE` | `/v2/<name>/blobs/<digest>` | `202` | `404`/`405` |

#### Error Codes

A `4XX` response code from the registry MAY return a body in any format. If the response body is in JSON format, it MUST
have the following format:

```json
    {
        "errors": [
            {
                "code": "<error identifier>",
                "message": "<message describing condition>",
                "detail": "<unstructured>"
            },
            ...
        ]
    }
```

The `code` field MUST be a unique identifier, containing only uppercase alphabetic characters and underscores. The
`message` field is OPTIONAL, and if present, it SHOULD be a human readable string or MAY be empty. The `detail` field is
OPTIONAL and MAY contain arbitrary JSON data providing information the client can use to resolve the issue.

The `code` field MUST be one of the following:

| Code                    | Description                                    |
|-------------------------|------------------------------------------------|
| `BLOB_UNKNOWN`          | blob unknown to registry                       |
| `BLOB_UPLOAD_INVALID`   | blob upload invalid                            |
| `BLOB_UPLOAD_UNKNOWN`   | blob upload unknown to registry                |
| `DIGEST_INVALID`        | provided digest did not match uploaded content |
| `MANIFEST_BLOB_UNKNOWN` | blob unknown to registry                       |
| `MANIFEST_INVALID`      | manifest invalid                               |
| `MANIFEST_UNKNOWN`      | manifest unknown                               |
| `MANIFEST_UNVERIFIED`   | manifest failed signature verification         |
| `NAME_INVALID`          | invalid repository name                        |
| `NAME_UNKNOWN`          | repository name not known to registry          |
| `SIZE_INVALID`          | provided length did not match content length   |
| `TAG_INVALID`           | manifest tag did not match URI                 |
| `UNAUTHORIZED`          | authentication required                        |
| `DENIED`                | requested access to the resource is denied     |
| `UNSUPPORTED`           | the operation is unsupported                  |
