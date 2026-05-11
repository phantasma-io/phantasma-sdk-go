package vm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/domain/types"
	sdkio "github.com/phantasma-io/phantasma-sdk-go/pkg/io"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/util"
)

var unitCoveredGen2Fixtures = []string{
	"gen2_csharp_vm_bigint_binary.tsv",
	"gen2_csharp_vm_bigint_decimal.tsv",
	"gen2_csharp_vmobject_arraytype.tsv",
	"gen2_csharp_vmobject_asbool.tsv",
	"gen2_csharp_vmobject_asbytes.tsv",
	"gen2_csharp_vmobject_asnumber.tsv",
	"gen2_csharp_vmobject_asstring.tsv",
	"gen2_csharp_vmobject_cast_struct.tsv",
	"gen2_csharp_vmobject_serde.tsv",
}

var liveRunnerCoveredGen2Fixtures = []string{
	"gen2_csharp_vm_scriptcontext_ops.tsv",
	"gen2_csharp_vm_scriptcontext_unary.tsv",
}

var notSDKUnitApplicableGen2Fixtures = []string{
	"gen2_csharp_vm_bigint_narrow_int.tsv",
	"gen2_csharp_vm_bigint_ops.tsv",
	"gen2_csharp_vm_bigint_unary_ops.tsv",
}

var gen2FixtureSHA256 = map[string]string{
	"gen2_csharp_vm_bigint_binary.tsv":       "a5be05751b35de8b7b3578577bb2769073ac7a2ddea3eaf9503d76d0302fa464",
	"gen2_csharp_vm_bigint_decimal.tsv":      "1bede4198883018817d94eceefe4e7b70a9f5c96c9d60d57481990ad21b027a9",
	"gen2_csharp_vm_bigint_narrow_int.tsv":   "b82315b4483c23ee7e3e9943b5c41cf8daf12c627c6e12f30b735ad7dbde1445",
	"gen2_csharp_vm_bigint_ops.tsv":          "997f3a935393358a89c7be785176e8528535111994bc5193c7d7ddc2429aa3d3",
	"gen2_csharp_vm_bigint_unary_ops.tsv":    "53719de8a1528897a083401aaad251cdb3e9e201f8639d29cd3708beeda93ea7",
	"gen2_csharp_vm_scriptcontext_ops.tsv":   "c87e4a5ec075b8efc0abe88a551ae8fe505df04167cb0e4f2714768c0a1e917f",
	"gen2_csharp_vm_scriptcontext_unary.tsv": "7198d33a84bd61c671dc1871f2b56e232748c41d69e957e8f994cd2dc9b5922c",
	"gen2_csharp_vmobject_arraytype.tsv":     "f6b7ce9cd92f464d260018ffb1a0ab01202ca908cf915dead3295e8270ddf532",
	"gen2_csharp_vmobject_asbool.tsv":        "a2979cc7eccd22760de82f8401de4b8b41c45fedf09b91a94871d3a3051c85d5",
	"gen2_csharp_vmobject_asbytes.tsv":       "dd326e18c94e2e116705893f742c708cfb1cd7b96c8a40a2ab6637b39ae409b9",
	"gen2_csharp_vmobject_asnumber.tsv":      "986cfc21658c66b04c1ffaaa7bb9fa08bc9a3acd929276d0d2496ba43c43bf69",
	"gen2_csharp_vmobject_asstring.tsv":      "eb14408b7e65fc417bf1bbfe4fb1e87c3d06d28734c7c25514a806f41fceede6",
	"gen2_csharp_vmobject_cast_struct.tsv":   "1580a9ec312619a7e2632076073ae80d57dcfc3defc0ef7b4876da34c0e231af",
	"gen2_csharp_vmobject_serde.tsv":         "0c74c90e83c5c20bed48b1d52ca5489d15a7c4f67874184c1d0a4f708ce5e42f",
}

func TestGen2FixtureManifestIsExplicit(t *testing.T) {
	// Keep every copied Gen2 fixture either covered by Go unit tests, covered by
	// a live runner, or explicitly classified as VM-runner-only behavior.
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	discovered := map[string]bool{}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "gen2_csharp_") && strings.HasSuffix(name, ".tsv") {
			discovered[name] = true
		}
	}
	classified := map[string]bool{}
	for _, group := range [][]string{unitCoveredGen2Fixtures, liveRunnerCoveredGen2Fixtures, notSDKUnitApplicableGen2Fixtures} {
		for _, name := range group {
			classified[name] = true
		}
	}
	if !reflect.DeepEqual(discovered, classified) {
		t.Fatalf("fixture classification mismatch:\ndiscovered=%v\nclassified=%v", discovered, classified)
	}
	for name, expected := range gen2FixtureSHA256 {
		data, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != expected {
			t.Fatalf("%s hash mismatch: want %s got %s", name, expected, got)
		}
	}
}

