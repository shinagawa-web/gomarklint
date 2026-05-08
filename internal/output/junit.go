package output

import (
	"encoding/xml"
	"fmt"
	"io"
)

// JUnitFormatter formats lint results as JUnit XML.
type JUnitFormatter struct{}

// NewJUnitFormatter creates a new JUnitFormatter.
func NewJUnitFormatter() *JUnitFormatter {
	return &JUnitFormatter{}
}

type junitTestSuites struct {
	XMLName  xml.Name         `xml:"testsuites"`
	Name     string           `xml:"name,attr"`
	Tests    int              `xml:"tests,attr"`
	Failures int              `xml:"failures,attr"`
	Time     string           `xml:"time,attr"`
	Suites   []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Cases    []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

// Format implements the Formatter interface for JUnit XML output.
func (f *JUnitFormatter) Format(w io.Writer, result *Result) error {
	totalTests := 0
	totalFailures := 0
	suites := make([]junitTestSuite, 0, len(result.OrderedPaths))

	for _, path := range result.OrderedPaths {
		errs := result.Details[path]
		suite := junitTestSuite{
			Name: path,
		}
		if len(errs) == 0 {
			suite.Tests = 1
			suite.Cases = []junitTestCase{{Name: path, Classname: path}}
		} else {
			suite.Tests = len(errs)
			suite.Failures = len(errs)
			suite.Cases = make([]junitTestCase, 0, len(errs))
			for _, e := range errs {
				severity := e.Severity
				if severity == "" {
					severity = "error"
				}
				tc := junitTestCase{
					Name:      fmt.Sprintf("line %d: %s", e.Line, e.Message),
					Classname: path,
					Failure: &junitFailure{
						Message: e.Message,
						Type:    severity,
					},
				}
				suite.Cases = append(suite.Cases, tc)
			}
		}
		totalTests += suite.Tests
		totalFailures += suite.Failures
		suites = append(suites, suite)
	}

	root := junitTestSuites{
		Name:     "gomarklint",
		Tests:    totalTests,
		Failures: totalFailures,
		Time:     fmt.Sprintf("%.2f", result.Duration.Seconds()),
		Suites:   suites,
	}

	if _, err := fmt.Fprint(w, xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(root); err != nil {
		return err
	}
	return enc.Flush()
}
