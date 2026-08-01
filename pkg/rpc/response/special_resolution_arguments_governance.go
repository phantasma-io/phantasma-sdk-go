package response

// Arguments of the governance module calls a special resolution can carry. Field-for-field mirrors
// of the call payloads; every numeric field is a string for the reason stated on
// SpecialResolutionArguments.

// GasConfigArguments are the arguments of governance.SetGasConfig.
type GasConfigArguments struct {
	Version                 string `json:"version"`
	MaxNameLength           string `json:"maxNameLength"`
	MaxTokenSymbolLength    string `json:"maxTokenSymbolLength"`
	FeeShift                string `json:"feeShift"`
	MaxStructureSize        string `json:"maxStructureSize"`
	FeeMultiplier           string `json:"feeMultiplier"`
	GasTokenID              string `json:"gasTokenId"`
	DataTokenID             string `json:"dataTokenId"`
	MinimumGasOffer         string `json:"minimumGasOffer"`
	DataEscrowPerRow        string `json:"dataEscrowPerRow"`
	GasFeeTransfer          string `json:"gasFeeTransfer"`
	GasFeeQuery             string `json:"gasFeeQuery"`
	GasFeeCreateTokenBase   string `json:"gasFeeCreateTokenBase"`
	GasFeeCreateTokenSymbol string `json:"gasFeeCreateTokenSymbol"`
	GasFeeCreateTokenSeries string `json:"gasFeeCreateTokenSeries"`
	GasFeePerByte           string `json:"gasFeePerByte"`
	GasFeeRegisterName      string `json:"gasFeeRegisterName"`
	GasBurnRatioMul         string `json:"gasBurnRatioMul"`
	GasBurnRatioShift       string `json:"gasBurnRatioShift"`

	// Gas-model-v2 tail: present only when the packaged config declares version >= 1, absent
	// otherwise, which is why these are pointers rather than empty strings.
	MinimumGasBill             *string `json:"minimumGasBill,omitempty"`
	GasProducerRatioMul        *string `json:"gasProducerRatioMul,omitempty"`
	GasProducerRatioShift      *string `json:"gasProducerRatioShift,omitempty"`
	GasDappRatioMul            *string `json:"gasDappRatioMul,omitempty"`
	GasDappRatioShift          *string `json:"gasDappRatioShift,omitempty"`
	PolicyFeeCreateTokenBase   *string `json:"policyFeeCreateTokenBase,omitempty"`
	PolicyFeeCreateTokenSymbol *string `json:"policyFeeCreateTokenSymbol,omitempty"`
	PolicyFeeCreateTokenSeries *string `json:"policyFeeCreateTokenSeries,omitempty"`
	PolicyFeeRegisterName      *string `json:"policyFeeRegisterName,omitempty"`
	LegacyDataEscrowPerRow     *string `json:"legacyDataEscrowPerRow,omitempty"`
}

func (GasConfigArguments) isSpecialResolutionArguments() {}

// ChainConfigArguments are the arguments of governance.SetChainConfig.
type ChainConfigArguments struct {
	Version         string `json:"version"`
	Reserved1       string `json:"reserved1"`
	Reserved2       string `json:"reserved2"`
	Reserved3       string `json:"reserved3"`
	AllowedTxTypes  string `json:"allowedTxTypes"`
	ExpiryWindow    string `json:"expiryWindow"`
	BlockRateTarget string `json:"blockRateTarget"`
}

func (ChainConfigArguments) isSpecialResolutionArguments() {}

// NestedResolutionArguments are the arguments of governance.SpecialResolution: a resolution nested
// inside another one. Its own calls are reported in the carrying call's Calls, not here.
type NestedResolutionArguments struct {
	// ResolutionID is rendered as a string here, unlike the numeric ResolutionID of the resolution
	// envelope.
	ResolutionID string `json:"resolutionId"`
}

func (NestedResolutionArguments) isSpecialResolutionArguments() {}

// MetadataArguments are the arguments of governance.SetMetadata.
type MetadataArguments struct {
	Metadata map[string]VMValue `json:"metadata"`
}

func (MetadataArguments) isSpecialResolutionArguments() {}

// ConsensusNode is one node of a governance.SetNodeConfig call.
type ConsensusNode struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// NodeConfigArguments are the arguments of governance.SetNodeConfig.
type NodeConfigArguments struct {
	Nodes []ConsensusNode `json:"nodes"`
}

func (NodeConfigArguments) isSpecialResolutionArguments() {}

// RegisterNameArguments are the arguments of governance.RegisterName.
type RegisterNameArguments struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

func (RegisterNameArguments) isSpecialResolutionArguments() {}

// AddressArguments is a single address argument, shared by governance.LookupName and
// token.GetBalances.
type AddressArguments struct {
	Address string `json:"address"`
}

func (AddressArguments) isSpecialResolutionArguments() {}

// NameArguments is a single name argument, shared by governance.LookupAddress and
// phantasma_vm.IsContractDeployed.
type NameArguments struct {
	Name string `json:"name"`
}

func (NameArguments) isSpecialResolutionArguments() {}
