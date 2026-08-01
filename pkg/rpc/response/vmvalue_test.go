package response

import (
	"encoding/json"
	"reflect"
	"testing"
)

// The shapes below come from getToken("SOUL", true) on devnet (2026-08-01): 15 scalar rows plus
// the "_ia" interest row, which is an array of structs. That row is what a string-typed model
// cannot represent, so it is the fixture that pins the whole VM value contract.
const soulMetadataWire = `[
	{"key":"name","value":"Phantasma Stake"},
	{"key":"_ia","value":[{"div":"10000","mul":"25","who":["P2K6h"]},{"fix":"1"}]}
]`

func TestTokenPropertyRowsDecodeScalarAndStructuredValues(t *testing.T) {
	var rows []TokenPropertyResult
	if err := json.Unmarshal([]byte(soulMetadataWire), &rows); err != nil {
		t.Fatalf("metadata decode failed: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 metadata rows, got %d", len(rows))
	}

	name, ok := rows[0].Value.AsText()
	if !ok || name != "Phantasma Stake" {
		t.Fatalf("scalar row must stay a scalar: %+v", rows[0])
	}

	interest, ok := rows[1].Value.AsItems()
	if !ok || len(interest) != 2 {
		t.Fatalf("_ia must decode as an array of 2 items: %+v", rows[1])
	}
	mul, ok := interest[0].Field("mul")
	if !ok {
		t.Fatalf("first interest entry must carry mul: %+v", interest[0])
	}
	if text, ok := mul.AsText(); !ok || text != "25" {
		t.Fatalf("mul must be the scalar 25: %+v", mul)
	}
	who, ok := interest[0].Field("who")
	if !ok {
		t.Fatalf("first interest entry must carry who: %+v", interest[0])
	}
	if items, ok := who.AsItems(); !ok || len(items) != 1 || items[0].Text != "P2K6h" {
		t.Fatalf("nested who array must survive: %+v", who)
	}
}

func TestVMValueRoundTripsToTheWireShape(t *testing.T) {
	// Re-serializing must reproduce the plain JSON value: nothing gets wrapped in an envelope and
	// nothing gets packed into a string, which is what makes the answers readable to consumers
	// that do not use this SDK.
	var rows []TokenPropertyResult
	if err := json.Unmarshal([]byte(soulMetadataWire), &rows); err != nil {
		t.Fatalf("metadata decode failed: %v", err)
	}

	encoded, err := json.Marshal(rows)
	if err != nil {
		t.Fatalf("metadata encode failed: %v", err)
	}
	assertSameJSON(t, []byte(soulMetadataWire), encoded)
}

func TestVMValueNormalizesUntypedScalars(t *testing.T) {
	// The node writes every scalar as a string, but an older node - or a hand-written response -
	// can still answer a bare number, boolean or null. Those must normalize instead of failing the
	// whole answer, and a number must keep its exact digits: chain values are big integers and a
	// float64 round trip would corrupt anything above 2^53.
	cases := []struct {
		wire string
		want string
	}{
		{wire: `"25"`, want: "25"},
		{wire: `25`, want: "25"},
		{wire: `123456789012345678901234567890`, want: "123456789012345678901234567890"},
		{wire: `true`, want: "true"},
		{wire: `false`, want: "false"},
		{wire: `null`, want: ""},
	}

	for _, testCase := range cases {
		var value VMValue
		if err := json.Unmarshal([]byte(testCase.wire), &value); err != nil {
			t.Fatalf("decoding %s failed: %v", testCase.wire, err)
		}
		text, ok := value.AsText()
		if !ok || text != testCase.want {
			t.Fatalf("%s must normalize to the scalar %q, got %+v", testCase.wire, testCase.want, value)
		}
	}
}

func TestVMValueAccessorsRejectTheWrongShape(t *testing.T) {
	scalar := VMText("25")
	if _, ok := scalar.AsItems(); ok {
		t.Fatal("a scalar must not answer as an array")
	}
	if _, ok := scalar.AsFields(); ok {
		t.Fatal("a scalar must not answer as a struct")
	}
	if _, ok := scalar.Field("mul"); ok {
		t.Fatal("a scalar has no fields")
	}

	structured := VMFields(map[string]VMValue{"mul": VMText("25")})
	if _, ok := structured.AsText(); ok {
		t.Fatal("a struct must not answer as a scalar")
	}
	if structured.IsText() {
		t.Fatal("a struct is not a scalar")
	}
	if _, ok := structured.Field("missing"); ok {
		t.Fatal("a missing field must report absent")
	}
}

func TestEmptyVMValueCollectionsEncodeAsEmptyNotNull(t *testing.T) {
	// The shape itself is the type discriminator on this wire, so an empty array must not
	// degenerate into null, which a reader would take for a scalar.
	encoded, err := json.Marshal(struct {
		Items  VMValue `json:"items"`
		Fields VMValue `json:"fields"`
	}{
		Items:  VMValue{Kind: VMValueItems},
		Fields: VMValue{Kind: VMValueFields},
	})
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	assertSameJSON(t, []byte(`{"items":[],"fields":{}}`), encoded)
}

// assertSameJSON compares two JSON documents by structure, so that key order and whitespace do not
// decide the outcome.
func assertSameJSON(t *testing.T, expected, actual []byte) {
	t.Helper()

	var want, got any
	if err := json.Unmarshal(expected, &want); err != nil {
		t.Fatalf("expected document is not valid JSON: %v", err)
	}
	if err := json.Unmarshal(actual, &got); err != nil {
		t.Fatalf("actual document is not valid JSON: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("JSON mismatch:\nwant %s\ngot  %s", expected, actual)
	}
}
