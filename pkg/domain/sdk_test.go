package domain

import "testing"

func TestSDKPayload(t *testing.T) {
	expected := "GO-SDK-v" + SDKVersion
	if string(SDKPayload) != expected {
		t.Fatalf("SDK payload mismatch: got %q want %q", string(SDKPayload), expected)
	}
}
