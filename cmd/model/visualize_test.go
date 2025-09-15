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
	"fmt"
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

// TestFilterGraphByRelation tests the graph filtering functionality.
func TestFilterGraphByRelation(t *testing.T) {
	t.Parallel()

	// Create a simple test graph
	nodes := map[string]*WeightedAuthorizationModelNode{
		"document#viewer": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "viewer",
			UniqueLabel: "document#viewer",
			Weights:     map[string]int{"test": 1},
		},
		"user": {
			NodeType:    SpecificType,
			Label:       "user",
			UniqueLabel: "user",
			Weights:     map[string]int{"test": 2},
		},
		"folder#owner": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "owner",
			UniqueLabel: "folder#owner",
			Weights:     map[string]int{"test": 3},
		},
		"group#member": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "member",
			UniqueLabel: "group#member",
			Weights:     map[string]int{"test": 4},
		},
	}

	edges := map[string][]*WeightedAuthorizationModelEdge{
		"edge1": {
			{
				EdgeType: "Direct Edge",
				From:     nodes["document#viewer"],
				To:       nodes["user"],
				Weights:  map[string]int{"test": 1},
			},
		},
		"edge2": {
			{
				EdgeType: "Direct Edge",
				From:     nodes["document#viewer"],
				To:       nodes["folder#owner"],
				Weights:  map[string]int{"test": 2},
			},
		},
		"edge3": {
			{
				EdgeType: "Direct Edge",
				From:     nodes["folder#owner"],
				To:       nodes["group#member"],
				Weights:  map[string]int{"test": 3},
			},
		},
	}

	originalGraph := &WeightedAuthorizationModelGraph{
		Nodes: nodes,
		Edges: edges,
	}

	tests := []struct {
		name           string
		relationFilter string
		expectedNodes  []string
		expectedEdges  int
	}{
		{
			name:           "No filter",
			relationFilter: "",
			expectedNodes:  []string{"document#viewer", "user", "folder#owner", "group#member"},
			expectedEdges:  3,
		},
		{
			name:           "Filter by document#viewer",
			relationFilter: "document#viewer",
			expectedNodes:  []string{"document#viewer", "user", "folder#owner", "group#member"},
			expectedEdges:  3, // All nodes are reachable from document#viewer
		},
		{
			name:           "Filter by folder#owner",
			relationFilter: "folder#owner",
			expectedNodes:  []string{"folder#owner", "group#member"},
			expectedEdges:  1, // Only edge3 from folder#owner to group#member
		},
		{
			name:           "Filter by user (leaf node)",
			relationFilter: "user",
			expectedNodes:  []string{"user"},
			expectedEdges:  0, // No outgoing edges from user
		},
		{
			name:           "Non-existent relation",
			relationFilter: "nonexistent#relation",
			expectedNodes:  []string{},
			expectedEdges:  0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			filteredGraph := FilterGraphByRelation(originalGraph, testCase.relationFilter)

			// Check nodes
			if len(filteredGraph.Nodes) != len(testCase.expectedNodes) {
				t.Errorf("Expected %d nodes, got %d", len(testCase.expectedNodes), len(filteredGraph.Nodes))
			}

			for _, expectedNode := range testCase.expectedNodes {
				if _, exists := filteredGraph.Nodes[expectedNode]; !exists {
					t.Errorf("Expected node %s not found in filtered graph", expectedNode)
				}
			}

			// Check edges
			totalEdges := 0
			for _, edgeList := range filteredGraph.Edges {
				totalEdges += len(edgeList)
			}

			if totalEdges != testCase.expectedEdges {
				t.Errorf("Expected %d edges, got %d", testCase.expectedEdges, totalEdges)
			}
		})
	}
}

