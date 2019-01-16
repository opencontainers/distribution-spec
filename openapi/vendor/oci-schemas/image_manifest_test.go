package oci

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

type imageManifestTestParams struct {
	file  string
	valid bool
}

func TestImageManifestValidation(t *testing.T) {
	tt := []imageManifestTestParams{
		{"golden.json", true},
		{"golden-minimal.json", true},
		{"schema-version-too-low.json", false},
		{"schema-version-too-high.json", false},
		{"schema-version-missing.json", false},
		{"config-media-type-invalid.json", false},
		{"layer-media-type-invalid.json", false},
		{"config-size-not-number.json", false},
		{"layer-size-not-number.json", false},
		{"layers-empty-array.json", false},
	}

	v := ImageManifestLoader
	for _, tc := range tt {
		tf := func(t *testing.T) {
			path := fmt.Sprintf("./test-fixtures/image-manifest/%v", tc.file)
			data, err := os.Open(path)
			if err != nil {
				t.Errorf("unable to open %v: %v", tc.file, err)
			}
			defer data.Close()

			b, err := ioutil.ReadAll(data)
			if err != nil {
				t.Errorf("unable to read data: %v", err)
			}

			loader := gojsonschema.NewBytesLoader(b)

			res, err := gojsonschema.Validate(v, loader)
			if err != nil {
				t.Errorf("unable to validate schema: %v", err)
			}

			if !res.Valid() && tc.valid {
				t.Errorf("expected valid schema, got invalid: %v", res.Errors())
			} else if res.Valid() && !tc.valid {
				t.Errorf("expected invalid schema, got valid")
			}
		}
		rmSuffix := strings.TrimSuffix(tc.file, ".json")
		name := strings.Replace(rmSuffix, "-", " ", -1)
		t.Run(name, tf)
	}
}
