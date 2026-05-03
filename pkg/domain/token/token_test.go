package token

import "testing"

func TestTokenFlagsRoundTrip(t *testing.T) {
	flags := TokenFlags(0).FromSlice([]string{"Transferable", "Fungible", "Burnable"})
	got := flags.ToSlice()
	want := []string{"Transferable", "Fungible", "Burnable"}
	if len(got) != len(want) {
		t.Fatalf("flag count mismatch: %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("flag mismatch at %d: got %s want %s", i, got[i], want[i])
		}
	}
	if none := TokenFlags(0).ToSlice(); len(none) != 1 || none[0] != "None" {
		t.Fatalf("empty flags mismatch: %v", none)
	}
}
