package scriptbuilder_test

import (
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-go/pkg/vm"
	scriptbuilder "github.com/phantasma-io/phantasma-go/pkg/vm/script_builder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expectedConsensusSingleVoteScriptHex = "0D00030350340303000D000302102703000D000223220000000000000000000000000000000000000000000000000000000000000000000003000D000223220100AA53BE71FC41BC0889B694F4D6D03F7906A3D9A21705943CAF9632EEAFBB489503000D000408416C6C6F7747617303000D0004036761732D00012E010D0003010003000D00041D73797374656D2E6E657875732E70726F746F636F6C2E76657273696F6E03000D00042F50324B464579466576705166536157384734566A536D6857555A585234517247395951523148624D7054554370434C03000D00040A53696E676C65566F746503000D000409636F6E73656E7375732D00012E010D000223220100AA53BE71FC41BC0889B694F4D6D03F7906A3D9A21705943CAF9632EEAFBB489503000D0004085370656E6447617303000D0004036761732D00012E010B"

func TestNewScript(t *testing.T) {
	fromAddress := cryptography.MustAddressFromString("P2KM9FjYrDXnPPAynLXAHdQ8wYz8de9VbDeybrLepnw6C5x")
	toAddress := cryptography.MustAddressFromString("P2KM9FjYrDXnPPAynLXAHdQ8wYz8de9VbDeybrLepnw6C5x")
	symbols := []string{"SOUL", "KCAL"}

	// Basic extension calls should compose without mutating shared builder state or panicking.
	assert.NotPanics(t, func() {
		sb := scriptbuilder.BeginScript()
		sb.AllowGas(fromAddress, cryptography.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
			TransferTokens(symbols[0], fromAddress, toAddress, big.NewInt(100000000)).
			SpendGas(fromAddress).
			EndScript()
	})

	assert.NotPanics(t, func() {
		sb := scriptbuilder.BeginScript()
		sb.CallInterop("Runtime.TransferToken", fromAddress, toAddress, symbols[0], "TOKEN_ID")
	})

	assert.NotPanics(t, func() {
		sb := scriptbuilder.BeginScript()
		sb.CallInterop("Runtime.TransferToken", fromAddress, toAddress, symbols, "TOKEN_ID").EndScript()
	})
}

func TestScriptBuilderTextAddressHelpersMatchTypedHelpers(t *testing.T) {
	fromText := "P2KM9FjYrDXnPPAynLXAHdQ8wYz8de9VbDeybrLepnw6C5x"
	toText := "P2KCineFiatZR8fDyU4pSfb3Bq1vtyW3zqBJi268YV5fH9e"
	from := cryptography.MustAddressFromString(fromText)
	to := cryptography.MustAddressFromString(toText)
	amount := big.NewInt(100000000)

	// String helpers parse Phantasma address text before loading VM bytes, keeping them equivalent to typed-address calls.
	typed := scriptbuilder.BeginScript().
		AllowGas(from, cryptography.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
		TransferTokens("SOUL", from, to, amount).
		TransferBalance("KCAL", from, to).
		TransferNFT("ART", from, to, big.NewInt(7)).
		Stake(from, amount).
		Unstake(from, amount).
		SpendGas(from).
		EndScript()

	text := scriptbuilder.BeginScript().
		AllowGasText(fromText, "NULL", big.NewInt(100000), big.NewInt(21000)).
		TransferTokensText("SOUL", fromText, toText, amount).
		TransferBalanceText("KCAL", fromText, toText).
		TransferNFTText("ART", fromText, toText, big.NewInt(7)).
		StakeText(fromText, amount).
		UnstakeText(fromText, amount).
		SpendGasText(fromText).
		EndScript()

	require.Equal(t, typed, text)
}

func TestScriptBuilderToTextHelpersParsePhantasmaDestination(t *testing.T) {
	fromText := "P2KM9FjYrDXnPPAynLXAHdQ8wYz8de9VbDeybrLepnw6C5x"
	toText := "P2KCineFiatZR8fDyU4pSfb3Bq1vtyW3zqBJi268YV5fH9e"
	from := cryptography.MustAddressFromString(fromText)
	to := cryptography.MustAddressFromString(toText)

	// Mixed typed/text helpers cover the C# string-destination overload while preserving Go's typed-address default.
	require.Equal(t,
		scriptbuilder.BeginScript().TransferTokens("SOUL", from, to, big.NewInt(1)).EndScript(),
		scriptbuilder.BeginScript().TransferTokensToText("SOUL", from, toText, big.NewInt(1)).EndScript(),
	)
	require.Equal(t,
		scriptbuilder.BeginScript().TransferNFT("ART", from, to, big.NewInt(7)).EndScript(),
		scriptbuilder.BeginScript().TransferNFTToText("ART", from, toText, big.NewInt(7)).EndScript(),
	)
}

func TestScriptBuilderMatchesSharedVector(t *testing.T) {
	keys, err := cryptography.FromWIF("L5UEVHBjujaR1721aZM5Zm5ayjDyamMZS9W35RE9Y9giRkdf3dVx")
	require.NoError(t, err)

	// This vector intentionally passes the vote address as a VM string, not as a binary Address argument.
	script := scriptbuilder.BeginScript().
		AllowGas(keys.Address(), cryptography.NullAddress(), big.NewInt(10000), big.NewInt(210000)).
		CallContract("consensus", "SingleVote", keys.Address().Text(), "system.nexus.protocol.version", 0).
		SpendGas(keys.Address()).
		EndScript()

	require.Equal(t, expectedConsensusSingleVoteScriptHex, strings.ToUpper(hex.EncodeToString(script)))
}

func TestScriptBuilderResolvesLabelsPerInstance(t *testing.T) {
	// Label resolution must be scoped to one builder so finished scripts do not leak labels into later builders.
	first := scriptbuilder.BeginScript().
		EmitJump(vm.JMP, "done", 0).
		EmitLoadString(0, "unused").
		EmitLabel("DONE").
		EndScript()

	target := int(binary.LittleEndian.Uint16(first[1:3]))
	require.Equal(t, len(first)-1, target)
	require.Equal(t, byte(vm.NOP), first[target-1])
	require.Equal(t, byte(vm.RET), first[target])

	second := scriptbuilder.BeginScript().EndScript()
	require.Equal(t, []byte{byte(vm.RET)}, second)
}

func TestScriptBuilderRejectsMissingLabel(t *testing.T) {
	// Unresolved labels are programmer errors and must fail before an invalid script can be broadcast.
	require.Panics(t, func() {
		scriptbuilder.BeginScript().EmitJump(vm.JMP, "missing", 0).EndScript()
	})
}

func TestScriptBuilderReturnsErrorsForInvalidArguments(t *testing.T) {
	// The checked finalizer is the non-panic path for user-supplied address text and dynamic arguments.
	_, err := scriptbuilder.BeginScript().
		TransferTokensText("SOUL", "not-an-address", "NULL", big.NewInt(1)).
		EndScriptWithError()
	require.Error(t, err)

	_, err = scriptbuilder.BeginScript().
		CallContract("test", "bad", struct{}{}).
		EndScriptWithError()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported type")

	_, err = scriptbuilder.BeginScript().
		CallContract("test", "nil", nil).
		EndScriptWithError()
	require.Error(t, err)
	require.Contains(t, err.Error(), "nil argument")
}

func TestEmitExtCallUsesExtCallOpcode(t *testing.T) {
	// Direct EXTCALL emission is required for low-level VM parity with C#/TS/C++ script vectors.
	script := scriptbuilder.BeginScript().EmitExtCall("Runtime.Time", 0).EndScript()

	require.GreaterOrEqual(t, len(script), 3)
	require.Equal(t, byte(vm.EXTCALL), script[len(script)-3])
	require.Equal(t, byte(0), script[len(script)-2])
	require.Equal(t, byte(vm.RET), script[len(script)-1])
}
