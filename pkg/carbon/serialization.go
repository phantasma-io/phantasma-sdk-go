package carbon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"unicode/utf8"
)

// Blob is the Carbon wire-format equivalent of the C#/TS/C++ ICarbonBlob.
// Implementations must write fields in validator wire order.
type Blob interface {
	WriteCarbon(*Writer)
	ReadCarbon(*Reader)
}

// Serialize encodes a Carbon blob using the validator binary format.
func Serialize(blob Blob) []byte {
	w := NewWriter()
	blob.WriteCarbon(w)
	return w.Bytes()
}

// Deserialize decodes a Carbon blob from the validator binary format.
func Deserialize[T Blob](data []byte, blob T) (out T, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			out = blob
			err = fmt.Errorf("carbon deserialization failed: %v", recovered)
		}
	}()
	r := NewReader(data)
	blob.ReadCarbon(r)
	r.AssertEOF()
	return blob, nil
}

// MustDeserialize decodes a Carbon blob and panics on malformed input.
func MustDeserialize[T Blob](data []byte, blob T) T {
	out, err := Deserialize(data, blob)
	if err != nil {
		panic(err)
	}
	return out
}

// Writer implements the Carbon binary writer contract used by the other SDKs.
// Integers are little-endian; strings are zero-terminated UTF-8; BigInt is the
// compact validator int256 representation.
type Writer struct {
	buf bytes.Buffer
}

// NewWriter creates an empty Carbon binary writer.
func NewWriter() *Writer {
	return &Writer{}
}

// Bytes returns the bytes written so far.
func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

// WriteRaw appends raw bytes.
func (w *Writer) WriteRaw(data []byte) {
	w.buf.Write(data)
}

// Write1 writes one byte.
func (w *Writer) Write1(v byte) {
	w.buf.WriteByte(v)
}

// Write2 writes a little-endian int16.
func (w *Writer) Write2(v int16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], uint16(v))
	w.WriteRaw(b[:])
}

// Write4 writes a little-endian int32.
func (w *Writer) Write4(v int32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], uint32(v))
	w.WriteRaw(b[:])
}

// Write4U writes a little-endian uint32.
func (w *Writer) Write4U(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	w.WriteRaw(b[:])
}

// Write8 writes a little-endian int64.
func (w *Writer) Write8(v int64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], uint64(v))
	w.WriteRaw(b[:])
}

// Write8U writes a little-endian uint64.
func (w *Writer) Write8U(v uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	w.WriteRaw(b[:])
}

// WriteFixed writes exactly count bytes or panics.
func (w *Writer) WriteFixed(data []byte, count int) {
	if len(data) != count {
		panic(fmt.Sprintf("incorrect input size: got %d, want %d", len(data), count))
	}
	w.WriteRaw(data)
}

// Write16 writes a 16-byte Carbon value.
func (w *Writer) Write16(data Bytes16) {
	w.WriteRaw(data[:])
}

// Write32 writes a 32-byte Carbon value.
func (w *Writer) Write32(data Bytes32) {
	w.WriteRaw(data[:])
}

// Write64 writes a 64-byte Carbon value.
func (w *Writer) Write64(data Bytes64) {
	w.WriteRaw(data[:])
}

// WriteBlob writes another Carbon blob.
func (w *Writer) WriteBlob(blob Blob) {
	blob.WriteCarbon(w)
}

// WriteBlobArray writes a length-prefixed Carbon blob array.
func (w *Writer) WriteBlobArray(items []Blob) {
	w.Write4(int32(len(items)))
	for _, item := range items {
		item.WriteCarbon(w)
	}
}

// WriteBigInt writes a Carbon int256 value.
func (w *Writer) WriteBigInt(value *big.Int) {
	if value == nil || value.Sign() == 0 {
		w.Write1(0)
		return
	}

	word := bigIntWord(value)
	fill := byte(0x00)
	if word[31]&0x80 != 0 {
		fill = 0xff
	}

	length := len(word)
	for length > 0 && word[length-1] == fill {
		length--
	}

	header := byte(length & 0x3f)
	if fill == 0xff {
		header |= 0x80
	}

	w.Write1(header)
	if length > 0 {
		w.WriteRaw(word[:length])
	}
}

// WriteBigIntArray writes a length-prefixed int256 array.
func (w *Writer) WriteBigIntArray(values []*big.Int) {
	w.Write4(int32(len(values)))
	for _, value := range values {
		w.WriteBigInt(value)
	}
}

