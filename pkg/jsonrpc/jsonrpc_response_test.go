package jsonrpc

import (
	"encoding/json"
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

func TestRPCResponseIDAcceptsNumberStringAndNull(t *testing.T) {
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
		{
			name: "null",
			body: `{"jsonrpc":"2.0","id":null,"result":true}`,
			id:   0,
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