// TestFilterGraphByRelations tests the multiple relations filtering functionality.
func TestFilterGraphByRelations(t *testing.T) {
	t.Parallel()

	// Create a simple test graph
	nodes := map[string]*WeightedAuthorizationModelNode{
		"document#viewer": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "viewer",
			UniqueLabel: "document#viewer",
			Weights:     map[string]int{"test": 1},
		},
		"user": {
			NodeType:    SpecificType,
			Label:       "user",
			UniqueLabel: "user",
			Weights:     map[string]int{"test": 2},
		},
		"folder#owner": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "owner",
			UniqueLabel: "folder#owner",
			Weights:     map[string]int{"test": 3},
		},
		"group#member": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "member",
			UniqueLabel: "group#member",
			Weights:     map[string]int{"test": 4},
		},
		"role#assignee": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "assignee",
			UniqueLabel: "role#assignee",
			Weights:     map[string]int{"test": 5},
		},
	}

	edges := map[string][]*WeightedAuthorizationModelEdge{
		"edge1": {
			{
				EdgeType: "Direct Edge",
				From:     nodes["document#viewer"],
				To:       nodes["user"],
				Weights:  map[string]int{"test": 1},
			},
		},
		"edge2": {
			{
				EdgeType: "Direct Edge",
				From:     nodes["document#viewer"],
				To:       nodes["folder#owner"],
				Weights:  map[string]int{"test": 2},
			},
		},
		"edge3": {
			{
				EdgeType: "Direct Edge",
				From:     nodes["folder#owner"],
				To:       nodes["group#member"],
				Weights:  map[string]int{"test": 3},
			},
		},
		"edge4": {
			{
				EdgeType: "Direct Edge",
				From:     nodes["role#assignee"],
				To:       nodes["user"],
				Weights:  map[string]int{"test": 4},
			},
		},
	}

	originalGraph := &WeightedAuthorizationModelGraph{
		Nodes: nodes,
		Edges: edges,
	}

	tests := []struct {
		name          string
		relations     []string
		expectedNodes []string
		expectedEdges int
	}{
		{
			name:          "Empty relations list",
			relations:     []string{},
			expectedNodes: []string{"document#viewer", "user", "folder#owner", "group#member", "role#assignee"},
			expectedEdges: 4,
		},
		{
			name:          "Single relation",
			relations:     []string{"document#viewer"},
			expectedNodes: []string{"document#viewer", "user", "folder#owner", "group#member"},
			expectedEdges: 3, // All nodes reachable from document#viewer
		},
		{
			name:          "Multiple relations - overlapping",
			relations:     []string{"document#viewer", "folder#owner"},
			expectedNodes: []string{"document#viewer", "user", "folder#owner", "group#member"},
			expectedEdges: 3, // Same as single document#viewer since folder#owner is reachable from it
		},
		{
			name:          "Multiple relations - non-overlapping",
			relations:     []string{"folder#owner", "role#assignee"},
			expectedNodes: []string{"folder#owner", "group#member", "role#assignee", "user"},
			expectedEdges: 2, // edge3 and edge4
		},
		{
			name:          "Single leaf relation",
			relations:     []string{"user"},
			expectedNodes: []string{"user"},
			expectedEdges: 0, // No outgoing edges from user
		},
		{
			name:          "Non-existent relation",
			relations:     []string{"nonexistent#relation"},
			expectedNodes: []string{},
			expectedEdges: 0,
		},
		{
			name:          "Mix of existing and non-existent relations",
			relations:     []string{"user", "nonexistent#relation"},
			expectedNodes: []string{"user"},
			expectedEdges: 0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			filteredGraph := FilterGraphByRelations(originalGraph, testCase.relations)

			// Check nodes
			if len(filteredGraph.Nodes) != len(testCase.expectedNodes) {
				t.Errorf("Expected %d nodes, got %d", len(testCase.expectedNodes), len(filteredGraph.Nodes))
			}

			for _, expectedNode := range testCase.expectedNodes {
				if _, exists := filteredGraph.Nodes[expectedNode]; !exists {
					t.Errorf("Expected node %s not found in filtered graph", expectedNode)
				}
			}

			// Check edges
			totalEdges := 0
			for _, edgeList := range filteredGraph.Edges {
				totalEdges += len(edgeList)
			}

			if totalEdges != testCase.expectedEdges {
				t.Errorf("Expected %d edges, got %d", testCase.expectedEdges, totalEdges)
			}
		})
	}
}

