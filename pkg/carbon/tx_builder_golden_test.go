package carbon

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
)

const carbonTxBuilderFixtureSHA256 = "efcb2d237ffd2ca3178b8c3b3106c7d035bc0f5e05959abb135163d637c3b11d"

func TestCarbonTxBuilderGoldenVectors(t *testing.T) {
	// This fixture suite pins public transaction builders, not just low-level
	// serializers, so SDK ergonomics cannot drift away from canonical bytes.
	for _, row := range carbonTxBuilderFixtureRows(t) {
		caseID, source, expectedHex := row[0], row[1], row[2]
		if source != "csharp_sdk" && source != "go_sdk" {
			t.Fatalf("%s unexpected source: %s", caseID, source)
		}
		if got := strings.ToUpper(hex.EncodeToString(carbonTxBuilderVector(t, caseID))); got != expectedHex {
			t.Fatalf("%s builder mismatch:\nwant %s\n got %s", caseID, expectedHex, got)
		}
		if row[3] == "" {
			t.Fatalf("%s missing fixture note", caseID)
		}
	}
}

func carbonTxBuilderFixtureRows(t *testing.T) [][]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "carbon_tx_builder_vectors.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != carbonTxBuilderFixtureSHA256 {
		t.Fatalf("carbon_tx_builder_vectors.tsv hash mismatch: want %s got %s", carbonTxBuilderFixtureSHA256, got)
	}

	rows := [][]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" || strings.HasPrefix(line, "case_id\t") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 4 {
			t.Fatalf("bad fixture row: %q", line)
		}
		rows = append(rows, parts)
	}
	return rows
}

func carbonTxBuilderVector(t *testing.T, caseID string) []byte {
	t.Helper()
	// Deterministic, non-funded fixture key used only to reproduce signed
	// Carbon transaction golden vectors.
	keys, err := cryptography.FromWIF("KwPpBSByydVKqStGHAnZzQofCqhDmD2bfRgc9BmZqM3ZmsdWJw4d")
	if err != nil {
		t.Fatal(err)
	}
	sender := MustBytes32(keys.PublicKey())
	receiver := testReceiverPublicKey(t)

	switch caseID {
	case "signed_transfer_fungible":
		msg := baseVectorTx(TxTypeTransferFungible, sender)
		msg.Msg = &TxMsgTransferFungible{To: receiver, TokenID: 1, Amount: 100_000_000}
		out, err := SignAndSerializeTxMsg(msg, keys)
		if err != nil {
			t.Fatal(err)
		}
		return out
	case "transfer_fungible_gas_payer":
		msg := baseVectorTx(TxTypeTransferFungibleGasPayer, sender)
		msg.Msg = &TxMsgTransferFungibleGasPayer{To: receiver, From: sender, TokenID: 1, Amount: 100_000_000}
		return Serialize(&msg)
	case "burn_fungible_gas_payer":
		msg := baseVectorTx(TxTypeBurnFungibleGasPayer, sender)
		msg.Msg = &TxMsgBurnFungibleGasPayer{TokenID: 1, From: sender, Amount: IntXFromInt64(100_000_000)}
		return Serialize(&msg)
	case "mint_fungible":
		msg := baseVectorTx(TxTypeMintFungible, sender)
		msg.Msg = &TxMsgMintFungible{TokenID: 1, To: receiver, Amount: IntXFromInt64(100_000_000)}
		return Serialize(&msg)
	case "create_token_nft":
		schemas := PrepareStandardTokenSchemas(false)
		tokenInfo, err := BuildTokenInfo(
			"MYNFT",
			IntXFromInt64(0),
			true,
			0,
			sender,
			MustBuildTokenMetadata(map[string]string{
				"name":        "My test token!",
				"icon":        samplePNGIconDataURI,
				"url":         "http://example.com",
				"description": "My test token description",
			}),
			SerializeTokenSchemas(schemas),
		)
		if err != nil {
			t.Fatal(err)
		}
		msg := BuildCreateTokenTx(tokenInfo, sender, DefaultCreateTokenFeeOptions(), 100_000_000, 1_759_711_416_000)
		return Serialize(&msg)
	case "create_token_series_u256_id":
		seriesInfo, err := BuildSeriesInfo(maxUint256(), 0, 0, sender)
		if err != nil {
			t.Fatal(err)
		}
		msg := BuildCreateTokenSeriesTx(^uint64(0), seriesInfo, sender, DefaultCreateSeriesFeeOptions(), 100_000_000, 1_759_711_416_000)
		return Serialize(&msg)
	case "mint_non_fungible_u256_nft_id":
		schemas := PrepareStandardTokenSchemas(false)
		rom, err := BuildNFTRom(schemas.ROM, maxUint256(), nftGoldenMetadata(true))
		if err != nil {
			t.Fatal(err)
		}
		msg := BuildMintNonFungibleTx(^uint64(0), ^uint32(0), sender, sender, rom, nil, DefaultMintNFTFeeOptions(), 100_000_000, 1_759_711_416_000)
		return Serialize(&msg)
	case "mint_phantasma_nft_single_u255_series":
		schemas := PrepareStandardTokenSchemas(false)
		publicRom, err := BuildPhantasmaNFTRom(schemas.ROM, nftGoldenMetadata(false))
		if err != nil {
			t.Fatal(err)
		}
		seriesID := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(1))
		msg, err := BuildMintPhantasmaNonFungibleSingleTx(42, seriesID, sender, receiver, publicRom, nil, DefaultMintNFTFeeOptions(), 123, 1_759_711_416_000)
		if err != nil {
			t.Fatal(err)
		}
		return Serialize(&msg)
	default:
		t.Fatalf("unhandled Carbon tx builder vector: %s", caseID)
		return nil
	}
}

func baseVectorTx(txType TxType, gasFrom Bytes32) TxMsg {
	return TxMsg{
		Type:    txType,
		Expiry:  1_759_711_416_000,
		MaxGas:  10_000_000,
		MaxData: 1000,
		GasFrom: gasFrom,
		Payload: MustSmallString("test-payload"),
	}
}

func nftGoldenMetadata(includeRawROM bool) []MetadataField {
	fields := []MetadataField{
		{Name: "name", Value: "My NFT #1"},
		{Name: "description", Value: "This is my first NFT!"},
		{Name: "imageURL", Value: "images-assets.nasa.gov/image/PIA13227/PIA13227~orig.jpg"},
		{Name: "infoURL", Value: "https://images.nasa.gov/details/PIA13227"},
		{Name: "royalties", Value: int32(10_000_000)},
	}
	if includeRawROM {
		fields = append(fields, MetadataField{Name: "rom", Value: []byte{0x01, 0x42}})
	}
	return fields
}
