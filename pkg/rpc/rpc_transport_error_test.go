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
	ctx      context.Context
	method   string
	params   []interface{}
	response *jsonrpc.RPCResponse
}

func (c *recordingRPCClient) Call(ctx context.Context, method string, params ...interface{}) (*jsonrpc.RPCResponse, error) {
	c.ctx = ctx
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
		_, err = client.GetAccount(context.Background(), "P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP")
	})
	require.Error(t, err)
}

func TestLookupNameCallsLookupNameAndReturnsAddress(t *testing.T) {
	recorder := &recordingRPCClient{
		response: &jsonrpc.RPCResponse{Result: "P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP"},
	}
	client := PhantasmaRPC{client: recorder}

	address, err := client.LookupName(context.Background(), "anonymous")

	require.NoError(t, err)
	require.Equal(t, "P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP", address)
	require.Equal(t, "lookUpName", recorder.method)
	require.Equal(t, []interface{}{"anonymous"}, recorder.params)
}

func TestTypedRPCWrapperUsesCallerContext(t *testing.T) {
	type contextKey struct{}

	recorder := &recordingRPCClient{
		response: &jsonrpc.RPCResponse{Result: []interface{}{}},
	}
	client := PhantasmaRPC{client: recorder}
	ctx := context.WithValue(context.Background(), contextKey{}, "caller-context")

	_, err := client.GetChains(ctx, true)

	require.NoError(t, err)
	require.Same(t, ctx, recorder.ctx)
}

func TestGetBlockHeightRejectsInvalidHeightWithoutNilResult(t *testing.T) {
	// RPC height parsing must return an explicit error instead of a nil *big.Int with nil error.
	client := PhantasmaRPC{client: stubRPCClient{response: &jsonrpc.RPCResponse{Result: "not-a-number"}}}

	height, err := client.GetBlockHeight(context.Background(), "main")

	require.Error(t, err)
	require.NotNil(t, height)
	require.Zero(t, height.Sign())
}

