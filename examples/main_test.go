package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/rpc/response"
)

func TestGetChainToken(t *testing.T) {
	chainTokens = []response.TokenResult{{Symbol: "SOUL"}}

	token := getChainToken("SOUL")
	if token.Symbol != "SOUL" {
		t.Fatalf("expected SOUL token, got %+v", token)
	}
	defer func() {
		if recover() == nil {
			t.Fatalf("missing token must panic")
		}
	}()
	_ = getChainToken("KCAL")
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
