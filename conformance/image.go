package conformance

import (
	digest "github.com/opencontainers/go-digest"
)

// These types are copied from github.com/opencontainers/image-spec/specs-go/v1
// Modifications have been made to remove fields that aren't used in these
// conformance tests, and to add new unspecified fields, to test registry
// conformance in handling unknown fields.

// manifest provides `application/vnd.oci.image.manifest.v1+json` mediatype structure when marshalled to JSON.
type manifest struct {
	// SchemaVersion is the image manifest schema that this image follows
	SchemaVersion int `json:"schemaVersion"`

	// MediaType specifies the type of this document data structure e.g. `application/vnd.oci.image.manifest.v1+json`
	MediaType string `json:"mediaType,omitempty"`

	// ArtifactType specifies the IANA media type of artifact when the manifest is used for an artifact.
	ArtifactType string `json:"artifactType,omitempty"`

	// Config references a configuration object for a container, by digest.
	// The referenced configuration object is a JSON blob that the runtime uses to set up the container.
	Config descriptor `json:"config"`

	// Layers is an indexed list of layers referenced by the manifest.
	Layers []descriptor `json:"layers"`

	// Subject is an optional link from the image manifest to another manifest forming an association between the image manifest and the other manifest.
	Subject *descriptor `json:"subject,omitempty"`
}

// descriptor describes the disposition of targeted content.
// This structure provides `application/vnd.oci.descriptor.v1+json` mediatype
// when marshalled to JSON.
type descriptor struct {
	// MediaType is the media type of the object this schema refers to.
	MediaType string `json:"mediaType"`

	// Digest is the digest of the targeted content.
	Digest digest.Digest `json:"digest"`

	// Size specifies the size in bytes of the blob.
	Size int64 `json:"size"`

	// Data specifies the data of the object described by the descriptor.
	Data []byte `json:"data,omitempty"`

	// NewUnspecifiedField is not covered by image-spec.
	// Registry implementations should still successfully store and serve
	// manifests containing this data.
	NewUnspecifiedField []byte `json:"newUnspecifiedField"`
}

// rootFS describes a layer content addresses
type rootFS struct {
	// Type is the type of the rootfs.
	Type string `json:"type"`

	// DiffIDs is an array of layer content hashes (DiffIDs), in order from bottom-most to top-most.
	DiffIDs []digest.Digest `json:"diff_ids"`
}

// image is the JSON structure which describes some basic information about the image.
// This provides the `application/vnd.oci.image.config.v1+json` mediatype when marshalled to JSON.
type image struct {
	// Author defines the name and/or email address of the person or entity which created and is responsible for maintaining the image.
	Author string `json:"author,omitempty"`

	// Architecture is the CPU architecture which the binaries in this image are built to run on.
	Architecture string `json:"architecture"`

	// Variant is the variant of the specified CPU architecture which image binaries are intended to run on.
	Variant string `json:"variant,omitempty"`

	// OS is the name of the operating system which the image is built to run on.
	OS string `json:"os"`

	// RootFS references the layer content addresses used by the image.
	RootFS rootFS `json:"rootfs"`
}

// index references manifests for various platforms.
// This structure provides `application/vnd.oci.image.index.v1+json` mediatype when marshalled to JSON.
type index struct {
	// SchemaVersion is the image manifest schema that this image follows
	SchemaVersion int `json:"schemaVersion"`

	// MediaType specifies the type of this document data structure e.g. `application/vnd.oci.image.index.v1+json`
	MediaType string `json:"mediaType,omitempty"`

	// Manifests references platform specific manifests.
	Manifests []descriptor `json:"manifests"`

	// Annotations contains arbitrary metadata for the image index.
	Annotations map[string]string `json:"annotations,omitempty"`
}
