package vm

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"slices"
	"strconv"
	"unicode/utf16"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/domain/types"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/io"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/util"
)

type VMObject struct {
	Type VMType
	Data interface{}
}

type VMObjectPair struct {
	Key   VMObject
	Value VMObject
}

type VMObjectStruct []VMObjectPair

// AsNumber() returns value stored in vm.VMObject structure, in .Data field, as a *big.Int number
func (v *VMObject) AsNumber() *big.Int {
	switch v.Type {
	case None:
		return big.NewInt(0)

	case String:
		b, ok := new(big.Int).SetString(v.Data.(string), 10)
		if !ok {
			panic("Invalid cast: expected number, got String")
		}
		return b

	case Bytes:
		if len(v.Data.([]byte)) == 0 {
			return big.NewInt(0)
		}
		return util.BigIntFromCsharpOrPhantasmaByteArray(v.Data.([]byte))

	case Enum:
		return big.NewInt(int64(v.Data.(uint32)))

	case Bool:
		var val = v.Data.(bool)
		if val {
			return big.NewInt(1)
		} else {
			return big.NewInt(0)
		}

	case Number:
		n := v.Data.(big.Int)
		return &n

	case Timestamp:
		return big.NewInt(int64(v.Data.(types.Timestamp).Value))

	case Object:
		switch data := v.Data.(type) {
		case []byte:
			if len(data) != 32 {
				panic("Invalid cast: expected number, got Object")
			}
			unsigned := append([]byte(nil), data...)
			slices.Reverse(unsigned)
			return new(big.Int).SetBytes(unsigned)
		case cryptography.Address, *cryptography.Address:
			panic("Invalid cast: expected number, got Object")
		default:
			panic("Invalid cast: expected number, got Object")
		}

	default:
		panic("Unsupported type")
	}
}

func (v *VMObject) AsBytes() []byte {
	switch v.Type {
	case Bytes:
		return append([]byte(nil), v.Data.([]byte)...)
	case String:
		return []byte(v.Data.(string))
	case Bool:
		if v.Data.(bool) {
			return []byte{1}
		}
		return []byte{0}
	case Enum:
		out := make([]byte, 4)
		binary.LittleEndian.PutUint32(out, v.Data.(uint32))
		return out
	case Number:
		n := v.Data.(big.Int)
		return util.BigIntToPhantasmaByteArray(&n)
	case Timestamp:
		out := make([]byte, 4)
		binary.LittleEndian.PutUint32(out, v.Data.(types.Timestamp).Value)
		return out
	case Object:
		switch data := v.Data.(type) {
		case []byte:
			return append([]byte(nil), data...)
		case cryptography.Address:
			return data.Bytes()
		case *cryptography.Address:
			return data.Bytes()
		default:
			panic("Invalid cast: expected bytes, got Object")
		}
	case Struct:
		return serializeVMObject(v)
	default:
		panic("Invalid cast: expected bytes, got " + VMTypeLookup[v.Type])
	}
}

func (v *VMObject) AsBool() bool {
	switch v.Type {
	case Bool:
		return v.Data.(bool)
	case Number:
		return v.AsNumber().Sign() != 0
	case Bytes:
		data := v.Data.([]byte)
		if len(data) == 1 {
			return data[0] != 0
		}
		panic("Invalid cast: expected bool, got Bytes")
	default:
		panic("Invalid cast: expected bool, got " + VMTypeLookup[v.Type])
	}
}

// AsString() returns value stored in vm.VMObject structure, in .Data field, as a string
func (v *VMObject) AsString() string {
	switch v.Type {
	case None:
		return "Null"

	case String:
		return v.Data.(string)

	case Bytes:
		return string(v.Data.([]byte))

	case Enum:
		return strconv.FormatInt(int64(v.Data.(uint32)), 10)

	case Bool:
		var val = v.Data.(bool)
		if val {
			return "true"
		} else {
			return "false"
		}

	case Number:
		n := v.Data.(big.Int)
		return n.String()

	case Timestamp:
		return strconv.FormatInt(int64(v.Data.(types.Timestamp).Value), 10)

	case Struct:
		if v.GetArrayType() == Number {
			children := v.structChildren()
			codeUnits := make([]uint16, 0, len(children))
			for i := 0; i < len(children); i++ {
				value := v.arrayValue(i)
				if value == nil {
					panic("Invalid cast: expected string, got Struct")
				}
				codeUnits = append(codeUnits, uint16(value.AsNumber().Uint64()))
			}
			return string(utf16.Decode(codeUnits))
		}
		return base64.StdEncoding.EncodeToString(v.AsBytes())

	default:
		panic("Unsupported type")
	}
}

