package response

import (
	"encoding/json"
	"reflect"
	"testing"
)

// callWire wraps one arguments payload into the call that carries it, which is the only way the
// decoder learns which shape to expect.
func callWire(module, method, arguments string) string {
	return `{"moduleId":0,"module":"` + module + `","methodId":0,"method":"` + method + `","arguments":` + arguments + `}`
}

func decodeCall(t *testing.T, wire string) SpecialResolutionCall {
	t.Helper()

	var call SpecialResolutionCall
	if err := json.Unmarshal([]byte(wire), &call); err != nil {
		t.Fatalf("call decode failed: %v", err)
	}
	return call
}

// argumentDispatchCases pins the shape every module/method pair decodes into. The pairs and their
// shapes mirror the C# reference converter and the node's SpecialResolutionHelper, which build
// these answers; a payload here carries real fields of its shape so that a wrong or missing JSON
// tag shows up as an all-zero struct rather than passing silently.
var argumentDispatchCases = []struct {
	module    string
	method    string
	arguments string
	want      SpecialResolutionArguments
}{
	{"governance", "SetGasConfig", `{"version":"1","feeMultiplier":"16"}`, GasConfigArguments{}},
	{"governance", "SetChainConfig", `{"version":"0","expiryWindow":"600"}`, ChainConfigArguments{}},
	{"governance", "SpecialResolution", `{"resolutionId":"31"}`, NestedResolutionArguments{}},
	{"governance", "SetMetadata", `{"metadata":{"name":"Phantasma"}}`, MetadataArguments{}},
	{"governance", "SetNodeConfig", `{"nodes":[{"id":"1","type":"Validator"}]}`, NodeConfigArguments{}},
	{"governance", "RegisterName", `{"address":"P2K6h","name":"alex"}`, RegisterNameArguments{}},
	{"governance", "LookupName", `{"address":"P2K6h"}`, AddressArguments{}},
	{"governance", "LookupAddress", `{"name":"alex"}`, NameArguments{}},

	{"phantasma_vm", "ExecuteScript", `{"maxGas":"10000","gasFrom":"P2K6h","script":"0B"}`, ExecuteScriptArguments{}},
	{"phantasma_vm", "RegisterTokenContract", `{"tokenId":"4","symbol":"CROWN","script":"0B","abi":"09"}`, RegisterTokenContractArguments{}},
	{"phantasma_vm", "DeployContract", `{"from":"P2K6h","contractName":"mail","script":"0B","abi":"09"}`, DeployContractArguments{}},
	{"phantasma_vm", "IsContractDeployed", `{"name":"mail"}`, NameArguments{}},
	{"phantasma_vm", "SetConfig", `{"featureLevel":"3","gasNexus":"100"}`, PhantasmaVMConfigArguments{}},
	{"phantasma_vm", "ImportContracts", `{"contractsCount":"70","contracts":[]}`, ImportContractsArguments{}},
	{"phantasma_vm", "RepairSeries", `{"supplementsCount":"3649","repairsCount":"8370"}`, RepairSeriesArguments{}},
	{"phantasma_vm", "RepairToken", `{"repairsCount":"2","repairs":[]}`, RepairTokenArguments{}},

	{"token", "TransferFungible", `{"from":"S3dPn","to":"S3dPn","amount":"1","token":"KCAL","tokenId":"1"}`, TransferFungibleArguments{}},
	{"token", "TransferNonFungible", `{"from":"S3dPn","to":"S3dPn","instanceIds":["7"],"token":"CROWN","tokenId":"4"}`, TransferNonFungibleArguments{}},
	{"token", "CreateToken", `{"symbol":"SOUL","owner":"P2K6h","maxSupply":"0","decimals":"8","flags":"199"}`, CreateTokenArguments{}},
	{"token", "MintFungible", `{"to":"S3dPn","amount":"5","token":"KCAL","tokenId":"1"}`, MintFungibleArguments{}},
	{"token", "BurnFungible", `{"from":"S3dPn","amount":"5","token":"KCAL","tokenId":"1"}`, BurnFungibleArguments{}},
	{"token", "GetBalance", `{"address":"P2K6h","token":"KCAL","tokenId":"1"}`, BalanceArguments{}},
	{"token", "CreateTokenSeries", `{"owner":"P2K6h","maxMint":"100","maxSupply":"1000","token":"CROWN","tokenId":"4"}`, TokenSeriesArguments{}},
	{"token", "DeleteTokenSeries", `{"seriesId":"1","token":"CROWN","tokenId":"4"}`, TokenSeriesReferenceArguments{}},
	{"token", "MintNonFungible", `{"owner":"P2K6h","tokens":[{"seriesId":"1","rom":"CA","ram":"FE"}],"token":"CROWN","tokenId":"4"}`, MintNonFungibleArguments{}},
	{"token", "BurnNonFungible", `{"address":"P2K6h","instanceIds":["7"],"token":"CROWN","tokenId":"4"}`, BurnNonFungibleArguments{}},
	{"token", "GetNonFungibleInfo", `{"instanceId":"7","getSchemas":"1","token":"CROWN","tokenId":"4"}`, NonFungibleInfoArguments{}},
	{"token", "GetNonFungibleInfoByRomId", `{"romId":"CAFE","getSchemas":"1","token":"CROWN","tokenId":"4"}`, NonFungibleInfoByRomIDArguments{}},
	{"token", "GetSeriesInfo", `{"seriesId":"1","token":"CROWN","tokenId":"4"}`, TokenSeriesReferenceArguments{}},
	{"token", "GetSeriesInfoByMetaId", `{"romId":"CAFE","token":"CROWN","tokenId":"4"}`, SeriesInfoByMetaIDArguments{}},
	{"token", "GetTokenInfo", `{"token":"CROWN","tokenId":"4"}`, TokenReferenceArguments{}},
	{"token", "GetTokenInfoBySymbol", `{"symbol":"CROWN"}`, SymbolArguments{}},
	{"token", "GetTokenSupply", `{"token":"CROWN","tokenId":"4"}`, TokenReferenceArguments{}},
	{"token", "GetSeriesSupply", `{"seriesId":"1","token":"CROWN","tokenId":"4"}`, TokenSeriesReferenceArguments{}},
	{"token", "GetTokenIdBySymbol", `{"symbol":"CROWN"}`, SymbolArguments{}},
	{"token", "GetBalances", `{"address":"P2K6h"}`, AddressArguments{}},
	{"token", "CreateMintedTokenSeries", `{"recipient":"P2K6h","roms":["CA"],"rams":["FE"],"owner":"P2K6h","maxMint":"1","maxSupply":"1","token":"CROWN","tokenId":"4"}`, CreateMintedTokenSeriesArguments{}},
	{"token", "ApplyInflation", `{"token":"SOUL","tokenId":"2"}`, TokenReferenceArguments{}},
	{"token", "UpdateTokenMetadata", `{"metadata":{"name":"Crown"},"token":"CROWN","tokenId":"4"}`, UpdateTokenMetadataArguments{}},
	{"token", "GetNextTokenInflation", `{"token":"SOUL","tokenId":"2"}`, TokenReferenceArguments{}},
	{"token", "SetTokensConfig", `{"flags":"3","flagsNames":["Transferable"]}`, TokensConfigArguments{}},
	{"token", "UpdateSeriesMetadata", `{"seriesId":"1","metadata":"CAFE","token":"CROWN","tokenId":"4"}`, UpdateSeriesMetadataArguments{}},
	{"token", "MintPhantasmaNonFungible", `{"owner":"P2K6h","tokens":[{"phantasmaSeriesId":"6472","rom":"CA","ram":"FE"}],"token":"CROWN","tokenId":"4"}`, MintPhantasmaNonFungibleArguments{}},
}

