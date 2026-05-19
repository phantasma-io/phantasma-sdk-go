package jsonrpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func newStrictResponseDecoder(body string) *json.Decoder {
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	return decoder
}

func TestRPCResponseIDAcceptsNumberAndString(t *testing.T) {
	tests := []struct {
		name string
		body string
		id   int
	}{
		{
			name: "number",
			body: `{"jsonrpc":"2.0","id":7,"result":true}`,
			id:   7,
		},
		{
			name: "string",
			body: `{"jsonrpc":"2.0","id":"7","result":true}`,
			id:   7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := decodeRPCResponse(newStrictResponseDecoder(tt.body))
			require.NoError(t, err)
			require.NotNil(t, response)
			require.Equal(t, tt.id, response.ID)
		})
	}
}

func TestRPCResponseIDRejectsMissingAndNull(t *testing.T) {
	// Missing and null ids cannot identify the request that produced the response.
	for _, body := range []string{
		`{"jsonrpc":"2.0","result":true}`,
		`{"jsonrpc":"2.0","id":null,"result":true}`,
	} {
		_, err := decodeRPCResponse(newStrictResponseDecoder(body))
		require.Error(t, err)
	}
}

func TestNewRequestUsesDefaultNumericID(t *testing.T) {
	// CallRaw uses NewRequest directly, so its default id must be parseable by response validation.
	require.Equal(t, "0", NewRequest("getBlockHeight", "main").ID)
}

func TestRPCResponsesHelpersUseFlexibleID(t *testing.T) {
	responses, err := decodeRPCResponses(newStrictResponseDecoder(`[
		{"jsonrpc":"2.0","id":"1","result":true},
		{"jsonrpc":"2.0","id":2,"result":false}
	]`))
	require.NoError(t, err)

	require.Equal(t, responses[0], responses.GetByID(1))
	require.Equal(t, responses[1], responses.GetByID(2))
	require.Equal(t, responses[0], responses.AsMap()[1])
	require.Equal(t, responses[1], responses.AsMap()[2])
}

func TestDecodeRPCResponsePreservesStrictDecoding(t *testing.T) {
	_, err := decodeRPCResponse(newStrictResponseDecoder(`{"jsonrpc":"2.0","id":"1","result":true,"extra":false}`))
	require.Error(t, err)
}

func TestDecodeRPCResponsePreservesJSONNumberResults(t *testing.T) {
	response, err := decodeRPCResponse(newStrictResponseDecoder(`{"jsonrpc":"2.0","id":"1","result":7}`))
	require.NoError(t, err)
	require.IsType(t, json.Number(""), response.Result)
}

func TestClientAcceptsMatchingSingleResponseID(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "number",
			body: `{"jsonrpc":"2.0","id":0,"result":true}`,
		},
		{
			name: "string",
			body: `{"jsonrpc":"2.0","id":"0","result":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("content-type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			response, err := NewClient(server.URL).Call(context.Background(), "getBlockHeight", "main")
			require.NoError(t, err)
			require.Equal(t, 0, response.ID)
		})
	}
}

func TestClientGeneratesDistinctSingleRequestIDsAndRejectsStaleID(t *testing.T) {
	// Normal Call() usage must advance the request id and reject a stale echo
	// from the previous call before the caller can consume the result body.
	var requestIDs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request RPCRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		requestIDs = append(requestIDs, request.ID)

		responseID := request.ID
		if len(requestIDs) == 2 {
			responseID = requestIDs[0]
		}

		w.Header().Set("content-type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      responseID,
			"result":  true,
		}))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	first, err := client.Call(context.Background(), "first")
	require.NoError(t, err)
	require.Equal(t, 0, first.ID)

	_, err = client.Call(context.Background(), "second")
	require.ErrorContains(t, err, "response id mismatch: got 0, expected 1")
	require.Equal(t, []string{"0", "1"}, requestIDs)
}

func TestClientRejectsMismatchedSingleResponseID(t *testing.T) {
	// A single-call response must match the exact numeric request id sent by the client.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1","result":true}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL).Call(context.Background(), "getBlockHeight", "main")
	require.ErrorContains(t, err, "response id mismatch")
}

func TestClientRejectsMissingAndNullSingleResponseID(t *testing.T) {
	for _, body := range []string{
		`{"jsonrpc":"2.0","result":true}`,
		`{"jsonrpc":"2.0","id":null,"result":true}`,
	} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("content-type", "application/json")
				_, _ = w.Write([]byte(body))
			}))
			defer server.Close()

			_, err := NewClient(server.URL).Call(context.Background(), "getBlockHeight", "main")
			require.Error(t, err)
			require.ErrorContains(t, err, "rpc response id")
		})
	}
}

func TestClientRejectsMismatchedIDBeforeRPCErrorAndHTTPStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1","error":{"code":-32603,"message":"Execution failed"}}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL).Call(context.Background(), "getBlockHeight", "main")
	require.ErrorContains(t, err, "response id mismatch")
	require.NotErrorIs(t, err, &HTTPError{})
	require.NotContains(t, err.Error(), "Execution failed")
}

func TestClientRejectsStaleDefaultResponseIDWhenRequestUsesDifferentID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"0","result":true}`))
	}))
	defer server.Close()

	client := NewClientWithOpts(server.URL, &RPCClientOpts{DefaultRequestID: 7})
	_, err := client.Call(context.Background(), "getBlockHeight", "main")
	require.ErrorContains(t, err, "response id mismatch: got 0, expected 7")
}

