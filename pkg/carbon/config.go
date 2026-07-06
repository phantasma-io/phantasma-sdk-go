package carbon

// ChainConfig stores Carbon chain timing and admission configuration.
type ChainConfig struct {
	Version         byte
	Reserved1       byte
	Reserved2       byte
	Reserved3       byte
	AllowedTxTypes  uint32
	ExpiryWindow    uint32
	BlockRateTarget uint32
}

// WriteCarbon writes the chain config to w.
func (c *ChainConfig) WriteCarbon(w *Writer) {
	w.Write1(c.Version)
	w.Write1(c.Reserved1)
	w.Write1(c.Reserved2)
	w.Write1(c.Reserved3)
	w.Write4U(c.AllowedTxTypes)
	w.Write4U(c.ExpiryWindow)
	w.Write4U(c.BlockRateTarget)
}

// ReadCarbon reads the chain config from r.
func (c *ChainConfig) ReadCarbon(r *Reader) {
	c.Version = r.Read1()
	c.Reserved1 = r.Read1()
	c.Reserved2 = r.Read1()
	c.Reserved3 = r.Read1()
	c.AllowedTxTypes = r.Read4U()
	c.ExpiryWindow = r.Read4U()
	c.BlockRateTarget = r.Read4U()
}

// GasConfig stores Carbon gas, data and fee configuration. The gas-model-v2 extension fields
// serialize only for Version >= 1, mirroring the node's data_blockchain.h wire format exactly:
// the version-0 byte image is frozen forever for historical replay, and a version>=1 image
// truncated to the v0 length fails to parse (the tail read panics on end of stream).
type GasConfig struct {
	Version                 byte
	MaxNameLength           byte
	MaxTokenSymbolLength    byte
	FeeShift                byte
	MaxStructureSize        uint32
	FeeMultiplier           uint64
	GasTokenID              uint64
	DataTokenID             uint64
	MinimumGasOffer         uint64
	DataEscrowPerRow        uint64
	GasFeeTransfer          uint64
	GasFeeQuery             uint64
	GasFeeCreateTokenBase   uint64
	GasFeeCreateTokenSymbol uint64
	GasFeeCreateTokenSeries uint64
	GasFeePerByte           uint64
	GasFeeRegisterName      uint64
	GasBurnRatioMul         uint64
	GasBurnRatioShift       byte

	// Gas-model-v2 extension (Version >= 1 only).

	// MinimumGasBill floors every settled gas bill (kcal-base). 0 = no floor (v1-equivalent).
	MinimumGasBill uint64
	// GasProducerRatioMul/Shift split a bill portion to the block producer; same mul/shift
	// fixed-point form as the burn ratio, 0/0 = v1-equivalent.
	GasProducerRatioMul   uint64
	GasProducerRatioShift byte
	// GasDappRatioMul/Shift split a bill portion to the tx gasTarget dapp address.
	GasDappRatioMul   uint64
	GasDappRatioShift byte
	// Policy fees are product-decision prices in kcal-base, charged directly with no fee
	// multiplier stage. Under v2 they replace the unit-priced GasFeeCreateToken* and
	// GasFeeRegisterName fields, which stay serialized for version-0 replay.
	PolicyFeeCreateTokenBase uint64
	// PolicyFeeCreateTokenSymbol is halved per symbol char after the first (v1 rule kept).
	PolicyFeeCreateTokenSymbol uint64
	PolicyFeeCreateTokenSeries uint64
	// PolicyFeeRegisterName is shifted right by (nameLength-1) like the v1 field (v1 rule kept).
	PolicyFeeRegisterName uint64
	// LegacyDataEscrowPerRow is the frozen pre-flip DataEscrowPerRow: storage rows existing
	// before the v2 flip refund at this price (exactly what they escrowed under v1). Immutable
	// after the flip.
	LegacyDataEscrowPerRow uint64
}

// HasGasModelV2 reports whether this config activates the gas-model-v2 billing rules (config
// version >= 1). The gas model is gated by the config version, not by a chain feature level.
func (c *GasConfig) HasGasModelV2() bool {
	return c.Version >= 1
}

