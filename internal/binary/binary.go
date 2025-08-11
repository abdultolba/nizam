package binary

import (
	"os/exec"
	"sync"
)

// ClientType represents a database client binary type
type ClientType string

const (
	PostgreSQL ClientType = "psql"
	MySQL      ClientType = "mysql"
	Redis      ClientType = "redis-cli"
	MongoDB    ClientType = "mongosh"
)

// cache stores the results of binary detection to avoid repeated LookPath calls
var (
	cache   = make(map[ClientType]bool)
	cacheMu sync.RWMutex
)

// HasBinary checks if a specific database client binary is available on the host system.
// It caches results to avoid repeated filesystem lookups.
func HasBinary(client ClientType) bool {
	cacheMu.RLock()
	if result, exists := cache[client]; exists {
		cacheMu.RUnlock()
		return result
	}
	cacheMu.RUnlock()

	// Check if binary exists in PATH
	_, err := exec.LookPath(string(client))
	result := err == nil

	// Cache the result
	cacheMu.Lock()
	cache[client] = result
	cacheMu.Unlock()

	return result
}

// ClearCache clears the binary detection cache.
// Useful for testing or when the PATH environment might have changed.
func ClearCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = make(map[ClientType]bool)
}

// GetBinaryPath returns the full path to the binary if it exists, empty string otherwise.
func GetBinaryPath(client ClientType) string {
	if !HasBinary(client) {
		return ""
	}
	
	path, err := exec.LookPath(string(client))
	if err != nil {
		return ""
	}
	return path
}
