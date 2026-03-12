package wg

import (
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	keys, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if keys == nil {
		t.Fatalf("Expected keys to be non-nil")
	}

	if len(keys.PrivateKey) == 0 {
		t.Errorf("Expected private key to be generated")
	}

	if len(keys.PublicKey) == 0 {
		t.Errorf("Expected public key to be generated")
	}
}
