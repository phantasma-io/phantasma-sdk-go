package carbon

import "testing"

// Wire-format tests for the gas-model-v2 GasConfig extension. The chain serializes the 10 v2
// fields only for version >= 1; the version-0 image is frozen forever (historical replay), so
// these tests pin both layouts and the truncation failure mode.

// liveV1GasConfig returns the mainnet v1 values (feeMultiplier 10000, transfer 10 units, byte
// fee 250000, escrow 2).
func liveV1GasConfig() GasConfig {
	return GasConfig{
		Version:                 0,
		MaxNameLength:           32,
		MaxTokenSymbolLength:    10,
		FeeShift:                0,
		MaxStructureSize:        65536,
		FeeMultiplier:           10_000,
		GasTokenID:              2,
		DataTokenID:             1,
		MinimumGasOffer:         10,
		DataEscrowPerRow:        2,
		GasFeeTransfer:          10,
		GasFeeQuery:             2,
		GasFeeCreateTokenBase:   10_000_000_000,
		GasFeeCreateTokenSymbol: 10_000_000_000,
		GasFeeCreateTokenSeries: 2_500_000_000,
		GasFeePerByte:           250_000,
		GasFeeRegisterName:      10_000_000_000_000,
		GasBurnRatioMul:         1,
		GasBurnRatioShift:       0,
	}
}

// v2GasConfig returns the spec activation-package values for the v2 tail.
func v2GasConfig() GasConfig {
	c := liveV1GasConfig()
	c.Version = 1
	c.DataEscrowPerRow = 200_000
	c.MinimumGasBill = 10_000_000
	c.PolicyFeeCreateTokenBase = 100_000_000_000_000
	c.PolicyFeeCreateTokenSymbol = 100_000_000_000_000
	c.PolicyFeeCreateTokenSeries = 25_000_000_000_000
	c.PolicyFeeRegisterName = 100_000_000_000_000_000
	c.LegacyDataEscrowPerRow = 2
	return c
}

func serializeGasConfig(t *testing.T, c GasConfig) []byte {
	t.Helper()
	w := NewWriter()
	c.WriteCarbon(w)
	return w.Bytes()
}

// Version-0 configs must keep the exact pre-v2 wire size (113 bytes); any growth would corrupt
// every historical block image.
func TestGasConfigV0KeepsLegacy113ByteLayout(t *testing.T) {
	got := serializeGasConfig(t, liveV1GasConfig())
	if len(got) != 113 {
		t.Fatalf("v0 image length = %d, want 113", len(got))
	}
}

// A version>=1 config appends the 66-byte v2 tail (8x u64 + 2x u8) after an unchanged head
// encoding - the tail is a pure wire extension, it must not disturb the first 113 bytes.
func TestGasConfigV2Appends66ByteTail(t *testing.T) {
	v2Bytes := serializeGasConfig(t, v2GasConfig())
	if len(v2Bytes) != 179 {
		t.Fatalf("v2 image length = %d, want 179", len(v2Bytes))
	}

	v0Twin := v2GasConfig()
	v0Twin.Version = 0 // same head values, version-0 layout
	v0Bytes := serializeGasConfig(t, v0Twin)
	if len(v0Bytes) != 113 {
		t.Fatalf("v0 twin image length = %d, want 113", len(v0Bytes))
	}

	if v2Bytes[0] != 1 || v0Bytes[0] != 0 {
		t.Fatalf("version bytes = %d/%d, want 1/0", v2Bytes[0], v0Bytes[0])
	}
	for i := 1; i < 113; i++ {
		if v2Bytes[i] != v0Bytes[i] {
			t.Fatalf("head byte %d diverged: %d != %d", i, v2Bytes[i], v0Bytes[i])
		}
	}
}

func TestGasConfigV2RoundtripPreservesAllFields(t *testing.T) {
	original := v2GasConfig()
	var decoded GasConfig
	decoded.ReadCarbon(NewReader(serializeGasConfig(t, original)))

	if decoded != original {
		t.Fatalf("v2 roundtrip mismatch:\n got %+v\nwant %+v", decoded, original)
	}
	if !decoded.HasGasModelV2() {
		t.Fatal("decoded v2 config must report HasGasModelV2")
	}
}

// Reading a version-0 image must zero the v2 fields even on a dirty instance - consumers must
// never see stale tail values on a v1 chain.
func TestGasConfigV0ReadZeroesV2Fields(t *testing.T) {
	dirty := v2GasConfig() // instance with nonzero v2 fields
	dirty.ReadCarbon(NewReader(serializeGasConfig(t, liveV1GasConfig())))

	if dirty.HasGasModelV2() {
		t.Fatal("v0 image must not report gas model v2")
	}
	if dirty.MinimumGasBill != 0 || dirty.PolicyFeeCreateTokenBase != 0 || dirty.LegacyDataEscrowPerRow != 0 {
		t.Fatalf("v2 fields not zeroed: %+v", dirty)
	}
}

// A version>=1 image truncated to the version-0 length must FAIL to parse, never silently
// produce a config with zeroed v2 prices (that would mean free product actions).
func TestGasConfigTruncatedV2ImageFailsToParse(t *testing.T) {
	truncated := serializeGasConfig(t, v2GasConfig())[:113]

	defer func() {
		if recover() == nil {
			t.Fatal("reading a truncated v2 image must panic (end of stream)")
		}
	}()
	var decoded GasConfig
	decoded.ReadCarbon(NewReader(truncated))
}
