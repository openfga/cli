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

	"github.com/spf13/cobra"

	"bytes"
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/openfga/cli/internal/output"

	"github.com/goccy/go-graphviz"
	"github.com/openfga/language/pkg/go/graph"
	language "github.com/openfga/language/pkg/go/transformer"
)

type Config struct {
	InputFile         string
	OutputFile        string
	ShowNodeType      bool
	ShowEdgeType      bool
	ShowNodeWeights   bool
	ShowEdgeWeights   bool
	ShowNodeWildcards bool
	ShowEdgeWildcards bool
	Format            string
}

type WeightedAuthorizationModelGraph struct {
	Edges map[string][]*WeightedAuthorizationModelEdge
	Nodes map[string]*WeightedAuthorizationModelNode
}

type WeightedAuthorizationModelEdge struct {
	Weights          map[string]int
	EdgeType         string
	TuplesetRelation string
	From             *WeightedAuthorizationModelNode
	To               *WeightedAuthorizationModelNode
	Wildcards        []string
	Conditions       []string
}

type WeightedAuthorizationModelNode struct {
	Weights     map[string]int
	NodeType    string
	Label       string
	UniqueLabel string
	Wildcards   []string
}

// visualizeCmd represents the visualize command.
var visualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "Visualize an Authorization Model",
	Long:  "Create a visual representation of an authorization model as an SVG or PNG image.",
	Example: `fga model visualize --model=model.fga
fga model visualize --model=model.fga --format=png
fga model visualize --model=model.fga --output-file=custom-name.svg
fga model visualize --model=model.fga --output-file=diagram.png`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Get command flags
		model, err := cmd.Flags().GetString("model")
		if err != nil {
			return fmt.Errorf("failed to get model parameter: %w", err)
		}

		outputFile, err := cmd.Flags().GetString("output-file")
		if err != nil {
			return fmt.Errorf("failed to get output-file parameter: %w", err)
		}

		format, err := cmd.Flags().GetString("format")
		if err != nil {
			return fmt.Errorf("failed to get format parameter: %w", err)
		}

		// If output-file is provided and has a valid extension, use that as the format
		if outputFile != "" {
			ext := strings.ToLower(filepath.Ext(outputFile))
			if ext == ".svg" || ext == ".png" {
				format = strings.TrimPrefix(ext, ".")
			}
		}

		// If output-file is not provided, derive it from the model filename and format
		if outputFile == "" {
			// Remove the extension from the model filename and add the format extension
			modelBase := strings.TrimSuffix(model, filepath.Ext(model))
			outputFile = modelBase + "." + format
		}

		// Read from file
		inputBytes, err := os.ReadFile(model)
		if err != nil {
			log.Fatalf("Error reading file %s: %v", model, err)
		}
		input := string(inputBytes)

		// Transform DSL to weighted graph
		weightedGraph, err := transformModelDSLToWeightedGraph(input)
		if err != nil {
			log.Fatalf("Error transforming model: %v", err)
		}

		// Generate DOT format
		dotContent := convertToGraphvizDOT(weightedGraph)

		// Generate diagram using Graphviz
		err = generateDiagram(dotContent, outputFile, format)
		if err != nil {
			log.Fatalf("Error generating diagram: %v", err)
		}

		fmt.Printf("Successfully generated diagram: %s\n", outputFile)

		return output.Display(output.EmptyStruct{})
	},
}

func transformModelDSLToWeightedGraph(dsl string) (*WeightedAuthorizationModelGraph, error) {
	authorizationModel, err := language.TransformDSLToProto(dsl)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization model DSL: %w", err)
	}

	wgb := graph.NewWeightedAuthorizationModelGraphBuilder()
	weightedGraph, err := wgb.Build(authorizationModel)
	if err != nil {
		return nil, fmt.Errorf("error building weighted graph: %w", err)
	}

	return translate(weightedGraph), nil
}

func translateNode(node *graph.WeightedAuthorizationModelNode) *WeightedAuthorizationModelNode {
	var nodeType string
	switch node.GetNodeType() {
	case graph.SpecificType:
		nodeType = "SpecificType"
	case graph.SpecificTypeAndRelation:
		nodeType = "SpecificTypeAndRelation"
	case graph.OperatorNode:
		nodeType = "OperatorNodeType"
	case graph.SpecificTypeWildcard:
		nodeType = "SpecificTypeWildcard"
	}
	return &WeightedAuthorizationModelNode{
		Weights:     node.GetWeights(),
		NodeType:    nodeType,
		Label:       node.GetLabel(),
		UniqueLabel: node.GetUniqueLabel(),
		Wildcards:   node.GetWildcards(),
	}
}

