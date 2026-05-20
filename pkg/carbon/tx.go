package carbon

import (
	"fmt"
	"math"
)

// TxType identifies a Carbon transaction message type.
type TxType byte

// Carbon transaction message types.
const (
	// TxTypeCall calls one Carbon module method.
	TxTypeCall TxType = iota
	TxTypeCallMulti
	TxTypeTrade
	TxTypeTransferFungible
	TxTypeTransferFungibleGasPayer
	TxTypeTransferNonFungibleSingle
	TxTypeTransferNonFungibleSingleGasPayer
	TxTypeTransferNonFungibleMulti
	TxTypeTransferNonFungibleMultiGasPayer
	TxTypeMintFungible
	TxTypeBurnFungible
	TxTypeBurnFungibleGasPayer
	TxTypeMintNonFungible
	TxTypeBurnNonFungible
	TxTypeBurnNonFungibleGasPayer
	TxTypePhantasma
	TxTypePhantasmaRaw
)

// TxPayload is implemented by all Carbon transaction payload types.
type TxPayload interface {
	Blob
	txPayload()
}

// TxMsg is an unsigned Carbon transaction message.
type TxMsg struct {
	Type    TxType
	Expiry  int64
	MaxGas  uint64
	MaxData uint64
	GasFrom Bytes32
	Payload SmallString
	Msg     TxPayload
}

// WriteCarbon writes the transaction message to w.
func (m *TxMsg) WriteCarbon(w *Writer) {
	if m.Msg == nil {
		panic("transaction payload is nil")
	}
	w.Write1(byte(m.Type))
	w.Write8(m.Expiry)
	w.Write8U(m.MaxGas)
	w.Write8U(m.MaxData)
	w.Write32(m.GasFrom)
	m.Payload.WriteCarbon(w)
	m.Msg.WriteCarbon(w)
}

// ReadCarbon reads the transaction message from r.
func (m *TxMsg) ReadCarbon(r *Reader) {
	m.Type = TxType(r.Read1())
	m.Expiry = r.Read8()
	m.MaxGas = r.Read8U()
	m.MaxData = r.Read8U()
	m.GasFrom = r.Read32()
	m.Payload.ReadCarbon(r)
	m.Msg = newPayloadForType(m.Type)
	m.Msg.ReadCarbon(r)
}

// Witness stores one Carbon transaction witness.
type Witness struct {
	Address   Bytes32
	Signature Bytes64
}

// WriteCarbon writes the witness to w.
func (witness *Witness) WriteCarbon(w *Writer) {
	w.Write32(witness.Address)
	w.Write64(witness.Signature)
}

// ReadCarbon reads the witness from r.
func (witness *Witness) ReadCarbon(r *Reader) {
	witness.Address = r.Read32()
	witness.Signature = r.Read64()
}

// SignedTxMsg stores a Carbon transaction and its witnesses.
type SignedTxMsg struct {
	Msg       TxMsg
	Witnesses []Witness
}

// WriteCarbon writes the signed transaction to w.
func (m *SignedTxMsg) WriteCarbon(w *Writer) {
	m.Msg.WriteCarbon(w)
	switch m.Msg.Type {
	case TxTypeTransferFungible,
		TxTypeTransferNonFungibleSingle,
		TxTypeTransferNonFungibleMulti,
		TxTypeMintFungible,
		TxTypeBurnFungible,
		TxTypeMintNonFungible,
		TxTypeBurnNonFungible:
		m.requireWitnessCount(1)
		if m.Witnesses[0].Address != m.Msg.GasFrom {
			panic("witness address mismatch")
		}
		w.Write64(m.Witnesses[0].Signature)
	case TxTypeTransferFungibleGasPayer,
		TxTypeTransferNonFungibleSingleGasPayer,
		TxTypeTransferNonFungibleMultiGasPayer,
		TxTypeBurnFungibleGasPayer,
		TxTypeBurnNonFungibleGasPayer:
		m.requireWitnessCount(2)
		if m.Witnesses[0].Address != m.Msg.GasFrom {
			panic("gas witness address mismatch")
		}
		w.Write64(m.Witnesses[0].Signature)
		w.Write64(m.Witnesses[1].Signature)
	case TxTypeCall, TxTypeCallMulti, TxTypeTrade, TxTypePhantasma:
		w.Write4(int32(len(m.Witnesses)))
		for i := range m.Witnesses {
			m.Witnesses[i].WriteCarbon(w)
		}
	case TxTypePhantasmaRaw:
		m.requireWitnessCount(0)
	default:
		panic(fmt.Sprintf("unsupported transaction type %d", m.Msg.Type))
	}
}

