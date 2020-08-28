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
digest	    := algorithm ":" hex
algorithm   := /[A-Za-z0-9_+.-]+/
hex	    := /[A-Fa-f0-9]+/
```

Some examples of _digests_ include the following:

| digest								  | description		       |
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
