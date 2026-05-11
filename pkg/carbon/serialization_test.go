package carbon

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
)

func TestCarbonVectors(t *testing.T) {
	// Shared TS/C++ vectors prove the primitive reader/writer stays byte-for-byte compatible across SDKs.
	rows := readVectorRows(t)
	for _, row := range rows {
		row := row
		t.Run(row.kind+"_"+row.hex, func(t *testing.T) {
			switch row.kind {
			case "U8":
				v := byte(parseUint(t, row.value, 8))
				assertWrite(t, row.hex, func(w *Writer) { w.Write1(v) })
				assertRead(t, row.hex, v, func(r *Reader) byte { return r.Read1() })
			case "I16":
				v := int16(parseInt(t, row.value, 16))
				assertWrite(t, row.hex, func(w *Writer) { w.Write2(v) })
				assertRead(t, row.hex, v, func(r *Reader) int16 { return r.Read2() })
			case "I32":
				v := int32(parseInt(t, row.value, 32))
				assertWrite(t, row.hex, func(w *Writer) { w.Write4(v) })
				assertRead(t, row.hex, v, func(r *Reader) int32 { return r.Read4() })
			case "U32":
				v := uint32(parseUint(t, row.value, 32))
				assertWrite(t, row.hex, func(w *Writer) { w.Write4U(v) })
				assertRead(t, row.hex, v, func(r *Reader) uint32 { return r.Read4U() })
			case "I64":
				v := parseInt(t, row.value, 64)
				assertWrite(t, row.hex, func(w *Writer) { w.Write8(v) })
				assertRead(t, row.hex, v, func(r *Reader) int64 { return r.Read8() })
			case "U64":
				v := parseUint(t, row.value, 64)
				assertWrite(t, row.hex, func(w *Writer) { w.Write8U(v) })
				assertRead(t, row.hex, v, func(r *Reader) uint64 { return r.Read8U() })
			case "FIX16":
				v := MustBytes16FromHex(row.value)
				assertWrite(t, row.hex, func(w *Writer) { w.Write16(v) })
				assertRead(t, row.hex, v, func(r *Reader) Bytes16 { return r.Read16() })
			case "FIX32":
				v := MustBytes32FromHex(row.value)
				assertWrite(t, row.hex, func(w *Writer) { w.Write32(v) })
				assertRead(t, row.hex, v, func(r *Reader) Bytes32 { return r.Read32() })
			case "FIX64":
				v := MustBytes64FromHex(row.value)
				assertWrite(t, row.hex, func(w *Writer) { w.Write64(v) })
				assertRead(t, row.hex, v, func(r *Reader) Bytes64 { return r.Read64() })
			case "SZ":
				assertWrite(t, row.hex, func(w *Writer) { w.WriteStringZ(row.value) })
				assertRead(t, row.hex, row.value, func(r *Reader) string { return r.ReadStringZ() })
			case "ARRSZ":
				values := strings.Split(row.value, ",")
				assertWrite(t, row.hex, func(w *Writer) { w.WriteStringZArray(values) })
				assertRead(t, row.hex, values, func(r *Reader) []string { return r.ReadStringZArray() })
			case "ARR8":
				values := parseInt8CSV(t, row.value)
				assertWrite(t, row.hex, func(w *Writer) { w.WriteInt8Array(values) })
				assertRead(t, row.hex, values, func(r *Reader) []int8 { return r.ReadInt8Array() })
			case "ARR16":
				values := parseInt16CSV(t, row.value)
				assertWrite(t, row.hex, func(w *Writer) { w.WriteInt16Array(values) })
				assertRead(t, row.hex, values, func(r *Reader) []int16 { return r.ReadInt16Array() })
			case "ARR32":
				values := parseInt32CSV(t, row.value)
				assertWrite(t, row.hex, func(w *Writer) { w.WriteInt32Array(values) })
				assertRead(t, row.hex, values, func(r *Reader) []int32 { return r.ReadInt32Array() })
			case "ARR64":
				values := parseInt64CSV(t, row.value)
				assertWrite(t, row.hex, func(w *Writer) { w.WriteInt64Array(values) })
				assertRead(t, row.hex, values, func(r *Reader) []int64 { return r.ReadInt64Array() })
			case "ARRU64":
				values := parseUint64CSV(t, row.value)
				assertWrite(t, row.hex, func(w *Writer) { w.WriteUint64Array(values) })
				assertRead(t, row.hex, values, func(r *Reader) []uint64 { return r.ReadUint64Array() })
			case "ARRBYTES-1D":
				value := decodeHex(t, row.value)
				assertWrite(t, row.hex, func(w *Writer) { w.WriteByteArray(value) })
				assertReadBytes(t, row.hex, value, func(r *Reader) []byte { return r.ReadByteArray() })
			case "ARRBYTES-2D":
				values := parseByteArrays(t, row.value)
				assertWrite(t, row.hex, func(w *Writer) { w.WriteByteArrays(values) })
				assertReadByteArrays(t, row.hex, values, func(r *Reader) [][]byte { return r.ReadByteArrays() })
			case "BI":
				writeValue := parseBigInt(t, row.value)
				readValue := parseBigInt(t, row.readValue())
				assertWrite(t, row.hex, func(w *Writer) { w.WriteBigInt(writeValue) })
				assertReadBigInt(t, row.hex, readValue, func(r *Reader) *big.Int { return r.ReadBigInt() })
			case "INTX":
				writeValue := MustIntXFromString(row.value)
				readValue := MustIntXFromString(row.readValue())
				assertWrite(t, row.hex, func(w *Writer) { writeValue.WriteCarbon(w) })
				assertReadIntX(t, row.hex, readValue, func(r *Reader) IntX {
					var out IntX
					out.ReadCarbon(r)
					return out
				})
			case "ARRBI":
				values := parseBigIntCSV(t, row.value)
				assertWrite(t, row.hex, func(w *Writer) { w.WriteBigIntArray(values) })
				assertReadBigIntSlice(t, row.hex, values, func(r *Reader) []*big.Int { return r.ReadBigIntArray() })
			case "TX1", "TX-CREATE-TOKEN", "TX-CREATE-TOKEN-SERIES", "TX-MINT-NON-FUNGIBLE":
				assertCarbonRoundTrip(t, row.hex, &TxMsg{})
			case "TX2":
				assertCarbonRoundTrip(t, row.hex, &SignedTxMsg{})
			case "VMSTRUCT01":
				assertCarbonRoundTrip(t, row.hex, &TokenSchemas{})
			case "VMSTRUCT02":
				assertCarbonRoundTrip(t, row.hex, &VMDynamicStruct{})
			default:
				t.Fatalf("unhandled vector kind %q", row.kind)
			}
		})
	}
}

