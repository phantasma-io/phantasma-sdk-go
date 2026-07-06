package carbon

import "testing"

// Tier-1 fee calculator tests. Expected numbers are hand-derived from the chain billing formula
// (bill = work-units and byte fee through the feeMultiplier/feeShift knob, plus v2 policy fees
// and the v2 minimum-bill floor), pinned as constants so any formula regression fails loudly.
// The same fixtures and expectations exist in every SDK (parity suite).

func mustEstimate(t *testing.T, kind NativeFeeKind, config GasConfig, params NativeFeeParams) NativeFeeEstimate {
	t.Helper()
	estimate, err := EstimateNativeFee(kind, &config, params)
	if err != nil {
		t.Fatalf("EstimateNativeFee: %v", err)
	}
	return estimate
}

// v1 transfer with an existing recipient row: bill is the pure work term 10 * 10000.
func TestEstimateV1TransferExistingRecipient(t *testing.T) {
	estimate := mustEstimate(t, NativeFeeTransferFungible, liveV1GasConfig(), NativeFeeParams{FreshRows: NoFreshRows})

	if estimate.ExpectedGasBill != 100_000 {
		t.Fatalf("expected bill = %d, want 100000", estimate.ExpectedGasBill)
	}
	// stdFee shape: 2x min offer + work + flat 1 KiB byte allowance.
	if want := uint64(10*2 + 100_000 + 1024*250_000); estimate.MaxGas != want {
		t.Fatalf("maxGas = %d, want %d", estimate.MaxGas, want)
	}
	if estimate.MaxData != 0 {
		t.Fatalf("maxData = %d, want 0", estimate.MaxData)
	}
}

// v1 transfer worst case (1 fresh row): the row quantum joins the byte fee and the escrow shows
// up in MaxData at the v1 price.
func TestEstimateV1TransferWithFreshRow(t *testing.T) {
	estimate := mustEstimate(t, NativeFeeTransferFungible, liveV1GasConfig(), NativeFeeParams{})

	if want := uint64(100_000 + 250_000); estimate.ExpectedGasBill != want {
		t.Fatalf("expected bill = %d, want %d", estimate.ExpectedGasBill, want)
	}
	if estimate.MaxData != 2 {
		t.Fatalf("maxData = %d, want 2", estimate.MaxData)
	}
}

// v2 transfer, default envelope 512 + 1 fresh row: blockData 513 -> 12825 byte units + 10 work
// units = 12835 units * 10000 = 128_350_000 kcal-base (above the 1e7 floor).
func TestEstimateV2TransferDefaultEnvelope(t *testing.T) {
	estimate := mustEstimate(t, NativeFeeTransferFungible, v2GasConfig(), NativeFeeParams{})

	if estimate.ExpectedGasBill != 128_350_000 {
		t.Fatalf("expected bill = %d, want 128350000", estimate.ExpectedGasBill)
	}
	if want := uint64(128_350_000 + 128_350_000/4); estimate.MaxGas != want {
		t.Fatalf("maxGas = %d, want %d", estimate.MaxGas, want)
	}
	if estimate.MaxData != 200_000 {
		t.Fatalf("maxData = %d, want 200000", estimate.MaxData)
	}
}

// v2 exact envelope: a measured 250-byte native transfer to an existing recipient bills
// (10 + 250*25) * 10000.
func TestEstimateV2TransferExactEnvelope(t *testing.T) {
	estimate := mustEstimate(t, NativeFeeTransferFungible, v2GasConfig(), NativeFeeParams{EnvelopeBytes: 250, FreshRows: NoFreshRows})

	if want := uint64((10 + 250*25) * 10_000); estimate.ExpectedGasBill != want {
		t.Fatalf("expected bill = %d, want %d", estimate.ExpectedGasBill, want)
	}
}

// A tiny v2 tx can never bill below the consensus floor; the offer must also respect the
// admission check maxGas >= minimumGasBill.
func TestEstimateV2FloorAppliesToSmallBills(t *testing.T) {
	config := v2GasConfig()
	config.MinimumGasBill = 10_000_000_000 // exaggerated floor above the computed bill

	estimate := mustEstimate(t, NativeFeeTransferFungible, config, NativeFeeParams{EnvelopeBytes: 250, FreshRows: NoFreshRows})

	if estimate.ExpectedGasBill != 10_000_000_000 {
		t.Fatalf("expected bill = %d, want the floor", estimate.ExpectedGasBill)
	}
	if estimate.MaxGas < 10_000_000_000 {
		t.Fatalf("maxGas = %d, must cover the floor", estimate.MaxGas)
	}
}

// NFT transfers scale the work term per instance; under v2 each instance also recreates its
// lookup row, so the escrow allowance is (count + 1) rows.
func TestEstimateV2NftMultiTransfer(t *testing.T) {
	estimate := mustEstimate(t, NativeFeeTransferNonFungible, v2GasConfig(),
		NativeFeeParams{Count: 5, EnvelopeBytes: 300})

	// work 5*10 units + bytes (300 envelope + 6 rows) * 25 units, all * 10000.
	if want := uint64((50 + 306*25) * 10_000); estimate.ExpectedGasBill != want {
		t.Fatalf("expected bill = %d, want %d", estimate.ExpectedGasBill, want)
	}
	if want := uint64(6 * 200_000); estimate.MaxData != want {
		t.Fatalf("maxData = %d, want %d", estimate.MaxData, want)
	}
}

