package logger

import (
	"testing"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantError bool
	}{
		{"valid debug", "debug", false},
		{"valid info", "info", false},
		{"invalid level", "notalevel", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.level)
			if (err != nil) != tt.wantError {
				t.Errorf("Initialize(%q) error = %v, wantError %v", tt.level, err, tt.wantError)
			}
		})
	}
}
