package main

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	confGoTag      = "conformance"
	envOCIConf     = "OCI"
	envOCIConfFile = "OCI_CONFIGURATION"
	envOCIVersion  = "OCI_VERSION"
	defaultOCIConf = "oci-conformance.yaml"
	chunkMin       = 1024
	truncateBody   = 4096
	biVCSCommit    = "vcs.revision"
)

type config struct {
	Registry   string       `conformance:"REGISTRY" yaml:"registry"`      // hostname:port of registry server
	TLS        tls          `conformance:"TLS" yaml:"tls"`                // tls configuration for communicating with the registry
	Repo1      string       `conformance:"REPO1" yaml:"repo1"`            // first repository for pushing content
	Repo2      string       `conformance:"REPO2" yaml:"repo2"`            // second repository for pushing content
	LoginUser  string       `conformance:"USERNAME" yaml:"username"`      // username for login, leave blank for anonymous
	LoginPass  string       `conformance:"PASSWORD" yaml:"password"`      // password for login, leave blank for anonymous
	LogLevel   string       `conformance:"LOG" yaml:"logging"`            // slog logging level, defaults to "warn"
	LogWriter  io.Writer    `yaml:"-"`                                    // writer used for logging, defaults to os.Stderr
	APIs       configAPI    `conformance:"API" yaml:"apis"`               // API tests to run
	Data       configData   `conformance:"DATA" yaml:"data"`              // data types to test
	ROData     configROData `conformance:"RO_DATA" yaml:"roData"`         // read-only data for registries that do not support push methods
	ResultsDir string       `conformance:"RESULTS_DIR" yaml:"resultsDir"` // directory to write results
	Version    string       `conformance:"VERSION" yaml:"version"`        // spec version used to set test defaults
	schemeReg  string       `yaml:"-"`                                    // base for url to access the registry
	Commit     string       `yaml:"commit"`                               // injected git commit hash from runtime
}

type tls int

const (
	tlsEnabled tls = iota
	tlsInsecure
	tlsDisabled
)

type configAPI struct {
	Pull      bool            `conformance:"PULL" yaml:"pull"`
	Push      bool            `conformance:"PUSH" yaml:"push"`
	Blobs     configBlobs     `conformance:"BLOBS" yaml:"blobs"`
	Manifests configManifests `conformance:"MANIFESTS" yaml:"manifests"`
	Tags      configTags      `conformance:"TAG" yaml:"tags"`
	Referrer  bool            `conformance:"REFERRER" yaml:"referrer"`
}

type configBlobs struct {
	Atomic         bool `config:"ATOMIC" yaml:"atomic"`
	Delete         bool `config:"DELETE" yaml:"delete"`
	MountAnonymous bool `config:"MOUNT_ANONYMOUS" yaml:"mountAnonymous"`
}

type configManifests struct {
	Atomic bool `config:"ATOMIC" yaml:"atomic"`
	Delete bool `config:"DELETE" yaml:"delete"`
}

type configTags struct {
	Atomic bool `config:"ATOMIC" yaml:"atomic"`
	Delete bool `config:"DELETE" yaml:"delete"`
	List   bool `config:"LIST" yaml:"list"`
}

type configData struct {
	Image            bool `conformance:"IMAGE" yaml:"image"`                       // standard OCI image
	Index            bool `conformance:"INDEX" yaml:"index"`                       // multi-platform manifest
	IndexList        bool `conformance:"INDEX_LIST" yaml:"indexList"`              // nested index
	Sparse           bool `conformance:"SPARSE" yaml:"sparse"`                     // TODO: multi-platform manifest with missing entries
	Artifact         bool `conformance:"ARTIFACT" yaml:"artifact"`                 // OCI artifact
	Subject          bool `conformance:"SUBJECT" yaml:"subject"`                   // artifact with the subject defined
	SubjectMissing   bool `conformance:"SUBJECT_MISSING" yaml:"subjectMissing"`    // artifact with a missing subject
	ArtifactList     bool `conformance:"ARTIFACT_LIST" yaml:"artifactList"`        // index of artifacts
	SubjectList      bool `conformance:"SUBJECT_LIST" yaml:"subjectList"`          // index with a subject
	DataField        bool `conformance:"DATA_FIELD" yaml:"dataField"`              // data field in descriptor
	Nondistributable bool `conformance:"NONDISTRIBUTABLE" yaml:"nondistributable"` // nondistributable image, deprecated in spec 1.1
	CustomFields     bool `conformance:"CUSTOM_FIELDS" yaml:"customFields"`        // fields added beyond the OCI spec
	Sha512           bool `conformance:"SHA512" yaml:"sha512"`                     // sha512 digest algorithm
}

