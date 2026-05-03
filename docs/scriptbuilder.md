# ScriptBuilder

`pkg/vm/script_builder` builds Phantasma VM scripts for transactions and read-only script invocation.

The builder intentionally follows the C#/TS/C++ SDK shape:

- `BeginScript()` / `NewScriptBuilder()` create an isolated builder instance;
- `EndScript()` appends `RET` and resolves labels;
- `ToScript()` resolves labels without appending `RET`;
- `CallContract(contract, method, args...)` pushes arguments, pushes the method name, switches to the contract context;
- `CallInterop(method, args...)` pushes arguments and emits `EXTCALL`.

Integer arguments are emitted as VM `Number` bytes compatible with .NET `BigInteger.ToByteArray()`. Array arguments are emitted through `CAST` and `PUT`, using the target register plus the next two temporary registers.

Raw `string` arguments are emitted as VM strings. When a contract or interop method expects a Phantasma address, pass `cryptography.Address`, `cryptography.MustAddressFromString(text)`, or one of the explicit `*Text` helpers.
Use `EndScriptWithError()` or `ToScript()` when address text, dynamic argument types, or labels come from external input. `EndScript()` keeps the compact helper style and panics on the same recorded builder errors.

## Common Transaction Script

```go
from := cryptography.MustAddressFromString("put sender address here")
to := cryptography.MustAddressFromString("put recipient address here")
amount := big.NewInt(100000000)

script := scriptbuilder.BeginScript().
    AllowGas(from, cryptography.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
    TransferTokens("SOUL", from, to, amount).
    SpendGas(from).
    EndScript()
```

## Labels

Labels are scoped to a single builder. Missing labels fail during `EndScript()` / `ToScript()`.

```go
script := scriptbuilder.BeginScript().
    EmitJump(vm.JMP, "done", 0).
    EmitLoadString(0, "unreachable").
    EmitLabel("done").
    EndScript()
```

## Extensions

Supported extension helpers:

- `AllowGas`
- `SpendGas`
- `MintTokens`
- `MintTokensText`
- `TransferTokens`
- `TransferTokensText`
- `TransferTokensToText`
- `TransferBalance`
- `TransferBalanceText`
- `TransferNFT`
- `TransferNFTText`
- `TransferNFTToText`
- `CrossTransferToken`
- `CrossTransferTokenText`
- `CrossTransferTokenToText`
- `CrossTransferNFT`
- `CrossTransferNFTText`
- `CrossTransferNFTToText`
- `Stake`
- `StakeText`
- `Unstake`
- `UnstakeText`
- `CallNFT`

`TransferTokensToText` and `TransferNFTToText` parse the `to` argument as Phantasma address text. Cross-chain text helpers parse the Phantasma destination-chain/sender addresses and keep the destination account as VM string for non-Phantasma destination formats.
