# Generic format for registry documents

## Why do we need standard registry document formats anyway?

A registry is a data store for content addressed data. At the lowest level it is just a large key (hash) to value (document) store, but one of the things we have learned about data stores is that they are more useful if they allow for structured data (a Merkle tree, technically a dag), with data being allowed to point to (hashes of) other data items. The complexity added by this is that if the data store needs to follow the links in documents it needs to know how to parse them. The primary use case for this parsing is for garbage collection: the usual storage model allows for an object that is not referenced internally or from a tag (a generic name that can be given to items in the store) may be garbage collected. Without this it is difficult to remove any items from the store. Tags exist to give human friendly names and to anchor items into the store while they exist.

Probably the best developed system along these lines is git. Git makes some different tradeoffs than registries. In particular it is optimised for handling many small (text) files, and serving these on a relatively small scale. There are only four types of content, commits, trees, blobs and annotated tags. Files are aggregated by into larger chunks (packfiles with indexes) in order to serve them more efficiently, and this is done dynamically to reduce network traffic. Git has well developed ways of handling metadata, which will be contrasted later. Like registries, the object model has got more complex over times, with trees now being able to point at commits not just trees or blobs. A registry is generally designed to handle larger files, with fewer links between them, and much higher volumes of traffic so the server is required to do less work. Over time there is likely to be a convergence, as there are use cases for which the git model of more aggressive file deduplication is useful (for example constrained network use cases), and we are likely to see more traffic akin to branch traversals in registries.