type configROData struct {
	Tags      []string `conformance:"TAGS" yaml:"tags"`           // tag names
	Manifests []string `conformance:"MANIFESTS" yaml:"manifests"` // manifest digests
	Blobs     []string `conformance:"BLOBS" yaml:"blobs"`         // blob digests
	Referrers []string `conformance:"REFERRERS" yaml:"referrers"` // referrers subject digests
}

func configLoad() (config, error) {
	// read config from yaml file if available
	loadFile := ""
	configFile := []byte{}
	if filename, ok := os.LookupEnv(envOCIConfFile); ok {
		loadFile = filename
	} else if fi, err := os.Stat(defaultOCIConf); err == nil && !fi.IsDir() {
		loadFile = defaultOCIConf
	}
	if loadFile != "" {
		fh, err := os.Open(loadFile)
		if err != nil {
			return config{}, err
		}
		configFile, err = io.ReadAll(fh)
		_ = fh.Close()
		if err != nil {
			return config{}, err
		}
	}
	// extract the version from the config file or env variable
	configVersion := ""
	if len(configFile) > 0 {
		verStruct := struct {
			Version string `yaml:"version"`
		}{}
		err := yaml.Unmarshal(configFile, &verStruct)
		if err != nil {
			return config{}, err
		}
		configVersion = verStruct.Version
	}
	configVersionEnv := os.Getenv(envOCIVersion)
	if configVersionEnv != "" {
		configVersion = configVersionEnv
	}
	// initialize config with default values based on spec version
	c := config{
		Registry:   "localhost:5000",
		Repo1:      "conformance/repo1",
		Repo2:      "conformance/repo2",
		LogLevel:   "warn",
		LogWriter:  os.Stderr,
		ResultsDir: "./results",
		APIs: configAPI{
			Pull: true,
			Push: true,
			Blobs: configBlobs{
				Atomic:         true,
				Delete:         true,
				MountAnonymous: true,
			},
			Manifests: configManifests{
				Atomic: true,
				Delete: true,
			},
			Tags: configTags{
				Atomic: true,
				Delete: true,
				List:   true,
			},
			Referrer: true,
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
			DataField:        true,
			Nondistributable: true,
			CustomFields:     true,
			Sha512:           true,
		},
	}
	switch configVersion {
	case "", "1.1":
		c.Version = "1.1"
	case "1.0":
		c.APIs.Blobs.MountAnonymous = false
		c.APIs.Referrer = false
		c.Version = "1.0"
	default:
		return config{}, fmt.Errorf("unsupported config version %s", configVersion)
	}
	// process legacy variables but warn user when they are seen
	err := confLegacyEnv(&c)
	if err != nil {
		return c, err
	}
	// read config from yaml file if available
	if len(configFile) > 0 {
		err := yaml.Unmarshal(configFile, &c)
		if err != nil {
			return c, err
		}
	}
	// parse config from environment variables, overriding any yaml settings
	err = confFromEnv(envOCIConf, confGoTag, reflect.ValueOf(&c))
	if err != nil {
		return c, err
	}
	// setup computed values
	scheme := "https"
	if c.TLS == tlsDisabled {
		scheme = "http"
	}
	c.schemeReg = fmt.Sprintf("%s://%s", scheme, c.Registry)
	// load the commit from the build info
	if bi, ok := debug.ReadBuildInfo(); ok && bi != nil {
		for _, setting := range bi.Settings {
			if setting.Key == biVCSCommit {
				c.Commit = setting.Value
				break
			}
		}
	}
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

	// fall back to extracting by the kind
	switch v.Kind() {
	case reflect.String:
		v.SetString(val)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("failed to parse bool value from environment %s=%s", env, val)
		}
		v.SetBool(b)
	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.String:
			valSlice := strings.Split(val, " ")
			newSlice := reflect.MakeSlice(v.Type(), len(valSlice), len(valSlice))
			for i, cur := range valSlice {
				newSlice.Index(i).SetString(cur)
			}
			v.Set(newSlice)
		default:
			return fmt.Errorf("unsupported slice of kind: %s", v.Type().Elem().Kind())
		}
	default:
		// unhandled type
		return fmt.Errorf("unsupported kind: %s", v.Kind())
	}
	return nil
}

