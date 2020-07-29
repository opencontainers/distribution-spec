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
- **Manifest**: a JSON document which defines an artifact
- **Config**: a section in the manifest (and associated blob) which contains artifact metadata
- **Artifact**: one conceptual piece of content stored as blobs with an accompanying manifest containing a config
- **Digest**: a unique identifier created from a cryptographic hash of a blob's content

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
`/v2/<name>/blobs/<digest>`

`<name>` is the namespace of the repository, and `<digest>` is the blob's digest.

A GET request to an existing blob URL MUST provide the expected blob, with a reponse code that MUST be `200 OK`.

If the blob is not found in the registry, the response code MUST be `404 Not Found`.

##### Pulling manifests

To pull a manifest, perform a `GET` request to a url in the following form:
`/v2/<name>/manifests/<reference>`

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

`/v2/<name>/blobs/uploads/?digest=<digest>`
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

`/v2/<name>/blobs/uploads/`

Here, `<name>` refers to the namespace of the repository. Upon success, the response MUST have a code of `202 Accepted`, and MUST include the following header:

```
Location: <location>
```

The `<location>` MUST contain a UUID representing a unique session ID for the upload to follow.

Optionally, the location MAY be absolute (containing the protocol and/or hostname), or it MAY be relative (containing just the URL path).

Once the `<location>` has been obtained, perform the upload proper by making a `PUT` request to the following URL path, and with the following headers and body:

`<location>?digest=<digest>`
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

URL path: `<location>`
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
`/v2/<name>/manifests/<reference>`
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
`/v2/<name>/tags/list`

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
`/v2/<name>/tags/list?n=<int>`

`<name>` is the namespace of the repository, and `<int>` is an integer specifying the number of tags requested. The response
to such a request MAY return fewer than `<int>` results, but only when the total number of tags attached to the repository
is less than `<int>`. Otherwise, the response MUST include `<int>` results. Without the `last` query parameter (described
next), the list returned will start at the beginning of the list and include `<int>` results. As above, the tags MUST be
in lexical order.

The `last` query parameter provides further means for limiting the number of tags. It is used exclusively in combination with the
`n` parameter:
```
/v2/<name>/tags/list?n=<int>&last=<tagname>
```

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
`/v2/<name>/manifests/<tag>`

##### Deleting Manifests
To delete a manifest, perform a `DELETE` request to a path in the following format:
`/v2/<name>/manifests/<digest>`

`<name>` is the namespace of the repository, and `<digest>` is the digest of the manifest to be deleted. Upon success, the registry
MUST respond with a `202 Accepted` code. If the repository does not exist, the response MUST return `404 Not Found`.

##### Deleting Blobs
To delete a blob, perform a `DELETE` request to a path in the following format:
`/v2/<name>/blobs/<digest>`

`<name>` is the namespace of the repository, and `<digest>` is the digest of the blob to be deleted. Upon success, the
registry MUST respond with code `202 Accepted`. If the blob is not found, a `404 Not Found` code MUST be returned.

## Scope

This specification covers URL layout and protocols for interaction between a registry and registry client.
Registry implementations MAY implement other API endpoints, but they are not covered by this specification.

This specification includes the following features:

- Namespace-oriented URI Layout
- PUSH/PULL registry server for V2 image manifest format
- Resumable layer PUSH support
- V2 Client (Consumer) requirements

### Future

The following is an incomplete list of features, discussed during the process of cutting this specification, which MAY be out of the scope of this specification, MAY be the purview of another specification, or MAY be deferred to a future version:

- Authentication and authorization support: While authentication and authorization support will influence this specification, those details MAY be left to a future specification. However, relevant header definitions and error codes are present to provide an indication of what a client MAY encounter.
- Immutable image references
- Multiple architecture support
- Migration from v2compatibility representation

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

## Changes

The V2 specification has been written to work as a living document, specifying only what is certain and leaving what is not specified open or to future changes.
Only non-conflicting additions SHOULD be made to the API and accepted changes SHOULD avoid preventing future changes from happening.

