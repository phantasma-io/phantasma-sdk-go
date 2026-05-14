package main

import (
	"fmt"
	"math/big"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/rpc/response"
)

func printBalance(address string) (int, []response.BalanceResult) {
	// Calling "GetAccount" method to get token balances of the address
	account, err := client.GetAccount(rpcContext, address)
	if err != nil {
		fmt.Println("GetAccount call failed:", err)
		return 0, nil
	}

	fmt.Println("Balances:")

	fmt.Println("- Fungible tokens:")
	j := 1
	for i := 0; i < len(account.Balances); i += 1 {
		t, ok := getChainToken(account.Balances[i].Symbol)
		if !ok {
			fmt.Printf("Token %s metadata not found\n", account.Balances[i].Symbol)
			continue
		}
		if t.IsFungible() {
			if account.Balances[i].Symbol == "SOUL" {
				unstakedSoul := account.Balances[i].ConvertDecimalsToFloat()
				stakedSoul := account.Stakes.ConvertDecimalsToFloat()
				fmt.Printf("#%02d: %s balance: %s [not staked: %s | staked: %s]\n", j, account.Balances[i].Symbol, (new(big.Float).Add(unstakedSoul, stakedSoul)).String(), unstakedSoul.String(), stakedSoul.String())
			} else {
				fmt.Printf("#%02d: %s balance: %s\n", j, account.Balances[i].Symbol, account.Balances[i].ConvertDecimals())
			}
			j += 1
		}
	}

	fmt.Println("- Non-fungible tokens (NFTs):")
	for i := 0; i < len(account.Balances); i += 1 {
		t, ok := getChainToken(account.Balances[i].Symbol)
		if !ok {
			continue
		}
		if !t.IsFungible() {
			fmt.Printf("#%02d: %s balance: %s\n", j, account.Balances[i].Symbol, account.Balances[i].ConvertDecimals())
			j += 1
		}
	}

	return len(account.Balances), account.Balances
}
