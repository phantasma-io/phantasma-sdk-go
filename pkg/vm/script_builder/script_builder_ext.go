package scriptbuilder

import (
	"math/big"
	"strings"

	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
)

func addressFromText(text string) (cryptography.Address, error) {
	if strings.EqualFold(strings.TrimSpace(text), "NULL") {
		return cryptography.NullAddress(), nil
	}
	return cryptography.FromString(text)
}

// AllowGas emits a gas AllowGas contract call.
func (s ScriptBuilder) AllowGas(from, to cryptography.Address, gasPrice, gasLimit *big.Int) ScriptBuilder {
	return s.CallContract("gas", "AllowGas", from, to, gasPrice, gasLimit)
}

// AllowGasText parses address text and emits a gas AllowGas contract call.
func (s ScriptBuilder) AllowGasText(from, to string, gasPrice, gasLimit *big.Int) ScriptBuilder {
	fromAddress, err := addressFromText(from)
	if err != nil {
		return s.withError(err)
	}
	toAddress, err := addressFromText(to)
	if err != nil {
		return s.withError(err)
	}
	return s.AllowGas(fromAddress, toAddress, gasPrice, gasLimit)
}

// SpendGas emits a gas SpendGas contract call.
func (s ScriptBuilder) SpendGas(address cryptography.Address) ScriptBuilder {
	return s.CallContract("gas", "SpendGas", address)
}

// SpendGasText parses address text and emits a gas SpendGas contract call.
func (s ScriptBuilder) SpendGasText(address string) ScriptBuilder {
	parsed, err := addressFromText(address)
	if err != nil {
		return s.withError(err)
	}
	return s.SpendGas(parsed)
}

// MintTokens emits a Runtime.MintTokens interop call.
func (s ScriptBuilder) MintTokens(symbol string, from, to cryptography.Address, amount *big.Int) ScriptBuilder {
	return s.CallInterop("Runtime.MintTokens", from, to, symbol, amount)
}

// MintTokensText parses address text and emits a Runtime.MintTokens interop call.
func (s ScriptBuilder) MintTokensText(symbol string, from, to string, amount *big.Int) ScriptBuilder {
	fromAddress, err := addressFromText(from)
	if err != nil {
		return s.withError(err)
	}
	toAddress, err := addressFromText(to)
	if err != nil {
		return s.withError(err)
	}
	return s.MintTokens(symbol, fromAddress, toAddress, amount)
}

// TransferTokensToText parses the destination address and emits a Runtime.TransferTokens interop call.
func (s ScriptBuilder) TransferTokensToText(symbol string, from cryptography.Address, to string, amount *big.Int) ScriptBuilder {
	toAddress, err := addressFromText(to)
	if err != nil {
		return s.withError(err)
	}
	return s.TransferTokens(symbol, from, toAddress, amount)
}

// Stake emits a stake contract Stake call.
func (s ScriptBuilder) Stake(address cryptography.Address, amount *big.Int) ScriptBuilder {
	return s.CallContract("stake", "Stake", address, amount)
}

// StakeText parses address text and emits a stake contract Stake call.
func (s ScriptBuilder) StakeText(address string, amount *big.Int) ScriptBuilder {
	parsed, err := addressFromText(address)
	if err != nil {
		return s.withError(err)
	}
	return s.Stake(parsed, amount)
}

// Unstake emits a stake contract Unstake call.
func (s ScriptBuilder) Unstake(address cryptography.Address, amount *big.Int) ScriptBuilder {
	return s.CallContract("stake", "Unstake", address, amount)
}

// UnstakeText parses address text and emits a stake contract Unstake call.
func (s ScriptBuilder) UnstakeText(address string, amount *big.Int) ScriptBuilder {
	parsed, err := addressFromText(address)
	if err != nil {
		return s.withError(err)
	}
	return s.Unstake(parsed, amount)
}

// TransferTokens emits a Runtime.TransferTokens interop call.
func (s ScriptBuilder) TransferTokens(symbol string, from, to cryptography.Address, amount *big.Int) ScriptBuilder {
	return s.CallInterop("Runtime.TransferTokens", from, to, symbol, amount)
}

// TransferTokensText parses address text and emits a Runtime.TransferTokens interop call.
func (s ScriptBuilder) TransferTokensText(symbol string, from, to string, amount *big.Int) ScriptBuilder {
	fromAddress, err := addressFromText(from)
	if err != nil {
		return s.withError(err)
	}
	toAddress, err := addressFromText(to)
	if err != nil {
		return s.withError(err)
	}
	return s.TransferTokens(symbol, fromAddress, toAddress, amount)
}

