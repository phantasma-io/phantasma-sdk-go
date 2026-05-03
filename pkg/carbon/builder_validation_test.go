package carbon

import (
	"strings"
	"testing"
)

func TestTokenInfoBuilderValidation(t *testing.T) {
	// Token builders should reject malformed public inputs before they can create invalid Carbon payloads.
	metadata := validTokenMetadata(t)
	creator := EmptyBytes32
	maxSupply := IntXFromInt64(0)

	tests := []struct {
		name    string
		symbol  string
		wantErr string
	}{
		{name: "empty", symbol: "", wantErr: "empty string"},
		{name: "too_long", symbol: strings.Repeat("A", 256), wantErr: "too long"},
		{name: "digit", symbol: "AB1", wantErr: "only A-Z"},
		{name: "lowercase", symbol: "AbC", wantErr: "only A-Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildTokenInfo(tt.symbol, maxSupply, false, 0, creator, metadata, nil)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected %q error, got %v", tt.wantErr, err)
			}
		})
	}

	if _, err := BuildTokenInfo("TEST", maxSupply, false, 0, creator, nil, nil); err == nil {
		t.Fatalf("missing metadata must be rejected")
	}

	_, err := BuildTokenInfo("FUNGIBLE", maxSupply, false, 8, creator, metadata, nil)
	if err != nil {
		t.Fatalf("valid fungible token rejected: %v", err)
	}

	_, err = BuildTokenInfo("NFT", MustIntXFromString("9223372036854775808"), true, 0, creator, metadata, SerializeTokenSchemas(PrepareStandardTokenSchemas(false)))
	if err == nil || !strings.Contains(err.Error(), "NFT maximum supply") {
		t.Fatalf("large NFT supply must be rejected, got %v", err)
	}

	_, err = BuildTokenInfo("NFT", maxSupply, true, 0, creator, metadata, nil)
	if err == nil || !strings.Contains(err.Error(), "token schemas") {
		t.Fatalf("NFT without schemas must be rejected, got %v", err)
	}
}

func TestTokenMetadataValidation(t *testing.T) {
	// Token-level metadata is mandatory and icon data must be a supported base64 image data URI.
	if _, err := BuildTokenMetadata(nil); err == nil {
		t.Fatalf("missing metadata must be rejected")
	}
	if _, err := BuildTokenMetadata(map[string]string{
		"name":        "My token",
		"icon":        "not-a-data-uri",
		"url":         "http://example.com",
		"description": "description",
	}); err == nil {
		t.Fatalf("invalid icon must be rejected")
	}
	if _, err := BuildTokenMetadata(map[string]string{
		"name":        "My token",
		"icon":        "data:image/png;base64,!!!!",
		"url":         "http://example.com",
		"description": "description",
	}); err == nil {
		t.Fatalf("invalid base64 icon must be rejected")
	}
}

func TestTokenSchemasVerification(t *testing.T) {
	// Standard schema verification catches missing, mistyped and case-mismatched NFT metadata fields.
	if err := VerifyTokenSchemas(PrepareStandardTokenSchemas(false)); err != nil {
		t.Fatalf("standard schemas rejected: %v", err)
	}

	missingStandard := TokenSchemas{
		SeriesMetadata: VMStructSchema{Fields: cloneSchemaFields(standardSeriesFields)},
		ROM:            VMStructSchema{Fields: cloneSchemaFields(standardNFTFields)},
		RAM:            VMStructSchema{},
	}
	if err := VerifyTokenSchemas(missingStandard); err == nil || !strings.Contains(err.Error(), "mandatory metadata field not found: name") {
		t.Fatalf("missing standard metadata must be reported, got %v", err)
	}

	typeMismatch := TokenSchemas{
		SeriesMetadata: VMStructSchema{Fields: []VMNamedVariableSchema{MustVMNamedVariableSchema("name", VMTypeInt32)}},
		ROM:            VMStructSchema{},
		RAM:            VMStructSchema{},
	}
	if err := VerifyTokenSchemas(typeMismatch); err == nil || !strings.Contains(err.Error(), "type mismatch for name field") {
		t.Fatalf("type mismatch must be reported, got %v", err)
	}

	caseMismatch := TokenSchemas{
		SeriesMetadata: VMStructSchema{Fields: []VMNamedVariableSchema{MustVMNamedVariableSchema("Name", VMTypeString)}},
		ROM:            VMStructSchema{},
		RAM:            VMStructSchema{},
	}
	if err := VerifyTokenSchemas(caseMismatch); err == nil || !strings.Contains(err.Error(), "case mismatch for name field") {
		t.Fatalf("case mismatch must be reported, got %v", err)
	}
}

