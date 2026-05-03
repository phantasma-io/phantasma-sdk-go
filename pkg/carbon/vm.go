package carbon

import (
	"fmt"
	"math/big"
	"sort"
)

// VMType identifies a Carbon VM metadata value type.
type VMType byte

// Carbon VM value type ids.
const (
	// VMTypeDynamic stores a dynamically typed value.
	VMTypeDynamic VMType = 0
	VMTypeArray   VMType = 1
	VMTypeBytes   VMType = 1 << 1
	VMTypeStruct  VMType = 2 << 1
	VMTypeInt8    VMType = 3 << 1
	VMTypeInt16   VMType = 4 << 1
	VMTypeInt32   VMType = 5 << 1
	VMTypeInt64   VMType = 6 << 1
	VMTypeInt256  VMType = 7 << 1
	VMTypeBytes16 VMType = 8 << 1
	VMTypeBytes32 VMType = 9 << 1
	VMTypeBytes64 VMType = 10 << 1
	VMTypeString  VMType = 11 << 1

	VMTypeArrayDynamic VMType = VMTypeArray | VMTypeDynamic
	VMTypeArrayBytes   VMType = VMTypeArray | VMTypeBytes
	VMTypeArrayStruct  VMType = VMTypeArray | VMTypeStruct
	VMTypeArrayInt8    VMType = VMTypeArray | VMTypeInt8
	VMTypeArrayInt16   VMType = VMTypeArray | VMTypeInt16
	VMTypeArrayInt32   VMType = VMTypeArray | VMTypeInt32
	VMTypeArrayInt64   VMType = VMTypeArray | VMTypeInt64
	VMTypeArrayInt256  VMType = VMTypeArray | VMTypeInt256
	VMTypeArrayBytes16 VMType = VMTypeArray | VMTypeBytes16
	VMTypeArrayBytes32 VMType = VMTypeArray | VMTypeBytes32
	VMTypeArrayBytes64 VMType = VMTypeArray | VMTypeBytes64
	VMTypeArrayString  VMType = VMTypeArray | VMTypeString
)

// VMStructFlags control Carbon VM struct schema encoding.
type VMStructFlags byte

// Carbon VM struct schema flags.
const (
	// VMStructFlagsNone disables optional struct schema behavior.
	VMStructFlagsNone VMStructFlags = 0
	// VMStructFlagsDynamicExtras allows fields not declared in the schema.
	VMStructFlagsDynamicExtras VMStructFlags = 1 << 0
	// VMStructFlagsIsSorted preserves schema field order instead of sorting.
	VMStructFlagsIsSorted VMStructFlags = 1 << 1
)

// VMVariableSchema describes one VM metadata value schema.
type VMVariableSchema struct {
	Type      VMType
	StructDef *VMStructSchema
}

// NewVMVariableSchema returns a VM value schema.
func NewVMVariableSchema(vmType VMType, structDef *VMStructSchema) VMVariableSchema {
	return VMVariableSchema{Type: vmType, StructDef: structDef}
}

// WriteCarbon writes the VM variable schema to w.
func (s *VMVariableSchema) WriteCarbon(w *Writer) {
	w.Write1(byte(s.Type))
	if s.Type == VMTypeStruct || s.Type == VMTypeArrayStruct {
		if s.StructDef == nil {
			(&VMStructSchema{}).WriteCarbon(w)
			return
		}
		s.StructDef.WriteCarbon(w)
	}
}

// ReadCarbon reads the VM variable schema from r.
func (s *VMVariableSchema) ReadCarbon(r *Reader) {
	s.Type = VMType(r.Read1())
	if s.Type == VMTypeStruct || s.Type == VMTypeArrayStruct {
		s.StructDef = &VMStructSchema{}
		s.StructDef.ReadCarbon(r)
	} else {
		s.StructDef = nil
	}
}

// VMNamedVariableSchema binds a metadata field name to its schema.
type VMNamedVariableSchema struct {
	Name   SmallString
	Schema VMVariableSchema
}

