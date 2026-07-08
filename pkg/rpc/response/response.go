package response

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	chain "github.com/phantasma-io/phantasma-sdk-go/pkg/blockchain"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/carbon"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/io"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/util"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/vm"
)

// ErrorResult is returned by endpoints that encode failures inside the result object.
type ErrorResult struct {
	Error string `json:"error"`
}

// SingleResult wraps endpoints that return a single scalar value under an error-shaped field.
type SingleResult struct {
	Value interface{} `json:"error"`
}

// BalanceResult describes a token balance on a chain.
type BalanceResult struct {
	Chain    string   `json:"chain"`
	Amount   string   `json:"amount"`
	Symbol   string   `json:"symbol"`
	Decimals uint     `json:"decimals"`
	Ids      []string `json:"ids,omitempty"`
}

// Clone returns a deep copy of the balance result.
func (b BalanceResult) Clone() *BalanceResult {
	clone := b
	if b.Ids != nil {
		clone.Ids = make([]string, len(b.Ids))
		copy(clone.Ids, b.Ids)
	}

	return &clone
}

// ConvertDecimals returns Amount formatted with the token decimal separator.
func (b BalanceResult) ConvertDecimals() string {
	return util.ConvertDecimalsEx(b.Amount, int(b.Decimals), ".")
}

// ConvertDecimalsToFloat returns Amount as a floating point value using Decimals.
func (b BalanceResult) ConvertDecimalsToFloat() *big.Float {
	f, _ := big.NewFloat(0).SetString(b.ConvertDecimals())
	return f
}

// InteropResult describes platform interop metadata.
type InteropResult struct {
	Local    string `json:"local"`
	External string `json:"external"`
}

// PlatformResult describes a platform connected to the nexus.
type PlatformResult struct {
	Platform string          `json:"platform"`
	Chain    string          `json:"chain"`
	Fuel     string          `json:"fuel"`
	Tokens   []string        `json:"tokens"`
	Interop  []InteropResult `json:"interop"`
}

// GovernanceResult describes a governance key/value setting.
type GovernanceResult struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// OrganizationResult describes organization metadata.
type OrganizationResult struct {
	Name        *string               `json:"name,omitempty"`
	Owner       *string               `json:"owner,omitempty"`
	CarbonOwner *string               `json:"carbonOwner,omitempty"`
	Metadata    []TokenPropertyResult `json:"metadata,omitempty"`
	MemberCount *string               `json:"memberCount,omitempty"`
}

// OrganizationMemberResult describes one organization membership lookup row.
type OrganizationMemberResult struct {
	Address       *string `json:"address,omitempty"`
	CarbonAddress *string `json:"carbonAddress,omitempty"`
	IsMember      bool    `json:"isMember"`
	MemberTime    *uint64 `json:"memberTime,omitempty"`
}

// CrowdsaleResult describes a crowdsale contract state.
type CrowdsaleResult struct {
	Hash          string `json:"hash"`
	Name          string `json:"name"`
	Creator       string `json:"creator"`
	Flags         string `json:"flags"`
	StartDate     uint   `json:"startDate"`
	EndDate       uint   `json:"endDate"`
	SellSymbol    string `json:"sellSymbol"`
	ReceiveSymbol string `json:"receiveSymbol"`
	Price         uint   `json:"price"`
	GlobalSoftCap string `json:"globalSoftCap"`
	GlobalHardCap string `json:"globalHardCap"`
	UserSoftCap   string `json:"userSoftCap"`
	UserHardCap   string `json:"userHardCap"`
}

// NexusResult describes nexus-level metadata.
type NexusResult struct {
	Name          *string            `json:"name,omitempty"`
	Protocol      uint               `json:"protocol"`
	Platforms     []PlatformResult   `json:"platforms,omitempty"`
	Tokens        []TokenResult      `json:"tokens,omitempty"`
	Chains        []ChainResult      `json:"chains,omitempty"`
	Governance    []GovernanceResult `json:"governance,omitempty"`
	Organizations []string           `json:"organizations,omitempty"`
}

// StakeResult describes an account stake and unclaimed reward state.
type StakeResult struct {
	Amount    string `json:"amount"`
	Time      uint   `json:"time"`
	Unclaimed string `json:"unclaimed"`
}

// ConvertDecimals returns Amount formatted with SOUL token decimals.
func (s StakeResult) ConvertDecimals() string {
	return util.ConvertDecimalsEx(s.Amount, 8, ".") // Phantasma Stake token (SOUL) has 8 decimals
}

// ConvertDecimalsToFloat returns Amount as a floating point SOUL value.
func (s StakeResult) ConvertDecimalsToFloat() *big.Float {
	f, _ := big.NewFloat(0).SetString(s.ConvertDecimals())
	return f
}

