<p align="center">
<img src="./.github/phantasma-go.jpg" width="300px" alt="logo">
</p>
<p align="center">
  <b>Go</b> SDK for the <a href="https://phantasma.io">Phantasma</a> blockchain.
</p>

<hr />

![License](https://img.shields.io/github/license/phantasma-io/phantasma-go.svg?style=popout)

# Overview

This project aims to be an easy to use SDK for the Phantasma blockchain.

# Documentation

## Installation

Requires Go 1.25 or newer. The module is developed with the Go 1.26 toolchain.

PhantasmaGo is distributed as a library that includes all the functionality provided.

```
go get -u github.com/phantasma-io/phantasma-go
```

## Getting started
To start interacting with Phantasma blockchain you need to choose network you are planning to use (mainnet or testnet) and create corresponding RPC client.

Creation of testnet RPC client:
```
client = rpc.NewRPCTestnet()
```

Creation of mainnet RPC client:
```
client = rpc.NewRPCMainnet()
```

To create a new key pair structure from private key in WIF format use following code:
```
keyPair, err := cryptography.FromWIF("put WIF here")
if err != nil {
    log.Fatalf("creating key pair: %v", err)
}
```

To get detailed description of tokens deployed on the chain you can use following code:

```
var chainTokens []response.TokenResult

func getChainToken(symbol string) (response.TokenResult, bool) {
    for _, t := range chainTokens {
        if t.Symbol == symbol {
            return t, true
        }
    }

    return response.TokenResult{}, false
}

chainTokens, err = client.GetTokens(false)
if err != nil {
    log.Fatal(err)
}
```

This will allow you to get token characteristics this way:

```
t, ok := getChainToken("SOUL")
if !ok {
    log.Fatal("token SOUL not found")
}
if t.IsFungible() {
    fmt.Println("Token SOUL is fungible")
}
```

Code samples in the following sections of this documentation use `client` and `keyPair` structures and method `getChainToken` which should be initialized in advance.

## Carbon transactions and token builders

The SDK includes `pkg/carbon` for Phantasma Phoenix Carbon transaction serialization and token-module calls. This package mirrors the C#/TS/C++ SDK model:

- fixed-width Carbon types: `Bytes16`, `Bytes32`, `Bytes64`, `SmallString`, `IntX`;
- `TxMsg`, `SignedTxMsg`, witnesses, call sections, and trade payloads;
- VM schema structures used by token metadata;
- token-module argument/result blobs for create token, create series, mint, transfer, burn and metadata update calls;
- high-level helpers for token metadata, NFT ROM/RAM, standard NFT schemas, transaction signing, and NFT address/instance-id conversion.

Example: build and sign a Carbon NFT mint transaction:

```go
signer, err := cryptography.FromWIF("put WIF here")
if err != nil {
    log.Fatal(err)
}

receiver, err := carbon.Bytes32FromPhantasmaAddressText("put receiver address here")
if err != nil {
    log.Fatal(err)
}

schemas := carbon.PrepareStandardTokenSchemas(false)
rom, err := carbon.BuildNFTRom(schemas.ROM, big.NewInt(1), []carbon.MetadataField{
    {Name: "name", Value: "Example NFT"},
    {Name: "description", Value: "Minted with phantasma-go Carbon helpers"},
    {Name: "imageURL", Value: "https://example.com/nft.png"},
    {Name: "infoURL", Value: "https://example.com/nft"},
    {Name: "royalties", Value: int32(0)},
})
if err != nil {
    log.Fatal(err)
}

signedTx, err := carbon.BuildMintNonFungibleTxAndSignHex(
    42,                 // Carbon token id
    1,                  // Carbon series id
    signer,
    receiver,
    rom,
    nil,                // RAM
    carbon.DefaultMintNFTFeeOptions(),
    100_000_000,
    time.Now().UTC().Add(20*time.Minute).UnixMilli(),
)
if err != nil {
    log.Fatal(err)
}

txHash, err := client.SendCarbonTransaction(signedTx)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Carbon tx hash:", txHash)
```

See `docs/carbon.md` for the package-level API map.
Carbon `Build...` helpers return validation errors for user input; `MustBuild...` variants are available only when panic-on-invalid-input is intentional.

## Carbon RPC wrappers

The RPC client now exposes the Carbon endpoints used by the current C#/TS SDKs. Important additions include:

- chain/block lookup: `GetChains`, `GetChain`, `GetNexus`, `GetBlockByHash`, `GetLatestBlock`, `GetTransactionByBlockHashAndIndex`;
- contract/organization lookup: `GetContracts`, `GetContractByName`, `GetContractByAddress`, `GetOrganization`, `GetOrganizationByName`, `GetOrganizations`;
- Carbon token views: `GetTokensByOwner`, `GetTokenWithID`, `GetTokenSeries`, `GetTokenSeriesByID`, `GetTokenNFTs`, `GetAccountFungibleTokens`, `GetAccountNFTs`, `GetAccountOwnedTokens`, `GetAccountOwnedTokenSeries`;
- archive and auction helpers: `GetArchive`, `ReadArchive`, `WriteArchive`, `GetAuctionsCount`, `GetAuctions`, `GetAuction`;
- transaction broadcast: `SendCarbonTransaction`, `SignAndSendCarbonTransaction`.

Cursor-paginated endpoints return `response.CursorPaginatedResult[T]`.
Carbon token id filters use `uint64`, Carbon series id filters use `uint32`, and `0` means no Carbon id filter. `GetTokenNFTsWithSeriesID` accepts the Phantasma Series ID string filter exposed by current RPC nodes.
Use `client.CallContext(ctx, method, params...)` for low-level calls that need request cancellation or deadlines.

## Script Builder

Building a script is the most important part of interacting with the Phantasma blockchain. Without a proper script, the Phantasma blockchain will not know what you are trying to do.

These functions, `CallContract` and `CallInterop`, are your bread and butter for creating new scripts.

```
func (s ScriptBuilder) CallContract(contractName, method string, args ...interface{}) ScriptBuilder
```

```
func (s ScriptBuilder) CallInterop(method string, args ...interface{}) ScriptBuilder
```

You can find out all the different `CallInterop` functions below.

For `CallContract`, you will have to look through the ABI's of all the different smart contracts currently deployed on the Phantasma 'mainnet': [Link Here](https://explorer.phantasma.info/en/nexus?tab=contracts). To see all methods of a contract, for example `stake`, you can check it with explorer: [Link Here](https://explorer.phantasma.info/en/contract?id=stake&tab=methods).

The Go builder now resolves labels per instance, emits integer values as VM `Number` payloads, supports array arguments, and matches the C#/TS/C++ shared script vectors.
Address arguments are intentionally typed as `cryptography.Address` in high-level helpers. Raw `string` arguments passed to `CallContract` or `CallInterop` are emitted as VM strings; use `cryptography.MustAddressFromString(text)` or the `*Text` helpers when the ABI expects a Phantasma address.

### Examples

Following code generates script to transfer `tokenAmount` amount of token `tokenSymbol` from wallet `from` to wallet `to`
```
from := cryptography.MustAddressFromString("put sender address here") // Phantasma address, starting with capital 'P'
to := cryptography.MustAddressFromString("put recipient address here") // Phantasma address, starting with capital 'P'
tokenAmount := big.NewInt(1000000000) // Token amount in the form of big integer
tokenSymbol := "SOUL"

sb := scriptbuilder.BeginScript()
script := sb.CallContract("gas", "AllowGas", from, cryptography.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
    CallInterop("Runtime.TransferTokens", from, to, tokenSymbol, tokenAmount).
    CallContract("gas", "SpendGas", from).
    EndScript()
```

And here we generate script to make a call which does not require transaction, for this we use `CallContract` method:

```
address := cryptography.MustAddressFromString("put caller address here") // Phantasma address, starting with capital 'P'
tokenAmount := big.NewInt(1000000000) // Token amount in the form of big integer

sb := scriptbuilder.BeginScript().
    CallContract("gas", "AllowGas", address, cryptography.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
    CallContract("stake", "Stake", address, tokenAmount).
    CallContract("gas", "SpendGas", address)
script := sb.EndScript()
```

## Script Builder Extensions

For some widely used contract calls SDK has special extension methods which make code more compact. The typed-address helpers are `AllowGas`, `SpendGas`, `MintTokens`, `Stake`, `Unstake`, `TransferTokens`, `TransferBalance`, `TransferNFT`, `CrossTransferToken`, `CrossTransferNFT`, and `CallNFT`.

String-address convenience helpers are `AllowGasText`, `SpendGasText`, `MintTokensText`, `StakeText`, `UnstakeText`, `TransferTokensText`, `TransferBalanceText`, `TransferNFTText`, `CrossTransferTokenText`, and `CrossTransferNFTText`.

The ordinary `*Text` helpers parse Phantasma address text before emitting VM address bytes. Cross-chain text helpers parse the Phantasma destination-chain/sender addresses and keep the destination account as a VM string for non-Phantasma destination formats.

### Examples

We can rewrite examples from previous section using `AllowGas()` and `SpendGas()` extensions:

```
sb := scriptbuilder.BeginScript()
script := sb.AllowGas(from, cryptography.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
    CallInterop("Runtime.TransferTokens", from, to, tokenSymbol, tokenAmount).
    SpendGas(from).
    EndScript()
```
```
sb := scriptbuilder.BeginScript().
    AllowGas(address, cryptography.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
    CallContract("stake", "Stake", address, tokenAmount).
    SpendGas(address)
script := sb.EndScript()
```

We can also rewrite main contract calls in these examples:

```
sb := scriptbuilder.BeginScript()
script := sb.AllowGas(from, cryptography.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
    TransferTokens(tokenSymbol, from, to, tokenAmount).
    SpendGas(from).
    EndScript()
```
```
sb := scriptbuilder.BeginScript().
    AllowGas(address, cryptography.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
    Stake(address, tokenAmount).
    SpendGas(address)
script := sb.EndScript()
```

## InvokeRawScript and decoding the result

Scripts which does not require transaction can be sent to the chain directly using `InvokeRawScript()` call.

Here's an example of such call to get SoulMaster count from the chain:

```
// Build script
sb := scriptbuilder.BeginScript().
    CallContract("stake", "GetMasterCount")
script := sb.EndScript()

// Before sending script to the chain we need to encode it into Base16 encoding (HEX)
encodedScript := hex.EncodeToString(script)

// Make the call itself
result, err := client.InvokeRawScript("main", encodedScript)

if err != nil {
    log.Fatalf("script invocation failed: %v", err)
}

// `DecodeResultWithError()` decodes HEX-encoded byte array result, stored in `.Result` field, into `vm.VMObject` structure
// `AsNumber()` returns value stored in `vm.VMObject` structure, in `.Data` field, as a *big.Int number (in our case value is stored in `vm.VMObject` as big integer serialized into byte array)
value, err := result.DecodeResultWithError()
if err != nil {
    log.Fatalf("script result decoding failed: %v", err)
}
fmt.Println("Current SoulMasters count: ", value.AsNumber().String())
```

## Building and sending transaction

### Building transaction
To build a transaction you will first need to build a script.

Note, building a transaction is for transactional scripts only. Non transactional scripts should use the RPC function `InvokeRawScript()`.

```
// Build script
sb := scriptbuilder.BeginScript()
script := sb.AllowGas(keyPair.Address(), cryptography.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
    TransferTokens(tokenSymbol, keyPair.Address(), to, tokenAmount).
    SpendGas(keyPair.Address()).
    EndScript()

// Build transaction
expire := time.Now().UTC().Add(time.Second * time.Duration(30)).Unix()
tx := blockchain.NewTransaction(netSelected, "main", script, uint32(expire), domain.SDKPayload)

// Sign transaction
tx.Sign(keyPair)

// Before sending script to the chain we need to encode it into Base16 encoding (HEX)
txHex := hex.EncodeToString(tx.Bytes())
```

### Sending transaction

Here we send transaction prepared in previous block of code and stored as HEX in `txHex` variable.

```
txHash, err := client.SendRawTransaction(txHex)
if err != nil {
    log.Fatalf("broadcasting tx failed: %v", err)
} else {
    if util.ErrorDetect(txHash) {
        log.Fatalf("broadcasting tx failed: %s", txHash)
    } else {
        fmt.Println("Tx successfully broadcasted! Tx hash: " + txHash)
    }
}
```

### Waiting for transaction execution result

We need to wait for transaction to be minted on the chain to get its status:

```
for {
    txResult, _ := client.GetTransaction(txHash)

    if txResult.StateIsSuccess() {
        fmt.Println("Transaction was successfully minted, tx hash: " + fmt.Sprint(txResult.Hash))
        break // Funds were transferred successfully
    }
    if txResult.StateIsFault() {
        fmt.Println("Transaction failed, tx hash: " + fmt.Sprint(txResult.Hash))
        break // Funds were not transferred, transaction failed
    }

    time.Sleep(200 * time.Millisecond)
}
```

## Staking SOUL token

Following code shows how to stake SOUL token:

```
// Build script
sb := scriptbuilder.BeginScript().
    AllowGas(address, cryptography.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
    Stake(address, tokenAmount).
    SpendGas(address)
script := sb.EndScript()

// Build transaction
expire := time.Now().UTC().Add(time.Second * time.Duration(30)).Unix()
tx := chain.NewTransaction(netSelected, "main", script, uint32(expire), domain.SDKPayload)

// Sign transaction
tx.Sign(keyPair)

// Before sending script to the chain we need to encode it into Base16 encoding (HEX)
txHex := hex.EncodeToString(tx.Bytes())

txHash, err := client.SendRawTransaction(txHex)
if err != nil {
    log.Fatalf("broadcasting tx failed: %v", err)
} else {
    if util.ErrorDetect(txHash) {
        log.Fatalf("broadcasting tx failed: %s", txHash)
    } else {
        fmt.Println("Tx successfully broadcasted! Tx hash: " + txHash)
    }
}

for {
    txResult, _ := client.GetTransaction(txHash)

    if txResult.StateIsSuccess() {
        fmt.Println("Transaction was successfully minted, tx hash: " + fmt.Sprint(txResult.Hash))
        break // Funds were transferred successfully
    }
    if txResult.StateIsFault() {
        fmt.Println("Transaction failed, tx hash: " + fmt.Sprint(txResult.Hash))
        break // Funds were not transferred, transaction failed
    }

    time.Sleep(200 * time.Millisecond)
}
```

## Scanning the blockchain for incoming transactions

In the following code we monitor the blockchain by checking all the new blocks minted on the blockchain and waiting for `TokenReceive` event for given address. This event for address means that address has received some tokens.

```
func onTransactionReceived(address, symbol, amount string) {
    fmt.Printf("Address %s received %s %s\n", address, amount, symbol)
}

func waitForIncomingTransfers(address string) {
    // Get current block height
    height, _ := client.GetBlockHeight("main")

    for {
        // Get block's data by its height
        block, err := client.GetBlockByHeight("main", height.String())
        if err != nil {
            log.Fatalf("GetBlockByHeight failed: %v", err)
        }

        // Iterate throough all transactions in the block
        for _, tx := range block.Txs {
            // Skip failed trasactions
            if !tx.StateIsSuccess() {
                continue
            }

            // Iterate throough all events in the transaction
            for _, e := range tx.Events {

                if e.Kind == event.TokenReceive.String() && e.Address == address {
                    // We found TokenReceive event for given address

                    // Decode event data into event.TokenEventData structure
                    decoded, _ := hex.DecodeString(e.Data)
                    data := io.Deserialize[*event.TokenEventData](decoded, &event.TokenEventData{})

                    // Apply decimals to the token amount
                    t, ok := getChainToken(data.Symbol)
                    if !ok {
                        log.Fatalf("token %s not found", data.Symbol)
                    }
                    tokenAmount := util.ConvertDecimals(data.Value, int(t.Decimals))

                    // Call our callback function
                    onTransactionReceived(e.Address, data.Symbol, tokenAmount)
                }
            }
        }

        // Wait for next block to appear on the blockchain
        for {
            newHeight, _ := client.GetBlockHeight("main")
            if newHeight.Cmp(height) == 1 {
                // New block was minted (at least 1 new block)
                height = height.Add(height, big.NewInt(1))
                break
            }

            // Wait 200 milliseconds before making next RPC call
            time.Sleep(200 * time.Millisecond)
        }
    }
}
```

## Examples

This repository has `examples` folder with some code which can be easily reused. Examples are grouped into a single console application.

To run this application switch to `examples` folder and run:

```
go run .
```

or

```
sh run.sh
```

Application entry point is `main()` function in `main.go` source file. Once launched it will display the following menu:

![image](https://github.com/user-attachments/assets/42c7f5a5-41ea-4dfa-9a05-45028c563be6)

Wallet submenu:

![image](https://github.com/user-attachments/assets/c6484554-3be8-415d-a8a0-1d885d0f9a37)

Chain stats submenu:

![image](https://github.com/user-attachments/assets/7a1c0fec-9b92-4a8a-8e16-cd54f42b066f)


# Contributing

Feel free to contribute to this project after reading the
[contributing guidelines](CONTRIBUTING.md).

Before starting to work on a certain topic, create an new issue first,
describing the feature/topic you are going to implement.

# Contact

- Get in contact with us on the [Phantasma Discord](https://discord.gg/JzSnmFZCcD)

# License

- Open-source [MIT](LICENSE.md)
