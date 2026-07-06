package response

import (
	"encoding/json"
	"strings"
	"testing"
)

// getGasConfig response decoding: the node serializes 64-bit config values as decimal strings
// (JSON-number precision) and omits the v2 tail fields for version-0 configs. ToGasConfig must
// produce the exact wire config the Tier-1 estimator consumes.

// Shape currently served by a v1 (pre-flip) chain: no v2 fields at all.
const gasConfigV1JSON = `{
  "gasModelVersion": 1,
  "blockRateTarget": 2000,
  "expiryWindow": 90000,
  "gasConfig": {
    "version": 0,
    "maxNameLength": 32,
    "maxTokenSymbolLength": 10,
    "feeShift": 0,
    "maxStructureSize": 65536,
    "feeMultiplier": "10000",
    "gasTokenId": "2",
    "dataTokenId": "1",
    "minimumGasOffer": "10",
    "dataEscrowPerRow": "2",
    "gasFeeTransfer": "10",
    "gasFeeQuery": "2",
    "gasFeeCreateTokenBase": "10000000000",
    "gasFeeCreateTokenSymbol": "10000000000",
    "gasFeeCreateTokenSeries": "2500000000",
    "gasFeePerByte": "250000",
    "gasFeeRegisterName": "10000000000000",
    "gasBurnRatioMul": "1",
    "gasBurnRatioShift": 0
  }
}`

// Post-flip shape: version 1 config with the full v2 tail (policyFeeRegisterName deliberately
// exceeds 2^53 to pin the strings-not-numbers decision).
const gasConfigV2JSON = `{
  "gasModelVersion": 2,
  "blockRateTarget": 2000,
  "expiryWindow": 90000,
  "unitsPerBlockDataByte": 25,
  "gasConfig": {
    "version": 1,
    "maxNameLength": 32,
    "maxTokenSymbolLength": 10,
    "feeShift": 0,
    "maxStructureSize": 65536,
    "feeMultiplier": "10000",
    "gasTokenId": "2",
    "dataTokenId": "1",
    "minimumGasOffer": "10",
    "dataEscrowPerRow": "200000",
    "gasFeeTransfer": "10",
    "gasFeeQuery": "2",
    "gasFeeCreateTokenBase": "10000000000",
    "gasFeeCreateTokenSymbol": "10000000000",
    "gasFeeCreateTokenSeries": "2500000000",
    "gasFeePerByte": "250000",
    "gasFeeRegisterName": "10000000000000",
    "gasBurnRatioMul": "1",
    "gasBurnRatioShift": 0,
    "minimumGasBill": "10000000",
    "gasProducerRatioMul": "0",
    "gasProducerRatioShift": 0,
    "gasDappRatioMul": "0",
    "gasDappRatioShift": 0,
    "policyFeeCreateTokenBase": "100000000000000",
    "policyFeeCreateTokenSymbol": "100000000000000",
    "policyFeeCreateTokenSeries": "25000000000000",
    "policyFeeRegisterName": "100000000000000000",
    "legacyDataEscrowPerRow": "2"
  }
}`

func TestGasConfigResultV1Decodes(t *testing.T) {
	var result GasConfigResult
	if err := json.Unmarshal([]byte(gasConfigV1JSON), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.GasModelVersion != 1 || result.BlockRateTarget != 2000 || result.ExpiryWindow != 90000 {
		t.Fatalf("header mismatch: %+v", result)
	}
	if result.UnitsPerBlockDataByte != nil {
		t.Fatal("unitsPerBlockDataByte must be absent under gas model v1")
	}

	config, err := result.ToGasConfig()
	if err != nil {
		t.Fatalf("ToGasConfig: %v", err)
	}
	if config.Version != 0 || config.HasGasModelV2() {
		t.Fatalf("version mismatch: %+v", config)
	}
	if config.FeeMultiplier != 10_000 || config.GasFeePerByte != 250_000 {
		t.Fatalf("v1 fields mismatch: %+v", config)
	}
	// v2 fields absent from the wire must map to zero, not garbage.
	if config.MinimumGasBill != 0 || config.PolicyFeeRegisterName != 0 {
		t.Fatalf("v2 fields not zeroed: %+v", config)
	}
}

func TestGasConfigResultV2Decodes(t *testing.T) {
	var result GasConfigResult
	if err := json.Unmarshal([]byte(gasConfigV2JSON), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.UnitsPerBlockDataByte == nil || *result.UnitsPerBlockDataByte != 25 {
		t.Fatalf("unitsPerBlockDataByte mismatch: %+v", result.UnitsPerBlockDataByte)
	}

	config, err := result.ToGasConfig()
	if err != nil {
		t.Fatalf("ToGasConfig: %v", err)
	}
	if !config.HasGasModelV2() || config.DataEscrowPerRow != 200_000 || config.MinimumGasBill != 10_000_000 {
		t.Fatalf("v2 fields mismatch: %+v", config)
	}
	// Above-2^53 value survives exactly because it rides a string.
	if config.PolicyFeeRegisterName != 100_000_000_000_000_000 {
		t.Fatalf("policyFeeRegisterName = %d", config.PolicyFeeRegisterName)
	}
	if config.LegacyDataEscrowPerRow != 2 {
		t.Fatalf("legacyDataEscrowPerRow = %d", config.LegacyDataEscrowPerRow)
	}
}

// A response claiming gas model v2 but missing tail fields must fail loudly - estimating fees
// from silently zeroed v2 prices would produce rejected transactions.
func TestGasConfigResultV2MissingTailFieldFails(t *testing.T) {
	broken := strings.Replace(gasConfigV2JSON, `"minimumGasBill": "10000000",`, "", 1)
	var result GasConfigResult
	if err := json.Unmarshal([]byte(broken), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, err := result.ToGasConfig(); err == nil || !strings.Contains(err.Error(), "minimumGasBill") {
		t.Fatalf("expected a minimumGasBill error, got %v", err)
	}
}

func TestGasConfigResultMissingSectionFails(t *testing.T) {
	var result GasConfigResult
	if err := json.Unmarshal([]byte(`{"gasModelVersion": 1}`), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, err := result.ToGasConfig(); err == nil {
		t.Fatal("missing gasConfig section must fail")
	}
}
