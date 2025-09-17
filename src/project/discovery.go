package tmvar

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// findProjectRoot discovers the project root by looking for go.mod or .git
func findProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}

	dir := currentDir
	for {
		// Check for go.mod
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Check for .git directory
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	// Fallback to current directory if no project markers found
	return currentDir, nil
}

// findTOMLFiles discovers TOML files in the project directory
func findTOMLFiles() ([]string, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, err
	}

	var tomlFiles []string

	err = filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for .toml extension
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".toml") {
			tomlFiles = append(tomlFiles, path)
		}

		return nil
	})

	return tomlFiles, err
}

// loadTOMLFile loads and parses a TOML file into a nested map
func loadTOMLFile(filename string) (map[string]interface{}, error) {
	var config map[string]interface{}

	_, err := toml.DecodeFile(filename, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TOML file %s: %v", filename, err)
	}

	return config, nil
}

// resolveKey looks up a value in the nested TOML structure using dot notation
func resolveKey(data map[string]interface{}, key string) (string, bool) {
	parts := strings.Split(key, ".")

	current := data
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - get the actual value
			if value, exists := current[part]; exists {
				// Convert to string
				return fmt.Sprintf("%v", value), true
			}
			return "", false
		} else {
			// Intermediate part - must be a map
			if nextMap, exists := current[part]; exists {
				if typedMap, ok := nextMap.(map[string]interface{}); ok {
					current = typedMap
				} else {
					// Path exists but is not a map
					return "", false
				}
			} else {
				// Path doesn't exist
				return "", false
			}
		}
	}

	return "", false
}

// findValueInFiles searches for a key across all TOML files
func findValueInFiles(key string) (string, error) {
	files, err := findTOMLFiles()
	if err != nil {
		return "", fmt.Errorf("failed to discover TOML files: %v", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("Error: Variable \"%s\" not found\n\nNo TOML files found in project", key)
	}

	var searchedFiles []string
	var availableKeys []string

	for _, file := range files {
		data, err := loadTOMLFile(file)
		if err != nil {
			continue // Skip files that can't be loaded
		}

		searchedFiles = append(searchedFiles, file)

		// Check if key exists in this file
		if value, found := resolveKey(data, key); found {
			return value, nil
		}

		// Collect available keys for error message
		collectKeys(data, "", &availableKeys)
	}

	// Generate helpful error message
	errorMsg := fmt.Sprintf("Error: Variable \"%s\" not found\n\nSearched in:", key)
	for _, file := range searchedFiles {
		errorMsg += fmt.Sprintf("\n- %s", file)
	}

	if len(availableKeys) > 0 {
		errorMsg += "\n\nAvailable variables:"
		for _, availableKey := range availableKeys {
			errorMsg += fmt.Sprintf("\n- %s", availableKey)
		}
	}

	return "", fmt.Errorf("%s", errorMsg)
}

// collectKeys recursively collects all available keys from a TOML structure
func collectKeys(data map[string]interface{}, prefix string, keys *[]string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if subMap, isMap := value.(map[string]interface{}); isMap {
			collectKeys(subMap, fullKey, keys)
		} else {
			*keys = append(*keys, fullKey)
		}
	}
}