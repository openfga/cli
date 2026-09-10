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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

const minimalTemplate = `version: "1"
rules:
  - name: "rule"
    tuples:
      - user: "user:{{ input.id }}"
        relation: "member"
        object: "org:{{ input.org }}"
`

const starterTemplate = `version: "1"
rules:
  - name: "add-member"
    tuples:
      - user: "user:{{ input.id }}"
        relation: "member"
        object: "org:{{ input.org }}"
  - name: "add-admin"
    when: "input.role == \"admin\""
    tuples:
      - user: "user:{{ input.id }}"
        relation: "admin"
        object: "org:{{ input.org }}"
tests:
  - name: "admin gets member and admin"
    input:
      id: "anne"
      org: "acme"
      role: "admin"
    expect_tuples:
      - user: "user:anne"
        relation: "member"
        object: "org:acme"
      - user: "user:anne"
        relation: "admin"
        object: "org:acme"
  - name: "regular user only gets member"
    input:
      id: "bob"
      org: "acme"
      role: "user"
    expect_tuples:
      - user: "user:bob"
        relation: "member"
        object: "org:acme"
`

// stdinIsTTY reports whether stdin is an interactive terminal.
func stdinIsTTY() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return stat.Mode()&os.ModeCharDevice != 0
}

// huhConfirm is the default confirm function: prompts on a TTY, errors otherwise.
func huhConfirm(path string) (bool, error) {
	if !stdinIsTTY() {
		return false, fmt.Errorf("%s already exists — use --force to overwrite: %w", path, fs.ErrExist)
	}

	var overwrite bool

	if err := huh.NewConfirm().
		Title(path + " already exists. Overwrite?").
		Value(&overwrite).
		Run(); err != nil {
		return false, fmt.Errorf("prompt: %w", err)
	}

	return overwrite, nil
}

func initMapping(path string, minimal, force bool, out io.Writer) error {
	return initMappingWithConfirm(path, minimal, force, out, huhConfirm)
}

func initMappingWithConfirm(path string, minimal, force bool, out io.Writer, confirm func(string) (bool, error)) error {
	if !force {
		_, err := os.Stat(path)
		if err == nil {
			overwrite, confirmErr := confirm(path)

			switch {
			case errors.Is(confirmErr, huh.ErrUserAborted):
				fmt.Fprintln(out, "Aborted.")

				return nil
			case confirmErr != nil:
				return confirmErr
			case !overwrite:
				fmt.Fprintln(out, "Aborted.")

				return nil
			}
		} else if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("checking %s: %w", path, err)
		}
	}

	content := starterTemplate
	if minimal {
		content = minimalTemplate
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil { //nolint:gosec
		return fmt.Errorf("writing %s: %w", path, err)
	}

	fmt.Fprintf(out, "Created %s\n", path)

	return nil
}

var (
	initMinimal bool
	initForce   bool
)

var initCmd = &cobra.Command{
	Use:   "init [mapping-file]",
	Short: "Scaffold a starter mapping file",
	Long:  "Creates a new mapping YAML file with a sample rule and embedded test. Defaults to mapping.yaml.",
	Example: `  fga mapping init
  fga mapping init my-mapping.yaml
  fga mapping init --minimal mapping.yaml
  fga mapping init --force mapping.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "mapping.yaml"
		if len(args) == 1 {
			path = args[0]
		}

		return initMapping(path, initMinimal, initForce, cmd.OutOrStdout())
	},
}

func init() {
	initCmd.Flags().BoolVar(&initMinimal, "minimal", false, "Emit a skeleton without the example test block")
	initCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite an existing file")
}