func TestVMObjectStringAsNumberMatchesGen2DecimalFixtures(t *testing.T) {
	// Decimal strings are parsed with Gen2 BigInteger rules: optional sign and
	// arbitrary precision are accepted, while malformed input must reject.
	for _, parts := range fixtureRows(t, "gen2_csharp_vm_bigint_decimal.tsv") {
		caseID, input, outcome, expected := parts[0], parts[1], parts[2], parts[3]
		result, panicked := callBigInt(func() *big.Int {
			return (&VMObject{Type: String, Data: input}).AsNumber()
		})
		if outcome == "ok" {
			requireNoPanic(t, caseID, panicked)
			requireBigIntString(t, caseID, result, expected)
		} else if !panicked {
			t.Fatalf("%s expected panic", caseID)
		}
	}
}

func TestVMObjectAsNumberMatchesGen2Fixtures(t *testing.T) {
	// VMObject.AsNumber must match the classic VM for all scalar/object source
	// types, including signed little-endian byte arrays and 32-byte hash objects.
	for _, parts := range fixtureRows(t, "gen2_csharp_vmobject_asnumber.tsv") {
		caseID, outcome, expected := parts[0], parts[4], parts[5]
		object := objectFromFixture(t, parts[1], parts[3])
		result, panicked := callBigInt(object.AsNumber)
		if outcome == "ok" {
			requireNoPanic(t, caseID, panicked)
			requireBigIntString(t, caseID, result, expected)
		} else if !panicked {
			t.Fatalf("%s expected panic", caseID)
		}
	}
}

func TestVMObjectAsBytesMatchesGen2Fixtures(t *testing.T) {
	// VMObject.AsBytes is the byte-level bridge used by script and contract
	// helpers, so the exact Gen2 output bytes are asserted row by row.
	for _, parts := range fixtureRows(t, "gen2_csharp_vmobject_asbytes.tsv") {
		caseID, outcome, expected := parts[0], parts[4], parts[5]
		object := objectFromFixture(t, parts[1], parts[3])
		result, panicked := callBytes(object.AsBytes)
		if outcome == "ok" {
			requireNoPanic(t, caseID, panicked)
			if got := hex.EncodeToString(result); got != expected {
				t.Fatalf("%s bytes mismatch: want %s got %s", caseID, expected, got)
			}
		} else if !panicked {
			t.Fatalf("%s expected panic", caseID)
		}
	}
}

func TestVMObjectAsStringMatchesGen2Fixtures(t *testing.T) {
	// Struct-to-string must decode number arrays as UTF-16 code units and fall
	// back to base64 for non-number structs, matching Gen2 observable behavior.
	for _, parts := range fixtureRows(t, "gen2_csharp_vmobject_asstring.tsv") {
		caseID, expected := parts[0], parts[5]
		if got := objectFromFixture(t, parts[1], parts[3]).AsString(); got != expected {
			t.Fatalf("%s string mismatch: want %q got %q", caseID, expected, got)
		}
	}
}

func TestVMObjectAsBoolMatchesGen2Fixtures(t *testing.T) {
	// Gen2 accepts booleans, numeric truthiness, and single-byte byte arrays
	// only; other source types must reject instead of guessing.
	for _, parts := range fixtureRows(t, "gen2_csharp_vmobject_asbool.tsv") {
		caseID, outcome, expected := parts[0], parts[4], parts[5]
		object := objectFromFixture(t, parts[1], parts[3])
		result, panicked := callBool(object.AsBool)
		if outcome == "ok" {
			requireNoPanic(t, caseID, panicked)
			if got := strconv.FormatBool(result); got != expected {
				t.Fatalf("%s bool mismatch: want %s got %s", caseID, expected, got)
			}
		} else if !panicked {
			t.Fatalf("%s expected panic", caseID)
		}
	}
}

