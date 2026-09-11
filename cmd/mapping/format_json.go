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
	"encoding/json"
	"fmt"
	"io"

	"github.com/openfga/mapper/language"
)

type jsonOutput struct {
	Summary jsonSummary  `json:"summary"`
	Results []jsonResult `json:"results"`
	Warning string       `json:"warning,omitempty"`
}

type jsonSummary struct {
	Passed     int `json:"passed"`
	Failed     int `json:"failed"`
	Errors     int `json:"errors"`
	Total      int `json:"total"`
	DurationMS int `json:"duration_ms"`
}

type jsonResult struct {
	Name       string      `json:"name"`
	Status     string      `json:"status"`
	DurationMS int         `json:"duration_ms"`
	Expected   []jsonTuple `json:"expected,omitempty"`
	Actual     []jsonTuple `json:"actual,omitempty"`
	Diff       *jsonDiff   `json:"diff,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type jsonTuple struct {
	User      string         `json:"user"`
	Relation  string         `json:"relation"`
	Object    string         `json:"object"`
	Action    string         `json:"action"`
	Condition string         `json:"condition,omitempty"`
	Context   map[string]any `json:"context,omitempty"`
}

type jsonDiff struct {
	Missing []jsonTuple `json:"missing"`
	Extra   []jsonTuple `json:"extra"`
}

type jsonCompileError struct {
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

func toJSONTuples(tuples []language.Tuple) []jsonTuple {
	if len(tuples) == 0 {
		return nil
	}

	out := make([]jsonTuple, len(tuples))

	for idx, tup := range tuples {
		out[idx] = jsonTuple{
			User:      tup.User,
			Relation:  tup.Relation,
			Object:    tup.Object,
			Action:    string(tup.Action),
			Condition: tup.Condition,
			Context:   tup.Context,
		}
	}

	return out
}

// formatJSON writes results as a JSON document.
func formatJSON(out io.Writer, run testRunResult) error {
	passed, failed, errCount := countResults(run.results)
	dur := totalDuration(run.results)

	output := jsonOutput{
		Summary: jsonSummary{
			Passed:     passed,
			Failed:     failed,
			Errors:     errCount,
			Total:      len(run.results) + run.filtered + run.skipped,
			DurationMS: int(dur.Milliseconds()),
		},
		Results: make([]jsonResult, 0, len(run.results)),
	}

	if run.filterUsed != "" && len(run.results) == 0 {
		output.Warning = fmt.Sprintf("no tests matched filter %q", run.filterUsed)
	}

	for _, result := range run.results {
		jsonRes := jsonResult{
			Name:       result.Name,
			DurationMS: int(result.Duration.Milliseconds()),
		}

		switch {
		case result.Error != nil:
			jsonRes.Status = "error"
			jsonRes.Error = result.Error.Error()
			jsonRes.Expected = toJSONTuples(result.Expected)
			jsonRes.Actual = toJSONTuples(result.Actual)
		case result.Passed:
			jsonRes.Status = "pass"
		default:
			jsonRes.Status = "fail"
			jsonRes.Expected = toJSONTuples(result.Expected)
			jsonRes.Actual = toJSONTuples(result.Actual)

			jsonRes.Diff = &jsonDiff{
				Missing: toJSONTuples(missingTuples(result.Expected, result.Actual)),
				Extra:   toJSONTuples(extraTuples(result.Expected, result.Actual)),
			}
			if jsonRes.Diff.Missing == nil {
				jsonRes.Diff.Missing = []jsonTuple{}
			}

			if jsonRes.Diff.Extra == nil {
				jsonRes.Diff.Extra = []jsonTuple{}
			}
		}

		output.Results = append(output.Results, jsonRes)
	}

	return encodeJSON(out, output)
}

// formatJSONCompileError writes a compilation error as a JSON envelope.
func formatJSONCompileError(out io.Writer, err error) error {
	return encodeJSON(out, jsonCompileError{
		Error:  "failed to compile mapping file",
		Detail: err.Error(),
	})
}

// encodeJSON marshals v as pretty-printed JSON to out.
func encodeJSON(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")

	return enc.Encode(v) //nolint:wrapcheck
}
