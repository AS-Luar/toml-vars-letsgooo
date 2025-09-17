package tmvar

import (
	"fmt"
	"regexp"
	"strings"
)

// variablePattern matches {{section.key}} patterns
var variablePattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// resolveVariables processes a TOML data structure and resolves all {{variable}} references
func resolveVariables(data map[string]interface{}) (map[string]interface{}, error) {
	// Create a working copy to avoid modifying original data
	resolved := deepCopyMap(data)

	// Multi-pass resolution algorithm
	maxPasses := 10 // Prevent infinite loops
	for pass := 0; pass < maxPasses; pass++ {
		changed := false

		err := processMapForVariables(resolved, resolved, &changed)
		if err != nil {
			return nil, err
		}

		// If no changes were made, check if we still have unresolved variables
		if !changed {
			if hasUnresolvedVariables(resolved) {
				return nil, detectCircularDependencies(resolved)
			}
			break
		}

		// If we've reached max passes, check for circular dependencies
		if pass == maxPasses-1 {
			return nil, detectCircularDependencies(resolved)
		}
	}

	return resolved, nil
}

// processMapForVariables recursively processes all values in a map for variable substitution
func processMapForVariables(current map[string]interface{}, root map[string]interface{}, changed *bool) error {
	for key, value := range current {
		switch v := value.(type) {
		case string:
			// Process string values for variable substitution
			resolved, hasVariables, err := resolveStringVariables(v, root)
			if err != nil {
				return err
			}
			if hasVariables && resolved != v {
				current[key] = resolved
				*changed = true
			}
		case map[string]interface{}:
			// Recursively process nested maps
			err := processMapForVariables(v, root, changed)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// resolveStringVariables resolves all {{variable}} patterns in a string
func resolveStringVariables(str string, data map[string]interface{}) (string, bool, error) {
	hasVariables := false
	result := str

	// Find all {{variable}} matches
	matches := variablePattern.FindAllStringSubmatch(str, -1)

	for _, match := range matches {
		if len(match) != 2 {
			continue
		}

		hasVariables = true
		placeholder := match[0] // Full match: {{section.key}}
		variablePath := match[1] // Just the path: section.key

		// Resolve the variable
		value, found := resolveVariablePath(variablePath, data)
		if !found {
			return "", hasVariables, fmt.Errorf("Error: Variable '%s' referenced but not found\n\nAvailable variables:\n%s",
				variablePath, getAvailableVariablesList(data))
		}

		// Check if the resolved value still contains variables (for multi-pass)
		if strings.Contains(value, "{{") {
			// Still has unresolved variables - will be processed in next pass
			continue
		}

		// Replace the placeholder with the resolved value
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, hasVariables, nil
}

// resolveVariablePath resolves a dot-notation path like "section.key" to its value
func resolveVariablePath(path string, data map[string]interface{}) (string, bool) {
	parts := strings.Split(path, ".")

	current := data
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - get the actual value
			if value, exists := current[part]; exists {
				return fmt.Sprintf("%v", value), true
			}
			return "", false
		} else {
			// Intermediate part - must be a map
			if nextMap, exists := current[part]; exists {
				if typedMap, ok := nextMap.(map[string]interface{}); ok {
					current = typedMap
				} else {
					return "", false
				}
			} else {
				return "", false
			}
		}
	}

	return "", false
}

// detectCircularDependencies checks for circular dependencies after max passes
func detectCircularDependencies(data map[string]interface{}) error {
	// Build dependency graph
	dependencies := make(map[string][]string)
	collectDependencies(data, "", dependencies)

	// Check for cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for variable := range dependencies {
		if hasCycle(variable, dependencies, visited, recStack) {
			cycle := findCycle(variable, dependencies)
			return fmt.Errorf("Error: Circular dependency detected: %s", strings.Join(cycle, " → "))
		}
	}

	return fmt.Errorf("Error: Maximum resolution passes exceeded. Possible unresolvable variable references")
}

// collectDependencies builds a map of variable dependencies
func collectDependencies(data map[string]interface{}, prefix string, deps map[string][]string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			// Find all variable references in this string
			matches := variablePattern.FindAllStringSubmatch(v, -1)
			for _, match := range matches {
				if len(match) == 2 {
					referencedVar := match[1]
					deps[fullKey] = append(deps[fullKey], referencedVar)
				}
			}
		case map[string]interface{}:
			collectDependencies(v, fullKey, deps)
		}
	}
}

// hasCycle performs DFS to detect cycles in dependency graph
func hasCycle(node string, graph map[string][]string, visited, recStack map[string]bool) bool {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			if hasCycle(neighbor, graph, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[node] = false
	return false
}

// findCycle finds and returns a cycle path for error reporting
func findCycle(start string, graph map[string][]string) []string {
	// Simplified cycle detection for error reporting
	path := []string{start}
	current := start
	visited := make(map[string]bool)

	for len(graph[current]) > 0 && !visited[current] {
		visited[current] = true
		next := graph[current][0] // Take first dependency
		path = append(path, next)
		current = next

		// If we've seen this node before, we found a cycle
		for i, node := range path {
			if node == current && i < len(path)-1 {
				return path[i:]
			}
		}
	}

	return path
}

// getAvailableVariablesList returns a formatted list of available variables
func getAvailableVariablesList(data map[string]interface{}) string {
	var keys []string
	collectKeys(data, "", &keys)

	if len(keys) == 0 {
		return "- (no variables found)"
	}

	result := ""
	for _, key := range keys {
		result += "- " + key + "\n"
	}
	return strings.TrimSuffix(result, "\n")
}

// hasUnresolvedVariables checks if any string values still contain {{}} patterns
func hasUnresolvedVariables(data map[string]interface{}) bool {
	for _, value := range data {
		switch v := value.(type) {
		case string:
			if strings.Contains(v, "{{") {
				return true
			}
		case map[string]interface{}:
			if hasUnresolvedVariables(v) {
				return true
			}
		}
	}
	return false
}

// deepCopyMap creates a deep copy of a map[string]interface{}
func deepCopyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})

	for key, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[key] = deepCopyMap(v)
		default:
			copy[key] = v
		}
	}

	return copy
}