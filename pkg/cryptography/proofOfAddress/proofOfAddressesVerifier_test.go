package proofOfAddress

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testMessage string = `I have signed this message with my Phantasma, Ethereum and Neo Legacy signatures to prove that following addresses belong to me and were derived from private key that belongs to me and to confirm my willingness to swap funds across these addresses upon my request. My public addresses are:
Phantasma address: P2KHhbVZWDv1ZLLoJccN3PUAb9x9BqRnUyH3ZEhu5YwBeJQ
Ethereum address: 0xDf738B927DA923fe0A5Fd3aD2192990C68913e6a
Ethereum public key: 025D3F7F469803C68C12B8F731576C74A9B5308484FD3B425D87C35CAED0A2E398
Neo Legacy address: Ae3aEA6CpvckvypAUShj2CLsy7sfynKUzj
Neo Legacy public key: 02183A301779007BF42DD7B5247587585B0524E13989F964C2A8E289A0CDC91F00

Phantasma signature: 35FDC3CE357CD099FCA8D8687D6B9DC6DE2DB1E4D752312EF5186627DE21E7091FF40ABC1CECB55D0E9FE15E0FEBFA73DD861ACF5E6EDA94B72B078F77112306
Ethereum signature: 50AEA773BF563991A9BE0F034442FABB5168DB192D123E43F990A38BE290D927E2925D02915D3ABA01D35B6E6D3795937BC4C786EFDDF4E5FA16842651A8BFFF
Neo Legacy signature: 605AD6099DA4C4E06122CC7A544B0CF18B94D4203B96604F68884996DB78220BF25E82B0CF8EC360FE9FD083D6268B6B441B64812805E9F03F2852EE50397FE6`

var testMessagePhaAddressIncorrect string = `I have signed this message with my Phantasma, Ethereum and Neo Legacy signatures to prove that following addresses belong to me and were derived from private key that belongs to me and to confirm my willingness to swap funds across these addresses upon my request. My public addresses are:
Phantasma address: P2KHhbVZWDv1ZLLoJccN3PUAb9x9BqRnUyH3ZEhu5YwBeJW
Ethereum address: 0xDf738B927DA923fe0A5Fd3aD2192990C68913e6a
Ethereum public key: 025D3F7F469803C68C12B8F731576C74A9B5308484FD3B425D87C35CAED0A2E398
Neo Legacy address: Ae3aEA6CpvckvypAUShj2CLsy7sfynKUzj
Neo Legacy public key: 02183A301779007BF42DD7B5247587585B0524E13989F964C2A8E289A0CDC91F00

Phantasma signature: 35FDC3CE357CD099FCA8D8687D6B9DC6DE2DB1E4D752312EF5186627DE21E7091FF40ABC1CECB55D0E9FE15E0FEBFA73DD861ACF5E6EDA94B72B078F77112306
Ethereum signature: 50AEA773BF563991A9BE0F034442FABB5168DB192D123E43F990A38BE290D927E2925D02915D3ABA01D35B6E6D3795937BC4C786EFDDF4E5FA16842651A8BFFF
Neo Legacy signature: 605AD6099DA4C4E06122CC7A544B0CF18B94D4203B96604F68884996DB78220BF25E82B0CF8EC360FE9FD083D6268B6B441B64812805E9F03F2852EE50397FE6`

var testMessagePhaSignatureIncorrect string = `I have signed this message with my Phantasma, Ethereum and Neo Legacy signatures to prove that following addresses belong to me and were derived from private key that belongs to me and to confirm my willingness to swap funds across these addresses upon my request. My public addresses are:
Phantasma address: P2KHhbVZWDv1ZLLoJccN3PUAb9x9BqRnUyH3ZEhu5YwBeJQ
Ethereum address: 0xDf738B927DA923fe0A5Fd3aD2192990C68913e6a
Ethereum public key: 025D3F7F469803C68C12B8F731576C74A9B5308484FD3B425D87C35CAED0A2E398
Neo Legacy address: Ae3aEA6CpvckvypAUShj2CLsy7sfynKUzj
Neo Legacy public key: 02183A301779007BF42DD7B5247587585B0524E13989F964C2A8E289A0CDC91F00

Phantasma signature: 35FDC3CE357CD099FCA8D8687D6B0DC6DE2DB1E4D752312EF5186627DE21E7091FF40ABC1CECB55D0E9FE15E0FEBFA73DD861ACF5E6EDA94B72B078F77112306
Ethereum signature: 50AEA773BF563991A9BE0F034442FABB5168DB192D123E43F990A38BE290D927E2925D02915D3ABA01D35B6E6D3795937BC4C786EFDDF4E5FA16842651A8BFFF
Neo Legacy signature: 605AD6099DA4C4E06122CC7A544B0CF18B94D4203B96604F68884996DB78220BF25E82B0CF8EC360FE9FD083D6268B6B441B64812805E9F03F2852EE50397FE6`

func TestProofOfAddressesVerifier(t *testing.T) {
	v, err := NewProofOfAddressesVerifier(testMessage)
	require.NoError(t, err)

	success, errorMessage, err := v.VerifyMessage()
	require.NoError(t, err)
	assert.Equal(t, true, success)
	assert.Equal(t, "", errorMessage)

	v, err = NewProofOfAddressesVerifier(testMessagePhaAddressIncorrect)
	require.NoError(t, err)
	success, errorMessage, err = v.VerifyMessage()
	require.NoError(t, err)
	assert.Equal(t, false, success)
	assert.Equal(t, "Phantasma signature is incorrect!\nEthereum signature is incorrect!\nNeo Legacy signature is incorrect!\n", errorMessage)

	v, err = NewProofOfAddressesVerifier(testMessagePhaSignatureIncorrect)
	require.NoError(t, err)
	success, errorMessage, err = v.VerifyMessage()
	require.NoError(t, err)
	assert.Equal(t, false, success)
	assert.Equal(t, "Phantasma signature is incorrect!\n", errorMessage)
}

func TestProofOfAddressesVerifierRejectsMalformedEthereumPublicKey(t *testing.T) {
	message := strings.Replace(testMessage, "Ethereum public key: 025D3F7F469803C68C12B8F731576C74A9B5308484FD3B425D87C35CAED0A2E398", "Ethereum public key: 010203", 1)
	v, err := NewProofOfAddressesVerifier(message)
	require.NoError(t, err)

	success, _, err := v.VerifyMessage()
	assert.False(t, success)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "public key length")
}
