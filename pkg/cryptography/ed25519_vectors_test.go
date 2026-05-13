package cryptography

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const ed25519FixtureSHA256 = "dd747f5c49b49a67f1c63d02351be669558bf9da65571ed7311bcd8cf8d2bd01"

type ed25519Vector struct {
	caseID       string
	seedHex      string
	publicKeyHex string
	messageHex   string
	signatureHex string
}

func TestEd25519GoldenVectors(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "ed25519_vectors.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256.Sum256(data); hex.EncodeToString(got[:]) != ed25519FixtureSHA256 {
		t.Fatalf("ed25519_vectors.tsv hash mismatch: got %s", hex.EncodeToString(got[:]))
	}

	for _, vector := range readEd25519Vectors(t, data) {
		seed := mustHex(t, vector.seedHex)
		message := mustHex(t, vector.messageHex)
		publicKey := mustHex(t, vector.publicKeyHex)
		expectedSignature := mustHex(t, vector.signatureHex)

		keys := NewPhantasmaKeys(seed)
		if got := hex.EncodeToString(keys.PublicKey()); got != vector.publicKeyHex {
			t.Fatalf("%s public key mismatch: got %s", vector.caseID, got)
		}
		if !NewEd25519Signature(expectedSignature).Verify(message, []Address{keys.Address()}) {
			t.Fatalf("%s expected signature does not verify through SDK address", vector.caseID)
		}
		signature := keys.Sign(message).Bytes()
		if len(signature) != 65 || signature[0] != 64 {
			t.Fatalf("%s serialized signature has unexpected shape", vector.caseID)
		}
		if got := hex.EncodeToString(signature[1:]); got != vector.signatureHex {
			t.Fatalf("%s signature mismatch: got %s", vector.caseID, got)
		}
		if hex.EncodeToString(publicKey) != vector.publicKeyHex {
			t.Fatalf("%s public key fixture failed hex round trip", vector.caseID)
		}
	}
}

func readEd25519Vectors(t *testing.T, data []byte) []ed25519Vector {
	t.Helper()
	var result []ed25519Vector
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "case_id\t") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 7 {
			t.Fatalf("malformed Ed25519 fixture row: %q", line)
		}
		result = append(result, ed25519Vector{
			caseID:       parts[0],
			seedHex:      parts[2],
			publicKeyHex: parts[3],
			messageHex:   parts[4],
			signatureHex: parts[5],
		})
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatal("ed25519_vectors.tsv has no data rows")
	}
	return result
}

func mustHex(t *testing.T, text string) []byte {
	t.Helper()
	bytes, err := hex.DecodeString(text)
	if err != nil {
		t.Fatalf("invalid hex %q: %v", text, err)
	}
	return bytes
}
