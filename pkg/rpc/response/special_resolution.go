package response

import "encoding/json"

// SpecialResolutionData is the payload of a SpecialResolution extended event.
type SpecialResolutionData struct {
	ResolutionID uint64 `json:"resolutionId"`
	// Description is absent on most resolutions; the node omits it instead of answering null.
	Description *string                 `json:"description,omitempty"`
	Calls       []SpecialResolutionCall `json:"calls"`
}

func (SpecialResolutionData) isEventData() {}

type specialResolutionDataWire struct {
	ResolutionID json.RawMessage `json:"resolutionId"`
	Description  json.RawMessage `json:"description"`
	Calls        json.RawMessage `json:"calls"`
}

// UnmarshalJSON reads the resolution envelope field by field.
//
// Every field is read leniently and a payload that is not an object at all yields the empty
// resolution rather than an error: by the time this runs the document is already valid JSON, so
// the only failures left are shape mismatches, and those must not take down a whole block answer.
func (s *SpecialResolutionData) UnmarshalJSON(data []byte) error {
	*s = SpecialResolutionData{}

	var wire specialResolutionDataWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil
	}

	s.ResolutionID = jsonUint64(wire.ResolutionID)
	s.Description = jsonOptionalString(wire.Description)
	s.Calls = jsonCalls(wire.Calls)
	return nil
}

// SpecialResolutionCall is one call carried by a special resolution.
//
// Arguments is typed per method: Module and Method decide the shape, and the decoder picks it
// during deserialization. Calls carries the calls of a nested resolution and is absent everywhere
// else.
type SpecialResolutionCall struct {
	ModuleID uint32 `json:"moduleId"`
	Module   string `json:"module"`
	MethodID uint32 `json:"methodId"`
	Method   string `json:"method"`
	// Arguments is nil when the call carries none.
	Arguments SpecialResolutionArguments `json:"arguments,omitempty"`
	Calls     []SpecialResolutionCall    `json:"calls,omitempty"`
}

type specialResolutionCallWire struct {
	ModuleID  json.RawMessage `json:"moduleId"`
	Module    json.RawMessage `json:"module"`
	MethodID  json.RawMessage `json:"methodId"`
	Method    json.RawMessage `json:"method"`
	Arguments json.RawMessage `json:"arguments"`
	Calls     json.RawMessage `json:"calls"`
}

// UnmarshalJSON reads one call and gives its arguments the type that belongs to the called method.
// Field reads are lenient for the same reason as on SpecialResolutionData: a single odd call must
// not fail the transaction it belongs to.
func (c *SpecialResolutionCall) UnmarshalJSON(data []byte) error {
	*c = SpecialResolutionCall{}

	var wire specialResolutionCallWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil
	}

	c.ModuleID = jsonUint32(wire.ModuleID)
	c.Module = jsonString(wire.Module)
	c.MethodID = jsonUint32(wire.MethodID)
	c.Method = jsonString(wire.Method)
	c.Arguments = DecodeSpecialResolutionArguments(c.Module, c.Method, wire.Arguments)
	c.Calls = jsonCalls(wire.Calls)
	return nil
}

// SpecialResolutionArguments is the decoded arguments of one call inside a special resolution,
// typed per module and method.
//
// The set of shapes is closed - only the argument types in this package implement the interface -
// so a type switch over them covers everything this build models:
//
//	switch arguments := call.Arguments.(type) {
//	case ImportContractsArguments:
//		for _, contract := range arguments.Contracts { use(contract.Script, contract.Tables) }
//	case RawArguments:
//		use(arguments.RawArgs)
//	case UnrecognizedArguments:
//		log(arguments.JSON)
//	}
//
// Shapes that repeat across methods share one type on purpose: a query by token id looks the same
// whichever query it is. Every numeric field of every shape is a string, because chain values are
// big integers and JSON numbers lose precision above 2^53.
//
// Decoding is total: a module/method pair this build does not model, and a modeled pair whose
// payload does not match its shape, both arrive as UnrecognizedArguments with the JSON preserved
// verbatim. The C# reference SDK drops those instead; this SDK keeps them so that data answered by
// a node newer than the SDK is never silently lost.
type SpecialResolutionArguments interface {
	isSpecialResolutionArguments()
}