// WriteCarbon writes the gas config to w.
func (c *GasConfig) WriteCarbon(w *Writer) {
	w.Write1(c.Version)
	w.Write1(c.MaxNameLength)
	w.Write1(c.MaxTokenSymbolLength)
	w.Write1(c.FeeShift)
	w.Write4U(c.MaxStructureSize)
	w.Write8U(c.FeeMultiplier)
	w.Write8U(c.GasTokenID)
	w.Write8U(c.DataTokenID)
	w.Write8U(c.MinimumGasOffer)
	w.Write8U(c.DataEscrowPerRow)
	w.Write8U(c.GasFeeTransfer)
	w.Write8U(c.GasFeeQuery)
	w.Write8U(c.GasFeeCreateTokenBase)
	w.Write8U(c.GasFeeCreateTokenSymbol)
	w.Write8U(c.GasFeeCreateTokenSeries)
	w.Write8U(c.GasFeePerByte)
	w.Write8U(c.GasFeeRegisterName)
	w.Write8U(c.GasBurnRatioMul)
	w.Write1(c.GasBurnRatioShift)
	if c.Version == 0 {
		// Version-0 wire image must stay byte-identical to the pre-v2 layout.
		return
	}
	w.Write8U(c.MinimumGasBill)
	w.Write8U(c.GasProducerRatioMul)
	w.Write1(c.GasProducerRatioShift)
	w.Write8U(c.GasDappRatioMul)
	w.Write1(c.GasDappRatioShift)
	w.Write8U(c.PolicyFeeCreateTokenBase)
	w.Write8U(c.PolicyFeeCreateTokenSymbol)
	w.Write8U(c.PolicyFeeCreateTokenSeries)
	w.Write8U(c.PolicyFeeRegisterName)
	w.Write8U(c.LegacyDataEscrowPerRow)
}

// ReadCarbon reads the gas config from r.
func (c *GasConfig) ReadCarbon(r *Reader) {
	c.Version = r.Read1()
	c.MaxNameLength = r.Read1()
	c.MaxTokenSymbolLength = r.Read1()
	c.FeeShift = r.Read1()
	c.MaxStructureSize = r.Read4U()
	c.FeeMultiplier = r.Read8U()
	c.GasTokenID = r.Read8U()
	c.DataTokenID = r.Read8U()
	c.MinimumGasOffer = r.Read8U()
	c.DataEscrowPerRow = r.Read8U()
	c.GasFeeTransfer = r.Read8U()
	c.GasFeeQuery = r.Read8U()
	c.GasFeeCreateTokenBase = r.Read8U()
	c.GasFeeCreateTokenSymbol = r.Read8U()
	c.GasFeeCreateTokenSeries = r.Read8U()
	c.GasFeePerByte = r.Read8U()
	c.GasFeeRegisterName = r.Read8U()
	c.GasBurnRatioMul = r.Read8U()
	c.GasBurnRatioShift = r.Read1()
	if c.Version == 0 {
		// Version-0 rows carry no v2 tail; zero it so a reused instance never leaks stale values.
		c.MinimumGasBill = 0
		c.GasProducerRatioMul = 0
		c.GasProducerRatioShift = 0
		c.GasDappRatioMul = 0
		c.GasDappRatioShift = 0
		c.PolicyFeeCreateTokenBase = 0
		c.PolicyFeeCreateTokenSymbol = 0
		c.PolicyFeeCreateTokenSeries = 0
		c.PolicyFeeRegisterName = 0
		c.LegacyDataEscrowPerRow = 0
		return
	}
	// Version >= 1: the tail is mandatory; a truncated image panics (end of stream).
	c.MinimumGasBill = r.Read8U()
	c.GasProducerRatioMul = r.Read8U()
	c.GasProducerRatioShift = r.Read1()
	c.GasDappRatioMul = r.Read8U()
	c.GasDappRatioShift = r.Read1()
	c.PolicyFeeCreateTokenBase = r.Read8U()
	c.PolicyFeeCreateTokenSymbol = r.Read8U()
	c.PolicyFeeCreateTokenSeries = r.Read8U()
	c.PolicyFeeRegisterName = r.Read8U()
	c.LegacyDataEscrowPerRow = r.Read8U()
}