// TransferBalance emits a Runtime.TransferBalance interop call.
func (s ScriptBuilder) TransferBalance(symbol string, from, to cryptography.Address) ScriptBuilder {
	return s.CallInterop("Runtime.TransferBalance", from, to, symbol)
}

// TransferBalanceText parses address text and emits a Runtime.TransferBalance interop call.
func (s ScriptBuilder) TransferBalanceText(symbol string, from, to string) ScriptBuilder {
	fromAddress, err := addressFromText(from)
	if err != nil {
		return s.withError(err)
	}
	toAddress, err := addressFromText(to)
	if err != nil {
		return s.withError(err)
	}
	return s.TransferBalance(symbol, fromAddress, toAddress)
}

// TransferNFT emits a Runtime.TransferToken interop call for a non-fungible token.
func (s ScriptBuilder) TransferNFT(symbol string, from, to cryptography.Address, tokenID *big.Int) ScriptBuilder {
	return s.CallInterop("Runtime.TransferToken", from, to, symbol, tokenID)
}

// TransferNFTText parses address text and emits a Runtime.TransferToken interop call.
func (s ScriptBuilder) TransferNFTText(symbol string, from, to string, tokenID *big.Int) ScriptBuilder {
	fromAddress, err := addressFromText(from)
	if err != nil {
		return s.withError(err)
	}
	toAddress, err := addressFromText(to)
	if err != nil {
		return s.withError(err)
	}
	return s.TransferNFT(symbol, fromAddress, toAddress, tokenID)
}

// TransferNFTToText parses the destination address and emits a Runtime.TransferToken interop call.
func (s ScriptBuilder) TransferNFTToText(symbol string, from cryptography.Address, to string, tokenID *big.Int) ScriptBuilder {
	toAddress, err := addressFromText(to)
	if err != nil {
		return s.withError(err)
	}
	return s.TransferNFT(symbol, from, toAddress, tokenID)
}

// CrossTransferToken emits a Runtime.SendTokens interop call.
func (s ScriptBuilder) CrossTransferToken(destinationChain cryptography.Address, symbol string, from, to cryptography.Address, amount *big.Int) ScriptBuilder {
	return s.CallInterop("Runtime.SendTokens", destinationChain, from, to, symbol, amount)
}

// CrossTransferTokenText parses address text and emits a Runtime.SendTokens interop call.
func (s ScriptBuilder) CrossTransferTokenText(destinationChain string, symbol string, from string, to string, amount *big.Int) ScriptBuilder {
	destinationAddress, err := addressFromText(destinationChain)
	if err != nil {
		return s.withError(err)
	}
	fromAddress, err := addressFromText(from)
	if err != nil {
		return s.withError(err)
	}
	return s.CrossTransferTokenToText(destinationAddress, symbol, fromAddress, to, amount)
}

// CrossTransferTokenToText emits a Runtime.SendTokens interop call with destination account as VM string.
func (s ScriptBuilder) CrossTransferTokenToText(destinationChain cryptography.Address, symbol string, from cryptography.Address, to string, amount *big.Int) ScriptBuilder {
	return s.CallInterop("Runtime.SendTokens", destinationChain, from, to, symbol, amount)
}

// CrossTransferNFT emits a Runtime.SendToken interop call for a non-fungible token.
func (s ScriptBuilder) CrossTransferNFT(destinationChain cryptography.Address, symbol string, from, to cryptography.Address, tokenID *big.Int) ScriptBuilder {
	return s.CallInterop("Runtime.SendToken", destinationChain, from, to, symbol, tokenID)
}

// CrossTransferNFTText parses address text and emits a Runtime.SendToken interop call.
func (s ScriptBuilder) CrossTransferNFTText(destinationChain string, symbol string, from string, to string, tokenID *big.Int) ScriptBuilder {
	destinationAddress, err := addressFromText(destinationChain)
	if err != nil {
		return s.withError(err)
	}
	fromAddress, err := addressFromText(from)
	if err != nil {
		return s.withError(err)
	}
	return s.CrossTransferNFTToText(destinationAddress, symbol, fromAddress, to, tokenID)
}

// CrossTransferNFTToText emits a Runtime.SendToken interop call with destination account as VM string.
func (s ScriptBuilder) CrossTransferNFTToText(destinationChain cryptography.Address, symbol string, from cryptography.Address, to string, tokenID *big.Int) ScriptBuilder {
	return s.CallInterop("Runtime.SendToken", destinationChain, from, to, symbol, tokenID)
}

// CallNFT emits a method call against the token series contract.
func (s ScriptBuilder) CallNFT(symbol string, seriesID *big.Int, method string, args ...interface{}) ScriptBuilder {
	return s.CallContract(symbol+"#"+seriesID.String(), method, args...)
}
