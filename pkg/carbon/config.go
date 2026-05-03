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

// GasConfig stores Carbon gas, data and fee configuration.
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
}
