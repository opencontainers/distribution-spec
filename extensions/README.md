---
tags: oci,distribution
breaks: false
---

# Extensions API for Distribution

Extensions, in general, to the distribution-spec API are allowed by the distribution-spec, with certain reservations disclosed herein.
This document outlines a path for registry implementations and clients to iterate on new APIs, establish them, and propose them to OCI for canonical reference.

The basis of the Extension API (`_oci`) should be emulated for all extensions.

## Table

_notice_: All new `/extensions/_<extention>.md` docs MUST be added to this table.

| Extension                | Summary                                              |
|:------------------------:|:----------------------------------------------------:|
| [`_oci`](./_oci.md)      | Extensions discovering extensions on registry server |
| `_catalog`               | Reserved prior extension                             |

## Name

Extension names MUST be unique.
Extensions recorded in this distribution-spec are considered canonical definitions.

Extensions are specified by extension name (`<extension>`) aligning with the project, followed by the `<component>`, and last by by the `<module>`.
This constitutes the URI segments of the extension endpoint.
Additional options may be passed as parameters to the endpoint.

```http
_<extension>/<component>/<module>[?<key>=<value>&...]
```

The values of `<extension>`, `<component>`, and `<module>` MUST conform to the following regular expression:

`[a-z0-9]+([._-][a-z0-9]+)`

Note: the leading `_` on the extension is a prefix before this regular expression.

### Registry-Level Extensions

Registry-level extensions are nested under `/v2`.

```http
GET /v2/_<extension>/<component>/<module>[?<key>=<value>&...]
```

A contrived example of a search extension may be of the form `/v2/_oci/catalog/search?pattern=foo&n=10`

### Repository-Level Extensions

Repository-level extension endpoints are scoped under a repository name.

Repository-level extensions follow the same naming rules as the [registry-level endpoints](#registry-level-extensions).

```http
GET /v2/<name>/_<extension>/<component>/<module>[?<key>=<value>&...]
```

### Reserved Namespaces

As of current, `_oci` and `_catalog` are considered as reserved namespaces and cannot be used by other extensions.

### Versioning

Data payloads being exchanged from extensions SHOULD handle versioned data structures this with [`Accept`](https://datatracker.ietf.org/doc/html/rfc7231#section-5.3.2) and [`Content-Type`](https://datatracker.ietf.org/doc/html/rfc7231#section-3.1.1.5) HTTP headers.

If an extension has fundamentally changed enough, then it SHOULD be introduced as a new `<component>` and/or `<component>/<module>`.
Whether or not there is versioning built into those names is up to the extension maintainer.

## Filename

Extension definitions SHOULD be placed under `./extensions/` in the specification repository.
Extension files SHOULD follow the `_<extension>.md` pattern.
Refer [_oci.md](./_oci.md) for more details.

## Detail

Extensions details MAY describe more endpoints and  APIs that it MAY support.
It is also recommended to call out error codes encountered and enumerated as in the
in the following table:

| Code                | Message              | Description                                         |
|---------------------|----------------------|-----------------------------------------------------|
| `EXTENSION_UNKNOWN` | Extension is unknown | This error MAY be returned when a extension is unknown to the registry in a specified repository. This can be returned with a standard get or if a manifest references an unknown layer during upload. |

## Pagination

Extensions implementing pagination, to limit the size of a list returned in a response, SHOULD consider aligning with the [pagination pattern described in the content discovery portion](../spec.md#content-discovery) of the specification.

Extension MAY provide enumeration without lexical ordering and in this case, it is not required to support the `last` parameter.
Clients are NOT RECOMMENDED to construct the `link` and SHOULD treat the URL as opaque.

## Prior Art

When considering the proposal structure for these extensions, the following processes were considered:

* [Python PEP](https://www.python.org/dev/peps/)
* [Kubernetes KEP](https://github.com/kubernetes/enhancements/tree/master/keps)
