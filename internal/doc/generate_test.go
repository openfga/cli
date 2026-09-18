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

package doc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openfga/cli/internal/doc"
)

// buildTestTree builds a minimal cobra tree mirroring the real fga structure.
func buildTestTree() *cobra.Command {
	root := &cobra.Command{Use: "fga", Short: "OpenFGA CLI"}

	storeGroup := &cobra.Command{
		Use:   "store",
		Short: "Stores",
		Long:  "Manage OpenFGA stores.",
	}

	createCmd := &cobra.Command{
		Use:     "create",
		Short:   "Create Store",
		Long:    "Create an OpenFGA store.",
		Example: `fga store create --name "FGA Demo Store"`,
		Annotations: map[string]string{
			"docs:response": `{"id": "01H0H015178Y2V4CX10C2KGHF4", "name": "FGA Demo Store"}`,
		},
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	createCmd.Flags().String("name", "", "Store Name")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List Stores",
		Long:  "List OpenFGA stores.",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}

	hiddenCmd := &cobra.Command{
		Use:    "hidden",
		Short:  "Hidden Command",
		Hidden: true,
		RunE:   func(cmd *cobra.Command, args []string) error { return nil },
	}

	storeGroup.AddCommand(createCmd, listCmd, hiddenCmd)
	root.AddCommand(storeGroup)

	return root
}

func TestGenerateCommandsSection_GroupHeader(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	section := doc.GenerateCommandsSection(root)

	assert.Contains(t, section, "#### Stores")
}

func TestGenerateCommandsSection_CommandHeader(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	section := doc.GenerateCommandsSection(root)

	assert.Contains(t, section, "##### Create Store")
}

func TestGenerateCommandsSection_CommandSection(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	section := doc.GenerateCommandsSection(root)

	assert.Contains(t, section, "###### Command")
	assert.Contains(t, section, "fga store create")
}

func TestGenerateCommandsSection_ParametersSection(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	section := doc.GenerateCommandsSection(root)

	assert.Contains(t, section, "###### Parameters")
	assert.Contains(t, section, "`--name`")
	assert.Contains(t, section, "Store Name")
}

func TestGenerateCommandsSection_ExampleSection(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	section := doc.GenerateCommandsSection(root)

	assert.Contains(t, section, "###### Example")
	assert.Contains(t, section, `FGA Demo Store`)
}

func TestGenerateCommandsSection_ResponseSection(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	section := doc.GenerateCommandsSection(root)

	assert.Contains(t, section, "###### Response")
	assert.Contains(t, section, "01H0H015178Y2V4CX10C2KGHF4")
}

func TestGenerateCommandsSection_OmitsParametersSectionWhenNoFlags(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	section := doc.GenerateCommandsSection(root)

	// list command has no flags — find its section and verify no Parameters header follows
	listIdx := strings.Index(section, "##### List Stores")
	require.NotEqual(t, -1, listIdx)

	// Find the next command header after list
	nextCmdIdx := strings.Index(section[listIdx+1:], "#####")
	var listSection string
	if nextCmdIdx == -1 {
		listSection = section[listIdx:]
	} else {
		listSection = section[listIdx : listIdx+1+nextCmdIdx]
	}

	assert.NotContains(t, listSection, "###### Parameters")
}

func TestGenerateCommandsSection_OmitsExampleSectionWhenEmpty(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	section := doc.GenerateCommandsSection(root)

	// list command has no Example field
	listIdx := strings.Index(section, "##### List Stores")
	require.NotEqual(t, -1, listIdx)

	nextCmdIdx := strings.Index(section[listIdx+1:], "#####")
	var listSection string
	if nextCmdIdx == -1 {
		listSection = section[listIdx:]
	} else {
		listSection = section[listIdx : listIdx+1+nextCmdIdx]
	}

	assert.NotContains(t, listSection, "###### Example")
}

func TestGenerateCommandsSection_OmitsResponseSectionWhenNoAnnotation(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	section := doc.GenerateCommandsSection(root)

	listIdx := strings.Index(section, "##### List Stores")
	require.NotEqual(t, -1, listIdx)

	nextCmdIdx := strings.Index(section[listIdx+1:], "#####")
	var listSection string
	if nextCmdIdx == -1 {
		listSection = section[listIdx:]
	} else {
		listSection = section[listIdx : listIdx+1+nextCmdIdx]
	}

	assert.NotContains(t, listSection, "###### Response")
}

func TestGenerateCommandsSection_HiddenCommandsExcluded(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	section := doc.GenerateCommandsSection(root)

	assert.NotContains(t, section, "Hidden Command")
}

