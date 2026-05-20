package carbon

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
)

// ModuleID identifies a Carbon chain module.
type ModuleID uint32

// Carbon module ids.
const (
	// ModuleIDGovernance is the governance module id.
	ModuleIDGovernance  ModuleID = 0
	ModuleIDToken       ModuleID = 1
	ModuleIDPhantasma   ModuleID = 2
	ModuleIDPhantasmaVM          = ModuleIDPhantasma
	ModuleIDOrg         ModuleID = 3
	ModuleIDMarket      ModuleID = 4
	ModuleIDInternal    ModuleID = 0xffffffff
)

// TokenContractMethod identifies a method in the Carbon token module.
type TokenContractMethod uint32

// Token module method ids.
const (
	// TokenMethodTransferFungible transfers fungible tokens.
	TokenMethodTransferFungible TokenContractMethod = iota
	TokenMethodTransferNonFungible
	TokenMethodCreateToken
	TokenMethodMintFungible
	TokenMethodBurnFungible
	TokenMethodGetBalance
	TokenMethodCreateTokenSeries
	TokenMethodDeleteTokenSeries
	TokenMethodMintNonFungible
	TokenMethodBurnNonFungible
	TokenMethodGetInstances
	TokenMethodGetNonFungibleInfo
	TokenMethodGetNonFungibleInfoByRomID
	TokenMethodGetSeriesInfo
	TokenMethodGetSeriesInfoByMetaID
	TokenMethodGetTokenInfo
	TokenMethodGetTokenInfoBySymbol
	TokenMethodGetTokenSupply
	TokenMethodGetSeriesSupply
	TokenMethodGetTokenIDBySymbol
	TokenMethodGetBalances
	TokenMethodCreateMintedTokenSeries
	TokenMethodApplyInflation
	TokenMethodUpdateTokenMetadata
	TokenMethodGetNextTokenInflation
	TokenMethodSetTokensConfig
	TokenMethodUpdateSeriesMetadata
	TokenMethodMintPhantasmaNonFungible
)

// TokenFlags describe Carbon token capabilities.
type TokenFlags byte

// Token capability flags.
const (
	// TokenFlagsNone disables token capability flags.
	TokenFlagsNone        TokenFlags = 0
	TokenFlagsBigFungible TokenFlags = 1 << 0
	TokenFlagsNonFungible TokenFlags = 1 << 1
)

// TokensConfigFlags describe token module policy requirements.
type TokensConfigFlags byte

// Token module configuration flags.
const (
	// TokensConfigFlagsNone disables token config flags.
	TokensConfigFlagsNone                       TokensConfigFlags = 0
	TokensConfigFlagsRequireMetadata            TokensConfigFlags = 1 << 0
	TokensConfigFlagsRequireSymbol              TokensConfigFlags = 1 << 1
	TokensConfigFlagsRequireNFTMetaID           TokensConfigFlags = 1 << 2
	TokensConfigFlagsRequireNFTStandard         TokensConfigFlags = 1 << 3
	TokensConfigFlagsAllowExplicitNFTMetaIDMint TokensConfigFlags = 1 << 4
)

// TokensConfig stores token module configuration flags.
type TokensConfig struct {
	Flags TokensConfigFlags
}

// WriteCarbon writes token config to w.
func (c *TokensConfig) WriteCarbon(w *Writer) {
	w.Write1(byte(c.Flags))
}

// ReadCarbon reads token config from r.
func (c *TokensConfig) ReadCarbon(r *Reader) {
	c.Flags = TokensConfigFlags(r.Read1())
}

// TokenInfo describes a token definition submitted to the token module.
type TokenInfo struct {
	MaxSupply    IntX
	Flags        TokenFlags
	Decimals     byte
	Owner        Bytes32
	Symbol       SmallString
	Metadata     []byte
	TokenSchemas []byte
}

