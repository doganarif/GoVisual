package profiling

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/google/pprof/profile"
)

// FlameGraphNode represents a node in the flame graph
type FlameGraphNode struct {
	Name     string            `json:"name"`
	Value    int64             `json:"value"` // Time spent in nanoseconds
	Children []*FlameGraphNode `json:"children"`
}

// FlameGraph generates flame graph data from CPU profile
type FlameGraph struct {
	Root        *FlameGraphNode `json:"root"`
	TotalTime   int64           `json:"total_time"`
	SampleCount int64           `json:"sample_count"`
}

// GenerateFlameGraph generates a flame graph from CPU profile data
func GenerateFlameGraph(profileData []byte) (*FlameGraph, error) {
	if len(profileData) == 0 {
		return nil, fmt.Errorf("no profile data available")
	}

	// Parse the profile
	prof, err := profile.Parse(bytes.NewReader(profileData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	// Build the flame graph
	root := &FlameGraphNode{
		Name:     "root",
		Value:    0,
		Children: make([]*FlameGraphNode, 0),
	}

	// Process each sample in the profile
	for _, sample := range prof.Sample {
		value := sample.Value[1] // CPU nanoseconds
		if value == 0 {
			continue
		}

		// Build the stack trace path
		path := make([]string, 0, len(sample.Location))
		for i := len(sample.Location) - 1; i >= 0; i-- {
			loc := sample.Location[i]
			for j := len(loc.Line) - 1; j >= 0; j-- {
				line := loc.Line[j]
				funcName := line.Function.Name
				if funcName != "" {
					// Simplify function names
					funcName = simplifyFunctionName(funcName)
					path = append(path, funcName)
				}
			}
		}

		// Add to the tree
		addToTree(root, path, value)
	}

	// Calculate total time
	var totalTime int64
	for _, child := range root.Children {
		totalTime += child.Value
	}

	return &FlameGraph{
		Root:        root,
		TotalTime:   totalTime,
		SampleCount: int64(len(prof.Sample)),
	}, nil
}

// addToTree adds a stack trace path to the flame graph tree
func addToTree(root *FlameGraphNode, path []string, value int64) {
	current := root

	for _, name := range path {
		// Find or create child node
		var child *FlameGraphNode
		for _, c := range current.Children {
			if c.Name == name {
				child = c
				break
			}
		}

		if child == nil {
			child = &FlameGraphNode{
				Name:     name,
				Value:    0,
				Children: make([]*FlameGraphNode, 0),
			}
			current.Children = append(current.Children, child)
		}

		child.Value += value
		current = child
	}
}

// simplifyFunctionName simplifies a function name for display
func simplifyFunctionName(name string) string {
	// Remove parameter types for cleaner display
	if idx := strings.Index(name, "("); idx > 0 {
		name = name[:idx]
	}

	// Shorten package paths
	parts := strings.Split(name, "/")
	if len(parts) > 2 {
		// Keep last two parts of the path
		name = ".../" + strings.Join(parts[len(parts)-2:], "/")
	}

	return name
}

// GetHotSpots identifies the hottest code paths
func (fg *FlameGraph) GetHotSpots(threshold float64) []HotSpot {
	if fg.TotalTime == 0 {
		return nil
	}

	hotspots := make([]HotSpot, 0)
	findHotSpots(fg.Root, nil, fg.TotalTime, threshold, &hotspots)

	// Sort by percentage descending
	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].Percentage > hotspots[j].Percentage
	})

	// Return top 10
	if len(hotspots) > 10 {
		hotspots = hotspots[:10]
	}

	return hotspots
}

// HotSpot represents a performance hot spot
type HotSpot struct {
	Path       []string `json:"path"`
	Name       string   `json:"name"`
	Time       int64    `json:"time"`
	Percentage float64  `json:"percentage"`
}

// findHotSpots recursively finds hot spots in the flame graph
func findHotSpots(node *FlameGraphNode, path []string, totalTime int64, threshold float64, hotspots *[]HotSpot) {
	if node.Value == 0 {
		return
	}

	percentage := float64(node.Value) / float64(totalTime) * 100

	// Add current path
	currentPath := append(path, node.Name)

	// Check if this node is a hotspot
	if percentage >= threshold {
		*hotspots = append(*hotspots, HotSpot{
			Path:       currentPath,
			Name:       node.Name,
			Time:       node.Value,
			Percentage: percentage,
		})
	}

	// Recurse to children
	for _, child := range node.Children {
		findHotSpots(child, currentPath, totalTime, threshold, hotspots)
	}
}

// ConvertToD3Format converts the flame graph to D3.js compatible format
func (fg *FlameGraph) ConvertToD3Format() map[string]interface{} {
	return convertNodeToD3(fg.Root, fg.TotalTime)
}

// convertNodeToD3 converts a node to D3.js format
func convertNodeToD3(node *FlameGraphNode, totalTime int64) map[string]interface{} {
	d3Node := map[string]interface{}{
		"name":  node.Name,
		"value": node.Value,
	}

	if totalTime > 0 {
		d3Node["percentage"] = fmt.Sprintf("%.2f%%", float64(node.Value)/float64(totalTime)*100)
	}

	if len(node.Children) > 0 {
		children := make([]map[string]interface{}, len(node.Children))
		for i, child := range node.Children {
			children[i] = convertNodeToD3(child, totalTime)
		}
		d3Node["children"] = children
	}

	return d3Node
}

// GenerateTextFlameGraph generates a text-based flame graph (folded stack format)
func (fg *FlameGraph) GenerateTextFlameGraph() string {
	var buf bytes.Buffer
	generateTextNode(&buf, fg.Root, []string{})
	return buf.String()
}

// generateTextNode recursively generates text representation
func generateTextNode(buf *bytes.Buffer, node *FlameGraphNode, stack []string) {
	if node.Name != "root" {
		stack = append(stack, node.Name)
	}

	// Write current stack with value
	if len(stack) > 0 && node.Value > 0 {
		fmt.Fprintf(buf, "%s %d\n", strings.Join(stack, ";"), node.Value)
	}

	// Process children
	for _, child := range node.Children {
		generateTextNode(buf, child, stack)
	}
}