The [changes.md](changes.md) doc SHOULD be updated when changes are made to the specification, indicating what is different.
Optionally, we MAY start marking parts of the specification to correspond with the versions enumerated here.

## Overview

This section covers client flows and details of the API endpoints.
The URI layout of the new API is structured to support a rich authentication and authorization model by leveraging namespaces.
All endpoints will be prefixed by the API version and the repository name:

```HTTP
/v2/<name>/
```

For example, an API endpoint that will work with the `library/ubuntu` repository, the URI prefix will be:

```HTTP
/v2/library/ubuntu/
```

This scheme provides rich access control over various operations and methods using the URI prefix and http methods that can be controlled in variety of ways.

Classically, repository names have always been two path components where each path component is less than 30 characters.
The V2 registry API does not enforce this.
The rules for a repository name are as follows:

1. A repository name is broken up into _path components_.
A component of a repository name MUST begin with one or more lowercase alpha-numeric characters. Subsequent lowercase alpha-numeric characters are OPTIONAL and MAY be separated by periods, dashes or underscores.
More strictly, it MUST match the regular expression `[a-z0-9]+(?:[._-][a-z0-9]+)*`.
2. If a repository name has two or more path components, they MUST be separated by a forward slash ("/").
3. The total length of a repository name, including slashes, MUST be less than 256 characters.

These name requirements _only_ apply to the registry API and SHOULD accept a superset of what is supported by other components.

All endpoints SHOULD support aggressive http caching, compression and range headers, where appropriate.
The new API attempts to leverage HTTP semantics where possible but MAY break from standards to implement targeted features.