func (v *VMObject) GetArrayType() VMType {
	if v.Type != Struct {
		return None
	}
	children := v.structChildren()
	result := None
	for i := 0; i < len(children); i++ {
		value := v.arrayValue(i)
		if value == nil {
			return None
		}
		if result == None {
			result = value.Type
		} else if value.Type != result {
			return None
		}
	}
	return result
}

// CastTo applies the same VMObject conversion rules used by the classic VM
// object model. It returns a new object so callers do not accidentally mutate
// the original value while preparing interop arguments or fixture comparisons.
func (v *VMObject) CastTo(vmtype VMType) *VMObject {
	out := &VMObject{}
	if v.Type == vmtype {
		out.Copy(v)
		return out
	}

	switch vmtype {
	case None:
		return out
	case Bool:
		out.Type = Bool
		out.Data = v.AsBool()
	case Bytes:
		out.Type = Bytes
		out.Data = v.AsBytes()
	case Number:
		out.Type = Number
		out.Data = *v.AsNumber()
	case String:
		out.Type = String
		out.Data = v.AsString()
	case Object:
		if v.Type != Object {
			panic("invalid cast: " + VMTypeLookup[v.Type] + " to Object")
		}
		out.Copy(v)
	case Struct:
		switch v.Type {
		case String:
			runes := []rune(v.Data.(string))
			codeUnits := utf16.Encode(runes)
			children := make(VMObjectStruct, len(codeUnits))
			for i, codeUnit := range codeUnits {
				children[i] = VMObjectPair{
					Key:   VMObject{Type: Number, Data: *big.NewInt(int64(i))},
					Value: VMObject{Type: Number, Data: *big.NewInt(int64(codeUnit))},
				}
			}
			out.Type = Struct
			out.Data = children
		case Object:
			out.Copy(v)
		default:
			panic("invalid cast: " + VMTypeLookup[v.Type] + " to Struct")
		}
	default:
		panic("unsupported cast target: " + VMTypeLookup[vmtype])
	}
	return out
}

func (v *VMObject) String() string {
	switch v.Type {
	case None:
		return "Null"
	case Struct:
		return "[Struct]"
	case Bytes:
		return "[Bytes] => " + hex.EncodeToString((v.Data.([]byte)))
	case Number:
		return "[Number] => " + v.AsString()
	case Timestamp:
		return "[Time] => " + v.AsString()
	case String:
		return "[String] => " + v.AsString()
	case Bool:
		return "[Bool] => " + v.AsString()
	case Enum:
		return "[Enum] => " + v.AsString()
	case Object:
		var r string
		if v.Data == nil {
			r = "null"
		} else {
			r = "object"
		}
		return r
	default:
		return "Unknown"
	}
}

func (v *VMObject) SetValue(val []byte, vmtype VMType) *VMObject {
	v.Type = vmtype
	// if val != nil {
	// 	v._localSize = len(val)
	// }

	switch vmtype {
	case Bytes:
		{
			v.Data = val
			break
		}

	case Number:
		{
			var n *big.Int
			if len(val) == 0 {
				n = big.NewInt(0)
			} else {
				n = util.BigIntFromCsharpOrPhantasmaByteArray(val)
			}

			v.Data = *n
			break
		}

	case String:
		{
			v.Data = string(val)
			break
		}

	case Enum:
		{
			switch {
			case len(val) == 0:
				v.Data = uint32(0)
			case len(val) <= 4:
				var padded [4]byte
				copy(padded[:], val)
				v.Data = binary.LittleEndian.Uint32(padded[:])
			default:
				panic("Invalid enum byte length")
			}
			break
		}

	case Timestamp:
		{
			var n uint32
			if val == nil {
				n = 0
			} else {
				n = binary.LittleEndian.Uint32(val)
			}
			v.Data = types.Timestamp{Value: n}
			break
		}

	case Bool:
		{
			v.Data = val[0] != 0
			break
		}

	default:
		panic("Unsupported value type")
	}

	return v
}

