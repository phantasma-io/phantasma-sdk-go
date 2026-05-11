package carbon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validatorInt256FixtureSHA256 = "785a0ec3a9f7c78c15a6e76029e2f34b37fec7626b1ace62d9b668529e686f03"

type validatorInt256FixtureSet struct {
	Int256               []validatorInt256Row       `json:"int256"`
	IntX                 []validatorIntXRow         `json:"intx"`
	VMDynamicInt256      []validatorDynamicIntRow   `json:"vmDynamicInt256"`
	VMDynamicInt256Array []validatorDynamicArrayRow `json:"vmDynamicInt256Array"`
	MetadataStructs      []validatorMetadataRow     `json:"metadataStructs"`
	TokenInfo            []validatorTokenInfoRow    `json:"tokenInfo"`
	SeriesInfo           []validatorSeriesInfoRow   `json:"seriesInfo"`
}

type validatorInt256Row struct {
	ID                string `json:"id"`
	SourceDec         string `json:"sourceDec"`
	ReadBackSignedDec string `json:"readBackSignedDec"`
	WireHex           string `json:"wireHex"`
	WireLen           int    `json:"wireLen"`
}

type validatorIntXRow struct {
	ID          string `json:"id"`
	SourceDec   string `json:"sourceDec"`
	ReadBackDec string `json:"readBackDec"`
	WireHex     string `json:"wireHex"`
	WireLen     int    `json:"wireLen"`
}

type validatorDynamicIntRow struct {
	ID        string `json:"id"`
	SourceDec string `json:"sourceDec"`
	WireHex   string `json:"wireHex"`
	WireLen   int    `json:"wireLen"`
}

type validatorDynamicArrayRow struct {
	ID      string   `json:"id"`
	Values  []string `json:"values"`
	WireHex string   `json:"wireHex"`
	WireLen int      `json:"wireLen"`
}

type validatorMetadataRow struct {
	ID      string `json:"id"`
	Shape   string `json:"shape"`
	IDec    string `json:"_iDec"`
	Mode    int8   `json:"mode"`
	RomHex  string `json:"romHex"`
	WireHex string `json:"wireHex"`
	WireLen int    `json:"wireLen"`
}

type validatorTokenInfoRow struct {
	ID           string `json:"id"`
	MaxSupplyDec string `json:"maxSupplyDec"`
	Flags        byte   `json:"flags"`
	Decimals     byte   `json:"decimals"`
	Symbol       string `json:"symbol"`
	MetadataHex  string `json:"metadataHex"`
	WireHex      string `json:"wireHex"`
	WireLen      int    `json:"wireLen"`
}

type validatorSeriesInfoRow struct {
	ID          string `json:"id"`
	MaxMint     uint32 `json:"maxMint"`
	MaxSupply   uint32 `json:"maxSupply"`
	MetadataHex string `json:"metadataHex"`
	WireHex     string `json:"wireHex"`
	WireLen     int    `json:"wireLen"`
}

