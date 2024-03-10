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

	"github.com/gocarina/gocsv"
	"gopkg.in/yaml.v3"

	"github.com/mattn/go-isatty"
	"github.com/nwidger/jsoncolor"
)

// Printer is a content type agnostic interface for displaying data.
type Printer interface {
	DisplayNoColor(data any) error
	DisplayColor(data any) error
}

// jsonPrinter implements the Printer interface for JSON output.
type jsonPrinter struct{}

// csvPrinter implements the Printer interface for CSV output.
type csvPrinter struct{}

// yamlPrinter implements the Printer interface for YAML output.
type yamlPrinter struct{}

func (prt *jsonPrinter) DisplayNoColor(data any) error {
	result, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal json with error %w", err)
	}

	fmt.Println(string(result))

	return nil
}

func (prt *jsonPrinter) DisplayColor(data any) error {
	// create custom formatter
	f := jsoncolor.NewFormatter()

	dst, err := jsoncolor.MarshalIndentWithFormatter(data, "", "  ", f)
	if err != nil {
		return fmt.Errorf("unable to display output with error %w", err)
	}

	fmt.Println(string(dst))

	return nil
}

func (prt *csvPrinter) DisplayColor(data any) error {
	return prt.DisplayNoColor(data)
}

func (prt *csvPrinter) DisplayNoColor(data any) error {
	b, err := gocsv.MarshalBytes(data)
	if err != nil {
		return fmt.Errorf("unable to marshal CSV with error: %w", err)
	}

	fmt.Println(string(b))

	return nil
}

func (prt *yamlPrinter) DisplayColor(data any) error {
	return prt.DisplayNoColor(data)
}

func (prt *yamlPrinter) DisplayNoColor(data any) error {
	result, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to marshal yaml with error %w", err)
	}

	fmt.Println(string(result))

	return nil
}

// UniPrinter is a universal printer that can handle different output formats.
type UniPrinter struct {
	Colorful bool
	Printer  Printer
}

// NewUniPrinter creates a new UniPrinter based on the specified output format and optional functional options.
func NewUniPrinter(outputFormat string) *UniPrinter {
	uniPrinter := UniPrinter{Colorful: true}
	if os.Getenv("NO_COLOR") != "" {
		uniPrinter.Colorful = false
	}

	switch outputFormat {
	case "yaml":
		uniPrinter.Printer = &yamlPrinter{}
	case "csv":
		uniPrinter.Printer = &csvPrinter{}
	default:
		uniPrinter.Printer = &jsonPrinter{}
	}

	return &uniPrinter
}

// Display prints the data using the configured printer and color settings.
func (prt UniPrinter) Display(data any) error {
	if prt.Colorful {
		err := prt.Printer.DisplayColor(data)
		if err != nil {
			return fmt.Errorf("failed to display colorful output: %w", err)
		}

		return nil
	}

	err := prt.Printer.DisplayNoColor(data)
	if err != nil {
		return fmt.Errorf("failed to display output: %w", err)
	}

	return nil
}

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
