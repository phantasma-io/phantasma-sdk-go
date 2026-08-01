package response

import (
	"bytes"
	"encoding/json"
)

// VMValueKind tells which of the three VM shapes a VMValue carries.
type VMValueKind uint8

const (
	// VMValueText is a scalar; the whole value is in Text.
	VMValueText VMValueKind = iota
	// VMValueItems is an array; the elements are in Items.
	VMValueItems
	// VMValueFields is a struct; the fields are in Fields.
	VMValueFields
)

// VMValue is a value decoded from VM storage: a scalar, an array, or a struct.
//
// VM values are dynamically typed, so the wire carries the plain JSON value - a string, an array
// or an object - and the shape itself says which of the three it is. Nothing is packed into a JSON
// string, which is what these values used to be before the 2026-08 node series.
//
// Scalars are always text: chain numbers are big integers and JSON numbers lose precision above
// 2^53, so the node writes them as decimal strings, and byte values arrive as hex. Struct field
// names arrive exactly as the chain stores them; the node does not rename dictionary keys.
//
// The zero value is an empty scalar, which is also what an explicit JSON null decodes to.
type VMValue struct {
	Kind   VMValueKind
	Text   string
	Items  []VMValue
	Fields map[string]VMValue
}

// VMText builds a scalar value.
func VMText(text string) VMValue {
	return VMValue{Kind: VMValueText, Text: text}
}

// VMItems builds an array value.
func VMItems(items ...VMValue) VMValue {
	return VMValue{Kind: VMValueItems, Items: items}
}

// VMFields builds a struct value.
func VMFields(fields map[string]VMValue) VMValue {
	return VMValue{Kind: VMValueFields, Fields: fields}
}

// IsText reports whether the value is a scalar.
func (v VMValue) IsText() bool {
	return v.Kind == VMValueText
}

// AsText returns the scalar content; ok is false for an array or a struct.
func (v VMValue) AsText() (text string, ok bool) {
	if v.Kind != VMValueText {
		return "", false
	}
	return v.Text, true
}

// AsItems returns the array elements; ok is false for a scalar or a struct.
func (v VMValue) AsItems() (items []VMValue, ok bool) {
	if v.Kind != VMValueItems {
		return nil, false
	}
	return v.Items, true
}

// AsFields returns the struct fields; ok is false for a scalar or an array.
func (v VMValue) AsFields() (fields map[string]VMValue, ok bool) {
	if v.Kind != VMValueFields {
		return nil, false
	}
	return v.Fields, true
}

// Field returns one field of a struct; ok is false for a scalar, an array, or a missing field.
func (v VMValue) Field(name string) (field VMValue, ok bool) {
	if v.Kind != VMValueFields {
		return VMValue{}, false
	}
	field, ok = v.Fields[name]
	return field, ok
}

// MarshalJSON writes the value as the plain JSON it represents: a string, an array or an object.
// A nil slice or map still writes as an empty array or object, never as null, because the shape
// itself is what tells a reader which of the three kinds this is.
func (v VMValue) MarshalJSON() ([]byte, error) {
	switch v.Kind {
	case VMValueItems:
		items := v.Items
		if items == nil {
			items = []VMValue{}
		}
		return json.Marshal(items)
	case VMValueFields:
		fields := v.Fields
		if fields == nil {
			fields = map[string]VMValue{}
		}
		return json.Marshal(fields)
	default:
		return json.Marshal(v.Text)
	}
}

// UnmarshalJSON maps the plain JSON value onto the three VM shapes.
func (v *VMValue) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	// Numbers have to survive as their exact text: chain values are big integers, and the float64
	// that encoding/json produces by default rounds anything above 2^53.
	decoder.UseNumber()

	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}

	*v = vmValueFromDecoded(decoded)
	return nil
}

// vmValueFromDecoded converts one already-parsed JSON value.
//
// Numbers and booleans are normalized to their JSON text: the node writes every scalar as a
// string, but a response from an older node - or a hand-written one - can still carry them
// untyped, and failing the whole answer over that would be worse than normalizing it. An explicit
// null becomes an empty scalar; the node omits empty values instead of answering null.
//
// Nesting depth is bounded by encoding/json itself, which rejects input nested deeper than its own
// recursion limit before anything reaches this function.
func vmValueFromDecoded(decoded any) VMValue {
	switch value := decoded.(type) {
	case string:
		return VMText(value)
	case []any:
		items := make([]VMValue, 0, len(value))
		for _, item := range value {
			items = append(items, vmValueFromDecoded(item))
		}
		return VMValue{Kind: VMValueItems, Items: items}
	case map[string]any:
		fields := make(map[string]VMValue, len(value))
		for name, field := range value {
			fields[name] = vmValueFromDecoded(field)
		}
		return VMValue{Kind: VMValueFields, Fields: fields}
	case json.Number:
		return VMText(value.String())
	case bool:
		if value {
			return VMText("true")
		}
		return VMText("false")
	default:
		return VMValue{}
	}
}