// WriteCarbon writes token info to w.
func (t *TokenInfo) WriteCarbon(w *Writer) {
	t.MaxSupply.WriteCarbon(w)
	w.Write1(byte(t.Flags))
	w.Write1(t.Decimals)
	w.Write32(t.Owner)
	t.Symbol.WriteCarbon(w)
	w.WriteByteArray(t.Metadata)
	if t.Flags&TokenFlagsNonFungible != 0 {
		w.WriteByteArray(t.TokenSchemas)
	}
}

// ReadCarbon reads token info from r.
func (t *TokenInfo) ReadCarbon(r *Reader) {
	t.MaxSupply.ReadCarbon(r)
	t.Flags = TokenFlags(r.Read1())
	t.Decimals = r.Read1()
	t.Owner = r.Read32()
	t.Symbol.ReadCarbon(r)
	t.Metadata = r.ReadByteArray()
	if t.Flags&TokenFlagsNonFungible != 0 {
		t.TokenSchemas = r.ReadByteArray()
	} else {
		t.TokenSchemas = nil
	}
}

// SeriesInfo describes one NFT series definition.
type SeriesInfo struct {
	MaxMint   uint32
	MaxSupply uint32
	Owner     Bytes32
	Metadata  []byte
	ROM       VMStructSchema
	RAM       VMStructSchema
}

// WriteCarbon writes series info to w.
func (s *SeriesInfo) WriteCarbon(w *Writer) {
	w.Write4U(s.MaxMint)
	w.Write4U(s.MaxSupply)
	w.Write32(s.Owner)
	w.WriteByteArray(s.Metadata)
	s.ROM.WriteCarbon(w)
	s.RAM.WriteCarbon(w)
}

// ReadCarbon reads series info from r.
func (s *SeriesInfo) ReadCarbon(r *Reader) {
	s.MaxMint = r.Read4U()
	s.MaxSupply = r.Read4U()
	s.Owner = r.Read32()
	s.Metadata = r.ReadByteArray()
	s.ROM.ReadCarbon(r)
	s.RAM.ReadCarbon(r)
}

// TokenSchemas stores token, ROM and RAM metadata schemas.
type TokenSchemas struct {
	SeriesMetadata VMStructSchema
	ROM            VMStructSchema
	RAM            VMStructSchema
}

// WriteCarbon writes token schemas to w.
func (s *TokenSchemas) WriteCarbon(w *Writer) {
	s.SeriesMetadata.WriteCarbon(w)
	s.ROM.WriteCarbon(w)
	s.RAM.WriteCarbon(w)
}

// ReadCarbon reads token schemas from r.
func (s *TokenSchemas) ReadCarbon(r *Reader) {
	s.SeriesMetadata.ReadCarbon(r)
	s.ROM.ReadCarbon(r)
	s.RAM.ReadCarbon(r)
}

// FeeOptions controls base transaction gas calculation.
type FeeOptions struct {
	GasFeeBase    uint64
	FeeMultiplier uint64
}

// NewFeeOptions returns explicit gas fee options.
func NewFeeOptions(gasFeeBase, feeMultiplier uint64) FeeOptions {
	return FeeOptions{GasFeeBase: gasFeeBase, FeeMultiplier: feeMultiplier}
}

// DefaultFeeOptions returns SDK default gas fee options.
func DefaultFeeOptions() FeeOptions {
	return NewFeeOptions(10_000, 1_000)
}

// CalculateMaxGas calculates max gas using defaults for zero-valued fields.
func (f FeeOptions) CalculateMaxGas() uint64 {
	f = f.withDefaults()
	return f.GasFeeBase * f.FeeMultiplier
}

// CalculateMaxGasForCount calculates max gas for count-sensitive transactions.
func (f FeeOptions) CalculateMaxGasForCount(count uint64) (uint64, error) {
	f = f.withDefaults()
	return multiplyGasForCount(f.GasFeeBase, f.FeeMultiplier, count, "FeeOptions.CalculateMaxGasForCount")
}

func (f FeeOptions) withDefaults() FeeOptions {
	defaults := DefaultFeeOptions()
	if f.GasFeeBase == 0 {
		f.GasFeeBase = defaults.GasFeeBase
	}
	if f.FeeMultiplier == 0 {
		f.FeeMultiplier = defaults.FeeMultiplier
	}
	return f
}

