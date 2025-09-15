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
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/goccy/go-graphviz"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/storetest"

	"github.com/openfga/language/pkg/go/graph"
	language "github.com/openfga/language/pkg/go/transformer"
)

const (
	// Node type constants.
	OperatorNodeType        = "OperatorNodeType"
	SpecificType            = "SpecificType"
	SpecificTypeAndRelation = "SpecificTypeAndRelation"
	SpecificTypeWildcard    = "SpecificTypeWildcard"
)

// Static error variables.
var (
	ErrNoModelInStoreData = errors.New("no model found in StoreData file")
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

// TypeDiagram represents a simplified graph with types as nodes and relations as edges.
type TypeDiagram struct {
	Types     map[string]*TypeNode
	Relations []*TypeRelation
}

// TypeNode represents a type in the type diagram.
type TypeNode struct {
	Name  string
	Label string
}

// TypeRelation represents a direct assignable relation between types.
type TypeRelation struct {
	FromType      string
	ToType        string
	RelationName  string
	RelationLabel string
}

// DetectFormatFromExtension determines the format based on the output file extension.
// If the extension is .svg or .png, it returns that format, otherwise returns the original format.
func DetectFormatFromExtension(outputFile, originalFormat string) string {
	if outputFile == "" {
		return originalFormat
	}

	ext := strings.ToLower(filepath.Ext(outputFile))
	if ext == ".svg" || ext == ".png" {
		return strings.TrimPrefix(ext, ".")
	}

	return originalFormat
}

// GenerateOutputFileName creates an output filename based on the model filename and format.
// If outputFile is provided, it returns that. Otherwise, it derives the name from the model file.
func GenerateOutputFileName(modelFile, outputFile, format string) string {
	if outputFile != "" {
		return outputFile
	}

	// Remove the extension from the model filename and add the format extension
	modelBase := strings.TrimSuffix(modelFile, filepath.Ext(modelFile))

	return modelBase + "." + format
}

// CalculateOutputParameters determines the final format and output filename based on inputs.
func CalculateOutputParameters(modelFile, outputFile, formatFlag string) (string, string) {
	// First, detect format from extension if output file is provided
	format := DetectFormatFromExtension(outputFile, formatFlag)

	// Then, generate the final output filename
	finalOutputFile := GenerateOutputFileName(modelFile, outputFile, format)

	return format, finalOutputFile
}

// LoadModelContent loads the model content from a file, handling both direct .fga files
// and .fga.yaml files that contain StoreData format.
func LoadModelContent(filePath string) (string, error) {
	// Check if this is a .fga.yaml file
	if strings.HasSuffix(strings.ToLower(filePath), ".fga.yaml") {
		return loadStoreDataModel(filePath)
	}

	// Load as regular file
	inputBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return string(inputBytes), nil
}

// loadStoreDataModel loads a model from a .fga.yaml store file.
func loadStoreDataModel(filePath string) (string, error) {
	// Load as StoreData format
	format, storeData, err := storetest.ReadFromFile(filePath, filepath.Dir(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to load StoreData from %s: %w", filePath, err)
	}

	if storeData.Model == "" {
		return "", fmt.Errorf("%w: %s", ErrNoModelInStoreData, filePath)
	}

	// Check if this is a modular model (format indicates modular and Model contains a file path)
	if format == authorizationmodel.ModelFormatModular {
		return convertModularModelToDSL(storeData.Model)
	}

	return storeData.Model, nil
}

// convertModularModelToDSL converts a modular model file to DSL format.
func convertModularModelToDSL(modFilePath string) (string, error) {
	// For modular models, storeData.Model contains the file path, not the content
	// We need to load the modular model and convert it to DSL
	model := &authorizationmodel.AuthzModel{}
	if err := model.ReadModelFromModFGA(modFilePath); err != nil {
		return "", fmt.Errorf("failed to read modular model from %s: %w", modFilePath, err)
	}

	// Convert the model to DSL format
	dslContent, err := model.DisplayAsDSL([]string{"model"})
	if err != nil {
		return "", fmt.Errorf("failed to convert modular model to DSL: %w", err)
	}

	return *dslContent, nil
}

// CreateTypeDiagram creates a simplified type diagram from a weighted graph.
func CreateTypeDiagram(graph *WeightedAuthorizationModelGraph, includeTypes []string, excludeTypes []string) *TypeDiagram {
	typeDiagram := &TypeDiagram{
		Types:     make(map[string]*TypeNode),
		Relations: []*TypeRelation{},
	}

	// Helper function to check if a type should be included
	shouldIncludeType := func(typeName string) bool {
		// If includeTypes is specified, only include types in that list
		if len(includeTypes) > 0 {
			found := false
			for _, includeType := range includeTypes {
				if typeName == includeType {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}

		// If excludeTypes is specified, exclude types in that list
		if len(excludeTypes) > 0 {
			for _, excludeType := range excludeTypes {
				if typeName == excludeType {
					return false
				}
			}
		}

		return true
	}

	// Collect all types (filtered) in deterministic order
	nodeKeys := make([]string, 0, len(graph.Nodes))
	for nodeKey := range graph.Nodes {
		nodeKeys = append(nodeKeys, nodeKey)
	}
	sort.Strings(nodeKeys)

	for _, nodeKey := range nodeKeys {
		node := graph.Nodes[nodeKey]
		if node.NodeType == SpecificType && shouldIncludeType(node.UniqueLabel) {
			typeDiagram.Types[node.UniqueLabel] = &TypeNode{
				Name:  node.UniqueLabel,
				Label: node.Label,
			}
		}
	}

	// Find direct assignable relations between types in deterministic order
	for _, nodeKey := range nodeKeys {
		node := graph.Nodes[nodeKey]
		if node.NodeType == SpecificTypeAndRelation {
			// Check if this relation is directly assignable
			if isDirectlyAssignable(graph, node.UniqueLabel) {
				// Find what types this relation connects to
				relationParts := strings.Split(node.UniqueLabel, "#")
				if len(relationParts) == 2 {
					fromType := relationParts[0]
					relationName := relationParts[1]

					// Only proceed if fromType should be included
					if !shouldIncludeType(fromType) {
						continue
					}

					// Find target types by following edges in deterministic order
					edgeKeys := make([]string, 0, len(graph.Edges))
					for edgeKey := range graph.Edges {
						edgeKeys = append(edgeKeys, edgeKey)
					}
					sort.Strings(edgeKeys)

					for _, edgeKey := range edgeKeys {
						edges := graph.Edges[edgeKey]
						for _, edge := range edges {
							if edge.From != nil && edge.From.UniqueLabel == node.UniqueLabel &&
								edge.To != nil && edge.To.NodeType == SpecificType {

								toType := edge.To.UniqueLabel

								// Only proceed if toType should be included
								if !shouldIncludeType(toType) {
									continue
								}

								// Add both types if they don't exist
								if _, exists := typeDiagram.Types[fromType]; !exists {
									typeDiagram.Types[fromType] = &TypeNode{
										Name:  fromType,
										Label: fromType,
									}
								}
								if _, exists := typeDiagram.Types[toType]; !exists {
									typeDiagram.Types[toType] = &TypeNode{
										Name:  toType,
										Label: toType,
									}
								}

								// Add the relation
								typeDiagram.Relations = append(typeDiagram.Relations, &TypeRelation{
									FromType:      fromType,
									ToType:        toType,
									RelationName:  relationName,
									RelationLabel: relationName,
								})
							}
						}
					}
				}
			}
		}
	}

	return typeDiagram
}

// ConvertTypeDiagramToGraphvizDOT converts a type diagram to Graphviz DOT format.
func ConvertTypeDiagramToGraphvizDOT(typeDiagram *TypeDiagram) string {
	var buffer bytes.Buffer

	buffer.WriteString("digraph TypeDiagram {\n")
	buffer.WriteString("  rankdir=TB;\n")
	buffer.WriteString("  node [shape=box, style=filled, fillcolor=lightblue];\n")
	buffer.WriteString("  edge [fontsize=10];\n\n")

	// Sort types by name for deterministic output
	typeNames := make([]string, 0, len(typeDiagram.Types))
	typeMap := make(map[string]*TypeNode)
	for _, typeNode := range typeDiagram.Types {
		typeNames = append(typeNames, typeNode.Name)
		typeMap[typeNode.Name] = typeNode
	}
	sort.Strings(typeNames)

	// Add type nodes in sorted order
	for _, typeName := range typeNames {
		typeNode := typeMap[typeName]
		buffer.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\"];\n",
			typeNode.Name, typeNode.Label))
	}

	buffer.WriteString("\n")

	// Sort relations for deterministic output
	sortedRelations := make([]*TypeRelation, len(typeDiagram.Relations))
	copy(sortedRelations, typeDiagram.Relations)
	sort.Slice(sortedRelations, func(i, j int) bool {
		if sortedRelations[i].FromType != sortedRelations[j].FromType {
			return sortedRelations[i].FromType < sortedRelations[j].FromType
		}
		if sortedRelations[i].ToType != sortedRelations[j].ToType {
			return sortedRelations[i].ToType < sortedRelations[j].ToType
		}
		return sortedRelations[i].RelationLabel < sortedRelations[j].RelationLabel
	})

	// Add relation edges in sorted order
	for _, relation := range sortedRelations {
		buffer.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n",
			relation.ToType, relation.FromType, relation.RelationLabel))
	}

	buffer.WriteString("}\n")
	return buffer.String()
}

