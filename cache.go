package tomv

import (
	"os"
	"sync"
	"time"
)

type cacheEntry struct {
	value     string
	timestamp time.Time
	err       error
}

var (
	cache      = make(map[string]cacheEntry)
	cacheMutex = sync.RWMutex{}
	fileCache  = make(map[string]time.Time) // Track file modification times
)

// getValueFromCache retrieves a value with smart file monitoring
func getValueFromCache(key string) (string, error) {
	cacheMutex.RLock()

	// Check if we have a cached value and if files haven't changed
	if entry, exists := cache[key]; exists && !filesChanged() {
		cacheMutex.RUnlock()
		return entry.value, entry.err
	}

	cacheMutex.RUnlock()

	// Files changed or no cache - need to reload
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if entry, exists := cache[key]; exists && !filesChanged() {
		return entry.value, entry.err
	}

	// Update file timestamps and reload value
	updateFileTimestamps()
	value, err := findValueInFiles(key)

	// Cache the result
	cache[key] = cacheEntry{
		value:     value,
		timestamp: time.Now(),
		err:       err,
	}

	return value, err
}

// filesChanged checks if any TOML files have been modified since last check
func filesChanged() bool {
	files, err := findTOMLFiles()
	if err != nil {
		return true // Assume changed if we can't check
	}

	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			return true // File might have been deleted
		}

		lastModified := stat.ModTime()
		if cachedTime, exists := fileCache[file]; !exists || lastModified.After(cachedTime) {
			return true
		}
	}

	return false
}

// updateFileTimestamps updates our record of file modification times
func updateFileTimestamps() {
	files, err := findTOMLFiles()
	if err != nil {
		return
	}

	// Clear cache when files change
	cache = make(map[string]cacheEntry)

	// Update file timestamps
	for _, file := range files {
		stat, err := os.Stat(file)
		if err == nil {
			fileCache[file] = stat.ModTime()
		}
	}
}

// clearCache clears all cached values (useful for testing)
func clearCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	cache = make(map[string]cacheEntry)
	fileCache = make(map[string]time.Time)
}