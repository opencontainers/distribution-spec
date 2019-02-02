# Pulling an Image

Pulling an image from a registry is a multi-step process.
To pull an image from a registry, the client MUST know the following information beforehand:
- The hostname of the registry.
- The name of the image's repository.
- A tag name or digest that specifies a specific version of the image.

## Step 1: Fetch the manifest

When fetching a manifest from a registry, there are two manifest types that clients MUST be aware of:

### Standard Manifests

The most widely used manifest type is the standard manifest.
The standard manifest contains meta-data that describes where to download content for a specific image.

### Index of Manifests

The second manifest type is an index of manifests.
An index allows multiple images to be referenced by the same tag.

The index contains meta-data that describes a collection of standard manifests.
A common use case for indices is to store binaries built for different operating systems.

If a registry returns an index, clients MUST use the index to fetch a standard manifest by its digest.

## Step 2: Fetch blobs referenced in the manifest.

Once a standard manifest is fetched, clients MUST fetch all of the content referenced within.

Image download is complete. No further steps required.
