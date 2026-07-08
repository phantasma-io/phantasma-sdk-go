package response

import (
	"encoding/json"
	"testing"
)

func TestAccountCloneAndTokenBalance(t *testing.T) {
	relay := "100"
	account := AccountResult{
		Relay:    &relay,
		Balances: []BalanceResult{{Chain: "main", Amount: "10", Symbol: "SOUL", Decimals: 8, Ids: []string{"1"}}},
		Txs:      []string{"tx1"},
	}

	clone := account.Clone()
	*clone.Relay = "200"
	clone.Balances[0].Amount = "20"
	clone.Balances[0].Ids[0] = "2"
	clone.Txs[0] = "tx2"

	if *account.Relay != "100" || account.Balances[0].Amount != "10" || account.Balances[0].Ids[0] != "1" || account.Txs[0] != "tx1" {
		t.Fatalf("clone must not alias nested slices")
	}

	balance := account.GetTokenBalance(TokenResult{Symbol: "KCAL", Decimals: 10})
	if balance.Symbol != "KCAL" || balance.Amount != "0" || balance.Decimals != 10 {
		t.Fatalf("unexpected synthetic balance: %+v", balance)
	}
}

func TestTokenResultFlagHelpers(t *testing.T) {
	token := TokenResult{Flags: "Fungible, Transferable, Burnable"}
	if !token.IsFungible() || !token.IsTransferable() || !token.IsBurnable() {
		t.Fatalf("expected token flags to be detected")
	}
	if token.IsMintable() || token.IsFuel() {
		t.Fatalf("unexpected token flag detected")
	}
}

func TestScriptResultCheckedDecode(t *testing.T) {
	// Result payloads returned by invokeRawScript are hex-encoded VM objects.
	result := ScriptResult{
		Result:  "0601",
		Results: []string{"0600"},
	}

	value, err := result.DecodeResultWithError()
	if err != nil {
		t.Fatalf("DecodeResultWithError failed: %v", err)
	}
	if value.AsString() != "true" {
		t.Fatalf("expected true VM bool, got %s", value.AsString())
	}

	indexed, err := result.DecodeResultsWithError(0)
	if err != nil {
		t.Fatalf("DecodeResultsWithError failed: %v", err)
	}
	if indexed.AsString() != "false" {
		t.Fatalf("expected false VM bool, got %s", indexed.AsString())
	}
}

func TestScriptResultCheckedDecodeRejectsBadInput(t *testing.T) {
	result := ScriptResult{Result: "not-hex"}
	if _, err := result.DecodeResultWithError(); err == nil {
		t.Fatalf("expected invalid hex to fail")
	}
	if _, err := result.DecodeResultsWithError(0); err == nil {
		t.Fatalf("expected missing indexed result to fail")
	}
}

