package oci

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

type imageIndexTestParams struct {
	file  string
	valid bool
}

// platformComboTestTable creates a table of all possible OS/Arch combinations.
var platformComboTestTable = func() []imageIndexTestParams {
	osList := []string{
		"android",
		"darwin",
		"dragonfly",
		"freebsd",
		"linux",
		"netbsd",
		"openbsd",
		"plan9",
		"solaris",
		"windows",
	}

	archList := []string{
		"386",
		"amd64",
		"arm",
		"arm64",
		"ppc64",
		"ppc64le",
		"mips64",
		"mips64le",
		"s390x",
	}

	validCombinations := map[string][]string{
		"android":   {"arm"},
		"darwin":    {"386", "amd64", "arm", "arm64"},
		"dragonfly": {"amd64"},
		"freebsd":   {"386", "amd64", "arm"},
		"linux":     {"386", "amd64", "arm", "arm64", "ppc64", "ppc64le", "mips64", "mips64le", "s390x"},
		"netbsd":    {"386", "amd64", "arm"},
		"openbsd":   {"386", "amd64", "arm"},
		"plan9":     {"386", "amd64"},
		"solaris":   {"amd64"},
		"windows":   {"386", "amd64"},
	}

	tt := []imageIndexTestParams{}
	for _, os := range osList {
		for _, arch := range archList {
			file := fmt.Sprintf("platform-%s-%s.json", os, arch)
			valid := func() bool {
				vc := validCombinations[os]
				for _, ok := range vc {
					if ok == arch {
						return true
					}
				}
				return false
			}()
			tt = append(tt, imageIndexTestParams{file: file, valid: valid})
		}
	}
	return tt

}()

func TestImageIndexValidation(t *testing.T) {
	tt := []imageIndexTestParams{
		{"golden.json", true},
		{"golden-minimal.json", true},
		{"custom-manifest-media-type.json", true},
		{"manifest-not-array.json", false},
		{"platform-combination-invalid.json", false},
		{"platform-os-invalid.json", false},
		{"platform-arch-invalid.json", false},
		{"schema-version-missing.json", false},
		{"schema-version-too-high.json", false},
		{"schema-version-too-low.json", false},
	}
	tt = append(tt, platformComboTestTable...)

	v := ImageIndexLoader
	for _, tc := range tt {
		tf := func(t *testing.T) {
			path := fmt.Sprintf("./test-fixtures/image-index/%v", tc.file)
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