For detail on individual endpoints, please see the [_Detail_](#detail) section.

### Errors

Actionable failure conditions, covered in detail in their relevant sections, are reported as part of 4xx responses, in a json response body.
One or more errors will be returned in the following format:

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

The `code` field will be a unique identifier, all caps with underscores by convention.
The `message` field will be a human readable string.
The OPTIONAL `detail` field MAY contain arbitrary json data providing information the client can use to resolve the issue.

While the client can take action on certain error codes, the registry MAY add new error codes over time.
All client implementations SHOULD treat unknown error codes as `UNKNOWN`, allowing future error codes to be added without breaking API compatibility.
For the purposes of the specification error codes will only be added and never removed.

For a complete account of all error codes, please see the [_Errors_](#errors-2) section.

### API Version Check

A minimal endpoint, mounted at `/v2/` will provide version support information based on its response statuses.
The request format is as follows:

```HTTP
GET /v2/
```

If a `200 OK` response is returned, the registry implements the V2(.1) registry API and the client MAY proceed safely with other V2 operations.
Optionally, the response MAY contain information about the supported paths in the response body.
The client SHOULD be prepared to ignore this data.

If a `401 Unauthorized` response is returned, the client SHOULD take action based on the contents of the "WWW-Authenticate" header and try the endpoint again.
Depending on access control setup, the client MAY still have to authenticate against different resources, even if this check succeeds.

If `404 Not Found` response status, or other unexpected status, is returned, the client SHOULD proceed with the assumption that the registry does not implement V2 of the API.

When a `200 OK` or `401 Unauthorized` response is returned, the "Docker-Distribution-API-Version" header SHOULD be set to "registry/2.0".
Clients MAY require this header value to determine if the endpoint serves this API.
When this header is omitted, clients MAY fallback to an older API version.

### Content Digests

This API design is driven heavily by [content addressability](http://en.wikipedia.org/wiki/Content-addressable_storage).
The core of this design is the concept of a content addressable identifier.
It uniquely identifies content by taking a collision-resistant hash of the bytes.

Such an identifier can be independently calculated and verified by selection of a common _algorithm_.
If such an identifier can be communicated in a secure manner, one can retrieve the content from an insecure source, calculate it independently and be certain that the correct content was obtained.
Put simply, the identifier is a property of the content.

To disambiguate from other concepts, we call this identifier a _digest_.
A _digest_ is a serialized hash result, consisting of a _algorithm_ and _hex_ portion.
The _algorithm_ identifies the methodology used to calculate the digest.
The _hex_ portion is the hex-encoded result of the hash.

We define a _digest_ string to match the following grammar:

```EBNF
digest      := algorithm ":" hex
algorithm   := /[A-Za-z0-9_+.-]+/
hex         := /[A-Fa-f0-9]+/
```

Some examples of _digests_ include the following:

| digest                                                                  | description                |
|-------------------------------------------------------------------------|----------------------------|
| `sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b` | Common sha256 based digest format |

Compliant implementations SHOULD use _algorithm_ of sha256.
The _algorithm_ does allow one to implement a wide variety of algorithms, though for portability sake it is best to keep this consistent because a translation of hashing algorithm used would produce a new object and new documents that reference it.
Heavy processing of input before calculating a hash is discouraged to avoid degrading the uniqueness of the _digest_ but some canonicalization MAY be performed to ensure consistent identifiers.

Let's use a simple example in pseudo-code to demonstrate a digest calculation:

```EBNF
let C = 'a small string'
let B = sha256(C)
let D = 'sha256:' + EncodeHex(B)
let ID(C) = D
```

Above, we have bytestring `C` passed into a function, `SHA256`, that returns a bytestring `B`, which is the hash of `C`.
`D` gets the algorithm concatenated with the hex encoding of `B`.
We then define the identifier of `C` to `ID(C)` as equal to `D`.
A digest can be verified by independently calculating `D` and comparing it with identifier `ID(C)`.

#### Digest Header

To provide verification of http content, any response MAY include a `Docker-Content-Digest` header.
This will include the digest of the target entity returned in the response.
For blobs, this is the entire blob content.
For manifests, this is the manifest body without the signature content, also known as the JWS payload.
Note that the commonly used canonicalization for digest calculation MAY be dependent on the mediatype of the content, such as with manifests.

The client MAY choose to ignore the header or MAY verify it to ensure content integrity and transport security.
This is most important when fetching by a digest.
To ensure security, the content SHOULD be verified against the digest used to fetch the content.
At times, the returned digest MAY differ from that used to initiate a request.
Such digests are considered to be from different _domains_, meaning they have different values for _algorithm_.
In such a case, the client MAY choose to verify the digests in both domains or ignore the server's digest.
To maintain security, the client MUST always verify the content against the _digest_ used to fetch the content.

> __IMPORTANT:__ If a _digest_ is used to fetch content, the client SHOULD use the same digest used to fetch the content to verify it.
> The header `Docker-Content-Digest` SHOULD NOT be trusted over the "local" digest.

### Pulling An Image

An "image" is a combination of a JSON manifest and individual layer files.
The process of pulling an image centers around retrieving these two components.

The first step in pulling an image is to retrieve the manifest.
For details on manifest formats and their content types, refer to the OCI Image Specification's [manifest property description](https://github.com/opencontainers/image-spec/blob/master/manifest.md#image-manifest-property-descriptions).

When the manifest is in hand, the client MUST verify the signature to ensure the names and layers are valid.
Once confirmed, the client will then use the digests to download the individual layers.
Layers are stored in as blobs in the V2 registry API, keyed by their digest.

#### Pulling an Image Manifest

The image manifest can be fetched with the following url:

```HTTP
GET /v2/<name>/manifests/<reference>
```

The `name` and `reference` parameter identify the image and are REQUIRED.
The reference MAY include a tag or digest.

The client SHOULD include an Accept header indicating which manifest content types it supports.
In a successful response, the Content-Type header will indicate which manifest type is being returned.

A `404 Not Found` response will be returned if the image is unknown to the registry.
If the image exists and the response is successful, the image manifest will be returned:

```json
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

The client SHOULD verify the returned manifest signature for authenticity before fetching layers.

##### Existing Manifests

The image manifest can be checked for existence with the following url:

```HTTP
HEAD /v2/<name>/manifests/<reference>
```

The `name` and `reference` parameter identify the image and are REQUIRED.
The reference MAY include a tag or digest.

A `404 Not Found` response will be returned if the image is unknown to the registry.
If the image exists and the response is successful the response will be as follows:

```HTTP
200 OK
Content-Length: <length of manifest>
Docker-Content-Digest: <digest>
```

#### Pulling a Layer

Layers are stored in the blob portion of the registry, keyed by digest.
Pulling a layer is carried out by a standard http request.
The URL is as follows:

```HTTP
GET /v2/<name>/blobs/<digest>
```

Access to a layer will be gated by the `name` of the repository but is identified uniquely in the registry by `digest`.

This endpoint MAY issue a 307 (302 for < HTTP 1.1) redirect to another service for downloading the layer and clients SHOULD be prepared to handle redirects.

This endpoint SHOULD support aggressive HTTP caching for image layers.
Support for Etags, modification dates and other cache control headers SHOULD be included.
To allow for incremental downloads, `Range` requests SHOULD be supported, as well.

### Pushing An Image

Pushing an image works in the opposite order as a pull.
After assembling the image manifest, the client MUST first push the individual layers.
When the layers are fully pushed into the registry, the client SHOULD upload the signed manifest.

The details of each step of the process are covered in the following sections.

#### Pushing a Layer

All layer uploads use two steps to manage the upload process.
The first step starts the upload in the registry service, returning a url to carry out the second step.
The second step uses the upload url to transfer the actual data.
Uploads are started with a POST request which returns a url that can be used to push data and check upload status.

The `Location` header will be used to communicate the upload location after each request.
While it won't change in the this specification, clients SHOULD use the most recent value returned by the API.
`Location` header value returned MUST either be absolute or relative as described in
[RFC 7231](https://tools.ietf.org/html/rfc7231#section-7.1.2).

##### Starting An Upload

To begin the process, a POST request SHOULD be issued in the following format:

```HTTP
POST /v2/<name>/blobs/uploads/
```

The parameters of this request are the image namespace under which the layer will be linked.
Responses to this request are covered below.

##### Existing Layers

The existence of a layer can be checked via a `HEAD` request to the blob store API.
The request SHOULD be formatted as follows:

```HTTP
HEAD /v2/<name>/blobs/<digest>
```

If the layer with the digest specified in `digest` is available, a 200 OK response will be received, with no actual body content (this is according to http specification).
The response will look as follows:

```HTTP
200 OK
Content-Length: <length of blob>
Docker-Content-Digest: <digest>
```

When this response is received, the client can assume that the layer is already available in the registry under the given name and SHOULD take no further action to upload the layer.
Note that the binary digests MAY differ for the existing registry layer, but the digests will be guaranteed to match.

##### Uploading the Layer

If the POST request is successful, a `202 Accepted` response will be returned with the upload URL in the `Location` header:

```HTTP
202 Accepted
Location: /v2/<name>/blobs/uploads/<session_id>
Range: bytes=0-<offset>
Content-Length: 0
```

The rest of the upload process can be carried out with the returned url, called the "Upload URL" from the `Location` header.
All responses to the upload url, whether sending data or getting status, will be in this format.

Though the URI format (`/v2/<name>/blobs/uploads/<session_id>`) for the `Location` header is specified, clients SHOULD treat it as an opaque url and SHOULD never try to assemble it.
While the `session_id` parameter MAY be an actual UUID, this proposal imposes no constraints on the format and clients SHOULD never impose any.

Header `Blob-Upload-Session-ID` OPTIONAL: If clients need to correlate local upload state with remote upload state, largely for resumable uploads.
Header `Docker-Upload-UUID` OPTIONAL: legacy compatibility. Not contstrained to being an official UUID.

##### Upload Progress

The progress and chunk coordination of the upload process will be coordinated through the `Range` header.
While this is a non-standard use of the `Range` header, there are examples of [similar approaches](https://developers.google.com/youtube/v3/guides/using_resumable_upload_protocol) in APIs with heavy use.
For an upload that just started, for an example with a 1000 byte layer file, the `Range` header would be as follows:

```HTTP
Range: bytes=0-0
```

To get the status of an upload, issue a GET request to the upload URL:

```HTTP
GET /v2/<name>/blobs/uploads/<session_id>
Host: <registry host>
```

The response will be similar to the above, except will return 204 status:

```HTTP
204 No Content
Location: /v2/<name>/blobs/uploads/<session_id>
Range: bytes=0-<offset>
```

Note that the HTTP `Range` header byte ranges are inclusive and that will be honored, even in non-standard use cases.

##### Monolithic Upload

A monolithic upload is simply a chunked upload with a single chunk and MAY be favored by clients that would like to avoided the complexity of chunking.
To carry out a "monolithic" upload, one can simply put the entire content blob to the provided URL:

```HTTP
PUT /v2/<name>/blobs/uploads/<session_id>?digest=<digest>
Content-Length: <size of layer>
Content-Type: application/octet-stream

<Layer Binary Data>
```

The "digest" parameter MUST be included with the PUT request.
Please see the [_Completed Upload_](#completed-upload) section for details on the parameters and expected responses.

##### Chunked Upload

To carry out an upload of a chunk, the client can specify a range header and only include that part of the layer file:

```HTTP
PATCH /v2/<name>/blobs/uploads/<session_id>
Content-Length: <size of chunk>
Content-Range: <start of range>-<end of range>
Content-Type: application/octet-stream

<Layer Chunk Binary Data>
```

There is no enforcement on layer chunk splits other than that the server MUST receive them in order.
The server MAY enforce a minimum chunk size.
If the server cannot accept the chunk, a `416 Requested Range Not Satisfiable` response will be returned and will include a `Range` header indicating the current status:

```HTTP
416 Requested Range Not Satisfiable
Location: /v2/<name>/blobs/uploads/<session_id>
Range: 0-<last valid range>
Content-Length: 0
Blob-Upload-Session-ID: <session_id>
```

If this response is received, the client SHOULD resume from the "last valid range" and upload the subsequent chunk.
A 416 will be returned under the following conditions:

- Invalid Content-Range header format
- Out of order chunk: the range of the next chunk MUST start immediately after the "last valid range" from the previous response.

When a chunk is accepted as part of the upload, a `202 Accepted` response will be returned, including a `Range` header with the current upload status:

```HTTP
202 Accepted
Location: /v2/<name>/blobs/uploads/<session_id>
Range: bytes=0-<offset>
Content-Length: 0
Blob-Upload-Session-ID: <session_id>
```

##### Completed Upload

For an upload to be considered complete, the client MUST submit a `PUT` request on the upload endpoint with a digest parameter.
If it is not provided, the upload will not be considered complete.
The format for the final chunk will be as follows:

```HTTP
PUT /v2/<name>/blobs/uploads/<session_id>?digest=<digest>
Content-Length: <size of chunk>
Content-Range: <start of range>-<end of range>
Content-Type: application/octet-stream

<Last Layer Chunk Binary Data>
```

Optionally, if all chunks have already been uploaded, a `PUT` request with a `digest` parameter and zero-length body MAY be sent to complete and validate the upload.
Multiple "digest" parameters MAY be provided with different digests.
The server MAY verify none or all of them but MUST notify the client if the content is rejected.

When the last chunk is received and the layer has been validated, the client will receive a `201 Created` response:

```HTTP
201 Created
Location: /v2/<name>/blobs/<digest>
Content-Length: 0
Docker-Content-Digest: <digest>
```

The `Location` header will contain the registry URL to access the accepted layer file.
The `Docker-Content-Digest` header returns the canonical digest of the uploaded blob which MAY differ from the provided digest.
Most clients MAY ignore the value but if it is used, the client SHOULD verify the value against the uploaded blob data.

###### Digest Parameter

The "digest" parameter is designed as an opaque parameter to support verification of a successful transfer.
For example, an HTTP URI parameter might be as follows:

```
sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b
```

Given this parameter, the registry will verify that the provided content does match this digest.

##### Canceling an Upload

An upload can be cancelled by issuing a DELETE request to the upload endpoint.
The format will be as follows:

```HTTP
DELETE /v2/<name>/blobs/uploads/<session_id>
```

After this request is issued, the upload `session_id` will no longer be valid and the registry server will dump all intermediate data.
While uploads will time out if not completed, clients SHOULD issue this request if they encounter a fatal error but still have the ability to issue an http request.

##### Cross Repository Blob Mount

A blob MAY be mounted from another repository that the client has read access to, removing the need to upload a blob already known to the registry.
To issue a blob mount instead of an upload, a POST request SHOULD be issued in the following format:

```HTTP
POST /v2/<name>/blobs/uploads/?mount=<digest>&from=<repository name>
Content-Length: 0
```

If the blob is successfully mounted, the client will receive a `201 Created` response:

```HTTP
201 Created
Location: /v2/<name>/blobs/<digest>
Content-Length: 0
Docker-Content-Digest: <digest>
```

The `Location` header will contain the registry URL to access the accepted layer file.
The `Docker-Content-Digest` header returns the canonical digest of the uploaded blob which MAY differ from the provided digest.
Most clients MAY ignore the value but if it is used, the client SHOULD verify the value against the uploaded blob data.

If a mount fails due to invalid repository or digest arguments, the registry will fall back to the standard upload behavior and return a `202 Accepted` with the upload URL in the `Location` header:

```HTTP
202 Accepted
Location: /v2/<name>/blobs/uploads/<session_id>
Range: bytes=0-<offset>
Content-Length: 0
Blob-Upload-Session-ID: <session_id>
```

This behavior is consistent with older versions of the registry, which do not recognize the repository mount query parameters.

Note: a client MAY issue a HEAD request to check existence of a blob in a source repository to distinguish between the registry not supporting blob mounts and the blob not existing in the expected repository.

##### Errors

If an 502, 503 or 504 error is received, the client SHOULD assume that the download can proceed due to a temporary condition, honoring the appropriate retry mechanism.
Other 5xx errors SHOULD be treated as terminal.

If there is a problem with the upload, a 4xx error will be returned indicating the problem.
After receiving a 4xx response (except 416, as called out above), the upload will be considered failed and the client SHOULD take appropriate action.

Note that the upload url will not be available forever.
If the upload `session_id` is unknown to the registry, a `404 Not Found` response will be returned and the client MUST restart the upload process.

#### Deleting a Layer

A layer MAY be deleted from the registry via its `name` and `digest`.
A delete MAY be issued with the following request format:

```HTTP
DELETE /v2/<name>/blobs/<digest>
```

If the blob exists and has been successfully deleted, the following response will be issued:

```HTTP
202 Accepted
Content-Length: None
```

If the blob had already been deleted or did not exist, a `404 Not Found` response will be issued instead.

If a layer is deleted which is referenced by a manifest in the registry, then the complete images will not be resolvable.

#### Pushing an Image Manifest

Once all of the layers for an image are uploaded, the client can upload the image manifest.
An image can be pushed using the following request format:

```HTTP
PUT /v2/<name>/manifests/<reference>
Content-Type: <manifest media type>

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

The `name` and `reference` fields of the response body MUST match those specified in the URL.
The `reference` field MAY be a "tag" or a "digest".
The content type SHOULD match the type of the manifest being uploaded, as specified in [manifest-v2-1.md](manifest-v2-1.md) and [manifest-v2-2.md](manifest-v2-2.md).

If there is a problem with pushing the manifest, a relevant 4xx response will be returned with a JSON error message.
Please see the [_PUT Manifest_](#put-manifest) section for details on possible error codes that MAY be returned.

If one or more layers are unknown to the registry, `BLOB_UNKNOWN` errors are returned.
The `detail` field of the error response will have a `digest` field identifying the missing blob.
An error is returned for each unknown blob.
The response format is as follows:

```json
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

### Listing Repositories

Repository listing is reserved for a future version of the distribution spec.

The `_catalog` api is reserved for historical usage. Registries MAY implement `_catalog`, but are NOT required.

### Listing Image Tags

It MAY be necessary to list all of the tags under a given repository.
The tags for an image repository can be retrieved with the following request:

```HTTP
GET /v2/<name>/tags/list
```

The response will be in the following format:

```HTTP
200 OK
Content-Type: application/json

{
  "name": "<name>",
  "tags": [
    "<tag>",
    ...
  ]
}
```

For repositories with a large number of tags, this response MAY be quite large.
If such a response is expected, one SHOULD use the pagination.

#### Pagination

Paginated tag results can be retrieved by adding an `n` parameter to the request URL, declaring that the response SHOULD be limited to `n` results. Starting a paginated flow MAY begin as follows:

```HTTP
GET /v2/<name>/tags/list?n=<integer>
```

The above specifies that a tags response SHOULD be returned, from the start of the result set, ordered lexically, limiting the number of results to `n`. The response to such a request would look as follows:

```HTTP
200 OK
Content-Type: application/json
Link: <<url>?n=<n from the request>&last=<last tag value from previous response>>; rel="next"

{
    "name": "<name>",
    "tags": [
      "<tag>",
      ...
    ]
}
```

To get the _next_ `n` entries, one can create a URL where the argument `last` has the value from `tags[len(tags)-1]`.
If there are indeed more results, the URL for the next block is encoded in an [RFC5988](https://tools.ietf.org/html/rfc5988) `Link` header, as a "next" relation.

The presence of the `Link` header communicates to the client that the entire result set has not been returned and another request MAY be issued.
If the header is not present, the client can assume that all results have been received.

> __NOTE:__ In the request template above, note that the brackets are required. For example, if the url is `http://example.com/v2/hello-world/tags/list?n=20&last=b`, the value of the header would be `<http://example.com/v2/hello-world/tags/list?n=20&last=b>; rel="next"`.
> Please see [RFC5988](https://tools.ietf.org/html/rfc5988) for details.

Compliant client implementations SHOULD always use the `Link` header value when proceeding through results linearly. The client MAY construct URLs to skip forward in the list of tags.

To get the next result set, a client would issue the request as follows, using the URL encoded in the described `Link` header:

```
GET /v2/<name>/tags/list?n=<n from the request>&last=<last tag value from previous response>
```

The above process should then be repeated until the `Link` header is no longer set in the response.

The tag list result set is represented abstractly as a lexically sorted list, where the position in that list can be specified by the query term `last`. The entries in the response start _after_ the term specified by `last`, up to `n`
entries.

The behavior of `last` is quite simple when demonstrated with an example. Let us say a repository has the following tags:

```
v1
v2
v3
v4
```

If the value of `n` is 2, _v1_ and _v2_ will be returned on the first response.
The `Link` header returned on the response will have `n` set to 2 and last set
to _v2_:

```
Link: <<url>?n=2&last=v2>; rel="next"
```

The client can then issue the request with the above value from the `Link`
header, receiving the values _v3_ and _v4_. Note that `n` may change on the second
to last response or be fully omitted, depending on the server implementation.


### Deleting an Image

An image MAY be deleted from the registry via its `name` and `reference`.
A delete MAY be issued with the following request format:

```HTTP
    DELETE /v2/<name>/manifests/<reference>
```

For deletes, `reference` MUST be a digest or the delete will fail.
If the image exists and has been successfully deleted, the following response will be issued:

```HTTP
    202 Accepted
    Content-Length: None
```

If the image had already been deleted or did not exist, a `404 Not Found` response will be issued instead.

> **Note**: When deleting a manifest from a registry version 2.3 or later, the following header MUST be used when `HEAD` or `GET`-ing the manifest to obtain the correct digest to delete:

    Accept: application/vnd.docker.distribution.manifest.v2+json

> for more details, see: [compatibility.md](https://github.com/docker/distribution/blob/master/docs/spec/manifest-v2-2.md#backward-compatibility)

## Detail

> **Note**: This section is still under construction.
> For the purposes of implementation, if any details below differ from the described request flows above, the section below SHOULD be corrected.
> When they match, this note SHOULD be removed.

The behavior of the endpoints are covered in detail in this section, organized by route and entity.
All aspects of the request and responses are covered, including headers, parameters and body formats.
Examples of requests and their corresponding responses, with success and failure, are enumerated.

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
