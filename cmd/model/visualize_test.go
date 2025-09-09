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

package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetectFormatFromExtension tests the format detection logic from visualize.go.
func TestDetectFormatFromExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		outputFile     string
		originalFormat string
		expectedFormat string
	}{
		{"PNG extension", "output.png", "svg", "png"},
		{"SVG extension", "output.svg", "png", "svg"},
		{"Uppercase PNG", "OUTPUT.PNG", "svg", "png"},
		{"Mixed case SVG", "Output.Svg", "png", "svg"},
		{"No extension", "output", "svg", "svg"},
		{"Empty output file", "", "png", "png"},
		{"TXT extension", "output.txt", "png", "png"},
		{"Multiple dots with PNG", "my.output.file.png", "svg", "png"},
		{"Multiple dots with SVG", "my.output.file.svg", "png", "svg"},
		{"Dot in path but no extension", "path/to/output", "svg", "svg"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actualFormat := DetectFormatFromExtension(testCase.outputFile, testCase.originalFormat)

			if actualFormat != testCase.expectedFormat {
				t.Errorf("Expected format %s, got %s for filename %s", testCase.expectedFormat, actualFormat, testCase.outputFile)
			}
		})
	}
}

// TestGenerateOutputFileName tests the output filename generation logic from visualize.go.
func TestGenerateOutputFileName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		modelFile    string
		outputFile   string
		format       string
		expectedFile string
	}{
		{"Output file provided", "model.fga", "custom.svg", "png", "custom.svg"},
		{"No output file - simple FGA", "model.fga", "", "svg", "model.svg"},
		{"No output file - complex path", "path/to/model.fga", "", "png", "path/to/model.png"},
		{"No output file - no extension", "model", "", "svg", "model.svg"},
		{"No output file - multiple dots", "model.v1.0.fga", "", "png", "model.v1.0.png"},
		{"No output file - different extension", "model.yaml", "", "svg", "model.svg"},
		{"No output file - hidden file", ".model.fga", "", "png", ".model.png"},
		{"No output file - current dir", "./model.fga", "", "svg", "./model.svg"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actualFile := GenerateOutputFileName(testCase.modelFile, testCase.outputFile, testCase.format)

			if actualFile != testCase.expectedFile {
				t.Errorf("Expected output file %s, got %s", testCase.expectedFile, actualFile)
			}
		})
	}
}

// TestCalculateOutputParameters tests the complete parameter calculation logic from visualize.go.
func TestCalculateOutputParameters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		modelFile      string
		outputFile     string
		formatFlag     string
		expectedFormat string
		expectedOutput string
	}{
		{
			name:           "Default format SVG with no output file",
			modelFile:      "model.fga",
			outputFile:     "",
			formatFlag:     "svg",
			expectedFormat: "svg",
			expectedOutput: "model.svg",
		},
		{
			name:           "Default format PNG with no output file",
			modelFile:      "auth_model.fga",
			outputFile:     "",
			formatFlag:     "png",
			expectedFormat: "png",
			expectedOutput: "auth_model.png",
		},
		{
			name:           "Output file with PNG extension overrides format",
			modelFile:      "model.fga",
			outputFile:     "diagram.png",
			formatFlag:     "svg",
			expectedFormat: "png",
			expectedOutput: "diagram.png",
		},
		{
			name:           "Output file with SVG extension overrides format",
			modelFile:      "model.fga",
			outputFile:     "custom.svg",
			formatFlag:     "png",
			expectedFormat: "svg",
			expectedOutput: "custom.svg",
		},
		{
			name:           "Output file with non-image extension keeps original format",
			modelFile:      "model.fga",
			outputFile:     "output.txt",
			formatFlag:     "png",
			expectedFormat: "png",
			expectedOutput: "output.txt",
		},
		{
			name:           "Complex model path with directory",
			modelFile:      "models/complex_model.fga",
			outputFile:     "",
			formatFlag:     "svg",
			expectedFormat: "svg",
			expectedOutput: "models/complex_model.svg",
		},
		{
			name:           "Model with multiple dots in filename",
			modelFile:      "model.v1.0.fga",
			outputFile:     "",
			formatFlag:     "png",
			expectedFormat: "png",
			expectedOutput: "model.v1.0.png",
		},
		{
			name:           "Model without extension",
			modelFile:      "model",
			outputFile:     "",
			formatFlag:     "svg",
			expectedFormat: "svg",
			expectedOutput: "model.svg",
		},
		{
			name:           "Uppercase extension should be handled",
			modelFile:      "model.fga",
			outputFile:     "OUTPUT.PNG",
			formatFlag:     "svg",
			expectedFormat: "png",
			expectedOutput: "OUTPUT.PNG",
		},
		{
			name:           "Mixed case extension should be handled",
			modelFile:      "model.fga",
			outputFile:     "output.Svg",
			formatFlag:     "png",
			expectedFormat: "svg",
			expectedOutput: "output.Svg",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actualFormat, actualOutput := CalculateOutputParameters(testCase.modelFile, testCase.outputFile, testCase.formatFlag)

			if actualFormat != testCase.expectedFormat {
				t.Errorf("Expected format %s, got %s", testCase.expectedFormat, actualFormat)
			}

			if actualOutput != testCase.expectedOutput {
				t.Errorf("Expected output file %s, got %s", testCase.expectedOutput, actualOutput)
			}
		})
	}
}

