package rpc

import (
	"context"
	"errors"
	"testing"

	"github.com/phantasma-io/phantasma-go/pkg/jsonrpc"
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

func TestGetAccountRejectsNilRPCResponseWithoutPanic(t *testing.T) {
	client := PhantasmaRPC{client: stubRPCClient{}}

	var err error
	require.NotPanics(t, func() {
		_, err = client.GetAccount("P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP")
	})
	require.Error(t, err)
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
			name: "GetAccountEx",
			call: func(client PhantasmaRPC) error {
				_, err := client.GetAccountEx("P2KA7yzB3uUncuAqP6tLut27iTKAC6ZTnAVM4myUuG57oQP")
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