func TestArgumentsDispatchCoversEveryModuleMethodPair(t *testing.T) {
	for _, testCase := range argumentDispatchCases {
		call := decodeCall(t, callWire(testCase.module, testCase.method, testCase.arguments))

		want := reflect.TypeOf(testCase.want)
		got := reflect.TypeOf(call.Arguments)
		if got != want {
			t.Fatalf("%s.%s must decode to %s, got %s", testCase.module, testCase.method, want, got)
		}
		if reflect.ValueOf(call.Arguments).IsZero() {
			t.Fatalf("%s.%s decoded to an all-zero %s: the payload did not reach any field",
				testCase.module, testCase.method, want)
		}
	}
}

func TestEveryDecodedMethodHasADispatchCase(t *testing.T) {
	// Guards the table above against drift: a method added to the decoder map without a case here
	// would otherwise ship untested.
	covered := make(map[string]map[string]bool, len(specialResolutionArgumentDecoders))
	for _, testCase := range argumentDispatchCases {
		if covered[testCase.module] == nil {
			covered[testCase.module] = map[string]bool{}
		}
		covered[testCase.module][testCase.method] = true
	}

	pairs := 0
	for module, methods := range specialResolutionArgumentDecoders {
		for method := range methods {
			pairs++
			if !covered[module][method] {
				t.Fatalf("%s.%s is decoded but has no dispatch case", module, method)
			}
		}
	}
	if pairs != len(argumentDispatchCases) {
		t.Fatalf("decoder covers %d pairs, the dispatch table lists %d", pairs, len(argumentDispatchCases))
	}
}