// StorageResult describes account storage quota and archive usage.
type StorageResult struct {
	Available uint            `json:"available"`
	Used      uint            `json:"used"`
	Avatar    string          `json:"avatar"`
	Archives  []ArchiveResult `json:"archives"`
}

// AccountResult describes account state returned by account endpoints.
type AccountResult struct {
	Address   string          `json:"address"`
	Name      string          `json:"name"`
	Stakes    StakeResult     `json:"stakes"`
	Stake     string          `json:"stake"`
	Unclaimed string          `json:"unclaimed"`
	Relay     *string         `json:"relay,omitempty"`
	Validator string          `json:"validator"`
	Storage   StorageResult   `json:"storage"`
	Balances  []BalanceResult `json:"balances"`
	Txs       []string        `json:"txs,omitempty"`
}

// Clone returns a deep copy of the account result.
func (a AccountResult) Clone() *AccountResult {
	clone := a
	if a.Relay != nil {
		relay := *a.Relay
		clone.Relay = &relay
	}
	if a.Txs != nil {
		clone.Txs = append([]string(nil), a.Txs...)
	}
	clone.Balances = make([]BalanceResult, len(a.Balances))
	for i, b := range a.Balances {
		clone.Balances[i] = *b.Clone()
	}

	return &clone
}

// GetTokenBalance returns the balance row for t, creating an empty row when absent.
func (a *AccountResult) GetTokenBalance(t TokenResult) *BalanceResult {
	for i, b := range a.Balances {
		if b.Symbol == t.Symbol {
			return &a.Balances[i]
		}
	}

	b := BalanceResult{Chain: "main", Amount: "0", Symbol: t.Symbol, Decimals: uint(t.Decimals), Ids: []string{}}

	if a.Balances == nil {
		a.Balances = []BalanceResult{}
	}
	a.Balances = append(a.Balances, b)

	return &a.Balances[len(a.Balances)-1]
}

// AddressTransactionsResult contains paged transactions for an address.
type AddressTransactionsResult struct {
	Address string              `json:"address"`
	Txs     []TransactionResult `json:"txs"`
}

// LeaderboardRowResult describes one leaderboard row.
type LeaderboardRowResult struct {
	Address string `json:"address"`
	Value   string `json:"value"`
}

// LeaderboardResult describes a named leaderboard.
type LeaderboardResult struct {
	Name *string                `json:"name,omitempty"`
	Rows []LeaderboardRowResult `json:"rows,omitempty"`
}

// DappResult describes an application deployed on a chain.
type DappResult struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
}

// ChainResult describes a chain in the nexus.
type ChainResult struct {
	Name         *string  `json:"name,omitempty"`
	Address      *string  `json:"address,omitempty"`
	Parent       *string  `json:"parent,omitempty"`
	Height       uint     `json:"height"`
	Organization *string  `json:"organization,omitempty"`
	Contracts    []string `json:"contracts,omitempty"`
	Dapps        []string `json:"dapps,omitempty"`
}

// EstimateTransactionResult is the estimateTransaction response: the exact fee bill of one
// serialized transaction envelope, computed by dry-running it against current chain state
// (gas-model-v2 Tier-2). 64-bit amounts arrive as decimal strings (they can exceed the 2^53
// precision of JSON numbers in JavaScript-facing tooling). Amounts are kcal-base atoms of the gas
// token; escrow amounts are data-token atoms. Service availability (routing, gas model, node
// budget) surfaces as a standard RPC error, never through this shape.
type EstimateTransactionResult struct {
	// WouldAbort is true when the transaction would not complete on-chain as submitted; see
	// AbortReason.
	WouldAbort bool `json:"wouldAbort"`
	// AbortReason holds the rejection or abort reason when WouldAbort is true; empty otherwise.
	AbortReason string `json:"abortReason,omitempty"`
	// GasBillKcalBase is the settled gas bill in kcal-base, including the minimum-bill floor and
	// the maxGas clamp (aborted transactions still pay).
	GasBillKcalBase *string `json:"gasBillKcalBase,omitempty"`
	// DataRows is the number of newly paid storage quanta the transaction creates
	// (DataRows * dataEscrowPerRow == DataEscrowAtoms).
	DataRows *string `json:"dataRows,omitempty"`
	// DataEscrowAtoms is the gross storage escrow paid for grown rows, in data-token atoms.
	DataEscrowAtoms *string `json:"dataEscrowAtoms,omitempty"`
	// DataRefundAtoms is the gross storage refunds for shrunk rows, in data-token atoms.
	DataRefundAtoms *string `json:"dataRefundAtoms,omitempty"`
	// RecommendedMaxGas is the recommended TxMsg maxGas: the bill plus a 15% state-drift margin,
	// floored at the chain minimums; "0" when WouldAbort.
	RecommendedMaxGas *string `json:"recommendedMaxGas,omitempty"`
	// RecommendedMaxData is the recommended TxMsg maxData: the net escrow plus a 15% margin,
	// aligned up to whole rows; "0" when WouldAbort or nothing is escrowed.
	RecommendedMaxData *string `json:"recommendedMaxData,omitempty"`
}

