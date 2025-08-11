package snapshot

import (
	"testing"

	"github.com/abdultolba/nizam/internal/dockerx"
)

func TestMongoDBEngine_GetEngineType(t *testing.T) {
	docker := &dockerx.Client{} // Mock client for testing
	engine := NewMongoDBEngine(docker)

	if got := engine.GetEngineType(); got != "mongo" {
		t.Errorf("GetEngineType() = %v, want %v", got, "mongo")
	}
}

func TestMongoDBEngine_CanHandle(t *testing.T) {
	docker := &dockerx.Client{} // Mock client for testing
	engine := NewMongoDBEngine(docker)

	tests := []struct {
		name       string
		engineType string
		want       bool
	}{
		{
			name:       "handles mongodb",
			engineType: "mongodb",
			want:       true,
		},
		{
			name:       "handles mongo",
			engineType: "mongo",
			want:       true,
		},
		{
			name:       "does not handle postgres",
			engineType: "postgres",
			want:       false,
		},
		{
			name:       "does not handle mysql",
			engineType: "mysql",
			want:       false,
		},
		{
			name:       "does not handle redis",
			engineType: "redis",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := engine.CanHandle(tt.engineType); got != tt.want {
				t.Errorf("CanHandle(%s) = %v, want %v", tt.engineType, got, tt.want)
			}
		})
	}
}
