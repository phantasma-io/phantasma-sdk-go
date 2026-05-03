package carbon

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func TestTokenSchemasFromJSON(t *testing.T) {
	jsonSchema := `{
		"seriesMetadata": [],
		"rom": [
			{"name":"name","type":"String"},
			{"name":"description","type":"String"},
			{"name":"imageURL","type":"String"},
			{"name":"infoURL","type":"String"},
			{"name":"royalties","type":"Int32"}
		],
		"ram": []
	}`

	schemas, err := TokenSchemasFromJSON(jsonSchema)
	if err != nil {
		t.Fatalf("JSON schemas rejected: %v", err)
	}
	if err := VerifyTokenSchemas(schemas); err != nil {
		t.Fatalf("parsed schemas failed verification: %v", err)
	}
	if schemas.RAM.Flags != VMStructFlagsDynamicExtras {
		t.Fatalf("empty RAM field list must preserve dynamic extras, got %d", schemas.RAM.Flags)
	}

	wantHex := strings.ToUpper(hex.EncodeToString(SerializeTokenSchemas(schemas)))
	if got := SerializeTokenSchemasHex(schemas); got != wantHex {
		t.Fatalf("schema hex mismatch: want %s got %s", wantHex, got)
	}

	parsedFields, err := ParseTokenSchemasJSON(jsonSchema)
	if err != nil {
		t.Fatalf("schema JSON parse failed: %v", err)
	}
	roundTripJSON, err := json.Marshal(parsedFields)
	if err != nil {
		t.Fatalf("schema JSON marshal failed: %v", err)
	}
	if !strings.Contains(string(roundTripJSON), `"type":"String"`) || !strings.Contains(string(roundTripJSON), `"seriesMetadata"`) {
		t.Fatalf("public schema JSON must use SDK field/type names, got %s", roundTripJSON)
	}

	if _, err := TokenSchemasFromJSON(`{"seriesMetadata":[],"rom":[]}`); err == nil {
		t.Fatalf("missing RAM array must be rejected")
	}
	if _, err := TokenSchemasFromJSON(`{"seriesMetadata":[],"rom":[{"name":"name","type":"Nope"}],"ram":[]}`); err == nil {
		t.Fatalf("unknown VM type must be rejected")
	}
}

func TestCarbonReaderRejectsLengthBeyondRemainingBytes(t *testing.T) {
	if _, err := ParseMintNonFungibleResult(9, "FFFFFF7F"); err == nil || !strings.Contains(err.Error(), "exceeds remaining bytes") {
		t.Fatalf("oversized result length must be rejected before allocation, got %v", err)
	}
}

func TestCarbonConfigAndMarketStructuresRoundTrip(t *testing.T) {
	if ModuleIDPhantasmaVM != 2 || MarketMethodGetTokenListingInfoByID != 8 {
		t.Fatalf("module or market method ids changed")
	}

	tests := []struct {
		name string
		blob Blob
	}{
		{
			name: "chain_config",
			blob: &ChainConfig{
				Version:         1,
				Reserved1:       2,
				Reserved2:       3,
				Reserved3:       4,
				AllowedTxTypes:  0x0a0b0c0d,
				ExpiryWindow:    0x11223344,
				BlockRateTarget: 0x55667788,
			},
		},
		{
			name: "gas_config",
			blob: &GasConfig{
				Version:                 1,
				MaxNameLength:           32,
				MaxTokenSymbolLength:    16,
				FeeShift:                4,
				MaxStructureSize:        2048,
				FeeMultiplier:           1000,
				GasTokenID:              1,
				DataTokenID:             2,
				MinimumGasOffer:         3,
				DataEscrowPerRow:        4,
				GasFeeTransfer:          5,
				GasFeeQuery:             6,
				GasFeeCreateTokenBase:   7,
				GasFeeCreateTokenSymbol: 8,
				GasFeeCreateTokenSeries: 9,
				GasFeePerByte:           10,
				GasFeeRegisterName:      11,
				GasBurnRatioMul:         12,
				GasBurnRatioShift:       2,
			},
		},
		{
			name: "market_config",
			blob: &MarketConfig{
				MinimumListingTime: 1000,
				MaximumListingTime: 2000,
				DelistingGrace:     3000,
				Flags:              MarketConfigFlagsPriceRequired | MarketConfigFlagsCanCancelEarly,
			},
		},
		{
			name: "token_listing",
			blob: &TokenListing{
				Type:         ListingTypeFixedPrice,
				Seller:       repeatedBytes32(0x11),
				QuoteTokenID: 9,
				Price:        IntXFromInt64(100),
				StartDate:    123,
				EndDate:      456,
			},
		},
		{
			name: "market_sell_args",
			blob: &MarketSellTokenArgs{
				From:         repeatedBytes32(0x22),
				TokenID:      9,
				InstanceID:   7,
				QuoteTokenID: 1,
				Price:        IntXFromInt64(100),
				EndDate:      456,
			},
		},
		{
			name: "market_buy_by_id_args",
			blob: &MarketBuyTokenByIDArgs{
				From:       repeatedBytes32(0x33),
				Symbol:     MustSmallString("NFT"),
				InstanceID: NewVMDynamicVariable(VMTypeInt256, IntXFromInt64(7).BigInt()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := Serialize(tt.blob)
			decoded, err := Deserialize(encoded, tt.blob)
			if err != nil {
				t.Fatalf("roundtrip failed: %v", err)
			}
			if got := Serialize(decoded); hex.EncodeToString(got) != hex.EncodeToString(encoded) {
				t.Fatalf("roundtrip bytes changed: want %x got %x", encoded, got)
			}
		})
	}
}
