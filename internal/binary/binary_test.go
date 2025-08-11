package binary

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHasBinary(t *testing.T) {
	// Clear cache before tests
	ClearCache()

	tests := []struct {
		name     string
		client   ClientType
		wantFunc func() bool // Function to determine expected result
	}{
		{
			name:   "PostgreSQL psql",
			client: PostgreSQL,
			wantFunc: func() bool {
				// Check if psql is actually available
				path := os.Getenv("PATH")
				for _, dir := range strings.Split(path, string(os.PathListSeparator)) {
					if _, err := os.Stat(filepath.Join(dir, "psql")); err == nil {
						return true
					}
				}
				return false
			},
		},
		{
			name:   "MySQL mysql",
			client: MySQL,
			wantFunc: func() bool {
				path := os.Getenv("PATH")
				for _, dir := range strings.Split(path, string(os.PathListSeparator)) {
					if _, err := os.Stat(filepath.Join(dir, "mysql")); err == nil {
						return true
					}
				}
				return false
			},
		},
		{
			name:   "Redis redis-cli",
			client: Redis,
			wantFunc: func() bool {
				path := os.Getenv("PATH")
				for _, dir := range strings.Split(path, string(os.PathListSeparator)) {
					if _, err := os.Stat(filepath.Join(dir, "redis-cli")); err == nil {
						return true
					}
				}
				return false
			},
		},
		{
			name:   "MongoDB mongosh",
			client: MongoDB,
			wantFunc: func() bool {
				path := os.Getenv("PATH")
				for _, dir := range strings.Split(path, string(os.PathListSeparator)) {
					if _, err := os.Stat(filepath.Join(dir, "mongosh")); err == nil {
						return true
					}
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.wantFunc()
			got := HasBinary(tt.client)
			if got != want {
				t.Errorf("HasBinary(%v) = %v, want %v", tt.client, got, want)
			}
		})
	}
}

func TestHasBinaryCache(t *testing.T) {
	ClearCache()

	// First call should populate cache
	client := PostgreSQL
	result1 := HasBinary(client)
	
	// Second call should use cache (same result)
	result2 := HasBinary(client)
	
	if result1 != result2 {
		t.Errorf("Cache inconsistency: first call = %v, second call = %v", result1, result2)
	}

	// Verify cache contains the result
	cacheMu.RLock()
	cachedResult, exists := cache[client]
	cacheMu.RUnlock()
	
	if !exists {
		t.Error("Expected result to be cached")
	}
	
	if cachedResult != result1 {
		t.Errorf("Cached result = %v, expected %v", cachedResult, result1)
	}
}

func TestClearCache(t *testing.T) {
	ClearCache()
	
	// Populate cache
	HasBinary(PostgreSQL)
	HasBinary(MySQL)
	
	// Verify cache has entries
	cacheMu.RLock()
	cacheSize := len(cache)
	cacheMu.RUnlock()
	
	if cacheSize == 0 {
		t.Error("Expected cache to have entries")
	}
	
	// Clear cache
	ClearCache()
	
	// Verify cache is empty
	cacheMu.RLock()
	cacheSize = len(cache)
	cacheMu.RUnlock()
	
	if cacheSize != 0 {
		t.Errorf("Expected cache to be empty, got %d entries", cacheSize)
	}
}

func TestGetBinaryPath(t *testing.T) {
	ClearCache()
	
	tests := []struct {
		name   string
		client ClientType
	}{
		{"PostgreSQL", PostgreSQL},
		{"MySQL", MySQL},
		{"Redis", Redis},
		{"MongoDB", MongoDB},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := GetBinaryPath(tt.client)
			hasBinary := HasBinary(tt.client)
			
			if hasBinary && path == "" {
				t.Errorf("HasBinary(%v) = true but GetBinaryPath(%v) = empty", tt.client, tt.client)
			}
			
			if !hasBinary && path != "" {
				t.Errorf("HasBinary(%v) = false but GetBinaryPath(%v) = %q", tt.client, tt.client, path)
			}
		})
	}
}

func TestClientTypeString(t *testing.T) {
	tests := []struct {
		client   ClientType
		expected string
	}{
		{PostgreSQL, "psql"},
		{MySQL, "mysql"},
		{Redis, "redis-cli"},
		{MongoDB, "mongosh"},
	}

	for _, tt := range tests {
		if string(tt.client) != tt.expected {
			t.Errorf("ClientType %v string = %q, want %q", tt.client, string(tt.client), tt.expected)
		}
	}
}
