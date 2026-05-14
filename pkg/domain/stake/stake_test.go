package stake

import (
	"math/big"
	"testing"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/domain/types"
	sdkio "github.com/phantasma-io/phantasma-sdk-go/pkg/io"
)

func TestEnergyStakeRoundTrip(t *testing.T) {
	stake := &EnergyStake{StakeAmount: big.NewInt(100), StakeTime: types.NewTimestamp(42)}
	decoded := sdkio.Deserialize[*EnergyStake](sdkio.Serialize[*EnergyStake](stake))
	if decoded.StakeAmount.Cmp(stake.StakeAmount) != 0 || decoded.StakeTime.Value != stake.StakeTime.Value {
		t.Fatalf("energy stake round trip mismatch: %+v", decoded)
	}
}

func TestEnergyClaimRoundTrip(t *testing.T) {
	claim := &EnergyClaim_S{StakeAmount: "100", ClaimDate: types.NewTimestamp(42), IsNew: true}
	decoded := sdkio.Deserialize[*EnergyClaim_S](sdkio.Serialize[*EnergyClaim_S](claim))
	if decoded.StakeAmount != claim.StakeAmount || decoded.ClaimDate.Value != claim.ClaimDate.Value || !decoded.IsNew {
		t.Fatalf("energy claim round trip mismatch: %+v", decoded)
	}
}