// NewVMNamedVariableSchema validates and returns a named VM variable schema.
func NewVMNamedVariableSchema(name string, vmType VMType) (VMNamedVariableSchema, error) {
	smallName, err := NewSmallString(name)
	if err != nil {
		return VMNamedVariableSchema{}, err
	}
	return VMNamedVariableSchema{
		Name:   smallName,
		Schema: NewVMVariableSchema(vmType, nil),
	}, nil
}

// MustVMNamedVariableSchema returns a VM variable schema and panics if the name is invalid.
func MustVMNamedVariableSchema(name string, vmType VMType) VMNamedVariableSchema {
	out, err := NewVMNamedVariableSchema(name, vmType)
	if err != nil {
		panic(err)
	}
	return out
}

// WriteCarbon writes the named VM variable schema to w.
func (s *VMNamedVariableSchema) WriteCarbon(w *Writer) {
	s.Name.WriteCarbon(w)
	s.Schema.WriteCarbon(w)
}

// ReadCarbon reads the named VM variable schema from r.
func (s *VMNamedVariableSchema) ReadCarbon(r *Reader) {
	s.Name.ReadCarbon(r)
	s.Schema.ReadCarbon(r)
}

// VMStructSchema describes a named-field VM metadata struct.
type VMStructSchema struct {
	Fields []VMNamedVariableSchema
	Flags  VMStructFlags
}

// WriteCarbon writes the VM struct schema to w.
func (s *VMStructSchema) WriteCarbon(w *Writer) {
	w.Write4(int32(len(s.Fields)))
	for i := range s.Fields {
		s.Fields[i].WriteCarbon(w)
	}
	w.Write1(byte(s.Flags))
}

// ReadCarbon reads the VM struct schema from r.
func (s *VMStructSchema) ReadCarbon(r *Reader) {
	count := r.ReadLength()
	s.Fields = make([]VMNamedVariableSchema, count)
	for i := range s.Fields {
		s.Fields[i].ReadCarbon(r)
	}
	s.Flags = VMStructFlags(r.Read1())
}

// VMNamedDynamicVariable binds a metadata field name to a dynamic value.
type VMNamedDynamicVariable struct {
	Name  SmallString
	Value VMDynamicVariable
}

// NewVMNamedDynamicVariable validates and returns a named VM dynamic value.
func NewVMNamedDynamicVariable(name string, vmType VMType, value any) (VMNamedDynamicVariable, error) {
	smallName, err := NewSmallString(name)
	if err != nil {
		return VMNamedDynamicVariable{}, err
	}
	return VMNamedDynamicVariable{
		Name:  smallName,
		Value: NewVMDynamicVariable(vmType, value),
	}, nil
}

// MustVMNamedDynamicVariable returns a named VM dynamic value and panics if inputs are invalid.
func MustVMNamedDynamicVariable(name string, vmType VMType, value any) VMNamedDynamicVariable {
	out, err := NewVMNamedDynamicVariable(name, vmType, value)
	if err != nil {
		panic(err)
	}
	return out
}

// WriteCarbon writes the named VM dynamic value to w.
func (v *VMNamedDynamicVariable) WriteCarbon(w *Writer) {
	v.Name.WriteCarbon(w)
	v.Value.WriteCarbon(w)
}

// ReadCarbon reads the named VM dynamic value from r.
func (v *VMNamedDynamicVariable) ReadCarbon(r *Reader) {
	v.Name.ReadCarbon(r)
	v.Value.ReadCarbon(r)
}

// VMDynamicStruct stores named VM metadata values.
type VMDynamicStruct struct {
	Fields []VMNamedDynamicVariable
}

// WriteCarbon writes the dynamic struct to w.
func (s *VMDynamicStruct) WriteCarbon(w *Writer) {
	s.sortFields()
	w.Write4(int32(len(s.Fields)))
	for i := range s.Fields {
		s.Fields[i].WriteCarbon(w)
	}
}

// ReadCarbon reads the dynamic struct from r.
func (s *VMDynamicStruct) ReadCarbon(r *Reader) {
	count := r.ReadLength()
	s.Fields = make([]VMNamedDynamicVariable, count)
	for i := range s.Fields {
		s.Fields[i].ReadCarbon(r)
	}
	s.sortFields()
}