// CreateTokenFeeOptions controls gas calculation for create-token transactions.
type CreateTokenFeeOptions struct {
	GasFeeBase              uint64
	FeeMultiplier           uint64
	GasFeeCreateTokenBase   uint64
	GasFeeCreateTokenSymbol uint64
}

// NewCreateTokenFeeOptions returns explicit create-token gas fee options.
func NewCreateTokenFeeOptions(gasFeeBase, feeMultiplier, createTokenBase, createTokenSymbol uint64) CreateTokenFeeOptions {
	return CreateTokenFeeOptions{
		GasFeeBase:              gasFeeBase,
		FeeMultiplier:           feeMultiplier,
		GasFeeCreateTokenBase:   createTokenBase,
		GasFeeCreateTokenSymbol: createTokenSymbol,
	}
}

// DefaultCreateTokenFeeOptions returns SDK default create-token fee options.
func DefaultCreateTokenFeeOptions() CreateTokenFeeOptions {
	return NewCreateTokenFeeOptions(10_000, 10_000, 10_000_000_000, 10_000_000_000)
}

// CalculateMaxGas calculates create-token max gas using defaults for zero-valued fields.
func (f CreateTokenFeeOptions) CalculateMaxGas(symbol SmallString) uint64 {
	f = f.withDefaults()
	shift := len(symbol.Bytes()) - 1
	if shift < 0 {
		shift = 0
	}
	symbolPart := f.GasFeeCreateTokenSymbol
	if shift < 64 {
		symbolPart >>= shift
	} else {
		symbolPart = 0
	}
	return (f.GasFeeBase + f.GasFeeCreateTokenBase + symbolPart) * f.FeeMultiplier
}

func (f CreateTokenFeeOptions) withDefaults() CreateTokenFeeOptions {
	defaults := DefaultCreateTokenFeeOptions()
	if f.GasFeeBase == 0 {
		f.GasFeeBase = defaults.GasFeeBase
	}
	if f.FeeMultiplier == 0 {
		f.FeeMultiplier = defaults.FeeMultiplier
	}
	if f.GasFeeCreateTokenBase == 0 {
		f.GasFeeCreateTokenBase = defaults.GasFeeCreateTokenBase
	}
	if f.GasFeeCreateTokenSymbol == 0 {
		f.GasFeeCreateTokenSymbol = defaults.GasFeeCreateTokenSymbol
	}
	return f
}

// CreateSeriesFeeOptions controls gas calculation for create-series transactions.
type CreateSeriesFeeOptions struct {
	GasFeeBase             uint64
	FeeMultiplier          uint64
	GasFeeCreateSeriesBase uint64
}

// NewCreateSeriesFeeOptions returns explicit create-series gas fee options.
func NewCreateSeriesFeeOptions(gasFeeBase, feeMultiplier, createSeriesBase uint64) CreateSeriesFeeOptions {
	return CreateSeriesFeeOptions{
		GasFeeBase:             gasFeeBase,
		FeeMultiplier:          feeMultiplier,
		GasFeeCreateSeriesBase: createSeriesBase,
	}
}

// DefaultCreateSeriesFeeOptions returns SDK default create-series fee options.
func DefaultCreateSeriesFeeOptions() CreateSeriesFeeOptions {
	return NewCreateSeriesFeeOptions(10_000, 10_000, 2_500_000_000)
}

// CalculateMaxGas calculates create-series max gas using defaults for zero-valued fields.
func (f CreateSeriesFeeOptions) CalculateMaxGas() uint64 {
	f = f.withDefaults()
	return (f.GasFeeBase + f.GasFeeCreateSeriesBase) * f.FeeMultiplier
}

func (f CreateSeriesFeeOptions) withDefaults() CreateSeriesFeeOptions {
	defaults := DefaultCreateSeriesFeeOptions()
	if f.GasFeeBase == 0 {
		f.GasFeeBase = defaults.GasFeeBase
	}
	if f.FeeMultiplier == 0 {
		f.FeeMultiplier = defaults.FeeMultiplier
	}
	if f.GasFeeCreateSeriesBase == 0 {
		f.GasFeeCreateSeriesBase = defaults.GasFeeCreateSeriesBase
	}
	return f
}

