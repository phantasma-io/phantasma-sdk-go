# Carbon

`pkg/carbon` contains Phantasma Phoenix Carbon serialization, transaction, signing, token and VM schema helpers.

## Serialization

Use `Writer` and `Reader` when implementing custom Carbon blobs:

```go
w := carbon.NewWriter()
w.Write8U(42)
symbol, err := carbon.NewSmallString("SOUL")
if err != nil {
    log.Fatal(err)
}
symbol.WriteCarbon(w)
bytes := w.Bytes()

r := carbon.NewReader(bytes)
value := r.Read8U()
```

Supported primitives include fixed-size byte arrays, signed/unsigned integers, Carbon BigInt, `IntX`, `SmallString`, byte arrays, and arrays of Carbon blobs.

## Transactions

Core transaction models:

- `TxMsg`
- `TxMsgCall`
- `TxMsgTransferFungible`
- `TxMsgTransferNonFungible`
- `TxMsgBurnFungible`
- `TxMsgBurnNonFungible`
- `TxMsgTrade`
- `SignedTxMsg`
- `Witness`
- `ChainConfig`
- `GasConfig`
- `TokenListing`
- `MarketConfig`

Signing helpers:

```go
signed, err := carbon.SignTxMsg(tx, keyPair)
if err != nil {
    log.Fatal(err)
}
encoded, err := carbon.SignAndSerializeTxMsg(tx, keyPair)
if err != nil {
    log.Fatal(err)
}
hexEncoded, err := carbon.SignAndSerializeTxMsgHex(tx, keyPair)
if err != nil {
    log.Fatal(err)
}
_, _, _ = signed, encoded, hexEncoded
```

## Token Module

Token builders:

- `BuildCreateTokenTx`
- `BuildCreateTokenTxAndSign`
- `BuildCreateTokenTxAndSignHex`
- `BuildCreateTokenSeriesTx`
- `BuildCreateTokenSeriesTxAndSign`
- `BuildCreateTokenSeriesTxAndSignHex`
- `BuildMintNonFungibleTx`
- `BuildMintNonFungibleTxAndSign`
- `BuildMintNonFungibleTxAndSignHex`
- `BuildMintPhantasmaNonFungibleTx`
- `BuildMintPhantasmaNonFungibleTxAndSign`
- `BuildMintPhantasmaNonFungibleTxAndSignHex`
- `BuildMintPhantasmaNonFungibleSingleTx`
- `BuildMintPhantasmaNonFungibleSingleTxAndSign`
- `BuildMintPhantasmaNonFungibleSingleTxAndSignHex`

Argument/result blobs:

- `MintFungibleArgs`
- `MintNonFungibleArgs`
- `MintPhantasmaNonFungibleArgs`
- `TransferFungibleArgs`
- `TransferNonFungibleArgs`
- `BurnFungibleArgs`
- `BurnNonFungibleArgs`
- `CreateTokenSeriesArgs`
- `CreateMintedTokenSeriesArgs`
- `UpdateTokenMetadataArgs`
- `UpdateSeriesMetadataArgs`
- `MintNonFungibleResult`
- `PhantasmaNFTMintResult`
- `ParseCreateTokenResult`
- `ParseCreateTokenSeriesResult`
- `ParseMintNonFungibleResult`
- `ParseMintPhantasmaNonFungibleResult`

Metadata helpers:

- `PrepareStandardTokenSchemas`
- `BuildTokenSchemasFromFields`
- `ParseTokenSchemasJSON`
- `TokenSchemasFromJSON`
- `SerializeTokenSchemas`
- `SerializeTokenSchemasHex`
- `BuildAndSerializeTokenSchemas`
- `VerifyTokenSchemas`
- `BuildTokenMetadata`
- `MustBuildTokenMetadata`
- `BuildTokenSeriesMetadata`
- `MustBuildTokenSeriesMetadata`
- `BuildNFTRom`
- `MustBuildNFTRom`
- `BuildPhantasmaNFTPublicMintSchema`
- `BuildPhantasmaNFTRom`
- `MustBuildPhantasmaNFTRom`
- `BuildTokenInfo`
- `BuildSeriesInfo`
- `MustBuildSeriesInfo`

`Build...` helpers return validation errors for user-supplied metadata and schema mismatches. `MustBuild...` helpers are for constants/tests and panic on the same invalid input.

Address helpers:

- `Bytes32FromPublicKey`
- `Bytes32FromPhantasmaAddress`
- `Bytes32FromPhantasmaAddressText`
- `MustBytes32FromPhantasmaAddressText`
- `GetNFTAddress`
- `UnpackNFTInstanceID`

## Example

```go
schemas := carbon.PrepareStandardTokenSchemas(false)
serializedSchemas := carbon.SerializeTokenSchemas(schemas)

metadata, err := carbon.BuildTokenMetadata(map[string]string{
    "name": "Example Art",
    "icon": "data:image/png;base64,iVBORw0KGgo=",
    "url": "https://example.com",
    "description": "Example Carbon NFT token",
})
if err != nil {
    log.Fatal(err)
}

owner, err := carbon.Bytes32FromPublicKey(keyPair.PublicKey())
if err != nil {
    log.Fatal(err)
}

tokenInfo, err := carbon.BuildTokenInfo(
    "ART",
    carbon.NewIntX(big.NewInt(1000)),
    true,
    0,
    owner,
    metadata,
    serializedSchemas,
)
if err != nil {
    log.Fatal(err)
}

signedTx, err := carbon.BuildCreateTokenTxAndSign(
    tokenInfo,
    keyPair,
    carbon.DefaultCreateTokenFeeOptions(),
    100_000_000,
    time.Now().UTC().Add(20*time.Minute).UnixMilli(),
)
```

For deterministic Phantasma NFT mints, build the public ROM without chain-owned `_i` and nested `rom` fields:

```go
publicRom, err := carbon.BuildPhantasmaNFTRom(schemas.ROM, []carbon.MetadataField{
    {Name: "name", Value: "My NFT #1"},
    {Name: "description", Value: "This is my first NFT!"},
    {Name: "imageURL", Value: "https://example.com/nft.png"},
    {Name: "infoURL", Value: "https://example.com/nft"},
    {Name: "royalties", Value: int32(0)},
})
if err != nil {
    log.Fatal(err)
}

receiver, err := carbon.Bytes32FromPhantasmaAddressText("put receiver address here")
if err != nil {
    log.Fatal(err)
}

signedTx, err := carbon.BuildMintPhantasmaNonFungibleSingleTxAndSign(
    42,
    big.NewInt(777),
    keyPair,
    receiver,
    publicRom,
    nil,
    carbon.DefaultMintNFTFeeOptions(),
    100_000_000,
    time.Now().UTC().Add(20*time.Minute).UnixMilli(),
)
```
