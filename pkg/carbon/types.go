package carbon

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Bytes16 is a fixed-width 16-byte Carbon value.
type Bytes16 [16]byte

// Bytes32 is a fixed-width 32-byte Carbon value.
type Bytes32 [32]byte

// Bytes64 is a fixed-width 64-byte Carbon value.
type Bytes64 [64]byte

var (
	// EmptyBytes16 is the zero 16-byte Carbon value.
	EmptyBytes16 Bytes16
	// EmptyBytes32 is the zero 32-byte Carbon value.
	EmptyBytes32 Bytes32
	// EmptyBytes64 is the zero 64-byte Carbon value.
	EmptyBytes64 Bytes64
)

// NewBytes16 returns a Bytes16 from exactly 16 bytes.
func NewBytes16(data []byte) (Bytes16, error) {
	var out Bytes16
	if len(data) != len(out) {
		return out, fmt.Errorf("Bytes16 length must be %d, got %d", len(out), len(data))
	}
	copy(out[:], data)
	return out, nil
}

// MustBytes16 returns a Bytes16 and panics if data is not exactly 16 bytes.
func MustBytes16(data []byte) Bytes16 {
	out, err := NewBytes16(data)
	if err != nil {
		panic(err)
	}
	return out
}

// NewBytes32 returns a Bytes32 from exactly 32 bytes.
func NewBytes32(data []byte) (Bytes32, error) {
	var out Bytes32
	if len(data) != len(out) {
		return out, fmt.Errorf("Bytes32 length must be %d, got %d", len(out), len(data))
	}
	copy(out[:], data)
	return out, nil
}

// MustBytes32 returns a Bytes32 and panics if data is not exactly 32 bytes.
func MustBytes32(data []byte) Bytes32 {
	out, err := NewBytes32(data)
	if err != nil {
		panic(err)
	}
	return out
}

// NewBytes64 returns a Bytes64 from exactly 64 bytes.
func NewBytes64(data []byte) (Bytes64, error) {
	var out Bytes64
	if len(data) != len(out) {
		return out, fmt.Errorf("Bytes64 length must be %d, got %d", len(out), len(data))
	}
	copy(out[:], data)
	return out, nil
}

// MustBytes64 returns a Bytes64 and panics if data is not exactly 64 bytes.
func MustBytes64(data []byte) Bytes64 {
	out, err := NewBytes64(data)
	if err != nil {
		panic(err)
	}
	return out
}

// Bytes16FromHex decodes a 16-byte Carbon value from hex.
func Bytes16FromHex(value string) (Bytes16, error) {
	data, err := decodeFixedHex(value, 16)
	if err != nil {
		return Bytes16{}, err
	}
	return NewBytes16(data)
}

// MustBytes16FromHex decodes a 16-byte Carbon value from hex and panics on error.
func MustBytes16FromHex(value string) Bytes16 {
	out, err := Bytes16FromHex(value)
	if err != nil {
		panic(err)
	}
	return out
}

// Bytes32FromHex decodes a 32-byte Carbon value from hex.
func Bytes32FromHex(value string) (Bytes32, error) {
	data, err := decodeFixedHex(value, 32)
	if err != nil {
		return Bytes32{}, err
	}
	return NewBytes32(data)
}

// MustBytes32FromHex decodes a 32-byte Carbon value from hex and panics on error.
func MustBytes32FromHex(value string) Bytes32 {
	out, err := Bytes32FromHex(value)
	if err != nil {
		panic(err)
	}
	return out
}

// Bytes64FromHex decodes a 64-byte Carbon value from hex.
func Bytes64FromHex(value string) (Bytes64, error) {
	data, err := decodeFixedHex(value, 64)
	if err != nil {
		return Bytes64{}, err
	}
	return NewBytes64(data)
}

// MustBytes64FromHex decodes a 64-byte Carbon value from hex and panics on error.
func MustBytes64FromHex(value string) Bytes64 {
	out, err := Bytes64FromHex(value)
	if err != nil {
		panic(err)
	}
	return out
}

