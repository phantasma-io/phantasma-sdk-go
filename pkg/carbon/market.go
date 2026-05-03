package carbon

// ListingType identifies a Carbon market listing kind.
type ListingType byte

const (
	// ListingTypeFixedPrice is a fixed-price market listing.
	ListingTypeFixedPrice ListingType = 0
)

// MarketContractMethod identifies a method in the Carbon market module.
type MarketContractMethod uint32

// Market module method ids.
const (
	// MarketMethodSellToken lists a token by numeric ids.
	MarketMethodSellToken MarketContractMethod = iota
	MarketMethodSellTokenByID
	MarketMethodCancelSale
	MarketMethodCancelSaleByID
	MarketMethodBuyToken
	MarketMethodBuyTokenByID
	MarketMethodGetTokenListingCount
	MarketMethodGetTokenListingInfo
	MarketMethodGetTokenListingInfoByID
)

// TokenListing stores one Carbon market listing.
type TokenListing struct {
	Type         ListingType
	Seller       Bytes32
	QuoteTokenID uint64
	Price        IntX
	StartDate    int64
	EndDate      int64
}

// WriteCarbon writes the token listing to w.
func (l *TokenListing) WriteCarbon(w *Writer) {
	w.Write1(byte(l.Type))
	w.Write32(l.Seller)
	w.Write8U(l.QuoteTokenID)
	l.Price.WriteCarbon(w)
	w.Write8(l.StartDate)
	w.Write8(l.EndDate)
}

// ReadCarbon reads the token listing from r.
func (l *TokenListing) ReadCarbon(r *Reader) {
	l.Type = ListingType(r.Read1())
	l.Seller = r.Read32()
	l.QuoteTokenID = r.Read8U()
	l.Price.ReadCarbon(r)
	l.StartDate = r.Read8()
	l.EndDate = r.Read8()
}

// MarketConfigFlags describe Carbon market policy flags.
type MarketConfigFlags uint32

const (
	// MarketConfigFlagsNone disables market policy flags.
	MarketConfigFlagsNone MarketConfigFlags = 0
	// MarketConfigFlagsPriceRequired requires sale prices to be configured.
	MarketConfigFlagsPriceRequired MarketConfigFlags = 1 << 0
	// MarketConfigFlagsEnforceRoyalties applies royalty checks to listings.
	MarketConfigFlagsEnforceRoyalties MarketConfigFlags = 1 << 1
	// MarketConfigFlagsCanCancelEarly allows cancelling listings before their start time.
	MarketConfigFlagsCanCancelEarly MarketConfigFlags = 1 << 2
	// MarketConfigFlagsCanPurchaseLate allows purchases during the late-purchase window.
	MarketConfigFlagsCanPurchaseLate MarketConfigFlags = 1 << 3
)

const (
	// MarketMinimumListingTimeMS is the on-chain default minimum listing time.
	MarketMinimumListingTimeMS uint64 = 1000
	// MarketMaximumListingTimeMS is the on-chain default maximum listing time.
	MarketMaximumListingTimeMS uint64 = 1000 * 60 * 60 * 24 * 90
	// MarketDelistingGraceMS is the on-chain default delisting grace period.
	MarketDelistingGraceMS uint64 = 1000 * 60 * 60 * 24
	// MarketRoyaltyOnePercent is the royalty percent scale unit.
	MarketRoyaltyOnePercent uint64 = 10_000_000
	// MarketRoyaltyHundredPercent is the royalty percent scale maximum.
	MarketRoyaltyHundredPercent uint64 = 100 * MarketRoyaltyOnePercent
)

// MarketConfig stores Carbon market module configuration.
type MarketConfig struct {
	MinimumListingTime uint64
	MaximumListingTime uint64
	DelistingGrace     uint64
	Flags              MarketConfigFlags
}

// DefaultMarketConfig returns the SDK copy of the on-chain default market config.
func DefaultMarketConfig() MarketConfig {
	return MarketConfig{
		MinimumListingTime: MarketMinimumListingTimeMS,
		MaximumListingTime: MarketMaximumListingTimeMS,
		DelistingGrace:     MarketDelistingGraceMS,
		Flags:              MarketConfigFlagsPriceRequired | MarketConfigFlagsEnforceRoyalties,
	}
}

// WriteCarbon writes the market config to w.
func (c *MarketConfig) WriteCarbon(w *Writer) {
	w.Write8U(c.MinimumListingTime)
	w.Write8U(c.MaximumListingTime)
	w.Write8U(c.DelistingGrace)
	w.Write4U(uint32(c.Flags))
}

// ReadCarbon reads the market config from r.
func (c *MarketConfig) ReadCarbon(r *Reader) {
	c.MinimumListingTime = r.Read8U()
	c.MaximumListingTime = r.Read8U()
	c.DelistingGrace = r.Read8U()
	c.Flags = MarketConfigFlags(r.Read4U())
}

// MarketSellTokenArgs stores the Carbon market sell-token call arguments.
type MarketSellTokenArgs struct {
	From         Bytes32
	TokenID      uint64
	InstanceID   uint64
	QuoteTokenID uint64
	Price        IntX
	EndDate      int64
}

// WriteCarbon writes sell-token arguments to w.
func (a *MarketSellTokenArgs) WriteCarbon(w *Writer) {
	w.Write32(a.From)
	w.Write8U(a.TokenID)
	w.Write8U(a.InstanceID)
	w.Write8U(a.QuoteTokenID)
	a.Price.WriteCarbon(w)
	w.Write8(a.EndDate)
}

