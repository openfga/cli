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

package output

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/itchyny/gojq"
)

// jqFilter is the jq expression set with the --jq flag, applied to the JSON output of every command.
var jqFilter *gojq.Code

// SetJQFilter compiles a jq expression that filters the output of Display and UniPrinter.Display.
// An empty expression turns filtering off.
func SetJQFilter(expression string) error {
	if expression == "" {
		jqFilter = nil

		return nil
	}

	query, err := gojq.Parse(expression)
	if err != nil {
		return fmt.Errorf("invalid jq expression %q: %w", expression, err)
	}

	code, err := gojq.Compile(query)
	if err != nil {
		return fmt.Errorf("invalid jq expression %q: %w", expression, err)
	}

	jqFilter = code

	return nil
}

// applyJQFilter runs the jq filter on the JSON form of data and returns its results.
func applyJQFilter(code *gojq.Code, data any) ([]any, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json with error %w", err)
	}

	var input any
	if err := json.Unmarshal(encoded, &input); err != nil {
		return nil, fmt.Errorf("unable to unmarshal json with error %w", err)
	}

	var results []any

	iter := code.Run(input)

	for {
		result, ok := iter.Next()
		if !ok {
			break
		}

		if err, isErr := result.(error); isErr {
			var haltErr *gojq.HaltError
			if errors.As(err, &haltErr) && haltErr.Value() == nil {
				break
			}

			return nil, fmt.Errorf("failed to apply jq expression: %w", err)
		}

		results = append(results, result)
	}

	return results, nil
}

// displayFiltered applies the jq filter to data and prints each result: strings as raw text, like
// `jq -r`, so they can be used directly in scripts, and any other value with display.
func displayFiltered(code *gojq.Code, data any, display func(any) error) error {
	results, err := applyJQFilter(code, data)
	if err != nil {
		return err
	}

	for _, result := range results {
		if text, isString := result.(string); isString {
			fmt.Println(text)

			continue
		}

		if err := display(result); err != nil {
			return err
		}
	}

	return nil
}