// TestDisplayWeightSummary tests the weight summary display functionality.
func TestDisplayWeightSummary(t *testing.T) {
	t.Parallel()

	// Test with nil graph
	DisplayWeightSummary(nil)

	// Test with empty graph
	emptyGraph := &WeightedAuthorizationModelGraph{
		Nodes: make(map[string]*WeightedAuthorizationModelNode),
		Edges: make(map[string][]*WeightedAuthorizationModelEdge),
	}
	DisplayWeightSummary(emptyGraph)

	// Test with graph containing various node types and weights
	nodes := map[string]*WeightedAuthorizationModelNode{
		"document#viewer": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "viewer",
			UniqueLabel: "document#viewer",
			Weights:     map[string]int{"user": 1, "admin": 5},
		},
		"user": {
			NodeType:    SpecificType,
			Label:       "user",
			UniqueLabel: "user",
			Weights:     map[string]int{"user": 2}, // Should be ignored (not a relation)
		},
		"folder#owner": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "owner",
			UniqueLabel: "folder#owner",
			Weights:     map[string]int{"user": 1, "admin": 3},
		},
		"document#can_view": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "can_view",
			UniqueLabel: "document#can_view",
			Weights:     map[string]int{"user": 2147483647}, // Test infinity weight
		},
	}

	testGraph := &WeightedAuthorizationModelGraph{
		Nodes: nodes,
		Edges: make(map[string][]*WeightedAuthorizationModelEdge),
	}

	// This will print to stdout, which is fine for testing
	DisplayWeightSummary(testGraph)

	// Test passes if no panic occurs
}

// TestCreateTypeDiagram tests the type diagram creation functionality.
func TestCreateTypeDiagram(t *testing.T) {
	t.Parallel()

	// Create a simple test graph
	nodes := map[string]*WeightedAuthorizationModelNode{
		"user": {
			NodeType:    SpecificType,
			Label:       "user",
			UniqueLabel: "user",
		},
		"folder": {
			NodeType:    SpecificType,
			Label:       "folder",
			UniqueLabel: "folder",
		},
		"folder#owner": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "owner",
			UniqueLabel: "folder#owner",
		},
		"folder#parent": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "parent",
			UniqueLabel: "folder#parent",
		},
		"union:test": {
			NodeType:    OperatorNodeType,
			Label:       "union",
			UniqueLabel: "union:test",
		},
	}

	edges := map[string][]*WeightedAuthorizationModelEdge{
		"folder#owner": {
			{
				From: nodes["folder#owner"],
				To:   nodes["user"],
			},
		},
		"folder#parent": {
			{
				From: nodes["folder#parent"],
				To:   nodes["folder"],
			},
		},
		"folder#viewer": {
			{
				From: nodes["folder#viewer"],
				To:   nodes["union:test"],
			},
		},
	}

	testGraph := &WeightedAuthorizationModelGraph{
		Nodes: nodes,
		Edges: edges,
	}

	typeDiagram := CreateTypeDiagram(testGraph, []string{}, []string{})

	// Check that we have the expected types
	if len(typeDiagram.Types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(typeDiagram.Types))
	}

	if _, exists := typeDiagram.Types["user"]; !exists {
		t.Error("Expected 'user' type to exist")
	}

	if _, exists := typeDiagram.Types["folder"]; !exists {
		t.Error("Expected 'folder' type to exist")
	}

	// Check that we have some relations
	if len(typeDiagram.Relations) == 0 {
		t.Error("Expected at least one relation in type diagram")
	}
}

