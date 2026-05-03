package vm

import (
	"math/big"
	"testing"
)

func TestVMObjectConversions(t *testing.T) {
	if got := new(VMObject).SetValue([]byte("42"), String).AsNumber().Int64(); got != 42 {
		t.Fatalf("string number conversion mismatch: %d", got)
	}
	if got := new(VMObject).SetValue([]byte{1}, Bool).AsString(); got != "true" {
		t.Fatalf("bool string conversion mismatch: %s", got)
	}

	number := new(VMObject).SetValue([]byte{0x2a}, Number)
	if number.AsNumber().Cmp(big.NewInt(42)) != 0 {
		t.Fatalf("number conversion mismatch: %s", number.AsNumber())
	}
	if number.String() != "[Number] => 42" {
		t.Fatalf("string representation mismatch: %s", number.String())
	}
}

func TestVMObjectCopy(t *testing.T) {
	source := new(VMObject).SetValue([]byte("hello"), String)
	var clone VMObject
	clone.Copy(source)
	if clone.Type != String || clone.AsString() != "hello" {
		t.Fatalf("copy mismatch: %+v", clone)
	}

	clone.Copy(nil)
	if clone.Type != None || clone.Data != nil {
		t.Fatalf("nil copy must reset object")
	}
}