func TestTx1VectorFields(t *testing.T) {
	// Decode the unsigned transaction vector to ensure round-trip tests also preserve semantic fields.
	row := vectorByKind(t, "TX1")
	var msg TxMsg
	readCarbon(t, row.hex, &msg)

	if msg.Type != TxTypeTransferFungible {
		t.Fatalf("type mismatch: got %d", msg.Type)
	}
	if msg.Expiry != 1759711416000 {
		t.Fatalf("expiry mismatch: got %d", msg.Expiry)
	}
	if msg.MaxGas != 10_000_000 || msg.MaxData != 1000 {
		t.Fatalf("gas/data mismatch: got %d/%d", msg.MaxGas, msg.MaxData)
	}
	if msg.Payload.String() != "test-payload" {
		t.Fatalf("payload mismatch: got %q", msg.Payload.String())
	}
	ft, ok := msg.Msg.(*TxMsgTransferFungible)
	if !ok {
		t.Fatalf("payload type mismatch: got %T", msg.Msg)
	}
	if ft.TokenID != 1 || ft.Amount != 100_000_000 {
		t.Fatalf("transfer mismatch: token %d amount %d", ft.TokenID, ft.Amount)
	}
}

func TestTx2SignedVectorFields(t *testing.T) {
	// The signed transfer vector must preserve the compact single-witness form used by Carbon transfers.
	row := vectorByKind(t, "TX2")
	var msg SignedTxMsg
	readCarbon(t, row.hex, &msg)

	if len(msg.Witnesses) != 1 {
		t.Fatalf("witness count mismatch: got %d", len(msg.Witnesses))
	}
	if msg.Witnesses[0].Address != msg.Msg.GasFrom {
		t.Fatalf("single witness address must match gasFrom")
	}
	if msg.Witnesses[0].Signature == EmptyBytes64 {
		t.Fatalf("signature must not be empty")
	}
}

