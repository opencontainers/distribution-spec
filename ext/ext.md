# _ext - Index of Distribution Extensions

## Summary

This base extension is used to discover and return an array of extensions.

## Reference Explanation

The base extension may be queried through a `GET` to discover available
extensions.

Registry level extensions may be discovered with a standard GET as follows.

```HTTP
    GET /v2/_ext/
```

Repository level extensions may be discovered with a standard GET as follows.

```HTTP
    GET /v2/{name}/_ext/
```

The base extension returns an array of supported extensions with details of the
endpoints as shown below.

```HTTP
200 OK
Content-Length: <length>
Content-Type: application/json

{
    "extensions: [
        "_<ns>/<ext>/<component>",
    ]    
}
```

### *Extensions* Property Descriptions

- **`extensions`** *array of strings*

    This REQUIRED property contains a list of supported extension endpoints.

    The [relative URI]((https://tools.ietf.org/doc/html/rfc1808))
    which MAY be constructed by replacing
    the base discovery extension path,`_ext`.

## Error Codes

Registry implementations MAY chose to not support any extension and the base
extension MAY return the following error message.

| Code          | Message                       | Description                                                                                 |
|---------------|-------------------------------|---------------------------------------------------------------------------------------------|
| `UNSUPPORTED` | The operation is unsupported. | The operation was unsupported due to a missing implementation or invalid set of parameters. |