// isDirectlyAssignable determines if a relation is directly assignable.
func isDirectlyAssignable(graph *WeightedAuthorizationModelGraph, relationKey string) bool {
	hasDirectTypeConnection := false
	hasOperatorConnection := false

	for _, edges := range graph.Edges {
		for _, edge := range edges {
			if edge.From != nil && edge.From.UniqueLabel == relationKey && edge.To != nil {
				targetNode := graph.Nodes[edge.To.UniqueLabel]
				if targetNode != nil {
					switch targetNode.NodeType {
					case SpecificType:
						hasDirectTypeConnection = true
					case OperatorNodeType:
						hasOperatorConnection = true
					}
				}
			}
		}
	}

	return hasDirectTypeConnection && !hasOperatorConnection
}

// FilterGraphByRelation filters the graph to show only nodes involved in evaluating a specific relation.
func FilterGraphByRelation(
	graph *WeightedAuthorizationModelGraph,
	relationKey string,
) *WeightedAuthorizationModelGraph {
	if relationKey == "" {
		return graph
	}

	// Find all nodes that are reachable from the target relation
	visitedNodes := findRelevantNodes(graph, relationKey)
	relevantEdges := findRelevantEdges(graph, visitedNodes)

	// Build filtered nodes map
	filteredNodes := make(map[string]*WeightedAuthorizationModelNode)

	for nodeKey := range visitedNodes {
		if node, exists := graph.Nodes[nodeKey]; exists {
			filteredNodes[nodeKey] = node
		}
	}

	return &WeightedAuthorizationModelGraph{
		Nodes: filteredNodes,
		Edges: relevantEdges,
	}
}

