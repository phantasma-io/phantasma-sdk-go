package scriptbuilder_test

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/vm"
	scriptbuilder "github.com/phantasma-io/phantasma-sdk-go/pkg/vm/script_builder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expectedConsensusSingleVoteScriptHex = "0D00030350340303000D000302102703000D000223220000000000000000000000000000000000000000000000000000000000000000000003000D000223220100AA53BE71FC41BC0889B694F4D6D03F7906A3D9A21705943CAF9632EEAFBB489503000D000408416C6C6F7747617303000D0004036761732D00012E010D0003010003000D00041D73797374656D2E6E657875732E70726F746F636F6C2E76657273696F6E03000D00042F50324B464579466576705166536157384734566A536D6857555A585234517247395951523148624D7054554370434C03000D00040A53696E676C65566F746503000D000409636F6E73656E7375732D00012E010D000223220100AA53BE71FC41BC0889B694F4D6D03F7906A3D9A21705943CAF9632EEAFBB489503000D0004085370656E6447617303000D0004036761732D00012E010B"
const scriptBuilderFixtureSHA256 = "81907a6b1df095b84599d8f8d709623e20dadeca2082ab9dffef114c7d0015e0"

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

func TestScriptBuilderMatchesGoldenVectors(t *testing.T) {
	// The TSV was generated from the Gen2 C# SDK and covers high-level helpers
	// plus array/timestamp argument loading that had drifted between SDKs.
	for _, row := range scriptBuilderFixtureRows(t) {
		caseID, source, expectedHex := row[0], row[1], row[2]
		require.Equal(t, "csharp_sdk", source)
		require.Equal(t, expectedHex, strings.ToUpper(hex.EncodeToString(scriptBuilderVector(t, caseID))), caseID)
		require.NotEmpty(t, row[3], caseID)
	}
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

func scriptBuilderFixtureRows(t *testing.T) [][]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "vm_script_builder_vectors.tsv"))
	require.NoError(t, err)
	require.Equal(t, scriptBuilderFixtureSHA256, fmt.Sprintf("%x", sha256.Sum256(data)))

	rows := [][]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" || strings.HasPrefix(line, "case_id\t") {
			continue
		}
		parts := strings.Split(line, "\t")
		require.Len(t, parts, 4, line)
		rows = append(rows, parts)
	}
	return rows
}

func scriptBuilderVector(t *testing.T, caseID string) []byte {
	t.Helper()
	// Deterministic, non-funded fixture keys used only to reproduce the shared
	// Gen2 C# golden scripts.
	mainKeys, err := cryptography.FromWIF("L5UEVHBjujaR1721aZM5Zm5ayjDyamMZS9W35RE9Y9giRkdf3dVx")
	require.NoError(t, err)
	helperKeys, err := cryptography.FromWIF("KxMn2TgXukYaNXx7tEdjh7qB2YaMgeuKy47j4rvKigHhBuZWeP3r")
	require.NoError(t, err)
	address := helperKeys.Address()
	nullAddress := cryptography.NullAddress()

	switch caseID {
	case "consensus_single_vote":
		return scriptbuilder.BeginScript().
			AllowGas(mainKeys.Address(), nullAddress, big.NewInt(10000), big.NewInt(210000)).
			CallContract("consensus", "SingleVote", mainKeys.Address().Text(), "system.nexus.protocol.version", 0).
			SpendGas(mainKeys.Address()).
			EndScript()
	case "gas_transfer_spend":
		return scriptbuilder.BeginScript().
			AllowGas(address, nullAddress, big.NewInt(100000), big.NewInt(21000)).
			TransferTokens("SOUL", address, nullAddress, big.NewInt(100000000)).
			SpendGas(address).
			EndScript()
	case "mint_tokens":
		return scriptbuilder.BeginScript().MintTokens("SOUL", address, nullAddress, big.NewInt(1)).EndScript()
	case "transfer_balance":
		return scriptbuilder.BeginScript().TransferBalance("KCAL", address, nullAddress).EndScript()
	case "transfer_nft":
		return scriptbuilder.BeginScript().TransferNFT("ART", address, nullAddress, big.NewInt(42)).EndScript()
	case "cross_transfer_token":
		return scriptbuilder.BeginScript().CrossTransferToken(nullAddress, "SOUL", address, nullAddress, big.NewInt(1)).EndScript()
	case "cross_transfer_nft":
		return scriptbuilder.BeginScript().CrossTransferNFT(nullAddress, "ART", address, nullAddress, big.NewInt(7)).EndScript()
	case "stake_unstake":
		return scriptbuilder.BeginScript().Stake(address, big.NewInt(7)).Unstake(address, big.NewInt(8)).EndScript()
	case "call_nft":
		return scriptbuilder.BeginScript().CallNFT("ART", big.NewInt(7), "mint", address).EndScript()
	case "runtime_array_timestamp":
		return scriptbuilder.BeginScript().
			CallInterop("Runtime.Test", []interface{}{"alpha", big.NewInt(7)}, time.Unix(1778330400, 0).UTC()).
			EndScript()
	default:
		t.Fatalf("unhandled script vector: %s", caseID)
		return nil
	}
}