// ToFeeEstimate converts a completed estimate into the same carbon.NativeFeeEstimate the Tier-1
// estimator produces, so wallet code consumes both tiers identically: MaxGas/MaxData are the
// recommended ceilings and ExpectedGasBill is the exact settled bill. It returns an error when
// WouldAbort is set - an aborted simulation has no recommendations (retry with a higher offer or
// fall back to the Tier-1 estimator) - and on malformed numeric strings.
func (e *EstimateTransactionResult) ToFeeEstimate() (carbon.NativeFeeEstimate, error) {
	if e.WouldAbort {
		return carbon.NativeFeeEstimate{}, fmt.Errorf("estimateTransaction reported the transaction would abort: %s", e.AbortReason)
	}
	var err error
	parse := func(value *string, fieldName string) uint64 {
		if err != nil {
			return 0
		}
		if value == nil || *value == "" {
			err = fmt.Errorf("estimateTransaction field %s is missing or empty", fieldName)
			return 0
		}
		v, parseErr := strconv.ParseUint(*value, 10, 64)
		if parseErr != nil {
			err = fmt.Errorf("estimateTransaction field %s: %w", fieldName, parseErr)
			return 0
		}
		return v
	}
	estimate := carbon.NativeFeeEstimate{
		MaxGas:          parse(e.RecommendedMaxGas, "recommendedMaxGas"),
		MaxData:         parse(e.RecommendedMaxData, "recommendedMaxData"),
		ExpectedGasBill: parse(e.GasBillKcalBase, "gasBillKcalBase"),
	}
	if err != nil {
		return carbon.NativeFeeEstimate{}, err
	}
	return estimate, nil
}

// GasConfigResult is the getGasConfig response: the current on-chain gas configuration plus the
// chain parameters fee estimation needs. 64-bit config values arrive as decimal strings (they
// can exceed the 2^53 precision of JSON numbers in JavaScript-facing tooling).
type GasConfigResult struct {
	// GasModelVersion is 1 for the original fee model, 2 for gas-model-v2 (config version >= 1).
	GasModelVersion uint32 `json:"gasModelVersion"`
	// GasConfig is the current on-chain config.
	GasConfig *GasConfigDataResult `json:"gasConfig,omitempty"`
	// BlockRateTarget is the chain block rate target in milliseconds.
	BlockRateTarget uint32 `json:"blockRateTarget"`
	// ExpiryWindow is the transaction expiry window in milliseconds.
	ExpiryWindow uint32 `json:"expiryWindow"`
	// UnitsPerBlockDataByte is the gas-model-v2 price of block-carried bytes in gas units per
	// byte; nil under gas model v1.
	UnitsPerBlockDataByte *uint32 `json:"unitsPerBlockDataByte,omitempty"`
}

// GasConfigDataResult is the JSON shape of the on-chain GasConfig. Fields after
// GasBurnRatioShift exist only when the config version is >= 1 (gas-model-v2) and are omitted
// from v1 responses.
type GasConfigDataResult struct {
	Version                 byte    `json:"version"`
	MaxNameLength           byte    `json:"maxNameLength"`
	MaxTokenSymbolLength    byte    `json:"maxTokenSymbolLength"`
	FeeShift                byte    `json:"feeShift"`
	MaxStructureSize        uint32  `json:"maxStructureSize"`
	FeeMultiplier           *string `json:"feeMultiplier,omitempty"`
	GasTokenID              *string `json:"gasTokenId,omitempty"`
	DataTokenID             *string `json:"dataTokenId,omitempty"`
	MinimumGasOffer         *string `json:"minimumGasOffer,omitempty"`
	DataEscrowPerRow        *string `json:"dataEscrowPerRow,omitempty"`
	GasFeeTransfer          *string `json:"gasFeeTransfer,omitempty"`
	GasFeeQuery             *string `json:"gasFeeQuery,omitempty"`
	GasFeeCreateTokenBase   *string `json:"gasFeeCreateTokenBase,omitempty"`
	GasFeeCreateTokenSymbol *string `json:"gasFeeCreateTokenSymbol,omitempty"`
	GasFeeCreateTokenSeries *string `json:"gasFeeCreateTokenSeries,omitempty"`
	GasFeePerByte           *string `json:"gasFeePerByte,omitempty"`
	GasFeeRegisterName      *string `json:"gasFeeRegisterName,omitempty"`
	GasBurnRatioMul         *string `json:"gasBurnRatioMul,omitempty"`
	GasBurnRatioShift       byte    `json:"gasBurnRatioShift"`

	MinimumGasBill             *string `json:"minimumGasBill,omitempty"`
	GasProducerRatioMul        *string `json:"gasProducerRatioMul,omitempty"`
	GasProducerRatioShift      *byte   `json:"gasProducerRatioShift,omitempty"`
	GasDappRatioMul            *string `json:"gasDappRatioMul,omitempty"`
	GasDappRatioShift          *byte   `json:"gasDappRatioShift,omitempty"`
	PolicyFeeCreateTokenBase   *string `json:"policyFeeCreateTokenBase,omitempty"`
	PolicyFeeCreateTokenSymbol *string `json:"policyFeeCreateTokenSymbol,omitempty"`
	PolicyFeeCreateTokenSeries *string `json:"policyFeeCreateTokenSeries,omitempty"`
	PolicyFeeRegisterName      *string `json:"policyFeeRegisterName,omitempty"`
	LegacyDataEscrowPerRow     *string `json:"legacyDataEscrowPerRow,omitempty"`
}