// DecodeHex decodes hex and accepts an optional 0x prefix.
func DecodeHex(value string) ([]byte, error) {
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	return hex.DecodeString(value)
}

// MustDecodeHex decodes hex and panics on error.
func MustDecodeHex(value string) []byte {
	out, err := DecodeHex(value)
	if err != nil {
		panic(err)
	}
	return out
}

// Bytes returns a copy of the fixed-width value.
func (b Bytes16) Bytes() []byte {
	out := make([]byte, len(b))
	copy(out, b[:])
	return out
}

// Bytes returns a copy of the fixed-width value.
func (b Bytes32) Bytes() []byte {
	out := make([]byte, len(b))
	copy(out, b[:])
	return out
}

// Bytes returns a copy of the fixed-width value.
func (b Bytes64) Bytes() []byte {
	out := make([]byte, len(b))
	copy(out, b[:])
	return out
}

func (b Bytes16) String() string {
	return hex.EncodeToString(b[:])
}

func (b Bytes32) String() string {
	return hex.EncodeToString(b[:])
}

func (b Bytes64) String() string {
	return hex.EncodeToString(b[:])
}

// WriteCarbon writes the fixed-width value to w.
func (b Bytes16) WriteCarbon(w *Writer) {
	w.Write16(b)
}

// ReadCarbon reads the fixed-width value from r.
func (b *Bytes16) ReadCarbon(r *Reader) {
	*b = r.Read16()
}

// WriteCarbon writes the fixed-width value to w.
func (b Bytes32) WriteCarbon(w *Writer) {
	w.Write32(b)
}

// ReadCarbon reads the fixed-width value from r.
func (b *Bytes32) ReadCarbon(r *Reader) {
	*b = r.Read32()
}

// WriteCarbon writes the fixed-width value to w.
func (b Bytes64) WriteCarbon(w *Writer) {
	w.Write64(b)
}

// ReadCarbon reads the fixed-width value from r.
func (b *Bytes64) ReadCarbon(r *Reader) {
	*b = r.Read64()
}

// SmallString is a Carbon length-prefixed UTF-8 string with a one-byte length.
type SmallString string

// NewSmallString validates and returns a Carbon small string.
func NewSmallString(value string) (SmallString, error) {
	if err := validateSmallString(value); err != nil {
		return "", err
	}
	return SmallString(value), nil
}

// MustSmallString returns a Carbon small string and panics if value is invalid.
func MustSmallString(value string) SmallString {
	out, err := NewSmallString(value)
	if err != nil {
		panic(err)
	}
	return out
}

func (s SmallString) String() string {
	return string(s)
}

// Bytes returns the UTF-8 bytes of the small string.
func (s SmallString) Bytes() []byte {
	return []byte(s)
}

// WriteCarbon writes the small string to w.
func (s SmallString) WriteCarbon(w *Writer) {
	if err := validateSmallString(string(s)); err != nil {
		panic(err)
	}
	data := []byte(s)
	w.Write1(byte(len(data)))
	w.WriteRaw(data)
}

// ReadCarbon reads the small string from r.
func (s *SmallString) ReadCarbon(r *Reader) {
	length := int(r.Read1())
	data := r.ReadRaw(length)
	if !utf8.Valid(data) {
		panic("SmallString is not valid UTF-8")
	}
	*s = SmallString(string(data))
}

func decodeFixedHex(value string, expected int) ([]byte, error) {
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	data, err := hex.DecodeString(value)
	if err != nil {
		return nil, err
	}
	if len(data) != expected {
		return nil, fmt.Errorf("decoded hex length must be %d, got %d", expected, len(data))
	}
	return data, nil
}

func validateSmallString(value string) error {
	data := []byte(value)
	if len(data) > 255 {
		return fmt.Errorf("SmallString too long")
	}
	if !utf8.Valid(data) {
		return fmt.Errorf("SmallString is not valid UTF-8")
	}
	return nil
}
