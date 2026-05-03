package main

import (
	"encoding/hex"
	"fmt"

	scriptbuilder "github.com/phantasma-io/phantasma-go/pkg/vm/script_builder"
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
	result, err := client.InvokeRawScript("main", encodedScript)

	if err != nil {
		panic("Script invocation failed! Error: " + err.Error())
	}

	value, err := result.DecodeResultWithError()
	if err != nil {
		panic("Script result decoding failed! Error: " + err.Error())
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
	result, err := client.InvokeRawScript("main", encodedScript)

	if err != nil {
		panic("Script invocation failed! Error: " + err.Error())
	}

	mastersCount, err := result.DecodeResultsWithError(0)
	if err != nil {
		panic("SoulMasters count decoding failed! Error: " + err.Error())
	}

	lastInflationDate, err := result.DecodeResultsWithError(1)
	if err != nil {
		panic("Last inflation date decoding failed! Error: " + err.Error())
	}

	fmt.Printf("Current SoulMasters count: %s, last inflation date: %s \n", mastersCount.AsString(), lastInflationDate.AsString())
}