func TestTokenBuilderVectors(t *testing.T) {
	// Builder vectors cover the transaction helpers most likely to drift from C#/TS/C++ wire format.
	sender := testSenderPublicKey(t)
	schemas := PrepareStandardTokenSchemas(false)
	schemasBytes := SerializeTokenSchemas(schemas)

	t.Run("create_token", func(t *testing.T) {
		metadata, err := BuildTokenMetadata(map[string]string{
			"name":        "My test token!",
			"icon":        samplePNGIconDataURI,
			"url":         "http://example.com",
			"description": "My test token description",
		})
		if err != nil {
			t.Fatal(err)
		}

		tokenInfo, err := BuildTokenInfo("MYNFT", IntXFromInt64(0), true, 0, sender, metadata, schemasBytes)
		if err != nil {
			t.Fatal(err)
		}
		msg := BuildCreateTokenTx(tokenInfo, sender, DefaultCreateTokenFeeOptions(), 100_000_000, 1_759_711_416_000)
		assertSerializedHex(t, vectorByKind(t, "TX-CREATE-TOKEN").hex, &msg)
	})

	t.Run("create_token_series", func(t *testing.T) {
		seriesInfo, err := BuildSeriesInfo(maxUint256(), 0, 0, sender)
		if err != nil {
			t.Fatal(err)
		}
		msg := BuildCreateTokenSeriesTx(^uint64(0), seriesInfo, sender, DefaultCreateSeriesFeeOptions(), 100_000_000, 1_759_711_416_000)
		assertSerializedHex(t, vectorByKind(t, "TX-CREATE-TOKEN-SERIES").hex, &msg)
	})

	t.Run("mint_non_fungible", func(t *testing.T) {
		rom, err := BuildNFTRom(schemas.ROM, maxUint256(), []MetadataField{
			{Name: "name", Value: "My NFT #1"},
			{Name: "description", Value: "This is my first NFT!"},
			{Name: "imageURL", Value: "images-assets.nasa.gov/image/PIA13227/PIA13227~orig.jpg"},
			{Name: "infoURL", Value: "https://images.nasa.gov/details/PIA13227"},
			{Name: "royalties", Value: int32(10_000_000)},
			{Name: "rom", Value: []byte{0x01, 0x42}},
		})
		if err != nil {
			t.Fatal(err)
		}
		msg := BuildMintNonFungibleTx(^uint64(0), ^uint32(0), sender, sender, rom, nil, DefaultMintNFTFeeOptions(), 100_000_000, 1_759_711_416_000)
		assertSerializedHex(t, vectorByKind(t, "TX-MINT-NON-FUNGIBLE").hex, &msg)
	})
}

type vectorRow struct {
	kind   string
	value  string
	hex    string
	fields []string
}

func (r vectorRow) readValue() string {
	if len(r.fields) >= 5 {
		return r.fields[4]
	}
	return r.value
}

func readVectorRows(t *testing.T) []vectorRow {
	t.Helper()
	data, err := os.ReadFile("testdata/carbon_vectors.tsv")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	rows := make([]vectorRow, 0, len(lines))
	for _, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			t.Fatalf("invalid vector row: %q", line)
		}
		rows = append(rows, vectorRow{
			kind:   fields[0],
			value:  fields[1],
			hex:    fields[2],
			fields: fields,
		})
	}
	return rows
}

func vectorByKind(t *testing.T, kind string) vectorRow {
	t.Helper()
	for _, row := range readVectorRows(t) {
		if row.kind == kind {
			return row
		}
	}
	t.Fatalf("vector %q not found", kind)
	return vectorRow{}
}

func assertWrite(t *testing.T, expectedHex string, write func(*Writer)) {
	t.Helper()
	w := NewWriter()
	write(w)
	if got := strings.ToUpper(hex.EncodeToString(w.Bytes())); got != expectedHex {
		t.Fatalf("write mismatch:\nwant %s\n got %s", expectedHex, got)
	}
}

func assertRead[T any](t *testing.T, sourceHex string, expected T, read func(*Reader) T) {
	t.Helper()
	r := NewReader(decodeHex(t, sourceHex))
	got := read(r)
	r.AssertEOF()
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("read mismatch: want %v got %v", expected, got)
	}
}

func assertReadBytes(t *testing.T, sourceHex string, expected []byte, read func(*Reader) []byte) {
	t.Helper()
	r := NewReader(decodeHex(t, sourceHex))
	got := read(r)
	r.AssertEOF()
	if !bytes.Equal(got, expected) {
		t.Fatalf("read mismatch: want %x got %x", expected, got)
	}
}

func assertReadByteArrays(t *testing.T, sourceHex string, expected [][]byte, read func(*Reader) [][]byte) {
	t.Helper()
	r := NewReader(decodeHex(t, sourceHex))
	got := read(r)
	r.AssertEOF()
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("read mismatch: want %x got %x", expected, got)
	}
}

func assertReadBigInt(t *testing.T, sourceHex string, expected *big.Int, read func(*Reader) *big.Int) {
	t.Helper()
	r := NewReader(decodeHex(t, sourceHex))
	got := read(r)
	r.AssertEOF()
	if got.Cmp(expected) != 0 {
		t.Fatalf("read mismatch: want %s got %s", expected, got)
	}
}

