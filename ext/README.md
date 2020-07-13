# Extensions

The basis of the Extension API is described in a document which should be emulated for all extensions.

## Table of Contents

<!-- toc -->
* [Table](#table)
* [Name](#name)
* [Filename](#filename)
* [Detail](#detail)
* [Prior Art](#prior-art)
<!-- /toc -->

## Table

_notice_: All new `./ext/ext-$name.md` docs MUST be added to this table.

| `$name` (and definition) | Summary |
|:--:|:--:|
| [0](./ext-0.md) | Extensions discovering extensions on registry server |
|  |  |


## Name

Extension names MUST be unique.
Names SHOULD include a version.

Each extension's endpoints will be nested below its name.

```HTTP
    GET /v2/ext/0/...
```

## Filename

XXX

## Detail

XXX acceptable error codes

## Prior Art

When considering the proposal structure for these extensions, the following processes were considered:

* [Python PEP](https://www.python.org/dev/peps/)
* [Kubernetes KEP](https://github.com/kubernetes/enhancements/tree/master/keps)