// ReadCarbon reads the signed transaction from r.
func (m *SignedTxMsg) ReadCarbon(r *Reader) {
	m.Msg.ReadCarbon(r)
	m.Witnesses = m.Witnesses[:0]
	switch m.Msg.Type {
	case TxTypeTransferFungible,
		TxTypeTransferNonFungibleSingle,
		TxTypeTransferNonFungibleMulti,
		TxTypeMintFungible,
		TxTypeBurnFungible,
		TxTypeMintNonFungible,
		TxTypeBurnNonFungible:
		m.Witnesses = append(m.Witnesses, Witness{
			Address:   m.Msg.GasFrom,
			Signature: r.Read64(),
		})
	case TxTypeTransferFungibleGasPayer,
		TxTypeTransferNonFungibleSingleGasPayer,
		TxTypeTransferNonFungibleMultiGasPayer,
		TxTypeBurnFungibleGasPayer,
		TxTypeBurnNonFungibleGasPayer:
		m.Witnesses = append(m.Witnesses,
			Witness{Address: m.Msg.GasFrom, Signature: r.Read64()},
			Witness{Address: m.fromAddress(), Signature: r.Read64()},
		)
	case TxTypeCall, TxTypeCallMulti, TxTypeTrade, TxTypePhantasma:
		count := r.ReadLengthFor(96)
		for i := 0; i < count; i++ {
			var witness Witness
			witness.ReadCarbon(r)
			m.Witnesses = append(m.Witnesses, witness)
		}
	case TxTypePhantasmaRaw:
		m.Witnesses = nil
	default:
		panic(fmt.Sprintf("unsupported transaction type %d", m.Msg.Type))
	}
}

func (m *SignedTxMsg) requireWitnessCount(count int) {
	if len(m.Witnesses) != count {
		panic(fmt.Sprintf("expected %d witnesses, got %d", count, len(m.Witnesses)))
	}
}

func (m *SignedTxMsg) fromAddress() Bytes32 {
	switch msg := m.Msg.Msg.(type) {
	case *TxMsgTransferFungibleGasPayer:
		return msg.From
	case *TxMsgTransferNonFungibleSingleGasPayer:
		return msg.From
	case *TxMsgTransferNonFungibleMultiGasPayer:
		return msg.From
	case *TxMsgBurnFungibleGasPayer:
		return msg.From
	case *TxMsgBurnNonFungibleGasPayer:
		return msg.From
	default:
		return EmptyBytes32
	}
}

// CallArgSection stores one segmented call-argument section.
type CallArgSection struct {
	RegisterOffset int32
	Args           []byte
}

// MsgCallArgSections stores segmented call arguments.
type MsgCallArgSections struct {
	Sections []CallArgSection
}

// HasSections reports whether any call-argument sections are present.
func (s MsgCallArgSections) HasSections() bool {
	return len(s.Sections) > 0
}

// WriteCarbon writes segmented call arguments to w.
func (s MsgCallArgSections) WriteCarbon(w *Writer) {
	if len(s.Sections) == 0 {
		panic("arg sections are empty")
	}
	w.Write4(-int32(len(s.Sections)))
	for _, section := range s.Sections {
		if section.RegisterOffset < 0 {
			w.Write4(section.RegisterOffset)
			continue
		}
		w.Write4(int32(len(section.Args)))
		w.WriteRaw(section.Args)
	}
}

