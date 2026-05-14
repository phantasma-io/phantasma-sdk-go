package event

import (
	"math/big"
	"testing"

	sdkio "github.com/phantasma-io/phantasma-sdk-go/pkg/io"
)

func TestEventKindHelpers(t *testing.T) {
	var kind EventKind
	kind.SetString("TokenReceive")
	if kind != TokenReceive || kind.String() != "TokenReceive" || !kind.IsTokenEvent() {
		t.Fatalf("token event kind helpers failed")
	}
	if !OrderBid.IsMarketEvent() || TokenMint.IsMarketEvent() {
		t.Fatalf("market event helpers failed")
	}
}

func TestTokenEventDataRoundTrip(t *testing.T) {
	event := &TokenEventData{Symbol: "SOUL", Value: big.NewInt(123), ChainName: "main"}
	decoded := sdkio.Deserialize[*TokenEventData](sdkio.Serialize[*TokenEventData](event))
	if decoded.Symbol != event.Symbol || decoded.Value.Cmp(event.Value) != 0 || decoded.ChainName != event.ChainName {
		t.Fatalf("token event round trip mismatch: %+v", decoded)
	}
}
