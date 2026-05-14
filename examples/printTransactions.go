package main

import (
	"fmt"
	"time"
)

func printTransactionsCount(address string) {
	// Calling "GetAddressTransactionCount" method to get transactions for the address
	transactionCount, err := client.GetAddressTransactionCount(rpcContext, address, "main")
	if err != nil {
		fmt.Println("GetAddressTransactionCount call failed:", err)
		return
	}
	fmt.Println("Transactions count: ", transactionCount)
}

func printTransactions(address string, page, pageSize int) {
	// Calling "GetAddressTransactions" method to get transactions for the address
	transactions, err := client.GetAddressTransactions(rpcContext, address, page, pageSize)
	if err != nil {
		fmt.Println("GetAddressTransactions call failed:", err)
		return
	}
	fmt.Println("Transactions:")
	txs := transactions.Result.Txs
	for i := 0; i < len(txs); i += 1 {
		fmt.Println("#", i+1, ": ", txs[i].Hash, " timestamp: ", time.Unix(int64(txs[i].Timestamp), 0))
	}
}
