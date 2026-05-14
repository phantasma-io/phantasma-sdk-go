package scriptbuilder

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	crypto "github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/io"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/util"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/vm"
)

const maxRegisterCount = 32

// ScriptBuilder incrementally builds Phantasma VM bytecode.
type ScriptBuilder struct {
	writer         *io.BufBinWriter
	jumpLocations  map[int]string
	labelLocations map[string]int
}

// BeginScript creates a new VM script builder.
func BeginScript() ScriptBuilder {
	return NewScriptBuilder()
}

// NewScriptBuilder creates a new VM script builder.
func NewScriptBuilder() ScriptBuilder {
	return ScriptBuilder{
		writer:         io.NewBufBinWriter(),
		jumpLocations:  make(map[int]string),
		labelLocations: make(map[string]int),
	}
}

// EndScript appends RET and returns the final script, panicking if the builder has an error.
func (s ScriptBuilder) EndScript() []byte {
	script, err := s.EndScriptWithError()
	if err != nil {
		panic(err)
	}
	return script
}

// EndScriptWithError appends RET and returns the final script or any builder error.
func (s ScriptBuilder) EndScriptWithError() ([]byte, error) {
	s.EmitS(vm.RET)
	return s.ToScript()
}

// ToScript resolves labels and returns the current script without appending RET.
func (s ScriptBuilder) ToScript() ([]byte, error) {
	if s.writer == nil {
		return nil, errors.New("script builder is not initialized")
	}

	script := s.writer.Bytes()
	if script == nil {
		if s.writer.Err != nil {
			return nil, s.writer.Err
		}
		return nil, errors.New("script builder returned empty script buffer")
	}

	for targetOffset, label := range s.jumpLocations {
		labelOffset, ok := s.labelLocations[normalizeLabel(label)]
		if !ok {
			return nil, fmt.Errorf("could not find label: %s", label)
		}
		if labelOffset > 0xFFFF {
			return nil, fmt.Errorf("label offset exceeds uint16: %d", labelOffset)
		}
		if targetOffset < 0 || targetOffset+1 >= len(script) {
			return nil, fmt.Errorf("invalid jump patch offset: %d", targetOffset)
		}

		binary.LittleEndian.PutUint16(script[targetOffset:], uint16(labelOffset))
	}

	return script, nil
}

// CurrentSize returns the current byte length of the script.
func (s ScriptBuilder) CurrentSize() int {
	if s.writer == nil {
		return 0
	}
	return s.writer.Len()
}

func normalizeLabel(label string) string {
	return strings.ToLower(label)
}

// EmitS emits an opcode with no immediate payload.
func (s ScriptBuilder) EmitS(opcode vm.Opcode) ScriptBuilder {
	s.writer.WriteOp(byte(opcode))
	return s
}

// EmitM emits an opcode followed by an immediate byte payload.
func (s ScriptBuilder) EmitM(opcode vm.Opcode, bytes []byte) ScriptBuilder {
	s.writer.WriteOp(byte(opcode))

	if len(bytes) > 0 {
		s.writer.WriteBytes(bytes)
	}

	return s
}

// EmitThrow emits a THROW instruction using reg.
func (s ScriptBuilder) EmitThrow(reg byte) ScriptBuilder {
	s.EmitS(vm.THROW)
	s.writer.WriteB(reg)
	return s
}

// EmitPush emits a PUSH instruction for reg.
func (s ScriptBuilder) EmitPush(reg byte) ScriptBuilder {
	s.EmitS(vm.PUSH)
	s.writer.WriteB(reg)
	return s
}

// EmitPop emits a POP instruction for reg.
func (s ScriptBuilder) EmitPop(reg byte) ScriptBuilder {
	s.EmitS(vm.POP)
	s.writer.WriteB(reg)
	return s
}

// EmitExtCall emits an EXTCALL instruction for method.
func (s ScriptBuilder) EmitExtCall(method string, reg byte) ScriptBuilder {
	bytes := []byte(method)
	s.EmitLoad(reg, bytes, vm.String)
	s.EmitS(vm.EXTCALL)
	s.writer.WriteB(reg)
	return s
}

// EmitLoadBool loads a bool value into reg.
func (s ScriptBuilder) EmitLoadBool(reg byte, toLoad bool) ScriptBuilder {

	var bytes []byte
	if toLoad {
		bytes = []byte{1}
	} else {
		bytes = []byte{0}
	}

	s.EmitLoad(reg, bytes, vm.Bool)
	return s
}