// ReadWithCount reads segmented call arguments after the negative count has been consumed.
func (s *MsgCallArgSections) ReadWithCount(r *Reader, countNegative int32) {
	if countNegative >= 0 {
		panic("arg sections count must be negative")
	}
	count64 := -int64(countNegative)
	if count64 > math.MaxInt32 {
		panic("invalid array length")
	}
	count := int(count64)
	if int64(count)*4 > int64(len(r.data)-r.off) {
		panic(fmt.Sprintf("array length %d exceeds remaining bytes %d", count, len(r.data)-r.off))
	}
	s.Sections = make([]CallArgSection, count)
	for i := range s.Sections {
		value := r.Read4()
		if value < 0 {
			s.Sections[i] = CallArgSection{RegisterOffset: value}
			continue
		}
		s.Sections[i] = CallArgSection{Args: r.ReadRaw(int(value))}
	}
}

// TxMsgCall calls one Carbon module method.
type TxMsgCall struct {
	ModuleID uint32
	MethodID uint32
	Args     []byte
	Sections *MsgCallArgSections
}

func (*TxMsgCall) txPayload() {}

// WriteCarbon writes the call payload to w.
func (m *TxMsgCall) WriteCarbon(w *Writer) {
	w.Write4U(m.ModuleID)
	w.Write4U(m.MethodID)
	if m.Sections != nil && m.Sections.HasSections() {
		m.Sections.WriteCarbon(w)
		return
	}
	w.Write4(int32(len(m.Args)))
	w.WriteRaw(m.Args)
}

// ReadCarbon reads the call payload from r.
func (m *TxMsgCall) ReadCarbon(r *Reader) {
	m.ModuleID = r.Read4U()
	m.MethodID = r.Read4U()
	length := r.Read4()
	if length >= 0 {
		m.Args = r.ReadRaw(int(length))
		m.Sections = nil
		return
	}
	m.Sections = &MsgCallArgSections{}
	m.Sections.ReadWithCount(r, length)
	m.Args = nil
}

// TxMsgCallMulti calls multiple Carbon module methods.
type TxMsgCallMulti struct {
	Calls []TxMsgCall
}

func (*TxMsgCallMulti) txPayload() {}

// WriteCarbon writes the multi-call payload to w.
func (m *TxMsgCallMulti) WriteCarbon(w *Writer) {
	w.Write4(int32(len(m.Calls)))
	for i := range m.Calls {
		m.Calls[i].WriteCarbon(w)
	}
}

// ReadCarbon reads the multi-call payload from r.
func (m *TxMsgCallMulti) ReadCarbon(r *Reader) {
	count := r.ReadLengthFor(12)
	m.Calls = make([]TxMsgCall, count)
	for i := range m.Calls {
		m.Calls[i].ReadCarbon(r)
	}
}

// TxMsgTrade groups token operations into one trade payload.
type TxMsgTrade struct {
	TransferF []TxMsgTransferFungibleGasPayer
	TransferN []TxMsgTransferNonFungibleSingleGasPayer
	MintF     []TxMsgMintFungible
	BurnF     []TxMsgBurnFungibleGasPayer
	MintN     []TxMsgMintNonFungible
	BurnN     []TxMsgBurnNonFungibleGasPayer
}

func (*TxMsgTrade) txPayload() {}

// WriteCarbon writes the trade payload to w.
func (m *TxMsgTrade) WriteCarbon(w *Writer) {
	writeTransferFungibleGasPayerSlice(w, m.TransferF)
	writeTransferNonFungibleSingleGasPayerSlice(w, m.TransferN)
	writeMintFungibleSlice(w, m.MintF)
	writeBurnFungibleGasPayerSlice(w, m.BurnF)
	writeMintNonFungibleSlice(w, m.MintN)
	writeBurnNonFungibleGasPayerSlice(w, m.BurnN)
}

// ReadCarbon reads the trade payload from r.
func (m *TxMsgTrade) ReadCarbon(r *Reader) {
	readTransferFungibleGasPayerSlice(r, &m.TransferF)
	readTransferNonFungibleSingleGasPayerSlice(r, &m.TransferN)
	readMintFungibleSlice(r, &m.MintF)
	readBurnFungibleGasPayerSlice(r, &m.BurnF)
	readMintNonFungibleSlice(r, &m.MintN)
	readBurnNonFungibleGasPayerSlice(r, &m.BurnN)
}

// TxMsgTransferFungible transfers fungible tokens from the gas payer.
type TxMsgTransferFungible struct {
	To      Bytes32
	TokenID uint64
	Amount  uint64
}