// findRelevantNodes performs DFS to find all nodes reachable from the target relation.
func findRelevantNodes(graph *WeightedAuthorizationModelGraph, relationKey string) map[string]bool {
	visitedNodes := make(map[string]bool)

	// Start DFS from the target relation node
	var dfs func(nodeKey string)

	dfs = func(nodeKey string) {
		if visitedNodes[nodeKey] {
			return
		}

		visitedNodes[nodeKey] = true

		// Find all edges that start from this node and traverse to their destinations
		for _, edges := range graph.Edges {
			for _, edge := range edges {
				if edge.From != nil && edge.From.UniqueLabel == nodeKey {
					// Continue DFS to the destination node
					if edge.To != nil {
						dfs(edge.To.UniqueLabel)
					}
				}
			}
		}
	}

	dfs(relationKey)

	return visitedNodes
}

// findRelevantEdges finds all edges that connect the relevant nodes.
func findRelevantEdges(
	graph *WeightedAuthorizationModelGraph,
	visitedNodes map[string]bool,
) map[string][]*WeightedAuthorizationModelEdge {
	relevantEdges := make(map[string][]*WeightedAuthorizationModelEdge)

	for edgeKey, edges := range graph.Edges {
		for _, edge := range edges {
			// Include edge if both From and To nodes are in our visited set
			if edge.From != nil && edge.To != nil &&
				visitedNodes[edge.From.UniqueLabel] && visitedNodes[edge.To.UniqueLabel] {
				if relevantEdges[edgeKey] == nil {
					relevantEdges[edgeKey] = []*WeightedAuthorizationModelEdge{}
				}

				relevantEdges[edgeKey] = append(relevantEdges[edgeKey], edge)
			}
		}
	}

	return relevantEdges
}

