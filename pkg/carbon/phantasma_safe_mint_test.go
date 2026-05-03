package carbon

import (
	"math/big"
	"strings"
	"testing"
)

func TestPhantasmaNFTRomBuilderSerializesPublicMintPayload(t *testing.T) {
	// Public mint ROM serialization must omit chain-owned deterministic fields while preserving user metadata.
	schemas := PrepareStandardTokenSchemas(false)
	rom, err := BuildPhantasmaNFTRom(schemas.ROM, standardNFTMetadata())
	if err != nil {
		t.Fatal(err)
	}
	publicSchema := BuildPhantasmaNFTPublicMintSchema(schemas.ROM)

	var romStruct VMDynamicStruct
	reader := NewReader(rom)
	romStruct.ReadWithSchema(&publicSchema, reader)
	reader.AssertEOF()

	// Deterministic Phantasma mints let the chain own `_i` and nested `rom`; callers only provide the public payload.
	if got := romStruct.Get("name"); got == nil || got.Data != "My NFT #1" {
		t.Fatalf("unexpected public name field: %#v", got)
	}
	if got := romStruct.Get("description"); got == nil || got.Data != "This is my first NFT!" {
		t.Fatalf("unexpected public description field: %#v", got)
	}
	if romStruct.Get(StandardMetaID) != nil {
		t.Fatalf("public mint ROM must not expose %q", StandardMetaID)
	}
	if romStruct.Get("rom") != nil {
		t.Fatalf("public mint ROM must not expose nested rom")
	}
}

func TestPhantasmaNFTRomBuilderRejectsReservedFields(t *testing.T) {
	schemas := PrepareStandardTokenSchemas(false)

	// Reserved field names are rejected before schema writing so caller input cannot override chain-owned ids.
	for _, fieldName := range []string{StandardMetaID, "rom", strings.ToUpper(StandardMetaID), "ROM"} {
		t.Run(fieldName, func(t *testing.T) {
			metadata := append(standardNFTMetadata(), MetadataField{Name: fieldName, Value: []byte{0x01}})
			if _, err := BuildPhantasmaNFTRom(schemas.ROM, metadata); err == nil {
				t.Fatalf("reserved metadata field %q must be rejected", fieldName)
			}
		})
	}
}

func TestMintPhantasmaNonFungibleSingleTxBuildsDeterministicCallArgs(t *testing.T) {
	// Single-mint convenience should produce the same Token.Call ABI as the multi-mint helper.
	schemas := PrepareStandardTokenSchemas(false)
	rom, err := BuildPhantasmaNFTRom(schemas.ROM, standardNFTMetadata())
	if err != nil {
		t.Fatal(err)
	}
	sender := repeatedBytes32(0x11)
	receiver := repeatedBytes32(0x22)

	tx, err := BuildMintPhantasmaNonFungibleSingleTx(
		42,
		big.NewInt(777),
		sender,
		receiver,
		rom,
		nil,
		DefaultMintNFTFeeOptions(),
		123,
		999,
	)
	if err != nil {
		t.Fatal(err)
	}

	// The single-item helper is a thin package over the Token.Call ABI used by the C#/TS SDKs.
	if tx.Type != TxTypeCall {
		t.Fatalf("unexpected tx type: %d", tx.Type)
	}
	if tx.Expiry != 999 || tx.MaxData != 123 || tx.GasFrom != sender {
		t.Fatalf("unexpected tx envelope: %#v", tx)
	}

	call, ok := tx.Msg.(*TxMsgCall)
	if !ok {
		t.Fatalf("expected call payload, got %T", tx.Msg)
	}
	if call.ModuleID != uint32(ModuleIDToken) || call.MethodID != uint32(TokenMethodMintPhantasmaNonFungible) {
		t.Fatalf("unexpected call target: module=%d method=%d", call.ModuleID, call.MethodID)
	}

	var args MintPhantasmaNonFungibleArgs
	Deserialize(call.Args, &args)
	if args.TokenID != 42 || args.Address != receiver || len(args.Tokens) != 1 {
		t.Fatalf("unexpected deterministic mint args: %#v", args)
	}
	if args.Tokens[0].PhantasmaSeriesID.String() != "777" {
		t.Fatalf("unexpected Phantasma series id: %s", args.Tokens[0].PhantasmaSeriesID.String())
	}
	if string(args.Tokens[0].ROM) != string(rom) || len(args.Tokens[0].RAM) != 0 {
		t.Fatalf("unexpected deterministic mint payload")
	}
}

func standardNFTMetadata() []MetadataField {
	return []MetadataField{
		{Name: "name", Value: "My NFT #1"},
		{Name: "description", Value: "This is my first NFT!"},
		{Name: "imageURL", Value: "images-assets.nasa.gov/image/PIA13227/PIA13227~orig.jpg"},
		{Name: "infoURL", Value: "https://images.nasa.gov/details/PIA13227"},
		{Name: "royalties", Value: int32(10_000_000)},
	}
}
