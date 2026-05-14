package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	chain "github.com/phantasma-io/phantasma-sdk-go/pkg/blockchain"
	crypto "github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/domain"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/util"
	scriptbuilder "github.com/phantasma-io/phantasma-sdk-go/pkg/vm/script_builder"
)

func unstakeSoulToken(address crypto.Address, tokenAmount *big.Int) {
	// build script
	sb := scriptbuilder.BeginScript().
		AllowGas(address, crypto.NullAddress(), big.NewInt(100000), big.NewInt(21000)).
		Unstake(address, tokenAmount).
		SpendGas(address)
	script := sb.EndScript()

	// build tx
	expire := time.Now().UTC().Add(time.Second * time.Duration(30)).Unix()
	tx := chain.NewTransaction(netSelected, "main", script, uint32(expire), domain.SDKPayload)

	// sign tx
	if err := tx.Sign(keyPair); err != nil {
		fmt.Println("Signing transaction failed:", err)
		return
	}

	fmt.Println("Tx script: " + hex.EncodeToString(script))

	// encode tx as hex
	txHex := hex.EncodeToString(tx.Bytes())

	fmt.Println("Tx: " + txHex)

	if !PromptYNChoice("Send transaction?") {
		return
	}

	txHash, err := client.SendRawTransaction(rpcContext, txHex)
	if err != nil {
		fmt.Println("Broadcasting tx failed:", err)
		return
	}
	if util.ErrorDetect(txHash) {
		fmt.Println("Broadcasting tx failed:", txHash)
		return
	}
	fmt.Println("Tx successfully broadcasted! Tx hash: " + txHash)

	waitForTransactionResult(txHash)
}