func TestRawArgumentsWinOverAKnownMethod(t *testing.T) {
	// An older node can answer a raw dump for a method this build models. Dispatching on the
	// method name first would decode that dump into a typed shape with every field empty, which a
	// consumer cannot tell apart from a genuinely empty call.
	call := decodeCall(t, callWire("token", "TransferFungible", `{"rawArgs":"0104040B534F554C"}`))

	raw, ok := call.Arguments.(RawArguments)
	if !ok {
		t.Fatalf("a rawArgs payload must decode as raw, got %T", call.Arguments)
	}
	if raw.RawArgs != "0104040B534F554C" {
		t.Fatalf("raw argument buffer lost: %+v", raw)
	}
}

func TestUnmodeledMethodKeepsItsArgumentsVerbatim(t *testing.T) {
	const payload = `{"brandNewField":"7","nested":{"deep":["1"]}}`
	call := decodeCall(t, callWire("token", "BrandNewMethod", payload))

	unrecognized, ok := call.Arguments.(UnrecognizedArguments)
	if !ok {
		t.Fatalf("an unmodeled method must keep its arguments, got %T", call.Arguments)
	}
	assertSameJSON(t, []byte(payload), unrecognized.JSON)
}

func TestMismatchedKnownShapeKeepsItsArgumentsVerbatim(t *testing.T) {
	// The pair is modeled but the payload is not the modeled shape - a node whose fields drifted.
	// Keeping the JSON makes the drift visible instead of silently reporting empty fields.
	const payload = `{"supplementsCount":["3649"]}`
	call := decodeCall(t, callWire("phantasma_vm", "RepairSeries", payload))

	unrecognized, ok := call.Arguments.(UnrecognizedArguments)
	if !ok {
		t.Fatalf("a mismatched payload must keep its arguments, got %T", call.Arguments)
	}
	assertSameJSON(t, []byte(payload), unrecognized.JSON)
}

func TestNestedResolutionCallsDecodeRecursively(t *testing.T) {
	// A resolution can carry another resolution: the outer call holds the nested id in its
	// arguments and the nested calls in its calls, which dispatch exactly like top-level ones.
	call := decodeCall(t, `{
		"moduleId": 0,
		"module": "governance",
		"methodId": 2,
		"method": "SpecialResolution",
		"arguments": {"resolutionId": "31"},
		"calls": [{
			"moduleId": 1,
			"module": "token",
			"methodId": 0,
			"method": "TransferFungible",
			"arguments": {"from":"S3dPn","to":"S3dPn","amount":"5","token":"KCAL","tokenId":"1"}
		}]
	}`)

	nested, ok := call.Arguments.(NestedResolutionArguments)
	if !ok {
		t.Fatalf("nested resolution arguments should decode to their typed shape, got %T", call.Arguments)
	}
	// The envelope reports resolutionId as a number, this argument shape as a string; both follow
	// the wire rather than being normalized to one of the two.
	if nested.ResolutionID != "31" {
		t.Fatalf("nested resolution id lost: %+v", nested)
	}
	if len(call.Calls) != 1 {
		t.Fatalf("expected 1 nested call, got %d", len(call.Calls))
	}
	transfer, ok := call.Calls[0].Arguments.(TransferFungibleArguments)
	if !ok {
		t.Fatalf("nested call arguments should dispatch like top-level ones, got %T", call.Calls[0].Arguments)
	}
	if transfer.Amount != "5" {
		t.Fatalf("nested transfer payload lost: %+v", transfer)
	}
}

func TestCreateTokenArgumentsCarryVMMetadata(t *testing.T) {
	// token.CreateToken metadata values are VM values, not plain strings: the interest array of
	// getToken("SOUL", true) is the shape that motivated VMValue (devnet, 2026-08-01).
	call := decodeCall(t, callWire("token", "CreateToken", `{
		"symbol": "SOUL",
		"owner": "P2KFNXEbt65rQiWqogAzqkVGMqFirPmqPw8mQyxvRKsrXV8",
		"maxSupply": "0",
		"decimals": "8",
		"flags": "199",
		"metadata": {"_ia": [{"mul": "25", "div": "10000"}], "name": "Phantasma Stake"}
	}`))

	created, ok := call.Arguments.(CreateTokenArguments)
	if !ok {
		t.Fatalf("CreateToken arguments should decode to their typed shape, got %T", call.Arguments)
	}
	if created.Decimals != "8" {
		t.Fatalf("decimals must stay the wire string: %+v", created)
	}
	if name, ok := created.Metadata["name"].AsText(); !ok || name != "Phantasma Stake" {
		t.Fatalf("scalar metadata lost: %+v", created.Metadata)
	}
	interest, ok := created.Metadata["_ia"].AsItems()
	if !ok || len(interest) != 1 {
		t.Fatalf("_ia must stay an array: %+v", created.Metadata["_ia"])
	}
	mul, ok := interest[0].Field("mul")
	if !ok {
		t.Fatalf("interest entry must carry mul: %+v", interest[0])
	}
	if text, _ := mul.AsText(); text != "25" {
		t.Fatalf("nested scalar lost: %+v", mul)
	}
	if created.TokenSchemas != nil {
		t.Fatalf("a fungible token carries no schema blob: %+v", created.TokenSchemas)
	}
}