// MintNFTFeeOptions controls gas calculation for NFT mint transactions.
type MintNFTFeeOptions struct {
	GasFeeBase    uint64
	FeeMultiplier uint64
}

// NewMintNFTFeeOptions returns explicit NFT mint gas fee options.
func NewMintNFTFeeOptions(gasFeeBase, feeMultiplier uint64) MintNFTFeeOptions {
	return MintNFTFeeOptions{GasFeeBase: gasFeeBase, FeeMultiplier: feeMultiplier}
}

// DefaultMintNFTFeeOptions returns SDK default NFT mint fee options.
func DefaultMintNFTFeeOptions() MintNFTFeeOptions {
	defaults := DefaultFeeOptions()
	return NewMintNFTFeeOptions(defaults.GasFeeBase, defaults.FeeMultiplier)
}

// CalculateMaxGasForCount calculates NFT mint max gas for the requested token count.
func (f MintNFTFeeOptions) CalculateMaxGasForCount(count uint64) (uint64, error) {
	f = f.withDefaults()
	return multiplyGasForCount(f.GasFeeBase, f.FeeMultiplier, count, "MintNFTFeeOptions.CalculateMaxGasForCount")
}

func (f MintNFTFeeOptions) calculateMaxGasForSingle() uint64 {
	f = f.withDefaults()
	return f.GasFeeBase * f.FeeMultiplier
}

func (f MintNFTFeeOptions) withDefaults() MintNFTFeeOptions {
	defaults := DefaultMintNFTFeeOptions()
	if f.GasFeeBase == 0 {
		f.GasFeeBase = defaults.GasFeeBase
	}
	if f.FeeMultiplier == 0 {
		f.FeeMultiplier = defaults.FeeMultiplier
	}
	return f
}

func multiplyGasForCount(gasFeeBase, feeMultiplier, count uint64, methodName string) (uint64, error) {
	if count == 0 {
		return 0, fmt.Errorf("%s count must be positive", methodName)
	}

	max := ^uint64(0)
	if gasFeeBase != 0 && feeMultiplier > max/gasFeeBase {
		return 0, fmt.Errorf("%s overflow", methodName)
	}
	baseGas := gasFeeBase * feeMultiplier
	if baseGas != 0 && count > max/baseGas {
		return 0, fmt.Errorf("%s overflow", methodName)
	}
	return baseGas * count, nil
}

// BuildCreateTokenTx builds an unsigned Carbon create-token transaction.
func BuildCreateTokenTx(tokenInfo TokenInfo, creator Bytes32, fees CreateTokenFeeOptions, maxData uint64, expiry int64) TxMsg {
	return TxMsg{
		Type:    TxTypeCall,
		Expiry:  effectiveExpiry(expiry),
		MaxGas:  fees.CalculateMaxGas(tokenInfo.Symbol),
		MaxData: maxData,
		GasFrom: creator,
		Payload: MustSmallString(""),
		Msg: &TxMsgCall{
			ModuleID: uint32(ModuleIDToken),
			MethodID: uint32(TokenMethodCreateToken),
			Args:     Serialize(&tokenInfo),
		},
	}
}

// BuildCreateTokenTxAndSign builds, signs and serializes a create-token transaction.
func BuildCreateTokenTxAndSign(tokenInfo TokenInfo, signer cryptography.KeyPair, fees CreateTokenFeeOptions, maxData uint64, expiry int64) ([]byte, error) {
	creator, err := publicKeyFromSigner(signer)
	if err != nil {
		return nil, err
	}
	msg := BuildCreateTokenTx(tokenInfo, creator, fees, maxData, expiry)
	return SignAndSerializeTxMsg(msg, signer)
}

// BuildCreateTokenTxAndSignHex builds, signs and hex-encodes a create-token transaction.
func BuildCreateTokenTxAndSignHex(tokenInfo TokenInfo, signer cryptography.KeyPair, fees CreateTokenFeeOptions, maxData uint64, expiry int64) (string, error) {
	return signAndSerializeHex(BuildCreateTokenTxAndSign(tokenInfo, signer, fees, maxData, expiry))
}