// FilterGraphByRelations filters the graph to show only nodes involved in evaluating any of the specified relations.
func FilterGraphByRelations(
	graph *WeightedAuthorizationModelGraph,
	relationKeys []string,
) *WeightedAuthorizationModelGraph {
	if len(relationKeys) == 0 {
		return graph
	}

	// Find all nodes that are reachable from any of the target relations
	allVisitedNodes := make(map[string]bool)

	for _, relationKey := range relationKeys {
		visitedNodes := findRelevantNodes(graph, relationKey)
		// Merge visited nodes from this relation into the overall set
		for nodeKey := range visitedNodes {
			allVisitedNodes[nodeKey] = true
		}
	}

	relevantEdges := findRelevantEdges(graph, allVisitedNodes)

	// Build filtered nodes map
	filteredNodes := make(map[string]*WeightedAuthorizationModelNode)

	for nodeKey := range allVisitedNodes {
		if node, exists := graph.Nodes[nodeKey]; exists {
			filteredNodes[nodeKey] = node
		}
	}

	return &WeightedAuthorizationModelGraph{
		Nodes: filteredNodes,
		Edges: relevantEdges,
	}
}

// DisplayWeightSummary displays a summary of weight distribution across all nodes in the graph.
func DisplayWeightSummary(graph *WeightedAuthorizationModelGraph) {
	if graph == nil || len(graph.Nodes) == 0 {
		return
	}

	// Count relations by weight across all weight types
	weightCounts := make(map[string]map[int]int) // weightType -> weight -> count

	for _, node := range graph.Nodes {
		// Only count nodes that represent relations (type#relation format)
		if node.NodeType == SpecificTypeAndRelation {
			for weightType, weight := range node.Weights {
				if weightCounts[weightType] == nil {
					weightCounts[weightType] = make(map[int]int)
				}
				weightCounts[weightType][weight]++
			}
		}
	}

	// Display the summary for each weight type in alphabetical order
	var weightTypes []string
	for weightType := range weightCounts {
		weightTypes = append(weightTypes, weightType)
	}

	// Sort weight types alphabetically
	for i := 0; i < len(weightTypes); i++ {
		for j := i + 1; j < len(weightTypes); j++ {
			if weightTypes[i] > weightTypes[j] {
				weightTypes[i], weightTypes[j] = weightTypes[j], weightTypes[i]
			}
		}
	}

	for _, weightType := range weightTypes {
		counts := weightCounts[weightType]
		fmt.Printf("\nWeight distribution (%s):\n", weightType)

		// Sort weights for consistent output
		var weights []int
		for weight := range counts {
			weights = append(weights, weight)
		}

		// Simple sort since Go doesn't have a built-in sort for small slices
		for i := 0; i < len(weights); i++ {
			for j := i + 1; j < len(weights); j++ {
				if weights[i] > weights[j] {
					weights[i], weights[j] = weights[j], weights[i]
				}
			}
		}

		for _, weight := range weights {
			count := counts[weight]
			if weight == 2147483647 {
				fmt.Printf("  weight ∞: %d relations\n", count)
			} else {
				fmt.Printf("  weight %d: %d relations\n", weight, count)
			}
		}
	}
}