// Get returns the dynamic value for name, or nil if it is absent.
func (s *VMDynamicStruct) Get(name string) *VMDynamicVariable {
	for i := range s.Fields {
		if s.Fields[i].Name.String() == name {
			return &s.Fields[i].Value
		}
	}
	return nil
}

// WriteWithSchema writes the dynamic struct using a static schema.
func (s *VMDynamicStruct) WriteWithSchema(schema *VMStructSchema, w *Writer) bool {
	ok := true
	fieldsFound := 0
	for i := range schema.Fields {
		fieldSchema := schema.Fields[i]
		field := s.Get(fieldSchema.Name.String())
		if field == nil {
			defaultValue := NewVMDynamicVariable(fieldSchema.Schema.Type, nil)
			ok = defaultValue.writeStatic(fieldSchema.Schema.Type, fieldSchema.Schema.StructDef, w) && ok
			continue
		}
		if field.writeStatic(fieldSchema.Schema.Type, fieldSchema.Schema.StructDef, w) {
			fieldsFound++
		} else {
			ok = false
		}
	}

	if schema.Flags&VMStructFlagsDynamicExtras == 0 {
		if schema.Flags&VMStructFlagsIsSorted == 0 {
			s.sortFields()
		}
		return ok
	}

	if fieldsFound == len(schema.Fields) && len(s.Fields) == len(schema.Fields) {
		w.Write4U(0)
		return ok
	}

	extras := make([]VMNamedDynamicVariable, 0)
	for _, field := range s.Fields {
		if !schemaHasField(schema, field.Name.String()) {
			extras = append(extras, field)
		}
	}
	writeNamedDynamicVariables(w, extras)
	return ok
}

// ReadWithSchema reads the dynamic struct using a static schema.
func (s *VMDynamicStruct) ReadWithSchema(schema *VMStructSchema, r *Reader) {
	s.Fields = make([]VMNamedDynamicVariable, len(schema.Fields))
	for i := range schema.Fields {
		s.Fields[i].Name = schema.Fields[i].Name
		s.Fields[i].Value.ReadWithSchema(&schema.Fields[i].Schema, r)
	}
	if schema.Flags&VMStructFlagsDynamicExtras == 0 {
		if schema.Flags&VMStructFlagsIsSorted == 0 {
			s.sortFields()
		}
		return
	}

	extras := readNamedDynamicVariables(r)
	s.Fields = append(s.Fields, extras...)
	s.sortFields()
}

func (s *VMDynamicStruct) sortFields() {
	if len(s.Fields) <= 1 {
		return
	}
	sort.Slice(s.Fields, func(i, j int) bool {
		return s.Fields[i].Name.String() < s.Fields[j].Name.String()
	})
}

// VMStructArray stores repeated VM structs sharing the same schema.
type VMStructArray struct {
	Schema  VMStructSchema
	Structs []VMDynamicStruct
}

// VMDynamicVariable stores one typed VM metadata value.
type VMDynamicVariable struct {
	Type VMType
	Data any
}

// NewVMDynamicVariable returns a dynamic VM value, using the type default when value is nil.
func NewVMDynamicVariable(vmType VMType, value any) VMDynamicVariable {
	if value == nil {
		value = defaultVMValue(vmType)
	}
	return VMDynamicVariable{Type: vmType, Data: value}
}

// WriteCarbon writes the dynamic VM value to w.
func (v *VMDynamicVariable) WriteCarbon(w *Writer) {
	w.Write1(byte(v.Type))
	v.writeStatic(v.Type, nil, w)
}

// ReadCarbon reads the dynamic VM value from r.
func (v *VMDynamicVariable) ReadCarbon(r *Reader) {
	v.Type = VMType(r.Read1())
	v.readStatic(v.Type, nil, r)
}

// ReadWithSchema reads the VM value using a static schema.
func (v *VMDynamicVariable) ReadWithSchema(schema *VMVariableSchema, r *Reader) {
	v.Type = schema.Type
	v.readStatic(schema.Type, schema.StructDef, r)
}