// BuildCreateTokenSeriesTx builds an unsigned Carbon create-series transaction.
func BuildCreateTokenSeriesTx(tokenID uint64, seriesInfo SeriesInfo, creator Bytes32, fees CreateSeriesFeeOptions, maxData uint64, expiry int64) TxMsg {
	argsWriter := NewWriter()
	argsWriter.Write8U(tokenID)
	seriesInfo.WriteCarbon(argsWriter)

	return TxMsg{
		Type:    TxTypeCall,
		Expiry:  effectiveExpiry(expiry),
		MaxGas:  fees.CalculateMaxGas(),
		MaxData: maxData,
		GasFrom: creator,
		Payload: MustSmallString(""),
		Msg: &TxMsgCall{
			ModuleID: uint32(ModuleIDToken),
			MethodID: uint32(TokenMethodCreateTokenSeries),
			Args:     argsWriter.Bytes(),
		},
	}
}

// BuildCreateTokenSeriesTxAndSign builds, signs and serializes a create-series transaction.
func BuildCreateTokenSeriesTxAndSign(tokenID uint64, seriesInfo SeriesInfo, signer cryptography.KeyPair, fees CreateSeriesFeeOptions, maxData uint64, expiry int64) ([]byte, error) {
	creator, err := publicKeyFromSigner(signer)
	if err != nil {
		return nil, err
	}
	msg := BuildCreateTokenSeriesTx(tokenID, seriesInfo, creator, fees, maxData, expiry)
	return SignAndSerializeTxMsg(msg, signer)
}

// BuildCreateTokenSeriesTxAndSignHex builds, signs and hex-encodes a create-series transaction.
func BuildCreateTokenSeriesTxAndSignHex(tokenID uint64, seriesInfo SeriesInfo, signer cryptography.KeyPair, fees CreateSeriesFeeOptions, maxData uint64, expiry int64) (string, error) {
	return signAndSerializeHex(BuildCreateTokenSeriesTxAndSign(tokenID, seriesInfo, signer, fees, maxData, expiry))
}

// BuildMintNonFungibleTx builds an unsigned Carbon NFT mint transaction.
func BuildMintNonFungibleTx(tokenID uint64, seriesID uint32, sender Bytes32, receiver Bytes32, rom []byte, ram []byte, fees MintNFTFeeOptions, maxData uint64, expiry int64) TxMsg {
	return TxMsg{
		Type:    TxTypeMintNonFungible,
		Expiry:  effectiveExpiry(expiry),
		MaxGas:  fees.calculateMaxGasForSingle(),
		MaxData: maxData,
		GasFrom: sender,
		Payload: MustSmallString(""),
		Msg: &TxMsgMintNonFungible{
			TokenID:  tokenID,
			To:       receiver,
			SeriesID: seriesID,
			ROM:      rom,
			RAM:      ram,
		},
	}
}

// BuildMintNonFungibleTxAndSign builds, signs and serializes an NFT mint transaction.
func BuildMintNonFungibleTxAndSign(tokenID uint64, seriesID uint32, signer cryptography.KeyPair, receiver Bytes32, rom []byte, ram []byte, fees MintNFTFeeOptions, maxData uint64, expiry int64) ([]byte, error) {
	sender, err := publicKeyFromSigner(signer)
	if err != nil {
		return nil, err
	}
	msg := BuildMintNonFungibleTx(tokenID, seriesID, sender, receiver, rom, ram, fees, maxData, expiry)
	return SignAndSerializeTxMsg(msg, signer)
}

// BuildMintNonFungibleTxAndSignHex builds, signs and hex-encodes an NFT mint transaction.
func BuildMintNonFungibleTxAndSignHex(tokenID uint64, seriesID uint32, signer cryptography.KeyPair, receiver Bytes32, rom []byte, ram []byte, fees MintNFTFeeOptions, maxData uint64, expiry int64) (string, error) {
	return signAndSerializeHex(BuildMintNonFungibleTxAndSign(tokenID, seriesID, signer, receiver, rom, ram, fees, maxData, expiry))
}

