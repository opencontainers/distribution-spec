---
tags: oci,distribution
breaks: false
---

# Extensions API for Distribution

The basis of the Extension API is described in a document which should
be emulated for all extensions.

## Table

_notice_: All new `./ext/ext-$ns-$name.md` docs MUST be added to this table.

| `$ns-$name`                  | Summary                                              |
|:------------------------:|:----------------------------------------------------:|
| [ext](./ext.md)          | Extensions discovering extensions on registry server |

## Name

Extension names MUST be unique. Extensions are  specified by
`namespace` aligning with the project, followed by the `extension` provided by the project and last by by the `component`. This constitutes the URI segments
of the extension endpoint. Additional options may be passed as parameters to the endpoint.

```http
_<ns>/<ext>/<component>[?<key>=<value>&...]
```

For versioning, extensions SHOULD use versioned `mediaType`.

### Registry Level Extensions

Registry level extensions are nested under `/v2`.

```http
GET /v2/_<ns>/<ext>/<component>[?<key>=<value>&...]
```

For example a search extension may be of the form `/v2/_oci/catalog/search?pattern=foo?n=10`

### Repository Level Extensions

Repository level extensions follow the same naming rules the endpoints
can be access under a repository.

```http
GET /v2/<name>/_<ns>/<ext>/<component>[?<key>=<value>&...]
```

## Filename

Extension definitions SHOULD be placed under `./ext/`. Extension files
SHOULD follow the `ext-$ns-$name.md`. Refer [ext.md](./ext.md) for more details.

## Detail

Extensions details MAY describe more endpoints and  APIs that it MAY support.
It is also recommended to call out error codes encountered and enumerated as in the
in the following table:

| Code                | Message              | Description                                                                                                                                                                                            |
|---------------------|----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `EXTENSION_UNKNOWN` | Extension is unknown | This error MAY be returned when a extension is unknown to the registry in a specified repository. This can be returned with a standard get or if a manifest references an unknown layer during upload. |

## Pagination

Extensions implementing pagination and SHOULD align with the
[pagination](./spec.md#pagination) specification.

Extension MAY provide enumeration without lexical ordering and in this case,
it is not required to support the `last` parameter. Clients are NOT RECOMMENDED to construct the `link` and SHOULD treat the URL as opaque.

## Prior Art

When considering the proposal structure for these extensions, the following processes were considered:

* [Python PEP](https://www.python.org/dev/peps/)
* [Kubernetes KEP](https://github.com/kubernetes/enhancements/tree/master/keps)