func (*TxMsgTransferFungible) txPayload() {}

// WriteCarbon writes the fungible transfer payload to w.
func (m *TxMsgTransferFungible) WriteCarbon(w *Writer) {
	w.Write32(m.To)
	w.Write8U(m.TokenID)
	w.Write8U(m.Amount)
}

// ReadCarbon reads the fungible transfer payload from r.
func (m *TxMsgTransferFungible) ReadCarbon(r *Reader) {
	m.To = r.Read32()
	m.TokenID = r.Read8U()
	m.Amount = r.Read8U()
}

// TxMsgTransferFungibleGasPayer transfers fungible tokens with a separate gas payer.
type TxMsgTransferFungibleGasPayer struct {
	To      Bytes32
	From    Bytes32
	TokenID uint64
	Amount  uint64
}

func (*TxMsgTransferFungibleGasPayer) txPayload() {}

// WriteCarbon writes the gas-payer fungible transfer payload to w.
func (m *TxMsgTransferFungibleGasPayer) WriteCarbon(w *Writer) {
	w.Write32(m.To)
	w.Write32(m.From)
	w.Write8U(m.TokenID)
	w.Write8U(m.Amount)
}

// ReadCarbon reads the gas-payer fungible transfer payload from r.
func (m *TxMsgTransferFungibleGasPayer) ReadCarbon(r *Reader) {
	m.To = r.Read32()
	m.From = r.Read32()
	m.TokenID = r.Read8U()
	m.Amount = r.Read8U()
}

// TxMsgTransferNonFungibleSingle transfers one NFT from the gas payer.
type TxMsgTransferNonFungibleSingle struct {
	To         Bytes32
	TokenID    uint64
	InstanceID uint64
}

func (*TxMsgTransferNonFungibleSingle) txPayload() {}

// WriteCarbon writes the single-NFT transfer payload to w.
func (m *TxMsgTransferNonFungibleSingle) WriteCarbon(w *Writer) {
	w.Write32(m.To)
	w.Write8U(m.TokenID)
	w.Write8U(m.InstanceID)
}

// ReadCarbon reads the single-NFT transfer payload from r.
func (m *TxMsgTransferNonFungibleSingle) ReadCarbon(r *Reader) {
	m.To = r.Read32()
	m.TokenID = r.Read8U()
	m.InstanceID = r.Read8U()
}

// TxMsgTransferNonFungibleSingleGasPayer transfers one NFT with a separate gas payer.
type TxMsgTransferNonFungibleSingleGasPayer struct {
	To         Bytes32
	From       Bytes32
	TokenID    uint64
	InstanceID uint64
}

func (*TxMsgTransferNonFungibleSingleGasPayer) txPayload() {}

// WriteCarbon writes the gas-payer single-NFT transfer payload to w.
func (m *TxMsgTransferNonFungibleSingleGasPayer) WriteCarbon(w *Writer) {
	w.Write32(m.To)
	w.Write32(m.From)
	w.Write8U(m.TokenID)
	w.Write8U(m.InstanceID)
}

// ReadCarbon reads the gas-payer single-NFT transfer payload from r.
func (m *TxMsgTransferNonFungibleSingleGasPayer) ReadCarbon(r *Reader) {
	m.To = r.Read32()
	m.From = r.Read32()
	m.TokenID = r.Read8U()
	m.InstanceID = r.Read8U()
}

// TxMsgTransferNonFungibleMulti transfers multiple NFTs from the gas payer.
type TxMsgTransferNonFungibleMulti struct {
	To          Bytes32
	TokenID     uint64
	InstanceIDs []uint64
}

func (*TxMsgTransferNonFungibleMulti) txPayload() {}

// WriteCarbon writes the multi-NFT transfer payload to w.
func (m *TxMsgTransferNonFungibleMulti) WriteCarbon(w *Writer) {
	w.Write32(m.To)
	w.Write8U(m.TokenID)
	w.WriteUint64Array(m.InstanceIDs)
}

// ReadCarbon reads the multi-NFT transfer payload from r.
func (m *TxMsgTransferNonFungibleMulti) ReadCarbon(r *Reader) {
	m.To = r.Read32()
	m.TokenID = r.Read8U()
	m.InstanceIDs = r.ReadUint64Array()
}

