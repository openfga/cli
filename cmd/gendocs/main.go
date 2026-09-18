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

// Command gendocs regenerates the Commands section of README.md from cobra
// command metadata. Run from the repository root: go run ./cmd/gendocs
package main

import (
	"log"

	"github.com/openfga/cli/cmd"
	"github.com/openfga/cli/internal/doc"
)

func main() {
	root := cmd.RootCmd()

	section := doc.GenerateCommandsSection(root)
	if err := doc.SpliceIntoReadme("README.md", section); err != nil {
		log.Fatalf("gendocs: %v", err)
	}

	toc := doc.GenerateCommandsTOC(root)
	if err := doc.SpliceCommandsTOC("README.md", toc); err != nil {
		log.Fatalf("gendocs: %v", err)
	}
}
