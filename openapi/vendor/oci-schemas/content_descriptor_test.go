package oci

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

type descriptorTestParams struct {
	file  string
	valid bool
}

func TestContentDescriptorValidation(t *testing.T) {
	tt := []descriptorTestParams{
		{"golden.json", true},
		{"digest-sha256.json", true},
		{"digest-sha256-invalid.json", false},
		{"digest-sha512.json", true},
		{"digest-sha512-invalid.json", false},
	}

	v := contentDescriptorLoader
	for _, tc := range tt {
		tf := func(t *testing.T) {
			path := fmt.Sprintf("./test-fixtures/content-descriptor/%v", tc.file)
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