// TxMsgTransferNonFungibleMultiGasPayer transfers multiple NFTs with a separate gas payer.
type TxMsgTransferNonFungibleMultiGasPayer struct {
	To          Bytes32
	From        Bytes32
	TokenID     uint64
	InstanceIDs []uint64
}

func (*TxMsgTransferNonFungibleMultiGasPayer) txPayload() {}

// WriteCarbon writes the gas-payer multi-NFT transfer payload to w.
func (m *TxMsgTransferNonFungibleMultiGasPayer) WriteCarbon(w *Writer) {
	w.Write32(m.To)
	w.Write32(m.From)
	w.Write8U(m.TokenID)
	w.WriteUint64Array(m.InstanceIDs)
}

// ReadCarbon reads the gas-payer multi-NFT transfer payload from r.
func (m *TxMsgTransferNonFungibleMultiGasPayer) ReadCarbon(r *Reader) {
	m.To = r.Read32()
	m.From = r.Read32()
	m.TokenID = r.Read8U()
	m.InstanceIDs = r.ReadUint64Array()
}

// TxMsgMintFungible mints fungible tokens.
type TxMsgMintFungible struct {
	TokenID uint64
	To      Bytes32
	Amount  IntX
}

func (*TxMsgMintFungible) txPayload() {}

// WriteCarbon writes the fungible mint payload to w.
func (m *TxMsgMintFungible) WriteCarbon(w *Writer) {
	w.Write8U(m.TokenID)
	w.Write32(m.To)
	m.Amount.WriteCarbon(w)
}

// ReadCarbon reads the fungible mint payload from r.
func (m *TxMsgMintFungible) ReadCarbon(r *Reader) {
	m.TokenID = r.Read8U()
	m.To = r.Read32()
	m.Amount.ReadCarbon(r)
}

// TxMsgBurnFungible burns fungible tokens from the gas payer.
type TxMsgBurnFungible struct {
	TokenID uint64
	Amount  IntX
}

func (*TxMsgBurnFungible) txPayload() {}

// WriteCarbon writes the fungible burn payload to w.
func (m *TxMsgBurnFungible) WriteCarbon(w *Writer) {
	w.Write8U(m.TokenID)
	m.Amount.WriteCarbon(w)
}

// ReadCarbon reads the fungible burn payload from r.
func (m *TxMsgBurnFungible) ReadCarbon(r *Reader) {
	m.TokenID = r.Read8U()
	m.Amount.ReadCarbon(r)
}

// TxMsgBurnFungibleGasPayer burns fungible tokens with a separate gas payer.
type TxMsgBurnFungibleGasPayer struct {
	TokenID uint64
	From    Bytes32
	Amount  IntX
}

func (*TxMsgBurnFungibleGasPayer) txPayload() {}

// WriteCarbon writes the gas-payer fungible burn payload to w.
func (m *TxMsgBurnFungibleGasPayer) WriteCarbon(w *Writer) {
	w.Write8U(m.TokenID)
	w.Write32(m.From)
	m.Amount.WriteCarbon(w)
}

// ReadCarbon reads the gas-payer fungible burn payload from r.
func (m *TxMsgBurnFungibleGasPayer) ReadCarbon(r *Reader) {
	m.TokenID = r.Read8U()
	m.From = r.Read32()
	m.Amount.ReadCarbon(r)
}

// TxMsgMintNonFungible mints one Carbon NFT.
type TxMsgMintNonFungible struct {
	TokenID  uint64
	To       Bytes32
	SeriesID uint32
	ROM      []byte
	RAM      []byte
}

func (*TxMsgMintNonFungible) txPayload() {}

// WriteCarbon writes the NFT mint payload to w.
func (m *TxMsgMintNonFungible) WriteCarbon(w *Writer) {
	w.Write8U(m.TokenID)
	w.Write32(m.To)
	w.Write4U(m.SeriesID)
	w.WriteByteArray(m.ROM)
	w.WriteByteArray(m.RAM)
}