func (v *VMObject) Copy(other *VMObject) {
	if other == nil || other.Type == None {
		v.Type = None
		v.Data = nil
		return
	}

	v.Type = other.Type

	if other.Type == Struct {
		children := other.structChildren()
		copied := make(VMObjectStruct, len(children))
		for i, pair := range children {
			copied[i].Key.Copy(&pair.Key)
			copied[i].Value.Copy(&pair.Value)
		}
		v.Data = copied
	} else {
		v.Data = other.Data
	}
}

// Serialize implements ther Serializable interface
func (v *VMObject) Serialize(writer *io.BinWriter) {
	writer.WriteB(byte(v.Type))
	switch v.Type {
	case None:
		return
	case Bool:
		writer.WriteBool(v.AsBool())
	case String:
		writer.WriteString(v.Data.(string))
	case Bytes:
		writer.WriteVarBytes(v.Data.([]byte))
	case Enum:
		writer.WriteVarUint(uint64(v.Data.(uint32)))
	case Number:
		n := v.Data.(big.Int)
		writer.WriteBigInteger(&n)
	case Timestamp:
		t := v.Data.(types.Timestamp)
		writer.WriteTimestamp(&t)
	case Object:
		inner := io.NewBufBinWriter()
		switch data := v.Data.(type) {
		case cryptography.Address:
			data.Serialize(inner.BinWriter)
		case []byte:
			inner.WriteVarBytes(data)
		case io.Serializer:
			data.Serialize(inner.BinWriter)
		default:
			panic("Unsupported object serialization")
		}
		writer.WriteVarBytes(inner.Bytes())
	case Struct:
		children := v.structChildren()
		writer.WriteVarUint(uint64(len(children)))
		for _, pair := range children {
			pair.Key.Serialize(writer)
			pair.Value.Serialize(writer)
		}
	}
}

func ValidateStructKey(key *VMObject) {
	if key.Type == None {
		panic("Cannot use value of type None as key for struct field")
	}
	if key.Type == Struct {
		panic("Cannot use value of type Struct as key for struct field")
	}
	if key.Type == Object {
		panic("Cannot use value of type Object as key for struct field")
	}
}

// Deserialize implements ther Serializable interface
func (v *VMObject) Deserialize(reader *io.BinReader) {
	v.Type = VMType(reader.ReadB())

	switch v.Type {
	case Bool:
		v.Data = reader.ReadBool()
	case Bytes:
		v.Data = reader.ReadVarBytes()
	case Enum:
		v.Data = uint32(reader.ReadVarUint())
	case Number:
		v.Data = *reader.ReadBigInteger()
	case Object:
		bytes := reader.ReadVarBytes()
		if len(bytes) == 35 {
			v.Data = io.Deserialize[*cryptography.Address](bytes)
			v.Type = Object
		} else {
			// NOTE object type information is lost during serialization, so we reconstruct it as byte array
			v.Type = Bytes
			v.Data = bytes
		}
	case String:
		v.Data = reader.ReadString()
	case Struct:
		childCount := reader.ReadVarUint()
		children := make(VMObjectStruct, 0, childCount)
		for {
			if childCount == 0 {
				break
			}

			key := &VMObject{}
			key.Deserialize(reader)

			ValidateStructKey(key)

			val := &VMObject{}
			val.Deserialize(reader)

			children = append(children, VMObjectPair{Key: *key, Value: *val})
			childCount--
		}

		v.Data = children
	case Timestamp:
		v.Data = *reader.ReadTimestamp()
	}
}

func (v *VMObject) structChildren() VMObjectStruct {
	switch children := v.Data.(type) {
	case VMObjectStruct:
		return children
	case []VMObjectPair:
		return VMObjectStruct(children)
	case map[VMObject]VMObject:
		out := make(VMObjectStruct, 0, len(children))
		for key, value := range children {
			out = append(out, VMObjectPair{Key: key, Value: value})
		}
		return out
	default:
		return nil
	}
}

func (v *VMObject) arrayValue(index int) *VMObject {
	for _, pair := range v.structChildren() {
		if pair.Key.Type == Number && pair.Key.AsNumber().Cmp(big.NewInt(int64(index))) == 0 {
			value := pair.Value
			return &value
		}
	}
	return nil
}

func serializeVMObject(v *VMObject) []byte {
	writer := io.NewBufBinWriter()
	v.Serialize(writer.BinWriter)
	return writer.Bytes()
}
