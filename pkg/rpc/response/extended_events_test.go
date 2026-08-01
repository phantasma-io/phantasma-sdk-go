package response

import (
	"encoding/json"
	"testing"
)

// Live fixtures were captured from https://devnet.phantasma.info/rpc on 2026-08-01 via
// getBlockByHeight("main", <height>); the height is stated on each test. Long hex payloads
// (contract scripts, ABIs, ROMs) are truncated to keep the fixtures readable - the field set, the
// field types and every other value are verbatim. Shape-only fixtures for the event kinds without
// a capturable live sample (token and market events) mirror the node's emission sites in
// RpcEventBuilder.TokenEvents.cs / RpcEventBuilder.MarketEvents.cs, whose serializer settings
// (camelCase, enum names as strings, nulls omitted) are the same ones verified live on the
// special-resolution family.

// Devnet block 8,736,259: one SpecialResolution event whose single call is
// phantasma_vm.RepairSeries with 3,649 supplements and 8,370 repairs. The fixture keeps the first
// supplement and the first repair.
const repairSeriesEventWire = `{
	"address": "P2KJPTC82NAFEzXg3X4eA83JvyWQ8PJVaBop2fUUsKPBcou",
	"contract": "governance",
	"kind": "SpecialResolution",
	"data": {
		"resolutionId": 37,
		"description": "Special Resolution",
		"calls": [{
			"moduleId": 2,
			"module": "phantasma_vm",
			"methodId": 6,
			"method": "RepairSeries",
			"arguments": {
				"supplementsCount": "3649",
				"supplements": [{
					"token": "BRC",
					"tokenId": "23",
					"phantasmaSeriesId": "6472",
					"maxSupply": "1000",
					"mintCount": "30",
					"mode": "1",
					"script": "0004010D000403524F4D0300",
					"abi": "080A67657443726561746564",
					"rom": "010804076372656174656405"
				}],
				"repairsCount": "8370",
				"repairs": [{
					"token": "CROWN",
					"tokenId": "4",
					"phantasmaSeriesId": "0",
					"importedLiveCount": "10998",
					"script": "0004000E0000040D01040743",
					"abi": "04076765744E616D65040100"
				}]
			}
		}]
	}
}`

func TestSpecialResolutionRepairSeriesDecodesFromTheDevnetAnswer(t *testing.T) {
	var event EventExResult
	if err := json.Unmarshal([]byte(repairSeriesEventWire), &event); err != nil {
		t.Fatalf("event decode failed: %v", err)
	}

	if event.Contract != "governance" || event.Kind != "SpecialResolution" {
		t.Fatalf("event envelope lost: %+v", event)
	}
	resolution, ok := event.Data.(SpecialResolutionData)
	if !ok {
		t.Fatalf("SpecialResolution data should decode to its typed shape, got %T", event.Data)
	}
	if resolution.ResolutionID != 37 {
		t.Fatalf("resolution id lost: %+v", resolution)
	}
	if resolution.Description == nil || *resolution.Description != "Special Resolution" {
		t.Fatalf("resolution description lost: %+v", resolution)
	}
	if len(resolution.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(resolution.Calls))
	}

	call := resolution.Calls[0]
	if call.ModuleID != 2 || call.Module != "phantasma_vm" || call.MethodID != 6 || call.Method != "RepairSeries" {
		t.Fatalf("call identity lost: %+v", call)
	}
	if call.Calls != nil {
		t.Fatalf("a non-nested call must carry no nested calls: %+v", call.Calls)
	}

	arguments, ok := call.Arguments.(RepairSeriesArguments)
	if !ok {
		t.Fatalf("RepairSeries arguments should decode to their typed shape, got %T", call.Arguments)
	}
	if arguments.SupplementsCount != "3649" || arguments.RepairsCount != "8370" {
		t.Fatalf("repair counts lost: %+v", arguments)
	}
	supplement := arguments.Supplements[0]
	if supplement.Token != "BRC" || supplement.TokenID != "23" || supplement.PhantasmaSeriesID != "6472" ||
		supplement.MaxSupply != "1000" || supplement.MintCount != "30" || supplement.Mode != "1" {
		t.Fatalf("supplement fields lost: %+v", supplement)
	}
	repair := arguments.Repairs[0]
	if repair.Token != "CROWN" || repair.TokenID != "4" || repair.PhantasmaSeriesID != "0" ||
		repair.ImportedLiveCount != "10998" {
		t.Fatalf("repair fields lost: %+v", repair)
	}
}

func TestSpecialResolutionRoundTripsToTheWireShape(t *testing.T) {
	// Serializing the decoded event must reproduce the wire object exactly: camelCase names,
	// numeric ids as numbers, string counts as strings, and no null keys for the absent nested
	// calls.
	var event EventExResult
	if err := json.Unmarshal([]byte(repairSeriesEventWire), &event); err != nil {
		t.Fatalf("event decode failed: %v", err)
	}

	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("event encode failed: %v", err)
	}
	assertSameJSON(t, []byte(repairSeriesEventWire), encoded)
}

