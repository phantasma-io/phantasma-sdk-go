package main

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
)

var consoleReader = bufio.NewReader(os.Stdin)

func readConsoleLine(reader *bufio.Reader) (string, bool) {
	line, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF && line != "" {
			return strings.TrimRight(line, "\r\n"), true
		}
		return "", false
	}

	return strings.TrimRight(line, "\r\n"), true
}

func exitOnClosedInput() {
	fmt.Println()
	os.Exit(0)
}

func PromptIndexedMenu(title string, items []string) (int, string) {
	if title != "" {
		fmt.Println(title)
	}

	for menuIndex, each := range items {
		if len(items) > 9 {
			fmt.Printf("%02d - %s\n", menuIndex+1, each)
		} else {
			fmt.Printf("%d - %s\n", menuIndex+1, each)
		}
	}

	menuIndex := 0
	for {
		fmt.Print("Enter menu index: ")
		menuIndexStr, ok := readConsoleLine(consoleReader)
		if !ok {
			exitOnClosedInput()
		}
		menuIndex, _ = strconv.Atoi(menuIndexStr)

		if menuIndex >= 1 && menuIndex <= len(items) {
			return menuIndex, items[menuIndex-1]
		}
		fmt.Printf("Please enter menu index in the range [%d-%d]\n", 1, len(items))
	}
}

func PromptYNChoice(message string) bool {
	for {
		fmt.Print(message, " Please enter 'y' or 'n': ")
		choiceYN, ok := readConsoleLine(consoleReader)
		if !ok {
			exitOnClosedInput()
		}
		if strings.ToLower(choiceYN) == "n" {
			return false
		}
		if strings.ToLower(choiceYN) == "y" {
			return true
		}
		fmt.Println("Please enter 'y' or 'n'")
	}
}

func PromptIntInput(message string, minValue int, maxValue int) int {
	for {
		fmt.Printf("%s [%d-%d]: ", message, minValue, maxValue)
		inputStr, ok := readConsoleLine(consoleReader)
		if !ok {
			exitOnClosedInput()
		}
		input, _ := strconv.Atoi(inputStr)

		if input < minValue || input > maxValue {
			fmt.Printf("Entered value '%d' is out of range [%d-%d]\n", input, minValue, maxValue)
		} else {
			return input
		}
	}
}

func PromptBigFloatInput(message string, minValue *big.Float, maxValue *big.Float) (*big.Float, string) {
	for {
		fmt.Printf("%s [%s-%s]: ", message, minValue.String(), maxValue.String())
		inputStr, ok := readConsoleLine(consoleReader)
		if !ok {
			exitOnClosedInput()
		}
		input, ok := big.NewFloat(0).SetString(inputStr)
		if !ok {
			fmt.Printf("Entered value '%s' is not a valid number\n", inputStr)
			continue
		}

		if input.Cmp(minValue) == -1 || input.Cmp(maxValue) == 1 {
			fmt.Printf("Entered value '%s' is out of range [%s-%s]\n", input.String(), minValue.String(), maxValue.String())
		} else {
			return input, inputStr
		}
	}
}

func PromptStringInput(message string) string {
	fmt.Print(message)
	input, ok := readConsoleLine(consoleReader)
	if !ok {
		exitOnClosedInput()
	}

	return input
}