// WriteStringZ writes a zero-terminated UTF-8 string.
func (w *Writer) WriteStringZ(value string) {
	if !utf8.ValidString(value) {
		panic("string is not valid UTF-8")
	}
	for i := 0; i < len(value); i++ {
		if value[i] == 0 {
			panic("string contains zero byte")
		}
	}
	w.WriteRaw([]byte(value))
	w.Write1(0)
}

// WriteStringZArray writes a length-prefixed zero-terminated string array.
func (w *Writer) WriteStringZArray(values []string) {
	w.Write4(int32(len(values)))
	for _, value := range values {
		w.WriteStringZ(value)
	}
}

// WriteByteArray writes a length-prefixed byte array.
func (w *Writer) WriteByteArray(data []byte) {
	w.Write4(int32(len(data)))
	w.WriteRaw(data)
}

// WriteByteArrays writes a length-prefixed array of byte arrays.
func (w *Writer) WriteByteArrays(values [][]byte) {
	w.Write4(int32(len(values)))
	for _, value := range values {
		w.WriteByteArray(value)
	}
}

// WriteInt8Array writes a length-prefixed int8 array.
func (w *Writer) WriteInt8Array(values []int8) {
	w.Write4(int32(len(values)))
	for _, value := range values {
		w.Write1(byte(value))
	}
}

// WriteInt16Array writes a length-prefixed little-endian int16 array.
func (w *Writer) WriteInt16Array(values []int16) {
	w.Write4(int32(len(values)))
	for _, value := range values {
		w.Write2(value)
	}
}

// WriteInt32Array writes a length-prefixed little-endian int32 array.
func (w *Writer) WriteInt32Array(values []int32) {
	w.Write4(int32(len(values)))
	for _, value := range values {
		w.Write4(value)
	}
}

// WriteInt64Array writes a length-prefixed little-endian int64 array.
func (w *Writer) WriteInt64Array(values []int64) {
	w.Write4(int32(len(values)))
	for _, value := range values {
		w.Write8(value)
	}
}

// WriteUint64Array writes a length-prefixed little-endian uint64 array.
func (w *Writer) WriteUint64Array(values []uint64) {
	w.Write4(int32(len(values)))
	for _, value := range values {
		w.Write8U(value)
	}
}

// Reader implements the Carbon binary reader contract used by the other SDKs.
type Reader struct {
	data []byte
	off  int
}

// NewReader creates a Carbon binary reader over data.
func NewReader(data []byte) *Reader {
	return &Reader{data: data}
}

// AssertEOF panics if unread bytes remain.
func (r *Reader) AssertEOF() {
	if r.off != len(r.data) {
		panic(fmt.Sprintf("unexpected trailing bytes: %d", len(r.data)-r.off))
	}
}

// ReadRaw reads count raw bytes.
func (r *Reader) ReadRaw(count int) []byte {
	if count < 0 || r.off+count > len(r.data) {
		panic("end of stream reached")
	}
	out := r.data[r.off : r.off+count]
	r.off += count
	return out
}

// ReadLength reads a Carbon array length.
func (r *Reader) ReadLength() int {
	length := r.Read4()
	if length < 0 {
		panic("negative array length")
	}
	if int64(length) > int64(len(r.data)-r.off) {
		panic(fmt.Sprintf("array length %d exceeds remaining bytes %d", length, len(r.data)-r.off))
	}
	return int(length)
}

// Read1 reads one byte.
func (r *Reader) Read1() byte {
	return r.ReadRaw(1)[0]
}

// Read2 reads a little-endian int16.
func (r *Reader) Read2() int16 {
	return int16(binary.LittleEndian.Uint16(r.ReadRaw(2)))
}

// Read4 reads a little-endian int32.
func (r *Reader) Read4() int32 {
	return int32(binary.LittleEndian.Uint32(r.ReadRaw(4)))
}

// Read4U reads a little-endian uint32.
func (r *Reader) Read4U() uint32 {
	return binary.LittleEndian.Uint32(r.ReadRaw(4))
}

// Read8 reads a little-endian int64.
func (r *Reader) Read8() int64 {
	return int64(binary.LittleEndian.Uint64(r.ReadRaw(8)))
}

// Read8U reads a little-endian uint64.
func (r *Reader) Read8U() uint64 {
	return binary.LittleEndian.Uint64(r.ReadRaw(8))
}

// Read16 reads a 16-byte Carbon value.
func (r *Reader) Read16() Bytes16 {
	return MustBytes16(r.ReadRaw(16))
}

// Read32 reads a 32-byte Carbon value.
func (r *Reader) Read32() Bytes32 {
	return MustBytes32(r.ReadRaw(32))
}

// Read64 reads a 64-byte Carbon value.
func (r *Reader) Read64() Bytes64 {
	return MustBytes64(r.ReadRaw(64))
}

