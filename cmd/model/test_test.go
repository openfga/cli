package model

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     modelTestCmd.Use,
		Short:   modelTestCmd.Short,
		Long:    modelTestCmd.Long,
		Example: modelTestCmd.Example,
		RunE:    modelTestCmd.RunE,
	}
	c.Flags().String("tests", "", "path or glob of YAML/JSON test files")
	c.Flags().String("api-url", "http://localhost:8080", "api url")
	c.Flags().Bool("verbose", false, "Print verbose JSON output")
	c.Flags().Bool("suppress-summary", false, "Suppress the plain text summary output")
	_ = c.MarkFlagRequired("tests")
	return c
}

func writeTestFile(t *testing.T, dir, name, expect string) string {
	t.Helper()
	content := `name: Sample
model: |
  model
    schema 1.1

  type user
  type doc
    relations
      define viewer: [user]
tuples:
  - user: user:anne
    relation: viewer
    object: doc:1
tests:
  - name: check
    check:
      - user: user:anne
        object: doc:1
        assertions:
          viewer: ` + expect + "\n"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return path
}

type executeResult struct {
	out     string
	err     string
	execErr error
}

func executeCmd(t *testing.T, cmd *cobra.Command, args ...string) executeResult {
	t.Helper()
	bufOut := &bytes.Buffer{}
	bufErr := &bytes.Buffer{}
	cmd.SetOut(bufOut)
	cmd.SetErr(bufErr)
	cmd.SetArgs(args)
	execErr := cmd.Execute()
	return executeResult{bufOut.String(), bufErr.String(), execErr}
}

func TestModelTestCmdMissingFlag(t *testing.T) {
	cmd := newTestCmd()
	res := executeCmd(t, cmd)
	if res.execErr == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestModelTestCmdInvalidGlob(t *testing.T) {
	cmd := newTestCmd()
	res := executeCmd(t, cmd, "--tests", "[")
	if res.execErr == nil {
		t.Fatalf("expected error for invalid glob")
	}
}

func TestModelTestCmdZeroMatches(t *testing.T) {
	cmd := newTestCmd()
	res := executeCmd(t, cmd, "--tests", "no-such-file-*.yaml")
	if res.execErr == nil {
		t.Fatalf("expected error for missing file")
	}
}

func TestModelTestCmdSingleFilePass(t *testing.T) {
	dir := t.TempDir()
	file := writeTestFile(t, dir, "ok.fga.yaml", "true")
	cmd := newTestCmd()
	res := executeCmd(t, cmd, "--tests", file)
	if res.execErr != nil {
		t.Fatalf("unexpected error: %v", res.execErr)
	}
	if res.err == "" {
		t.Fatalf("expected summary on stderr")
	}
}

func TestModelTestCmdSingleFileFail(t *testing.T) {
	dir := t.TempDir()
	file := writeTestFile(t, dir, "fail.fga.yaml", "false")
	cmd := newTestCmd()
	res := executeCmd(t, cmd, "--tests", file)
	if res.execErr == nil {
		t.Fatalf("expected failure error")
	}
}

func TestModelTestCmdMultiFile(t *testing.T) {
	dir := t.TempDir()
	_ = writeTestFile(t, dir, "a.fga.yaml", "true")
	_ = writeTestFile(t, dir, "b.fga.yaml", "true")
	pattern := filepath.Join(dir, "*.fga.yaml")
	cmd := newTestCmd()
	res := executeCmd(t, cmd, "--tests", pattern)
	if res.execErr != nil {
		t.Fatalf("unexpected error: %v", res.execErr)
	}
	if count := strings.Count(res.err, "# "); count != 3 { // two files + summary
		t.Fatalf("expected per-file and summary output, got %q", res.err)
	}
}

func TestModelTestCmdVerboseSuppress(t *testing.T) {
	dir := t.TempDir()
	file := writeTestFile(t, dir, "ok.fga.yaml", "true")
	cmd := newTestCmd()
	res := executeCmd(t, cmd, "--tests", file, "--verbose", "--suppress-summary")
	if res.execErr != nil {
		t.Fatalf("unexpected error: %v", res.execErr)
	}
	if strings.Contains(res.err, "# Test Summary") {
		t.Fatalf("summary should be suppressed")
	}
}

func TestModelTestCmdContextCancel(t *testing.T) {
	dir := t.TempDir()
	file := writeTestFile(t, dir, "ok.fga.yaml", "true")
	cmd := newTestCmd()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	bufOut := &bytes.Buffer{}
	bufErr := &bytes.Buffer{}
	cmd.SetOut(bufOut)
	cmd.SetErr(bufErr)
	cmd.SetArgs([]string{"--tests", file})
	err := cmd.ExecuteContext(ctx)
	if err == nil {
		t.Fatalf("expected cancellation error")
	}
}