func TestGasConfigV2TailIsOptional(t *testing.T) {
	const version0 = `{
		"version": "0",
		"maxNameLength": "255",
		"maxTokenSymbolLength": "10",
		"feeShift": "10",
		"maxStructureSize": "65535",
		"feeMultiplier": "16",
		"gasTokenId": "1",
		"dataTokenId": "0",
		"minimumGasOffer": "10000",
		"dataEscrowPerRow": "1000000",
		"gasFeeTransfer": "1000",
		"gasFeeQuery": "100",
		"gasFeeCreateTokenBase": "100000000",
		"gasFeeCreateTokenSymbol": "10000000",
		"gasFeeCreateTokenSeries": "1000000",
		"gasFeePerByte": "10",
		"gasFeeRegisterName": "100000",
		"gasBurnRatioMul": "1",
		"gasBurnRatioShift": "1"
	}`

	// A version 0 config has no gas-model-v2 tail on the wire; the optional fields must stay
	// absent and re-serializing the call must not invent null keys for them.
	call := decodeCall(t, callWire("governance", "SetGasConfig", version0))
	config, ok := call.Arguments.(GasConfigArguments)
	if !ok {
		t.Fatalf("SetGasConfig arguments should decode to their typed shape, got %T", call.Arguments)
	}
	if config.FeeMultiplier != "16" || config.MinimumGasBill != nil {
		t.Fatalf("version 0 config must carry no v2 tail: %+v", config)
	}

	encoded, err := json.Marshal(call)
	if err != nil {
		t.Fatalf("call encode failed: %v", err)
	}
	assertSameJSON(t, []byte(callWire("governance", "SetGasConfig", version0)), encoded)

	// A version 1 config carries the tail; the same shape must pick it up.
	withTail := `{"version":"1","feeMultiplier":"16","minimumGasBill":"21000","gasProducerRatioMul":"45"}`
	call = decodeCall(t, callWire("governance", "SetGasConfig", withTail))
	config, ok = call.Arguments.(GasConfigArguments)
	if !ok {
		t.Fatalf("SetGasConfig arguments should decode to their typed shape, got %T", call.Arguments)
	}
	if config.MinimumGasBill == nil || *config.MinimumGasBill != "21000" {
		t.Fatalf("v2 minimum gas bill lost: %+v", config)
	}
	if config.GasProducerRatioMul == nil || *config.GasProducerRatioMul != "45" {
		t.Fatalf("v2 producer ratio lost: %+v", config)
	}
}

func TestAbsentAndNullArgumentsCarryNoShape(t *testing.T) {
	// The node omits absent arguments (its serializer drops nulls), but both spellings must land
	// on no arguments at all rather than on a fabricated empty shape.
	call := decodeCall(t, `{"moduleId":1,"module":"token","methodId":0,"method":"TransferFungible"}`)
	if call.Arguments != nil || call.Calls != nil {
		t.Fatalf("a call without arguments must carry none: %+v", call)
	}

	call = decodeCall(t, `{"moduleId":1,"module":"token","methodId":0,"method":"TransferFungible","arguments":null}`)
	if call.Arguments != nil {
		t.Fatalf("explicit null arguments must carry no shape: %+v", call.Arguments)
	}
}

func TestCallToleratesOddEnvelopeFields(t *testing.T) {
	// One malformed call must not fail the transaction it belongs to: ids that are not numbers and
	// names that are not strings fall back to their zero values, and the arguments still dispatch
	// on whatever module and method could be read.
	call := decodeCall(t, `{"moduleId":"two","module":["token"],"methodId":0,"method":"TransferFungible","arguments":{"amount":"5"},"calls":{}}`)

	if call.ModuleID != 0 || call.Module != "" {
		t.Fatalf("unreadable identity fields must fall back to zero: %+v", call)
	}
	if call.Calls != nil {
		t.Fatalf("a non-array calls field must yield no nested calls: %+v", call.Calls)
	}
	// Module was unreadable, so the pair is unknown and the arguments keep their JSON.
	if _, ok := call.Arguments.(UnrecognizedArguments); !ok {
		t.Fatalf("arguments of an unknown pair must stay verbatim, got %T", call.Arguments)
	}
}