// visualizeCmd represents the visualize command.
var visualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "Visualize an Authorization Model",
	Long: "Create a visual representation of an authorization model as an SVG or PNG image. " +
		"Supports both .fga model files and .fga.yaml store files.",
	Example: `fga model visualize --file=model.fga
fga model visualize --file=model.fga --format=png
fga model visualize --file=model.fga --output-file=custom-name.svg
fga model visualize --file=model.fga --output-file=diagram.png
fga model visualize --file=store.fga.yaml
fga model visualize --file=store.fga.yaml --format=png
fga model visualize --file=model.fga --include-relations="document#viewer"
fga model visualize --file=model.fga --include-relations="folder#owner" --format=png
fga model visualize --file=model.fga --include-relations="document#viewer,folder#owner"
fga model visualize --file=model.fga --graph-type=type
fga model visualize --file=model.fga --graph-type=type --include-types="user,document"
fga model visualize --file=model.fga --graph-type=type --exclude-types="folder"`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Get command flags
		model, err := cmd.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("failed to get file parameter: %w", err)
		}

		outputFile, err := cmd.Flags().GetString("output-file")
		if err != nil {
			return fmt.Errorf("failed to get output-file parameter: %w", err)
		}

		format, err := cmd.Flags().GetString("format")
		if err != nil {
			return fmt.Errorf("failed to get format parameter: %w", err)
		}

		includeRelations, err := cmd.Flags().GetStringSlice("include-relations")
		if err != nil {
			return fmt.Errorf("failed to get include-relations parameter: %w", err)
		}

		graphType, err := cmd.Flags().GetString("graph-type")
		if err != nil {
			return fmt.Errorf("failed to get graph-type parameter: %w", err)
		}

		includeTypes, err := cmd.Flags().GetStringSlice("include-types")
		if err != nil {
			return fmt.Errorf("failed to get include-types parameter: %w", err)
		}

		excludeTypes, err := cmd.Flags().GetStringSlice("exclude-types")
		if err != nil {
			return fmt.Errorf("failed to get exclude-types parameter: %w", err)
		}

		// Validate graph type
		if graphType != "weighted" && graphType != "type" {
			return fmt.Errorf("invalid graph-type '%s': must be 'weighted' or 'type'", graphType)
		}

		// Validate flag combinations
		if graphType == "type" && len(includeRelations) > 0 {
			return fmt.Errorf("--include-relations cannot be used with --graph-type=type")
		}

		if graphType == "weighted" && (len(includeTypes) > 0 || len(excludeTypes) > 0) {
			return fmt.Errorf("--include-types and --exclude-types can only be used with --graph-type=type")
		}

		// Calculate the final format and output filename
		format, outputFile = CalculateOutputParameters(model, outputFile, format)

		// Load model content (handles both .fga and .fga.yaml files)
		input, err := LoadModelContent(model)
		if err != nil {
			log.Fatalf("Error loading model: %v", err)
		}

		// Transform DSL to weighted graph
		weightedGraph, err := TransformModelDSLToWeightedGraph(input)
		if err != nil {
			log.Fatalf("Error transforming model: %v", err)
		}

		// Handle type diagram mode
		if graphType == "type" {
			typeDiagramGraph := CreateTypeDiagram(weightedGraph, includeTypes, excludeTypes)
			dotContent := ConvertTypeDiagramToGraphvizDOT(typeDiagramGraph)

			err = GenerateDiagram(dotContent, outputFile, format)
			if err != nil {
				log.Fatalf("Error generating diagram: %v", err)
			}

			fmt.Printf("Successfully generated type diagram: %s\n", outputFile)
			return output.Display(output.EmptyStruct{})
		}

		// Apply relation filter if specified
		if len(includeRelations) > 0 {
			weightedGraph = FilterGraphByRelations(weightedGraph, includeRelations)
		}

		// Generate DOT format
		dotContent := ConvertToGraphvizDOT(weightedGraph)

		// Generate diagram using Graphviz
		err = GenerateDiagram(dotContent, outputFile, format)
		if err != nil {
			log.Fatalf("Error generating diagram: %v", err)
		}

		// Display weight distribution summary
		DisplayWeightSummary(weightedGraph)

		fmt.Printf("Successfully generated diagram: %s\n", outputFile)

		return output.Display(output.EmptyStruct{})
	},
}

