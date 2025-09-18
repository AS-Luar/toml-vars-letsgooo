package tomv

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

// FileData represents a loaded TOML file with its resolved data
type FileData struct {
	Path          string
	Prefix        string
	Data          map[string]interface{}
	Resolved      map[string]interface{}
	resolvedValue string // Used for storing resolved value during search
}

// extractFilePrefix extracts filename prefix from path (app.toml -> app)
func extractFilePrefix(filePath string) string {
	filename := filepath.Base(filePath)
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	return name
}

// loadAllTOMLFiles loads all TOML files with namespaced architecture
func loadAllTOMLFiles() ([]FileData, error) {
	files, err := findTOMLFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to discover TOML files: %v", err)
	}

	var fileDataList []FileData
	namespacedData := make(map[string]interface{})

	// First pass: Load all files and namespace them
	for _, file := range files {
		data, err := loadTOMLFile(file)
		if err != nil {
			continue // Skip files that can't be loaded
		}

		prefix := extractFilePrefix(file)
		fileData := FileData{
			Path:   file,
			Prefix: prefix,
			Data:   data,
		}

		fileDataList = append(fileDataList, fileData)

		// Add to namespaced structure: namespacedData[filePrefix][section][key]
		namespacedData[prefix] = data
	}

	// Second pass: Resolve variables using namespaced structure
	resolvedNamespaced, err := resolveVariables(namespacedData)
	if err != nil {
		return nil, fmt.Errorf("error resolving cross-file variables: %v", err)
	}

	// Third pass: Extract resolved data back to individual files
	for i := range fileDataList {
		prefix := fileDataList[i].Prefix
		if resolvedFileData, exists := resolvedNamespaced[prefix]; exists {
			if resolvedMap, ok := resolvedFileData.(map[string]interface{}); ok {
				fileDataList[i].Resolved = resolvedMap
			} else {
				fileDataList[i].Resolved = fileDataList[i].Data
			}
		} else {
			fileDataList[i].Resolved = fileDataList[i].Data
		}
	}

	return fileDataList, nil
}

// findValueInFiles searches for a key with smart lookup and conflict detection
func findValueInFiles(key string) (string, error) {
	fileDataList, err := loadAllTOMLFiles()
	if err != nil {
		return "", err
	}

	if len(fileDataList) == 0 {
		return "", fmt.Errorf("variable \"%s\" not found\n\nNo TOML files found in project", key)
	}

	// Check if key uses explicit file prefix (filename.section.key)
	if strings.Contains(key, ".") {
		parts := strings.SplitN(key, ".", 2)
		if len(parts) == 2 {
			potentialFilePrefix := parts[0]
			remainingKey := parts[1]

			// Check if this is actually a file prefix
			var fileFound bool
			for _, fileData := range fileDataList {
				if fileData.Prefix == potentialFilePrefix {
					fileFound = true
					// This is explicit file syntax
					if value, found := resolveKey(fileData.Resolved, remainingKey); found {
						return value, nil
					}
					// Key not found in specified file
					return "", fmt.Errorf("variable \"%s\" not found in file %s\n\nAvailable variables in %s:\n%s",
						remainingKey, fileData.Path, fileData.Path, getFileVariablesList(fileData.Resolved))
				}
			}

			// If potentialFilePrefix looks like a file prefix but doesn't exist, and remainingKey contains dots
			if !fileFound && strings.Contains(remainingKey, ".") {
				// This looks like an explicit file syntax with invalid prefix
				return "", fmt.Errorf("file prefix \"%s\" not found\n\nAvailable file prefixes:\n%s",
					potentialFilePrefix, getAvailableFilePrefixes(fileDataList))
			}
			// Otherwise, continue with regular search (might be section.key format)
		}
	}

	// Smart lookup: Search for key across all files
	var foundFiles []FileData
	var searchedFiles []string
	var allAvailableKeys []string

	for _, fileData := range fileDataList {
		searchedFiles = append(searchedFiles, fileData.Path)

		if value, found := resolveKey(fileData.Resolved, key); found {
			foundFiles = append(foundFiles, fileData)
			// Store the actual value for potential return
			foundFiles[len(foundFiles)-1].resolvedValue = value
		}

		// Collect available keys for error message
		collectKeys(fileData.Resolved, "", &allAvailableKeys)
	}

	// Handle results based on how many files contain the key
	switch len(foundFiles) {
	case 0:
		// Variable not found in any file
		errorMsg := fmt.Sprintf("variable \"%s\" not found\n\nSearched in:", key)
		for _, file := range searchedFiles {
			errorMsg += fmt.Sprintf("\n- %s", file)
		}
		if len(allAvailableKeys) > 0 {
			errorMsg += "\n\nAvailable variables:\n" + formatVariablesList(allAvailableKeys)
		}
		return "", fmt.Errorf("%s", errorMsg)

	case 1:
		// Variable found in exactly one file - return it
		return foundFiles[0].resolvedValue, nil

	default:
		// Variable found in multiple files - conflict error with helpful message
		errorMsg := fmt.Sprintf("variable \"%s\" found in multiple files:", key)
		for _, fileData := range foundFiles {
			errorMsg += fmt.Sprintf("\n- %s", fileData.Path)
		}
		errorMsg += "\n\nUse explicit syntax:"
		for _, fileData := range foundFiles {
			errorMsg += fmt.Sprintf("\n- tomv.Get(\"%s.%s\")", fileData.Prefix, key)
		}
		return "", fmt.Errorf("%s", errorMsg)
	}
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

// getFileVariablesList returns a formatted list of variables in a specific file
func getFileVariablesList(data map[string]interface{}) string {
	var keys []string
	collectKeys(data, "", &keys)

	if len(keys) == 0 {
		return "- (no variables found)"
	}

	return formatVariablesList(keys)
}

// getAvailableFilePrefixes returns a formatted list of available file prefixes
func getAvailableFilePrefixes(fileDataList []FileData) string {
	result := ""
	for _, fileData := range fileDataList {
		result += fmt.Sprintf("- %s (from %s)\n", fileData.Prefix, fileData.Path)
	}
	return strings.TrimSuffix(result, "\n")
}

// formatVariablesList formats a list of variables for error messages
func formatVariablesList(keys []string) string {
	result := ""
	for _, key := range keys {
		result += "- " + key + "\n"
	}
	return strings.TrimSuffix(result, "\n")
}