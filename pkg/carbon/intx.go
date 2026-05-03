package carbon

import "math/big"

var (
	minI64 = new(big.Int).Lsh(big.NewInt(-1), 63)
	maxI64 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 63), big.NewInt(1))
)

// IntX represents Carbon's variable-width signed integer encoding.
type IntX struct {
	value big.Int
}

// NewIntX copies value into a Carbon IntX.
func NewIntX(value *big.Int) IntX {
	var out IntX
	if value != nil {
		out.value.Set(value)
	}
	return out
}

// IntXFromInt64 returns an IntX from a signed 64-bit integer.
func IntXFromInt64(value int64) IntX {
	return NewIntX(big.NewInt(value))
}

// MustIntXFromString parses a base-10 IntX and panics on invalid input.
func MustIntXFromString(value string) IntX {
	v, ok := new(big.Int).SetString(value, 10)
	if !ok {
		panic("invalid IntX decimal string")
	}
	return NewIntX(v)
}

// BigInt returns a copy of the IntX value.
func (x IntX) BigInt() *big.Int {
	return new(big.Int).Set(&x.value)
}

func (x IntX) String() string {
	return x.value.String()
}

// Is8ByteSafe reports whether the value can use Carbon's compact int64 form.
func (x IntX) Is8ByteSafe() bool {
	return x.value.Cmp(minI64) >= 0 && x.value.Cmp(maxI64) <= 0
}

// WriteCarbon writes the IntX value to w.
func (x IntX) WriteCarbon(w *Writer) {
	if x.Is8ByteSafe() {
		header := byte(0x08)
		if x.value.Sign() < 0 {
			header = 0x88
		}
		w.Write1(header)
		w.Write8(x.value.Int64())
		return
	}
	w.WriteBigInt(&x.value)
}

// ReadCarbon reads the IntX value from r.
func (x *IntX) ReadCarbon(r *Reader) {
	header := r.Read1()
	length := header & 0x3f
	if length < 8 {
		panic("invalid IntX packing")
	}

	if length == 8 {
		raw := r.ReadRaw(8)
		value := int64FromLittleEndian(raw)
		headerNegative := header&0x80 != 0
		valueNegative := value < 0
		if headerNegative == valueNegative {
			x.value.SetInt64(value)
			return
		}

		fill := byte(0x00)
		if headerNegative {
			fill = 0xff
		}
		var word [32]byte
		copy(word[:], raw)
		for i := len(raw); i < len(word); i++ {
			word[i] = fill
		}
		x.value.Set(intFromWord(word))
		return
	}

	x.value.Set(r.ReadBigIntWithHeader(int(header)))
}

func int64FromLittleEndian(raw []byte) int64 {
	var v uint64
	for i := 0; i < len(raw); i++ {
		v |= uint64(raw[i]) << (8 * i)
	}
	return int64(v)
}
