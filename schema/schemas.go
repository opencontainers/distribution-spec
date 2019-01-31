package schema

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

var schemaFS = _escFS(false)

func loadSchema(file string) *gojsonschema.Schema {
	path := fmt.Sprintf("file:///%s", file)
	loader := gojsonschema.NewReferenceLoaderFileSystem(path, schemaFS)
	schema, _ := gojsonschema.NewSchema(loader)
	return schema
}

// CatalogSchema provides a gojsonschema.Schema that can be used to validate
// a catalog of repositories.
func CatalogSchema() *gojsonschema.Schema {
	return loadSchema("catalog.schema.json")
}

// ErrorSchema provides a gojsonschema.Schema that can be used to validate an
// error returned from a registry.
func ErrorSchema() *gojsonschema.Schema {
	return loadSchema("errors.schema.json")
}

// TagsListSchema provides a gojsonschem.Schema that can be used to validate
// a list of tags returned from a registry.
func TagsListSchema() *gojsonschema.Schema {
	return loadSchema("tags-list.schema.json")
}
