package response

import (
	"encoding/json"
	"strings"
	"testing"
)

// estimateTransaction response decoding: the node serializes 64-bit amounts as decimal strings
// (JSON-number precision), and a completed estimate must convert into the same
// carbon.NativeFeeEstimate the Tier-1 estimator produces so wallet code consumes both tiers
// identically.

// Completed dry run: recommendations present, no abort. recommendedMaxGas deliberately exceeds
// 2^53 to pin the strings-not-numbers decision.
const estimateCompletedJSON = `{
  "wouldAbort": false,
  "abortReason": "",
  "gasBillKcalBase": "10000000",
  "dataRows": "1",
  "dataEscrowAtoms": "200000",
  "dataRefundAtoms": "0",
  "recommendedMaxGas": "100000000000000000",
  "recommendedMaxData": "400000"
}`

// Aborted dry run: the settled abort bill is still reported (aborts pay), recommendations are 0.
const estimateAbortedJSON = `{
  "wouldAbort": true,
  "abortReason": "gas fees [gas=3125 max=40]",
  "gasBillKcalBase": "40",
  "dataRows": "0",
  "dataEscrowAtoms": "0",
  "dataRefundAtoms": "0",
  "recommendedMaxGas": "0",
  "recommendedMaxData": "0"
}`

func TestEstimateTransactionCompletedConverts(t *testing.T) {
	var result EstimateTransactionResult
	if err := json.Unmarshal([]byte(estimateCompletedJSON), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.WouldAbort {
		t.Fatalf("completed estimate should not report wouldAbort")
	}
	estimate, err := result.ToFeeEstimate()
	if err != nil {
		t.Fatalf("ToFeeEstimate: %v", err)
	}
	// Above-2^53 value survives exactly because it rides a string.
	if estimate.MaxGas != 100_000_000_000_000_000 {
		t.Fatalf("MaxGas = %d, want 100000000000000000", estimate.MaxGas)
	}
	if estimate.MaxData != 400_000 {
		t.Fatalf("MaxData = %d, want 400000", estimate.MaxData)
	}
	if estimate.ExpectedGasBill != 10_000_000 {
		t.Fatalf("ExpectedGasBill = %d, want 10000000", estimate.ExpectedGasBill)
	}
}

// An aborted simulation has no recommendations; converting must fail rather than yield zero
// ceilings a wallet could sign with.
func TestEstimateTransactionAbortedRefusesConversion(t *testing.T) {
	var result EstimateTransactionResult
	if err := json.Unmarshal([]byte(estimateAbortedJSON), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !result.WouldAbort {
		t.Fatalf("aborted estimate should report wouldAbort")
	}
	if _, err := result.ToFeeEstimate(); err == nil || !strings.Contains(err.Error(), "gas fees") {
		t.Fatalf("expected an abort error mentioning the reason, got %v", err)
	}
}

// A malformed server response (lost field) must not silently become a zero ceiling.
func TestEstimateTransactionMissingFieldFails(t *testing.T) {
	var result EstimateTransactionResult
	if err := json.Unmarshal([]byte(`{"wouldAbort": false}`), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, err := result.ToFeeEstimate(); err == nil {
		t.Fatalf("expected a missing-field error, got nil")
	}
}
