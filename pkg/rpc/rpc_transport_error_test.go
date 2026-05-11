package rpc

import (
	"context"
	"errors"
	"testing"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/jsonrpc"
	resp "github.com/phantasma-io/phantasma-sdk-go/pkg/rpc/response"
	"github.com/stretchr/testify/require"
)

type stubRPCClient struct {
	response *jsonrpc.RPCResponse
	err      error
}

func (c stubRPCClient) Call(ctx context.Context, method string, params ...interface{}) (*jsonrpc.RPCResponse, error) {
	return c.response, c.err
}

func (c stubRPCClient) CallRaw(ctx context.Context, request *jsonrpc.RPCRequest) (*jsonrpc.RPCResponse, error) {
	panic("unexpected CallRaw")
}

func (c stubRPCClient) CallFor(ctx context.Context, out interface{}, method string, params ...interface{}) error {
	panic("unexpected CallFor")
}

func (c stubRPCClient) CallBatch(ctx context.Context, requests jsonrpc.RPCRequests) (jsonrpc.RPCResponses, error) {
	panic("unexpected CallBatch")
}

func (c stubRPCClient) CallBatchRaw(ctx context.Context, requests jsonrpc.RPCRequests) (jsonrpc.RPCResponses, error) {
	panic("unexpected CallBatchRaw")
}

type recordingRPCClient struct {
	method   string
	params   []interface{}
	response *jsonrpc.RPCResponse
}

func (c *recordingRPCClient) Call(ctx context.Context, method string, params ...interface{}) (*jsonrpc.RPCResponse, error) {
	c.method = method
	c.params = append([]interface{}{}, params...)
	return c.response, nil
}

func (c *recordingRPCClient) CallRaw(ctx context.Context, request *jsonrpc.RPCRequest) (*jsonrpc.RPCResponse, error) {
	panic("unexpected CallRaw")
}

func (c *recordingRPCClient) CallFor(ctx context.Context, out interface{}, method string, params ...interface{}) error {
	panic("unexpected CallFor")
}

func (c *recordingRPCClient) CallBatch(ctx context.Context, requests jsonrpc.RPCRequests) (jsonrpc.RPCResponses, error) {
	panic("unexpected CallBatch")
}

func (c *recordingRPCClient) CallBatchRaw(ctx context.Context, requests jsonrpc.RPCRequests) (jsonrpc.RPCResponses, error) {
	panic("unexpected CallBatchRaw")
}

func TestGetAccountRejectsNilRPCResponseWithoutPanic(t *testing.T) {
	client := PhantasmaRPC{client: stubRPCClient{}}

	var err error
	require.NotPanics(t, func() {
		_, err = client.GetAccount("P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP")
	})
	require.Error(t, err)
}

func TestLookupNameCallsLookupNameAndReturnsAddress(t *testing.T) {
	recorder := &recordingRPCClient{
		response: &jsonrpc.RPCResponse{Result: "P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP"},
	}
	client := PhantasmaRPC{client: recorder}

	address, err := client.LookupName("anonymous")

	require.NoError(t, err)
	require.Equal(t, "P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP", address)
	require.Equal(t, "lookUpName", recorder.method)
	require.Equal(t, []interface{}{"anonymous"}, recorder.params)
}

func TestGetBlockHeightRejectsInvalidHeightWithoutNilResult(t *testing.T) {
	// RPC height parsing must return an explicit error instead of a nil *big.Int with nil error.
	client := PhantasmaRPC{client: stubRPCClient{response: &jsonrpc.RPCResponse{Result: "not-a-number"}}}

	height, err := client.GetBlockHeight("main")

	require.Error(t, err)
	require.NotNil(t, height)
	require.Zero(t, height.Sign())
}