func TestMetadataBuildersReturnErrorsForInvalidInput(t *testing.T) {
	schemas := PrepareStandardTokenSchemas(false)

	// Public builders return errors for user-supplied metadata problems; Must variants are the only panic path.
	if _, err := BuildNFTRom(schemas.ROM, IntXFromInt64(1).BigInt(), nil); err == nil || !strings.Contains(err.Error(), "metadata field \"name\" is mandatory") {
		t.Fatalf("missing NFT metadata must be returned as an error, got %v", err)
	}
	if _, err := BuildNFTRom(schemas.ROM, IntXFromInt64(1).BigInt(), []MetadataField{{Name: "Name", Value: "wrong case"}}); err == nil || !strings.Contains(err.Error(), "incorrect case") {
		t.Fatalf("case-mismatched metadata must be returned as an error, got %v", err)
	}
	if _, err := BuildTokenSeriesMetadata(schemas.SeriesMetadata, IntXFromInt64(1).BigInt(), []MetadataField{{Name: "rom", Value: 7}}); err == nil || !strings.Contains(err.Error(), "must be bytes or hex string") {
		t.Fatalf("invalid rom type must be returned as an error, got %v", err)
	}
	if _, err := BuildPhantasmaNFTRom(schemas.ROM, append(standardNFTMetadata(), MetadataField{Name: StandardMetaID, Value: "reserved"})); err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("reserved public mint metadata must be returned as an error, got %v", err)
	}
}

func TestMustMetadataBuildersPanicOnInvalidInput(t *testing.T) {
	schemas := PrepareStandardTokenSchemas(false)

	// Must builders are useful for constants/tests but deliberately panic instead of returning errors.
	assertPanics(t, func() { MustBuildNFTRom(schemas.ROM, IntXFromInt64(1).BigInt(), nil) })
	assertPanics(t, func() { MustBuildPhantasmaNFTRom(schemas.ROM, nil) })
}

func TestRequiredPhantasmaIDsRejectNil(t *testing.T) {
	schemas := PrepareStandardTokenSchemas(false)

	if _, err := BuildSeriesInfo(nil, 0, 0, EmptyBytes32); err == nil || !strings.Contains(err.Error(), "phantasmaSeriesID") {
		t.Fatalf("nil series id must be rejected, got %v", err)
	}
	if _, err := BuildTokenSeriesMetadata(schemas.SeriesMetadata, nil, nil); err == nil || !strings.Contains(err.Error(), "phantasmaSeriesID") {
		t.Fatalf("nil series metadata id must be rejected, got %v", err)
	}
	if _, err := BuildNFTRom(schemas.ROM, nil, standardNFTMetadata()); err == nil || !strings.Contains(err.Error(), "phantasmaNFTID") {
		t.Fatalf("nil NFT id must be rejected, got %v", err)
	}
	if _, err := BuildMintPhantasmaNonFungibleSingleTx(9, nil, EmptyBytes32, EmptyBytes32, nil, nil, MintNFTFeeOptions{}, 0, 0); err == nil || !strings.Contains(err.Error(), "phantasmaSeriesID") {
		t.Fatalf("nil mint series id must be rejected, got %v", err)
	}
}

func validTokenMetadata(t *testing.T) []byte {
	t.Helper()
	metadata, err := BuildTokenMetadata(map[string]string{
		"name":        "My test token!",
		"icon":        samplePNGIconDataURI,
		"url":         "http://example.com",
		"description": "My test token description",
	})
	if err != nil {
		t.Fatal(err)
	}
	return metadata
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	fn()
}
