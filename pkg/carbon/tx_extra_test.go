package carbon

import (
	"encoding/hex"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
)

func TestCarbonTxExtraVectors(t *testing.T) {
	// Extra transaction vectors cover non-helper payloads so helpers cannot silently corrupt base tx encoding.
	sender := testSenderPublicKey(t)
	receiver := testReceiverPublicKey(t)
	expiry := int64(1_759_711_416_000)
	maxGas := uint64(10_000_000)
	maxData := uint64(1000)
	payload := MustSmallString("test-payload")

	tests := []struct {
		name     string
		msg      TxMsg
		expected string
	}{
		{
			name: "transfer_fungible_gas_payer",
			msg: TxMsg{
				Type:    TxTypeTransferFungibleGasPayer,
				Expiry:  expiry,
				MaxGas:  maxGas,
				MaxData: maxData,
				GasFrom: sender,
				Payload: payload,
				Msg: &TxMsgTransferFungibleGasPayer{
					To:      receiver,
					From:    sender,
					TokenID: 1,
					Amount:  100_000_000,
				},
			},
			expected: "04C04EF9B6990100008096980000000000E803000000000000F94A8E45BDF1E37A8466B951849E92D1BAF870F49D1D04CD204D0BC9FE4308960C746573742D7061796C6F6164D4C5061B81C4682B27A0CFC6459CD9D7892EB60A43F73DD1060B6C478AA7C3D8F94A8E45BDF1E37A8466B951849E92D1BAF870F49D1D04CD204D0BC9FE430896010000000000000000E1F50500000000",
		},
		{
			name: "burn_fungible_gas_payer",
			msg: TxMsg{
				Type:    TxTypeBurnFungibleGasPayer,
				Expiry:  expiry,
				MaxGas:  maxGas,
				MaxData: maxData,
				GasFrom: sender,
				Payload: payload,
				Msg: &TxMsgBurnFungibleGasPayer{
					TokenID: 1,
					From:    sender,
					Amount:  IntXFromInt64(100_000_000),
				},
			},
			expected: "0BC04EF9B6990100008096980000000000E803000000000000F94A8E45BDF1E37A8466B951849E92D1BAF870F49D1D04CD204D0BC9FE4308960C746573742D7061796C6F61640100000000000000F94A8E45BDF1E37A8466B951849E92D1BAF870F49D1D04CD204D0BC9FE4308960800E1F50500000000",
		},
		{
			name: "mint_fungible",
			msg: TxMsg{
				Type:    TxTypeMintFungible,
				Expiry:  expiry,
				MaxGas:  maxGas,
				MaxData: maxData,
				GasFrom: sender,
				Payload: payload,
				Msg: &TxMsgMintFungible{
					TokenID: 1,
					To:      receiver,
					Amount:  IntXFromInt64(100_000_000),
				},
			},
			expected: "09C04EF9B6990100008096980000000000E803000000000000F94A8E45BDF1E37A8466B951849E92D1BAF870F49D1D04CD204D0BC9FE4308960C746573742D7061796C6F61640100000000000000D4C5061B81C4682B27A0CFC6459CD9D7892EB60A43F73DD1060B6C478AA7C3D80800E1F50500000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSerializedHex(t, tt.expected, &tt.msg)

			var decoded TxMsg
			readCarbon(t, tt.expected, &decoded)
			assertSerializedHex(t, tt.expected, &decoded)
		})
	}
}

func TestSignAndSerializeTxMsgMatchesSignedTransferVector(t *testing.T) {
	// Signing must match the shared signed-transfer vector, including witness order and compact signature layout.
	keys, err := cryptography.FromWIF("KwPpBSByydVKqStGHAnZzQofCqhDmD2bfRgc9BmZqM3ZmsdWJw4d")
	if err != nil {
		t.Fatal(err)
	}
	msg := TxMsg{
		Type:    TxTypeTransferFungible,
		Expiry:  1_759_711_416_000,
		MaxGas:  10_000_000,
		MaxData: 1000,
		GasFrom: testSenderPublicKey(t),
		Payload: MustSmallString("test-payload"),
		Msg: &TxMsgTransferFungible{
			To:      testReceiverPublicKey(t),
			TokenID: 1,
			Amount:  100_000_000,
		},
	}

	signed, err := SignAndSerializeTxMsg(msg, keys)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToUpper(hex.EncodeToString(signed))
	if expected := vectorByKind(t, "TX2").hex; got != expected {
		t.Fatalf("signed tx mismatch:\nwant %s\n got %s", expected, got)
	}

	signedHex, err := SignAndSerializeTxMsgHex(msg, keys)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ToUpper(signedHex) != vectorByKind(t, "TX2").hex {
		t.Fatalf("signed tx hex mismatch")
	}
}