// RawArguments carries the argument buffer of a call the answering node itself could not decode.
type RawArguments struct {
	RawArgs string `json:"rawArgs"`
}

func (RawArguments) isSpecialResolutionArguments() {}

// UnrecognizedArguments carries the verbatim arguments of a module/method pair this build does not
// model, or of a modeled pair whose payload did not match its shape.
type UnrecognizedArguments struct {
	JSON json.RawMessage
}

func (UnrecognizedArguments) isSpecialResolutionArguments() {}

// MarshalJSON writes the preserved arguments back unchanged.
func (u UnrecognizedArguments) MarshalJSON() ([]byte, error) {
	if len(u.JSON) == 0 {
		return []byte("null"), nil
	}
	return u.JSON, nil
}

type specialResolutionArgumentsDecoder func(json.RawMessage) SpecialResolutionArguments

// specialResolutionArgumentDecoders maps module and method to the shape of that call's arguments.
// It mirrors the converter of the C# SDK and the node's SpecialResolutionHelper, which build these
// answers. A pair missing here is not an error: the node answers the raw argument buffer for
// anything it cannot decode, and an unmodeled pair keeps its JSON.
var specialResolutionArgumentDecoders = map[string]map[string]specialResolutionArgumentsDecoder{
	"governance": {
		"SetGasConfig":      typedArguments[GasConfigArguments],
		"SetChainConfig":    typedArguments[ChainConfigArguments],
		"SpecialResolution": typedArguments[NestedResolutionArguments],
		"SetMetadata":       typedArguments[MetadataArguments],
		"SetNodeConfig":     typedArguments[NodeConfigArguments],
		"RegisterName":      typedArguments[RegisterNameArguments],
		"LookupName":        typedArguments[AddressArguments],
		"LookupAddress":     typedArguments[NameArguments],
	},
	"phantasma_vm": {
		"ExecuteScript":         typedArguments[ExecuteScriptArguments],
		"RegisterTokenContract": typedArguments[RegisterTokenContractArguments],
		"DeployContract":        typedArguments[DeployContractArguments],
		"IsContractDeployed":    typedArguments[NameArguments],
		"SetConfig":             typedArguments[PhantasmaVMConfigArguments],
		"ImportContracts":       typedArguments[ImportContractsArguments],
		"RepairSeries":          typedArguments[RepairSeriesArguments],
		"RepairToken":           typedArguments[RepairTokenArguments],
	},
	"token": {
		"TransferFungible":          typedArguments[TransferFungibleArguments],
		"TransferNonFungible":       typedArguments[TransferNonFungibleArguments],
		"CreateToken":               typedArguments[CreateTokenArguments],
		"MintFungible":              typedArguments[MintFungibleArguments],
		"BurnFungible":              typedArguments[BurnFungibleArguments],
		"GetBalance":                typedArguments[BalanceArguments],
		"CreateTokenSeries":         typedArguments[TokenSeriesArguments],
		"DeleteTokenSeries":         typedArguments[TokenSeriesReferenceArguments],
		"MintNonFungible":           typedArguments[MintNonFungibleArguments],
		"BurnNonFungible":           typedArguments[BurnNonFungibleArguments],
		"GetNonFungibleInfo":        typedArguments[NonFungibleInfoArguments],
		"GetNonFungibleInfoByRomId": typedArguments[NonFungibleInfoByRomIDArguments],
		"GetSeriesInfo":             typedArguments[TokenSeriesReferenceArguments],
		"GetSeriesInfoByMetaId":     typedArguments[SeriesInfoByMetaIDArguments],
		"GetTokenInfo":              typedArguments[TokenReferenceArguments],
		"GetTokenInfoBySymbol":      typedArguments[SymbolArguments],
		"GetTokenSupply":            typedArguments[TokenReferenceArguments],
		"GetSeriesSupply":           typedArguments[TokenSeriesReferenceArguments],
		"GetTokenIdBySymbol":        typedArguments[SymbolArguments],
		"GetBalances":               typedArguments[AddressArguments],
		"CreateMintedTokenSeries":   typedArguments[CreateMintedTokenSeriesArguments],
		"ApplyInflation":            typedArguments[TokenReferenceArguments],
		"UpdateTokenMetadata":       typedArguments[UpdateTokenMetadataArguments],
		"GetNextTokenInflation":     typedArguments[TokenReferenceArguments],
		"SetTokensConfig":           typedArguments[TokensConfigArguments],
		"UpdateSeriesMetadata":      typedArguments[UpdateSeriesMetadataArguments],
		"MintPhantasmaNonFungible":  typedArguments[MintPhantasmaNonFungibleArguments],
	},
}