func (v *VMDynamicVariable) writeStatic(vmType VMType, schema *VMStructSchema, w *Writer) bool {
	if v.Type != vmType {
		defaultValue := NewVMDynamicVariable(vmType, nil)
		return defaultValue.writeStatic(vmType, schema, w)
	}

	if vmType == VMTypeDynamic {
		if v.Data == nil {
			w.Write1(byte(VMTypeArrayDynamic))
			w.Write4U(0)
			return true
		}
		inner := expectType[VMDynamicVariable](v.Data, vmType)
		inner.WriteCarbon(w)
		return true
	}
	if v.Data == nil {
		panic("invalid VM dynamic variable")
	}

	switch vmType {
	case VMTypeBytes:
		w.WriteByteArray(expectBytes(v.Data, vmType))
	case VMTypeStruct:
		structValue := expectStruct(v.Data, vmType)
		if schema != nil {
			return structValue.WriteWithSchema(schema, w)
		}
		structValue.WriteCarbon(w)
	case VMTypeInt8:
		w.Write1(byte(expectInt64(v.Data, vmType)))
	case VMTypeInt16:
		w.Write2(int16(expectInt64(v.Data, vmType)))
	case VMTypeInt32:
		w.Write4(int32(expectInt64(v.Data, vmType)))
	case VMTypeInt64:
		w.Write8(expectInt64(v.Data, vmType))
	case VMTypeInt256:
		w.WriteBigInt(expectBigInt(v.Data, vmType))
	case VMTypeBytes16:
		w.Write16(expectType[Bytes16](v.Data, vmType))
	case VMTypeBytes32:
		w.Write32(expectType[Bytes32](v.Data, vmType))
	case VMTypeBytes64:
		w.Write64(expectType[Bytes64](v.Data, vmType))
	case VMTypeString:
		w.WriteStringZ(expectString(v.Data, vmType))
	case VMTypeArrayDynamic:
		writeDynamicVariables(w, expectType[[]VMDynamicVariable](v.Data, vmType))
	case VMTypeArrayBytes:
		w.WriteByteArrays(expectType[[][]byte](v.Data, vmType))
	case VMTypeArrayStruct:
		array := expectType[VMStructArray](v.Data, vmType)
		w.Write4U(uint32(len(array.Structs)))
		usedSchema := schema
		if usedSchema == nil {
			array.Schema.WriteCarbon(w)
			if len(array.Schema.Fields) > 0 {
				usedSchema = &array.Schema
			}
		}
		ok := true
		for i := range array.Structs {
			if usedSchema != nil {
				ok = array.Structs[i].WriteWithSchema(usedSchema, w) && ok
			} else {
				array.Structs[i].WriteCarbon(w)
			}
		}
		return ok
	case VMTypeArrayInt8:
		w.WriteInt8Array(expectType[[]int8](v.Data, vmType))
	case VMTypeArrayInt16:
		w.WriteInt16Array(expectType[[]int16](v.Data, vmType))
	case VMTypeArrayInt32:
		w.WriteInt32Array(expectType[[]int32](v.Data, vmType))
	case VMTypeArrayInt64:
		w.WriteInt64Array(expectType[[]int64](v.Data, vmType))
	case VMTypeArrayInt256:
		w.WriteBigIntArray(expectType[[]*big.Int](v.Data, vmType))
	case VMTypeArrayBytes16:
		writeBytes16Array(w, expectType[[]Bytes16](v.Data, vmType))
	case VMTypeArrayBytes32:
		writeBytes32Array(w, expectType[[]Bytes32](v.Data, vmType))
	case VMTypeArrayBytes64:
		writeBytes64Array(w, expectType[[]Bytes64](v.Data, vmType))
	case VMTypeArrayString:
		w.WriteStringZArray(expectType[[]string](v.Data, vmType))
	default:
		panic(fmt.Sprintf("invalid VM dynamic variable type %d", vmType))
	}
	return true
}

