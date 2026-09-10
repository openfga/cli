/*
Copyright © 2023 OpenFGA

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mapping

import (
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/openfga/mapper"
	"github.com/openfga/mapper/language"
)

type junitTestSuites struct {
	XMLName    xml.Name     `xml:"testsuites"`
	TestSuites []junitSuite `xml:"testsuite"`
}

type junitSuite struct {
	Name      string      `xml:"name,attr"`
	Tests     int         `xml:"tests,attr"`
	Failures  int         `xml:"failures,attr"`
	Errors    int         `xml:"errors,attr"`
	Skipped   int         `xml:"skipped,attr,omitempty"`
	Time      string      `xml:"time,attr"`
	TestCases []junitCase `xml:"testcase"`
}

type junitCase struct {
	Name      string        `xml:"name,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
	Error     *junitError   `xml:"error,omitempty"`
	SystemOut *string       `xml:"system-out,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

type junitError struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

// junitSuiteName extracts a suite name from the mapping file path.
func junitSuiteName(mappingFile string) string {
	if mappingFile == "" {
		return "mapping"
	}

	base := filepath.Base(mappingFile)
	ext := filepath.Ext(base)

	return strings.TrimSuffix(base, ext)
}

// formatJUnit writes results as JUnit XML.
func formatJUnit(out io.Writer, run testRunResult, verbose bool) error {
	_, failed, errCount := countResults(run.results)
	dur := totalDuration(run.results)

	suite := junitSuite{
		Name:     junitSuiteName(run.mappingFile),
		Tests:    len(run.results) + run.filtered + run.skipped,
		Failures: failed,
		Errors:   errCount,
		Skipped:  run.filtered + run.skipped,
		Time:     fmt.Sprintf("%.3f", dur.Seconds()),
	}

	for _, result := range run.results {
		testCase := junitCase{
			Name: result.Name,
			Time: fmt.Sprintf("%.3f", result.Duration.Seconds()),
		}

		switch {
		case result.Error != nil:
			testCase.Error = &junitError{
				Message: result.Error.Error(),
			}
		case !result.Passed:
			missing := sortTuples(missingTuples(result.Expected, result.Actual))
			extra := sortTuples(extraTuples(result.Expected, result.Actual))
			testCase.Failure = &junitFailure{
				Message: fmt.Sprintf("expected %d tuples, got %d", len(result.Expected), len(result.Actual)),
				Body:    buildJUnitFailureBody(missing, extra),
			}
		}

		if verbose && result.Trace != nil {
			if sysOut := buildJUnitSystemOut(result.Trace); sysOut != "" {
				testCase.SystemOut = &sysOut
			}
		}

		suite.TestCases = append(suite.TestCases, testCase)
	}

	return writeJUnitXML(out, junitTestSuites{TestSuites: []junitSuite{suite}})
}

// formatJUnitCompileError writes a compilation error as valid JUnit XML.
func formatJUnitCompileError(out io.Writer, err error) error {
	suite := junitSuite{
		Name:   "mapping",
		Tests:  1,
		Errors: 1,
		Time:   "0.000",
		TestCases: []junitCase{{
			Name: "compilation",
			Time: "0.000",
			Error: &junitError{
				Message: "failed to compile mapping file",
				Body:    err.Error(),
			},
		}},
	}

	return writeJUnitXML(out, junitTestSuites{TestSuites: []junitSuite{suite}})
}

// writeJUnitXML marshals the given testsuites to out as indented XML with a header.
func writeJUnitXML(out io.Writer, suites junitTestSuites) error {
	xmlBytes, err := xml.MarshalIndent(suites, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling junit xml: %w", err)
	}

	_, err = fmt.Fprintf(out, "%s%s\n", xml.Header, xmlBytes)

	return err //nolint:wrapcheck
}

// buildJUnitSystemOut builds the system-out content from trace information.
func buildJUnitSystemOut(trace *mapper.Trace) string {
	var builder strings.Builder

	var matched, skipped []string

	for _, rt := range trace.Rules {
		switch rt.Status {
		case mapper.RuleMatched:
			matched = append(matched, fmt.Sprintf("%s -> %d tuples", rt.Name, rt.EmittedN))
		case mapper.RuleSkipped:
			skipped = append(skipped, rt.Name)
		case mapper.RuleErrored:
		}
	}

	if len(matched) > 0 {
		fmt.Fprintf(&builder, "rules matched: %s\n", strings.Join(matched, ", "))
	}

	if len(skipped) > 0 {
		fmt.Fprintf(&builder, "rules skipped: %s (when guard false)\n", strings.Join(skipped, ", "))
	}

	return builder.String()
}

// buildJUnitFailureBody renders the human-readable diff for a JUnit failure element.
func buildJUnitFailureBody(missing, extra []language.Tuple) string {
	var builder strings.Builder

	builder.WriteString("\n")

	if len(missing) > 0 {
		builder.WriteString("missing (expected but not produced):\n")

		for _, tup := range missing {
			fmt.Fprintf(&builder, "  %s  %s  %s  [%s]\n", tup.User, tup.Relation, tup.Object, tup.Action)
		}
	}

	if len(extra) > 0 {
		if len(missing) > 0 {
			builder.WriteString("\n")
		}

		builder.WriteString("extra (produced but not expected):\n")

		for _, tup := range extra {
			fmt.Fprintf(&builder, "  %s  %s  %s  [%s]\n", tup.User, tup.Relation, tup.Object, tup.Action)
		}
	}

	return builder.String()
}