func TestRPCDTOsDecodeCurrentResponseShapes(t *testing.T) {
	debugComment := "mint"
	rawTransaction := []byte(`{
		"hash": "HASH",
		"chainAddress": "CHAIN",
		"timestamp": 123,
		"blockHeight": 456,
		"blockHash": "BLOCK",
		"script": "",
		"payload": "CAFE",
		"carbonTxType": 9,
		"carbonTxData": "BEEF",
		"debugComment": "mint",
		"events": [{"address": "Pevent", "contract": "gas", "kind": "GasEscrow", "name": "GasEscrow", "data": "00"}],
		"extendedEvents": [{"address": "Pevent", "contract": "market", "kind": "Order", "data": {"id": "7"}}],
		"state": "Halt",
		"result": "",
		"fee": "2600000",
		"signatures": [{"kind": "Ed25519", "data": "AA"}],
		"sender": "Psender",
		"gasPayer": "Pgas",
		"gasTarget": "Ptarget",
		"gasPrice": "1",
		"gasLimit": "100000000",
		"expiration": 789
	}`)
	var tx TransactionResult
	if err := json.Unmarshal(rawTransaction, &tx); err != nil {
		t.Fatalf("transaction decode failed: %v", err)
	}
	if tx.CarbonTxType != 9 || tx.CarbonTxData != "BEEF" || tx.DebugComment == nil || *tx.DebugComment != debugComment {
		t.Fatalf("carbon transaction fields lost: %+v", tx)
	}
	if tx.Sender != "Psender" || tx.GasPayer != "Pgas" || tx.GasTarget != "Ptarget" || tx.GasPrice != "1" || tx.GasLimit != "100000000" {
		t.Fatalf("gas/sender fields lost: %+v", tx)
	}
	if len(tx.Events) != 1 || tx.Events[0].Name != "GasEscrow" || tx.Events[0].Kind != "GasEscrow" {
		t.Fatalf("event fields lost: %+v", tx.Events)
	}
	if len(tx.ExtendedEvents) != 1 || tx.ExtendedEvents[0].Kind != "Order" {
		t.Fatalf("extended events lost: %+v", tx.ExtendedEvents)
	}
	if len(tx.Signatures) != 1 || tx.Signatures[0].Kind != "Ed25519" || tx.Signatures[0].Data != "AA" {
		t.Fatalf("signature fields lost: %+v", tx.Signatures)
	}

	var block BlockResult
	if err := json.Unmarshal([]byte(`{"hash":"BLOCK","height":456,"txs":[`+string(rawTransaction)+`],"reward":"0"}`), &block); err != nil {
		t.Fatalf("block decode failed: %v", err)
	}
	if block.Events != nil || block.Oracles != nil {
		t.Fatalf("omitted block arrays must remain nil: %+v", block)
	}
	if len(block.Txs) != 1 || block.Txs[0].CarbonTxType != 9 {
		t.Fatalf("nested transaction fields lost: %+v", block.Txs)
	}
	// Pre-gas-model-v2 block: the producerAddress key is omitted, so the pointer stays nil.
	if block.ProducerAddress != nil {
		t.Fatalf("omitted producerAddress must decode to nil: %+v", block.ProducerAddress)
	}

	// Gas-model-v2 block: producerAddress is present and decodes verbatim; distinct in meaning
	// from ValidatorAddress (the consensus-log leader).
	var v2Block BlockResult
	if err := json.Unmarshal([]byte(`{"hash":"BLOCK","height":457,"txs":[],"reward":"0","validatorAddress":"Pvalidator","producerAddress":"Pproducer"}`), &v2Block); err != nil {
		t.Fatalf("v2 block decode failed: %v", err)
	}
	if v2Block.ValidatorAddress != "Pvalidator" {
		t.Fatalf("validatorAddress lost: %+v", v2Block.ValidatorAddress)
	}
	if v2Block.ProducerAddress == nil || *v2Block.ProducerAddress != "Pproducer" {
		t.Fatalf("producerAddress not decoded: %+v", v2Block.ProducerAddress)
	}

	var token TokenResult
	if err := json.Unmarshal([]byte(`{
		"symbol":"CROWN",
		"name":"Crown",
		"decimals":0,
		"currentSupply":"1",
		"maxSupply":"0",
		"burnedSupply":"0",
		"address":"S-token",
		"owner":"Powner",
		"flags":"Transferable, NonFungible",
		"carbonId":"4",
		"metadata":[{"key":"name","value":"Crown"}],
		"series":[{"seriesId":"0","carbonTokenId":"4","carbonSeriesId":"1"}]
	}`), &token); err != nil {
		t.Fatalf("token decode failed: %v", err)
	}
	if token.CarbonID != "4" || token.Metadata[0].Key != "name" || token.Metadata[0].Value != "Crown" {
		t.Fatalf("token fields lost: %+v", token)
	}
	if token.Script != nil || token.External != nil || token.Price != nil {
		t.Fatalf("omitted token optional fields must remain nil: %+v", token)
	}
	if token.Series[0].SeriesID != "0" || token.Series[0].CarbonSeriesID != "1" {
		t.Fatalf("series identity fields lost: %+v", token.Series[0])
	}

	var nft TokenDataResult
	if err := json.Unmarshal([]byte(`{
		"id":"114421",
		"series":"0",
		"carbonTokenId":"4",
		"carbonSeriesId":"1",
		"carbonNftAddress":"ABCDEF",
		"mint":"1",
		"chainName":"main",
		"ownerAddress":"Powner",
		"creatorAddress":"Pcreator",
		"ram":"",
		"rom":"CAFE",
		"status":"Active",
		"infusion":[],
		"properties":[{"key":"name","value":"Crown #1"}]
	}`), &nft); err != nil {
		t.Fatalf("NFT decode failed: %v", err)
	}
	if nft.ID != "114421" || nft.Series != "0" || nft.CarbonSeriesID != "1" || nft.Properties[0].Value != "Crown #1" {
		t.Fatalf("NFT identity fields lost: %+v", nft)
	}

	var chain ChainResult
	if err := json.Unmarshal([]byte(`{"height":0}`), &chain); err != nil {
		t.Fatalf("chain decode failed: %v", err)
	}
	if chain.Name != nil || chain.Contracts != nil || chain.Height != 0 {
		t.Fatalf("stub chain optional fields must remain nil: %+v", chain)
	}

	var archive ArchiveResult
	if err := json.Unmarshal([]byte(`{"time":0,"size":0,"blockCount":0}`), &archive); err != nil {
		t.Fatalf("archive decode failed: %v", err)
	}
	if archive.Name != nil || archive.MissingBlocks != nil {
		t.Fatalf("stub archive optional fields must remain nil: %+v", archive)
	}

	var script ScriptResult
	if err := json.Unmarshal([]byte(`{"events":[],"result":"0601","results":["0601"],"oracles":[]}`), &script); err != nil {
		t.Fatalf("script decode failed: %v", err)
	}
	if script.Error != nil || script.State != nil || script.Gas != nil {
		t.Fatalf("omitted script optional fields must remain nil: %+v", script)
	}
}

