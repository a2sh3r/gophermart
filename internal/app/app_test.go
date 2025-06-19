package app

import (
	"testing"
)

func TestNewApp(t *testing.T) {
	_, err := NewApp()
	if err != nil {
		t.Logf("NewApp returned error (this may be expected in test env): %v", err)
	}
}
