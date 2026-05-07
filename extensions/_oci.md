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

### Component: `repositories`

This component is for endpoints relating to repository listing operations.

This endpoint returns repositories under a namespace path, in lexical (i.e. case-insensitive alphanumeric order) or "ASCIIbetical" ([Go's `sort.Strings`](https://pkg.go.dev/sort#Strings)) order.

Repository listing MAY be retrieved with a standard `GET` as follows.

```HTTP
GET /v2/<name>/_oci/repositories
```

Registry implementations MAY allow `<name>` to be empty in order to list all repositories available on the registry.
When `<name>` is empty, the path is:

```HTTP
GET /v2/_oci/repositories
```

A registry that does not support listing repositories with an empty `<name>` MUST return a `405 Method Not Allowed` response code for this request.

When non-empty, `<name>` is a `/` delimited repository namespace path.
For this endpoint, a non-empty `<name>` is used as a namespace segment prefix and MUST include one or more complete namespace segments.
For requests with a non-empty `<name>`, returned repositories MUST have names that are equal to `<name>` or that begin with `<name>/`.
Registries MUST NOT treat a non-empty `<name>` as a partial namespace segment prefix.
For example, given repositories `foo/bar`, `foo/baz`, and `bar/baz`, a request to `/v2/foo/_oci/repositories` returns `foo/bar` and `foo/baz`.

A successful request MUST return a `200 OK` response code.
If the namespace path is invalid or the registry cannot list repositories under it, the registry MUST return an appropriate error response.
The list of repositories MAY be empty if there are no repositories under the namespace path.
Registry implementations MAY limit the repositories returned based on implementation policy, access control, or other registry-specific behavior.
The presence of a repository in the response only indicates that the registry may provide access to that repository at the time of the request.
The absence of a repository from the response does not indicate that the repository does not exist.

Upon success, the response body MUST be a JSON array of repository objects in the following format:

```HTTP
200 OK
Content-Length: <length>
Content-Type: application/json

[
    {
        "name": "foo/bar"
    },
    {
        "name": "foo/baz"
    }
]
```

Each repository object in the response MUST include the following properties:

- **`name`** *string*, OPTIONAL

  The full repository name.
  The value MUST match the repository name syntax defined for `<name>` in this specification.
  The value MUST be equal to the requested namespace path or have the requested namespace path as a complete namespace segment prefix.

Repository objects MAY include additional properties.

##### Query Parameters

The following query parameters MAY be provided:

- **`n`** *integer*, OPTIONAL

  Specifies the maximum number of results to return.
  The registry MAY return fewer results than requested if fewer matching repositories exist or a `Link` header is provided.
  Otherwise, the response MUST include `n` results.
  When `n` is zero, this endpoint MUST return an empty array and MUST NOT include a `Link` header.
  When `n` is not specified, the registry MAY apply a default limit.

- **`last`** *string*, OPTIONAL

  Specifies the `name` value of the last repository object received by the client.
  When provided, the response MUST begin non-inclusively after `last` in the ordered repository list.
  The `last` value MUST NOT be a numerical index.
  When using the `last` query parameter, the `n` parameter is OPTIONAL.

A `Link` header MAY be included in the response when additional repositories are available.
If included, the `Link` header MUST be set according to [RFC5988](https://www.rfc-editor.org/rfc/rfc5988.html) with the Relation Type `rel="next"`.

##### Pagination Example

To fetch the first page of up to 10 repositories:

```HTTP
GET /v2/<name>/_oci/repositories?n=10
```

To fetch the next page, pass the `name` value of the last repository object from the prior response as the `last` parameter:

```HTTP
GET /v2/<name>/_oci/repositories?n=10&last=foo/bar
```

When a `Link` header is provided, clients SHOULD prefer the `Link` header over constructing the next request with the `last` parameter.

## Code representations

Golang structures for these JSON structures is available at [`github.com/opencontainers/distribution-spec/specs-go/v1/extensions`](https://github.com/opencontainers/distribution-spec/tree/main/specs-go/v1/extensions/)

## Error Codes

Registry implementations MAY chose to not support any extension and the base extension MAY return the following error message.

| Code          | Message                       | Description                                                                                 |
|---------------|-------------------------------|---------------------------------------------------------------------------------------------|
| `UNSUPPORTED` | The operation is unsupported. | The operation was unsupported due to a missing implementation or invalid set of parameters. |
