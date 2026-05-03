# RPC

`pkg/rpc` wraps Phantasma JSON-RPC calls with typed response models from `pkg/rpc/response`.

## Client

```go
client := rpc.NewRPCTestnet()
client = rpc.NewRPCMainnet()
client = rpc.NewRPC("http://localhost:5172/rpc")
raw, err := client.CallContext(ctx, "getBlockHeight", "main")
```

## Core Calls

Existing RPC calls remain available:

- `GetAccount`, `GetAccounts`, `LookupName`
- `GetAddressTransactions`, `GetAddressTransactionCount`
- `GetBlockByHeight`, `GetBlockHeight`
- `InvokeRawScript`
- `SendRawTransaction`
- `GetTransaction`
- `GetTokens`, `GetToken`

## Carbon and Current SDK Parity Calls

The wrapper also includes the Carbon endpoints exposed by the current C#/TS SDKs:

- `GetVersion`
- `GetPhantasmaVMConfig`
- `GetChains`, `GetChain`, `GetNexus`
- `GetBlockByHash`, `GetLatestBlock`
- `GetBlockTransactionCountByHash`, `GetBlockTransactionCountByHashOnChain`
- `GetTransactionByBlockHashAndIndex`, `GetTransactionByBlockHashAndIndexOnChain`
- `GetContracts`, `GetContractByName`, `GetContractByAddress`
- `GetOrganization`, `GetOrganizationByName`, `GetOrganizations`
- `GetLeaderboard`
- `GetTokensByOwner`, `GetTokensByOwnerWithAddressType`, `GetTokenWithID`
- `GetTokenData`, `GetTokenBalance`, `GetTokenBalanceChecked`, `GetTokenBalanceWithAddressType`
- `GetTokenSeries`, `GetTokenSeriesByID`, `GetTokenNFTs`
- `GetAccountFungibleTokens`, `GetAccountNFTs`, `GetAccountOwnedTokens`, `GetAccountOwnedTokenSeries`
- `GetAccountsWithAddressType`, `GetAccountWithAddressType`
- `GetAccountFungibleTokensWithAddressType`, `GetAccountNFTsWithAddressType`, `GetAccountOwnedTokensWithAddressType`, `GetAccountOwnedTokenSeriesWithAddressType`
- `GetAuctionsCount`, `GetAuctions`, `GetAuction`
- `GetArchive`, `ReadArchive`, `WriteArchive`
- `GetNFT`, `GetNFTs`
- `SendCarbonTransaction`, `SignAndSendCarbonTransaction`
- `SignAndSendTransaction`, `SignAndSendTransactionWithExpiration`

Cursor-paginated methods return `response.CursorPaginatedResult[T]`. Page-paginated methods return `response.PaginatedResult[T]`.
Carbon token/series id filters are numeric (`uint64` for token ids, `uint32` for Carbon series ids). Pass `0` when the RPC endpoint should use its default/no-filter behavior. `GetTokenNFTsWithSeriesID` also accepts a Phantasma Series ID string filter.
`GetAccounts` accepts variadic addresses and `GetNFTs` accepts slices, then joins them for the wire call. Use `GetAccountsText` or `GetNFTsText` only when you already have the comma-separated RPC string. `WriteArchive` accepts raw bytes; use `WriteArchiveBase64` only when the block is already encoded.
Address-type variants accept `rpc.AddressTypePhantasma` or `rpc.AddressTypeCarbon` for RPC calls where C# exposes explicit Phantasma-vs-Carbon address interpretation.

## Carbon Broadcast

```go
signedTx, err := carbon.SignAndSerializeTxMsgHex(tx, keyPair)
if err != nil {
    log.Fatal(err)
}

hash, err := client.SendCarbonTransaction(signedTx)
if err != nil {
    log.Fatal(err)
}
```

For the common sign-and-broadcast flow:

```go
hash, err := client.SignAndSendCarbonTransaction(tx, keyPair)
if err != nil {
    log.Fatal(err)
}
```