func TestSpliceIntoReadme_ReplacesContentBetweenMarkers(t *testing.T) {
	t.Parallel()

	original := "# Title\n\n<!-- BEGIN_COMMANDS -->\nold content here\n<!-- END_COMMANDS -->\n\n# Footer\n"
	newSection := "new generated content"

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(path, []byte(original), 0o600))

	err := doc.SpliceIntoReadme(path, newSection)
	require.NoError(t, err)

	result, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(result)
	assert.Contains(t, content, "# Title")
	assert.Contains(t, content, "<!-- BEGIN_COMMANDS -->")
	assert.Contains(t, content, newSection)
	assert.Contains(t, content, "<!-- END_COMMANDS -->")
	assert.Contains(t, content, "# Footer")
	assert.NotContains(t, content, "old content here")
}

func TestSpliceIntoReadme_ErrorsWhenMarkersNotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(path, []byte("# No markers here\n"), 0o600))

	err := doc.SpliceIntoReadme(path, "content")
	require.Error(t, err)
}

func TestGenerateCommandsTOC_StartsWithCommandsHeader(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	toc := doc.GenerateCommandsTOC(root)

	assert.True(t, strings.HasPrefix(toc, "  - [Commands](#commands)\n"))
}

func TestGenerateCommandsTOC_GroupEntries(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	toc := doc.GenerateCommandsTOC(root)

	assert.Contains(t, toc, "- [Stores](#stores)")
}

func TestGenerateCommandsTOC_CommandEntries(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	toc := doc.GenerateCommandsTOC(root)

	assert.Contains(t, toc, "- [Create Store](#create-store)")
	assert.Contains(t, toc, "- [List Stores](#list-stores)")
}

func TestGenerateCommandsTOC_CommandsIndentedUnderGroup(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	toc := doc.GenerateCommandsTOC(root)

	// Group at 4-space indent, commands at 6-space indent
	assert.Contains(t, toc, "    - [Stores](#stores)\n")
	assert.Contains(t, toc, "      - [Create Store](#create-store)\n")
}

func TestGenerateCommandsTOC_HiddenCommandsExcluded(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	toc := doc.GenerateCommandsTOC(root)

	assert.NotContains(t, toc, "Hidden Command")
}

func TestSpliceCommandsTOC_ReplacesContentBetweenMarkers(t *testing.T) {
	t.Parallel()

	original := "# Title\n\n<!-- BEGIN_COMMANDS_TOC -->\nold toc\n<!-- END_COMMANDS_TOC -->\n\n# Footer\n"
	newTOC := "new toc content"

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(path, []byte(original), 0o600))

	err := doc.SpliceCommandsTOC(path, newTOC)
	require.NoError(t, err)

	result, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(result)
	assert.Contains(t, content, newTOC)
	assert.NotContains(t, content, "old toc")
}

func TestSpliceIntoReadme_ErrorsWhenMarkersOutOfOrder(t *testing.T) {
	t.Parallel()

	// END before BEGIN — splice must error rather than produce garbled output
	original := "<!-- END_COMMANDS -->\nsome content\n<!-- BEGIN_COMMANDS -->\n"

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(path, []byte(original), 0o600))

	err := doc.SpliceIntoReadme(path, "content")
	require.Error(t, err)
}

func TestGenerateCommandsSection_TopLevelRunnableExcluded(t *testing.T) {
	t.Parallel()

	root := buildTestTree()
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(versionCmd)

	section := doc.GenerateCommandsSection(root)

	assert.NotContains(t, section, "Print version")
}

func TestGenerateCommandsSection_NonJSONResponseLang(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "fga"}
	group := &cobra.Command{Use: "model", Short: "Models"}
	cmd := &cobra.Command{
		Use:   "transform",
		Short: "Transform Model",
		Annotations: map[string]string{
			"docs:response":      "model\n  schema 1.1",
			"docs:response:lang": "fga",
		},
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	group.AddCommand(cmd)
	root.AddCommand(group)

	section := doc.GenerateCommandsSection(root)

	assert.Contains(t, section, "```fga\nmodel\n  schema 1.1\n```")
}

func TestAnchor_PunctuationStripped(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{"Read Relationship Tuple Changes (Watch)", "read-relationship-tuple-changes-watch"},
		{"Authorization Models", "authorization-models"},
		{"List Objects", "list-objects"},
	}

	for _, tc := range cases {
		section := doc.GenerateCommandsTOC(func() *cobra.Command {
			root := &cobra.Command{Use: "fga"}
			group := &cobra.Command{Use: "grp", Short: tc.input}
			group.AddCommand(&cobra.Command{
				Use:   "sub",
				Short: "Sub",
				RunE:  func(cmd *cobra.Command, args []string) error { return nil },
			})
			root.AddCommand(group)
			return root
		}())
		assert.Contains(t, section, "#"+tc.want, "anchor for %q", tc.input)
	}
}
