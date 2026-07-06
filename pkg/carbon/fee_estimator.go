package carbon

import (
	"fmt"
	"math/bits"
)

// NativeFeeKind enumerates the operations supported by the Tier-1 fee estimator.
type NativeFeeKind int

const (
	// NativeFeeTransferFungible is a fungible token transfer (TxTypes TransferFungible / _GasPayer).
	NativeFeeTransferFungible NativeFeeKind = iota
	// NativeFeeTransferNonFungible is an NFT transfer of Count instances.
	NativeFeeTransferNonFungible
	// NativeFeeMintFungible is a fungible mint.
	NativeFeeMintFungible
	// NativeFeeMintNonFungible is an NFT mint of Count instances.
	NativeFeeMintNonFungible
	// NativeFeeBurnFungible is a fungible burn.
	NativeFeeBurnFungible
	// NativeFeeBurnNonFungible is an NFT burn of Count instances.
	NativeFeeBurnNonFungible
	// NativeFeeCreateToken is a TokenContract.CreateToken call; set SymbolLength when a symbol is used.
	NativeFeeCreateToken
	// NativeFeeCreateTokenSeries is a TokenContract.CreateTokenSeries call.
	NativeFeeCreateTokenSeries
	// NativeFeeRegisterName is a GovernanceContract.RegisterName call; NameLength is required.
	NativeFeeRegisterName
	// NativeFeeScript is a generic Phantasma VM script transaction (AllowGas/SpendGas pattern:
	// stake, marketplace, custom contract calls). Script opcode costs are not closed-form in
	// Tier-1; the estimate budgets ScriptUnitsAllowance VM work units on top of the byte fee.
	// For an exact script bill use the node-side Tier-2 estimator once available.
	NativeFeeScript
)

// GasModelV2UnitsPerBlockDataByte is the gas-model-v2 price of block-carried bytes, in gas
// units per byte. A versioned consensus constant of the v2 gas model (node data_blockchain.h
// kGasModelV2UnitsPerBlockDataByte), deliberately not part of the on-chain config.
const GasModelV2UnitsPerBlockDataByte = 25

// WitnessArrayEntryBytes is the serialized size of one witness-array entry (32-byte address +
// 64-byte signature).
const WitnessArrayEntryBytes = 96

// NativeSignatureBytes is the serialized size of one bare signature (native TxTypes carry no
// witness array).
const NativeSignatureBytes = 64

// NativeFeeParams are optional inputs for EstimateNativeFee. The zero value produces a safe
// single-signer estimate; set fields for exactness (the unused part of a gas offer is always
// refunded, so generous values only lock the balance for the block, they cost nothing).
type NativeFeeParams struct {
	// Count is the instance count for NFT kinds. 0 means 1.
	Count uint32
	// SymbolLength is the token symbol length in characters (CreateToken). 0 = no symbol.
	SymbolLength int
	// NameLength is the registered name length in characters (RegisterName kind, required).
	NameLength int
	// EnvelopeBytes is the full signed transaction size in bytes (the envelope carried in the
	// block). 0 = use a conservative per-kind default. Under gas model v2 every envelope byte
	// is billed, so pass the real size (see EnvelopeBytesFor) for exact numbers.
	EnvelopeBytes uint32
	// PayloadBytes is the user payload size attached to the tx (billed under gas model v1).
	PayloadBytes uint32
	// FreshRows is the maximum number of paid storage rows the transaction can create (fresh
	// token-balance rows, NFT lookup rows, ...). 0 (the zero value) = use the per-kind
	// worst-case default; NoFreshRows = the tx cannot create paid rows (SOUL/KCAL-only flows).
	// SOUL/KCAL balance rows are free and never count. Determines MaxData:
	// escrow = rows * DataEscrowPerRow.
	FreshRows int64
	// RomRamBytes is the per-instance ROM+RAM size for MintNonFungible (stored state, drives escrow).
	RomRamBytes uint32
	// ScriptUnitsAllowance is the VM work-unit allowance for the Script kind. 0 means the
	// default 5000, which exceeds every script seen in mainnet history (max 3392 units).
	ScriptUnitsAllowance uint64
}

// NoFreshRows marks a transaction that cannot create paid storage rows (MaxData 0), because
// the zero value of NativeFeeParams.FreshRows selects the per-kind worst-case default instead.
const NoFreshRows int64 = -1