func TestSpecialResolutionTransferFungibleDecodesFromTheDevnetAnswer(t *testing.T) {
	// Devnet block 8,736,266: resolution 44 "Repair imported NFT fungible infusions" carries 9,600
	// token.TransferFungible calls; this is its first call verbatim.
	var event EventExResult
	if err := json.Unmarshal([]byte(`{
		"address": "P2KJPTC82NAFEzXg3X4eA83JvyWQ8PJVaBop2fUUsKPBcou",
		"contract": "governance",
		"kind": "SpecialResolution",
		"data": {
			"resolutionId": 44,
			"description": "Repair imported NFT fungible infusions",
			"calls": [{
				"moduleId": 1,
				"module": "token",
				"methodId": 0,
				"method": "TransferFungible",
				"arguments": {
					"from": "S3dPnV8dfdkHDHDcJiHY255FEUZCM7oAmDW78LpYZ4jveGW",
					"to": "S3dPnV8dfdkHDHDcJiHY255FEUZCM7oAmDW78LpYZ4jveGW",
					"amount": "10000000000",
					"token": "KCAL",
					"tokenId": "1"
				}
			}]
		}
	}`), &event); err != nil {
		t.Fatalf("event decode failed: %v", err)
	}

	resolution, ok := event.Data.(SpecialResolutionData)
	if !ok {
		t.Fatalf("SpecialResolution data should decode to its typed shape, got %T", event.Data)
	}
	if resolution.ResolutionID != 44 {
		t.Fatalf("resolution id lost: %+v", resolution)
	}
	transfer, ok := resolution.Calls[0].Arguments.(TransferFungibleArguments)
	if !ok {
		t.Fatalf("TransferFungible arguments should decode to their typed shape, got %T", resolution.Calls[0].Arguments)
	}
	if transfer.From != "S3dPnV8dfdkHDHDcJiHY255FEUZCM7oAmDW78LpYZ4jveGW" || transfer.To != transfer.From {
		t.Fatalf("transfer parties lost: %+v", transfer)
	}
	// Token and TokenID come from the embedded identity pair, which is where the wire carries them.
	if transfer.Amount != "10000000000" || transfer.Token != "KCAL" || transfer.TokenID != "1" {
		t.Fatalf("transfer payload lost: %+v", transfer)
	}
}

func TestImportContractsDecodesFromTheDevnetAnswer(t *testing.T) {
	// Devnet block 8,736,257: phantasma_vm.ImportContracts restoring 70 contracts. The fixture
	// keeps the "mail" contract (empty storage) and the "pharming" contract's first root variable
	// and first table row. This is the shape a flat string map cannot carry at all.
	var event EventExResult
	if err := json.Unmarshal([]byte(`{
		"address": "P2KJPTC82NAFEzXg3X4eA83JvyWQ8PJVaBop2fUUsKPBcou",
		"contract": "governance",
		"kind": "SpecialResolution",
		"data": {
			"resolutionId": 36,
			"description": "Special Resolution",
			"calls": [{
				"moduleId": 2,
				"module": "phantasma_vm",
				"methodId": 5,
				"method": "ImportContracts",
				"arguments": {
					"contractsCount": "70",
					"contracts": [
						{
							"name": "mail",
							"address": "S3d6cUXRwJbudV4ADbRtMz3P9527ts7D2Lh9h2J96m48FPW",
							"owner": "P2KFNXEbt65rQiWqogAzqkVGMqFirPmqPw8mQyxvRKsrXV8",
							"script": "0B",
							"abi": "090B507573684D657373616765",
							"rootVariables": [],
							"tables": []
						},
						{
							"name": "pharming",
							"address": "S3d6cUXRwJbudV4ADbRtMz3P9527ts7D2Lh9h2J96m48FPW",
							"owner": "P2KFNXEbt65rQiWqogAzqkVGMqFirPmqPw8mQyxvRKsrXV8",
							"script": "0B",
							"abi": "0906676574546F6B656E",
							"rootVariables": [{
								"key": "6D616E61676572",
								"value": "0100E9F4F69F677473684D2E201672A6AC30CA8F2A238C68"
							}],
							"tables": [{
								"name": "addrs_kcal_bnb",
								"rows": [{
									"key": "3C003E",
									"value": "0104040B534F554C41646472657373"
								}]
							}]
						}
					]
				}
			}]
		}
	}`), &event); err != nil {
		t.Fatalf("event decode failed: %v", err)
	}

	resolution, ok := event.Data.(SpecialResolutionData)
	if !ok {
		t.Fatalf("SpecialResolution data should decode to its typed shape, got %T", event.Data)
	}
	imported, ok := resolution.Calls[0].Arguments.(ImportContractsArguments)
	if !ok {
		t.Fatalf("ImportContracts arguments should decode to their typed shape, got %T", resolution.Calls[0].Arguments)
	}
	if imported.ContractsCount != "70" || len(imported.Contracts) != 2 {
		t.Fatalf("contract list lost: %+v", imported)
	}
	if imported.Contracts[0].Name != "mail" || len(imported.Contracts[0].RootVariables) != 0 ||
		len(imported.Contracts[0].Tables) != 0 {
		t.Fatalf("empty-storage contract lost: %+v", imported.Contracts[0])
	}
	pharming := imported.Contracts[1]
	if pharming.Name != "pharming" || pharming.RootVariables[0].Key != "6D616E61676572" {
		t.Fatalf("root variables lost: %+v", pharming)
	}
	if pharming.Tables[0].Name != "addrs_kcal_bnb" || pharming.Tables[0].Rows[0].Key != "3C003E" {
		t.Fatalf("storage tables lost: %+v", pharming.Tables)
	}
}