func TestVMObjectArrayTypeMatchesGen2Fixtures(t *testing.T) {
	// Array type detection is value-based on numeric 0..N keys, not Go object
	// identity or map iteration order.
	for _, parts := range fixtureRows(t, "gen2_csharp_vmobject_arraytype.tsv") {
		caseID, expected := parts[0], parts[4]
		if got := VMTypeLookup[objectFromFixture(t, parts[1], parts[3]).GetArrayType()]; got != expected {
			t.Fatalf("%s array type mismatch: want %s got %s", caseID, expected, got)
		}
	}
}

func TestVMObjectSerdeMatchesGen2Fixtures(t *testing.T) {
	// Serialization fixtures pin the exact VMObject wire encoding and the
	// lossy Object-vs-Bytes roundtrip behavior used by the Gen2 VM serializer.
	for _, parts := range fixtureRows(t, "gen2_csharp_vmobject_serde.tsv") {
		caseID, serializedHex, roundtripType, descriptor := parts[0], parts[4], parts[5], parts[6]
		object := objectFromFixture(t, parts[1], parts[3])
		if got := hex.EncodeToString(serializeVMObject(object)); got != serializedHex {
			t.Fatalf("%s serialize mismatch: want %s got %s", caseID, serializedHex, got)
		}
		roundtrip := objectFromBytes(t, serializedHex)
		if got := VMTypeLookup[roundtrip.Type]; got != roundtripType {
			t.Fatalf("%s roundtrip type mismatch: want %s got %s", caseID, roundtripType, got)
		}
		if got := objectDescriptor(roundtrip); got != descriptor {
			t.Fatalf("%s descriptor mismatch: want %s got %s", caseID, descriptor, got)
		}
	}
}

func TestVMObjectCastToStructMatchesGen2Fixtures(t *testing.T) {
	// String-to-Struct casts use UTF-16 code units in numbered struct slots;
	// existing Object values are preserved and invalid scalar casts reject.
	for _, parts := range fixtureRows(t, "gen2_csharp_vmobject_cast_struct.tsv") {
		caseID, outcome, resultType, descriptor := parts[0], parts[4], parts[5], parts[6]
		object := objectFromFixture(t, parts[1], parts[3])
		result, panicked := callObject(func() *VMObject { return object.CastTo(Struct) })
		if outcome == "ok" {
			requireNoPanic(t, caseID, panicked)
			if got := VMTypeLookup[result.Type]; got != resultType {
				t.Fatalf("%s result type mismatch: want %s got %s", caseID, resultType, got)
			}
			if got := objectDescriptor(result); got != descriptor {
				t.Fatalf("%s descriptor mismatch: want %s got %s", caseID, descriptor, got)
			}
		} else if !panicked {
			t.Fatalf("%s expected panic", caseID)
		}
	}
}

func fixtureRows(t *testing.T, name string) [][]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	var rows [][]string
	width := 0
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if strings.HasPrefix(line, "case_id\t") {
			width = len(parts)
			continue
		}
		for width != 0 && len(parts) < width {
			parts = append(parts, "")
		}
		rows = append(rows, parts)
	}
	return rows
}