// BuildMintPhantasmaNonFungibleTx builds an unsigned Phantasma NFT mint transaction.
func BuildMintPhantasmaNonFungibleTx(tokenID uint64, sender Bytes32, receiver Bytes32, tokens []PhantasmaNFTMintInfo, fees MintNFTFeeOptions, maxData uint64, expiry int64) (TxMsg, error) {
	maxGas, err := fees.CalculateMaxGasForCount(uint64(len(tokens)))
	if err != nil {
		return TxMsg{}, err
	}

	args := MintPhantasmaNonFungibleArgs{
		TokenID: tokenID,
		Address: receiver,
		Tokens:  tokens,
	}
	return TxMsg{
		Type:    TxTypeCall,
		Expiry:  effectiveExpiry(expiry),
		MaxGas:  maxGas,
		MaxData: maxData,
		GasFrom: sender,
		Payload: MustSmallString(""),
		Msg: &TxMsgCall{
			ModuleID: uint32(ModuleIDToken),
			MethodID: uint32(TokenMethodMintPhantasmaNonFungible),
			Args:     Serialize(&args),
		},
	}, nil
}

// BuildMintPhantasmaNonFungibleSingleTx builds an unsigned single-Phantasma-NFT mint transaction.
func BuildMintPhantasmaNonFungibleSingleTx(tokenID uint64, phantasmaSeriesID *big.Int, sender Bytes32, receiver Bytes32, publicRom []byte, ram []byte, fees MintNFTFeeOptions, maxData uint64, expiry int64) (TxMsg, error) {
	if err := requireBigInt("phantasmaSeriesID", phantasmaSeriesID); err != nil {
		return TxMsg{}, err
	}
	if ram == nil {
		ram = []byte{}
	}
	return BuildMintPhantasmaNonFungibleTx(
		tokenID,
		sender,
		receiver,
		[]PhantasmaNFTMintInfo{
			{
				PhantasmaSeriesID: NewIntX(phantasmaSeriesID),
				ROM:               publicRom,
				RAM:               ram,
			},
		},
		fees,
		maxData,
		expiry,
	)
}

// MustBuildMintPhantasmaNonFungibleSingleTx builds a single-Phantasma-NFT mint transaction and panics on validation errors.
func MustBuildMintPhantasmaNonFungibleSingleTx(tokenID uint64, phantasmaSeriesID *big.Int, sender Bytes32, receiver Bytes32, publicRom []byte, ram []byte, fees MintNFTFeeOptions, maxData uint64, expiry int64) TxMsg {
	out, err := BuildMintPhantasmaNonFungibleSingleTx(tokenID, phantasmaSeriesID, sender, receiver, publicRom, ram, fees, maxData, expiry)
	if err != nil {
		panic(err)
	}
	return out
}

// BuildMintPhantasmaNonFungibleSingleTxAndSign builds, signs and serializes a single-Phantasma-NFT mint transaction.
func BuildMintPhantasmaNonFungibleSingleTxAndSign(tokenID uint64, phantasmaSeriesID *big.Int, signer cryptography.KeyPair, receiver Bytes32, publicRom []byte, ram []byte, fees MintNFTFeeOptions, maxData uint64, expiry int64) ([]byte, error) {
	sender, err := publicKeyFromSigner(signer)
	if err != nil {
		return nil, err
	}
	msg, err := BuildMintPhantasmaNonFungibleSingleTx(tokenID, phantasmaSeriesID, sender, receiver, publicRom, ram, fees, maxData, expiry)
	if err != nil {
		return nil, err
	}
	return SignAndSerializeTxMsg(msg, signer)
}

