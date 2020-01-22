package conformance

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

const (
	htmlTemplate string = `<html>
  <head>
    <title>OCI Distribution Conformance Tests</title>
    <style>
	  body {
	    padding: 30px;
	  }
      .result {
        padding: 1.25em 0 .25em 2em;
        border: 2px solid #e2e2e2;
        border-radius: 5px;
        margin-top: 10px;
      }
      .red {
        background: #ffc8c8;
      }
      pre.fail-message {
        background: #f9a5a5;
        padding: 20px;
        display: inline-block;
        border-radius: 4px;
        font-size: 1.25em;
      }
      .green {
        background: #c8ffc8;
        padding: 1.25em 0 1.25em 2em;
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
        background: white;
        padding: 10px;
        border: 1px solid gray;
        display: inline-block;
        border-radius: 4px;
        width: 97%;
        font-size: 1.25em;
        overflow-x: auto;
        max-height: 60em;
        overflow-y: auto;
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
    <div>
      {{with .SpecSummaryMap}}
        {{$x := .M}}
        {{range $i, $k := .Keys}}
          <h2>{{$k}}</h2>
          {{$v := index $x $k}}
          {{range $z, $s := $v}}
            {{if eq $s.State 4}}
              <div class="result red">
                <div id="output-box-{{$s.ID}}-button" class="toggle"
                  onclick="javascript:toggleOutput('output-box-{{$s.ID}}')">+</div>
                <h3 style="display: inline;">{{$s.Title}}</h3>
                <br>
                <div>
                  <div id="output-box-{{$s.ID}}" style="display: none;">
                    <pre class="pre-box">{{$s.CapturedOutput}}</pre>
                  </div>
                </div>
                <pre class="fail-message">{{$s.Failure.Message}}</pre>
                <br>
              </div>
    </div>
            {{else if eq $s.State 3}}
              <div class="result green">
                <div id="output-box-{{$s.ID}}-button" class="toggle"
                  onclick="javascript:toggleOutput('output-box-{{$s.ID}}')">+</div>
                <h3 style="display: inline;">{{$s.Title}}</h3>
                <br>
                <div id="output-box-{{$s.ID}}" style="display: none;">
                  <pre class="pre-box">{{$s.CapturedOutput}}</pre>
                </div>
			  </div>
            {{end}}
          {{end}}
        {{end}}
      {{end}}
  </body>
</html>
`
)

type (
	summaryMap struct {
		M    map[string][]specSnapshot
		Keys []string
		Size int
	}

	specSnapshot struct {
		types.SpecSummary
		ID    int
		Title string
	}

	httpDebugWriter struct {
		CapturedOutput []string
		debug          bool
	}

	httpDebugLogger struct {
		l *log.Logger
		w io.Writer
	}
)

func (sm *summaryMap) Add(key string, sum *specSnapshot) {
	sm.M[key] = append(sm.M[key], *sum)
	sm.Size++

	if !sm.containsKey(key) {
		sm.Keys = append(sm.Keys, key)
	}
}

func (sm *summaryMap) containsKey(key string) bool {
	var containsKey bool
	for _, k := range sm.Keys {
		if k == key {
			containsKey = true
			break
		}
	}
	return containsKey
}

func newSpecSnapshot(sum *types.SpecSummary, id int) *specSnapshot {
	return &specSnapshot{SpecSummary: *sum, Title: sum.ComponentTexts[3], ID: id}
}

func newHTTPDebugWriter(debug bool) *httpDebugWriter {
	return &httpDebugWriter{debug: debug}
}

func (writer *httpDebugWriter) Write(b []byte) (int, error) {
	s := string(b)
	writer.CapturedOutput = append(writer.CapturedOutput, s)
	if writer.debug {
		fmt.Println(s)
	}

	return len(b), nil
}

func newHTTPDebugLogger(f io.Writer) *httpDebugLogger {
	debugLogger := &httpDebugLogger{w: f, l: log.New(f, "", log.Ldate|log.Lmicroseconds)}
	return debugLogger
}

func (l *httpDebugLogger) Errorf(format string, v ...interface{}) {
	l.output("ERROR "+format, v...)
}

func (l *httpDebugLogger) Warnf(format string, v ...interface{}) {
	l.output("WARN "+format, v...)
}

func (l *httpDebugLogger) Debugf(format string, v ...interface{}) {
	re := regexp.MustCompile("(?i)(\"?\\w*(authorization|token|state)\\w*\"?(:|=)\\s*)(\")?\\s*((bearer|basic)? )?[^\\s&\"]*(\")?")
	format = re.ReplaceAllString(format, "$1$4$5*****$7")
	l.output("DEBUG "+format, v...)
}

func (l *httpDebugLogger) output(format string, v ...interface{}) {
	if len(v) == 0 {
		l.l.Print(format)
		return
	}
	l.w.Write([]byte(fmt.Sprintf(format, v...)))
}

type (
	HTMLReporter struct {
		htmlReportFilename string
		SpecSummaryMap     summaryMap
		config             config.DefaultReporterConfigType
		debugLogger        *httpDebugWriter
		debugIndex         int
	}
)

func newHTMLReporter(htmlReportFilename string) *HTMLReporter {
	return &HTMLReporter{
		htmlReportFilename: htmlReportFilename,
		debugLogger:        httpWriter,
		SpecSummaryMap:     summaryMap{M: make(map[string][]specSnapshot)},
	}
}

func (reporter *HTMLReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	b := new(bytes.Buffer)
	for _, co := range httpWriter.CapturedOutput[reporter.debugIndex:] {
		b.WriteString(co)
		b.WriteString("\n")
	}
	specSummary.CapturedOutput = b.String()

	header := specSummary.ComponentTexts[2]
	summary := newSpecSnapshot(specSummary, reporter.SpecSummaryMap.Size)
	reporter.SpecSummaryMap.Add(header, summary)
	reporter.debugIndex = len(reporter.debugLogger.CapturedOutput)
}

func (reporter *HTMLReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	t, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		log.Fatal(err)
	}

	htmlReportFilenameAbsPath, err := filepath.Abs(reporter.htmlReportFilename)
	if err != nil {
		log.Fatal(err)
	}

	htmlReportFile, err := os.Create(htmlReportFilenameAbsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer htmlReportFile.Close()

	err = t.ExecuteTemplate(htmlReportFile, "report", &reporter)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("HTML report was created: %s", htmlReportFilenameAbsPath)
}

//unused by HTML reporter
func (reporter *HTMLReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (reporter *HTMLReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
}

func (reporter *HTMLReporter) SpecWillRun(specSummary *types.SpecSummary) {
}

func (reporter *HTMLReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
}