Registry formats started off just for a very specific container image use case. Originally there was one special document type, the image manifest, which pointed at a list of blobs, and a configuration blob. To find all the referenced blobs, you only had to track all the image manifests that had tags pointing at them, and then find all the blobs pointed at by those. Then the image index was added, which pointed at image manifests, for multi architecture selection. Technically this could also point at other image manifests, although this has not used much. A garbage collector had to traverse these additional links as well. The issue with these two formats is that they were very specific to container images, and although they have some ability to add generic metadata, it is difficult to adapt them to new types of stored data. The OCI Artifacts specification still requires objects to be defined as a configuration blob and a list of blobs. The registry may be expected to parse the configuration blob for display purposes. However there is no way to define objects that point to other objects, rather than blobs, so the format is not very generic. An [OCI artifact manifest](https://github.com/opencontainers/artifacts/pull/29) has been proposed that supports blobs and references, but even that is not fully generic, and more formats a re likely to be needed in future.

If we look at the problem from the point of view of not artifact authors and current usage but from the point of view of what metadata the registry needs to operate garbage collection and traversal operations, what we need is very simple. The registry needs to have blob pointers, and it needs to know about pointers to other items. This is essentially the blobs and references lists from the artifact manifest. However there is no need to distinguish the config blob from other blobs other than with metadata.

From the point of view of someone describing a new type of artifact, it is flexibility of attaching metadata that matters most. You may want links to blobs and other objects, labelled with their types and use cases. Examples of things that people want to construct now are a link to another object and a bill of materials for that object, or a manifest for a piece of software for multiple architectures, all combined in a single object so there is a choice of layers to download. You might also want to add an existing format that has links in the underlying blobs but you need to include them in the manifest so that they are visible to the registry without teaching it about the underlying format. So the important thing is that there are highly flexible ways to attach metadata to describe every link and blob, and anything else that is useful for the format to include to avoid having to parse more blobs.

So to meet both use cases, we should have a format that has a lot of extensible metadata but it has a very simple structure for the operator to use for management operations. Let us look at the management side data model first.

- **`mediaType`** *string*

  This field contains the `mediaType` of this document. This MUST be `application/vnd.oci.object.manifest.v1+json` for the JSON encoding or `application/vnd.oci.object.manifest.v1+jwt` for JSON web signature (RFC7515) encoding. The server needs this to return the media type to the application requesting this object.

- **`blobs`** *array of objects*

  An optional set of references to blobs. Each object MUST have a descriptor. From the server point of view the only pieces that matter for the descriptor are **`digest`**, used to find the object in the content store, and **`size`** which is used to check the expected size if the server needs to fetch the object, and potentially for hash collision attack detection. The server data model does not need to know the media type, as it simply needs to know this is a blob so it does not need to track further links. The **`urls`** property in the descriptor, and the **`annotations`** are just for client side use. Note there is no seperate model for `config` here; as far as the server data model is concerned this is simply another blob. Below we will discuss how the client distinguishes the different sorts of blob. The server is also not concerned about the ordering of blobs or references, so it can store these in a relation.

- **`references`** *array of objects*

  An optional set of references to other artifacts, of this type or of other supported types that may themselves point to blobs or other references. Each object MUST have a descriptor. The descriptor MUST have **`digest`** and **`size`**. These will be traversed for garbage collection. A registry will generally want to reject uploads of manifests with references that point to objects that it cannot parse or interpret as valid.

This is all the data that is needed to manage the content store from the registry operators side, and this can be used as the data model. It is easy to see how existing image manifests and index can be mapped down to this data model. However, it is too minimal for the client to be able to process it so we need to enhance it with metadata. This is where the design space gets much more complicated. We will have simple processing rules that let the server simply extract this data model from a format that is more suitable for client processing.

The existing formats assume that one document corresponds to one version of one specification. This has caused all sorts of issues, such as how to add new forms of compression to container images in a way that new clients can get enhancements while old clients continue to work. The only partial upgrade path that has worked has been for multi arch images, where an additional layer of indirection is required for a client to make a choice of which version it supports, although here the specification is very rigid on the set of allowed choices and existing clients only worked if pointed at the single architecture image. Because the registry is a content addressed store, content negotiation does not work well either, as the client often needs to check the content hash; the only place where it has been used was on upgrading to content adressablity, and it should not be used again.

For that reason this proposal has a much more generic approach, such that a document can contain several different versions of a single piece of content, suitable for different users or clients. This also means that it could include both a container image and a Helm chart; this would generally be discouraged but is a necessary byproduct of a flexible schema that supports generic upgrade paths. Upgrade will never work smoothly unless you can include both old and new versions in the same object and let a client pick the one that it understands. This proposal is also generic enough to support mappings of all existing artifact types to its format, as it is sufficiently generic, and we would recommend that we switch to using this for every type of object, rather than having specialised types. A registry can internally convert other types to this format before processing, to make them conform to the data model of this object manifest type.

We also want to improve the structure of generic metadata that applications can use for their own purposes. In particular we want to have another optional generic list of items that make up an object that does not include descriptors, in addition to `blobs` and `references` above. These could be used to refer to a list of components that are not necessarily in the repository, or referring to components by tag, or components that are purely inline in the manifest, such as data URLs or additional signatures. This allows additional flexibility. In addition the current annotations are restrictive in that they do not allow annotations to be lists, while some of the existing structures that we want to potentially map to annotations are lists. For this reason we allow all annotations to be lists (or if simpler, require them all to be singleton lists if only one item). For example an item that applies to both Linux and FreeBSD but not windows could be tagged with the list of operating systems it supports. We want to be able to represent something like a config within an object if that is more conevnient than using an external reference. This can be done by using the generic item list as well.


Here is a draft of a generic metadata format that should work for any sort of content in future. Below we show how existing manifests map to it and some generic ones that clients should understand. This is an outline design of what I think is approximately the right level of generality, but it can probably be tweaked to make things easier for clients. The design requirements are as follows:
1. Any number of objects can be stored in a manifest. This can include different versions of the same image format, either for schema upgrades, or different versions for different architectures as we have for image index now. Other use cases are possible such as an index document that simply covers a number of blobs, or a Helm chart that has the chart information and a set of images in the standrad image format in the same manifest.
2. Clients rank their preference for objects and search for a match, much as happens with multi arch now. Clients will generally prefer newer formats over old ones, and their specific architecture matches over worse ones. Generic clients might have user configuration. Note that multi arch can be included in a single manifest rather than indirecting to image manifests, making fetching more efficient.
3. A few generic object types are supported. Notably these include "Pointer", "Relation" and "Property". Pointer is simply an object that points at another one which should be dereferenced. A signed Pointer object can be used for a detached signature. Relation is a pointer, a type, and a second object which is related, used for adding metadata such as an SBOM to an existing object; again a client that is not interested in the SBOM should just follow the link to the primary object. "Property" is like a pointer but is stored within a manifest and links to a property of that manifest itself, used if you want to say store an SBOM at build time in the same object, to avoid an extra redirection.
4. We support items in a manifest that have neither blobs nor references, so that the server side does not care about these. This allows manifests to do things like reference foreign layers or images by tag that are not actually references at all, but make the client side processing more consistent.
5. User interface specific items, such as links to icons for the image to be displayed can be added as an additional object that the client will ignore but can be used by the registry UI only.
6. The manifest as a whole does not have a type, only the objects in it.

- **`schemaVersion`** *int*

  This is a REQUIRED property. The design of this document is such that this should not need to be bumped.

- **`mediaType`** *string*

  This field contains the `mediaType` of this document. This MUST be `application/vnd.oci.object.manifest.v1+json` for the JSON encoding or `application/vnd.oci.object.manifest.v1+jwt` for JSON web signature (RFC7515) encoding. The JSON web signature encoding allows an inline signature on the object. Other formattings could be allowed in future.

- **`objects`** *array of objects*

  An OPTIONAL list of objects that make up this manifest. These will potentially include multiple objects of the same or differetn versions that the client will pick. Blob and reference links are extracted from this list for the server housekeeping activity. Note we could use a map rather than a list here, with keys corresponding to object `type`.

The `objects` each correspond to items that the client may wish to process, and are defined as follows.

- **`type`** *string*

  A REQUIRED type for this object. Official allocations will be registered with OCI, such as for container images (eg `org.oci.image`), and other allocations will be recommended to be registered as standard. These should be reverse DNS allocations to avoid conflicts.

- **`version`** *string*

  A REQUIRED version string for the type. Clients will normally look for an object with the latest version they understand, and fall back to older versions if they support them. Different versions may have different structures.

- **`filters`** *map string - string*

  An array of OPTIONAL keys and values for selecting among multiple object options that a client has. These are filters that the client will use to select which version of an image is most appropriate, for example selecting by architecture ro any other appropriate choice. Common keys and values should be used where possible, such as `org.oci.architecture` so that common code and documentation can be reused. These are collected into a single map so a user interface can display them without understanding the object type.

- **`components`** *array of components*

  An array of OPTIONAL components of the object, such as layers and config for an image. See below for details, this is parsed to extract the `blob` and `reference` links for gc.

- **`annotations`** *string-string map*

  Standard OCI annotations as per existing standards, this is metadata that may have common fields across different objects. Arguably this should be a map of string to string array so that annotation keys can be repeated, so they map better to other structures (such as http headers) that allow repetition.

- **any other data** *unspecified*

  Any other fields with any other structure. The client can use this data in any way that is appropriate, but generic tooling will not usually be able to do anything with it as there are no standards.

The `components` of an object correspond to individual parts, potentially of multiple types, for example an image has a config and some layers. The parts that are needed to extract generic data are specified but again any type specific data can be added.

- **`rtype`** *string*

  The reference type of the component, which MUST be `blob`, `reference` or not specified if the component has neither a blob or reference link. This would be for a case where it is a component that does not have a descriptor, eg it refers to something by tag not hash so is not tracked as a reference. I am also wondering if we should add `data` for a "data URI" type reference where there would be a blob but it is inlined into the object instead, useful for small unique objects to save an external blob lookup, but the client can treat them exactly like blobs.

- **`descriptor`** *descriptor*

  Standard OCI descriptor. If the component is not a `blob` or `reference` this will be ignored but the client can use it, eg for a foreign layer or for some other reference type that still wants to verify a hash, such as a bridge to another system.

- **`ctype`** *string*

  The OPTIONAL component type, for an image this might be `org.oci.layer` or `org.oci.config`. These benefit from standardisation, as lots of artifacts use configs.

- **any other data** *unspecified*

  Any other fields with any other structure. The client can use this data in any way that is appropriate, but generic tooling will not usually be able to do anything with it as there are no standards.

Now let us look at how we map various existing and proposed types to this manifest format.

The simplest type is the Pointer that simply points to another object. This looks useless, but if it is signed it can act as a detached signature for the item that it points at.
```
{
	"schemaVersion": "1",
	"mediaType": "application/vnd.oci.object.manifest.v1+json",
	"objects": [
	{
		"type": "org.oci.pointer",
		"version": "1",
		"components": [
		{
			"rtype": "reference",
			"descriptor": {
  				"mediaType": "application/vnd.oci.image.manifest.v1+json",
  				"size": 7682,
 				"digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270"
			}
		}
		]
	}
	]
}
```

Next we will show a Relation, used to store an SBOM for example. This has a custom field `relation` for the relation type, which has some values registered with OCI as part of the format. A client that does not care about the SBOM will simply go to the first component in a relation without decoding it or fetching the SBOM. You could also store a detached signature as a relation, or indeed any other metadata about an item, which is why they are provided as a standard object type.
```
{
	"schemaVersion": "1",
	"mediaType": "application/vnd.oci.object.manifest.v1+json",
	"objects": [
	{
		"type": "org.oci.relation",
		"version": "1",
		"relation": "org.example.sbom",
		"components": [
		{
			"rtype": "reference",
			"descriptor": {
  				"mediaType": "application/vnd.oci.image.manifest.v1+json",
  				"size": 7682,
 				"digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270"
			}
		},
		{
			"rtype": "reference",
			"descriptor": {
  				"mediaType": "application/vnd.oci.object.manifest.v1+json",
  				"size": 1682,
 				"digest": "sha256:3b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270"
			}
		}
		]
	}
	]
}
```

Now let us look at how this would map the following image manifest
```
{
  "schemaVersion": 2,
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "size": 7023,
    "digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
  },
  "layers": [
    {
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      "size": 32654,
      "digest": "sha256:9834876dcfb05cb167a5c24953eba58c4ac89b1adf57f28f2f9d09af107ee8f0"
    },
    {
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      "size": 16724,
      "digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
    },
    {
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      "size": 73109,
      "digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736"
    }
  ],
  "annotations": {
    "com.example.key1": "value1",
    "com.example.key2": "value2"
  }
}
```
Note that we should get clients to support this generic manifest for images too so we can transition and support upgradeability.
```
{
	"schemaVersion": "1",
	"mediaType": "application/vnd.oci.object.manifest.v1+json",
	"objects": [
	{
		"type": "org.oci.image",
		"version": "2",
		"components": [
		{
			"rtype": "blob",
			"ctype": "org.oci.config",
			"descriptor": {
			   "mediaType": "application/vnd.oci.image.config.v1+json",
    			"size": 7023,
    			"digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
  			},
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 32654,
      			"digest": "sha256:9834876dcfb05cb167a5c24953eba58c4ac89b1adf57f28f2f9d09af107ee8f0"
  			}
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 16724,
      			"digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
  			}
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 73109,
      			"digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736"
    		}
		},
		]
	}
	],
	"annotations": {
    	"com.example.key1": "value1",
    	"com.example.key2": "value2"
  }
}
```
If we want to add an SBOM to this image we can add a Property, like a pointer but pointing at the object itself.
```
{
	"schemaVersion": "1",
	"mediaType": "application/vnd.oci.object.manifest.v1+json",
	"objects": [
	{
		"type": "org.oci.image",
		"version": "2",
		"components": [
		{
			"rtype": "blob",
			"ctype": "org.oci.config",
			"descriptor": {
			   "mediaType": "application/vnd.oci.image.config.v1+json",
    			"size": 7023,
    			"digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
  			},
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 32654,
      			"digest": "sha256:9834876dcfb05cb167a5c24953eba58c4ac89b1adf57f28f2f9d09af107ee8f0"
  			}
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 16724,
      			"digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
  			}
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 73109,
      			"digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736"
    		}
		},
		]
	},
	{
		"type": "org.oci.property",
		"version": "1",
		"components": [
		{
			"rtype": "reference",
			"descriptor": {
  				"mediaType": "application/vnd.oci.object.manifest.v1+json",
  				"size": 1682,
 				"digest": "sha256:3b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270"
			}
		}
		],
		"relation": "org.example.sbom"
	}
	],
	"annotations": {
    	"com.example.key1": "value1",
    	"com.example.key2": "value2"
  }
}
```
To make a multiarch image directly, without indirecting via images, we can use filters.
```
{
	"schemaVersion": "1",
	"mediaType": "application/vnd.oci.object.manifest.v1+json",
	"objects": [
	{
		"type": "org.oci.image",
		"version": "2",
		"filter": {
			"org.oci.os": "linux",
			"org.oci.architecture": "amd64"
		}
		"components": [
		{
			"rtype": "blob",
			"ctype": "org.oci.config",
			"descriptor": {
			   "mediaType": "application/vnd.oci.image.config.v1+json",
    			"size": 7023,
    			"digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
  			},
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 32654,
      			"digest": "sha256:9834876dcfb05cb167a5c24953eba58c4ac89b1adf57f28f2f9d09af107ee8f0"
  			}
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 16724,
      			"digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
  			}
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 73109,
      			"digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736"
    		}
		},
		]
	},
		{
		"type": "org.oci.image",
		"version": "2",
		"filter": {
			"org.oci.os": "linux",
			"org.oci.architecture": "ppc64"
		}
		"components": [
		{
			"rtype": "blob",
			"ctype": "org.oci.config",
			"descriptor": {
			   "mediaType": "application/vnd.oci.image.config.v1+json",
    			"size": 7023,
    			"digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
  			},
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 32654,
      			"digest": "sha256:9834876dcfb05cb167a5c24953eba58c4ac89b1adf57f28f2f9d09af107ee8f0"
  			}
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 16724,
      			"digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
  			}
		},
		{
			"rtype": "blob",
			"ctype": "org.oci.layer",
			"descriptor": {
    			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      			"size": 73109,
      			"digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736"
    		}
		},
		]
	}
	],
	"annotations": {
    	"com.example.key1": "value1",
    	"com.example.key2": "value2"
  }
}
```
This assumes that the annotations are global for the manifest we could move them to the component level. For the SBOM, if we were to attach one as a Property here it would apply to everything, so would be assuming it was a multi arch SBOM say. If we want to split things up with different SBOM per architecture, we would be better off splitting the images into different manifests, although we could devise something like Property that only applies to a single component here, for example we could add SBOM support as well as configs and layers in the image format, as it is just another reference type. This is the kind of flexibility this architecture gives us.
