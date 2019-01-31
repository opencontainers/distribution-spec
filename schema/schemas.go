//   Copyright 2018 The Linux Foundation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

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
