package carbon

// NFTMintInfo describes one Carbon NFT mint payload.
type NFTMintInfo struct {
	SeriesID uint32
	ROM      []byte
	RAM      []byte
}

// WriteCarbon writes the NFT mint payload to w.
func (i *NFTMintInfo) WriteCarbon(w *Writer) {
	w.Write4U(i.SeriesID)
	w.WriteByteArray(i.ROM)
	w.WriteByteArray(i.RAM)
}

// ReadCarbon reads the NFT mint payload from r.
func (i *NFTMintInfo) ReadCarbon(r *Reader) {
	i.SeriesID = r.Read4U()
	i.ROM = r.ReadByteArray()
	i.RAM = r.ReadByteArray()
}

// MintNonFungibleArgs are token-module arguments for minting Carbon NFTs.
type MintNonFungibleArgs struct {
	TokenID uint64
	Address Bytes32
	Tokens  []NFTMintInfo
}

// WriteCarbon writes the mint non-fungible arguments to w.
func (a *MintNonFungibleArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
	w.Write32(a.Address)
	writeNFTMintInfoArray(w, a.Tokens)
}

// ReadCarbon reads the mint non-fungible arguments from r.
func (a *MintNonFungibleArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
	a.Address = r.Read32()
	a.Tokens = readNFTMintInfoArray(r)
}

// CreateTokenSeriesArgs are token-module arguments for creating a token series.
type CreateTokenSeriesArgs struct {
	TokenID uint64
	Info    SeriesInfo
}

// WriteCarbon writes the create-series arguments to w.
func (a *CreateTokenSeriesArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
	a.Info.WriteCarbon(w)
}

// ReadCarbon reads the create-series arguments from r.
func (a *CreateTokenSeriesArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
	a.Info.ReadCarbon(r)
}

// CreateMintedTokenSeriesArgs are token-module arguments for creating and minting a series.
type CreateMintedTokenSeriesArgs struct {
	TokenID uint64
	Info    SeriesInfo
	Address Bytes32
	ROMs    [][]byte
	RAMs    [][]byte
}

// WriteCarbon writes the create-and-mint series arguments to w.
func (a *CreateMintedTokenSeriesArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
	a.Info.WriteCarbon(w)
	w.Write32(a.Address)
	w.WriteByteArrays(a.ROMs)
	w.WriteByteArrays(a.RAMs)
}

// ReadCarbon reads the create-and-mint series arguments from r.
func (a *CreateMintedTokenSeriesArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
	a.Info.ReadCarbon(r)
	a.Address = r.Read32()
	a.ROMs = r.ReadByteArrays()
	a.RAMs = r.ReadByteArrays()
}

// PhantasmaNFTMintInfo describes one Phantasma NFT mint payload.
type PhantasmaNFTMintInfo struct {
	PhantasmaSeriesID IntX
	ROM               []byte
	RAM               []byte
}

// WriteCarbon writes the Phantasma NFT mint payload to w.
func (i *PhantasmaNFTMintInfo) WriteCarbon(w *Writer) {
	i.PhantasmaSeriesID.WriteCarbon(w)
	w.WriteByteArray(i.ROM)
	w.WriteByteArray(i.RAM)
}

// ReadCarbon reads the Phantasma NFT mint payload from r.
func (i *PhantasmaNFTMintInfo) ReadCarbon(r *Reader) {
	i.PhantasmaSeriesID.ReadCarbon(r)
	i.ROM = r.ReadByteArray()
	i.RAM = r.ReadByteArray()
}

// MintPhantasmaNonFungibleArgs are token-module arguments for minting Phantasma NFTs.
type MintPhantasmaNonFungibleArgs struct {
	TokenID uint64
	Address Bytes32
	Tokens  []PhantasmaNFTMintInfo
}

// WriteCarbon writes the Phantasma NFT mint arguments to w.
func (a *MintPhantasmaNonFungibleArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
	w.Write32(a.Address)
	writePhantasmaNFTMintInfoArray(w, a.Tokens)
}

// ReadCarbon reads the Phantasma NFT mint arguments from r.
func (a *MintPhantasmaNonFungibleArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
	a.Address = r.Read32()
	a.Tokens = readPhantasmaNFTMintInfoArray(r)
}

// PhantasmaNFTMintResult describes one Phantasma NFT mint result row.
type PhantasmaNFTMintResult struct {
	PhantasmaNFTID   Bytes32
	CarbonInstanceID uint64
}

// WriteCarbon writes the Phantasma NFT mint result to w.
func (r0 *PhantasmaNFTMintResult) WriteCarbon(w *Writer) {
	w.Write32(r0.PhantasmaNFTID)
	w.Write8U(r0.CarbonInstanceID)
}

// ReadCarbon reads the Phantasma NFT mint result from r.
func (r0 *PhantasmaNFTMintResult) ReadCarbon(r *Reader) {
	r0.PhantasmaNFTID = r.Read32()
	r0.CarbonInstanceID = r.Read8U()
}

// MintFungibleArgs are token-module arguments for minting fungible tokens.
type MintFungibleArgs struct {
	TokenID uint64
	To      Bytes32
	Amount  IntX
}