func TestSignAndSendTransactionRejectsNilKeyWithoutPanic(t *testing.T) {
	client := PhantasmaRPC{client: stubRPCClient{}}

	var err error
	require.NotPanics(t, func() {
		_, err = client.SignAndSendTransaction(nil, "testnet", []byte{0x01}, "main", nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "key pair")
}

func TestRPCWrapperParameterParity(t *testing.T) {
	objectResult := map[string]interface{}{}
	arrayResult := []interface{}{}
	cursorResult := map[string]interface{}{
		"result": []interface{}{},
		"cursor": "next",
	}
	paginatedArrayResult := map[string]interface{}{
		"page":       1,
		"pageSize":   10,
		"total":      0,
		"totalPages": 0,
		"result":     []interface{}{},
	}

	tests := []struct {
		name     string
		response interface{}
		call     func(PhantasmaRPC) error
		method   string
		params   []interface{}
	}{
		{
			name:     "GetAccountsWithAddressType",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountsWithAddressType([]string{"Pone", "Ptwo"}, true, true, AddressTypeCarbon)
				return err
			},
			method: "getAccounts",
			params: []interface{}{"Pone,Ptwo", true, true, AddressTypeCarbon},
		},
		{
			name:     "GetAccountWithAddressType",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountWithAddressType("Paccount", true, false, AddressTypeCarbon)
				return err
			},
			method: "getAccount",
			params: []interface{}{"Paccount", true, false, AddressTypeCarbon},
		},
		{
			name:     "GetBlockTransactionCountByHash",
			response: "7",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockTransactionCountByHash("block-hash")
				return err
			},
			method: "getBlockTransactionCountByHash",
			params: []interface{}{"main", "block-hash"},
		},
		{
			name:     "GetBlockTransactionCountByHashOnChain",
			response: "7",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockTransactionCountByHashOnChain("custom", "block-hash")
				return err
			},
			method: "getBlockTransactionCountByHash",
			params: []interface{}{"custom", "block-hash"},
		},
		{
			name:     "GetBlockByHash",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockByHash("block-hash")
				return err
			},
			method: "getBlockByHash",
			params: []interface{}{"block-hash"},
		},
		{
			name:     "GetLatestBlock",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetLatestBlock("main")
				return err
			},
			method: "getLatestBlock",
			params: []interface{}{"main"},
		},
		{
			name:     "GetTransactionByBlockHashAndIndex",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTransactionByBlockHashAndIndex("block-hash", 2)
				return err
			},
			method: "getTransactionByBlockHashAndIndex",
			params: []interface{}{"main", "block-hash", 2},
		},
		{
			name:     "GetTransactionByBlockHashAndIndexOnChain",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTransactionByBlockHashAndIndexOnChain("custom", "block-hash", 3)
				return err
			},
			method: "getTransactionByBlockHashAndIndex",
			params: []interface{}{"custom", "block-hash", 3},
		},
		{
			name:     "GetChains",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetChains(true)
				return err
			},
			method: "getChains",
			params: []interface{}{true},
		},
		{
			name:     "GetChain",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetChain("main", true)
				return err
			},
			method: "getChain",
			params: []interface{}{"main", true},
		},
		{
			name:     "GetNexus",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetNexus(true)
				return err
			},
			method: "getNexus",
			params: []interface{}{true},
		},
		{
			name:     "GetContracts",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetContracts("main", true)
				return err
			},
			method: "getContracts",
			params: []interface{}{"main", true},
		},
		{
			name:     "GetContractByName",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetContractByName("main", "gas")
				return err
			},
			method: "getContract",
			params: []interface{}{"main", "gas"},
		},
		{
			name:     "GetContractByAddress",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetContractByAddress("main", "Pcontract")
				return err
			},
			method: "getContractByAddress",
			params: []interface{}{"main", "Pcontract"},
		},
		{
			name:     "GetOrganization",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetOrganization("org-id", true)
				return err
			},
			method: "getOrganization",
			params: []interface{}{"org-id", true},
		},
		{
			name:     "GetOrganizationByName",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetOrganizationByName("validators", true)
				return err
			},
			method: "getOrganizationByName",
			params: []interface{}{"validators", true},
		},
		{
			name:     "GetOrganizations",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetOrganizations(true)
				return err
			},
			method: "getOrganizations",
			params: []interface{}{true},
		},
		{
			name:     "GetLeaderboard",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetLeaderboard("top")
				return err
			},
			method: "getLeaderboard",
			params: []interface{}{"top"},
		},
		{
			name:     "SendCarbonTransaction",
			response: "tx-hash",
			call: func(client PhantasmaRPC) error {
				_, err := client.SendCarbonTransaction("00")
				return err
			},
			method: "sendCarbonTransaction",
			params: []interface{}{"00"},
		},
		{
			name:     "GetTokensByOwner",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokensByOwner(true, "Powner")
				return err
			},
			method: "getTokens",
			params: []interface{}{true, "Powner"},
		},
		{
			name:     "GetTokensByOwnerWithAddressType",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokensByOwnerWithAddressType(true, "Powner", AddressTypeCarbon)
				return err
			},
			method: "getTokens",
			params: []interface{}{true, "Powner", AddressTypeCarbon},
		},
		{
			name:     "GetTokenWithID",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenWithID("NFT", true, 42)
				return err
			},
			method: "getToken",
			params: []interface{}{"NFT", true, uint64(42)},
		},
		// Carbon id filters use numeric zero as the default value accepted by current RPC nodes.
		{
			name:     "GetTokenWithIDDefaultCarbonID",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenWithID("NFT", true, 0)
				return err
			},
			method: "getToken",
			params: []interface{}{"NFT", true, uint64(0)},
		},
		{
			name:     "GetTokenData",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenData("NFT", "100")
				return err
			},
			method: "getTokenData",
			params: []interface{}{"NFT", "100"},
		},
		{
			name:     "GetTokenBalance",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenBalance("Paccount", "SOUL", "main")
				return err
			},
			method: "getTokenBalance",
			params: []interface{}{"Paccount", "SOUL", "main"},
		},
		{
			name:     "GetTokenBalanceChecked",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenBalanceChecked("Paccount", "SOUL", "main", true)
				return err
			},
			method: "getTokenBalance",
			params: []interface{}{"Paccount", "SOUL", "main", true},
		},
		{
			name:     "GetTokenBalanceWithAddressType",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenBalanceWithAddressType("Paccount", "SOUL", "main", true, AddressTypeCarbon)
				return err
			},
			method: "getTokenBalance",
			params: []interface{}{"Paccount", "SOUL", "main", true, AddressTypeCarbon},
		},
		{
			name:     "GetTokenSeries",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenSeries("NFT", 42, 10, "cursor")
				return err
			},
			method: "getTokenSeries",
			params: []interface{}{"NFT", uint64(42), 10, "cursor"},
		},
		{
			name:     "GetTokenSeriesDefaultCarbonID",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenSeries("NFT", 0, 10, "cursor")
				return err
			},
			method: "getTokenSeries",
			params: []interface{}{"NFT", uint64(0), 10, "cursor"},
		},
		{
			name:     "GetTokenSeriesByID",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenSeriesByID("NFT", 42, "9", 3)
				return err
			},
			method: "getTokenSeriesById",
			params: []interface{}{"NFT", uint64(42), "9", uint32(3)},
		},
		{
			name:     "GetTokenNFTs",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenNFTs(42, 3, 10, "cursor", true)
				return err
			},
			method: "getTokenNFTs",
			params: []interface{}{uint64(42), uint32(3), 10, "cursor", true, ""},
		},
		{
			name:     "GetTokenNFTsWithSeriesID",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenNFTsWithSeriesID(42, 0, "9", 10, "cursor", true)
				return err
			},
			method: "getTokenNFTs",
			params: []interface{}{uint64(42), uint32(0), 10, "cursor", true, "9"},
		},
		{
			name:     "GetAccountFungibleTokens",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountFungibleTokens("Paccount", "SOUL", 1, 10, "cursor", true)
				return err
			},
			method: "getAccountFungibleTokens",
			params: []interface{}{"Paccount", "SOUL", uint64(1), 10, "cursor", true},
		},
		{
			name:     "GetAccountFungibleTokensWithAddressType",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountFungibleTokensWithAddressType("Paccount", "SOUL", 1, 10, "cursor", true, AddressTypeCarbon)
				return err
			},
			method: "getAccountFungibleTokens",
			params: []interface{}{"Paccount", "SOUL", uint64(1), 10, "cursor", true, AddressTypeCarbon},
		},
		{
			name:     "GetAccountNFTs",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountNFTs("Paccount", "NFT", 42, 3, 10, "cursor", true, true)
				return err
			},
			method: "getAccountNFTs",
			params: []interface{}{"Paccount", "NFT", uint64(42), uint32(3), 10, "cursor", true, true},
		},
		{
			name:     "GetAccountNFTsWithAddressType",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountNFTsWithAddressType("Paccount", "NFT", 42, 3, 10, "cursor", true, true, AddressTypeCarbon)
				return err
			},
			method: "getAccountNFTs",
			params: []interface{}{"Paccount", "NFT", uint64(42), uint32(3), 10, "cursor", true, true, AddressTypeCarbon},
		},
		{
			name:     "GetAccountOwnedTokens",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountOwnedTokens("Paccount", "NFT", 42, 10, "cursor", true)
				return err
			},
			method: "getAccountOwnedTokens",
			params: []interface{}{"Paccount", "NFT", uint64(42), 10, "cursor", true},
		},
		{
			name:     "GetAccountOwnedTokensWithAddressType",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountOwnedTokensWithAddressType("Paccount", "NFT", 42, 10, "cursor", true, AddressTypeCarbon)
				return err
			},
			method: "getAccountOwnedTokens",
			params: []interface{}{"Paccount", "NFT", uint64(42), 10, "cursor", true, AddressTypeCarbon},
		},
		{
			name:     "GetAccountOwnedTokenSeries",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountOwnedTokenSeries("Paccount", "NFT", 42, 10, "cursor", true)
				return err
			},
			method: "getAccountOwnedTokenSeries",
			params: []interface{}{"Paccount", "NFT", uint64(42), 10, "cursor", true},
		},
		{
			name:     "GetAccountOwnedTokenSeriesWithAddressType",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountOwnedTokenSeriesWithAddressType("Paccount", "NFT", 42, 10, "cursor", true, AddressTypeCarbon)
				return err
			},
			method: "getAccountOwnedTokenSeries",
			params: []interface{}{"Paccount", "NFT", uint64(42), 10, "cursor", true, AddressTypeCarbon},
		},
		{
			name:     "GetAuctionsCount",
			response: "7",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAuctionsCount("main", "NFT")
				return err
			},
			method: "getAuctionsCount",
			params: []interface{}{"main", "NFT"},
		},
		{
			name:     "GetAuctions",
			response: paginatedArrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAuctions("main", "NFT", 1, 10)
				return err
			},
			method: "getAuctions",
			params: []interface{}{"main", "NFT", 1, 10},
		},
		{
			name:     "GetAuction",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAuction("main", "NFT", "100")
				return err
			},
			method: "getAuction",
			params: []interface{}{"main", "NFT", "100"},
		},
		{
			name:     "GetNFT",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetNFT("NFT", "100", true)
				return err
			},
			method: "getNFT",
			params: []interface{}{"NFT", "100", true},
		},
		{
			name:     "GetNFTs",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetNFTs("NFT", []string{"100", "101"}, true)
				return err
			},
			method: "getNFTs",
			params: []interface{}{"NFT", "100,101", true},
		},
		{
			name:     "GetArchive",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetArchive("archive-hash")
				return err
			},
			method: "getArchive",
			params: []interface{}{"archive-hash"},
		},
		{
			name:     "WriteArchive",
			response: "true",
			call: func(client PhantasmaRPC) error {
				_, err := client.WriteArchive("archive-hash", 4, []byte("block"))
				return err
			},
			method: "writeArchive",
			params: []interface{}{"archive-hash", 4, "YmxvY2s="},
		},
		{
			name:     "ReadArchive",
			response: "YmxvY2s=",
			call: func(client PhantasmaRPC) error {
				_, err := client.ReadArchive("archive-hash", 4)
				return err
			},
			method: "readArchive",
			params: []interface{}{"archive-hash", 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := &recordingRPCClient{response: &jsonrpc.RPCResponse{Result: tt.response}}
			client := PhantasmaRPC{client: recorder}

			require.NoError(t, tt.call(client))
			require.Equal(t, tt.method, recorder.method)
			require.Equal(t, tt.params, recorder.params)
		})
	}
}