func TestClientCallBatchAssignsDistinctRequestIDs(t *testing.T) {
	// CallBatch owns automatic ids for the whole batch, so callers get one
	// response per generated request id even when responses arrive unordered.
	var requestIDs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requests []RPCRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&requests))
		require.Len(t, requests, 2)
		requestIDs = []string{requests[0].ID, requests[1].ID}

		w.Header().Set("content-type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode([]map[string]interface{}{
			{"jsonrpc": "2.0", "id": requestIDs[1], "result": false},
			{"jsonrpc": "2.0", "id": requestIDs[0], "result": true},
		}))
	}))
	defer server.Close()

	responses, err := NewClient(server.URL).CallBatch(context.Background(), RPCRequests{
		NewRequest("first"),
		NewRequest("second"),
	})
	require.NoError(t, err)
	require.Equal(t, []string{"0", "1"}, requestIDs)
	require.Equal(t, true, responses.GetByID(0).Result)
	require.Equal(t, false, responses.GetByID(1).Result)
}

func TestClientRejectsUnexpectedBatchResponseID(t *testing.T) {
	// Batch responses may arrive unordered, but every response id must belong to the submitted batch.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`[
			{"jsonrpc":"2.0","id":"0","result":true},
			{"jsonrpc":"2.0","id":"7","result":false}
		]`))
	}))
	defer server.Close()

	requests := RPCRequests{
		NewRequestWithID(0, "first"),
		NewRequestWithID(1, "second"),
	}
	_, err := NewClient(server.URL).CallBatchRaw(context.Background(), requests)
	require.ErrorContains(t, err, "unexpected id 7")
}

func TestClientAcceptsUnorderedBatchResponseIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`[
			{"jsonrpc":"2.0","id":"1","result":false},
			{"jsonrpc":"2.0","id":"0","result":true}
		]`))
	}))
	defer server.Close()

	requests := RPCRequests{
		NewRequestWithID(0, "first"),
		NewRequestWithID(1, "second"),
	}
	responses, err := NewClient(server.URL).CallBatchRaw(context.Background(), requests)
	require.NoError(t, err)
	require.Equal(t, true, responses.GetByID(0).Result)
	require.Equal(t, false, responses.GetByID(1).Result)
}

func TestClientRejectsDuplicateAndMissingBatchResponseIDs(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		errorText string
	}{
		{
			name: "duplicate",
			body: `[
				{"jsonrpc":"2.0","id":"0","result":true},
				{"jsonrpc":"2.0","id":"0","result":false}
			]`,
			errorText: "duplicated",
		},
		{
			name: "missing one response",
			body: `[
				{"jsonrpc":"2.0","id":"0","result":true}
			]`,
			errorText: "response count mismatch",
		},
		{
			name: "missing response id",
			body: `[
				{"jsonrpc":"2.0","id":"0","result":true},
				{"jsonrpc":"2.0","result":false}
			]`,
			errorText: "rpc response id missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("content-type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			requests := RPCRequests{
				NewRequestWithID(0, "first"),
				NewRequestWithID(1, "second"),
			}
			_, err := NewClient(server.URL).CallBatchRaw(context.Background(), requests)
			require.ErrorContains(t, err, tt.errorText)
		})
	}
}