// WriteCarbon writes the fungible mint arguments to w.
func (a *MintFungibleArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
	w.Write32(a.To)
	a.Amount.WriteCarbon(w)
}

// ReadCarbon reads the fungible mint arguments from r.
func (a *MintFungibleArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
	a.To = r.Read32()
	a.Amount.ReadCarbon(r)
}

// TransferFungibleArgs are token-module arguments for transferring fungible tokens.
type TransferFungibleArgs struct {
	To      Bytes32
	From    Bytes32
	TokenID uint64
	Amount  IntX
}

// WriteCarbon writes the fungible transfer arguments to w.
func (a *TransferFungibleArgs) WriteCarbon(w *Writer) {
	w.Write32(a.To)
	w.Write32(a.From)
	w.Write8U(a.TokenID)
	a.Amount.WriteCarbon(w)
}

// ReadCarbon reads the fungible transfer arguments from r.
func (a *TransferFungibleArgs) ReadCarbon(r *Reader) {
	a.To = r.Read32()
	a.From = r.Read32()
	a.TokenID = r.Read8U()
	a.Amount.ReadCarbon(r)
}

// TransferNonFungibleArgs are token-module arguments for transferring NFTs.
type TransferNonFungibleArgs struct {
	To          Bytes32
	From        Bytes32
	TokenID     uint64
	InstanceIDs []uint64
}

// WriteCarbon writes the NFT transfer arguments to w.
func (a *TransferNonFungibleArgs) WriteCarbon(w *Writer) {
	w.Write32(a.To)
	w.Write32(a.From)
	w.Write8U(a.TokenID)
	w.WriteUint64Array(a.InstanceIDs)
}

// ReadCarbon reads the NFT transfer arguments from r.
func (a *TransferNonFungibleArgs) ReadCarbon(r *Reader) {
	a.To = r.Read32()
	a.From = r.Read32()
	a.TokenID = r.Read8U()
	a.InstanceIDs = r.ReadUint64Array()
}

// BurnFungibleArgs are token-module arguments for burning fungible tokens.
type BurnFungibleArgs struct {
	TokenID uint64
	From    Bytes32
	Amount  IntX
}

// WriteCarbon writes the fungible burn arguments to w.
func (a *BurnFungibleArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
	w.Write32(a.From)
	a.Amount.WriteCarbon(w)
}

// ReadCarbon reads the fungible burn arguments from r.
func (a *BurnFungibleArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
	a.From = r.Read32()
	a.Amount.ReadCarbon(r)
}

// BurnNonFungibleArgs are token-module arguments for burning NFTs.
type BurnNonFungibleArgs struct {
	TokenID     uint64
	From        Bytes32
	InstanceIDs []uint64
}

// WriteCarbon writes the NFT burn arguments to w.
func (a *BurnNonFungibleArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
	w.Write32(a.From)
	w.WriteUint64Array(a.InstanceIDs)
}

// ReadCarbon reads the NFT burn arguments from r.
func (a *BurnNonFungibleArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
	a.From = r.Read32()
	a.InstanceIDs = r.ReadUint64Array()
}

// UpdateTokenMetadataArgs are token-module arguments for updating token metadata.
type UpdateTokenMetadataArgs struct {
	TokenID  uint64
	Metadata VMDynamicStruct
}

// WriteCarbon writes the token metadata update arguments to w.
func (a *UpdateTokenMetadataArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
	a.Metadata.WriteCarbon(w)
}

// ReadCarbon reads the token metadata update arguments from r.
func (a *UpdateTokenMetadataArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
	a.Metadata.ReadCarbon(r)
}

// UpdateSeriesMetadataArgs are token-module arguments for updating series metadata.
type UpdateSeriesMetadataArgs struct {
	TokenID  uint64
	SeriesID uint32
	Metadata []byte
}

// WriteCarbon writes the series metadata update arguments to w.
func (a *UpdateSeriesMetadataArgs) WriteCarbon(w *Writer) {
	w.Write8U(a.TokenID)
	w.Write4U(a.SeriesID)
	w.WriteByteArray(a.Metadata)
}

// ReadCarbon reads the series metadata update arguments from r.
func (a *UpdateSeriesMetadataArgs) ReadCarbon(r *Reader) {
	a.TokenID = r.Read8U()
	a.SeriesID = r.Read4U()
	a.Metadata = r.ReadByteArray()
}

func writeNFTMintInfoArray(w *Writer, values []NFTMintInfo) {
	w.Write4(int32(len(values)))
	for i := range values {
		values[i].WriteCarbon(w)
	}
}

func readNFTMintInfoArray(r *Reader) []NFTMintInfo {
	count := r.ReadLength()
	out := make([]NFTMintInfo, count)
	for i := range out {
		out[i].ReadCarbon(r)
	}
	return out
}

func writePhantasmaNFTMintInfoArray(w *Writer, values []PhantasmaNFTMintInfo) {
	w.Write4(int32(len(values)))
	for i := range values {
		values[i].WriteCarbon(w)
	}
}

func readPhantasmaNFTMintInfoArray(r *Reader) []PhantasmaNFTMintInfo {
	count := r.ReadLength()
	out := make([]PhantasmaNFTMintInfo, count)
	for i := range out {
		out[i].ReadCarbon(r)
	}
	return out
}