// ReadCarbon reads the NFT mint payload from r.
func (m *TxMsgMintNonFungible) ReadCarbon(r *Reader) {
	m.TokenID = r.Read8U()
	m.To = r.Read32()
	m.SeriesID = r.Read4U()
	m.ROM = r.ReadByteArray()
	m.RAM = r.ReadByteArray()
}

// TxMsgBurnNonFungible burns one NFT from the gas payer.
type TxMsgBurnNonFungible struct {
	TokenID    uint64
	InstanceID uint64
}

func (*TxMsgBurnNonFungible) txPayload() {}

// WriteCarbon writes the NFT burn payload to w.
func (m *TxMsgBurnNonFungible) WriteCarbon(w *Writer) {
	w.Write8U(m.TokenID)
	w.Write8U(m.InstanceID)
}

// ReadCarbon reads the NFT burn payload from r.
func (m *TxMsgBurnNonFungible) ReadCarbon(r *Reader) {
	m.TokenID = r.Read8U()
	m.InstanceID = r.Read8U()
}

// TxMsgBurnNonFungibleGasPayer burns one NFT with a separate gas payer.
type TxMsgBurnNonFungibleGasPayer struct {
	TokenID    uint64
	From       Bytes32
	InstanceID uint64
}

func (*TxMsgBurnNonFungibleGasPayer) txPayload() {}

// WriteCarbon writes the gas-payer NFT burn payload to w.
func (m *TxMsgBurnNonFungibleGasPayer) WriteCarbon(w *Writer) {
	w.Write8U(m.TokenID)
	w.Write32(m.From)
	w.Write8U(m.InstanceID)
}

// ReadCarbon reads the gas-payer NFT burn payload from r.
func (m *TxMsgBurnNonFungibleGasPayer) ReadCarbon(r *Reader) {
	m.TokenID = r.Read8U()
	m.From = r.Read32()
	m.InstanceID = r.Read8U()
}

// TxMsgPhantasma wraps a classic Phantasma VM transaction script.
type TxMsgPhantasma struct {
	Nexus  SmallString
	Chain  SmallString
	Script []byte
}

func (*TxMsgPhantasma) txPayload() {}

// WriteCarbon writes the Phantasma VM wrapper payload to w.
func (m *TxMsgPhantasma) WriteCarbon(w *Writer) {
	m.Nexus.WriteCarbon(w)
	m.Chain.WriteCarbon(w)
	w.WriteByteArray(m.Script)
}

// ReadCarbon reads the Phantasma VM wrapper payload from r.
func (m *TxMsgPhantasma) ReadCarbon(r *Reader) {
	m.Nexus.ReadCarbon(r)
	m.Chain.ReadCarbon(r)
	m.Script = r.ReadByteArray()
}

// TxMsgPhantasmaRaw wraps a pre-serialized Phantasma transaction.
type TxMsgPhantasmaRaw struct {
	Transaction []byte
}

func (*TxMsgPhantasmaRaw) txPayload() {}

// WriteCarbon writes the raw Phantasma transaction payload to w.
func (m *TxMsgPhantasmaRaw) WriteCarbon(w *Writer) {
	w.WriteByteArray(m.Transaction)
}

// ReadCarbon reads the raw Phantasma transaction payload from r.
func (m *TxMsgPhantasmaRaw) ReadCarbon(r *Reader) {
	m.Transaction = r.ReadByteArray()
}

