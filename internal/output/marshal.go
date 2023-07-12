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

// Package output handles functions relating to displaying the final output
package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/nwidger/jsoncolor"
)

// EmptyStruct is used when we wish to return an empty object.
type EmptyStruct struct{}

func displayColorTerminal(data any) error {
	// create custom formatter
	f := jsoncolor.NewFormatter()

	dst, err := jsoncolor.MarshalIndentWithFormatter(data, "", "  ", f)
	if err != nil {
		return fmt.Errorf("unable to display output with error %w", err)
	}

	fmt.Println(string(dst))

	return nil
}

func displayNoColorTerminal(data any) error {
	result, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal json with error %w", err)
	}

	fmt.Println(string(result))

	return nil
}

func outputNonTerminal(data any) error {
	result, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to marshal json with error %w", err)
	}

	fmt.Println(string(result))

	return nil
}

// Display will decorate the output if possible.  Otherwise, will print out the standard JSON.
func Display(data any) error {
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		if os.Getenv("NO_COLOR") != "" {
			return displayNoColorTerminal(data)
		}

		return displayColorTerminal(data)
	}

	return outputNonTerminal(data)
}