// ReadCarbon reads sell-token arguments from r.
func (a *MarketSellTokenArgs) ReadCarbon(r *Reader) {
	a.From = r.Read32()
	a.TokenID = r.Read8U()
	a.InstanceID = r.Read8U()
	a.QuoteTokenID = r.Read8U()
	a.Price.ReadCarbon(r)
	a.EndDate = r.Read8()
}

// MarketSellTokenByIDArgs stores symbol-based market sell-token call arguments.
type MarketSellTokenByIDArgs struct {
	From        Bytes32
	Symbol      SmallString
	InstanceID  VMDynamicVariable
	QuoteSymbol SmallString
	Price       IntX
	EndDate     int64
}

// WriteCarbon writes symbol-based sell-token arguments to w.
func (a *MarketSellTokenByIDArgs) WriteCarbon(w *Writer) {
	w.Write32(a.From)
	a.Symbol.WriteCarbon(w)
	a.InstanceID.WriteCarbon(w)
	a.QuoteSymbol.WriteCarbon(w)
	a.Price.WriteCarbon(w)
	w.Write8(a.EndDate)
}

// ReadCarbon reads symbol-based sell-token arguments from r.
func (a *MarketSellTokenByIDArgs) ReadCarbon(r *Reader) {
	a.From = r.Read32()
	a.Symbol.ReadCarbon(r)
	a.InstanceID.ReadCarbon(r)
	a.QuoteSymbol.ReadCarbon(r)
	a.Price.ReadCarbon(r)
	a.EndDate = r.Read8()
}

// MarketCancelSaleArgs stores numeric market cancel-sale arguments.
type MarketCancelSaleArgs struct {
	TokenID    uint64
	InstanceID uint64
}

// WriteCarbon writes cancel-sale arguments to w.
func (a *MarketCancelSaleArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
	w.Write8U(a.InstanceID)
}

// ReadCarbon reads cancel-sale arguments from r.
func (a *MarketCancelSaleArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
	a.InstanceID = r.Read8U()
}

// MarketCancelSaleByIDArgs stores symbol-based market cancel-sale arguments.
type MarketCancelSaleByIDArgs struct {
	Symbol     SmallString
	InstanceID VMDynamicVariable
}

// WriteCarbon writes symbol-based cancel-sale arguments to w.
func (a *MarketCancelSaleByIDArgs) WriteCarbon(w *Writer) {
	a.Symbol.WriteCarbon(w)
	a.InstanceID.WriteCarbon(w)
}

// ReadCarbon reads symbol-based cancel-sale arguments from r.
func (a *MarketCancelSaleByIDArgs) ReadCarbon(r *Reader) {
	a.Symbol.ReadCarbon(r)
	a.InstanceID.ReadCarbon(r)
}

// MarketBuyTokenArgs stores numeric market buy-token arguments.
type MarketBuyTokenArgs struct {
	From       Bytes32
	TokenID    uint64
	InstanceID uint64
}

// WriteCarbon writes buy-token arguments to w.
func (a *MarketBuyTokenArgs) WriteCarbon(w *Writer) {
	w.Write32(a.From)
	w.Write8U(a.TokenID)
	w.Write8U(a.InstanceID)
}

// ReadCarbon reads buy-token arguments from r.
func (a *MarketBuyTokenArgs) ReadCarbon(r *Reader) {
	a.From = r.Read32()
	a.TokenID = r.Read8U()
	a.InstanceID = r.Read8U()
}

// MarketBuyTokenByIDArgs stores symbol-based market buy-token arguments.
type MarketBuyTokenByIDArgs struct {
	From       Bytes32
	Symbol     SmallString
	InstanceID VMDynamicVariable
}

// WriteCarbon writes symbol-based buy-token arguments to w.
func (a *MarketBuyTokenByIDArgs) WriteCarbon(w *Writer) {
	w.Write32(a.From)
	a.Symbol.WriteCarbon(w)
	a.InstanceID.WriteCarbon(w)
}

// ReadCarbon reads symbol-based buy-token arguments from r.
func (a *MarketBuyTokenByIDArgs) ReadCarbon(r *Reader) {
	a.From = r.Read32()
	a.Symbol.ReadCarbon(r)
	a.InstanceID.ReadCarbon(r)
}

// MarketGetTokenListingCountArgs stores market listing-count arguments.
type MarketGetTokenListingCountArgs struct {
	TokenID uint64
}

// WriteCarbon writes listing-count arguments to w.
func (a *MarketGetTokenListingCountArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
}

// ReadCarbon reads listing-count arguments from r.
func (a *MarketGetTokenListingCountArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
}

// MarketGetTokenListingInfoArgs stores numeric listing-info arguments.
type MarketGetTokenListingInfoArgs struct {
	TokenID    uint64
	InstanceID uint64
}

// WriteCarbon writes listing-info arguments to w.
func (a *MarketGetTokenListingInfoArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
	w.Write8U(a.InstanceID)
}

// ReadCarbon reads listing-info arguments from r.
func (a *MarketGetTokenListingInfoArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
	a.InstanceID = r.Read8U()
}

// MarketGetTokenListingInfoByIDArgs stores symbol-based listing-info arguments.
type MarketGetTokenListingInfoByIDArgs struct {
	Symbol     SmallString
	InstanceID VMDynamicVariable
}

// WriteCarbon writes symbol-based listing-info arguments to w.
func (a *MarketGetTokenListingInfoByIDArgs) WriteCarbon(w *Writer) {
	a.Symbol.WriteCarbon(w)
	a.InstanceID.WriteCarbon(w)
}

// ReadCarbon reads symbol-based listing-info arguments from r.
func (a *MarketGetTokenListingInfoByIDArgs) ReadCarbon(r *Reader) {
	a.Symbol.ReadCarbon(r)
	a.InstanceID.ReadCarbon(r)
}