// NativeFeeEstimate is the result of a Tier-1 fee estimate. Gas values are kcal-base (1 KCAL =
// 1e10 kcal-base); MaxData is in data-token atoms (SOUL, 1 SOUL = 1e8 atoms).
type NativeFeeEstimate struct {
	// MaxGas is the recommended gas offer (TxMsg maxGas). Includes deterministic headroom; the
	// unused part is refunded.
	MaxGas uint64
	// MaxData is the recommended storage-escrow ceiling (TxMsg maxData). Only the actually
	// created rows are escrowed.
	MaxData uint64
	// ExpectedGasBill is the bill the chain formula yields for exactly the provided inputs (no
	// headroom).
	ExpectedGasBill uint64
}

// EstimateNativeFee is the Tier-1 static fee calculator: closed-form gas/data offers for native
// operations, exact under both gas models (selected by GasConfig.Version). It mirrors the
// validator billing formula (node blockchain.cpp settlement + token/governance contract gas
// sites); any change to those formulas ships as a new gas-model version, never silently.
func EstimateNativeFee(kind NativeFeeKind, config *GasConfig, params NativeFeeParams) (NativeFeeEstimate, error) {
	if config == nil {
		return NativeFeeEstimate{}, fmt.Errorf("EstimateNativeFee: config is nil")
	}
	count := params.Count
	if count == 0 {
		count = 1
	}
	v2 := config.HasGasModelV2()

	// Work units consumed by the operation itself (the ConsumeGas amounts in the token /
	// governance contracts) and, under v2, the direct kcal-base policy fee that replaces the
	// v1 unit-priced product prices.
	var workUnits, policyFee uint64
	switch kind {
	case NativeFeeTransferFungible, NativeFeeMintFungible, NativeFeeBurnFungible:
		workUnits = config.GasFeeTransfer
	case NativeFeeTransferNonFungible, NativeFeeMintNonFungible, NativeFeeBurnNonFungible:
		workUnits = saturatingMul(config.GasFeeTransfer, uint64(count))
	case NativeFeeCreateToken:
		// Symbol price halves per character after the first; shift is validated by the chain
		// against MaxTokenSymbolLength, mirror that bound here.
		shift, err := symbolShift(params.SymbolLength, config.MaxTokenSymbolLength, "SymbolLength")
		if err != nil {
			return NativeFeeEstimate{}, err
		}
		if v2 {
			policyFee = config.PolicyFeeCreateTokenBase
			if params.SymbolLength > 0 {
				policyFee = saturatingAdd(policyFee, config.PolicyFeeCreateTokenSymbol>>shift)
			}
		} else {
			workUnits = config.GasFeeCreateTokenBase
			if params.SymbolLength > 0 {
				workUnits = saturatingAdd(workUnits, config.GasFeeCreateTokenSymbol>>shift)
			}
		}
	case NativeFeeCreateTokenSeries:
		if v2 {
			policyFee = config.PolicyFeeCreateTokenSeries
		} else {
			workUnits = config.GasFeeCreateTokenSeries
		}
	case NativeFeeRegisterName:
		if params.NameLength <= 0 {
			return NativeFeeEstimate{}, fmt.Errorf("EstimateNativeFee: NameLength is required for RegisterName")
		}
		shift, err := symbolShift(params.NameLength, config.MaxNameLength, "NameLength")
		if err != nil {
			return NativeFeeEstimate{}, err
		}
		if v2 {
			policyFee = config.PolicyFeeRegisterName >> shift
		} else {
			workUnits = config.GasFeeRegisterName >> shift
		}
	case NativeFeeScript:
		workUnits = params.ScriptUnitsAllowance
		if workUnits == 0 {
			workUnits = 5000
		}
	default:
		return NativeFeeEstimate{}, fmt.Errorf("EstimateNativeFee: unknown fee kind %d", kind)
	}

	rows := params.FreshRows
	if rows < 0 {
		rows = 0 // NoFreshRows: the tx cannot create paid rows
	} else if rows == 0 {
		rows = defaultFreshRows(kind, count, params.RomRamBytes)
	}
	envelope := uint64(params.EnvelopeBytes)
	if envelope == 0 {
		envelope = defaultEnvelopeBytes(kind, count, params.RomRamBytes)
	}
	// Native TxTypes store no events in the block; script txs do (Notify), so budget some.
	var eventBytes uint64
	if kind == NativeFeeScript {
		eventBytes = 512
	}

	var expected, maxGas uint64
	if v2 {
		// v2 formula: bill = mulShiftSat(workUnits + blockData*25, mult, shift) + policyFee,
		// floored at MinimumGasBill. blockData = envelope + events + net storage quanta (quanta
		// are added to the byte count by the chain formula).
		expected = billV2(workUnits, envelope, eventBytes, uint64(rows), policyFee, config)
		// Offer headroom: +25% over the padded bill covers witness-size wiggle and event
		// variance; deterministic and always refunded down to the actual bill.
		padded := billV2(workUnits, roundUp(envelope, 128), eventBytes, uint64(rows), policyFee, config)
		maxGas = saturatingAdd(padded, padded/4)
		if maxGas < config.MinimumGasBill {
			maxGas = config.MinimumGasBill
		}
		if maxGas < config.MinimumGasOffer {
			maxGas = config.MinimumGasOffer
		}
	} else {
		// v1 formula: bill = (workUnits * mult >> shift) + blockData * GasFeePerByte where
		// blockData = payload + events + net storage quanta (no envelope term, no floor).
		work := mulShift(workUnits, config.FeeMultiplier, config.FeeShift)
		blockData := saturatingAdd(saturatingAdd(uint64(params.PayloadBytes), eventBytes), uint64(rows))
		expected = saturatingAdd(work, saturatingMul(blockData, config.GasFeePerByte))
		// Offer shape mirrors the validator's own test-agent stdFee: a 2x minimum-offer pad
		// plus a flat 1 KiB block-data allowance on top of the work term.
		byteAllowance := blockData
		if byteAllowance < 1024 {
			byteAllowance = 1024
		}
		maxGas = saturatingAdd(saturatingAdd(config.MinimumGasOffer*2, work),
			saturatingMul(byteAllowance, config.GasFeePerByte))
	}

	return NativeFeeEstimate{
		MaxGas:          maxGas,
		MaxData:         saturatingMul(uint64(rows), config.DataEscrowPerRow),
		ExpectedGasBill: expected,
	}, nil
}

