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

// Package doc generates the Commands section of README.md from cobra command metadata.
package doc

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	beginMarker    = "<!-- BEGIN_COMMANDS -->"
	endMarker      = "<!-- END_COMMANDS -->"
	beginTOCMarker = "<!-- BEGIN_COMMANDS_TOC -->"
	endTOCMarker   = "<!-- END_COMMANDS_TOC -->"
)

// GenerateCommandsSection walks the cobra command tree rooted at rootCmd and
// returns a markdown string suitable for splicing between the README markers.
// Hidden and unavailable commands are excluded.
func GenerateCommandsSection(rootCmd *cobra.Command) string {
	var b strings.Builder

	for _, cmd := range rootCmd.Commands() {
		if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		if cmd.Runnable() {
			continue // top-level runnables (e.g. version) have no group heading — skip, consistent with TOC
		}

		writeSection(&b, cmd)
	}

	return b.String()
}

func writeSection(b *strings.Builder, cmd *cobra.Command) {
	if !cmd.Runnable() {
		fmt.Fprintf(b, "#### %s\n\n", cmd.Short)

		for _, sub := range cmd.Commands() {
			if !sub.IsAvailableCommand() || sub.IsAdditionalHelpTopicCommand() {
				continue
			}

			writeSection(b, sub)
		}

		return
	}

	fmt.Fprintf(b, "##### %s\n\n", cmd.Short)

	fmt.Fprintf(b, "###### Command\n\n```\n%s\n```\n\n", cmd.UseLine())

	if hasVisibleLocalFlags(cmd) {
		fmt.Fprintf(b, "###### Parameters\n\n")

		cmd.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
			if f.Hidden {
				return
			}

			fmt.Fprintf(b, "* `--%s`: %s\n", f.Name, f.Usage)
		})

		fmt.Fprintf(b, "\n")
	}

	if cmd.Example != "" {
		fmt.Fprintf(b, "###### Example\n\n```bash\n%s\n```\n\n", strings.TrimSpace(cmd.Example))
	}

	if response, ok := cmd.Annotations["docs:response"]; ok && response != "" {
		lang := cmd.Annotations["docs:response:lang"]
		if lang == "" {
			lang = "json"
		}

		fmt.Fprintf(b, "###### Response\n\n```%s\n%s\n```\n\n", lang, response)
	}
}

func hasVisibleLocalFlags(cmd *cobra.Command) bool {
	hasVisible := false

	cmd.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			hasVisible = true
		}
	})

	return hasVisible
}

// SpliceIntoReadme replaces the content between the BEGIN_COMMANDS and
// END_COMMANDS HTML comment markers in the file at readmePath with section.
func SpliceIntoReadme(readmePath, section string) error {
	return splice(readmePath, beginMarker, endMarker, section)
}

// SpliceCommandsTOC replaces the content between the BEGIN_COMMANDS_TOC and
// END_COMMANDS_TOC HTML comment markers in the file at readmePath with toc.
func SpliceCommandsTOC(readmePath, toc string) error {
	return splice(readmePath, beginTOCMarker, endTOCMarker, toc)
}

func splice(readmePath, begin, end, section string) error {
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", readmePath, err)
	}

	text := string(content)
	beginIdx := strings.Index(text, begin)
	endIdx := strings.Index(text, end)

	if beginIdx == -1 || endIdx == -1 {
		return fmt.Errorf("splice markers not found in %s: expected %q and %q", readmePath, begin, end)
	}

	if beginIdx > endIdx {
		return fmt.Errorf("splice markers out of order in %s: %q appears after %q", readmePath, begin, end)
	}

	newContent := text[:beginIdx+len(begin)] + "\n" + section + text[endIdx:]

	return os.WriteFile(readmePath, []byte(newContent), 0o644) //nolint:wrapcheck
}

// GenerateCommandsTOC returns the markdown list entry for Commands and all its
// subgroups, suitable for splicing between the BEGIN_COMMANDS_TOC markers.
// The marker wraps the entire - [Commands] subtree so no HTML comment
// interrupts the list and triggers code-block parsing.
func GenerateCommandsTOC(rootCmd *cobra.Command) string {
	var b strings.Builder

	b.WriteString("  - [Commands](#commands)\n")

	for _, cmd := range rootCmd.Commands() {
		if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
			continue
		}

		writeTOCSection(&b, cmd)
	}

	return b.String()
}

func writeTOCSection(b *strings.Builder, cmd *cobra.Command) {
	if cmd.Runnable() {
		return // top-level runnables (e.g. version) are not command groups — skip
	}

	fmt.Fprintf(b, "    - [%s](#%s)\n", cmd.Short, anchor(cmd.Short))

	for _, sub := range cmd.Commands() {
		if !sub.IsAvailableCommand() || sub.IsAdditionalHelpTopicCommand() {
			continue
		}

		fmt.Fprintf(b, "      - [%s](#%s)\n", sub.Short, anchor(sub.Short))
	}
}

// anchor converts a heading string to a GitHub markdown anchor.
// "Read a Single Authorization Model" -> "read-a-single-authorization-model".
func anchor(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder

	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}

	return b.String()
}