// EmitLoadTime loads a timestamp value into reg.
func (s ScriptBuilder) EmitLoadTime(reg byte, toLoad time.Time) ScriptBuilder {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(toLoad.Unix()))
	s.EmitLoad(reg, bytes, vm.Timestamp)
	return s
}

// EmitLoadInt loads an int value as a VM Number into reg.
func (s ScriptBuilder) EmitLoadInt(reg byte, toLoad int) ScriptBuilder {
	return s.EmitLoadNumberAsBinary(reg, big.NewInt(int64(toLoad)))
}

// EmitLoadNumberAsBinary loads an integer as a VM Number into reg.
func (s ScriptBuilder) EmitLoadNumberAsBinary(reg byte, toLoad *big.Int) ScriptBuilder {
	b := util.BigIntToCsharpByteArray(toLoad)
	s.EmitLoad(reg, b, vm.Number)
	return s
}

// EmitLoadString loads a string value into reg.
func (s ScriptBuilder) EmitLoadString(reg byte, toLoad string) ScriptBuilder {
	bytes := []byte(toLoad)
	s.EmitLoad(reg, bytes, vm.String)
	return s
}

// EmitLoad emits a LOAD instruction for raw VM data.
func (s ScriptBuilder) EmitLoad(reg byte, bytes []byte, _type vm.VMType) ScriptBuilder {
	if len(bytes) > 0xFFFF {
		return s.failf("tried to load too much data: %d bytes", len(bytes))
	}

	s.EmitS(vm.LOAD)
	s.writer.WriteB(reg)
	s.writer.WriteB(byte(_type))

	s.writer.WriteVarUint(uint64(len(bytes)))
	s.writer.WriteBytes(bytes)
	return s
}

// EmitMove emits a MOVE instruction from srcReg to dstReg.
func (s ScriptBuilder) EmitMove(srcReg byte, dstReg byte) ScriptBuilder {
	s.EmitS(vm.MOVE)
	s.writer.WriteB(srcReg)
	s.writer.WriteB(dstReg)
	return s
}

// EmitCopy emits a COPY instruction from srcReg to dstReg.
func (s ScriptBuilder) EmitCopy(srcReg byte, dstReg byte) ScriptBuilder {
	s.EmitS(vm.COPY)
	s.writer.WriteB(srcReg)
	s.writer.WriteB(dstReg)
	return s
}

// EmitLabel records a jump target label at the current script offset.
func (s ScriptBuilder) EmitLabel(label string) ScriptBuilder {
	s.EmitS(vm.NOP)
	s.labelLocations[normalizeLabel(label)] = s.writer.Len()
	return s
}

// EmitJump emits an unconditional or conditional jump to label.
func (s ScriptBuilder) EmitJump(opcode vm.Opcode, label string, reg byte) ScriptBuilder {
	switch opcode {
	case vm.JMP, vm.JMPIF, vm.JMPNOT:
		s.EmitS(opcode)
	default:
		return s.failf("invalid jump opcode: %v", opcode)
	}

	if opcode != vm.JMP {
		s.writer.WriteB(reg)
	}

	ofs := s.writer.Len()
	s.writer.WriteU16LE(0)
	s.jumpLocations[ofs] = label

	return s
}

// EmitCall emits a CALL instruction to label using regCnt registers.
func (s ScriptBuilder) EmitCall(label string, regCnt byte) ScriptBuilder {
	if regCnt < 1 || regCnt > maxRegisterCount {
		return s.failf("invalid number of registers: %d", regCnt)
	}

	ofs := s.writer.Len() + 2
	s.EmitS(vm.CALL)
	s.writer.WriteB(regCnt)
	s.writer.WriteU16LE(0)

	s.jumpLocations[ofs] = label

	return s
}

// EmitConditionalJump emits JMPIF or JMPNOT using srcReg and label.
func (s ScriptBuilder) EmitConditionalJump(opcode vm.Opcode, srcReg byte, label string) ScriptBuilder {
	if opcode != vm.JMPIF && opcode != vm.JMPNOT {
		return s.failf("opcode is not a conditional jump: %v", opcode)
	}

	ofs := s.writer.Len() + 2
	s.EmitS(opcode)
	s.writer.WriteB(srcReg)
	s.writer.WriteU16LE(0)
	s.jumpLocations[ofs] = label
	return s
}

// EmitVarBytes emits value as a variable-length unsigned integer.
func (s ScriptBuilder) EmitVarBytes(value int) ScriptBuilder {
	s.writer.WriteVarUint(uint64(value))
	return s
}

// EmitRaw appends raw bytes to the script.
func (s ScriptBuilder) EmitRaw(value []byte) ScriptBuilder {
	s.writer.WriteBytes(value)
	return s
}

