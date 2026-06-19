package rpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func echoIDHandler(capture func(*http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		capture(r)
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)
		out, _ := json.Marshal(map[string]interface{}{"jsonrpc": "2.0", "id": req["id"], "result": "8804017"})
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write(out)
	}
}

func TestNewRPCWithApiKeySendsHeader(t *testing.T) {
	var got string
	server := httptest.NewServer(echoIDHandler(func(r *http.Request) { got = r.Header.Get("X-Api-Key") }))
	defer server.Close()

	client := NewRPCWithApiKey(server.URL, "test-key")
	_, err := client.Call(context.Background(), "getBlockHeight", "main")

	require.NoError(t, err)
	require.Equal(t, "test-key", got)
}

func TestNewRPCOmitsApiKeyHeader(t *testing.T) {
	var present bool
	server := httptest.NewServer(echoIDHandler(func(r *http.Request) {
		_, present = r.Header["X-Api-Key"]
	}))
	defer server.Close()

	client := NewRPC(server.URL)
	_, err := client.Call(context.Background(), "getBlockHeight", "main")

	require.NoError(t, err)
	require.False(t, present)
}