// ToGasConfig converts the getGasConfig JSON response to the wire-format carbon.GasConfig
// consumed by the Tier-1 fee estimator (carbon.EstimateNativeFee). It fails on malformed
// numeric strings and on a v2 response missing tail fields: estimating fees from silently
// zeroed v2 prices would produce rejected transactions.
func (g *GasConfigResult) ToGasConfig() (carbon.GasConfig, error) {
	if g.GasConfig == nil {
		return carbon.GasConfig{}, fmt.Errorf("getGasConfig response has no gasConfig section")
	}
	c := g.GasConfig
	var config carbon.GasConfig
	var err error
	parse := func(value *string, fieldName string) uint64 {
		if err != nil {
			return 0
		}
		if value == nil || *value == "" {
			err = fmt.Errorf("getGasConfig field %s is missing or empty", fieldName)
			return 0
		}
		v, parseErr := strconv.ParseUint(*value, 10, 64)
		if parseErr != nil {
			err = fmt.Errorf("getGasConfig field %s: %w", fieldName, parseErr)
			return 0
		}
		return v
	}
	config.Version = c.Version
	config.MaxNameLength = c.MaxNameLength
	config.MaxTokenSymbolLength = c.MaxTokenSymbolLength
	config.FeeShift = c.FeeShift
	config.MaxStructureSize = c.MaxStructureSize
	config.FeeMultiplier = parse(c.FeeMultiplier, "feeMultiplier")
	config.GasTokenID = parse(c.GasTokenID, "gasTokenId")
	config.DataTokenID = parse(c.DataTokenID, "dataTokenId")
	config.MinimumGasOffer = parse(c.MinimumGasOffer, "minimumGasOffer")
	config.DataEscrowPerRow = parse(c.DataEscrowPerRow, "dataEscrowPerRow")
	config.GasFeeTransfer = parse(c.GasFeeTransfer, "gasFeeTransfer")
	config.GasFeeQuery = parse(c.GasFeeQuery, "gasFeeQuery")
	config.GasFeeCreateTokenBase = parse(c.GasFeeCreateTokenBase, "gasFeeCreateTokenBase")
	config.GasFeeCreateTokenSymbol = parse(c.GasFeeCreateTokenSymbol, "gasFeeCreateTokenSymbol")
	config.GasFeeCreateTokenSeries = parse(c.GasFeeCreateTokenSeries, "gasFeeCreateTokenSeries")
	config.GasFeePerByte = parse(c.GasFeePerByte, "gasFeePerByte")
	config.GasFeeRegisterName = parse(c.GasFeeRegisterName, "gasFeeRegisterName")
	config.GasBurnRatioMul = parse(c.GasBurnRatioMul, "gasBurnRatioMul")
	config.GasBurnRatioShift = c.GasBurnRatioShift
	if config.Version >= 1 {
		config.MinimumGasBill = parse(c.MinimumGasBill, "minimumGasBill")
		config.GasProducerRatioMul = parse(c.GasProducerRatioMul, "gasProducerRatioMul")
		config.GasDappRatioMul = parse(c.GasDappRatioMul, "gasDappRatioMul")
		config.PolicyFeeCreateTokenBase = parse(c.PolicyFeeCreateTokenBase, "policyFeeCreateTokenBase")
		config.PolicyFeeCreateTokenSymbol = parse(c.PolicyFeeCreateTokenSymbol, "policyFeeCreateTokenSymbol")
		config.PolicyFeeCreateTokenSeries = parse(c.PolicyFeeCreateTokenSeries, "policyFeeCreateTokenSeries")
		config.PolicyFeeRegisterName = parse(c.PolicyFeeRegisterName, "policyFeeRegisterName")
		config.LegacyDataEscrowPerRow = parse(c.LegacyDataEscrowPerRow, "legacyDataEscrowPerRow")
		if err == nil && c.GasProducerRatioShift == nil {
			err = fmt.Errorf("getGasConfig field gasProducerRatioShift is missing")
		}
		if err == nil && c.GasDappRatioShift == nil {
			err = fmt.Errorf("getGasConfig field gasDappRatioShift is missing")
		}
		if err == nil {
			config.GasProducerRatioShift = *c.GasProducerRatioShift
			config.GasDappRatioShift = *c.GasDappRatioShift
		}
	}
	if err != nil {
		return carbon.GasConfig{}, err
	}
	return config, nil
}