func (s ScriptBuilder) loadIntoReg(dstReg byte, arg interface{}) {
	switch e := arg.(type) {
	case string:
		s.EmitLoadString(dstReg, e)
	case bool:
		s.EmitLoadBool(dstReg, e)
	case []byte:
		s.EmitLoad(dstReg, e, vm.Bytes)
	case int:
		s.EmitLoadInt(dstReg, e)
	case int8:
		s.EmitLoadNumberAsBinary(dstReg, big.NewInt(int64(e)))
	case int16:
		s.EmitLoadNumberAsBinary(dstReg, big.NewInt(int64(e)))
	case int32:
		s.EmitLoadNumberAsBinary(dstReg, big.NewInt(int64(e)))
	case int64:
		s.EmitLoadNumberAsBinary(dstReg, big.NewInt(e))
	case uint:
		s.EmitLoadNumberAsBinary(dstReg, new(big.Int).SetUint64(uint64(e)))
	case uint8:
		s.EmitLoadNumberAsBinary(dstReg, new(big.Int).SetUint64(uint64(e)))
	case uint16:
		s.EmitLoadNumberAsBinary(dstReg, new(big.Int).SetUint64(uint64(e)))
	case uint32:
		s.EmitLoadNumberAsBinary(dstReg, new(big.Int).SetUint64(uint64(e)))
	case uint64:
		s.EmitLoadNumberAsBinary(dstReg, new(big.Int).SetUint64(e))
	case *big.Int:
		s.EmitLoadNumberAsBinary(dstReg, e)
	case big.Int:
		s.EmitLoadNumberAsBinary(dstReg, &e)
	case time.Time:
		s.EmitLoadTime(dstReg, e)
	case crypto.Address:
		s.EmitLoad(dstReg, e.BytesPrefixed(), vm.Bytes)
	case *crypto.Address:
		if e == nil {
			s.failf("unsupported nil address pointer")
			return
		}
		s.EmitLoad(dstReg, e.BytesPrefixed(), vm.Bytes)
	default:
		if arg == nil {
			s.failf("unsupported nil argument")
			return
		}
		if s.loadArrayIntoReg(dstReg, arg) {
			return
		}
		s.failf("unsupported type %T", e)
		return
	}
}

func (s ScriptBuilder) loadArrayIntoReg(dstReg byte, arg interface{}) bool {
	value := reflect.ValueOf(arg)
	if value.Kind() != reflect.Array && value.Kind() != reflect.Slice {
		return false
	}

	if dstReg > maxRegisterCount-3 {
		s.failf("array load needs three registers starting at %d", dstReg)
		return true
	}

	s.EmitM(vm.CAST, []byte{dstReg, dstReg, byte(vm.None)})
	for i := 0; i < value.Len(); i++ {
		valueReg := dstReg + 1
		keyReg := dstReg + 2
		s.loadIntoReg(valueReg, value.Index(i).Interface())
		s.loadIntoReg(keyReg, i)
		s.EmitM(vm.PUT, []byte{valueReg, dstReg, keyReg})
	}
	return true
}

func (s ScriptBuilder) failf(format string, args ...interface{}) ScriptBuilder {
	return s.withError(fmt.Errorf(format, args...))
}

func (s ScriptBuilder) withError(err error) ScriptBuilder {
	if s.writer != nil {
		if s.writer.Err == nil {
			s.writer.Err = err
		}
	}
	return s
}

func (s ScriptBuilder) insertMethodArgs(args []interface{}) {
	var tempReg byte = 0

	for i := len(args) - 1; i >= 0; i-- {
		arg := args[i]
		s.loadIntoReg(tempReg, arg)
		s.EmitPush(tempReg)
	}
}

// CallInterop emits an interop call with arguments pushed in VM call order.
func (s ScriptBuilder) CallInterop(method string, args ...interface{}) ScriptBuilder {
	s.insertMethodArgs(args)

	var dstReg byte = 0
	s.EmitLoadString(dstReg, method)

	s.EmitM(vm.EXTCALL, []byte{dstReg})

	return s
}

// CallContract emits a contract call with arguments pushed in VM call order.
func (s ScriptBuilder) CallContract(contractName, method string, args ...interface{}) ScriptBuilder {
	s.insertMethodArgs(args)

	var tmpReg byte = 0
	s.EmitLoadString(tmpReg, method)
	s.EmitPush(tmpReg)

	var srcReg byte = 0
	var dstReg byte = 1
	s.EmitLoadString(srcReg, contractName)
	s.EmitM(vm.CTX, []byte{srcReg, dstReg})

	s.EmitM(vm.SWITCH, []byte{dstReg})
	return s
}