// BuildMintPhantasmaNonFungibleSingleTxAndSignHex builds, signs and hex-encodes a single-Phantasma-NFT mint transaction.
func BuildMintPhantasmaNonFungibleSingleTxAndSignHex(tokenID uint64, phantasmaSeriesID *big.Int, signer cryptography.KeyPair, receiver Bytes32, publicRom []byte, ram []byte, fees MintNFTFeeOptions, maxData uint64, expiry int64) (string, error) {
	return signAndSerializeHex(BuildMintPhantasmaNonFungibleSingleTxAndSign(tokenID, phantasmaSeriesID, signer, receiver, publicRom, ram, fees, maxData, expiry))
}

// BuildMintPhantasmaNonFungibleTxAndSign builds, signs and serializes a Phantasma NFT mint transaction.
func BuildMintPhantasmaNonFungibleTxAndSign(tokenID uint64, signer cryptography.KeyPair, receiver Bytes32, tokens []PhantasmaNFTMintInfo, fees MintNFTFeeOptions, maxData uint64, expiry int64) ([]byte, error) {
	sender, err := publicKeyFromSigner(signer)
	if err != nil {
		return nil, err
	}
	msg, err := BuildMintPhantasmaNonFungibleTx(tokenID, sender, receiver, tokens, fees, maxData, expiry)
	if err != nil {
		return nil, err
	}
	return SignAndSerializeTxMsg(msg, signer)
}

// BuildMintPhantasmaNonFungibleTxAndSignHex builds, signs and hex-encodes a Phantasma NFT mint transaction.
func BuildMintPhantasmaNonFungibleTxAndSignHex(tokenID uint64, signer cryptography.KeyPair, receiver Bytes32, tokens []PhantasmaNFTMintInfo, fees MintNFTFeeOptions, maxData uint64, expiry int64) (string, error) {
	return signAndSerializeHex(BuildMintPhantasmaNonFungibleTxAndSign(tokenID, signer, receiver, tokens, fees, maxData, expiry))
}

// GetNFTAddress derives a Carbon NFT address from token and instance ids.
func GetNFTAddress(carbonTokenID uint64, instanceID uint64) Bytes32 {
	var address Bytes32
	address[15] = 1
	putUint64LE(address[16:24], carbonTokenID)
	putUint64LE(address[24:32], instanceID)
	return address
}

// UnpackNFTInstanceID splits a Carbon instance id into series id and mint number.
func UnpackNFTInstanceID(instanceID uint64) (seriesID uint32, mintNumber uint32) {
	return uint32(instanceID & 0xffffffff), uint32((instanceID >> 32) & 0xffffffff)
}

// ParseCreateTokenResult parses a create-token result into the Carbon token id.
func ParseCreateTokenResult(resultHex string) (uint64, error) {
	return parseCarbonResult(resultHex, func(r *Reader) (uint64, error) {
		return r.Read8U(), nil
	})
}

// MustParseCreateTokenResult parses a create-token result and panics on error.
func MustParseCreateTokenResult(resultHex string) uint64 {
	out, err := ParseCreateTokenResult(resultHex)
	if err != nil {
		panic(err)
	}
	return out
}

// ParseCreateTokenSeriesResult parses a create-series result into the Carbon series id.
func ParseCreateTokenSeriesResult(resultHex string) (uint32, error) {
	return parseCarbonResult(resultHex, func(r *Reader) (uint32, error) {
		return r.Read4U(), nil
	})
}

// MustParseCreateTokenSeriesResult parses a create-series result and panics on error.
func MustParseCreateTokenSeriesResult(resultHex string) uint32 {
	out, err := ParseCreateTokenSeriesResult(resultHex)
	if err != nil {
		panic(err)
	}
	return out
}

// ParseMintNonFungibleResult parses a Carbon NFT mint result into NFT addresses.
func ParseMintNonFungibleResult(carbonTokenID uint64, resultHex string) ([]Bytes32, error) {
	return parseCarbonResult(resultHex, func(r *Reader) ([]Bytes32, error) {
		count := r.ReadLengthFor(8)
		out := make([]Bytes32, count)
		for i := range out {
			out[i] = GetNFTAddress(carbonTokenID, r.Read8U())
		}
		return out, nil
	})
}

