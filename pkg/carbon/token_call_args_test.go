package carbon

import (
	"math/big"
	"testing"
)

const (
	transferFungibleArgsHex = "1111111111111111111111111111111111111111111111111111111111111111222222222222222222222222222222222222222222222222222222222222222201000000000000000800E1F50500000000"
	mintFungibleArgsHex     = "010000000000000011111111111111111111111111111111111111111111111111111111111111110800E1F50500000000"
	transferNFTArgsHex      = "1111111111111111111111111111111111111111111111111111111111111111222222222222222222222222222222222222222222222222222222222222222201000000000000000200000007000000000000000800000000000000"
	burnFungibleArgsHex     = "010000000000000022222222222222222222222222222222222222222222222222222222222222220800E1F50500000000"
	burnNFTArgsHex          = "010000000000000022222222222222222222222222222222222222222222222222222222222222220200000007000000000000000800000000000000"
	emptySchemaHex          = "0000000000"
	seriesInfoHex           = "0300000009000000333333333333333333333333333333333333333333333333333333333333333302000000AABB" + emptySchemaHex + emptySchemaHex
	createSeriesArgsHex     = "0900000000000000" + seriesInfoHex
	createMintedSeriesHex   = "0900000000000000" + seriesInfoHex + "4444444444444444444444444444444444444444444444444444444444444444" + "0200000002000000010200000000" + "010000000100000003"
	updateTokenMetadataHex  = "0900000000000000" + "01000000016E16616C70686100"
	updateSeriesMetadataHex = "09000000000000000700000004000000DEADBEEF"
	mintPhantasmaNFTHex     = "0900000000000000" + "4444444444444444444444444444444444444444444444444444444444444444" + "02000000" + "082A0000000000000002000000AABB01000000CC" + "082B000000000000000000000002000000DDEE"
	phantasmaNFTResultHex   = "5555555555555555555555555555555555555555555555555555555555555555" + "7B00000000000000"
)

func TestTokenCallArgsVectors(t *testing.T) {
	// Token-call argument vectors cover the Carbon token module ABI used by create/mint/transfer/burn calls.
	one := repeatedBytes32(0x11)
	two := repeatedBytes32(0x22)
	four := repeatedBytes32(0x44)

	tests := []struct {
		name string
		hex  string
		blob Blob
	}{
		{
			name: "mint_fungible",
			hex:  mintFungibleArgsHex,
			blob: &MintFungibleArgs{TokenID: 1, To: one, Amount: IntXFromInt64(100_000_000)},
		},
		{
			name: "transfer_fungible",
			hex:  transferFungibleArgsHex,
			blob: &TransferFungibleArgs{To: one, From: two, TokenID: 1, Amount: IntXFromInt64(100_000_000)},
		},
		{
			name: "transfer_non_fungible",
			hex:  transferNFTArgsHex,
			blob: &TransferNonFungibleArgs{To: one, From: two, TokenID: 1, InstanceIDs: []uint64{7, 8}},
		},
		{
			name: "burn_fungible",
			hex:  burnFungibleArgsHex,
			blob: &BurnFungibleArgs{TokenID: 1, From: two, Amount: IntXFromInt64(100_000_000)},
		},
		{
			name: "burn_non_fungible",
			hex:  burnNFTArgsHex,
			blob: &BurnNonFungibleArgs{TokenID: 1, From: two, InstanceIDs: []uint64{7, 8}},
		},
		{
			name: "tokens_config",
			hex:  "11",
			blob: &TokensConfig{Flags: TokensConfigFlagsRequireMetadata | TokensConfigFlagsAllowExplicitNFTMetaIDMint},
		},
		{
			name: "create_token_series",
			hex:  createSeriesArgsHex,
			blob: &CreateTokenSeriesArgs{TokenID: 9, Info: vectorSeriesInfo()},
		},
		{
			name: "create_minted_token_series",
			hex:  createMintedSeriesHex,
			blob: &CreateMintedTokenSeriesArgs{
				TokenID: 9,
				Info:    vectorSeriesInfo(),
				Address: four,
				ROMs:    [][]byte{{0x01, 0x02}, {}},
				RAMs:    [][]byte{{0x03}},
			},
		},
		{
			name: "update_token_metadata",
			hex:  updateTokenMetadataHex,
			blob: &UpdateTokenMetadataArgs{
				TokenID: 9,
				Metadata: VMDynamicStruct{Fields: []VMNamedDynamicVariable{
					MustVMNamedDynamicVariable("n", VMTypeString, "alpha"),
				}},
			},
		},
		{
			name: "update_series_metadata",
			hex:  updateSeriesMetadataHex,
			blob: &UpdateSeriesMetadataArgs{TokenID: 9, SeriesID: 7, Metadata: []byte{0xde, 0xad, 0xbe, 0xef}},
		},
		{
			name: "mint_phantasma_non_fungible",
			hex:  mintPhantasmaNFTHex,
			blob: &MintPhantasmaNonFungibleArgs{
				TokenID: 9,
				Address: four,
				Tokens: []PhantasmaNFTMintInfo{
					{PhantasmaSeriesID: IntXFromInt64(42), ROM: []byte{0xaa, 0xbb}, RAM: []byte{0xcc}},
					{PhantasmaSeriesID: IntXFromInt64(43), ROM: nil, RAM: []byte{0xdd, 0xee}},
				},
			},
		},
		{
			name: "phantasma_nft_mint_result",
			hex:  phantasmaNFTResultHex,
			blob: &PhantasmaNFTMintResult{PhantasmaNFTID: repeatedBytes32(0x55), CarbonInstanceID: 123},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSerializedHex(t, tt.hex, tt.blob)
			readCarbon(t, tt.hex, tt.blob)
			assertSerializedHex(t, tt.hex, tt.blob)
		})
	}
}

