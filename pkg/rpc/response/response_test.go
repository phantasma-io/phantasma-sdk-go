package response

import "testing"

func TestAccountCloneAndTokenBalance(t *testing.T) {
	account := AccountResult{
		Balances: []BalanceResult{{Chain: "main", Amount: "10", Symbol: "SOUL", Decimals: 8, Ids: []string{"1"}}},
	}

	clone := account.Clone()
	clone.Balances[0].Amount = "20"
	clone.Balances[0].Ids[0] = "2"

	if account.Balances[0].Amount != "10" || account.Balances[0].Ids[0] != "1" {
		t.Fatalf("clone must not alias nested slices")
	}

	balance := account.GetTokenBalance(TokenResult{Symbol: "KCAL", Decimals: 10})
	if balance.Symbol != "KCAL" || balance.Amount != "0" || balance.Decimals != 10 {
		t.Fatalf("unexpected synthetic balance: %+v", balance)
	}
}

func TestTokenResultFlagHelpers(t *testing.T) {
	token := TokenResult{Flags: "Fungible, Transferable, Burnable"}
	if !token.IsFungible() || !token.IsTransferable() || !token.IsBurnable() {
		t.Fatalf("expected token flags to be detected")
	}
	if token.IsMintable() || token.IsFuel() {
		t.Fatalf("unexpected token flag detected")
	}
}

func TestScriptResultCheckedDecode(t *testing.T) {
	// Result payloads returned by invokeRawScript are hex-encoded VM objects.
	result := ScriptResult{
		Result:  "0601",
		Results: []string{"0600"},
	}

	value, err := result.DecodeResultWithError()
	if err != nil {
		t.Fatalf("DecodeResultWithError failed: %v", err)
	}
	if value.AsString() != "true" {
		t.Fatalf("expected true VM bool, got %s", value.AsString())
	}

	indexed, err := result.DecodeResultsWithError(0)
	if err != nil {
		t.Fatalf("DecodeResultsWithError failed: %v", err)
	}
	if indexed.AsString() != "false" {
		t.Fatalf("expected false VM bool, got %s", indexed.AsString())
	}
}

func TestScriptResultCheckedDecodeRejectsBadInput(t *testing.T) {
	result := ScriptResult{Result: "not-hex"}
	if _, err := result.DecodeResultWithError(); err == nil {
		t.Fatalf("expected invalid hex to fail")
	}
	if _, err := result.DecodeResultsWithError(0); err == nil {
		t.Fatalf("expected missing indexed result to fail")
	}
}