// DecodeSpecialResolutionArguments types the arguments of one call from its module and method.
//
// Exported so that a consumer decoding calls through its own transport gets the same dispatch.
// Absent arguments yield nil; everything else follows the totality rule documented on
// SpecialResolutionArguments.
func DecodeSpecialResolutionArguments(module, method string, data json.RawMessage) SpecialResolutionArguments {
	if len(data) == 0 {
		return nil
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return UnrecognizedArguments{JSON: data}
	}
	// An explicit null decodes into a nil map without error, and means the same as an absent field.
	if object == nil {
		return nil
	}

	// The undecoded case is recognised by its content, not by the method name: a method this build
	// knows can still arrive as a raw dump from an older node, and reading that as the typed shape
	// would silently produce an object with every field empty.
	if _, ok := object["rawArgs"]; ok {
		return typedArguments[RawArguments](data)
	}

	if methods, ok := specialResolutionArgumentDecoders[module]; ok {
		if decode, ok := methods[method]; ok {
			return decode(data)
		}
	}

	return UnrecognizedArguments{JSON: data}
}

// typedArguments decodes one modeled shape, keeping the raw arguments when they do not match it.
func typedArguments[T SpecialResolutionArguments](data json.RawMessage) SpecialResolutionArguments {
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return UnrecognizedArguments{JSON: data}
	}
	return value
}

// jsonCalls reads a nested call list; anything that is not an array yields no calls.
func jsonCalls(raw json.RawMessage) []SpecialResolutionCall {
	if len(raw) == 0 {
		return nil
	}
	var calls []SpecialResolutionCall
	if err := json.Unmarshal(raw, &calls); err != nil {
		return nil
	}
	return calls
}

// jsonString reads a string field; a missing field or a non-string value yields the empty string.
func jsonString(raw json.RawMessage) string {
	if value := jsonOptionalString(raw); value != nil {
		return *value
	}
	return ""
}

// jsonOptionalString reads a string field, keeping absent and non-string apart from an empty one.
func jsonOptionalString(raw json.RawMessage) *string {
	if len(raw) == 0 {
		return nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	return &value
}

// jsonUint64 reads a numeric id field; a missing or non-numeric value yields 0, as in the C# SDK.
func jsonUint64(raw json.RawMessage) uint64 {
	if len(raw) == 0 {
		return 0
	}
	var value uint64
	if err := json.Unmarshal(raw, &value); err != nil {
		return 0
	}
	return value
}

// jsonUint32 reads a numeric id field; a missing or non-numeric value yields 0, as in the C# SDK.
func jsonUint32(raw json.RawMessage) uint32 {
	if len(raw) == 0 {
		return 0
	}
	var value uint32
	if err := json.Unmarshal(raw, &value); err != nil {
		return 0
	}
	return value
}
