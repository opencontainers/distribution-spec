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
	defaultOCIConf = "oci-conformance.yml"
	chunkMin       = 1024
	truncateBody   = 4096
	biVCSCommit    = "vcs.revision"
)

type config struct {
	Registry   string     `conformance:"REGISTRY" yaml:"registry"`      // hostname:port of registry server
	TLS        tls        `conformance:"TLS" yaml:"tls"`                // tls configuration for communicating with the registry
	Repo1      string     `conformance:"REPO1" yaml:"repo1"`            // first repository for pushing content
	Repo2      string     `conformance:"REPO2" yaml:"repo2"`            // second repository for pushing content
	LoginUser  string     `conformance:"USERNAME" yaml:"username"`      // username for login, leave blank for anonymous
	LoginPass  string     `conformance:"PASSWORD" yaml:"password"`      // password for login, leave blank for anonymous
	LogLevel   string     `conformance:"LOG" yaml:"logging"`            // slog logging level, defaults to "warn"
	LogWriter  io.Writer  `yaml:"-"`                                    // writer used for logging, defaults to os.Stderr
	APIs       configAPI  `conformance:"API" yaml:"apis"`               // API tests to run
	Data       configData `conformance:"DATA" yaml:"data"`              // data types to test
	ResultsDir string     `conformance:"RESULTS_DIR" yaml:"resultsDir"` // directory to write results
	Version    string     `conformance:"VERSION" yaml:"version"`        // spec version used to set test defaults
	schemeReg  string     `yaml:"-"`                                    // base for url to access the registry
	Commit     string     `yaml:"commit"`                               // injected git commit hash from runtime
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
	Sparse           bool `conformance:"SPARSE" yaml:"sparse"`                     // multi-platform manifest with missing entries
	Artifact         bool `conformance:"ARTIFACT" yaml:"artifact"`                 // OCI artifact
	Subject          bool `conformance:"SUBJECT" yaml:"subject"`                   // artifact with the subject defined
	SubjectMissing   bool `conformance:"SUBJECT_MISSING" yaml:"subjectMissing"`    // artifact with a missing subject
	ArtifactList     bool `conformance:"ARTIFACT_LIST" yaml:"artifactList"`        // index of artifacts
	SubjectList      bool `conformance:"SUBJECT_LIST" yaml:"subjectList"`          // index with a subject
	Nondistributable bool `conformance:"NONDISTRIBUTABLE" yaml:"nondistributable"` // nondistributable image, deprecated in spec 1.1
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
	}
	switch configVersion {
	case "", "1.1":
		c.APIs = configAPI{
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
		}
		c.Data = configData{
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
		}
		c.Version = "1.1"
	case "1.0":
		c.APIs = configAPI{
			Pull: true,
			Push: true,
			Blobs: configBlobs{
				Atomic:         true,
				Delete:         true,
				MountAnonymous: false,
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
			Referrer: false,
		}
		c.Data = configData{
			Image:            true,
			Index:            true,
			IndexList:        true,
			Sparse:           false,
			Artifact:         true,
			Subject:          true,
			SubjectMissing:   true,
			ArtifactList:     true,
			SubjectList:      true,
			Nondistributable: true,
		}
		c.Version = "1.0"
	default:
		return config{}, fmt.Errorf("unsupported config version %s", configVersion)
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
	if len(configFile) > 0 {
		err := yaml.Unmarshal(configFile, &c)
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
	default:
		// unhandled type
		return fmt.Errorf("unsupported kind: %s", v.Kind())
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
        border-bottom: 1px solid #ddd;
      }
      tr:hover {
        background-color: #ffe39b;
      }
      .result {
        padding: 1.25em 0 .25em 0.8em;
        border: 1px solid #e1e1e1;
        border-radius: 5px;
        margin-top: 10px;
      }
      .red {
        background: #ffc8c8;
      }
      pre.fail-message {
        background: #f9a5a5;
        padding: 20px;
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
        padding: 1.25em 0 1.25em 0.8em;
      }
      .grey {
        background: lightgrey;
        padding: 1.25em 0 1.25em 0.8em;
      }
      .toggle {
        border: 2px solid #3e3e3e;
        cursor: pointer;
        width: 1em;
        text-align: center;
        font-weight: bold;
        display: inline;
        font-family: monospace;
        padding: 0 .25em 0 .25em;
        margin: 1em 1em 1em 0;
        font-size: 12pt;
        color: #3e3e3e;
        border-radius: 3px;
      }
      pre.pre-box {
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
      .summary {
        width: 100%;
        height: auto;
        padding: 0 0 .5em 0;
        border-radius: 6px;
        border: 1px solid #cccddd;
        background: white;
      }
      .summary-bullet {
        width: 100%;
        height: auto;
        display: flex;
        flex-wrap: wrap;
        padding: .5em .1em .1em .5em;
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
      .subcategory {
        background: white;
        padding: 0px 20px 20px 20px;
        border: 1px solid #cccddd;
        border-radius: 6px;
      }
      h2 {
        margin-top: 45px;
      }
      h4 {
        vertical-align: bottom;
        cursor: pointer;
      }
    </style>
    <script>
      function toggleOutput(id) {
        var elem = document.getElementById(id);
        var button = document.getElementById(id + "-button");
        if (elem.style['display'] === 'block') {
          button.innerHTML = "+";
          elem.style['display'] = 'none';
        } else {
          button.innerHTML = "-";
          elem.style['display'] = 'block';
        }
      }
    </script>
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
    <h4 style="display: inline;">{{ .Name }}</h4>
    {{- if ne .Output.String "" }}
    <details><summary>Output</summary>
    <pre class="pre-box">{{- html .Output.String -}}</pre>
    </details>
    {{- end }}
    {{- if ne ( len .Errs ) 0 }}
    <details><summary>Errors</summary>
    {{- range $err := .Errs }}
    <pre class="fail-message">{{ html $err.Error }}</pre>
    {{- end }}
    </details>
    {{- end }}
    {{- range $result := .Children }}
    {{template "results" $result }}
    {{- end }}
    </div>
  `,
	"status-color": `
  {{- if eq .String "Pass" }}green
  {{- else if eq .String "Fail" }}red
  {{- else if eq .String "Error" }}red
  {{- else if eq .String "Skip" }}grey
  {{- else if eq .String "Disabled" }}grey
  {{- end }}`,
}
