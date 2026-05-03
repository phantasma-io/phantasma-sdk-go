package contract

import (
	"testing"

	sdkio "github.com/phantasma-io/phantasma-go/pkg/io"
	"github.com/phantasma-io/phantasma-go/pkg/vm"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func TestContractInterfaceRoundTrip(t *testing.T) {
	iface := &ContractInterface{
		Methods: orderedmap.New[string, ContractMethod](),
		Events:  []ContractEvent{{Value: 1, Name: "Evt", ReturnType: vm.String, Description: []byte("desc")}},
	}
	iface.Methods.Set("main", ContractMethod{
		Name:       "main",
		ReturnType: vm.Number,
		Offset:     7,
		Parameters: []ContractParameter{{Name: "amount", Type: vm.Number}},
	})

	decoded := sdkio.Deserialize[*ContractInterface](sdkio.Serialize[*ContractInterface](iface))
	method, ok := decoded.Methods.Get("main")
	if !ok || method.ReturnType != vm.Number || method.Offset != 7 || len(decoded.Events) != 1 {
		t.Fatalf("contract interface round trip mismatch: %+v", decoded)
	}
}
