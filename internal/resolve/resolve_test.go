package resolve

import (
	"testing"

	"github.com/abdultolba/nizam/internal/config"
)

func TestDetermineEngine(t *testing.T) {
	tests := []struct {
		image    string
		service  string
		expected string
	}{
		{"postgres:16", "db", "postgres"},
		{"mysql:8", "database", "mysql"},
		{"redis:7", "cache", "redis"},
		{"mongo:7", "mongodb", "mongo"},
		{"mariadb:11", "db", "mysql"},
		{"unknown:1", "postgres-service", "postgres"},
		{"unknown:1", "mysql-db", "mysql"},
		{"unknown:1", "redis-cache", "redis"},
		{"unknown:1", "mongo-store", "mongo"},
		{"unknown:1", "unknown", "postgres"}, // default
	}

	for _, test := range tests {
		result := DetermineEngine(test.image, test.service)
		if result != test.expected {
			t.Errorf("DetermineEngine(%s, %s): expected %s, got %s",
				test.image, test.service, test.expected, result)
		}
	}
}

func TestGetServiceInfo(t *testing.T) {
	cfg := &config.Config{
		Profile: "test",
		Services: map[string]config.Service{
			"postgres": {
				Image: "postgres:16",
				Ports: []string{"5432:5432"},
				Environment: map[string]string{
					"POSTGRES_USER":     "testuser",
					"POSTGRES_PASSWORD": "testpass",
					"POSTGRES_DB":       "testdb",
				},
			},
			"redis": {
				Image:       "redis:7",
				Ports:       []string{"6379:6379"},
				Environment: map[string]string{},
			},
		},
	}

	// Test PostgreSQL service
	info, err := GetServiceInfo(cfg, "postgres")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if info.Name != "postgres" {
		t.Errorf("Expected name postgres, got %s", info.Name)
	}
	if info.Engine != "postgres" {
		t.Errorf("Expected engine postgres, got %s", info.Engine)
	}
	if info.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", info.Port)
	}
	if info.User != "testuser" {
		t.Errorf("Expected user testuser, got %s", info.User)
	}
	if info.Database != "testdb" {
		t.Errorf("Expected database testdb, got %s", info.Database)
	}
	if info.Container != "nizam_postgres" {
		t.Errorf("Expected container nizam_postgres, got %s", info.Container)
	}

	// Test Redis service
	info, err = GetServiceInfo(cfg, "redis")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if info.Engine != "redis" {
		t.Errorf("Expected engine redis, got %s", info.Engine)
	}
	if info.Port != 6379 {
		t.Errorf("Expected port 6379, got %d", info.Port)
	}

	// Test nonexistent service
	_, err = GetServiceInfo(cfg, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent service")
	}
}

func TestGetConnectionString(t *testing.T) {
	tests := []struct {
		info     ServiceInfo
		expected string
	}{
		{
			ServiceInfo{
				Engine:   "postgres",
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				Database: "db",
			},
			"postgresql://user:pass@localhost:5432/db?sslmode=disable",
		},
		{
			ServiceInfo{
				Engine:   "redis",
				Host:     "localhost",
				Port:     6379,
				Password: "",
			},
			"redis://localhost:6379",
		},
		{
			ServiceInfo{
				Engine:   "redis",
				Host:     "localhost",
				Port:     6379,
				Password: "secret",
			},
			"redis://:secret@localhost:6379",
		},
	}

	for _, test := range tests {
		result := test.info.GetConnectionString()
		if result != test.expected {
			t.Errorf("GetConnectionString(): expected %s, got %s", test.expected, result)
		}
	}
}
