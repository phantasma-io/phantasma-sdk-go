package domain

import (
	"regexp"
	"testing"
)

// The SDK payload tag ("GO-SDK-v"+SDKVersion) is embedded in on-chain transactions, so its shape is a
// wire contract: exactly "GO-SDK-v" followed by a bare MAJOR.MINOR.PATCH version. These tests catch a
// malformed SDKVersion (empty, partial, or a stray leading "v" that would yield "GO-SDK-vv...") from
// reaching the wire. They intentionally do NOT assert the version equals the released git tag -
// SDKVersion is a hardcoded constant and keeping it in sync with the tag is a manual release step.

func TestSDKVersionIsBareSemver(t *testing.T) {
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(SDKVersion) {
		t.Fatalf("SDKVersion must be a bare MAJOR.MINOR.PATCH semver (no leading 'v'), got %q", SDKVersion)
	}
}

func TestSDKPayloadIsWellFormed(t *testing.T) {
	got := string(SDKPayload)
	if !regexp.MustCompile(`^GO-SDK-v\d+\.\d+\.\d+$`).MatchString(got) {
		t.Fatalf("SDK payload must be GO-SDK-v<MAJOR.MINOR.PATCH>, got %q", got)
	}
}