func newPayloadForType(txType TxType) TxPayload {
	switch txType {
	case TxTypeCall:
		return &TxMsgCall{}
	case TxTypeCallMulti:
		return &TxMsgCallMulti{}
	case TxTypeTrade:
		return &TxMsgTrade{}
	case TxTypeTransferFungible:
		return &TxMsgTransferFungible{}
	case TxTypeTransferFungibleGasPayer:
		return &TxMsgTransferFungibleGasPayer{}
	case TxTypeTransferNonFungibleSingle:
		return &TxMsgTransferNonFungibleSingle{}
	case TxTypeTransferNonFungibleSingleGasPayer:
		return &TxMsgTransferNonFungibleSingleGasPayer{}
	case TxTypeTransferNonFungibleMulti:
		return &TxMsgTransferNonFungibleMulti{}
	case TxTypeTransferNonFungibleMultiGasPayer:
		return &TxMsgTransferNonFungibleMultiGasPayer{}
	case TxTypeMintFungible:
		return &TxMsgMintFungible{}
	case TxTypeBurnFungible:
		return &TxMsgBurnFungible{}
	case TxTypeBurnFungibleGasPayer:
		return &TxMsgBurnFungibleGasPayer{}
	case TxTypeMintNonFungible:
		return &TxMsgMintNonFungible{}
	case TxTypeBurnNonFungible:
		return &TxMsgBurnNonFungible{}
	case TxTypeBurnNonFungibleGasPayer:
		return &TxMsgBurnNonFungibleGasPayer{}
	case TxTypePhantasma:
		return &TxMsgPhantasma{}
	case TxTypePhantasmaRaw:
		return &TxMsgPhantasmaRaw{}
	default:
		panic(fmt.Sprintf("unsupported transaction type %d", txType))
	}
}

func writeTransferFungibleGasPayerSlice(w *Writer, values []TxMsgTransferFungibleGasPayer) {
	w.Write4(int32(len(values)))
	for i := range values {
		values[i].WriteCarbon(w)
	}
}

func readTransferFungibleGasPayerSlice(r *Reader, values *[]TxMsgTransferFungibleGasPayer) {
	count := r.ReadLengthFor(80)
	out := make([]TxMsgTransferFungibleGasPayer, count)
	for i := range out {
		out[i].ReadCarbon(r)
	}
	*values = out
}

func writeTransferNonFungibleSingleGasPayerSlice(w *Writer, values []TxMsgTransferNonFungibleSingleGasPayer) {
	w.Write4(int32(len(values)))
	for i := range values {
		values[i].WriteCarbon(w)
	}
}

func readTransferNonFungibleSingleGasPayerSlice(r *Reader, values *[]TxMsgTransferNonFungibleSingleGasPayer) {
	count := r.ReadLengthFor(80)
	out := make([]TxMsgTransferNonFungibleSingleGasPayer, count)
	for i := range out {
		out[i].ReadCarbon(r)
	}
	*values = out
}

func writeMintFungibleSlice(w *Writer, values []TxMsgMintFungible) {
	w.Write4(int32(len(values)))
	for i := range values {
		values[i].WriteCarbon(w)
	}
}

func readMintFungibleSlice(r *Reader, values *[]TxMsgMintFungible) {
	count := r.ReadLengthFor(49)
	out := make([]TxMsgMintFungible, count)
	for i := range out {
		out[i].ReadCarbon(r)
	}
	*values = out
}

func writeBurnFungibleGasPayerSlice(w *Writer, values []TxMsgBurnFungibleGasPayer) {
	w.Write4(int32(len(values)))
	for i := range values {
		values[i].WriteCarbon(w)
	}
}

func readBurnFungibleGasPayerSlice(r *Reader, values *[]TxMsgBurnFungibleGasPayer) {
	count := r.ReadLengthFor(49)
	out := make([]TxMsgBurnFungibleGasPayer, count)
	for i := range out {
		out[i].ReadCarbon(r)
	}
	*values = out
}

func writeMintNonFungibleSlice(w *Writer, values []TxMsgMintNonFungible) {
	w.Write4(int32(len(values)))
	for i := range values {
		values[i].WriteCarbon(w)
	}
}

func readMintNonFungibleSlice(r *Reader, values *[]TxMsgMintNonFungible) {
	count := r.ReadLengthFor(52)
	out := make([]TxMsgMintNonFungible, count)
	for i := range out {
		out[i].ReadCarbon(r)
	}
	*values = out
}

func writeBurnNonFungibleGasPayerSlice(w *Writer, values []TxMsgBurnNonFungibleGasPayer) {
	w.Write4(int32(len(values)))
	for i := range values {
		values[i].WriteCarbon(w)
	}
}

func readBurnNonFungibleGasPayerSlice(r *Reader, values *[]TxMsgBurnNonFungibleGasPayer) {
	count := r.ReadLengthFor(48)
	out := make([]TxMsgBurnNonFungibleGasPayer, count)
	for i := range out {
		out[i].ReadCarbon(r)
	}
	*values = out
}