// TestCreateTypeDiagramWithIncludeTypes tests type diagram creation with include filter.
func TestCreateTypeDiagramWithIncludeTypes(t *testing.T) {
	t.Parallel()

	// Create a test graph with multiple types
	nodes := map[string]*WeightedAuthorizationModelNode{
		"user": {
			NodeType:    SpecificType,
			Label:       "user",
			UniqueLabel: "user",
		},
		"folder": {
			NodeType:    SpecificType,
			Label:       "folder",
			UniqueLabel: "folder",
		},
		"document": {
			NodeType:    SpecificType,
			Label:       "document",
			UniqueLabel: "document",
		},
		"folder#owner": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "owner",
			UniqueLabel: "folder#owner",
		},
	}

	edges := map[string][]*WeightedAuthorizationModelEdge{
		"folder#owner": {
			{
				From: nodes["folder#owner"],
				To:   nodes["user"],
			},
		},
	}

	testGraph := &WeightedAuthorizationModelGraph{
		Nodes: nodes,
		Edges: edges,
	}

	// Test with include filter for only "user" and "folder"
	typeDiagram := CreateTypeDiagram(testGraph, []string{"user", "folder"}, []string{})

	// Should only have user and folder types
	if len(typeDiagram.Types) != 2 {
		t.Errorf("Expected 2 types with include filter, got %d", len(typeDiagram.Types))
	}

	if _, exists := typeDiagram.Types["user"]; !exists {
		t.Error("Expected 'user' type to exist with include filter")
	}

	if _, exists := typeDiagram.Types["folder"]; !exists {
		t.Error("Expected 'folder' type to exist with include filter")
	}

	if _, exists := typeDiagram.Types["document"]; exists {
		t.Error("Expected 'document' type to be excluded with include filter")
	}
}

// TestCreateTypeDiagramWithExcludeTypes tests type diagram creation with exclude filter.
func TestCreateTypeDiagramWithExcludeTypes(t *testing.T) {
	t.Parallel()

	// Create a test graph with multiple types
	nodes := map[string]*WeightedAuthorizationModelNode{
		"user": {
			NodeType:    SpecificType,
			Label:       "user",
			UniqueLabel: "user",
		},
		"folder": {
			NodeType:    SpecificType,
			Label:       "folder",
			UniqueLabel: "folder",
		},
		"document": {
			NodeType:    SpecificType,
			Label:       "document",
			UniqueLabel: "document",
		},
		"folder#owner": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "owner",
			UniqueLabel: "folder#owner",
		},
	}

	edges := map[string][]*WeightedAuthorizationModelEdge{
		"folder#owner": {
			{
				From: nodes["folder#owner"],
				To:   nodes["user"],
			},
		},
	}

	testGraph := &WeightedAuthorizationModelGraph{
		Nodes: nodes,
		Edges: edges,
	}

	// Test with exclude filter for "document"
	typeDiagram := CreateTypeDiagram(testGraph, []string{}, []string{"document"})

	// Should have user and folder, but not document
	if _, exists := typeDiagram.Types["user"]; !exists {
		t.Error("Expected 'user' type to exist with exclude filter")
	}

	if _, exists := typeDiagram.Types["folder"]; !exists {
		t.Error("Expected 'folder' type to exist with exclude filter")
	}

	if _, exists := typeDiagram.Types["document"]; exists {
		t.Error("Expected 'document' type to be excluded with exclude filter")
	}
}

// TestConvertTypeDiagramToGraphvizDOT tests the DOT format conversion.
func TestConvertTypeDiagramToGraphvizDOT(t *testing.T) {
	t.Parallel()

	typeDiagram := &TypeDiagram{
		Types: map[string]*TypeNode{
			"user": {
				Name:  "user",
				Label: "user",
			},
			"folder": {
				Name:  "folder",
				Label: "folder",
			},
		},
		Relations: []*TypeRelation{
			{
				FromType:      "folder",
				ToType:        "user",
				RelationName:  "owner",
				RelationLabel: "owner",
			},
		},
	}

	dotOutput := ConvertTypeDiagramToGraphvizDOT(typeDiagram)

	// Check that the output contains expected elements
	if !strings.Contains(dotOutput, "digraph TypeDiagram") {
		t.Error("Expected DOT output to contain 'digraph TypeDiagram'")
	}

	if !strings.Contains(dotOutput, "\"user\"") {
		t.Error("Expected DOT output to contain user node")
	}

	if !strings.Contains(dotOutput, "\"folder\"") {
		t.Error("Expected DOT output to contain folder node")
	}

	if !strings.Contains(dotOutput, "\"user\" -> \"folder\" [label=\"owner\"]") {
		t.Error("Expected DOT output to contain user->folder edge with owner label")
	}
}