// ReadBigInt reads a Carbon int256 value.
func (r *Reader) ReadBigInt() *big.Int {
	return r.ReadBigIntWithHeader(-1)
}

// ReadBigIntWithHeader reads a Carbon int256 value with an optional pre-read header.
func (r *Reader) ReadBigIntWithHeader(preReadHeader int) *big.Int {
	header := byte(preReadHeader)
	if preReadHeader < 0 {
		header = r.Read1()
	}

	if header == 0 {
		return big.NewInt(0)
	}

	length := int(header & 0x3f)
	if header&0x40 != 0 || length > 32 {
		panic("BigInt too big")
	}

	fill := byte(0x00)
	if header&0x80 != 0 {
		fill = 0xff
	}

	var word [32]byte
	if length > 0 {
		copy(word[:], r.ReadRaw(length))
	}
	for i := length; i < len(word); i++ {
		word[i] = fill
	}
	if word[31]&0x80 != header&0x80 {
		panic("non-standard BigInt header")
	}

	return intFromWord(word)
}

// ReadBigIntArray reads a length-prefixed int256 array.
func (r *Reader) ReadBigIntArray() []*big.Int {
	length := r.ReadLength()
	out := make([]*big.Int, length)
	for i := range out {
		out[i] = r.ReadBigInt()
	}
	return out
}

// ReadStringZ reads a zero-terminated UTF-8 string.
func (r *Reader) ReadStringZ() string {
	start := r.off
	for {
		if r.off >= len(r.data) {
			panic("end of stream reached")
		}
		if r.data[r.off] == 0 {
			break
		}
		r.off++
	}
	out := string(r.data[start:r.off])
	r.off++
	return out
}

// ReadStringZArray reads a length-prefixed zero-terminated string array.
func (r *Reader) ReadStringZArray() []string {
	length := r.ReadLength()
	out := make([]string, length)
	for i := range out {
		out[i] = r.ReadStringZ()
	}
	return out
}

// ReadByteArray reads a length-prefixed byte array.
func (r *Reader) ReadByteArray() []byte {
	length := r.ReadLength()
	raw := r.ReadRaw(length)
	out := make([]byte, length)
	copy(out, raw)
	return out
}

// ReadByteArrays reads a length-prefixed array of byte arrays.
func (r *Reader) ReadByteArrays() [][]byte {
	length := r.ReadLength()
	out := make([][]byte, length)
	for i := range out {
		out[i] = r.ReadByteArray()
	}
	return out
}

// ReadInt8Array reads a length-prefixed int8 array.
func (r *Reader) ReadInt8Array() []int8 {
	length := r.ReadLength()
	out := make([]int8, length)
	for i := range out {
		out[i] = int8(r.Read1())
	}
	return out
}

// ReadInt16Array reads a length-prefixed little-endian int16 array.
func (r *Reader) ReadInt16Array() []int16 {
	length := r.ReadLength()
	out := make([]int16, length)
	for i := range out {
		out[i] = r.Read2()
	}
	return out
}

// ReadInt32Array reads a length-prefixed little-endian int32 array.
func (r *Reader) ReadInt32Array() []int32 {
	length := r.ReadLength()
	out := make([]int32, length)
	for i := range out {
		out[i] = r.Read4()
	}
	return out
}

// ReadInt64Array reads a length-prefixed little-endian int64 array.
func (r *Reader) ReadInt64Array() []int64 {
	length := r.ReadLength()
	out := make([]int64, length)
	for i := range out {
		out[i] = r.Read8()
	}
	return out
}

// ReadUint64Array reads a length-prefixed little-endian uint64 array.
func (r *Reader) ReadUint64Array() []uint64 {
	length := r.ReadLength()
	out := make([]uint64, length)
	for i := range out {
		out[i] = r.Read8U()
	}
	return out
}

func bigIntWord(value *big.Int) [32]byte {
	if value.BitLen() > 256 {
		panic("BigInt overflow")
	}

	mod := new(big.Int).Lsh(big.NewInt(1), 256)
	unsigned := new(big.Int).Mod(new(big.Int).Set(value), mod)
	be := unsigned.Bytes()

	var word [32]byte
	for i := 0; i < len(be); i++ {
		word[i] = be[len(be)-1-i]
	}
	return word
}

func intFromWord(word [32]byte) *big.Int {
	be := make([]byte, 32)
	for i := range word {
		be[len(word)-1-i] = word[i]
	}

	value := new(big.Int).SetBytes(be)
	if word[31]&0x80 != 0 {
		mod := new(big.Int).Lsh(big.NewInt(1), 256)
		value.Sub(value, mod)
	}
	return value
}
