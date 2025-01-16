package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	confGoTag      = "conformance"
	envOCIConf     = "OCI"
	envOCIConfFile = "OCI_CONFIGURATION"
	defaultOCIConf = "oci-conformance.yml"
)

type config struct {
	Registry  string     `conformance:"REGISTRY" yaml:"registry"` // hostname:port of registry server
	TLS       tls        `conformance:"TLS" yaml:"tls"`           // tls configuration for communicating with the registry
	Repo1     string     `conformance:"REPO1" yaml:"repo1"`       // first repository for pushing content
	Repo2     string     `conformance:"REPO2" yaml:"repo2"`       // second repository for pushing content
	LoginUser string     `conformance:"USERNAME" yaml:"username"` // username for login, leave blank for anonymous
	LoginPass string     `conformance:"PASSWORD" yaml:"password"` // password for login, leave blank for anonymous
	LogLevel  string     `conformance:"LOG" yaml:"logging"`       // TODO: logging level, use slog levels
	APIs      configAPI  `conformance:"API" yaml:"apis"`          // API tests to run
	Data      configData `conformance:"DATA" yaml:"data"`
	schemeReg string     // base for url to access the registry
}

type tls int

const (
	tlsEnabled tls = iota
	tlsInsecure
	tlsDisabled
)

type configAPI struct {
	Pull     bool          `conformance:"PULL" yaml:"pull"`
	Push     bool          `conformance:"PUSH" yaml:"push"`
	Tags     bool          `conformance:"TAG_LIST" yaml:"tags"`
	Referrer bool          `conformance:"REFERRER" yaml:"referrer"`
	Delete   configObjects `conformance:"DELETE" yaml:"delete"`
}

type configObjects struct {
	Tag      bool `conformance:"TAG" yaml:"tag"`
	Manifest bool `conformance:"MANIFEST" yaml:"manifest"`
	Blob     bool `conformance:"BLOB" yaml:"blob"`
}

type configData struct {
	Image            bool `conformance:"IMAGE" yaml:"image"`                       // standard OCI image
	Index            bool `conformance:"INDEX" yaml:"index"`                       // multi-platform manifest
	IndexList        bool `conformance:"INDEX_LIST" yaml:"indexList"`              // nested index
	Sparse           bool `conformance:"SPARSE" yaml:"sparse"`                     // multi-platform manifest with missing entries
	Artifact         bool `conformance:"ARTIFACT" yaml:"artifact"`                 // OCI artifact
	Subject          bool `conformance:"SUBJECT" yaml:"subject"`                   // artifact with the subject defined
	SubjectMissing   bool `conformance:"SUBJECT_MISSING" yaml:"subjectMissing"`    // artifact with a missing subject
	ArtifactList     bool `conformance:"ARTIFACT_LIST" yaml:"artifactList"`        // index of artifacts
	SubjectList      bool `conformance:"SUBJECT_LIST" yaml:"subjectList"`          // index with a subject
	Nondistributable bool `conformance:"NONDISTRIBUTABLE" yaml:"nondistributable"` // nondistributable image, deprecated
}