func confLegacyEnv(c *config) error {
	// Note: some legacy variables are not converted:
	// export OCI_HIDE_SKIPPED_WORKFLOWS=0
	// export OCI_DELETE_MANIFEST_BEFORE_BLOBS=0
	if v := os.Getenv("OCI_ROOT_URL"); v != "" {
		fmt.Fprintf(os.Stderr, "WARNING: OCI_REPO_URL is deprecated, use OCI_REGISTRY and OCI_TLS instead\n")
		v := strings.TrimSuffix(strings.TrimSpace(strings.ToLower(v)), "/")
		vSplit := strings.SplitN(v, "://", 2)
		scheme := "https"
		reg := v
		if len(vSplit) == 2 {
			scheme = vSplit[0]
			reg = vSplit[1]
		}
		switch scheme {
		case "http":
			c.TLS = tlsDisabled
		default:
			c.TLS = tlsEnabled
		}
		c.Registry = reg
	}
	if v := os.Getenv("OCI_NAMESPACE"); v != "" {
		fmt.Fprintf(os.Stderr, "WARNING: OCI_NAMESPACE is deprecated, use OCI_REPO1 instead\n")
		c.Repo1 = strings.TrimSuffix(strings.TrimSpace(strings.ToLower(v)), "/")
	}
	if v := os.Getenv("OCI_CROSSMOUNT_NAMESPACE"); v != "" {
		fmt.Fprintf(os.Stderr, "WARNING: OCI_CROSSMOUNT_NAMESPACE is deprecated, use OCI_REPO2 instead\n")
		c.Repo2 = strings.TrimSuffix(strings.TrimSpace(strings.ToLower(v)), "/")
	}
	if v := os.Getenv("OCI_TEST_PULL"); v != "" {
		fmt.Fprintf(os.Stderr, "WARNING: OCI_PULL is deprecated, use OCI_API_PULL instead\n")
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("failed ot parse bool value from environment %s=%s", "OCI_TEST_PULL", v)
		}
		c.APIs.Pull = b
	}
	if v := os.Getenv("OCI_TEST_PUSH"); v != "" {
		fmt.Fprintf(os.Stderr, "WARNING: OCI_PUSH is deprecated, use OCI_API_PUSH instead\n")
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("failed ot parse bool value from environment %s=%s", "OCI_TEST_PUSH", v)
		}
		c.APIs.Push = b
	}
	if v := os.Getenv("OCI_TEST_CONTENT_DISCOVERY"); v != "" {
		fmt.Fprintf(os.Stderr, "WARNING: OCI_CONTENT_DISCOVERY is deprecated, use OCI_API_TAG_LIST and OCI_API_REFERRER instead\n")
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("failed ot parse bool value from environment %s=%s", "OCI_TEST_CONTENT_DISCOVERY", v)
		}
		c.APIs.Tags.List = b
		c.APIs.Referrer = b
	}
	if v := os.Getenv("OCI_TEST_CONTENT_MANAGEMENT"); v != "" {
		fmt.Fprintf(os.Stderr, "WARNING: OCI_CONTENT_MANAGEMENT is deprecated, use OCI_API_TAG_DELETE, OCI_API_MANIFEST_DELETE, and OCI_API_BLOB_DELETE instead\n")
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("failed ot parse bool value from environment %s=%s", "OCI_TEST_CONTENT_MANAGEMENT", v)
		}
		c.APIs.Tags.Delete = b
		c.APIs.Manifests.Delete = b
		c.APIs.Blobs.Delete = b
	}
	if v := os.Getenv("OCI_DEBUG"); v != "" {
		fmt.Fprintf(os.Stderr, "WARNING: OCI_DEBUG is deprecated, use OCI_LOG=debug instead\n")
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("failed ot parse bool value from environment %s=%s", "OCI_DEBUG", v)
		}
		if b {
			c.LogLevel = "debug"
		}
	}
	if v := os.Getenv("OCI_TAG_NAME"); v != "" {
		fmt.Fprintf(os.Stderr, "WARNING: OCI_TAG_NAME is deprecated, use OCI_RO_DATA_TAGS instead\n")
		c.ROData.Tags = append(c.ROData.Tags, strings.Split(v, " ")...)
	}
	if v := os.Getenv("OCI_MANIFEST_DIGEST"); v != "" {
		fmt.Fprintf(os.Stderr, "WARNING: OCI_MANIFEST_DIGEST is deprecated, use OCI_RO_DATA_MANIFESTS instead\n")
		c.ROData.Manifests = append(c.ROData.Manifests, strings.Split(v, " ")...)
	}
	if v := os.Getenv("OCI_BLOB_DIGEST"); v != "" {
		fmt.Fprintf(os.Stderr, "WARNING: OCI_BLOB_DIGEST is deprecated, use OCI_RO_DATA_BLOBS instead\n")
		c.ROData.Blobs = append(c.ROData.Blobs, strings.Split(v, " ")...)
	}
	return nil
}