func TestCarbonResponseFieldsDecode(t *testing.T) {
	result := &jsonrpc.RPCResponse{Result: map[string]interface{}{
		"symbol":   "ART",
		"carbonId": "42",
		"metadata": []interface{}{map[string]interface{}{
			"key":   "name",
			"value": "Carbon Art",
		}},
		"tokenSchemas": map[string]interface{}{
			"seriesMetadata": map[string]interface{}{"flags": 0, "fields": []interface{}{}},
			"rom":            map[string]interface{}{"flags": 0, "fields": []interface{}{}},
			"ram":            map[string]interface{}{"flags": 0, "fields": []interface{}{}},
		},
		"series": []interface{}{map[string]interface{}{
			"seriesId":       "9",
			"carbonTokenId":  "42",
			"carbonSeriesId": "3",
			"ownerAddress":   "Powner",
			"maxMint":        "100",
			"mintCount":      "7",
			"metadata": []interface{}{map[string]interface{}{
				"key":   "series",
				"value": "Genesis",
			}},
		}},
	}}

	var token resp.TokenResult
	require.NoError(t, result.GetObject(&token))

	require.Equal(t, "ART", token.Symbol)
	require.Equal(t, "42", token.CarbonID)
	require.NotNil(t, token.TokenSchemas)
	require.Equal(t, "name", token.Metadata[0].Key)
	require.Equal(t, "Carbon Art", token.Metadata[0].Value)
	require.Equal(t, "9", token.Series[0].SeriesID)
	require.Equal(t, "42", token.Series[0].CarbonTokenID)
	require.Equal(t, "3", token.Series[0].CarbonSeriesID)
	require.Equal(t, "Powner", token.Series[0].OwnerAddress)
	require.Equal(t, "100", token.Series[0].MaxMint)
	require.Equal(t, "7", token.Series[0].MintCount)
	require.Equal(t, "series", token.Series[0].Metadata[0].Key)
}

