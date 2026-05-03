package response

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"slices"
	"strings"

	chain "github.com/phantasma-io/phantasma-go/pkg/blockchain"
	"github.com/phantasma-io/phantasma-go/pkg/io"
	"github.com/phantasma-io/phantasma-go/pkg/util"
	"github.com/phantasma-io/phantasma-go/pkg/vm"
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
	Ids      []string `json:"ids"`
}

// Clone returns a deep copy of the balance result.
func (b BalanceResult) Clone() *BalanceResult {
	clone := b
	clone.Ids = make([]string, len(b.Ids))
	copy(clone.Ids, b.Ids)

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

// OrganizationResult describes an organization and its members.
type OrganizationResult struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
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
	Name          string             `json:"name"`
	Protocol      uint               `json:"protocol"`
	Platforms     []PlatformResult   `json:"platforms"`
	Tokens        []TokenResult      `json:"tokens"`
	Chains        []ChainResult      `json:"chains"`
	Governance    []GovernanceResult `json:"governance"`
	Organizations []string           `json:"organizations"`
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
	Relay     string          `json:"relay"`
	Validator string          `json:"validator"`
	Storage   StorageResult   `json:"storage"`
	Balances  []BalanceResult `json:"balances"`
}

// Clone returns a deep copy of the account result.
func (a AccountResult) Clone() *AccountResult {
	clone := a
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
	Name string                 `json:"name"`
	Rows []LeaderboardRowResult `json:"rows"`
}

// DappResult describes an application deployed on a chain.
type DappResult struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
}

// ChainResult describes a chain in the nexus.
type ChainResult struct {
	Name         string   `json:"name"`
	Address      string   `json:"address"`
	Parent       string   `json:"parent"`
	Height       uint     `json:"height"`
	Organization string   `json:"organization"`
	Contracts    []string `json:"contracts"`
	Dapps        []string `json:"dapps"`
}

// EventResult describes one event emitted by a transaction or script invocation.
type EventResult struct {
	Address  string `json:"address"`
	Contract string `json:"contract"`
	Kind     string `json:"kind"`
	Data     string `json:"data"`
}

// OracleResult describes oracle data attached to a block or script result.
type OracleResult struct {
	URL     string `json:"url"`
	Content string `json:"content"`
}

// SignatureResult describes a transaction signature.
type SignatureResult struct {
	Kind string `json:"Kind"`
	Data string `json:"Data"`
}

// TransactionResult describes a transaction returned by RPC.
type TransactionResult struct {
	Hash         string            `json:"hash"`
	ChainAddress string            `json:"chainAddress"`
	Timestamp    uint              `json:"timestamp"`
	BlockHeight  int               `json:"blockHeight"`
	BlockHash    string            `json:"blockHash"`
	Script       string            `json:"script"`
	Payload      string            `json:"payload"`
	Events       []EventResult     `json:"events"`
	State        string            `json:"state"`
	Result       string            `json:"result"`
	Fee          string            `json:"fee"`
	Signatures   []SignatureResult `json:"signatures"`
	Expiration   uint              `json:"expiration"`
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
	Result T      `json:"result"`
	Cursor string `json:"cursor"`
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
	Reward           string              `json:"reward"`
	Events           []EventResult       `json:"events"`
	Oracles          []OracleResult      `json:"oracles"`
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
	Script        string                `json:"script"`
	Series        []TokenSeriesResult   `json:"series"`
	CarbonID      string                `json:"carbonId"`
	Metadata      []TokenPropertyResult `json:"metadata"`
	TokenSchemas  *TokenSchemasResult   `json:"tokenSchemas"`
	External      []TokenExternalResult `json:"external"`
	Price         []TokenPriceResult    `json:"price"`
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
	BurnedSupply   string                `json:"burnedSupply"`
	Mode           string                `json:"mode"`
	Script         string                `json:"script"`
	Methods        []ABIMethodResult     `json:"methods"`
	Metadata       []TokenPropertyResult `json:"metadata"`
}

// TokenPropertyResult describes a token metadata key/value pair.
type TokenPropertyResult struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
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
	ID               string                `json:"ID"`
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
	Result  string         `json:"result"`
	Results []string       `json:"results"`
	Oracles []OracleResult `json:"oracles"`
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
	Name          string   `json:"name"`
	Hash          string   `json:"hash"`
	Time          uint     `json:"time"`
	Size          uint     `json:"size"`
	Encryption    string   `json:"encryption"`
	BlockCount    int      `json:"blockCount"`
	MissingBlocks []int    `json:"missingBlocks"`
	Owners        []string `json:"owners"`
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
	Methods []ABIMethodResult `json:"methods"`
	Events  []ABIEventResult  `json:"events"`
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
