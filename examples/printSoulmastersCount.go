package main

import (
	"encoding/hex"
	"fmt"

	scriptbuilder "github.com/phantasma-io/phantasma-sdk-go/pkg/vm/script_builder"
)

func printSoulmastersCount() {
	// Build script
	sb := scriptbuilder.BeginScript().
		CallContract("stake", "GetMasterCount")
	script := sb.EndScript()

	// Before sending script to the chain we need to encode it into Base16 encoding (HEX)
	encodedScript := hex.EncodeToString(script)
	fmt.Println("Script: " + encodedScript)

	if !PromptYNChoice("Invoke script?") {
		return
	}

	// Make the call itself
	result, err := client.InvokeRawScript(rpcContext, "main", encodedScript)

	if err != nil {
		fmt.Println("Script invocation failed:", err)
		return
	}

	value, err := result.DecodeResultWithError()
	if err != nil {
		fmt.Println("Script result decoding failed:", err)
		return
	}

	fmt.Println("Current SoulMasters count: ", value.AsNumber().String())
}

func printSoulmastersCountAndLastInflationDate() {
	// Build script
	sb := scriptbuilder.BeginScript().
		CallContract("stake", "GetMasterCount")

	sb.CallContract("gas", "GetLastInflationDate")

	script := sb.EndScript()

	// Before sending script to the chain we need to encode it into Base16 encoding (HEX)
	encodedScript := hex.EncodeToString(script)
	fmt.Println("Script: " + encodedScript)

	if !PromptYNChoice("Invoke script?") {
		return
	}

	// Make the call itself
	result, err := client.InvokeRawScript(rpcContext, "main", encodedScript)

	if err != nil {
		fmt.Println("Script invocation failed:", err)
		return
	}

	mastersCount, err := result.DecodeResultsWithError(0)
	if err != nil {
		fmt.Println("SoulMasters count decoding failed:", err)
		return
	}

	lastInflationDate, err := result.DecodeResultsWithError(1)
	if err != nil {
		fmt.Println("Last inflation date decoding failed:", err)
		return
	}

	fmt.Printf("Current SoulMasters count: %s, last inflation date: %s \n", mastersCount.AsString(), lastInflationDate.AsString())
}
