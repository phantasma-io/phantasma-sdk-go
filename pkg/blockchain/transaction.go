package blockchain

import (
	"fmt"
	"strings"

	crypto "github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/io"
	hashing "github.com/phantasma-io/phantasma-sdk-go/pkg/util/hashing"
)

// MaxTransactionSignatureCount limits transaction signature allocation while decoding untrusted bytes.
const MaxTransactionSignatureCount = 1024

// Transaction represents a classic Phantasma VM transaction.
type Transaction struct {

	// Code to run in PhantasmaVM for this transaction.
	Script []byte

	NexusName string

	ChainName string

	Expiration uint32

	Payload []byte

	Signatures []crypto.Signature

	Hash crypto.Hash
}

// NewTransaction creates a new transaction object.
func NewTransaction(nexusName, chainName string, script []byte, timestamp uint32, payload []byte) Transaction {

	tx := Transaction{
		NexusName:  nexusName,
		ChainName:  chainName,
		Script:     script,
		Expiration: timestamp,
		Payload:    payload,
		Signatures: []crypto.Signature{},
		Hash:       crypto.Hash{},
	}

	tx.updateHash()

	return tx
}

// updateHash sets the hash of the transaction.
func (tx *Transaction) updateHash() {
	data := tx.BytesEx(false)
	bytes := hashing.Sha256(data)
	hash, err := crypto.HashFromBytes(bytes)
	if err != nil {
		panic("Updating hash on tx failed!")
	}
	tx.Hash = hash
}

// HasSignatures checks if the transaction was signed already.
func (tx *Transaction) HasSignatures() bool {
	return len(tx.Signatures) > 0
}

// SerializeEx writes the transaction with or without the signature section.
func (tx *Transaction) SerializeEx(writer *io.BinWriter, withSignatures bool) {
	writer.WriteString(tx.NexusName)
	writer.WriteString(tx.ChainName)
	writer.WriteVarBytes(tx.Script)
	writer.WriteU32LE(tx.Expiration)
	writer.WriteVarBytes(tx.Payload)

	if withSignatures {
		signatureCount := 0
		for _, signature := range tx.Signatures {
			if signature != nil {
				signatureCount++
			}
		}

		writer.WriteVarUint(uint64(signatureCount))
		for _, signature := range tx.Signatures {
			if signature == nil {
				continue
			}
			writer.WriteB(byte(signature.Kind()))
			signature.Serialize(writer)
		}
	}
}

// Serialize implements the Serializable interface.
func (tx *Transaction) Serialize(writer *io.BinWriter) {
	tx.SerializeEx(writer, true)
}

// Deserialize implements the Serializable interface.
func (tx *Transaction) Deserialize(reader *io.BinReader) {
	tx.NexusName = reader.ReadString()
	tx.ChainName = reader.ReadString()
	tx.Script = reader.ReadVarBytes()
	tx.Expiration = reader.ReadU32LE()
	tx.Payload = reader.ReadVarBytes()

	signatureCount := reader.ReadVarUint()
	if reader.Err != nil {
		return
	}
	if signatureCount > MaxTransactionSignatureCount {
		reader.Err = fmt.Errorf("transaction signature count is too large: got %d, max %d", signatureCount, MaxTransactionSignatureCount)
		return
	}

	if signatureCount == 0 {
		tx.Signatures = []crypto.Signature{}
	} else {
		tx.Signatures = make([]crypto.Signature, 0, signatureCount)
		for i := uint64(0); i < signatureCount; i++ {
			signature := deserializeSignature(reader)
			if reader.Err != nil {
				return
			}
			tx.Signatures = append(tx.Signatures, signature)
		}
	}

	if reader.Err != nil {
		return
	}
	tx.updateHash()
}

func deserializeSignature(reader *io.BinReader) crypto.Signature {
	kind := crypto.SignatureKind(reader.ReadB())
	if reader.Err != nil {
		return nil
	}

	switch kind {
	case crypto.Ed25519:
		signature := &crypto.Ed25519Signature{}
		signature.Deserialize(reader)
		return signature
	default:
		reader.Err = fmt.Errorf("unsupported transaction signature kind: %d", kind)
		return nil
	}
}

// String returns the transaction hash string.
func (tx *Transaction) String() string {
	return tx.Hash.String()
}

// BytesEx returns the serialized transaction with or without signatures.
func (tx *Transaction) BytesEx(withSignatures bool) []byte {
	bw := *io.NewBufBinWriter()
	tx.SerializeEx(bw.BinWriter, withSignatures)
	return bw.Bytes()
}

// Bytes returns the serialized transaction including signatures.
func (tx *Transaction) Bytes() []byte {
	bw := *io.NewBufBinWriter()
	tx.Serialize(bw.BinWriter)
	return bw.Bytes()
}

// Sign appends a signature to the transaction.
func (tx *Transaction) Sign(keyPair crypto.KeyPair) error {
	if keyPair == nil {
		return fmt.Errorf("key pair is required")
	}

	msg := tx.BytesEx(false)

	signature, err := keyPair.Sign(msg)
	if err != nil {
		return err
	}

	tx.Signatures = append(tx.Signatures, signature)
	return nil
}

// IsSignedBy checks if a transaction is signed by a specific address.
func (tx *Transaction) IsSignedBy(addresses []crypto.Address) bool {
	if !tx.HasSignatures() {
		return false
	}

	msg := tx.BytesEx(false)

	for _, signature := range tx.Signatures {
		if signature.Verify(msg, addresses) {
			return true
		}
	}

	return false
}

// Mine searches for a payload nonce that satisfies the requested hash difficulty.
func (tx *Transaction) Mine(difficulty int) {
	if difficulty == 0 {
		return
	}

	var nonce uint32 = 0

	for {
		if tx.Hash.GetDifficulty() >= difficulty {
			break
		}

		if nonce == 0 {
			tx.Payload = make([]byte, 4)
		}

		nonce++

		tx.Payload[0] = byte((nonce >> 0) & 0xFF)
		tx.Payload[1] = byte((nonce >> 8) & 0xFF)
		tx.Payload[2] = byte((nonce >> 16) & 0xFF)
		tx.Payload[3] = byte((nonce >> 24) & 0xFF)
		tx.updateHash()
	}
}

// TxStateIsSuccess reports whether a transaction VM state is successful.
func TxStateIsSuccess(state string) bool {
	return strings.ToUpper(state) == "HALT"
}

// TxStateIsFault reports whether a transaction VM state is a fault state.
func TxStateIsFault(state string) bool {
	state = strings.ToUpper(state)
	return state == "FAULT" || state == "BREAK"
}