func TestTokenResultParsers(t *testing.T) {
	// Result parsers should return typed values for valid RPC result blobs and errors for malformed payloads.
	tokenID, err := ParseCreateTokenResult("0900000000000000")
	if err != nil || tokenID != 9 {
		t.Fatalf("unexpected create token result: %d/%v", tokenID, err)
	}
	seriesID, err := ParseCreateTokenSeriesResult("07000000")
	if err != nil || seriesID != 7 {
		t.Fatalf("unexpected create series result: %d/%v", seriesID, err)
	}

	mintResult, err := ParseMintNonFungibleResult(9, "0200000007000000000000000800000000000000")
	if err != nil {
		t.Fatalf("mint result parser failed: %v", err)
	}
	if len(mintResult) != 2 || mintResult[0] != GetNFTAddress(9, 7) || mintResult[1] != GetNFTAddress(9, 8) {
		t.Fatalf("unexpected mint result: %v", mintResult)
	}

	payloadWriter := NewWriter()
	payloadWriter.Write4(2)
	(&PhantasmaNFTMintResult{PhantasmaNFTID: repeatedBytes32(0x55), CarbonInstanceID: 7}).WriteCarbon(payloadWriter)
	high := repeatedBytes32(0)
	high[0] = 0x2a
	high[31] = 0x80
	(&PhantasmaNFTMintResult{PhantasmaNFTID: high, CarbonInstanceID: 8}).WriteCarbon(payloadWriter)

	results, err := ParseMintPhantasmaNonFungibleResult(hexString(payloadWriter.Bytes()))
	if err != nil {
		t.Fatalf("phantasma mint result parser failed: %v", err)
	}
	if len(results) != 2 || results[0].CarbonInstanceID != 7 || results[1].PhantasmaNFTID != high {
		t.Fatalf("unexpected phantasma mint result: %v", results)
	}

	if _, err := ParseMintNonFungibleResult(9, "not-hex"); err == nil {
		t.Fatalf("malformed result hex must return an error")
	}
	if _, err := ParseMintPhantasmaNonFungibleResult("01000000"); err == nil {
		t.Fatalf("truncated result payload must return an error")
	}
}

func TestIntXIs8ByteSafeBoundaries(t *testing.T) {
	// The 8-byte safety check protects call paths that must match validator int64 boundaries exactly.
	min := new(big.Int).Lsh(big.NewInt(-1), 63)
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 63), big.NewInt(1))

	if !NewIntX(min).Is8ByteSafe() || !NewIntX(max).Is8ByteSafe() || !IntXFromInt64(0).Is8ByteSafe() {
		t.Fatalf("int64 boundary values must be safe")
	}
	if NewIntX(new(big.Int).Add(max, big.NewInt(1))).Is8ByteSafe() {
		t.Fatalf("max+1 must not fit into int64")
	}
	if NewIntX(new(big.Int).Sub(min, big.NewInt(1))).Is8ByteSafe() {
		t.Fatalf("min-1 must not fit into int64")
	}
}

func vectorSeriesInfo() SeriesInfo {
	return SeriesInfo{
		MaxMint:   3,
		MaxSupply: 9,
		Owner:     repeatedBytes32(0x33),
		Metadata:  []byte{0xaa, 0xbb},
		ROM:       VMStructSchema{},
		RAM:       VMStructSchema{},
	}
}

func repeatedBytes32(value byte) Bytes32 {
	var out Bytes32
	for i := range out {
		out[i] = value
	}
	return out
}

func hexString(data []byte) string {
	const alphabet = "0123456789ABCDEF"
	out := make([]byte, len(data)*2)
	for i, b := range data {
		out[i*2] = alphabet[b>>4]
		out[i*2+1] = alphabet[b&0x0f]
	}
	return string(out)
}