func TestValidatorInt256Fixtures(t *testing.T) {
	fixtures := loadValidatorInt256Fixtures(t)

	t.Run("raw_int256", func(t *testing.T) {
		// Raw Writer/Reader int256 coverage includes unsigned validator inputs
		// whose canonical signed readback intentionally wraps at 256 bits.
		for _, row := range fixtures.Int256 {
			w := NewWriter()
			w.WriteBigInt(parseBigInt(t, row.SourceDec))
			assertValidatorFixtureBytes(t, row.ID, row.WireHex, row.WireLen, w.Bytes())

			r := NewReader(w.Bytes())
			requireBigIntDecimal(t, row.ID, r.ReadBigInt(), row.ReadBackSignedDec)
			r.AssertEOF()
		}
	})

	t.Run("intx", func(t *testing.T) {
		// IntX is the public variable-width integer wrapper used in token and
		// market payloads; it must match validator packing for 64-bit and wide values.
		for _, row := range fixtures.IntX {
			value := NewIntX(parseBigInt(t, row.SourceDec))
			w := NewWriter()
			value.WriteCarbon(w)
			assertValidatorFixtureBytes(t, row.ID, row.WireHex, row.WireLen, w.Bytes())

			var decoded IntX
			r := NewReader(w.Bytes())
			decoded.ReadCarbon(r)
			r.AssertEOF()
			requireBigIntDecimal(t, row.ID, decoded.BigInt(), row.ReadBackDec)
		}
	})

	t.Run("dynamic_int256", func(t *testing.T) {
		// Dynamic metadata fields carry the VM type byte before the same int256
		// payload, so they are tested separately from the primitive writer.
		for _, row := range fixtures.VMDynamicInt256 {
			value := NewVMDynamicVariable(VMTypeInt256, parseBigInt(t, row.SourceDec))
			encoded := Serialize(&value)
			assertValidatorFixtureBytes(t, row.ID, row.WireHex, row.WireLen, encoded)

			var decoded VMDynamicVariable
			r := NewReader(encoded)
			decoded.ReadCarbon(r)
			r.AssertEOF()
			requireBigIntDecimal(t, row.ID, decoded.Data.(*big.Int), expectedSignedInt256Decimal(t, row.WireHex[2:]))
		}
	})

	t.Run("dynamic_int256_array", func(t *testing.T) {
		// Array_Int256 values are used by VM metadata schemas; this pins length
		// encoding plus every element's validator int256 packing.
		for _, row := range fixtures.VMDynamicInt256Array {
			values := make([]*big.Int, len(row.Values))
			for i, value := range row.Values {
				values[i] = parseBigInt(t, value)
			}
			dynamic := NewVMDynamicVariable(VMTypeArrayInt256, values)
			encoded := Serialize(&dynamic)
			assertValidatorFixtureBytes(t, row.ID, row.WireHex, row.WireLen, encoded)

			var decoded VMDynamicVariable
			r := NewReader(encoded)
			decoded.ReadCarbon(r)
			r.AssertEOF()
			decodedValues := decoded.Data.([]*big.Int)
			if len(decodedValues) != len(values) {
				t.Fatalf("%s decoded array length mismatch: want %d got %d", row.ID, len(values), len(decodedValues))
			}
		}
	})

	t.Run("metadata_structs", func(t *testing.T) {
		// The default NFT and series metadata structs are the surfaces where
		// Phantasma ID int256 values enter token/module calls.
		for _, row := range fixtures.MetadataStructs {
			structure := metadataFixtureStruct(t, row)
			encoded := Serialize(&structure)
			assertValidatorFixtureBytes(t, row.ID, row.WireHex, row.WireLen, encoded)

			var decoded VMDynamicStruct
			r := NewReader(encoded)
			decoded.ReadCarbon(r)
			r.AssertEOF()
			assertValidatorFixtureBytes(t, row.ID+" roundtrip", row.WireHex, row.WireLen, Serialize(&decoded))
		}
	})

	t.Run("token_info", func(t *testing.T) {
		// TokenInfo stores max supply as IntX and is signed inside create-token
		// transactions, so exact metadata bytes and owner fields are pinned.
		for _, row := range fixtures.TokenInfo {
			info := TokenInfo{
				MaxSupply:    NewIntX(parseBigInt(t, row.MaxSupplyDec)),
				Flags:        TokenFlags(row.Flags),
				Decimals:     row.Decimals,
				Owner:        tokenInfoFixtureOwner(t, row.ID),
				Symbol:       MustSmallString(row.Symbol),
				Metadata:     MustDecodeHex(row.MetadataHex),
				TokenSchemas: nil,
			}
			encoded := Serialize(&info)
			assertValidatorFixtureBytes(t, row.ID, row.WireHex, row.WireLen, encoded)

			var decoded TokenInfo
			r := NewReader(encoded)
			decoded.ReadCarbon(r)
			r.AssertEOF()
			assertValidatorFixtureBytes(t, row.ID+" roundtrip", row.WireHex, row.WireLen, Serialize(&decoded))
		}
	})

	t.Run("series_info", func(t *testing.T) {
		// SeriesInfo combines owner, max mint/supply, metadata int256 values,
		// and empty ROM/RAM schemas in the create-series payload.
		for _, row := range fixtures.SeriesInfo {
			info := SeriesInfo{
				MaxMint:   row.MaxMint,
				MaxSupply: row.MaxSupply,
				Owner:     seriesInfoFixtureOwner(t, row.ID),
				Metadata:  MustDecodeHex(row.MetadataHex),
				ROM:       VMStructSchema{},
				RAM:       VMStructSchema{},
			}
			encoded := Serialize(&info)
			assertValidatorFixtureBytes(t, row.ID, row.WireHex, row.WireLen, encoded)

			var decoded SeriesInfo
			r := NewReader(encoded)
			decoded.ReadCarbon(r)
			r.AssertEOF()
			assertValidatorFixtureBytes(t, row.ID+" roundtrip", row.WireHex, row.WireLen, Serialize(&decoded))
		}
	})
}