// CreateToken under v1 charges unit-priced product fees through the multiplier; under v2 it
// pays the direct kcal-base policy fee (no multiplier) plus the byte fee for its envelope. The
// policy magnitude equals the v1 price by design.
func TestEstimateCreateTokenBothModels(t *testing.T) {
	v1 := mustEstimate(t, NativeFeeCreateToken, liveV1GasConfig(),
		NativeFeeParams{SymbolLength: 4, FreshRows: NoFreshRows, EnvelopeBytes: 1000})
	if want := uint64((10_000_000_000 + 1_250_000_000) * 10_000); v1.ExpectedGasBill != want {
		t.Fatalf("v1 bill = %d, want %d", v1.ExpectedGasBill, want)
	}

	v2 := mustEstimate(t, NativeFeeCreateToken, v2GasConfig(),
		NativeFeeParams{SymbolLength: 4, FreshRows: NoFreshRows, EnvelopeBytes: 1000})
	policy := uint64(100_000_000_000_000 + 100_000_000_000_000>>3)
	byteFee := uint64(1000 * 25 * 10_000)
	if want := policy + byteFee; v2.ExpectedGasBill != want {
		t.Fatalf("v2 bill = %d, want %d", v2.ExpectedGasBill, want)
	}
}

// RegisterName halves the price per character after the first, under both models.
func TestEstimateRegisterNameLengthDiscount(t *testing.T) {
	v1 := mustEstimate(t, NativeFeeRegisterName, liveV1GasConfig(),
		NativeFeeParams{NameLength: 8, FreshRows: NoFreshRows, EnvelopeBytes: 300})
	if want := uint64(10_000_000_000_000>>7) * 10_000; v1.ExpectedGasBill != want {
		t.Fatalf("v1 bill = %d, want %d", v1.ExpectedGasBill, want)
	}

	v2 := mustEstimate(t, NativeFeeRegisterName, v2GasConfig(),
		NativeFeeParams{NameLength: 8, FreshRows: NoFreshRows, EnvelopeBytes: 300})
	if want := uint64(100_000_000_000_000_000>>7) + 300*25*10_000; v2.ExpectedGasBill != want {
		t.Fatalf("v2 bill = %d, want %d", v2.ExpectedGasBill, want)
	}
}

// The Script kind budgets a generous VM unit allowance (default 5000 exceeds every script in
// mainnet history) instead of pretending opcode costs are closed-form.
func TestEstimateScriptBudgetsVmAllowance(t *testing.T) {
	estimate := mustEstimate(t, NativeFeeScript, v2GasConfig(), NativeFeeParams{EnvelopeBytes: 568, FreshRows: NoFreshRows})

	// (5000 vm units + (568 + 512 events) * 25) * 10000
	if want := uint64((5000 + 1080*25) * 10_000); estimate.ExpectedGasBill != want {
		t.Fatalf("expected bill = %d, want %d", estimate.ExpectedGasBill, want)
	}
}

// Envelope arithmetic mirrors SignedTxMsg: native kinds append bare 64-byte signatures,
// call/script kinds append a length-prefixed 96-byte witness array.
func TestEnvelopeBytesFollowWitnessLayout(t *testing.T) {
	cases := []struct {
		kind    NativeFeeKind
		msgLen  int
		signers int
		want    uint32
	}{
		{NativeFeeTransferFungible, 150, 1, 150 + 64},
		{NativeFeeTransferFungible, 150, 2, 150 + 128},
		{NativeFeeCreateToken, 900, 1, 900 + 4 + 96},
		{NativeFeeScript, 500, 2, 500 + 4 + 192},
	}
	for _, tc := range cases {
		got, err := EnvelopeBytesFor(tc.kind, tc.msgLen, tc.signers)
		if err != nil {
			t.Fatalf("EnvelopeBytesFor(%v): %v", tc.kind, err)
		}
		if got != tc.want {
			t.Fatalf("EnvelopeBytesFor(%v, %d, %d) = %d, want %d", tc.kind, tc.msgLen, tc.signers, got, tc.want)
		}
	}
}

// Guard rails: impossible inputs are rejected instead of quoting fees for txs the chain would
// never admit.
func TestEstimateRejectsInvalidInputs(t *testing.T) {
	config := liveV1GasConfig()
	if _, err := EstimateNativeFee(NativeFeeRegisterName, &config, NativeFeeParams{}); err == nil {
		t.Fatal("RegisterName without NameLength must fail")
	}
	// maxTokenSymbolLength is 10
	if _, err := EstimateNativeFee(NativeFeeCreateToken, &config, NativeFeeParams{SymbolLength: 11}); err == nil {
		t.Fatal("oversized symbol must fail")
	}
	if _, err := EstimateNativeFee(NativeFeeTransferFungible, nil, NativeFeeParams{}); err == nil {
		t.Fatal("nil config must fail")
	}
}

// feeShift semantics: the chain clamps shifts >= 64 to a zero work delta; the estimator must
// match rather than undercharge/overcharge.
func TestEstimateOversizedFeeShiftZeroesScaledTerms(t *testing.T) {
	config := liveV1GasConfig()
	config.FeeShift = 64

	estimate := mustEstimate(t, NativeFeeTransferFungible, config, NativeFeeParams{FreshRows: NoFreshRows})

	// v1: work term zeroed, byte fee (raw kcal-base knob) unaffected by the shift.
	if estimate.ExpectedGasBill != 0 {
		t.Fatalf("expected bill = %d, want 0", estimate.ExpectedGasBill)
	}
}