func TestTokenCreateDataDecodesAndRoundTrips(t *testing.T) {
	// Shape per RpcEventBuilder.TokenEvents.cs:153: carbonTokenId is a JSON number, maxSupply a
	// string, and the metadata values are rendered to strings by the node - unlike the metadata of
	// a token response, which carries VM values.
	const wire = `{
		"address": "P2KFNXEbt65rQiWqogAzqkVGMqFirPmqPw8mQyxvRKsrXV8",
		"contract": "token",
		"kind": "TokenCreate",
		"data": {
			"symbol": "CROWN",
			"maxSupply": "0",
			"decimals": 0,
			"isNonFungible": true,
			"carbonTokenId": 4,
			"metadata": {"name": "Crown", "description": "Phantasma Crown"}
		}
	}`

	var event EventExResult
	if err := json.Unmarshal([]byte(wire), &event); err != nil {
		t.Fatalf("event decode failed: %v", err)
	}

	created, ok := event.Data.(TokenCreateData)
	if !ok {
		t.Fatalf("TokenCreate data should decode to its typed shape, got %T", event.Data)
	}
	if created.Symbol != "CROWN" || created.MaxSupply != "0" || created.Decimals != 0 ||
		!created.IsNonFungible || created.CarbonTokenID != 4 {
		t.Fatalf("token creation fields lost: %+v", created)
	}
	if created.Metadata["name"] != "Crown" || created.Metadata["description"] != "Phantasma Crown" {
		t.Fatalf("token metadata lost: %+v", created.Metadata)
	}

	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("event encode failed: %v", err)
	}
	assertSameJSON(t, []byte(wire), encoded)
}

func TestTokenSeriesCreateDataDecodes(t *testing.T) {
	// Shape per RpcEventBuilder.TokenEvents.cs:360: seriesId is the Phantasma id as a string while
	// the carbon ids and the mint bounds are JSON numbers.
	var event EventExResult
	if err := json.Unmarshal([]byte(`{
		"address": "P2KFNXEbt65rQiWqogAzqkVGMqFirPmqPw8mQyxvRKsrXV8",
		"contract": "token",
		"kind": "TokenSeriesCreate",
		"data": {
			"symbol": "CROWN",
			"seriesId": "6472",
			"maxMint": 100,
			"maxSupply": 1000,
			"owner": "P2KFNXEbt65rQiWqogAzqkVGMqFirPmqPw8mQyxvRKsrXV8",
			"carbonTokenId": 4,
			"carbonSeriesId": 1,
			"metadata": {"name": "Crown Series"}
		}
	}`), &event); err != nil {
		t.Fatalf("event decode failed: %v", err)
	}

	series, ok := event.Data.(TokenSeriesCreateData)
	if !ok {
		t.Fatalf("TokenSeriesCreate data should decode to its typed shape, got %T", event.Data)
	}
	if series.SeriesID != "6472" || series.MaxMint != 100 || series.MaxSupply != 1000 ||
		series.CarbonTokenID != 4 || series.CarbonSeriesID != 1 {
		t.Fatalf("series creation fields lost: %+v", series)
	}
	if series.Metadata["name"] != "Crown Series" {
		t.Fatalf("series metadata lost: %+v", series.Metadata)
	}
}

