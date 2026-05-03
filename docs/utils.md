# Utils

## Signing

Signing helpers live in the domain-specific packages:

- `pkg/cryptography` signs classic VM transactions and provides address/key primitives.
- `pkg/carbon` signs Carbon `TxMsg` values with `SignTxMsg`, `SignAndSerializeTxMsg`, and `SignAndSerializeTxMsgHex`.
- `pkg/rpc` can broadcast signed transactions through `SendRawTransaction`, `SendCarbonTransaction`, `SignAndSendTransaction`, and `SignAndSendCarbonTransaction`.