func TestSignAndSendTransactionRejectsNilKeyWithoutPanic(t *testing.T) {
	client := PhantasmaRPC{client: stubRPCClient{}}

	var err error
	require.NotPanics(t, func() {
		_, err = client.SignAndSendTransaction(context.Background(), nil, "testnet", []byte{0x01}, "main", nil)
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
				_, err := client.GetAccountsWithAddressType(context.Background(), []string{"Pone", "Ptwo"}, true, true, AddressTypeCarbon)
				return err
			},
			method: "getAccounts",
			params: []interface{}{"Pone,Ptwo", true, true, AddressTypeCarbon},
		},
		{
			name:     "GetAccountInfo",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountInfo(context.Background(), "Paccount")
				return err
			},
			method: "getAccountInfo",
			params: []interface{}{"Paccount"},
		},
		{
			name:     "GetAccountInfoWithAddressType",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountInfoWithAddressType(context.Background(), "001122", false, AddressTypeCarbon)
				return err
			},
			method: "getAccountInfo",
			params: []interface{}{"001122", false, AddressTypeCarbon},
		},
		{
			name:     "GetAccountWithAddressType",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountWithAddressType(context.Background(), "Paccount", true, false, AddressTypeCarbon)
				return err
			},
			method: "getAccount",
			params: []interface{}{"Paccount", true, false, AddressTypeCarbon},
		},
		{
			name:     "GetBlockTransactionCountByHash",
			response: "7",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockTransactionCountByHash(context.Background(), "block-hash")
				return err
			},
			method: "getBlockTransactionCountByHash",
			params: []interface{}{"main", "block-hash"},
		},
		{
			name:     "GetBlockTransactionCountByHashOnChain",
			response: "7",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockTransactionCountByHashOnChain(context.Background(), "custom", "block-hash")
				return err
			},
			method: "getBlockTransactionCountByHash",
			params: []interface{}{"custom", "block-hash"},
		},
		{
			name:     "GetBlockByHash",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockByHash(context.Background(), "block-hash")
				return err
			},
			method: "getBlockByHash",
			params: []interface{}{"block-hash"},
		},
		{
			name:     "GetLatestBlock",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetLatestBlock(context.Background(), "main")
				return err
			},
			method: "getLatestBlock",
			params: []interface{}{"main"},
		},
		{
			name:     "GetTransactionByBlockHashAndIndex",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTransactionByBlockHashAndIndex(context.Background(), "block-hash", 2)
				return err
			},
			method: "getTransactionByBlockHashAndIndex",
			params: []interface{}{"main", "block-hash", 2},
		},
		{
			name:     "GetTransactionByBlockHashAndIndexOnChain",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTransactionByBlockHashAndIndexOnChain(context.Background(), "custom", "block-hash", 3)
				return err
			},
			method: "getTransactionByBlockHashAndIndex",
			params: []interface{}{"custom", "block-hash", 3},
		},
		{
			name:     "GetChains",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetChains(context.Background(), true)
				return err
			},
			method: "getChains",
			params: []interface{}{true},
		},
		{
			name:     "GetChain",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetChain(context.Background(), "main", true)
				return err
			},
			method: "getChain",
			params: []interface{}{"main", true},
		},
		{
			name:     "GetNexus",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetNexus(context.Background(), true)
				return err
			},
			method: "getNexus",
			params: []interface{}{true},
		},
		{
			name:     "GetContracts",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetContracts(context.Background(), "main", true)
				return err
			},
			method: "getContracts",
			params: []interface{}{"main", true},
		},
		{
			name:     "GetContractByName",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetContractByName(context.Background(), "main", "gas")
				return err
			},
			method: "getContract",
			params: []interface{}{"main", "gas"},
		},
		{
			name:     "GetContractByAddress",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetContractByAddress(context.Background(), "main", "Pcontract")
				return err
			},
			method: "getContractByAddress",
			params: []interface{}{"main", "Pcontract"},
		},
		{
			name:     "GetOrganization",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetOrganization(context.Background(), "masters", true)
				return err
			},
			method: "getOrganization",
			params: []interface{}{"masters", true},
		},
		{
			name:     "GetOrganizations",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetOrganizations(context.Background(), 2, "cursor", true)
				return err
			},
			method: "getOrganizations",
			params: []interface{}{2, "cursor", true},
		},
		{
			name:     "GetOrganizationMembers",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetOrganizationMembers(context.Background(), "masters", 2, "cursor", false)
				return err
			},
			method: "getOrganizationMembers",
			params: []interface{}{"masters", 2, "cursor", false},
		},
		{
			name:     "GetOrganizationMember",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetOrganizationMember(context.Background(), "masters", "Pmember", true, AddressTypePhantasma)
				return err
			},
			method: "getOrganizationMember",
			params: []interface{}{"masters", "Pmember", true, AddressTypePhantasma},
		},
		{
			name:     "GetLeaderboard",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetLeaderboard(context.Background(), "top")
				return err
			},
			method: "getLeaderboard",
			params: []interface{}{"top"},
		},
		{
			name:     "SendCarbonTransaction",
			response: "tx-hash",
			call: func(client PhantasmaRPC) error {
				_, err := client.SendCarbonTransaction(context.Background(), "00")
				return err
			},
			method: "sendCarbonTransaction",
			params: []interface{}{"00"},
		},
		{
			name:     "GetTokensByOwner",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokensByOwner(context.Background(), true, "Powner")
				return err
			},
			method: "getTokens",
			params: []interface{}{true, "Powner"},
		},
		{
			name:     "GetTokensByOwnerWithAddressType",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokensByOwnerWithAddressType(context.Background(), true, "Powner", AddressTypeCarbon)
				return err
			},
			method: "getTokens",
			params: []interface{}{true, "Powner", AddressTypeCarbon},
		},
		{
			name:     "GetTokenWithID",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenWithID(context.Background(), "NFT", true, 42)
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
				_, err := client.GetTokenWithID(context.Background(), "NFT", true, 0)
				return err
			},
			method: "getToken",
			params: []interface{}{"NFT", true, uint64(0)},
		},
		{
			name:     "GetTokenData",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenData(context.Background(), "NFT", "100")
				return err
			},
			method: "getTokenData",
			params: []interface{}{"NFT", "100"},
		},
		{
			name:     "GetTokenBalance",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenBalance(context.Background(), "Paccount", "SOUL", "main")
				return err
			},
			method: "getTokenBalance",
			params: []interface{}{"Paccount", "SOUL", "main"},
		},
		{
			name:     "GetTokenBalanceChecked",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenBalanceChecked(context.Background(), "Paccount", "SOUL", "main", true)
				return err
			},
			method: "getTokenBalance",
			params: []interface{}{"Paccount", "SOUL", "main", true},
		},
		{
			name:     "GetTokenBalanceWithAddressType",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenBalanceWithAddressType(context.Background(), "Paccount", "SOUL", "main", true, AddressTypeCarbon)
				return err
			},
			method: "getTokenBalance",
			params: []interface{}{"Paccount", "SOUL", "main", true, AddressTypeCarbon},
		},
		{
			name:     "GetTokenSeries",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenSeries(context.Background(), "NFT", 42, 10, "cursor")
				return err
			},
			method: "getTokenSeries",
			params: []interface{}{"NFT", uint64(42), 10, "cursor"},
		},
		{
			name:     "GetTokenSeriesDefaultCarbonID",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenSeries(context.Background(), "NFT", 0, 10, "cursor")
				return err
			},
			method: "getTokenSeries",
			params: []interface{}{"NFT", uint64(0), 10, "cursor"},
		},
		{
			name:     "GetTokenSeriesByID",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenSeriesByID(context.Background(), "NFT", 42, "9", 3)
				return err
			},
			method: "getTokenSeriesById",
			params: []interface{}{"NFT", uint64(42), "9", uint32(3)},
		},
		{
			name:     "GetTokenNFTs",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenNFTs(context.Background(), 42, 3, 10, "cursor", true)
				return err
			},
			method: "getTokenNFTs",
			params: []interface{}{uint64(42), uint32(3), 10, "cursor", true, ""},
		},
		{
			name:     "GetTokenNFTsWithSeriesID",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenNFTsWithSeriesID(context.Background(), 42, 0, "9", 10, "cursor", true)
				return err
			},
			method: "getTokenNFTs",
			params: []interface{}{uint64(42), uint32(0), 10, "cursor", true, "9"},
		},
		{
			name:     "GetAccountFungibleTokens",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountFungibleTokens(context.Background(), "Paccount", "SOUL", 1, 10, "cursor", true)
				return err
			},
			method: "getAccountFungibleTokens",
			params: []interface{}{"Paccount", "SOUL", uint64(1), 10, "cursor", true},
		},
		{
			name:     "GetAccountFungibleTokensWithAddressType",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountFungibleTokensWithAddressType(context.Background(), "Paccount", "SOUL", 1, 10, "cursor", true, AddressTypeCarbon)
				return err
			},
			method: "getAccountFungibleTokens",
			params: []interface{}{"Paccount", "SOUL", uint64(1), 10, "cursor", true, AddressTypeCarbon},
		},
		{
			name:     "GetAccountNFTs",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountNFTs(context.Background(), "Paccount", "NFT", 42, 3, 10, "cursor", true, true)
				return err
			},
			method: "getAccountNFTs",
			params: []interface{}{"Paccount", "NFT", uint64(42), uint32(3), 10, "cursor", true, true},
		},
		{
			name:     "GetAccountNFTsWithAddressType",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountNFTsWithAddressType(context.Background(), "Paccount", "NFT", 42, 3, 10, "cursor", true, true, AddressTypeCarbon)
				return err
			},
			method: "getAccountNFTs",
			params: []interface{}{"Paccount", "NFT", uint64(42), uint32(3), 10, "cursor", true, true, AddressTypeCarbon},
		},
		{
			name:     "GetAccountOwnedTokens",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountOwnedTokens(context.Background(), "Paccount", "NFT", 42, 10, "cursor", true)
				return err
			},
			method: "getAccountOwnedTokens",
			params: []interface{}{"Paccount", "NFT", uint64(42), 10, "cursor", true},
		},
		{
			name:     "GetAccountOwnedTokensWithAddressType",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountOwnedTokensWithAddressType(context.Background(), "Paccount", "NFT", 42, 10, "cursor", true, AddressTypeCarbon)
				return err
			},
			method: "getAccountOwnedTokens",
			params: []interface{}{"Paccount", "NFT", uint64(42), 10, "cursor", true, AddressTypeCarbon},
		},
		{
			name:     "GetAccountOwnedTokenSeries",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountOwnedTokenSeries(context.Background(), "Paccount", "NFT", 42, 10, "cursor", true)
				return err
			},
			method: "getAccountOwnedTokenSeries",
			params: []interface{}{"Paccount", "NFT", uint64(42), 10, "cursor", true},
		},
		{
			name:     "GetAccountOwnedTokenSeriesWithAddressType",
			response: cursorResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountOwnedTokenSeriesWithAddressType(context.Background(), "Paccount", "NFT", 42, 10, "cursor", true, AddressTypeCarbon)
				return err
			},
			method: "getAccountOwnedTokenSeries",
			params: []interface{}{"Paccount", "NFT", uint64(42), 10, "cursor", true, AddressTypeCarbon},
		},
		{
			name:     "GetAuctionsCount",
			response: "7",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAuctionsCount(context.Background(), "main", "NFT")
				return err
			},
			method: "getAuctionsCount",
			params: []interface{}{"main", "NFT"},
		},
		{
			name:     "GetAuctions",
			response: paginatedArrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAuctions(context.Background(), "main", "NFT", 1, 10)
				return err
			},
			method: "getAuctions",
			params: []interface{}{"main", "NFT", 1, 10},
		},
		{
			name:     "GetAuction",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAuction(context.Background(), "main", "NFT", "100")
				return err
			},
			method: "getAuction",
			params: []interface{}{"main", "NFT", "100"},
		},
		{
			name:     "GetNFT",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetNFT(context.Background(), "NFT", "100", true)
				return err
			},
			method: "getNFT",
			params: []interface{}{"NFT", "100", true},
		},
		{
			name:     "GetNFTs",
			response: arrayResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetNFTs(context.Background(), "NFT", []string{"100", "101"}, true)
				return err
			},
			method: "getNFTs",
			params: []interface{}{"NFT", "100,101", true},
		},
		{
			name:     "GetArchive",
			response: objectResult,
			call: func(client PhantasmaRPC) error {
				_, err := client.GetArchive(context.Background(), "archive-hash")
				return err
			},
			method: "getArchive",
			params: []interface{}{"archive-hash"},
		},
		{
			name:     "WriteArchive",
			response: "true",
			call: func(client PhantasmaRPC) error {
				_, err := client.WriteArchive(context.Background(), "archive-hash", 4, []byte("block"))
				return err
			},
			method: "writeArchive",
			params: []interface{}{"archive-hash", 4, "YmxvY2s="},
		},
		{
			name:     "ReadArchive",
			response: "YmxvY2s=",
			call: func(client PhantasmaRPC) error {
				_, err := client.ReadArchive(context.Background(), "archive-hash", 4)
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

func TestAddressTypeMarshalJSONUsesRPCEnumNames(t *testing.T) {
	payload, err := AddressTypeCarbon.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, []byte(`"Carbon"`), payload)
}

// getAccountInfo names the staking object "stake", while getAccount carries the same object under
// "stakes" and uses "stake" for a deprecated flat scalar. Binding the wrong one would silently yield
// a zero stake, so the mapping is pinned against the exact wire shape the node returns.
func TestAccountInfoResponseFieldsDecode(t *testing.T) {
	result := &jsonrpc.RPCResponse{Result: map[string]interface{}{
		"address": "P2Kaccount",
		"name":    "myname",
		"stake": map[string]interface{}{
			"amount":    "1500000000000",
			"time":      1743520000,
			"unclaimed": "42000000000",
		},
	}}

	client := PhantasmaRPC{client: &recordingRPCClient{response: result}}

	account, err := client.GetAccountInfo(context.Background(), "P2Kaccount")
	require.NoError(t, err)
	require.Equal(t, "P2Kaccount", account.Address)
	require.Equal(t, "myname", account.Name)
	require.Equal(t, "1500000000000", account.Stake.Amount)
	require.Equal(t, uint(1743520000), account.Stake.Time)
	require.Equal(t, "42000000000", account.Stake.Unclaimed)
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

func TestOrganizationResponseFieldsDecode(t *testing.T) {
	memberCount := "2"
	orgOwner := "Powner"
	carbonOwner := "0xowner"
	orgResult := &jsonrpc.RPCResponse{Result: map[string]interface{}{
		"name":        "masters",
		"owner":       orgOwner,
		"carbonOwner": carbonOwner,
		"memberCount": memberCount,
		"metadata": []interface{}{map[string]interface{}{
			"key":   "role",
			"value": "validators",
		}},
	}}
	var organization resp.OrganizationResult
	require.NoError(t, orgResult.GetObject(&organization))
	require.Equal(t, orgOwner, *organization.Owner)
	require.Equal(t, carbonOwner, *organization.CarbonOwner)
	require.Equal(t, memberCount, *organization.MemberCount)
	require.Equal(t, "role", organization.Metadata[0].Key)

	memberTime := uint64(123)
	memberResult := &jsonrpc.RPCResponse{Result: map[string]interface{}{
		"address":       "Pmember",
		"carbonAddress": "0xmember",
		"isMember":      true,
		"memberTime":    memberTime,
	}}
	var member resp.OrganizationMemberResult
	require.NoError(t, memberResult.GetObject(&member))
	require.True(t, member.IsMember)
	require.Equal(t, memberTime, *member.MemberTime)
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
				_, err := client.GetPlatforms(context.Background())
				return err
			},
		},
		{
			name: "GetAccounts",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccounts(context.Background(), "P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP")
				return err
			},
		},
		{
			name: "LookupName",
			call: func(client PhantasmaRPC) error {
				_, err := client.LookupName(context.Background(), "anonymous")
				return err
			},
		},
		{
			name: "GetAccount",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccount(context.Background(), "P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP")
				return err
			},
		},
		{
			name: "GetAccountInfo",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountInfo(context.Background(), "P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP")
				return err
			},
		},
		{
			name: "GetAddressTransactions",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAddressTransactions(context.Background(), "P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP", 1, 10)
				return err
			},
		},
		{
			name: "GetAddressTransactionCount",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAddressTransactionCount(context.Background(), "P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP", "main")
				return err
			},
		},
		{
			name: "GetBlockByHeight",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockByHeight(context.Background(), "main", "1")
				return err
			},
		},
		{
			name: "GetBlockHeight",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockHeight(context.Background(), "main")
				return err
			},
		},
		{
			name: "GetContract",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetContract(context.Background(), "gas", "main")
				return err
			},
		},
		{
			name: "GetBlockByHash",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetBlockByHash(context.Background(), "block-hash")
				return err
			},
		},
		{
			name: "SendCarbonTransaction",
			call: func(client PhantasmaRPC) error {
				_, err := client.SendCarbonTransaction(context.Background(), "00")
				return err
			},
		},
		{
			name: "GetTokenSeries",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokenSeries(context.Background(), "NFT", 42, 10, "")
				return err
			},
		},
		{
			name: "GetAuctionsCount",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAuctionsCount(context.Background(), "main", "NFT")
				return err
			},
		},
		{
			name: "WriteArchive",
			call: func(client PhantasmaRPC) error {
				_, err := client.WriteArchive(context.Background(), "archive-hash", 0, nil)
				return err
			},
		},
		{
			name: "InvokeRawScript",
			call: func(client PhantasmaRPC) error {
				_, err := client.InvokeRawScript(context.Background(), "main", "00")
				return err
			},
		},
		{
			name: "SendRawTransaction",
			call: func(client PhantasmaRPC) error {
				_, err := client.SendRawTransaction(context.Background(), "00")
				return err
			},
		},
		{
			name: "GetTransaction",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTransaction(context.Background(), "00")
				return err
			},
		},
		{
			name: "GetTokens",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokens(context.Background(), false)
				return err
			},
		},
		{
			name: "GetTokensAsMap",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetTokensAsMap(context.Background(), false)
				return err
			},
		},
		{
			name: "GetToken",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetToken(context.Background(), "SOUL", false)
				return err
			},
		},
		{
			name: "GetVersion",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetVersion(context.Background())
				return err
			},
		},
		{
			name: "GetPhantasmaVMConfig",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetPhantasmaVMConfig(context.Background(), "main")
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
