package types

import "testing"

func TestNewTimestamp(t *testing.T) {
	if got := NewTimestamp(123); got.Value != 123 {
		t.Fatalf("timestamp value mismatch: %d", got.Value)
	}
}
