package carbon

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
)

// SignTxMsg signs a Carbon transaction message with a single witness.
func SignTxMsg(msg TxMsg, keys cryptography.KeyPair) (SignedTxMsg, error) {
	if keys == nil {
		return SignedTxMsg{}, fmt.Errorf("key pair is required")
	}

	publicKey, err := Bytes32FromPublicKey(keys.PublicKey())
	if err != nil {
		return SignedTxMsg{}, err
	}

	privateKey := keys.ExpandedPrivateKey()
	if len(privateKey) != ed25519.PrivateKeySize {
		return SignedTxMsg{}, fmt.Errorf("expanded private key length must be %d, got %d", ed25519.PrivateKeySize, len(privateKey))
	}

	signature := ed25519.Sign(ed25519.PrivateKey(privateKey), Serialize(&msg))
	return SignedTxMsg{
		Msg: msg,
		Witnesses: []Witness{{
			Address:   publicKey,
			Signature: MustBytes64(signature),
		}},
	}, nil
}

// SignAndSerializeTxMsg signs msg and returns the serialized signed transaction bytes.
func SignAndSerializeTxMsg(msg TxMsg, keys cryptography.KeyPair) ([]byte, error) {
	signed, err := SignTxMsg(msg, keys)
	if err != nil {
		return nil, err
	}
	return Serialize(&signed), nil
}

// SignAndSerializeTxMsgHex signs msg and returns the serialized signed transaction as hex.
func SignAndSerializeTxMsgHex(msg TxMsg, keys cryptography.KeyPair) (string, error) {
	signed, err := SignAndSerializeTxMsg(msg, keys)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signed), nil
}