// TransformModelDSLToWeightedGraph transforms a DSL string into a weighted graph structure.
func TransformModelDSLToWeightedGraph(dsl string) (*WeightedAuthorizationModelGraph, error) {
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
		nodeType = SpecificType
	case graph.SpecificTypeAndRelation:
		nodeType = SpecificTypeAndRelation
	case graph.OperatorNode:
		nodeType = OperatorNodeType
	case graph.SpecificTypeWildcard:
		nodeType = SpecificTypeWildcard
	}

	return &WeightedAuthorizationModelNode{
		Weights:     node.GetWeights(),
		NodeType:    nodeType,
		Label:       node.GetLabel(),
		UniqueLabel: node.GetUniqueLabel(),
		Wildcards:   node.GetWildcards(),
	}
}

func translateEdge(edge *graph.WeightedAuthorizationModelEdge) *WeightedAuthorizationModelEdge {
	var edgeType string

	switch edge.GetEdgeType() {
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
		Weights:          edge.GetWeights(),
		EdgeType:         edgeType,
		TuplesetRelation: edge.GetTuplesetRelation(),
		From:             translateNode(edge.GetFrom()),
		To:               translateNode(edge.GetTo()),
		Wildcards:        edge.GetWildcards(),
		Conditions:       edge.GetConditions(),
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

// ConvertToGraphvizDOT converts a weighted graph to DOT format for graphviz.
func ConvertToGraphvizDOT(graph *WeightedAuthorizationModelGraph) string {
	var dotBuilder strings.Builder

	dotBuilder.WriteString("digraph G {\n")
	dotBuilder.WriteString("  rankdir=TB;\n")
	dotBuilder.WriteString("  node [fontname=\"Arial\", fontsize=10];\n")
	dotBuilder.WriteString("  edge [fontname=\"Arial\", fontsize=8];\n")
	dotBuilder.WriteString("  graph [fontname=\"Arial\"];\n\n")

	// Track processed nodes to avoid duplicates
	processedNodes := make(map[string]bool)

	// Get edge keys and sort them for deterministic output
	edgeKeys := make([]string, 0, len(graph.Edges))
	for edgeKey := range graph.Edges {
		edgeKeys = append(edgeKeys, edgeKey)
	}
	sort.Strings(edgeKeys)

	// Process edges in sorted order
	for _, edgeKey := range edgeKeys {
		edgeGroup := graph.Edges[edgeKey]
		for _, edge := range edgeGroup {
			// Process From node
			if edge.From != nil && !processedNodes[edge.From.UniqueLabel] {
				dotBuilder.WriteString(generateNodeDOT(edge.From))
				processedNodes[edge.From.UniqueLabel] = true
			}

			// Process To node
			if edge.To != nil && !processedNodes[edge.To.UniqueLabel] {
				dotBuilder.WriteString(generateNodeDOT(edge.To))
				processedNodes[edge.To.UniqueLabel] = true
			}

			// Process edge
			if edge.From != nil && edge.To != nil {
				dotBuilder.WriteString(generateEdgeDOT(edge))
			}
		}
	}

	dotBuilder.WriteString("}\n")

	return dotBuilder.String()
}

func generateNodeDOT(node *WeightedAuthorizationModelNode) string {
	nodeLabel := generateNodeLabel(node)
	shape, fillColor, fontColor := getNodeAttributes(node)

	return fmt.Sprintf("  \"%s\" [label=<%s>, shape=%s, fillcolor=\"%s\", style=filled, fontcolor=\"%s\"];\n",
		escapeNodeName(node.UniqueLabel), nodeLabel, shape, fillColor, fontColor)
}

func generateNodeLabel(node *WeightedAuthorizationModelNode) string {
	var labelParts []string

	// Node label (name)
	var label string
	if node.NodeType == OperatorNodeType {
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
		labelParts = append(labelParts, generateWeightLabel(node.Weights))
	}

	// Node wildcards
	if len(node.Wildcards) > 0 {
		labelParts = append(labelParts, "<B>Wildcards:</B> "+escapeHTML(strings.Join(node.Wildcards, ", ")))
	}

	return strings.Join(labelParts, "<BR/>")
}

func generateWeightLabel(weights map[string]int) string {
	weightStrs := make([]string, 0, len(weights))

	for key, weight := range weights {
		weightStr := strconv.Itoa(weight)
		if weight == 2147483647 { // Max int32, representing infinity
			weightStr = "∞"
		}

		weightStrs = append(weightStrs, fmt.Sprintf("<B>%s</B>: %s", escapeHTML(key), weightStr))
	}

	return strings.Join(weightStrs, ", ")
}

func getNodeAttributes(node *WeightedAuthorizationModelNode) (string, string, string) {
	// Default values
	shape := "ellipse"
	fontColor := "black"

	var fillColor string

	// Only color nodes based on weight if they are NOT specific types or wildcards
	if node.NodeType == SpecificType || node.NodeType == SpecificTypeWildcard {
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

	if node.NodeType == OperatorNodeType {
		shape = "box"
	}

	return shape, fillColor, fontColor
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
		labelParts = append(labelParts, "<B>Wildcards:</B> "+escapeHTML(strings.Join(edge.Wildcards, ", ")))
	}

	edgeLabel := strings.Join(labelParts, "<BR/>")

	labelAttr := ""
	if edgeLabel != "" {
		labelAttr = fmt.Sprintf(" [label=<%s>]", edgeLabel)
	}

	return fmt.Sprintf("  \"%s\" -> \"%s\"%s;\n",
		escapeNodeName(edge.From.UniqueLabel), escapeNodeName(edge.To.UniqueLabel), labelAttr)
}

func escapeHTML(content string) string {
	content = strings.ReplaceAll(content, "&", "&amp;")
	content = strings.ReplaceAll(content, "<", "&lt;")
	content = strings.ReplaceAll(content, ">", "&gt;")
	content = strings.ReplaceAll(content, "\"", "&quot;")

	return content
}

func escapeNodeName(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

// getMaxWeight returns the highest weight value from a weights map.
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
// Light green for low weights (1-2) to dark red for high weights (20+).
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
	red := int(144 + (139-144)*normalizedWeight)
	green := int(238 + (0-238)*normalizedWeight)
	blue := int(144 + (0-144)*normalizedWeight)

	// Ensure values are in valid range
	if red < 0 {
		red = 0
	} else if red > 255 {
		red = 255
	}

	if green < 0 {
		green = 0
	} else if green > 255 {
		green = 255
	}

	if blue < 0 {
		blue = 0
	} else if blue > 255 {
		blue = 255
	}

	return fmt.Sprintf("#%02X%02X%02X", red, green, blue)
}

// GenerateDiagram generates a visual diagram from DOT content in the specified format.
func GenerateDiagram(dotContent, outputFile, format string) error {
	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputFile)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	ctx := context.Background()

	graphInstance, err := graphviz.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create graphviz instance: %w", err)
	}
	defer graphInstance.Close()

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
		return fmt.Errorf( //nolint:err113
			"unsupported format: %s. Supported formats: png, svg", format)
	}

	// Render to buffer
	var buf bytes.Buffer
	if err := graphInstance.Render(ctx, graph, gvFormat, &buf); err != nil {
		return fmt.Errorf("failed to render %s: %w", format, err)
	}

	err = os.WriteFile(outputFile, buf.Bytes(), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write output file file: %w", err)
	}

	return nil
}

func init() {
	visualizeCmd.Flags().String("file", "", "Authorization model file path (.fga or .fga.yaml)")
	visualizeCmd.Flags().String("output-file", "", "Output file path for the visualization"+
		"(defaults to model filename with format extension)")
	visualizeCmd.Flags().String("format", "svg", "Output format (svg or png)")
	visualizeCmd.Flags().StringSlice("include-relations", []string{},
		"Filter graph to show only nodes involved in evaluating these relations (comma-separated list, format: type#relation)")
	visualizeCmd.Flags().String("graph-type", "weighted",
		"Type of graph to generate: 'weighted' (default, shows full model with weights) or 'type' (simplified diagram with types as nodes)")
	visualizeCmd.Flags().StringSlice("include-types", []string{},
		"Include only these types in the type diagram (comma-separated list, only applies to 'type' graph)")
	visualizeCmd.Flags().StringSlice("exclude-types", []string{},
		"Exclude these types from the type diagram (comma-separated list, only applies to 'type' graph)")

	if err := visualizeCmd.MarkFlagRequired("file"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/model/visualize", err)
	}
}
