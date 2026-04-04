package templates

import (
	"testing"
	"time"
)

func TestIsStuckProcessing(t *testing.T) {
	fn := funcMap["isStuckProcessing"].(func(string, time.Time) bool)

	now := time.Now()

	tests := []struct {
		name      string
		status    string
		updatedAt time.Time
		want      bool
	}{
		{
			name:      "processing for 6 minutes is stuck",
			status:    "processing",
			updatedAt: now.Add(-6 * time.Minute),
			want:      true,
		},
		{
			name:      "processing for 4 minutes is not stuck",
			status:    "processing",
			updatedAt: now.Add(-4 * time.Minute),
			want:      false,
		},
		{
			name:      "failed status is never stuck-processing",
			status:    "failed",
			updatedAt: now.Add(-1 * time.Hour),
			want:      false,
		},
		{
			name:      "pending status is never stuck-processing",
			status:    "pending",
			updatedAt: now.Add(-1 * time.Hour),
			want:      false,
		},
		{
			name:      "completed status is never stuck-processing",
			status:    "completed",
			updatedAt: now.Add(-1 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fn(tt.status, tt.updatedAt)
			if got != tt.want {
				t.Errorf("isStuckProcessing(%q, %v) = %v, want %v", tt.status, tt.updatedAt, got, tt.want)
			}
		})
	}
}