func (v *VMDynamicVariable) readStatic(vmType VMType, schema *VMStructSchema, r *Reader) {
	switch vmType {
	case VMTypeDynamic:
		var inner VMDynamicVariable
		inner.ReadCarbon(r)
		v.Data = inner
	case VMTypeBytes:
		v.Data = r.ReadByteArray()
	case VMTypeStruct:
		var s VMDynamicStruct
		if schema != nil {
			s.ReadWithSchema(schema, r)
		} else {
			s.ReadCarbon(r)
		}
		v.Data = s
	case VMTypeInt8:
		v.Data = int8(r.Read1())
	case VMTypeInt16:
		v.Data = r.Read2()
	case VMTypeInt32:
		v.Data = r.Read4()
	case VMTypeInt64:
		v.Data = r.Read8()
	case VMTypeInt256:
		v.Data = r.ReadBigInt()
	case VMTypeBytes16:
		v.Data = r.Read16()
	case VMTypeBytes32:
		v.Data = r.Read32()
	case VMTypeBytes64:
		v.Data = r.Read64()
	case VMTypeString:
		v.Data = r.ReadStringZ()
	case VMTypeArrayDynamic:
		v.Data = readDynamicVariables(r)
	case VMTypeArrayBytes:
		v.Data = r.ReadByteArrays()
	case VMTypeArrayStruct:
		count := r.ReadLength()
		usedSchema := schema
		array := VMStructArray{}
		if usedSchema == nil {
			array.Schema.ReadCarbon(r)
			if len(array.Schema.Fields) > 0 {
				usedSchema = &array.Schema
			}
		} else {
			array.Schema = *usedSchema
		}
		array.Structs = make([]VMDynamicStruct, count)
		for i := range array.Structs {
			if usedSchema != nil {
				array.Structs[i].ReadWithSchema(usedSchema, r)
			} else {
				array.Structs[i].ReadCarbon(r)
			}
		}
		v.Data = array
	case VMTypeArrayInt8:
		v.Data = r.ReadInt8Array()
	case VMTypeArrayInt16:
		v.Data = r.ReadInt16Array()
	case VMTypeArrayInt32:
		v.Data = r.ReadInt32Array()
	case VMTypeArrayInt64:
		v.Data = r.ReadInt64Array()
	case VMTypeArrayInt256:
		v.Data = r.ReadBigIntArray()
	case VMTypeArrayBytes16:
		v.Data = readBytes16Array(r)
	case VMTypeArrayBytes32:
		v.Data = readBytes32Array(r)
	case VMTypeArrayBytes64:
		v.Data = readBytes64Array(r)
	case VMTypeArrayString:
		v.Data = r.ReadStringZArray()
	default:
		panic(fmt.Sprintf("unknown VM dynamic variable type %d", vmType))
	}
}

func defaultVMValue(vmType VMType) any {
	switch vmType {
	case VMTypeDynamic:
		return nil
	case VMTypeBytes:
		return []byte{}
	case VMTypeStruct:
		return VMDynamicStruct{}
	case VMTypeInt8:
		return int8(0)
	case VMTypeInt16:
		return int16(0)
	case VMTypeInt32:
		return int32(0)
	case VMTypeInt64:
		return int64(0)
	case VMTypeInt256:
		return big.NewInt(0)
	case VMTypeBytes16:
		return EmptyBytes16
	case VMTypeBytes32:
		return EmptyBytes32
	case VMTypeBytes64:
		return EmptyBytes64
	case VMTypeString:
		return ""
	case VMTypeArrayDynamic:
		return []VMDynamicVariable{}
	case VMTypeArrayBytes:
		return [][]byte{}
	case VMTypeArrayStruct:
		return VMStructArray{}
	case VMTypeArrayInt8:
		return []int8{}
	case VMTypeArrayInt16:
		return []int16{}
	case VMTypeArrayInt32:
		return []int32{}
	case VMTypeArrayInt64:
		return []int64{}
	case VMTypeArrayInt256:
		return []*big.Int{}
	case VMTypeArrayBytes16:
		return []Bytes16{}
	case VMTypeArrayBytes32:
		return []Bytes32{}
	case VMTypeArrayBytes64:
		return []Bytes64{}
	case VMTypeArrayString:
		return []string{}
	default:
		panic(fmt.Sprintf("unknown VM type %d", vmType))
	}
}