func translateEdge(e *graph.WeightedAuthorizationModelEdge) *WeightedAuthorizationModelEdge {
	var edgeType string
	switch e.GetEdgeType() {
	case graph.DirectEdge:
		edgeType = "Direct Edge"
	case graph.RewriteEdge:
		edgeType = "Rewrite Edge"
	case graph.TTUEdge:
		edgeType = "TTU Edge"
	case graph.ComputedEdge:
		edgeType = "Computed Edge"
	}

	return &WeightedAuthorizationModelEdge{
		Weights:          e.GetWeights(),
		EdgeType:         edgeType,
		TuplesetRelation: e.GetTuplesetRelation(),
		From:             translateNode(e.GetFrom()),
		To:               translateNode(e.GetTo()),
		Wildcards:        e.GetWildcards(),
		Conditions:       e.GetConditions(),
	}
}

func translate(weightedGraph *graph.WeightedAuthorizationModelGraph) *WeightedAuthorizationModelGraph {
	nodes := map[string]*WeightedAuthorizationModelNode{}
	for key, node := range weightedGraph.GetNodes() {
		nodes[key] = translateNode(node)
	}

	edges := map[string][]*WeightedAuthorizationModelEdge{}
	for key, edgeSlice := range weightedGraph.GetEdges() {
		transformedEdges := []*WeightedAuthorizationModelEdge{}
		for _, e := range edgeSlice {
			transformedEdges = append(transformedEdges, translateEdge(e))
		}
		edges[key] = transformedEdges
	}

	return &WeightedAuthorizationModelGraph{
		Nodes: nodes,
		Edges: edges,
	}
}