func TestMarketOrderDataDecodesForEachOrderKind(t *testing.T) {
	// Shape per RpcEventBuilder.MarketEvents.cs:403; all three order kinds share it, which is why
	// the carrying event's kind is the only way to tell them apart. On a cancel the node repeats
	// the seller in buyer, so that field is asserted as-is rather than expected to be empty.
	for _, kind := range []string{"OrderCreated", "OrderCancelled", "OrderFilled"} {
		var event EventExResult
		if err := json.Unmarshal([]byte(`{
			"address": "P2KFNXEbt65rQiWqogAzqkVGMqFirPmqPw8mQyxvRKsrXV8",
			"contract": "market",
			"kind": "`+kind+`",
			"data": {
				"baseSymbol": "CROWN",
				"quoteSymbol": "SOUL",
				"tokenId": "114421",
				"carbonBaseTokenId": 4,
				"carbonQuoteTokenId": 2,
				"carbonInstanceId": 7,
				"seller": "P2KFNXEbt65rQiWqogAzqkVGMqFirPmqPw8mQyxvRKsrXV8",
				"buyer": "P2K6hJ8LQ4dqiiTBUxKfLwCbGKRDvcnCkTQ8vHjaBGnDy5B",
				"price": "1000000000",
				"endPrice": "0",
				"startDate": 1785000000,
				"endDate": 1785600000,
				"type": "Fixed"
			}
		}`), &event); err != nil {
			t.Fatalf("%s decode failed: %v", kind, err)
		}

		order, ok := event.Data.(MarketOrderData)
		if !ok {
			t.Fatalf("%s data should decode to MarketOrderData, got %T", kind, event.Data)
		}
		if order.TokenID != "114421" || order.CarbonBaseTokenID != 4 || order.CarbonQuoteTokenID != 2 ||
			order.CarbonInstanceID != 7 {
			t.Fatalf("%s identity fields lost: %+v", kind, order)
		}
		if order.Price != "1000000000" || order.EndPrice != "0" || order.StartDate != 1785000000 ||
			order.EndDate != 1785600000 || order.Type != "Fixed" {
			t.Fatalf("%s auction fields lost: %+v", kind, order)
		}
	}
}

func TestUnknownEventKindKeepsItsPayloadVerbatim(t *testing.T) {
	// A node newer than this SDK can answer an event kind that is not modeled here. Losing that
	// payload - or failing the whole block answer over it - would be worse than handing the raw
	// JSON to the caller.
	const payload = `{"somethingNew":"7","nested":{"deep":true}}`
	var event EventExResult
	if err := json.Unmarshal([]byte(`{"address":"P","contract":"x","kind":"BrandNewKind","data":`+payload+`}`), &event); err != nil {
		t.Fatalf("event decode failed: %v", err)
	}

	unknown, ok := event.Data.(UnknownEventData)
	if !ok {
		t.Fatalf("an unmodeled kind must keep its payload, got %T", event.Data)
	}
	assertSameJSON(t, []byte(payload), unknown.JSON)
}

func TestMismatchedKnownKindFallsBackToTheRawPayload(t *testing.T) {
	// The kind names a modeled shape but the payload does not match it: keeping the JSON is what
	// lets a consumer detect that the node's shape drifted, instead of reading a half-empty struct
	// as if it were complete.
	const payload = `{"symbol":"CROWN","carbonTokenId":"not-a-number"}`
	var event EventExResult
	if err := json.Unmarshal([]byte(`{"address":"P","contract":"token","kind":"TokenCreate","data":`+payload+`}`), &event); err != nil {
		t.Fatalf("event decode failed: %v", err)
	}

	unknown, ok := event.Data.(UnknownEventData)
	if !ok {
		t.Fatalf("a mismatched payload must fall back to the raw shape, got %T", event.Data)
	}
	assertSameJSON(t, []byte(payload), unknown.JSON)
}

func TestExtendedEventWithoutDataDecodesToNoPayload(t *testing.T) {
	var event EventExResult
	if err := json.Unmarshal([]byte(`{"address":"P","contract":"gas","kind":"GasEscrow"}`), &event); err != nil {
		t.Fatalf("event decode failed: %v", err)
	}
	if event.Data != nil {
		t.Fatalf("an event without data must carry no payload: %+v", event.Data)
	}
}

func TestSpecialResolutionEnvelopeToleratesMissingFields(t *testing.T) {
	// Response models are default-tolerant across this package; an empty data object decodes to
	// the empty resolution instead of failing the transaction it belongs to.
	var event EventExResult
	if err := json.Unmarshal([]byte(`{"address":"P","contract":"governance","kind":"SpecialResolution","data":{}}`), &event); err != nil {
		t.Fatalf("event decode failed: %v", err)
	}

	resolution, ok := event.Data.(SpecialResolutionData)
	if !ok {
		t.Fatalf("SpecialResolution data should decode to its typed shape, got %T", event.Data)
	}
	if resolution.ResolutionID != 0 || resolution.Description != nil || len(resolution.Calls) != 0 {
		t.Fatalf("empty resolution expected, got %+v", resolution)
	}
}
