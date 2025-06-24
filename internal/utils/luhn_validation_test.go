package utils

import "testing"

func TestIsValidLuhn(t *testing.T) {
	tests := []struct {
		input   string
		want    bool
		comment string
	}{
		{"79927398713", true, "classic valid Luhn"},
		{"1234567812345670", true, "valid credit card number"},
		{"1234567812345678", false, "invalid credit card number"},
		{"49927398716", true, "another valid Luhn"},
		{"49927398717", false, "invalid Luhn"},
		{"abcdefg", false, "non-numeric input"},
		{"", false, "empty string"},
		{"0", true, "single zero is valid"},
		{"059", true, "valid short Luhn"},
		{"0591", false, "invalid short Luhn"},
	}

	for _, tt := range tests {
		t.Run(tt.input+"_"+tt.comment, func(t *testing.T) {
			got := IsValidLuhn(tt.input)
			if got != tt.want {
				t.Errorf("IsValidLuhn(%q) = %v, want %v (%s)", tt.input, got, tt.want, tt.comment)
			}
		})
	}
}