func configLoad() (config, error) {
	// initialize config with default values
	c := config{
		Registry: "localhost:5000",
		Repo1:    "conformance/repo1",
		Repo2:    "conformance/repo2",
		APIs: configAPI{
			Pull:     true,
			Push:     true,
			Tags:     true,
			Referrer: true,
			Delete: configObjects{
				Tag:      true,
				Manifest: true,
				Blob:     true,
			},
		},
		Data: configData{
			Image:            true,
			Index:            true,
			IndexList:        true,
			Sparse:           false,
			Artifact:         true,
			Subject:          true,
			SubjectMissing:   true,
			ArtifactList:     true,
			SubjectList:      true,
			Nondistributable: false,
		},
	}

	// TODO:
	// read legacy environment variables to set defaults, but warn if seen
	// export OCI_ROOT_URL="https://r.myreg.io"
	// export OCI_NAMESPACE="myorg/myrepo"
	// export OCI_CROSSMOUNT_NAMESPACE="myorg/other"
	// export OCI_TEST_PULL=1
	// export OCI_TEST_PUSH=1
	// export OCI_TEST_CONTENT_DISCOVERY=1
	// export OCI_TEST_CONTENT_MANAGEMENT=1
	// export OCI_HIDE_SKIPPED_WORKFLOWS=0
	// export OCI_DEBUG=0
	// export OCI_DELETE_MANIFEST_BEFORE_BLOBS=0 // exclude option until a requirement is found

	// read config from yaml file if available
	loadFile := ""
	if filename, ok := os.LookupEnv(envOCIConfFile); ok {
		loadFile = filename
	} else if fi, err := os.Stat(defaultOCIConf); err == nil && !fi.IsDir() {
		loadFile = defaultOCIConf
	}
	if loadFile != "" {
		fh, err := os.Open(loadFile)
		if err != nil {
			return c, err
		}
		err = yaml.NewDecoder(fh).Decode(&c)
		_ = fh.Close()
		if err != nil {
			return c, err
		}
	}

	// for each config option, check if env var is set to override value
	err := confFromEnv(envOCIConf, confGoTag, reflect.ValueOf(&c))
	if err != nil {
		return c, err
	}

	// setup computed values
	scheme := "https"
	if c.TLS == tlsDisabled {
		scheme = "http"
	}
	c.schemeReg = fmt.Sprintf("%s://%s", scheme, c.Registry)

	return c, nil
}

func (t tls) MarshalText() ([]byte, error) {
	var s string
	switch t {
	default:
		s = "enabled" // by default, TLS is enabled
	case tlsInsecure:
		s = "insecure"
	case tlsDisabled:
		s = "disabled"
	}
	return []byte(s), nil
}

func (t *tls) UnmarshalText(b []byte) error {
	switch strings.ToLower(string(b)) {
	default:
		*t = tlsEnabled
	case "insecure":
		*t = tlsInsecure
	case "disabled":
		*t = tlsDisabled
	}
	return nil
}

func confFromEnv(env, tag string, vp reflect.Value) error {
	vpt := vp.Type()
	if vpt.Kind() != reflect.Pointer {
		return fmt.Errorf("confFromEnv requires a pointer input")
	}
	if vp.IsZero() {
		return nil // nil pointer
	}
	v := reflect.Indirect(vp)
	if v.Kind() == reflect.Pointer {
		// pointer to a pointer, recurse
		return confFromEnv(env, tag, v)
	}
	if v.Kind() == reflect.Struct {
		// expand each field, adding to prefix and recursing on pointer to the entry
		for i := 0; i < v.NumField(); i++ {
			vtf := v.Type().Field(i)
			tagVal := vtf.Tag.Get(tag)
			if tagVal != "" {
				if !v.Field(i).CanAddr() {
					return fmt.Errorf("unable to generate address on %s", v.Field(i).Type().Name())
				}
				tagEnv := fmt.Sprintf("%s_%s", env, tagVal)
				err := confFromEnv(tagEnv, tag, v.Field(i).Addr())
				if err != nil {
					return fmt.Errorf("field failed \"%s\": %w", v.Field(i).Type().Name(), err)
				}
			}
		}
		return nil
	}

	// get the value from the environment
	val := os.Getenv(env)
	if val == "" {
		// skip undefined env variables
		return nil
	}

	// try to unmarshal with a built in method
	if mt, ok := vp.Interface().(interface{ UnmarshalText(b []byte) error }); ok {
		err := mt.UnmarshalText([]byte(val))
		if err != nil {
			return fmt.Errorf("failed to unmarshal \"%s\": %w", env, err)
		}
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(val)

	// TODO:
	// case reflect.Int:
	// case reflect.Bool:

	default:
		// unhandled type
		return fmt.Errorf("unsupported kind: %s", v.Kind())
	}
	return nil
}
