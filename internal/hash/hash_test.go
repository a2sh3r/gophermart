package hash

import (
	"testing"
)

func TestCalculateHash(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		key       string
		wantEmpty bool
	}{
		{"empty key", "data", "", true},
		{"empty data", "", "key", false},
		{"normal", "data", "key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateHash(tt.data, tt.key)
			if tt.wantEmpty && got != "" {
				t.Errorf("CalculateHash(%q, %q) = %q, want empty string", tt.data, tt.key, got)
			}
			if !tt.wantEmpty && got == "" {
				t.Errorf("CalculateHash(%q, %q) = empty, want non-empty", tt.data, tt.key)
			}
		})
	}
}

func TestVerifyHash(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		key       string
		hash      string
		wantError bool
	}{
		{"empty key", "data", "", "any", false},
		{"correct hash", "data", "key", CalculateHash("data", "key"), false},
		{"wrong hash", "data", "key", "wronghash", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyHash(tt.data, tt.key, tt.hash)
			if (err != nil) != tt.wantError {
				t.Errorf("VerifyHash(%q, %q, %q) error = %v, wantError %v", tt.data, tt.key, tt.hash, err, tt.wantError)
			}
		})
	}
}
