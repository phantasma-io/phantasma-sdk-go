package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/rpc/response"
)

func TestGetChainToken(t *testing.T) {
	chainTokens = []response.TokenResult{{Symbol: "SOUL"}}

	token, ok := getChainToken("SOUL")
	if !ok {
		t.Fatal("expected SOUL token to be found")
	}
	if token.Symbol != "SOUL" {
		t.Fatalf("expected SOUL token, got %+v", token)
	}
	_, ok = getChainToken("KCAL")
	if ok {
		t.Fatal("missing token must not be found")
	}
}

func TestReadConsoleLineAcceptsFinalEOFToken(t *testing.T) {
	line, ok := readConsoleLine(bufio.NewReader(strings.NewReader("5")))
	if !ok {
		t.Fatalf("expected final token before EOF to be accepted")
	}
	if line != "5" {
		t.Fatalf("expected final token 5, got %q", line)
	}
}

func TestReadConsoleLineReportsEmptyEOF(t *testing.T) {
	line, ok := readConsoleLine(bufio.NewReader(strings.NewReader("")))
	if ok {
		t.Fatalf("expected empty EOF to close input, got %q", line)
	}
}
