package snapshot

import (
	"testing"

	"github.com/abdultolba/nizam/internal/dockerx"
)

func TestMySQLEngine_GetEngineType(t *testing.T) {
	docker := &dockerx.Client{} // Mock client
	engine := NewMySQLEngine(docker)

	if got := engine.GetEngineType(); got != "mysql" {
		t.Errorf("GetEngineType() = %v, want %v", got, "mysql")
	}
}

func TestMySQLEngine_CanHandle(t *testing.T) {
	docker := &dockerx.Client{} // Mock client
	engine := NewMySQLEngine(docker)

	tests := []struct {
		name     string
		engine   string
		expected bool
	}{
		{
			name:     "handles mysql",
			engine:   "mysql",
			expected: true,
		},
		{
			name:     "handles mariadb",
			engine:   "mariadb",
			expected: true,
		},
		{
			name:     "does not handle postgres",
			engine:   "postgres",
			expected: false,
		},
		{
			name:     "does not handle redis",
			engine:   "redis",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := engine.CanHandle(tt.engine); got != tt.expected {
				t.Errorf("CanHandle(%v) = %v, want %v", tt.engine, got, tt.expected)
			}
		})
	}
}