func TestRPCDTOsIgnoreStaleWireFieldNamesWithoutAliasMapping(t *testing.T) {
	var sig SignatureResult
	if err := json.Unmarshal([]byte(`{"Kind":"Ed25519","Data":"AA"}`), &sig); err != nil {
		t.Fatalf("signature decode failed: %v", err)
	}
	if sig.Kind != "" || sig.Data != "" {
		t.Fatalf("stale signature fields must not populate current fields: %+v", sig)
	}

	var event EventResult
	if err := json.Unmarshal([]byte(`{"address":"P","contract":"gas","Kind":"GasEscrow","Data":"00"}`), &event); err != nil {
		t.Fatalf("event decode failed: %v", err)
	}
	if event.Kind != "" || event.Data != "" {
		t.Fatalf("stale event fields must not populate current fields: %+v", event)
	}

	var extended EventExResult
	if err := json.Unmarshal([]byte(`{"address":"P","contract":"gas","Kind":"TokenCreate","Data":{"symbol":"CROWN"}}`), &extended); err != nil {
		t.Fatalf("extended event decode failed: %v", err)
	}
	if extended.Kind != "" || extended.Data != nil {
		t.Fatalf("stale extended event fields must not populate current fields: %+v", extended)
	}

	var property TokenPropertyResult
	if err := json.Unmarshal([]byte(`{"Key":"name","Value":"Crown"}`), &property); err != nil {
		t.Fatalf("token property decode failed: %v", err)
	}
	if property.Key != "" || property.Value != "" {
		t.Fatalf("stale metadata fields must not populate current fields: %+v", property)
	}

	var nft TokenDataResult
	if err := json.Unmarshal([]byte(`{"ID":"114421","series":"0"}`), &nft); err != nil {
		t.Fatalf("NFT decode failed: %v", err)
	}
	if nft.ID != "" || nft.Series != "0" {
		t.Fatalf("stale NFT ID must not populate current id field: %+v", nft)
	}

	var token TokenResult
	if err := json.Unmarshal([]byte(`{"symbol":"CROWN","carbonID":"4"}`), &token); err != nil {
		t.Fatalf("token decode failed: %v", err)
	}
	if token.CarbonID != "" || token.Symbol != "CROWN" {
		t.Fatalf("stale carbonID must not populate current carbonId field: %+v", token)
	}
}
