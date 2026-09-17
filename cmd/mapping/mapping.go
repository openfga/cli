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

// Package mapping implements the fga mapping command group.
package mapping

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// MappingCmd is the root of the fga mapping command group.
var MappingCmd = &cobra.Command{
	Use:   "mapping",
	Short: "Manage JSON-to-tuple mappings",
	Long:  "Validate, test, and run JSON-to-tuple mapping files.",
}

// promptMappingFile resolves the mapping file path for a command that accepts an
// optional path argument. An explicit argument is returned unchanged. With no
// argument and an interactive stdin it prompts with a .yaml/.yml-scoped file
// picker; the picker renders to stderr (huh's default) so a piped stdout is never
// corrupted. A missing path with a non-interactive stdin, a cancelled picker, or
// an empty selection is a usage error: it is reported to errOut and exits with
// status 2.
func promptMappingFile(args []string, errOut io.Writer) string {
	if len(args) > 0 {
		return args[0]
	}

	if !isatty.IsTerminal(os.Stdin.Fd()) {
		fmt.Fprintln(errOut, "Error: mapping file path is required")
		os.Exit(2)
	}

	path := ""

	if err := huh.NewFilePicker().
		Title("Mapping file").
		CurrentDirectory(".").
		AllowedTypes([]string{".yaml", ".yml"}).
		ShowHidden(false).
		Picking(true).
		Height(15).
		Value(&path).
		Run(); err != nil || path == "" {
		fmt.Fprintln(errOut, "Error: mapping file path is required")
		os.Exit(2)
	}

	return path
}

func init() {
	MappingCmd.AddCommand(validateCmd)
	MappingCmd.AddCommand(testCmd)
	MappingCmd.AddCommand(initCmd)
	MappingCmd.AddCommand(runCmd)
}
