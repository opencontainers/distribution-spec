## Conditional Requests

The [Distribution Specification](spec.md) describes mutable tag identifiers, which can be used when pushing, pulling, and deleting manifests.
Given that HTTP is a stateless protocol for distributed systems, the mutability of tag references leads to many possible race conditions.
Fortunately, HTTP has a mechanism for dealing with race conditions: conditional requests.

[RFC 7232](https://tools.ietf.org/html/rfc7232) defines the semantics of Conditional Requests in great detail, so this document will simply highlight the relevant portions of the RFC as applied to the Distribution Specification.

### Safely Mutating Manifests

The `If-Match` header from [section 3.1](https://tools.ietf.org/html/rfc7232#section-3.1):

>   makes the request method conditional on
>   the recipient origin server either having at least one current
>   representation of the target resource, when the field-value is "\*",
>   or having a current representation of the target resource that has an
>   entity-tag matching a member of the list of entity-tags provided in
>   the field-value.

Specifically interesting for mutating manifests:

>   If-Match is most often used with state-changing methods (e.g., POST,
>   PUT, DELETE) to prevent accidental overwrites when multiple user
>   agents might be acting in parallel on the same resource (i.e., to
>   prevent the "lost update" problem).  It can also be used with safe
>   methods to abort a request if the selected representation does not
>   match one already stored (or partially stored) from a prior request.

A client wishing to safely mutate a manifest SHOULD include in the manifest PUT request the following header:

```
If-Match: "<ETag>"
```

Where `<ETag>` is the entity-tag that matches the representation of the manifest as expected by the client, i.e. the header returned by the registry, as described in [section 2.3](https://tools.ietf.org/html/rfc7232#section-2.3):

>   The "ETag" header field in a response provides the current entity-tag
>   for the selected representation, as determined at the conclusion of
>   handling the request.  An entity-tag is an opaque validator for
>   differentiating between multiple representations of the same
>   resource, regardless of whether those multiple representations are
>   due to resource state changes over time, content negotiation
>   resulting in multiple representations being valid at the same time,
>   or both.  An entity-tag consists of an opaque quoted string, possibly
>   prefixed by a weakness indicator.

If the state of the manifest in the registry does not match the supplied ETag, the registry MUST return a `412 Precondition Failed` response.

### Avoiding Overwriting Tags

The `If-None-Match` header from [section 3.2](https://tools.ietf.org/html/rfc7232#section-3.2):

>   makes the request method conditional
>   on a recipient cache or origin server either not having any current
>   representation of the target resource, when the field-value is "\*",
>   or having a selected representation with an entity-tag that does not
>   match any of those listed in the field-value

Specifically interesting for avoiding tag overwriting:

>   If-None-Match can also be used with a value of "\*" to prevent an
>   unsafe request method (e.g., PUT) from inadvertently modifying an
>   existing representation of the target resource when the client
>   believes that the resource does not have a current representation
>   (Section 4.2.1 of [RFC7231]).  This is a variation on the "lost
>   update" problem that might arise if more than one client attempts to
>   create an initial representation for the target resource.

A client wishing to avoid overwriting any existing tags SHOULD include in the manifest PUT request the following header:

```
If-None-Match: *
```

If there already exists a manifest in the registry with a matching tag identifier, the registry MUST return a `412 Precondition Failed` response.