func objectFromFixture(t *testing.T, sourceKind, payload string) *VMObject {
	t.Helper()
	switch sourceKind {
	case "serialized_vmobject":
		return objectFromBytes(t, payload)
	case "empty":
		return &VMObject{Type: None}
	case "string":
		return &VMObject{Type: String, Data: payload}
	case "bytes":
		return &VMObject{Type: Bytes, Data: mustHex(t, payload)}
	case "bool":
		return &VMObject{Type: Bool, Data: payload == "true"}
	case "enum":
		value, err := strconv.ParseUint(payload, 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		return &VMObject{Type: Enum, Data: uint32(value)}
	case "timestamp":
		value, err := strconv.ParseUint(payload, 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		return &VMObject{Type: Timestamp, Data: types.Timestamp{Value: uint32(value)}}
	case "number":
		value, ok := new(big.Int).SetString(payload, 10)
		if !ok {
			t.Fatalf("bad number fixture: %s", payload)
		}
		return &VMObject{Type: Number, Data: *value}
	case "object":
		data := mustHex(t, payload)
		if len(data) == cryptography.Length {
			return &VMObject{Type: Object, Data: cryptography.NewAddress(data)}
		}
		return &VMObject{Type: Object, Data: data}
	case "struct":
		return &VMObject{Type: Struct, Data: VMObjectStruct{
			{Key: VMObject{Type: String, Data: "name"}, Value: VMObject{Type: String, Data: "neo"}},
			{Key: VMObject{Type: String, Data: "count"}, Value: VMObject{Type: Number, Data: *big.NewInt(7)}},
		}}
	default:
		t.Fatalf("unsupported fixture source kind: %s", sourceKind)
		return nil
	}
}

func objectFromBytes(t *testing.T, value string) *VMObject {
	t.Helper()
	object := &VMObject{}
	object.Deserialize(sdkio.NewBinReaderFromBuf(mustHex(t, value)))
	return object
}

func objectDescriptor(object *VMObject) string {
	switch object.Type {
	case None:
		return "None"
	case Struct:
		return "Struct:" + hex.EncodeToString(serializeVMObject(object))
	case Bytes:
		return "Bytes:" + hex.EncodeToString(object.Data.([]byte))
	case Number:
		return "Number:" + object.AsNumber().String()
	case String:
		return "String:" + object.AsString()
	case Timestamp:
		return fmt.Sprintf("Timestamp:%d", object.Data.(types.Timestamp).Value)
	case Bool:
		return "Bool:" + strconv.FormatBool(object.AsBool())
	case Enum:
		return fmt.Sprintf("Enum:%d", object.Data.(uint32))
	case Object:
		switch data := object.Data.(type) {
		case cryptography.Address:
			return "Object.Address:" + hex.EncodeToString(data.Bytes())
		case *cryptography.Address:
			return "Object.Address:" + hex.EncodeToString(data.Bytes())
		case []byte:
			if len(data) == 32 {
				return "Object.Hash:" + hex.EncodeToString(data)
			}
			return "Object:" + hex.EncodeToString(data)
		default:
			return fmt.Sprintf("Object:%T", data)
		}
	default:
		return "Unknown"
	}
}

func mustHex(t *testing.T, value string) []byte {
	t.Helper()
	if value == "" {
		return nil
	}
	out, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func callBigInt(fn func() *big.Int) (value *big.Int, panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	return fn(), false
}

func callBytes(fn func() []byte) (value []byte, panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	return fn(), false
}

func callBool(fn func() bool) (value bool, panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	return fn(), false
}

func callObject(fn func() *VMObject) (value *VMObject, panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	return fn(), false
}

func requireNoPanic(t *testing.T, caseID string, panicked bool) {
	t.Helper()
	if panicked {
		t.Fatalf("%s unexpected panic", caseID)
	}
}

func requireBigIntString(t *testing.T, caseID string, value *big.Int, expected string) {
	t.Helper()
	want, ok := new(big.Int).SetString(expected, 10)
	if !ok {
		t.Fatalf("%s bad expected integer: %s", caseID, expected)
	}
	if value.Cmp(want) != 0 {
		t.Fatalf("%s number mismatch: want %s got %s", caseID, want, value)
	}
}

func TestPhantasmaBigIntSharedVectors(t *testing.T) {
	// The shared bigint TSV is the canonical source for both Phantasma and C#
	// two's-complement byte helpers; this hash gate prevents silent drift.
	data, err := os.ReadFile(filepath.Join("..", "util", "testdata", "phantasma_bigint_vectors.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != "75e955c8a727336b1dcfb1c5f1ae47e75aae10823a493dab3f8a6b0a5a8b8402" {
		t.Fatalf("phantasma_bigint_vectors.tsv hash mismatch: %s", got)
	}
	for lineNumber, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if lineNumber == 0 || line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			t.Fatalf("bad bigint fixture row: %q", line)
		}
		value, ok := new(big.Int).SetString(parts[0], 10)
		if !ok {
			t.Fatalf("bad bigint value: %s", parts[0])
		}
		requireBytes(t, parts[0]+" phantasma", util.BigIntToPhantasmaByteArray(value), decimalBytes(t, parts[1]))
		requireBytes(t, parts[0]+" csharp", util.BigIntToCsharpByteArray(value), decimalBytes(t, parts[2]))
	}
}

func decimalBytes(t *testing.T, value string) []byte {
	t.Helper()
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var out []byte
	for _, part := range strings.Fields(value) {
		item, err := strconv.ParseUint(part, 10, 8)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, byte(item))
	}
	return out
}

func requireBytes(t *testing.T, label string, got, want []byte) {
	t.Helper()
	if hex.EncodeToString(got) != hex.EncodeToString(want) {
		t.Fatalf("%s mismatch: want %x got %x", label, want, got)
	}
}
