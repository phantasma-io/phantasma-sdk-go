package response

// Arguments of the phantasma_vm module calls a special resolution can carry, including the
// migration calls that rebuild contracts and series. Every numeric field is a string for the
// reason stated on SpecialResolutionArguments.

// ExecuteScriptArguments are the arguments of phantasma_vm.ExecuteScript.
type ExecuteScriptArguments struct {
	MaxGas  string `json:"maxGas"`
	GasFrom string `json:"gasFrom"`
	Script  string `json:"script"`
}

func (ExecuteScriptArguments) isSpecialResolutionArguments() {}

// RegisterTokenContractArguments are the arguments of phantasma_vm.RegisterTokenContract.
type RegisterTokenContractArguments struct {
	TokenID string `json:"tokenId"`
	Symbol  string `json:"symbol"`
	Script  string `json:"script"`
	Abi     string `json:"abi"`
	// Token is the resolved token symbol; absent when the token could not be resolved at answer
	// time.
	Token *string `json:"token,omitempty"`
}

func (RegisterTokenContractArguments) isSpecialResolutionArguments() {}

// DeployContractArguments are the arguments of phantasma_vm.DeployContract.
type DeployContractArguments struct {
	From         string `json:"from"`
	ContractName string `json:"contractName"`
	Script       string `json:"script"`
	Abi          string `json:"abi"`
}

func (DeployContractArguments) isSpecialResolutionArguments() {}

// PhantasmaVMConfigArguments are the arguments of phantasma_vm.SetConfig.
type PhantasmaVMConfigArguments struct {
	FeatureLevel          string `json:"featureLevel"`
	GasConstructor        string `json:"gasConstructor"`
	GasNexus              string `json:"gasNexus"`
	GasOrganization       string `json:"gasOrganization"`
	GasAccount            string `json:"gasAccount"`
	GasLeaderboard        string `json:"gasLeaderboard"`
	GasStandard           string `json:"gasStandard"`
	GasOracle             string `json:"gasOracle"`
	FuelPerContractDeploy string `json:"fuelPerContractDeploy"`
}

func (PhantasmaVMConfigArguments) isSpecialResolutionArguments() {}

// ContractStorageRow is a key/value row of contract storage. Both sides are hex-encoded because
// they hold arbitrary bytes.
type ContractStorageRow struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ContractStorageTable is one map or list table of a contract, with every row it carries.
type ContractStorageTable struct {
	Name string               `json:"name"`
	Rows []ContractStorageRow `json:"rows"`
}

// ImportedContract is one contract restored by a migration: identity, code and the whole of its
// stored state.
type ImportedContract struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Owner   string `json:"owner"`
	Script  string `json:"script"`
	Abi     string `json:"abi"`
	// RootVariables are the root-level contract variables.
	RootVariables []ContractStorageRow `json:"rootVariables"`
	// Tables are the map and list tables, including their backing rows.
	Tables []ContractStorageTable `json:"tables"`
}

// ImportContractsArguments are the arguments of phantasma_vm.ImportContracts.
type ImportContractsArguments struct {
	ContractsCount string             `json:"contractsCount"`
	Contracts      []ImportedContract `json:"contracts"`
}

func (ImportContractsArguments) isSpecialResolutionArguments() {}

// SeriesSupplement is the definition needed to rebuild one Phantasma series.
type SeriesSupplement struct {
	Token             string `json:"token"`
	TokenID           string `json:"tokenId"`
	PhantasmaSeriesID string `json:"phantasmaSeriesId"`
	MaxSupply         string `json:"maxSupply"`
	MintCount         string `json:"mintCount"`
	Mode              string `json:"mode"`
	Script            string `json:"script"`
	Abi               string `json:"abi"`
	Rom               string `json:"rom"`
}

// SeriesMintCountRepair is the mint-count repair of one Phantasma series.
type SeriesMintCountRepair struct {
	Token             string `json:"token"`
	TokenID           string `json:"tokenId"`
	PhantasmaSeriesID string `json:"phantasmaSeriesId"`
	ImportedLiveCount string `json:"importedLiveCount"`
	Script            string `json:"script"`
	Abi               string `json:"abi"`
}

// RepairSeriesArguments are the arguments of phantasma_vm.RepairSeries.
type RepairSeriesArguments struct {
	SupplementsCount string                  `json:"supplementsCount"`
	Supplements      []SeriesSupplement      `json:"supplements"`
	RepairsCount     string                  `json:"repairsCount"`
	Repairs          []SeriesMintCountRepair `json:"repairs"`
}

func (RepairSeriesArguments) isSpecialResolutionArguments() {}

// TokenRepair is the repair of one token definition.
type TokenRepair struct {
	Token      string `json:"token"`
	TokenID    string `json:"tokenId"`
	Symbol     string `json:"symbol"`
	Script     string `json:"script"`
	Abi        string `json:"abi"`
	TokenFlags string `json:"tokenFlags"`
	// RepairMask is the bitmask of the repair operations the chain was asked to perform. It stays
	// numeric on purpose: a new chain-side operation must not silently render as an unrelated name
	// here.
	RepairMask string `json:"repairMask"`
}

// RepairTokenArguments are the arguments of phantasma_vm.RepairToken.
type RepairTokenArguments struct {
	RepairsCount string        `json:"repairsCount"`
	Repairs      []TokenRepair `json:"repairs"`
}

func (RepairTokenArguments) isSpecialResolutionArguments() {}