func loadValidatorInt256Fixtures(t *testing.T) validatorInt256FixtureSet {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "validator_int256_fixtures.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != validatorInt256FixtureSHA256 {
		t.Fatalf("validator_int256_fixtures.json hash mismatch: want %s got %s", validatorInt256FixtureSHA256, got)
	}

	var fixtures validatorInt256FixtureSet
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	return fixtures
}

func metadataFixtureStruct(t *testing.T, row validatorMetadataRow) VMDynamicStruct {
	t.Helper()
	fields := []VMNamedDynamicVariable{
		MustVMNamedDynamicVariable(StandardMetaID, VMTypeInt256, parseBigInt(t, row.IDec)),
	}
	switch row.Shape {
	case "nft-default":
		fields = append(fields, MustVMNamedDynamicVariable("rom", VMTypeBytes, MustDecodeHex(row.RomHex)))
	case "series-default":
		fields = append(fields,
			MustVMNamedDynamicVariable("mode", VMTypeInt8, row.Mode),
			MustVMNamedDynamicVariable("rom", VMTypeBytes, MustDecodeHex(row.RomHex)),
		)
	default:
		t.Fatalf("unsupported metadata fixture shape: %s", row.Shape)
	}
	return VMDynamicStruct{Fields: fields}
}

func tokenInfoFixtureOwner(t *testing.T, id string) Bytes32 {
	t.Helper()
	switch id {
	case "fungible_zero_supply":
		return sequentialBytes32(0x10)
	case "big_fungible_u64max_supply":
		return sequentialBytes32(0x20)
	default:
		t.Fatalf("unsupported token info fixture owner: %s", id)
		return EmptyBytes32
	}
}

func seriesInfoFixtureOwner(t *testing.T, id string) Bytes32 {
	t.Helper()
	switch id {
	case "series_zero_metaid":
		return sequentialBytes32(0x30)
	case "series_problematic_metaid":
		return sequentialBytes32(0x40)
	default:
		t.Fatalf("unsupported series info fixture owner: %s", id)
		return EmptyBytes32
	}
}

func sequentialBytes32(start byte) Bytes32 {
	var out Bytes32
	for i := range out {
		out[i] = start + byte(i)
	}
	return out
}

func expectedSignedInt256Decimal(t *testing.T, wireHex string) string {
	t.Helper()
	r := NewReader(MustDecodeHex(wireHex))
	return r.ReadBigInt().String()
}

func requireBigIntDecimal(t *testing.T, id string, got *big.Int, expected string) {
	t.Helper()
	want := parseBigInt(t, expected)
	if got.Cmp(want) != 0 {
		t.Fatalf("%s decimal mismatch: want %s got %s", id, want, got)
	}
}

func assertValidatorFixtureBytes(t *testing.T, id string, expectedHex string, expectedLen int, got []byte) {
	t.Helper()
	if len(got) != expectedLen {
		t.Fatalf("%s wire length mismatch: want %d got %d", id, expectedLen, len(got))
	}
	if actual := strings.ToUpper(hex.EncodeToString(got)); actual != expectedHex {
		t.Fatalf("%s wire mismatch:\nwant %s\n got %s", id, expectedHex, actual)
	}
}