// EnvelopeBytesFor computes the envelope size (signed tx bytes as carried in the block) from a
// serialized unsigned message length and the number of signers. Use with
// NativeFeeParams.EnvelopeBytes for exact v2 estimates. Witness layout mirrors SignedTxMsg:
// native TxTypes append bare 64-byte signatures (one, or two for the _GasPayer variants);
// Call/Trade/Phantasma txs append a length-prefixed witness array (32-byte address + 64-byte
// signature per entry).
func EnvelopeBytesFor(kind NativeFeeKind, serializedMessageLength int, witnessCount int) (uint32, error) {
	if serializedMessageLength < 0 {
		return 0, fmt.Errorf("EnvelopeBytesFor: serializedMessageLength must not be negative")
	}
	if witnessCount < 0 {
		return 0, fmt.Errorf("EnvelopeBytesFor: witnessCount must not be negative")
	}
	switch kind {
	case NativeFeeTransferFungible, NativeFeeTransferNonFungible, NativeFeeMintFungible,
		NativeFeeMintNonFungible, NativeFeeBurnFungible, NativeFeeBurnNonFungible:
		return uint32(serializedMessageLength + NativeSignatureBytes*witnessCount), nil
	default:
		// CreateToken/CreateTokenSeries/RegisterName ride TxTypes.Call; Script rides
		// TxTypes.Phantasma - both carry the witness array form.
		return uint32(serializedMessageLength + 4 + WitnessArrayEntryBytes*witnessCount), nil
	}
}

func billV2(workUnits, envelope, eventBytes, rows, policyFee uint64, config *GasConfig) uint64 {
	blockData := saturatingAdd(saturatingAdd(envelope, eventBytes), rows)
	byteUnits := saturatingMul(blockData, GasModelV2UnitsPerBlockDataByte)
	bill := mulShift(saturatingAdd(workUnits, byteUnits), config.FeeMultiplier, config.FeeShift)
	bill = saturatingAdd(bill, policyFee)
	if bill < config.MinimumGasBill {
		return config.MinimumGasBill
	}
	return bill
}