// MustParseMintNonFungibleResult parses a Carbon NFT mint result and panics on error.
func MustParseMintNonFungibleResult(carbonTokenID uint64, resultHex string) []Bytes32 {
	out, err := ParseMintNonFungibleResult(carbonTokenID, resultHex)
	if err != nil {
		panic(err)
	}
	return out
}

// ParseMintPhantasmaNonFungibleResult parses a Phantasma NFT mint result.
func ParseMintPhantasmaNonFungibleResult(resultHex string) ([]PhantasmaNFTMintResult, error) {
	return parseCarbonResult(resultHex, func(r *Reader) ([]PhantasmaNFTMintResult, error) {
		count := r.ReadLengthFor(40)
		out := make([]PhantasmaNFTMintResult, count)
		for i := range out {
			out[i].ReadCarbon(r)
		}
		return out, nil
	})
}

// MustParseMintPhantasmaNonFungibleResult parses a Phantasma NFT mint result and panics on error.
func MustParseMintPhantasmaNonFungibleResult(resultHex string) []PhantasmaNFTMintResult {
	out, err := ParseMintPhantasmaNonFungibleResult(resultHex)
	if err != nil {
		panic(err)
	}
	return out
}

// BuildTokenInfo validates inputs and builds a Carbon token definition.
func BuildTokenInfo(symbol string, maxSupply IntX, isNFT bool, decimals byte, owner Bytes32, metadata []byte, schemas []byte) (TokenInfo, error) {
	if err := CheckTokenSymbol(symbol); err != nil {
		return TokenInfo{}, err
	}
	if metadata == nil {
		return TokenInfo{}, fmt.Errorf("metadata is required for all tokens")
	}

	flags := TokenFlagsNone
	if isNFT {
		if !maxSupply.Is8ByteSafe() {
			return TokenInfo{}, fmt.Errorf("NFT maximum supply must fit into Int64")
		}
		if schemas == nil {
			return TokenInfo{}, fmt.Errorf("token schemas are required for NFTs")
		}
		flags = TokenFlagsNonFungible
	} else if maxSupply.BigInt().Sign() == 0 || !maxSupply.Is8ByteSafe() {
		flags = TokenFlagsBigFungible
	}

	symbolValue, err := NewSmallString(symbol)
	if err != nil {
		return TokenInfo{}, err
	}

	return TokenInfo{
		MaxSupply:    maxSupply,
		Flags:        flags,
		Decimals:     decimals,
		Owner:        owner,
		Symbol:       symbolValue,
		Metadata:     metadata,
		TokenSchemas: schemas,
	}, nil
}

// CheckTokenSymbol validates the Carbon token symbol format.
func CheckTokenSymbol(symbol string) error {
	if symbol == "" {
		return fmt.Errorf("symbol validation error: empty string is invalid")
	}
	if len(symbol) > 255 {
		return fmt.Errorf("symbol validation error: too long")
	}
	for i := 0; i < len(symbol); i++ {
		if symbol[i] < 'A' || symbol[i] > 'Z' {
			return fmt.Errorf("symbol validation error: only A-Z uppercase ASCII letters are allowed")
		}
	}
	return nil
}

func effectiveExpiry(expiry int64) int64 {
	if expiry != 0 {
		return expiry
	}
	return NowUnixMillis() + 60_000
}

func signAndSerializeHex(data []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

func parseCarbonResult[T any](resultHex string, read func(*Reader) (T, error)) (out T, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("carbon result parse failed: %v", recovered)
		}
	}()

	data, err := DecodeHex(resultHex)
	if err != nil {
		return out, err
	}
	r := NewReader(data)
	out, err = read(r)
	if err != nil {
		return out, err
	}
	r.AssertEOF()
	return out, nil
}

func putUint64LE(dst []byte, value uint64) {
	for i := 0; i < 8; i++ {
		dst[i] = byte(value >> (8 * i))
	}
}

func publicKeyFromSigner(signer cryptography.KeyPair) (Bytes32, error) {
	if signer == nil {
		return Bytes32{}, fmt.Errorf("key pair is required")
	}
	return Bytes32FromPublicKey(signer.PublicKey())
}
