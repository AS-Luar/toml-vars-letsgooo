package main

import (
	"fmt"
	"github.com/yourusername/toml-vars-letsgooo/src/project"
)

func main() {
	fmt.Println("=== tmvar Phase 1 Demo ===")

	// Test basic value retrieval
	fmt.Printf("Server Host: %s\n", tmvar.Get("server.host"))
	fmt.Printf("Server Port: %d\n", tmvar.GetInt("server.port"))
	fmt.Printf("Debug Mode: %t\n", tmvar.GetBool("server.debug"))

	// Test GetOr functions
	fmt.Printf("API Timeout (with default): %d\n", tmvar.GetIntOr("api.timeout", 30))
	fmt.Printf("Cache TTL (with default): %s\n", tmvar.GetOr("cache.ttl", "5m"))

	// Test Exists function
	fmt.Printf("Database URL exists: %t\n", tmvar.Exists("database.url"))
	fmt.Printf("Missing key exists: %t\n", tmvar.Exists("missing.key"))

	fmt.Println("\n=== Phase 1 Complete! ===")
}