// mulShift is the chain fee scaling: (value * FeeMultiplier) >> FeeShift with a 128-bit
// intermediate, saturating to uint64. Matches the validator's v2 MulShiftSaturateU64; for sane
// v1 configs (live values never overflow 64 bits) it is also bit-identical to the v1 math.
func mulShift(value, multiplier uint64, shift byte) uint64 {
	if shift >= 64 {
		return 0 // the chain clamps oversized shifts to a zero delta
	}
	hi, lo := bits.Mul64(value, multiplier)
	if shift == 0 {
		if hi != 0 {
			return ^uint64(0)
		}
		return lo
	}
	if hi>>shift != 0 {
		return ^uint64(0)
	}
	return hi<<(64-shift) | lo>>shift
}

func symbolShift(length int, maxLength byte, paramName string) (uint, error) {
	if length < 0 {
		return 0, fmt.Errorf("EstimateNativeFee: %s must not be negative", paramName)
	}
	if length == 0 {
		return 0, nil
	}
	shift := length - 1
	// The chain asserts shift < MaxNameLength / MaxTokenSymbolLength; a longer input could
	// never be admitted, so reject it here instead of quoting a fee for an impossible tx.
	if maxLength != 0 && shift >= int(maxLength) {
		return 0, fmt.Errorf("EstimateNativeFee: %s %d exceeds the chain maximum %d", paramName, length, maxLength)
	}
	return uint(shift), nil
}

// defaultFreshRows is the worst-case paid rows the operation can create (drives MaxData).
// Refund-only operations (burns) need no escrow allowance: refunds never require maxData budget.
func defaultFreshRows(kind NativeFeeKind, count uint32, romRamBytes uint32) int64 {
	switch kind {
	case NativeFeeTransferFungible, NativeFeeMintFungible:
		return 1 // recipient balance row may be fresh (SOUL/KCAL rows would be free)
	case NativeFeeTransferNonFungible:
		// Per instance the chain deletes the sender's NFT-lookup row and creates the
		// recipient's (creation escrows at the current price, the deletion refunds the old
		// row's own deposit) + possibly a fresh recipient balance row.
		return int64(count) + 1
	case NativeFeeMintNonFungible:
		// Per instance: owner row + lookup row + the instance state rows holding ROM/RAM
		// (1 KiB quanta), plus possibly a fresh recipient balance row.
		return int64(count)*(2+int64((romRamBytes+1023)/1024)) + 1
	case NativeFeeBurnFungible, NativeFeeBurnNonFungible:
		return 0
	case NativeFeeCreateToken:
		return 8 // token info + symbol lookup + supply/config rows, metadata-dependent
	case NativeFeeCreateTokenSeries:
		return 4
	case NativeFeeRegisterName:
		return 2 // name->address and address->name rows
	default:
		return 4
	}
}

// defaultEnvelopeBytes gives conservative single-witness envelope defaults per kind; used only
// when the caller did not measure the real signed size. Generous by design: under v2 an
// oversized estimate only raises the refunded offer, never the settled bill.
func defaultEnvelopeBytes(kind NativeFeeKind, count uint32, romRamBytes uint32) uint64 {
	switch kind {
	case NativeFeeTransferFungible, NativeFeeMintFungible, NativeFeeBurnFungible, NativeFeeRegisterName:
		return 512
	case NativeFeeTransferNonFungible, NativeFeeBurnNonFungible:
		return 512 + 8*uint64(count) // 8 bytes per carried instance id
	case NativeFeeMintNonFungible:
		return 512 + uint64(count)*(64+uint64(romRamBytes)) // ROM/RAM ride the envelope
	case NativeFeeCreateToken:
		return 4096 // token metadata (icons, descriptions) dominates
	case NativeFeeCreateTokenSeries:
		return 2048
	default:
		return 1024
	}
}

func roundUp(value, step uint64) uint64 {
	rem := value % step
	if rem == 0 {
		return value
	}
	return value + (step - rem)
}

func saturatingAdd(a, b uint64) uint64 {
	if a > ^uint64(0)-b {
		return ^uint64(0)
	}
	return a + b
}

func saturatingMul(a, b uint64) uint64 {
	if a == 0 || b == 0 {
		return 0
	}
	if a > ^uint64(0)/b {
		return ^uint64(0)
	}
	return a * b
}