// EventResult describes one event emitted by a transaction or script invocation.
type EventResult struct {
	Address  string `json:"address"`
	Contract string `json:"contract"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Data     string `json:"data"`
}

func (e *EventResult) UnmarshalJSON(data []byte) error {
	data, err := stripExactFields(data, "Kind", "Data")
	if err != nil {
		return err
	}

	type eventResult EventResult
	return json.Unmarshal(data, (*eventResult)(e))
}

// OracleResult describes oracle data attached to a block or script result.
type OracleResult struct {
	URL     string `json:"url"`
	Content string `json:"content"`
}

// SignatureResult describes a transaction signature.
type SignatureResult struct {
	Kind string `json:"kind"`
	Data string `json:"data"`
}

func (s *SignatureResult) UnmarshalJSON(data []byte) error {
	data, err := stripExactFields(data, "Kind", "Data")
	if err != nil {
		return err
	}

	type signatureResult SignatureResult
	return json.Unmarshal(data, (*signatureResult)(s))
}

// EventExResult describes one extended transaction event returned by current RPC nodes.
type EventExResult struct {
	Address  string      `json:"address"`
	Contract string      `json:"contract"`
	Kind     string      `json:"kind"`
	Data     interface{} `json:"data"`
}

func (e *EventExResult) UnmarshalJSON(data []byte) error {
	data, err := stripExactFields(data, "Kind", "Data")
	if err != nil {
		return err
	}

	type eventExResult EventExResult
	return json.Unmarshal(data, (*eventExResult)(e))
}

// TransactionResult describes a transaction returned by RPC.
type TransactionResult struct {
	Hash           string            `json:"hash"`
	ChainAddress   string            `json:"chainAddress"`
	Timestamp      uint              `json:"timestamp"`
	BlockHeight    int               `json:"blockHeight"`
	BlockHash      string            `json:"blockHash"`
	Script         string            `json:"script"`
	Payload        string            `json:"payload"`
	CarbonTxType   uint32            `json:"carbonTxType"`
	CarbonTxData   string            `json:"carbonTxData"`
	DebugComment   *string           `json:"debugComment,omitempty"`
	Events         []EventResult     `json:"events"`
	ExtendedEvents []EventExResult   `json:"extendedEvents"`
	State          string            `json:"state"`
	Result         string            `json:"result"`
	Fee            string            `json:"fee"`
	Signatures     []SignatureResult `json:"signatures"`
	Sender         string            `json:"sender"`
	GasPayer       string            `json:"gasPayer"`
	GasTarget      string            `json:"gasTarget"`
	GasPrice       string            `json:"gasPrice"`
	GasLimit       string            `json:"gasLimit"`
	Expiration     uint              `json:"expiration"`
}

// StateIsSuccess reports whether the transaction state is a success state.
func (t TransactionResult) StateIsSuccess() bool {
	return chain.TxStateIsSuccess(t.State)
}

// StateIsFault reports whether the transaction state is a fault state.
func (t TransactionResult) StateIsFault() bool {
	return chain.TxStateIsFault(t.State)
}

// PaginatedResult represents page-based RPC pagination.
type PaginatedResult[T any] struct {
	Page       uint `json:"page"`
	PageSize   uint `json:"pageSize"`
	Total      uint `json:"total"`
	TotalPages uint `json:"totalPages"`

	Result T `json:"result"`
}

// CursorPaginatedResult represents Carbon cursor pagination.
type CursorPaginatedResult[T any] struct {
	Result T       `json:"result"`
	Cursor *string `json:"cursor,omitempty"`
}

// BlockResult describes a block and, when requested, its transactions/events.
type BlockResult struct {
	Hash             string              `json:"hash"`
	PreviousHash     string              `json:"previousHash"`
	Timestamp        uint                `json:"timestamp"`
	Height           uint                `json:"height"`
	ChainAddress     string              `json:"chainAddress"`
	Protocol         uint                `json:"protocol"`
	Txs              []TransactionResult `json:"txs"`
	ValidatorAddress string              `json:"validatorAddress"`
	// ProducerAddress is the fee payout address stamped by the block producer inside the hashed
	// block input. Non-nil on gas-model-v2 blocks, nil on earlier blocks. Distinct from
	// ValidatorAddress (the consensus-log leader): usually equal today, but a configurable payout
	// address is a planned compatible extension, so consumers must not assume equality.
	ProducerAddress *string        `json:"producerAddress,omitempty"`
	Reward          string         `json:"reward"`
	Events          []EventResult  `json:"events,omitempty"`
	Oracles         []OracleResult `json:"oracles,omitempty"`
}

// TokenExternalResult describes an external platform mapping for a token.
type TokenExternalResult struct {
	Platform string `json:"platform"`
	Hash     string `json:"hash"`
}

// TokenPriceResult describes a token price candle.
type TokenPriceResult struct {
	Timestamp uint   `json:"Timestamp"`
	Open      string `json:"Open"`
	High      string `json:"High"`
	Low       string `json:"Low"`
	Close     string `json:"Close"`
}

// TokenResult describes a token definition.
type TokenResult struct {
	Symbol        string                `json:"symbol"`
	Name          string                `json:"name"`
	Decimals      int                   `json:"decimals"`
	CurrentSupply string                `json:"currentSupply"`
	MaxSupply     string                `json:"maxSupply"`
	BurnedSupply  string                `json:"burnedSupply"`
	Address       string                `json:"address"`
	Owner         string                `json:"owner"`
	Flags         string                `json:"flags"`
	Script        *string               `json:"script,omitempty"`
	Series        []TokenSeriesResult   `json:"series"`
	CarbonID      string                `json:"carbonId"`
	Metadata      []TokenPropertyResult `json:"metadata,omitempty"`
	TokenSchemas  *TokenSchemasResult   `json:"tokenSchemas"`
	External      []TokenExternalResult `json:"external,omitempty"`
	Price         []TokenPriceResult    `json:"price,omitempty"`
}

func (t *TokenResult) UnmarshalJSON(data []byte) error {
	data, err := stripExactFields(data, "carbonID")
	if err != nil {
		return err
	}

	type tokenResult TokenResult
	return json.Unmarshal(data, (*tokenResult)(t))
}

// IsBurnable reports whether the token has the Burnable flag.
func (t TokenResult) IsBurnable() bool {
	return slices.Contains(strings.Split(t.Flags, ", "), "Burnable")
}

// IsDivisible reports whether the token has the Divisible flag.
func (t TokenResult) IsDivisible() bool {
	return slices.Contains(strings.Split(t.Flags, ", "), "Divisible")
}

// IsFiat reports whether the token has the Fiat flag.
func (t TokenResult) IsFiat() bool {
	return slices.Contains(strings.Split(t.Flags, ", "), "Fiat")
}

// IsFinite reports whether the token has the Finite flag.
func (t TokenResult) IsFinite() bool {
	return slices.Contains(strings.Split(t.Flags, ", "), "Finite")
}

// IsFuel reports whether the token has the Fuel flag.
func (t TokenResult) IsFuel() bool {
	return slices.Contains(strings.Split(t.Flags, ", "), "Fuel")
}

// IsFungible reports whether the token has the Fungible flag.
func (t TokenResult) IsFungible() bool {
	return slices.Contains(strings.Split(t.Flags, ", "), "Fungible")
}

// IsMintable reports whether the token has the Mintable flag.
func (t TokenResult) IsMintable() bool {
	return slices.Contains(strings.Split(t.Flags, ", "), "Mintable")
}

// IsStakable reports whether the token has the Stakable flag.
func (t TokenResult) IsStakable() bool {
	return slices.Contains(strings.Split(t.Flags, ", "), "Stakable")
}

// IsTransferable reports whether the token has the Transferable flag.
func (t TokenResult) IsTransferable() bool {
	return slices.Contains(strings.Split(t.Flags, ", "), "Transferable")
}

// TokenSeriesResult describes a non-fungible token series.
type TokenSeriesResult struct {
	SeriesID       string                `json:"seriesId"`
	CarbonTokenID  string                `json:"carbonTokenId"`
	CarbonSeriesID string                `json:"carbonSeriesId"`
	OwnerAddress   string                `json:"ownerAddress"`
	MaxMint        string                `json:"maxMint"`
	MintCount      string                `json:"mintCount"`
	CurrentSupply  string                `json:"currentSupply"`
	MaxSupply      string                `json:"maxSupply"`
	BurnedSupply   *string               `json:"burnedSupply,omitempty"`
	Mode           *string               `json:"mode,omitempty"`
	Script         *string               `json:"script,omitempty"`
	Methods        []ABIMethodResult     `json:"methods,omitempty"`
	Metadata       []TokenPropertyResult `json:"metadata"`
}

// TokenPropertyResult describes a token metadata key/value pair.
type TokenPropertyResult struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (p *TokenPropertyResult) UnmarshalJSON(data []byte) error {
	data, err := stripExactFields(data, "Key", "Value")
	if err != nil {
		return err
	}

	type tokenPropertyResult TokenPropertyResult
	return json.Unmarshal(data, (*tokenPropertyResult)(p))
}

// VMVariableSchemaResult describes a VM metadata field schema.
type VMVariableSchemaResult struct {
	Type   string                `json:"type"`
	Schema *VMStructSchemaResult `json:"schema,omitempty"`
}

// VMNamedVariableSchemaResult describes a named VM metadata schema field.
type VMNamedVariableSchemaResult struct {
	Name   string                 `json:"name"`
	Schema VMVariableSchemaResult `json:"schema"`
}

// VMStructSchemaResult describes a VM metadata struct schema.
type VMStructSchemaResult struct {
	Fields []VMNamedVariableSchemaResult `json:"fields"`
	Flags  uint                          `json:"flags"`
}

// TokenSchemasResult describes token series, ROM and RAM metadata schemas.
type TokenSchemasResult struct {
	SeriesMetadata VMStructSchemaResult `json:"seriesMetadata"`
	ROM            VMStructSchemaResult `json:"rom"`
	RAM            VMStructSchemaResult `json:"ram"`
}

// TokenDataResult describes one NFT instance.
type TokenDataResult struct {
	ID               string                `json:"id"`
	Series           string                `json:"series"`
	CarbonTokenID    string                `json:"carbonTokenId"`
	CarbonSeriesID   string                `json:"carbonSeriesId"`
	CarbonNFTAddress string                `json:"carbonNftAddress"`
	Mint             string                `json:"mint"`
	ChainName        string                `json:"chainName"`
	OwnerAddress     string                `json:"ownerAddress"`
	CreatorAddress   string                `json:"creatorAddress"`
	RAM              string                `json:"ram"`
	ROM              string                `json:"rom"`
	Status           string                `json:"status"`
	Infusion         []TokenPropertyResult `json:"infusion"`
	Properties       []TokenPropertyResult `json:"properties"`
}

func (t *TokenDataResult) UnmarshalJSON(data []byte) error {
	data, err := stripExactFields(data, "ID")
	if err != nil {
		return err
	}

	type tokenDataResult TokenDataResult
	return json.Unmarshal(data, (*tokenDataResult)(t))
}

// SendRawTxResult is the classic send-transaction response shape.
type SendRawTxResult struct {
	Hash  string `json:"hash"`
	Error string `json:"error"`
}

// BuildInfoResult describes the node build/version metadata.
type BuildInfoResult struct {
	Version      string `json:"version"`
	Commit       string `json:"commit"`
	BuildTimeUTC string `json:"buildTimeUtc"`
}

// PhantasmaVMConfigResult describes the active Phantasma VM gas/config values.
type PhantasmaVMConfigResult struct {
	IsStored              bool   `json:"isStored"`
	FeatureLevel          int    `json:"featureLevel"`
	GasConstructor        string `json:"gasConstructor"`
	GasNexus              string `json:"gasNexus"`
	GasOrganization       string `json:"gasOrganization"`
	GasAccount            string `json:"gasAccount"`
	GasLeaderboard        string `json:"gasLeaderboard"`
	GasStandard           string `json:"gasStandard"`
	GasOracle             string `json:"gasOracle"`
	FuelPerContractDeploy string `json:"fuelPerContractDeploy"`
}

// AuctionResult describes an auction listing.
type AuctionResult struct {
	CreatorAddress  string `json:"creatorAddress"`
	ChainAddress    string `json:"chainAddress"`
	StartDate       uint   `json:"startDate"`
	EndDate         uint   `json:"endDate"`
	BaseSymbol      string `json:"baseSymbol"`
	QuoteSymbol     string `json:"quoteSymbol"`
	TokenID         string `json:"tokenId"`
	Price           string `json:"price"`
	EndPrice        string `json:"endPrice"`
	ExtensionPeriod string `json:"extensionPeriod"`
	Type            string `json:"type"`
	ROM             string `json:"rom"`
	RAM             string `json:"ram"`
	ListingFee      string `json:"listingFee"`
	CurrentWinner   string `json:"currentWinner"`
}

// ScriptResult describes a read-only script invocation result.
type ScriptResult struct {
	Events  []EventResult  `json:"events"`
	Result  string         `json:"result,omitempty"`
	Error   *string        `json:"error,omitempty"`
	Results []string       `json:"results"`
	Oracles []OracleResult `json:"oracles"`
	State   *string        `json:"state,omitempty"`
	Gas     *string        `json:"gas,omitempty"`
}

// DecodeResultWithError decodes the hex-encoded Result field into a VMObject.
func (s ScriptResult) DecodeResultWithError() (*vm.VMObject, error) {
	return decodeVMObjectHex(s.Result)
}

// DecodeResultsWithError decodes the hex-encoded Results entry at index into a VMObject.
func (s ScriptResult) DecodeResultsWithError(index int) (*vm.VMObject, error) {
	if index < 0 || index >= len(s.Results) {
		return nil, fmt.Errorf("script result index %d out of range", index)
	}
	return decodeVMObjectHex(s.Results[index])
}

func decodeVMObjectHex(value string) (*vm.VMObject, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, err
	}

	reader := io.NewBinReaderFromBuf(decoded)
	object := &vm.VMObject{}
	object.Deserialize(reader)
	if reader.Err != nil {
		return nil, reader.Err
	}
	return object, nil
}

// ArchiveResult describes archive metadata.
type ArchiveResult struct {
	Name          *string  `json:"name,omitempty"`
	Hash          *string  `json:"hash,omitempty"`
	Time          uint     `json:"time"`
	Size          uint     `json:"size"`
	Encryption    *string  `json:"encryption,omitempty"`
	BlockCount    int      `json:"blockCount"`
	MissingBlocks []int    `json:"missingBlocks,omitempty"`
	Owners        []string `json:"owners,omitempty"`
}

// ABIParameterResult describes a contract ABI parameter.
type ABIParameterResult struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ABIMethodResult describes a contract ABI method.
type ABIMethodResult struct {
	Name       string               `json:"name"`
	ReturnType string               `json:"returnType"`
	Parameters []ABIParameterResult `json:"parameters"`
}

// ABIEventResult describes a contract ABI event.
type ABIEventResult struct {
	Value       int    `json:"value"`
	Name        string `json:"name"`
	ReturnType  string `json:"returnType"`
	Description string `json:"description"`
}

// ContractResult describes a deployed contract.
type ContractResult struct {
	Name    string            `json:"name"`
	Address string            `json:"address"`
	Script  string            `json:"script"`
	Owner   *string           `json:"owner,omitempty"`
	Methods []ABIMethodResult `json:"methods,omitempty"`
	Events  []ABIEventResult  `json:"events,omitempty"`
}

// ChannelResult describes a payment/storage channel.
type ChannelResult struct {
	CreatorAddress string `json:"creatorAddress"`
	TargetAddress  string `json:"targetAddress"`
	Name           string `json:"name"`
	Chain          string `json:"chain"`
	CreationTime   uint   `json:"creationTime"`
	Symbol         string `json:"symbol"`
	Fee            string `json:"fee"`
	Balance        string `json:"balance"`
	Active         bool   `json:"active"`
	Index          int    `json:"index"`
}

// ReceiptResult describes a channel receipt.
type ReceiptResult struct {
	Nexus     string `json:"nexus"`
	Channel   string `json:"channel"`
	Index     string `json:"index"`
	Timestamp uint   `json:"timestamp"`
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Script    string `json:"script"`
}

// PeerResult describes a network peer.
type PeerResult struct {
	URL     string `json:"url"`
	Version string `json:"version"`
	Flags   string `json:"flags"`
	Fee     string `json:"fee"`
	Pow     uint   `json:"pow"`
}

// ValidatorResult describes a validator entry.
type ValidatorResult struct {
	Address string `json:"address"`
	Type    string `json:"type"`
}

// SwapResult describes a cross-chain swap.
type SwapResult struct {
	SourcePlatform      string `json:"sourcePlatform"`
	SourceChain         string `json:"sourceChain"`
	SourceHash          string `json:"sourceHash"`
	SourceAddress       string `json:"sourceAddress"`
	DestinationPlatform string `json:"destinationPlatform"`
	DestinationChain    string `json:"destinationChain"`
	DestinationHash     string `json:"destinationHash"`
	DestinationAddress  string `json:"destinationAddress"`
	Symbol              string `json:"symbol"`
	Value               string `json:"value"`
}

func stripExactFields(data []byte, fields ...string) ([]byte, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	changed := false
	for _, field := range fields {
		if _, ok := raw[field]; ok {
			delete(raw, field)
			changed = true
		}
	}

	if !changed {
		return data, nil
	}
	return json.Marshal(raw)
}