func TestRPCMethodsReturnTransportErrorWithoutPanic(t *testing.T) {
	transportErr := errors.New("transport failed")

	tests := []struct {
		name string
		call func(PhantasmaRPC) error
	}{
		{
			name: "GetPlatforms",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetPlatforms()
				return err
			},
		},
		{
			name: "GetAccounts",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccounts("P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP")
				return err
			},
		},
		{
			name: "LookupName",
			call: func(client PhantasmaRPC) error {
				_, err := client.LookupName("anonymous")
				return err
			},
		},
		{
			name: "GetAccount",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccount("P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP")
				return err
			},
		},
		{
			name: "GetAddressTransactions",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAddressTransactions("P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP", 1, 10)
				return err
			},
		},
		{
			name: "GetAddressTransactionCount",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAddressTransactionCount("P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP", "main")
				return err
			},
		},
		{
			name: "GetBlockByHeight",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockByHeight("main", "1")
				return err
			},
		},
		{
			name: "GetBlockHeight",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockHeight("main")
				return err
			},
		},
		{
			name: "GetContract",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetContract("gas", "main")
				return err
			},
		},
		{
			name: "GetBlockByHash",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockByHash("block-hash")
				return err
			},
		},
		{
			name: "SendCarbonTransaction",
			call: func(client PhantasmaRPC) error {
				_, err := client.SendCarbonTransaction("00")
				return err
			},
		},
		{
			name: "GetTokenSeries",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenSeries("NFT", 42, 10, "")
				return err
			},
		},
		{
			name: "GetAuctionsCount",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAuctionsCount("main", "NFT")
				return err
			},
		},
		{
			name: "WriteArchive",
			call: func(client PhantasmaRPC) error {
				_, err := client.WriteArchive("archive-hash", 0, nil)
				return err
			},
		},
		{
			name: "InvokeRawScript",
			call: func(client PhantasmaRPC) error {
				_, err := client.InvokeRawScript("main", "00")
				return err
			},
		},
		{
			name: "SendRawTransaction",
			call: func(client PhantasmaRPC) error {
				_, err := client.SendRawTransaction("00")
				return err
			},
		},
		{
			name: "GetTransaction",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTransaction("00")
				return err
			},
		},
		{
			name: "GetTokens",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokens(false)
				return err
			},
		},
		{
			name: "GetTokensAsMap",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokensAsMap(false)
				return err
			},
		},
		{
			name: "GetToken",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetToken("SOUL", false)
				return err
			},
		},
		{
			name: "GetVersion",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetVersion()
				return err
			},
		},
		{
			name: "GetPhantasmaVMConfig",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetPhantasmaVMConfig("main")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := PhantasmaRPC{client: stubRPCClient{err: transportErr}}

			var err error
			require.NotPanics(t, func() {
				err = tt.call(client)
			})
			require.ErrorIs(t, err, transportErr)
		})
	}
}
