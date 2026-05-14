# Utils

## Signing

Signing helpers live in the domain-specific packages:

- `pkg/cryptography` signs classic VM transactions and provides address/key primitives.
- `pkg/carbon` signs Carbon `TxMsg` values with `SignTxMsg`, `SignAndSerializeTxMsg`, and `SignAndSerializeTxMsgHex`.
- `pkg/rpc` can broadcast signed transactions through context-first methods such as `SendRawTransaction(ctx, txHex)`, `SendCarbonTransaction(ctx, txHex)`, `SignAndSendTransaction(ctx, ...)`, and `SignAndSendCarbonTransaction(ctx, ...)`.

## Amounts and Addresses

`util.ConvertDecimalsBack` and `util.ConvertDecimalsBackEx` return explicit errors when user input would require rounding beyond the token decimal precision.

`cryptography.NewAddress`, `cryptography.NewPhantasmaKeys`, and `cryptography.GeneratePhantasmaKeys` return errors for invalid input instead of panicking.
