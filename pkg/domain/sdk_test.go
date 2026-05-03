package domain

import "testing"

func TestSDKPayload(t *testing.T) {
	if string(SDKPayload) == "" {
		t.Fatalf("SDK payload must not be empty")
	}
}
