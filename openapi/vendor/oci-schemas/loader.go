package oci

import (
	"github.com/xeipuuv/gojsonschema"
)

var schemaFS = FS(false)

// ImageIndexLoader provides a JSONLoader that can be used to validate an image
// index.
var ImageIndexLoader = func() gojsonschema.JSONLoader {
	path := "file:///image-index.schema.json"
	return gojsonschema.NewReferenceLoaderFileSystem(path, schemaFS)
}

// ImageManifestLoader provides a JSONLoader that can be used to validate an
// image manifest.
var ImageManifestLoader = func() gojsonschema.JSONLoader {
	path := "file:///image-manifest.schema.json"
	return gojsonschema.NewReferenceLoaderFileSystem(path, schemaFS)
}

// ImageConfigLoader provides a JSONLoader that can be used to validate an image
// config.
var ImageConfigLoader = func() gojsonschema.JSONLoader {
	path := "file:///image-config.schema.json"
	return gojsonschema.NewReferenceLoaderFileSystem(path, schemaFS)
}

// TagListLoader provides a JSONLoader that can be used to validate an tag list.
var TagListLoader = func() gojsonschema.JSONLoader {
	path := "file:///tag-list.schema.json"
	return gojsonschema.NewReferenceLoaderFileSystem(path, schemaFS)
}

// CatalogLoader provides a JSONLoader that can be used to validate a catalog of
// repositories.
var CatalogLoader = func() gojsonschema.JSONLoader {
	path := "file:///catalog.schema.json"
	return gojsonschema.NewReferenceLoaderFileSystem(path, schemaFS)
}

// ErrorsLoader provides a JSONLoader that can be used to validate an
// OCI-compliant error response.
var ErrorsLoader = func() gojsonschema.JSONLoader {
	path := "file:///errors.schema.json"
	return gojsonschema.NewReferenceLoaderFileSystem(path, schemaFS)
}

// contentDescriptorLoader provides a JSONLoader that can be used to validate a
// content descriptor.
var contentDescriptorLoader = func() gojsonschema.JSONLoader {
	path := "file:///content-descriptor.schema.json"
	return gojsonschema.NewReferenceLoaderFileSystem(path, schemaFS)
}