func (c config) Report() string {
	// censor credentials
	if c.LoginUser != "" {
		c.LoginUser = "***"
	}
	if c.LoginPass != "" {
		c.LoginPass = "***"
	}
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("failed to marshal config: %v", err)
	}
	return string(b)
}

var confHTMLTemplates = map[string]string{
	"report": `<html>
  <head>
    <title>OCI Distribution Conformance Tests</title>
    <style>
      body {
        padding: 10px 20px 10px 20px;
        font-family: -apple-system,BlinkMacSystemFont,Segoe UI,PingFang SC,Hiragino Sans GB,Microsoft YaHei,Helvetica Neue,Helvetica,Arial,sans-serif,Apple Color Emoji,Segoe UI Emoji,Segoe UI Symbol;
        background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAYAAACp8Z5+AAAAG0lEQVQYV2Pce7zwv7NlPyMDFMAZGAIwlRgqAFydCAVv5m4UAAAAAElFTkSuQmCC") repeat;
        /* background made with http://www.patternify.com/ */
      }
      table {
        border-collapse: collapse;
        width: 100%;
        background-color: white;
      }
      th, td {
        padding: 12px;
        text-align: left;
		  vertical-align: top;
        border-bottom: 1px solid #ddd;
      }
      tr:hover {
        background-color: #ffe39b;
      }
		p {
		  margin: .1em 0em .1em 0em;
		}
		summary {
		  cursor: pointer;
		}
      .result {
        padding: .25em 0 .25em .75em;
        border: 0px solid #e1e1e1;
        border-radius: 5px;
        margin-top: .1em;
      }
		.result h4 {
		  margin: .1em 0em .1em 0em;
      }
      .red {
        background: #ffc8c8;
      }
      pre.fail-message {
		  margin: .1em 0em .1em 0em;
        background: #f9a5a5;
        padding: 10px;
        margin-right: 10px;
        display: inline-block;
        border-radius: 4px;
        font-size: 1.25em;
        width: 94%;
        overflow-x: auto;
        max-width: 85%;
      }
      .green {
        background: #c8ffc8;
        /* padding: 1.25em 0 1.25em 0.8em; */
      }
      .grey {
        background: lightgrey;
        /* padding: 1.25em 0 1.25em 0.8em; */
      }
      pre.pre-box {
		  margin: .1em 0em .1em 0em;
        background: #343a40;
        color: #fff;
        padding: 10px;
        border: 1px solid gray;
        display: inline-block;
        border-radius: 4px;
        width: 97%;
        font-size: 1.25em;
        overflow-x: auto;
        max-height: 60em;
        overflow-y: auto;
        max-width: 85%;
      }
      .bullet-left {
        width: 25%;
        font-weight: bold;
        font-size: 100%;
      }
      .bullet-right {
        width: auto;
        font-family: monospace;
        font-size: 110%;
      }
      .quick-summary {
        width: 70%;
        display: flex;
        margin: 0 auto 0 0;
        font-weight: bold;
        font-size: 1.2em;
      }
      .darkgreen {
        color: green;
      }
      .darkred {
        color: red;
        padding: 0 0 0 2em;
      }
      .darkgrey {
        color: grey;
        padding: 0 0 0 2em;
      }
      .meter {
        border: 1px solid black;
        margin: 0 .5em 0 auto;
        display: flex;
        height: 25px;
        width: 45%;
      }
      @media only screen and (max-width: 600px) {
        .meter {
          display: none;
        }
      }
      .meter-green {
        height: 100%;
        background: green;
        width: {{ .PercentPassed -}}%;
      }
      .meter-red {
        height: 100%;
        background: red;
        width: {{ .PercentFailed -}}%;
      }
      .meter-grey {
        height: 100%;
        background: grey;
        width: {{ .PercentSkipped -}}%;
      }
    </style>
  </head>
  <body>
    <h1>OCI Distribution Conformance Tests</h1>
    {{ template "summary" . }}
    <div>
    {{ template "results" .Results }}
    </div>
  </body>
</html>`,
	"summary": `<table>
      <tr>
      </tr>
      <tr>
        <td class="bullet-left">Summary</td>
        <td>
          <div class="quick-summary">
            {{- if gt .NumPassed 0 -}}
              <span class="darkgreen">
              {{- if .AllPassed -}}All {{ end -}}{{ .NumPassed }} passed</span>
            {{- end -}}
            {{- if gt .NumFailed 0 -}}
              <span class="darkred">
              {{- if .AllFailed -}}All {{ end -}}{{ .NumFailed }} failed</span>
            {{- end -}}
            {{- if gt .NumSkipped 0 -}}
              <span class="darkgrey">
              {{- if .AllSkipped -}}All {{ end -}}{{ .NumSkipped }} skipped</span>
            {{- end -}}
            <div class="meter">
              <div class="meter-green"></div>
              <div class="meter-red"></div>
              <div class="meter-grey"></div>
            </div>
          </div>
        </td>
      </tr>
      <tr>
        <td class="bullet-left">Start Time</td>
        <td>{{ .StartTimeString }}</td>
      </tr>
      <tr>
        <td class="bullet-left">End Time</td>
        <td>{{ .EndTimeString }}</td>
      </tr>
      <tr>
        <td class="bullet-left">Time Elapsed</td>
        <td>{{ .RunTime }}</td>
      </tr>
      <tr>
        <td class="bullet-left">Tested Spec</td>
        <td>{{ .Config.Version }}</td>
      </tr>
      <tr>
        <td class="bullet-left">Conformance Commit</td>
        <td>{{ .Config.Commit }}</td>
      </tr>
      <tr>
        <td class="bullet-left">Configuration</td>
        <td><pre>{{ .Config.Report }}</pre></td>
      </tr>
    </table>`,
	"results": `
    <div class="result {{ template "status-color" .Status }}">
    <details {{- if ne ( len .Children ) 0 }} open="true" {{- end }}><summary><h4 style="display: inline;">{{ .Name }}</h4></summary>
    {{- if ne .Output.String "" }}
    <p>Output:</p>
    <pre class="pre-box">{{- html .Output.String -}}</pre>
    {{- end }}
    {{- if ne ( len .Errs ) 0 }}
    <p>Errors:</p>
    {{- range $err := .Errs }}
    <pre class="fail-message">{{ html $err.Error }}</pre>
    {{- end }}
    {{- end }}
    {{- range $result := .Children }}
    {{template "results" $result }}
    {{- end }}
    </details>
    </div>
  `,
	"status-color": `
  {{- if eq .String "Pass" }}green
  {{- else if eq .String "FAIL" }}red
  {{- else if eq .String "Error" }}red
  {{- else if eq .String "Skip" }}grey
  {{- else if eq .String "Disabled" }}grey
  {{- end }}`,
}