func assertReadBigIntSlice(t *testing.T, sourceHex string, expected []*big.Int, read func(*Reader) []*big.Int) {
	t.Helper()
	r := NewReader(decodeHex(t, sourceHex))
	got := read(r)
	r.AssertEOF()
	if len(got) != len(expected) {
		t.Fatalf("read length mismatch: want %d got %d", len(expected), len(got))
	}
	for i := range got {
		if got[i].Cmp(expected[i]) != 0 {
			t.Fatalf("read mismatch at %d: want %s got %s", i, expected[i], got[i])
		}
	}
}

func assertReadIntX(t *testing.T, sourceHex string, expected IntX, read func(*Reader) IntX) {
	t.Helper()
	r := NewReader(decodeHex(t, sourceHex))
	got := read(r)
	r.AssertEOF()
	if got.BigInt().Cmp(expected.BigInt()) != 0 {
		t.Fatalf("read mismatch: want %s got %s", expected.String(), got.String())
	}
}

func assertCarbonRoundTrip(t *testing.T, sourceHex string, blob Blob) {
	t.Helper()
	readCarbon(t, sourceHex, blob)
	assertSerializedHex(t, sourceHex, blob)
}

func assertSerializedHex(t *testing.T, expectedHex string, blob Blob) {
	t.Helper()
	if got := strings.ToUpper(hex.EncodeToString(Serialize(blob))); got != expectedHex {
		t.Fatalf("serialization mismatch:\nwant %s\n got %s", expectedHex, got)
	}
}

func readCarbon(t *testing.T, sourceHex string, blob Blob) {
	t.Helper()
	r := NewReader(decodeHex(t, sourceHex))
	blob.ReadCarbon(r)
	r.AssertEOF()
}

func parseInt(t *testing.T, value string, bits int) int64 {
	t.Helper()
	out, err := strconv.ParseInt(value, 10, bits)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func parseUint(t *testing.T, value string, bits int) uint64 {
	t.Helper()
	out, err := strconv.ParseUint(value, 10, bits)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func parseBigInt(t *testing.T, value string) *big.Int {
	t.Helper()
	out, ok := new(big.Int).SetString(value, 10)
	if !ok {
		t.Fatalf("invalid BigInt %q", value)
	}
	return out
}

func parseInt8CSV(t *testing.T, value string) []int8 {
	t.Helper()
	parts := splitCSV(value)
	out := make([]int8, len(parts))
	for i, part := range parts {
		out[i] = int8(parseInt(t, part, 8))
	}
	return out
}

func parseInt16CSV(t *testing.T, value string) []int16 {
	t.Helper()
	parts := splitCSV(value)
	out := make([]int16, len(parts))
	for i, part := range parts {
		out[i] = int16(parseInt(t, part, 16))
	}
	return out
}

func parseInt32CSV(t *testing.T, value string) []int32 {
	t.Helper()
	parts := splitCSV(value)
	out := make([]int32, len(parts))
	for i, part := range parts {
		out[i] = int32(parseInt(t, part, 32))
	}
	return out
}

func parseInt64CSV(t *testing.T, value string) []int64 {
	t.Helper()
	parts := splitCSV(value)
	out := make([]int64, len(parts))
	for i, part := range parts {
		out[i] = parseInt(t, part, 64)
	}
	return out
}

func parseUint64CSV(t *testing.T, value string) []uint64 {
	t.Helper()
	parts := splitCSV(value)
	out := make([]uint64, len(parts))
	for i, part := range parts {
		out[i] = parseUint(t, part, 64)
	}
	return out
}

func parseBigIntCSV(t *testing.T, value string) []*big.Int {
	t.Helper()
	parts := splitCSV(value)
	out := make([]*big.Int, len(parts))
	for i, part := range parts {
		out[i] = parseBigInt(t, part)
	}
	return out
}

func parseByteArrays(t *testing.T, value string) [][]byte {
	t.Helper()
	value = strings.TrimPrefix(strings.TrimSuffix(value, "]]"), "[[")
	if value == "" {
		return nil
	}
	parts := strings.Split(value, "],[")
	out := make([][]byte, len(parts))
	for i, part := range parts {
		out[i] = decodeHex(t, strings.ReplaceAll(part, ",", ""))
	}
	return out
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}

func decodeHex(t *testing.T, value string) []byte {
	t.Helper()
	out, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func testSenderPublicKey(t *testing.T) Bytes32 {
	t.Helper()
	keys, err := cryptography.FromWIF("KwPpBSByydVKqStGHAnZzQofCqhDmD2bfRgc9BmZqM3ZmsdWJw4d")
	if err != nil {
		t.Fatal(err)
	}
	return MustBytes32(keys.PublicKey())
}

func maxUint256() *big.Int {
	return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
}

const samplePNGIconDataURI = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR4nGMAAQAABQABDQottAAAAABJRU5ErkJggg=="