// TestTemporaryFileCreation tests that we can create temp files for testing.
func TestTemporaryFileCreation(t *testing.T) {
	t.Parallel()

	// Create a temporary file to test with
	tmpDir := t.TempDir()
	modelFile := filepath.Join(tmpDir, "test_model.fga")

	// Create a simple model file content
	modelContent := `model
  schema 1.1

type user

type document
  relations
    define viewer: [user]`

	err := os.WriteFile(modelFile, []byte(modelContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test model file: %v", err)
	}

	// Test that the file exists
	if _, err := os.Stat(modelFile); os.IsNotExist(err) {
		t.Errorf("Test model file was not created")
	}

	// Test default parameter calculation with real file using the actual function
	format, outputFile := CalculateOutputParameters(modelFile, "", "svg")

	expectedOutput := filepath.Join(tmpDir, "test_model.svg")

	if outputFile != expectedOutput {
		t.Errorf("Expected output file %s, got %s", expectedOutput, outputFile)
	}

	if format != "svg" {
		t.Errorf("Expected format svg, got %s", format)
	}
}

// TestBasicModelVisualization tests the complete visualization workflow with the basic model fixture.
func TestBasicModelVisualization(t *testing.T) {
	t.Parallel()

	modelPath := "../../tests/fixtures/basic-model.fga"
	expectedSVGPath := "../../tests/fixtures/basic-model.svg"

	// Read the model file
	modelBytes, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("Failed to read model file: %v", err)
	}

	modelContent := string(modelBytes)

	// Transform DSL to weighted graph
	weightedGraph, err := TransformModelDSLToWeightedGraph(modelContent)
	if err != nil {
		t.Fatalf("Failed to transform model DSL to weighted graph: %v", err)
	}

	// Generate DOT format
	dotContent := ConvertToGraphvizDOT(weightedGraph)

	// Create a temporary file for the generated SVG
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "generated.svg")

	// Generate the SVG diagram
	err = GenerateDiagram(dotContent, outputFile, "svg")
	if err != nil {
		t.Fatalf("Failed to generate diagram: %v", err)
	}

	// Read the generated SVG
	generatedSVG, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated SVG: %v", err)
	}

	// Read the expected SVG
	expectedSVG, err := os.ReadFile(expectedSVGPath)
	if err != nil {
		t.Fatalf("Failed to read expected SVG: %v", err)
	}

	// Compare the SVG content (normalize line endings and whitespace)
	normalizedGenerated := normalizeXML(string(generatedSVG))
	normalizedExpected := normalizeXML(string(expectedSVG))

	if normalizedGenerated != normalizedExpected {
		t.Errorf("Generated SVG does not match expected SVG")
		// Write the generated content to a file for debugging
		debugFile := filepath.Join(tmpDir, "debug-generated.svg")
		if writeErr := os.WriteFile(debugFile, generatedSVG, 0600); writeErr != nil {
			t.Logf("Failed to write debug file: %v", writeErr)
		}

		t.Logf("Generated SVG written to: %s", debugFile)
		t.Logf("Expected SVG path: %s", expectedSVGPath)
	}
}

// normalizeXML normalizes XML content for comparison by removing
// version-specific comments and normalizing whitespace.
func normalizeXML(content string) string {
	lines := strings.Split(content, "\n")

	var normalized []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip version-specific comments that might change between graphviz versions
		if strings.Contains(trimmed, "Generated by graphviz version") {
			continue
		}

		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}

	return strings.Join(normalized, "\n")
}

// TestLoadModelContent tests the model content loading logic for both .fga and .fga.yaml files.
func TestLoadModelContent(t *testing.T) {
	t.Parallel()

	// Test regular .fga file
	t.Run("Regular FGA file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		fgaFile := filepath.Join(tmpDir, "test.fga")
		testModel := `
model
  schema 1.1

type user

type document
  relations
    define viewer: [user]
`
		err := os.WriteFile(fgaFile, []byte(testModel), 0600)
		if err != nil {
			t.Fatalf("Failed to write test FGA file: %v", err)
		}

		content, err := LoadModelContent(fgaFile)
		if err != nil {
			t.Fatalf("LoadModelContent failed: %v", err)
		}

		if content != testModel {
			t.Errorf("Content mismatch. Expected %q, got %q", testModel, content)
		}
	})

	// Test .fga.yaml file
	t.Run("FGA YAML file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		yamlFile := filepath.Join(tmpDir, "test.fga.yaml")

		yamlContent := `
name: test
model: |-
  model
    schema 1.1

  type user

  type document
    relations
      define viewer: [user]
tests: []
`
		err := os.WriteFile(yamlFile, []byte(yamlContent), 0600)
		if err != nil {
			t.Fatalf("Failed to write test YAML file: %v", err)
		}

		content, err := LoadModelContent(yamlFile)
		if err != nil {
			t.Fatalf("LoadModelContent failed: %v", err)
		}

		if !strings.Contains(content, "type user") || !strings.Contains(content, "type document") {
			t.Errorf("Content doesn't contain expected model elements. Got: %q", content)
		}
	})

	// Test non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		t.Parallel()

		_, err := LoadModelContent("/non/existent/file.fga")
		if err == nil {
			t.Fatal("Expected error for non-existent file, got nil")
		}
	})

	// Test .fga.yaml file without model
	t.Run("FGA YAML file without model", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		yamlFile := filepath.Join(tmpDir, "no-model.fga.yaml")

		yamlContent := `
name: test
tests: []
`
		err := os.WriteFile(yamlFile, []byte(yamlContent), 0600)
		if err != nil {
			t.Fatalf("Failed to write test YAML file: %v", err)
		}

		_, err = LoadModelContent(yamlFile)
		if err == nil {
			t.Fatal("Expected error for YAML file without model, got nil")
		}

		if !strings.Contains(err.Error(), "no model found") {
			t.Errorf("Expected 'no model found' error, got: %v", err)
		}
	})
}
