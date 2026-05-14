package main

import (
	"fmt"
	"time"
)

func waitForTransactionResult(txHash string) {
	for {
		txResult, err := client.GetTransaction(rpcContext, txHash)
		if err != nil {
			fmt.Println("GetTransaction failed:", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		fmt.Println("Tx state: " + fmt.Sprint(txResult.State))

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
}
