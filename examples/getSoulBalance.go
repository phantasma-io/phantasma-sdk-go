package main

import (
	"fmt"
	"math/big"
)

func getSoulBalance(address string) (*big.Float, *big.Float) {
	// Calling "GetAccount" method to get token balances of the address
	account, err := client.GetAccount(rpcContext, address)
	if err != nil {
		fmt.Println("GetAccount call failed:", err)
		return big.NewFloat(0), big.NewFloat(0)
	}

	for i := 0; i < len(account.Balances); i += 1 {
		if account.Balances[i].Symbol == "SOUL" {
			return account.Balances[i].ConvertDecimalsToFloat(), account.Stakes.ConvertDecimalsToFloat()
		}
	}

	return big.NewFloat(0), big.NewFloat(0)
}