func convertToGraphvizDOT(graph *WeightedAuthorizationModelGraph) string {
	var sb strings.Builder

	sb.WriteString("digraph G {\n")
	sb.WriteString("  rankdir=TB;\n")
	sb.WriteString("  node [fontname=\"Arial\", fontsize=10];\n")
	sb.WriteString("  edge [fontname=\"Arial\", fontsize=8];\n")
	sb.WriteString("  graph [fontname=\"Arial\"];\n\n")

	// Track processed nodes to avoid duplicates
	processedNodes := make(map[string]bool)

	// Process edges and their connected nodes
	for _, edgeGroup := range graph.Edges {
		for _, edge := range edgeGroup {
			// Process From node
			if edge.From != nil && !processedNodes[edge.From.UniqueLabel] {
				sb.WriteString(generateNodeDOT(edge.From))
				processedNodes[edge.From.UniqueLabel] = true
			}

			// Process To node
			if edge.To != nil && !processedNodes[edge.To.UniqueLabel] {
				sb.WriteString(generateNodeDOT(edge.To))
				processedNodes[edge.To.UniqueLabel] = true
			}

			// Process edge
			if edge.From != nil && edge.To != nil {
				sb.WriteString(generateEdgeDOT(edge))
			}
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

func generateNodeDOT(node *WeightedAuthorizationModelNode) string {
	var labelParts []string

	// Node label (name)
	var label string
	if node.NodeType == "OperatorNodeType" {
		label = node.Label
	} else {
		label = node.UniqueLabel
	}
	labelParts = append(labelParts, fmt.Sprintf("<B>%s</B>", escapeHTML(label)))

	// Node type
	if node.NodeType != "" {
		labelParts = append(labelParts, escapeHTML(node.NodeType))
	}

	// Node weights
	if len(node.Weights) > 0 {
		var weightStrs []string
		for key, weight := range node.Weights {
			weightStr := strconv.Itoa(weight)
			if weight == 2147483647 { // Max int32, representing infinity
				weightStr = "∞"
			}
			weightStrs = append(weightStrs, fmt.Sprintf("<B>%s</B>: %s", escapeHTML(key), weightStr))
		}
		labelParts = append(labelParts, strings.Join(weightStrs, ", "))
	}

	// Node wildcards
	if len(node.Wildcards) > 0 {
		labelParts = append(labelParts, fmt.Sprintf("<B>Wildcards:</B> %s", escapeHTML(strings.Join(node.Wildcards, ", "))))
	}

	// Node attributes
	shape := "ellipse"
	var fillColor string
	var fontColor string = "black"

	// Only color nodes based on weight if they are NOT specific types or wildcards
	// Specific types (like user, organization, etc.) and wildcards should remain uncolored
	if node.NodeType == "SpecificType" || node.NodeType == "SpecificTypeWildcard" {
		// Specific types and wildcards get default coloring
		fillColor = "lightgray"
	} else {
		// Find the highest weight for coloring relations and other node types
		maxWeight := getMaxWeight(node.Weights)
		fillColor = getWeightColor(maxWeight)

		// Use white font for darker backgrounds
		if maxWeight > 10 {
			fontColor = "white"
		}
	}

	if node.NodeType == "OperatorNodeType" {
		shape = "box"
	}

	nodeLabel := strings.Join(labelParts, "<BR/>")
	return fmt.Sprintf("  \"%s\" [label=<%s>, shape=%s, fillcolor=\"%s\", style=filled, fontcolor=\"%s\"];\n",
		escapeNodeName(node.UniqueLabel), nodeLabel, shape, fillColor, fontColor)
}

func generateEdgeDOT(edge *WeightedAuthorizationModelEdge) string {
	var labelParts []string

	// Edge weights
	if len(edge.Weights) > 0 {
		var weightStrs []string
		for key, weight := range edge.Weights {
			weightStr := strconv.Itoa(weight)
			if weight == 2147483647 { // Max int32, representing infinity
				weightStr = "∞"
			}
			weightStrs = append(weightStrs, fmt.Sprintf("<B>%s</B>: %s", escapeHTML(key), weightStr))
		}
		labelParts = append(labelParts, strings.Join(weightStrs, ", "))
	}

	// Edge type
	if edge.EdgeType != "" {
		labelParts = append(labelParts, escapeHTML(edge.EdgeType))
	}

	// Edge wildcards
	if len(edge.Wildcards) > 0 {
		labelParts = append(labelParts, fmt.Sprintf("<B>Wildcards:</B> %s", escapeHTML(strings.Join(edge.Wildcards, ", "))))
	}

	edgeLabel := strings.Join(labelParts, "<BR/>")
	labelAttr := ""
	if edgeLabel != "" {
		labelAttr = fmt.Sprintf(" [label=<%s>]", edgeLabel)
	}

	return fmt.Sprintf("  \"%s\" -> \"%s\"%s;\n",
		escapeNodeName(edge.From.UniqueLabel), escapeNodeName(edge.To.UniqueLabel), labelAttr)
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func escapeNodeName(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

// getMaxWeight returns the highest weight value from a weights map
func getMaxWeight(weights map[string]int) int {
	if len(weights) == 0 {
		return 1 // Default weight if no weights are present
	}

	maxWeight := 0
	for _, weight := range weights {
		if weight > maxWeight && weight != 2147483647 { // Ignore infinity values for color calculation
			maxWeight = weight
		}
	}

	// If all weights are infinity or zero, return a reasonable default
	if maxWeight == 0 {
		for _, weight := range weights {
			if weight == 2147483647 {
				return 100 // Use high value for infinity in color calculation
			}
		}
		return 1 // Default to minimum weight
	}

	return maxWeight
}

// getWeightColor returns a color based on weight value
// Light green for low weights (1-2) to dark red for high weights (20+)
func getWeightColor(weight int) string {
	// Handle special cases
	if weight >= 100 { // Infinity or very high weights
		return "#8B0000" // Dark red
	}

	// Normalize weight to 0-1 range for color interpolation
	// Weights typically range from 1 to ~10 in most models
	normalizedWeight := float64(weight-1) / 9.0 // 1-10 maps to 0-1
	if normalizedWeight > 1.0 {
		normalizedWeight = 1.0
	}
	if normalizedWeight < 0.0 {
		normalizedWeight = 0.0
	}

	// Color interpolation from light green to dark red
	// Light green: #90EE90 (RGB: 144, 238, 144)
	// Dark red: #8B0000 (RGB: 139, 0, 0)

	// Calculate RGB values
	r := int(144 + (139-144)*normalizedWeight)
	g := int(238 + (0-238)*normalizedWeight)
	b := int(144 + (0-144)*normalizedWeight)

	// Ensure values are in valid range
	if r < 0 {
		r = 0
	} else if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	} else if g > 255 {
		g = 255
	}
	if b < 0 {
		b = 0
	} else if b > 255 {
		b = 255
	}

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func generateDiagram(dotContent, outputFile, format string) error {
	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputFile)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	ctx := context.Background()
	g, err := graphviz.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create graphviz instance: %w", err)
	}
	defer g.Close()

	graph, err := graphviz.ParseBytes([]byte(dotContent))
	if err != nil {
		return fmt.Errorf("failed to parse DOT content: %w", err)
	}
	defer graph.Close()

	// Determine output format
	var gvFormat graphviz.Format
	switch strings.ToLower(format) {
	case "png":
		gvFormat = graphviz.PNG
	case "svg":
		gvFormat = graphviz.SVG
	default:
		return fmt.Errorf("unsupported format: %s. Supported formats: png, svg", format)
	}

	// Render to buffer
	var buf bytes.Buffer
	if err := g.Render(ctx, graph, gvFormat, &buf); err != nil {
		return fmt.Errorf("failed to render %s: %w", format, err)
	}

	// Write to file
	return os.WriteFile(outputFile, buf.Bytes(), 0644)
}

func init() {
	visualizeCmd.Flags().String("model", "", "Authorization model file path")
	visualizeCmd.Flags().String("output-file", "", "Output file path for the visualization (defaults to model filename with format extension)")
	visualizeCmd.Flags().String("format", "svg", "Output format (svg or png)")

	if err := visualizeCmd.MarkFlagRequired("model"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/model/visualize", err)
	}
}