func schemaHasField(schema *VMStructSchema, name string) bool {
	for i := range schema.Fields {
		if schema.Fields[i].Name.String() == name {
			return true
		}
	}
	return false
}

func writeNamedDynamicVariables(w *Writer, values []VMNamedDynamicVariable) {
	w.Write4(int32(len(values)))
	for i := range values {
		values[i].WriteCarbon(w)
	}
}

func readNamedDynamicVariables(r *Reader) []VMNamedDynamicVariable {
	count := r.ReadLength()
	out := make([]VMNamedDynamicVariable, count)
	for i := range out {
		out[i].ReadCarbon(r)
	}
	return out
}

func writeDynamicVariables(w *Writer, values []VMDynamicVariable) {
	w.Write4(int32(len(values)))
	for i := range values {
		values[i].WriteCarbon(w)
	}
}

func readDynamicVariables(r *Reader) []VMDynamicVariable {
	count := r.ReadLength()
	out := make([]VMDynamicVariable, count)
	for i := range out {
		out[i].ReadCarbon(r)
	}
	return out
}

func writeBytes16Array(w *Writer, values []Bytes16) {
	w.Write4(int32(len(values)))
	for _, value := range values {
		w.Write16(value)
	}
}

func readBytes16Array(r *Reader) []Bytes16 {
	count := r.ReadLength()
	out := make([]Bytes16, count)
	for i := range out {
		out[i] = r.Read16()
	}
	return out
}

func writeBytes32Array(w *Writer, values []Bytes32) {
	w.Write4(int32(len(values)))
	for _, value := range values {
		w.Write32(value)
	}
}

func readBytes32Array(r *Reader) []Bytes32 {
	count := r.ReadLength()
	out := make([]Bytes32, count)
	for i := range out {
		out[i] = r.Read32()
	}
	return out
}

func writeBytes64Array(w *Writer, values []Bytes64) {
	w.Write4(int32(len(values)))
	for _, value := range values {
		w.Write64(value)
	}
}

func readBytes64Array(r *Reader) []Bytes64 {
	count := r.ReadLength()
	out := make([]Bytes64, count)
	for i := range out {
		out[i] = r.Read64()
	}
	return out
}

func expectBytes(value any, vmType VMType) []byte {
	switch v := value.(type) {
	case []byte:
		return v
	case Bytes16:
		return v.Bytes()
	case Bytes32:
		return v.Bytes()
	case Bytes64:
		return v.Bytes()
	default:
		panic(fmt.Sprintf("VM type %d expects bytes, got %T", vmType, value))
	}
}

func expectStruct(value any, vmType VMType) *VMDynamicStruct {
	switch v := value.(type) {
	case VMDynamicStruct:
		return &v
	case *VMDynamicStruct:
		return v
	default:
		panic(fmt.Sprintf("VM type %d expects struct, got %T", vmType, value))
	}
}

func expectString(value any, vmType VMType) string {
	v, ok := value.(string)
	if !ok {
		panic(fmt.Sprintf("VM type %d expects string, got %T", vmType, value))
	}
	return v
}

func expectInt64(value any, vmType VMType) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		if v > uint64(maxI64.Int64()) {
			panic(fmt.Sprintf("VM type %d integer overflows int64", vmType))
		}
		return int64(v)
	default:
		panic(fmt.Sprintf("VM type %d expects integer, got %T", vmType, value))
	}
}

func expectBigInt(value any, vmType VMType) *big.Int {
	switch v := value.(type) {
	case *big.Int:
		return v
	case big.Int:
		return &v
	case IntX:
		return v.BigInt()
	case int64:
		return big.NewInt(v)
	case uint64:
		return new(big.Int).SetUint64(v)
	case int:
		return big.NewInt(int64(v))
	default:
		panic(fmt.Sprintf("VM type %d expects BigInt, got %T", vmType, value))
	}
}

func expectType[T any](value any, vmType VMType) T {
	v, ok := value.(T)
	if !ok {
		panic(fmt.Sprintf("VM type %d expects %T, got %T", vmType, *new(T), value))
	}
	return v
}
