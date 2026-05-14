package util

import (
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

// Internal utils

func isInteger(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func trimWholePrefix(s string, prefix string) string {
	for strings.HasPrefix(s, prefix) {
		s = s[len(prefix):]
	}
	return s
}

func trimWholeSuffix(s string, suffix string) string {
	for strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func stringIsZeroOrEmptyBigint(number string) bool {
	for _, c := range number {
		if c != '0' {
			return false
		}
	}

	return true
}

func cutOffSignPrefix(number string) (string, bool) {
	isPositive := true
	if number[:1] == "-" {
		isPositive = false
		number = number[1:]
	}

	return number, isPositive
}

func addSignPrefix(number string, isPositive bool) string {
	if isPositive || number == "0" {
		return number
	}

	return "-" + number
}

// ConvertDecimalsEx converts big integer number to decimal number, both serialized as a string.
// Example: ConvertDecimalsEx("90000", 10, ".") call returns "0.000009" string
func ConvertDecimalsEx(number string, decimals int, separator string) string {
	if len(number) == 0 {
		return "0"
	}

	number, isPositive := cutOffSignPrefix(number)

	if stringIsZeroOrEmptyBigint(number) {
		return "0"
	}

	number = trimWholePrefix(number, "0")

	if decimals == 0 {
		return addSignPrefix(number, isPositive)
	}

	if len(number) <= decimals {
		return addSignPrefix("0"+separator+strings.Repeat("0", decimals-len(number))+trimWholeSuffix(number, "0"), isPositive)
	}

	integerPart := number[:len(number)-decimals]
	if integerPart == "" {
		integerPart = "0"
	}

	fractionalPart := number[len(number)-decimals:]

	if stringIsZeroOrEmptyBigint(fractionalPart) {
		return addSignPrefix(integerPart, isPositive)
	}
	return addSignPrefix(integerPart+separator+trimWholeSuffix(fractionalPart, "0"), isPositive)
}

// ConvertDecimals converts big integer number to decimal number, serialized as a string.
// Example: ConvertDecimals(*big.NewInt(90000), 10) call returns "0.000009" string
func ConvertDecimals(number *big.Int, decimals int) string {
	return ConvertDecimalsEx(number.String(), decimals, ".")
}

// ConvertDecimalsBackEx converts a decimal string to a raw integer string.
// It returns an error when the input has non-zero fractional precision beyond decimals.
// Example: ConvertDecimalsBackEx("0.000009", 10, ".") returns "90000".
func ConvertDecimalsBackEx(number string, decimals int, separator string) (string, error) {
	if len(number) == 0 {
		return "0", nil
	}

	number, isPositive := cutOffSignPrefix(number)

	if stringIsZeroOrEmptyBigint(number) {
		return "0", nil
	}

	number = trimWholePrefix(number, "0")

	if !strings.Contains(number, separator) {
		// No fractional part found, we need to put zeroes instead
		return addSignPrefix(number+strings.Repeat("0", decimals), isPositive), nil
	}

	split := strings.SplitN(number, separator, 2)

	integerPart := split[0]
	fractionalPart := split[1]

	if decimals == 0 {
		// Nothing to do, only to check if passed number is correct
		if !isInteger(number) && !stringIsZeroOrEmptyBigint(fractionalPart) {
			return "", fmt.Errorf("fractional amount is not allowed when decimals is 0")
		}
		if integerPart == "" {
			integerPart = "0"
		}
		return addSignPrefix(integerPart, isPositive), nil
	}

	if len(fractionalPart) < decimals {
		// We need to add more zeroes to fractional part
		fractionalPart = fractionalPart + strings.Repeat("0", decimals-len(fractionalPart))
	} else if len(fractionalPart) > decimals {
		if stringIsZeroOrEmptyBigint(fractionalPart[decimals:]) {
			// We can safely drop zeroes
			fractionalPart = fractionalPart[:decimals]
		} else {
			return "", fmt.Errorf("fractional part exceeds %d decimals", decimals)
		}
	}

	result := trimWholePrefix(integerPart+fractionalPart, "0")
	if result == "" { // We killed all zeroes, we need one
		result = "0"
	}

	return addSignPrefix(result, isPositive), nil
}

// ConvertDecimalsBack converts decimal number, serialized as a string, to big integer number.
// Example: ConvertDecimalsBack("0.000009", 10) call returns *big.Int with value 90000
func ConvertDecimalsBack(number string, decimals int) (*big.Int, error) {
	raw, err := ConvertDecimalsBackEx(number, decimals, ".")
	if err != nil {
		return nil, err
	}
	n := new(big.Int)
	if _, ok := n.SetString(raw, 10); !ok {
		return nil, fmt.Errorf("invalid integer amount: %q", raw)
	}
	return n, nil
}