func TestAddressAndTokenHelpers(t *testing.T) {
	// Address and NFT id helpers bridge public Phantasma addresses with Carbon 32-byte token addresses.
	keys, err := cryptography.FromWIF("KwPpBSByydVKqStGHAnZzQofCqhDmD2bfRgc9BmZqM3ZmsdWJw4d")
	if err != nil {
		t.Fatal(err)
	}

	fromPublicKey := MustBytes32FromPublicKey(keys.PublicKey())
	fromAddress := MustBytes32FromPhantasmaAddress(keys.Address())
	if fromAddress != fromPublicKey {
		t.Fatalf("address conversion mismatch")
	}

	nftAddress := GetNFTAddress(9, 7)
	if nftAddress[15] != 1 || nftAddress[16] != 9 || nftAddress[24] != 7 {
		t.Fatalf("unexpected NFT address encoding: %s", nftAddress.String())
	}

	seriesID, mintNumber := UnpackNFTInstanceID(0x0000000800000007)
	if seriesID != 7 || mintNumber != 8 {
		t.Fatalf("unexpected unpacked instance id: %d/%d", seriesID, mintNumber)
	}

	raw, err := BuildMintNonFungibleTxAndSign(9, 7, keys, testReceiverPublicKey(t), []byte{0xaa}, nil, MintNFTFeeOptions{}, 0, 1_759_711_416_000)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := BuildMintNonFungibleTxAndSignHex(9, 7, keys, testReceiverPublicKey(t), []byte{0xaa}, nil, MintNFTFeeOptions{}, 0, 1_759_711_416_000)
	if err != nil {
		t.Fatal(err)
	}
	if encoded != hex.EncodeToString(raw) {
		t.Fatalf("hex helper mismatch")
	}
}

func TestFeeOptionZeroValuesUseDefaults(t *testing.T) {
	// Zero-valued option structs must stay safe defaults instead of silently creating zero-gas transactions.
	symbol := MustSmallString("TOKEN")
	if got, want := (CreateTokenFeeOptions{}).CalculateMaxGas(symbol), DefaultCreateTokenFeeOptions().CalculateMaxGas(symbol); got != want {
		t.Fatalf("zero create-token fees must use defaults: got %d want %d", got, want)
	}
	if got, want := (CreateSeriesFeeOptions{}).CalculateMaxGas(), DefaultCreateSeriesFeeOptions().CalculateMaxGas(); got != want {
		t.Fatalf("zero create-series fees must use defaults: got %d want %d", got, want)
	}
	if got, want := (MintNFTFeeOptions{}).calculateMaxGasForSingle(), DefaultMintNFTFeeOptions().calculateMaxGasForSingle(); got != want {
		t.Fatalf("zero mint fees must use defaults: got %d want %d", got, want)
	}
}

func TestNFTMintFeeOptionsScaleByCount(t *testing.T) {
	baseFees := NewFeeOptions(10, 1_000)
	if got := baseFees.CalculateMaxGas(); got != 10_000 {
		t.Fatalf("single base fee mismatch: got %d", got)
	}
	if got, err := baseFees.CalculateMaxGasForCount(3); err != nil || got != 30_000 {
		t.Fatalf("counted base fee mismatch: got %d err %v", got, err)
	}

	mintFees := NewMintNFTFeeOptions(10, 1_000)
	if got, err := mintFees.CalculateMaxGasForCount(3); err != nil || got != 30_000 {
		t.Fatalf("counted mint fee mismatch: got %d err %v", got, err)
	}
	if _, err := mintFees.CalculateMaxGasForCount(0); err == nil || !strings.Contains(err.Error(), "count must be positive") {
		t.Fatalf("zero mint count should fail, got %v", err)
	}

	tokens := []PhantasmaNFTMintInfo{
		{PhantasmaSeriesID: NewIntX(big.NewInt(1)), ROM: []byte{0x01}},
		{PhantasmaSeriesID: NewIntX(big.NewInt(2)), ROM: []byte{0x02}},
		{PhantasmaSeriesID: NewIntX(big.NewInt(3)), ROM: []byte{0x03}},
	}
	tx, err := BuildMintPhantasmaNonFungibleTx(42, testSenderPublicKey(t), testReceiverPublicKey(t), tokens, mintFees, 123, 999)
	if err != nil {
		t.Fatal(err)
	}
	if tx.MaxGas != 30_000 {
		t.Fatalf("multi Phantasma NFT mint max gas mismatch: got %d", tx.MaxGas)
	}
	if _, err := BuildMintPhantasmaNonFungibleTx(42, testSenderPublicKey(t), testReceiverPublicKey(t), nil, mintFees, 123, 999); err == nil || !strings.Contains(err.Error(), "count must be positive") {
		t.Fatalf("empty Phantasma NFT mint should fail, got %v", err)
	}
}

func TestFeeOptionMethodShapes(t *testing.T) {
	// Only count-sensitive fee types should expose count-sensitive helpers.
	if _, ok := reflect.TypeOf(CreateTokenFeeOptions{}).MethodByName("CalculateMaxGasForCount"); ok {
		t.Fatalf("create-token fees must not expose count-sensitive max gas")
	}
	if _, ok := reflect.TypeOf(CreateSeriesFeeOptions{}).MethodByName("CalculateMaxGasForCount"); ok {
		t.Fatalf("create-series fees must not expose count-sensitive max gas")
	}
	if _, ok := reflect.TypeOf(MintNFTFeeOptions{}).MethodByName("CalculateMaxGas"); ok {
		t.Fatalf("NFT mint fees must require an explicit count")
	}
	if _, ok := reflect.TypeOf(MintNFTFeeOptions{}).MethodByName("CalculateMaxGasForCount"); !ok {
		t.Fatalf("NFT mint fees must expose count-sensitive max gas")
	}
}

func testReceiverPublicKey(t *testing.T) Bytes32 {
	t.Helper()
	keys, err := cryptography.FromWIF("KwVG94yjfVg1YKFyRxAGtug93wdRbmLnqqrFV6Yd2CiA9KZDAp4H")
	if err != nil {
		t.Fatal(err)
	}
	return MustBytes32(keys.PublicKey())
}
