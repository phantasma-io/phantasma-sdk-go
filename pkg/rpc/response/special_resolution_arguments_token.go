package response

// Arguments of the token module calls a special resolution can carry. Shapes that repeat across
// methods share one type on purpose: a query by token id looks the same whichever query it is.
//
// The token identity pair is embedded rather than repeated, which mirrors the inheritance of the
// C# reference models and keeps the wire order (the fields of the concrete call first, the token
// identity last). Embedding also promotes the interface marker, so every shape that embeds
// TokenReferenceArguments is a SpecialResolutionArguments without repeating the method.

// TokenReferenceArguments is a token identity: the resolved symbol plus the numeric id it was
// resolved from. It is also the arguments of the plain token queries (GetTokenInfo,
// GetTokenSupply, ApplyInflation, GetNextTokenInflation).
type TokenReferenceArguments struct {
	Token   string `json:"token"`
	TokenID string `json:"tokenId"`
}

func (TokenReferenceArguments) isSpecialResolutionArguments() {}

// TokenSeriesReferenceArguments addresses one series of a token, shared by DeleteTokenSeries,
// GetSeriesInfo and GetSeriesSupply.
type TokenSeriesReferenceArguments struct {
	SeriesID string `json:"seriesId"`
	TokenReferenceArguments
}

// SymbolArguments is a single symbol argument, shared by GetTokenInfoBySymbol and
// GetTokenIdBySymbol.
type SymbolArguments struct {
	Symbol string `json:"symbol"`
}

func (SymbolArguments) isSpecialResolutionArguments() {}

// TransferFungibleArguments are the arguments of token.TransferFungible.
type TransferFungibleArguments struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	TokenReferenceArguments
}

// TransferNonFungibleArguments are the arguments of token.TransferNonFungible.
type TransferNonFungibleArguments struct {
	From        string   `json:"from"`
	To          string   `json:"to"`
	InstanceIDs []string `json:"instanceIds"`
	TokenReferenceArguments
}

// MintFungibleArguments are the arguments of token.MintFungible.
type MintFungibleArguments struct {
	To     string `json:"to"`
	Amount string `json:"amount"`
	TokenReferenceArguments
}

// BurnFungibleArguments are the arguments of token.BurnFungible.
type BurnFungibleArguments struct {
	From   string `json:"from"`
	Amount string `json:"amount"`
	TokenReferenceArguments
}

// BalanceArguments are the arguments of token.GetBalance.
type BalanceArguments struct {
	Address string `json:"address"`
	TokenReferenceArguments
}

// CreateTokenArguments are the arguments of token.CreateToken.
type CreateTokenArguments struct {
	Symbol    string `json:"symbol"`
	Owner     string `json:"owner"`
	MaxSupply string `json:"maxSupply"`
	Decimals  string `json:"decimals"`
	Flags     string `json:"flags"`
	// Metadata holds the decoded metadata fields; absent when the token carries none.
	Metadata map[string]VMValue `json:"metadata,omitempty"`
	// TokenSchemas is the NFT schema blob as hex; absent for fungible tokens.
	TokenSchemas *string `json:"tokenSchemas,omitempty"`
}

func (CreateTokenArguments) isSpecialResolutionArguments() {}

// TokenSeriesArguments is a series definition, as carried by token.CreateTokenSeries.
type TokenSeriesArguments struct {
	Owner     string `json:"owner"`
	MaxMint   string `json:"maxMint"`
	MaxSupply string `json:"maxSupply"`
	// Metadata holds the decoded series metadata; absent when the token declares no schema for it.
	Metadata map[string]VMValue `json:"metadata,omitempty"`
	// SeriesID is the Phantasma series id taken from the decoded metadata, when the schema carries
	// one.
	SeriesID *string `json:"seriesId,omitempty"`
	// MetadataRaw is the metadata blob as hex, reported instead of Metadata when it cannot be
	// decoded.
	MetadataRaw *string `json:"metadataRaw,omitempty"`
	TokenReferenceArguments
}

// CreateMintedTokenSeriesArguments are the arguments of token.CreateMintedTokenSeries.
type CreateMintedTokenSeriesArguments struct {
	Recipient string   `json:"recipient"`
	Roms      []string `json:"roms"`
	Rams      []string `json:"rams"`
	TokenSeriesArguments
}

// NFTMint is one NFT to mint, addressed by the carbon series id.
type NFTMint struct {
	SeriesID string `json:"seriesId"`
	Rom      string `json:"rom"`
	Ram      string `json:"ram"`
}

// PhantasmaNFTMint is one NFT to mint, addressed by the 32-byte Phantasma series id.
type PhantasmaNFTMint struct {
	PhantasmaSeriesID string `json:"phantasmaSeriesId"`
	Rom               string `json:"rom"`
	Ram               string `json:"ram"`
}

// MintNonFungibleArguments are the arguments of token.MintNonFungible.
type MintNonFungibleArguments struct {
	Owner  string    `json:"owner"`
	Tokens []NFTMint `json:"tokens"`
	TokenReferenceArguments
}

// MintPhantasmaNonFungibleArguments are the arguments of token.MintPhantasmaNonFungible.
type MintPhantasmaNonFungibleArguments struct {
	Owner  string             `json:"owner"`
	Tokens []PhantasmaNFTMint `json:"tokens"`
	TokenReferenceArguments
}

// BurnNonFungibleArguments are the arguments of token.BurnNonFungible.
type BurnNonFungibleArguments struct {
	Address     string   `json:"address"`
	InstanceIDs []string `json:"instanceIds"`
	TokenReferenceArguments
}

// NonFungibleInfoArguments are the arguments of token.GetNonFungibleInfo.
type NonFungibleInfoArguments struct {
	InstanceID string `json:"instanceId"`
	GetSchemas string `json:"getSchemas"`
	TokenReferenceArguments
}

// NonFungibleInfoByRomIDArguments are the arguments of token.GetNonFungibleInfoByRomId.
type NonFungibleInfoByRomIDArguments struct {
	RomID      string `json:"romId"`
	GetSchemas string `json:"getSchemas"`
	TokenReferenceArguments
}

// SeriesInfoByMetaIDArguments are the arguments of token.GetSeriesInfoByMetaId.
type SeriesInfoByMetaIDArguments struct {
	RomID string `json:"romId"`
	TokenReferenceArguments
}

// TokensConfigArguments are the arguments of token.SetTokensConfig.
type TokensConfigArguments struct {
	Flags string `json:"flags"`
	// FlagsNames lists the names of the flags that are set, including a Reserved0xNN entry for
	// unknown bits.
	FlagsNames []string `json:"flagsNames"`
}

func (TokensConfigArguments) isSpecialResolutionArguments() {}

// UpdateTokenMetadataArguments are the arguments of token.UpdateTokenMetadata.
type UpdateTokenMetadataArguments struct {
	Metadata map[string]VMValue `json:"metadata,omitempty"`
	TokenReferenceArguments
}

// UpdateSeriesMetadataArguments are the arguments of token.UpdateSeriesMetadata.
type UpdateSeriesMetadataArguments struct {
	SeriesID string `json:"seriesId"`
	// Metadata is the metadata blob as hex: this call carries it unschematized.
	Metadata string `json:"metadata"`
	TokenReferenceArguments
}