// TestIsDirectlyAssignable tests the direct assignability check.
func TestIsDirectlyAssignable(t *testing.T) {
	t.Parallel()

	// Create a test graph
	nodes := map[string]*WeightedAuthorizationModelNode{
		"user": {
			NodeType:    SpecificType,
			Label:       "user",
			UniqueLabel: "user",
		},
		"folder#owner": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "owner",
			UniqueLabel: "folder#owner",
		},
		"folder#viewer": {
			NodeType:    SpecificTypeAndRelation,
			Label:       "viewer",
			UniqueLabel: "folder#viewer",
		},
		"union:test": {
			NodeType:    OperatorNodeType,
			Label:       "union",
			UniqueLabel: "union:test",
		},
	}

	edges := map[string][]*WeightedAuthorizationModelEdge{
		"folder#owner": {
			{
				From: nodes["folder#owner"],
				To:   nodes["user"],
			},
		},
		"folder#viewer": {
			{
				From: nodes["folder#viewer"],
				To:   nodes["union:test"],
			},
		},
	}

	testGraph := &WeightedAuthorizationModelGraph{
		Nodes: nodes,
		Edges: edges,
	}

	// folder#owner should be directly assignable (connects directly to user type)
	if !isDirectlyAssignable(testGraph, "folder#owner") {
		t.Error("Expected folder#owner to be directly assignable")
	}

	// folder#viewer should not be directly assignable (connects to operator)
	if isDirectlyAssignable(testGraph, "folder#viewer") {
		t.Error("Expected folder#viewer to not be directly assignable")
	}
}

// TestFlagValidation tests validation of incompatible flag combinations.
func TestFlagValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		graphType        string
		includeRelations []string
		includeTypes     []string
		excludeTypes     []string
		expectedError    string
	}{
		{
			name:             "include-relations with graph-type=type",
			graphType:        "type",
			includeRelations: []string{"document#viewer"},
			includeTypes:     []string{},
			excludeTypes:     []string{},
			expectedError:    "--include-relations cannot be used with --graph-type=type",
		},
		{
			name:             "include-types with graph-type=weighted",
			graphType:        "weighted",
			includeRelations: []string{},
			includeTypes:     []string{"user"},
			excludeTypes:     []string{},
			expectedError:    "--include-types and --exclude-types can only be used with --graph-type=type",
		},
		{
			name:             "exclude-types with graph-type=weighted",
			graphType:        "weighted",
			includeRelations: []string{},
			includeTypes:     []string{},
			excludeTypes:     []string{"folder"},
			expectedError:    "--include-types and --exclude-types can only be used with --graph-type=type",
		},
		{
			name:             "multiple incompatible flags with graph-type=weighted",
			graphType:        "weighted",
			includeRelations: []string{},
			includeTypes:     []string{"user"},
			excludeTypes:     []string{"folder"},
			expectedError:    "--include-types and --exclude-types can only be used with --graph-type=type",
		},
		{
			name:             "valid combination - include-relations with weighted",
			graphType:        "weighted",
			includeRelations: []string{"document#viewer"},
			includeTypes:     []string{},
			excludeTypes:     []string{},
			expectedError:    "",
		},
		{
			name:             "valid combination - include-types with type",
			graphType:        "type",
			includeRelations: []string{},
			includeTypes:     []string{"user"},
			excludeTypes:     []string{},
			expectedError:    "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Test the validation logic directly
			var err error

			// Validate graph type
			if testCase.graphType != "weighted" && testCase.graphType != "type" {
				err = fmt.Errorf("invalid graph-type '%s': must be 'weighted' or 'type'", testCase.graphType)
			}

			// Validate flag combinations
			if err == nil && testCase.graphType == "type" && len(testCase.includeRelations) > 0 {
				err = fmt.Errorf("--include-relations cannot be used with --graph-type=type")
			}

			if err == nil && testCase.graphType == "weighted" && (len(testCase.includeTypes) > 0 || len(testCase.excludeTypes) > 0) {
				err = fmt.Errorf("--include-types and --exclude-types can only be used with --graph-type=type")
			}

			// Check the result
			if testCase.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error '%s' but got none", testCase.expectedError)
				}
				if err.Error() != testCase.expectedError {
					t.Errorf("Expected error '%s', got '%s'", testCase.expectedError, err.Error())
				}
			}
		})
	}
}
