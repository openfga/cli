/*
Copyright © 2026 OpenFGA

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

package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/itchyny/gojq"
)

// JQUnsupportedAnnotation marks a command, and the commands under it, as never printing through
// Display, so that --jq is rejected before the command runs.
const JQUnsupportedAnnotation = "jq-unsupported"

// ErrJQNotApplied is returned by CheckJQFilterApplied when --jq is set but the command printed
// its output without going through the filter.
var ErrJQNotApplied = errors.New("--jq is only supported by commands that print JSON or YAML output")

var (
	// jqFilter is the jq expression set with the --jq flag, applied to the output of Display and
	// UniPrinter.Display.
	jqFilter *gojq.Code

	// jqApplied records whether the filter has been applied to any output.
	jqApplied bool
)

// SetJQFilter compiles a jq expression that filters the output of Display and UniPrinter.Display.
// An empty expression turns filtering off.
func SetJQFilter(expression string) error {
	jqApplied = false

	if expression == "" {
		jqFilter = nil

		return nil
	}

	query, err := gojq.Parse(expression)
	if err != nil {
		return fmt.Errorf("invalid jq expression %q: %w", expression, err)
	}

	// Pass the environment, so filters can use env.NAME and $ENV.NAME like in jq.
	code, err := gojq.Compile(query, gojq.WithEnvironLoader(os.Environ))
	if err != nil {
		return fmt.Errorf("invalid jq expression %q: %w", expression, err)
	}

	jqFilter = code

	return nil
}

// JQFilterSet reports whether a --jq filter is set.
func JQFilterSet() bool {
	return jqFilter != nil
}

// CheckJQFilterApplied returns ErrJQNotApplied if a --jq filter is set but no output went
// through it, so that commands whose output is not filtered do not silently ignore the flag.
func CheckJQFilterApplied() error {
	if jqFilter != nil && !jqApplied {
		return ErrJQNotApplied
	}

	return nil
}

// runJQFilter runs the jq filter on the JSON form of data and calls emit with each result as it
// is produced, so results emitted before an error are not lost.
func runJQFilter(code *gojq.Code, data any, emit func(any) error) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to marshal json with error %w", err)
	}

	// Decode numbers as json.Number, so integers above 2^53 keep their precision in gojq.
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()

	var input any
	if err := decoder.Decode(&input); err != nil {
		return fmt.Errorf("unable to unmarshal json with error %w", err)
	}

	iter := code.Run(input)

	for {
		result, ok := iter.Next()
		if !ok {
			return nil
		}

		if err, isErr := result.(error); isErr {
			// halt stops the filter successfully, halt_error with a non-zero exit code is an error.
			var haltErr *gojq.HaltError
			if errors.As(err, &haltErr) && haltErr.Value() == nil && haltErr.ExitCode() == 0 {
				return nil
			}

			return fmt.Errorf("failed to apply jq expression: %w", err)
		}

		if err := emit(result); err != nil {
			return err
		}
	}
}

// displayFiltered applies the jq filter to data and prints each result: strings as raw text, like
// `jq -r`, so they can be used directly in scripts, and any other value with display.
func displayFiltered(code *gojq.Code, data any, display func(any) error) error {
	jqApplied = true

	return runJQFilter(code, data, func(result any) error {
		if text, isString := result.(string); isString {
			fmt.Println(text)

			return nil
		}

		return display(result)
	})
}
