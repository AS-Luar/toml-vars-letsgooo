package tomv

import (
	"strconv"
	"strings"
	"time"
)

// Get retrieves a string value by key, panics if not found
func Get(key string) string {
	value, err := getValue(key)
	if err != nil {
		panic(err)
	}
	return value
}

// GetInt retrieves an integer value by key, panics if not found or invalid
func GetInt(key string) int {
	value := Get(key)
	result, err := strconv.Atoi(value)
	if err != nil {
		panic("variable \"" + key + "\" is not a valid integer: " + value)
	}
	return result
}

// GetBool retrieves a boolean value by key, panics if not found or invalid
func GetBool(key string) bool {
	value := Get(key)
	switch value {
	case "true", "True", "TRUE", "yes", "Yes", "YES", "1":
		return true
	case "false", "False", "FALSE", "no", "No", "NO", "0":
		return false
	default:
		panic("variable \"" + key + "\" is not a valid boolean: " + value)
	}
}

// GetFloat retrieves a float64 value by key, panics if not found or invalid
func GetFloat(key string) float64 {
	value := Get(key)
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		panic("variable \"" + key + "\" is not a valid float: " + value)
	}
	return result
}

// GetDuration retrieves a time.Duration value by key, panics if not found or invalid
func GetDuration(key string) time.Duration {
	value := Get(key)
	result, err := time.ParseDuration(value)
	if err != nil {
		panic("variable \"" + key + "\" is not a valid duration: " + value)
	}
	return result
}

// GetStringSlice retrieves a comma-separated string value as a slice, panics if not found
func GetStringSlice(key string) []string {
	value := Get(key)
	if value == "" {
		return []string{}
	}

	// Split on comma and trim whitespace
	parts := strings.Split(value, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}
	return result
}

// GetIntSlice retrieves a comma-separated string value as an int slice, panics if not found or invalid
func GetIntSlice(key string) []int {
	stringSlice := GetStringSlice(key)
	result := make([]int, len(stringSlice))

	for i, str := range stringSlice {
		if str == "" {
			continue // Skip empty strings
		}
		num, err := strconv.Atoi(str)
		if err != nil {
			panic("variable \"" + key + "\" contains invalid integer: " + str)
		}
		result[i] = num
	}
	return result
}

// GetOr retrieves a string value by key, returns default if not found
func GetOr(key string, defaultValue string) string {
	value, err := getValue(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetIntOr retrieves an integer value by key, returns default if not found or invalid
func GetIntOr(key string, defaultValue int) int {
	value, err := getValue(key)
	if err != nil {
		return defaultValue
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return result
}

// GetBoolOr retrieves a boolean value by key, returns default if not found or invalid
func GetBoolOr(key string, defaultValue bool) bool {
	value, err := getValue(key)
	if err != nil {
		return defaultValue
	}
	switch value {
	case "true", "True", "TRUE", "yes", "Yes", "YES", "1":
		return true
	case "false", "False", "FALSE", "no", "No", "NO", "0":
		return false
	default:
		return defaultValue
	}
}

// GetFloatOr retrieves a float64 value by key, returns default if not found or invalid
func GetFloatOr(key string, defaultValue float64) float64 {
	value, err := getValue(key)
	if err != nil {
		return defaultValue
	}
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return result
}

// GetDurationOr retrieves a time.Duration value by key, returns default if not found or invalid
func GetDurationOr(key string, defaultValue time.Duration) time.Duration {
	value, err := getValue(key)
	if err != nil {
		return defaultValue
	}
	result, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return result
}

// GetStringSliceOr retrieves a comma-separated string value as a slice, returns default if not found
func GetStringSliceOr(key string, defaultValue []string) []string {
	value, err := getValue(key)
	if err != nil {
		return defaultValue
	}
	if value == "" {
		return []string{}
	}

	// Split on comma and trim whitespace
	parts := strings.Split(value, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}
	return result
}

// GetIntSliceOr retrieves a comma-separated string value as an int slice, returns default if not found or invalid
func GetIntSliceOr(key string, defaultValue []int) []int {
	value, err := getValue(key)
	if err != nil {
		return defaultValue
	}
	if value == "" {
		return []int{}
	}

	// Split on comma and trim whitespace
	parts := strings.Split(value, ",")
	result := make([]int, len(parts))

	for i, str := range parts {
		str = strings.TrimSpace(str)
		if str == "" {
			continue // Skip empty strings
		}
		num, err := strconv.Atoi(str)
		if err != nil {
			return defaultValue // Return default if any conversion fails
		}
		result[i] = num
	}
	return result
}

// Exists checks if a variable exists without retrieving its value
func Exists(key string) bool {
	_, err := getValue(key)
	return err == nil
}

// getValue is the internal function that handles the actual value retrieval
// This will be implemented by discovery.go and cache.go
func getValue(key string) (string, error) {
	return getValueFromCache(key)
}
