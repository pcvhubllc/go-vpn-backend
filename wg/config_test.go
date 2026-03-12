package wg

import (
	"os"
	"strings"
	"testing"
)

func TestAddPeerToConfig_NoExistingPeers(t *testing.T) {
	// Create a temporary config file
	f, err := os.CreateTemp("", "wg0_test_*.conf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	// Write initial server config
	initialConfig := `[Interface]
Address = 10.8.0.1/24
ListenPort = 51820
PrivateKey = server_private_key
`
	if _, err := f.WriteString(initialConfig); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	f.Close()

	// Add peer
	assignedIP, err := AddPeerToConfig(f.Name(), "dummy_public_key")
	if err != nil {
		t.Fatalf("AddPeerToConfig failed: %v", err)
	}

	if assignedIP != "10.8.0.2" {
		t.Errorf("Expected assigned IP to be 10.8.0.2, got %s", assignedIP)
	}

	// Verify file content
	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Failed to read updated config file: %v", err)
	}

	if !strings.Contains(string(content), "PublicKey = dummy_public_key") {
		t.Errorf("Expected config to contain dummy_public_key")
	}
	if !strings.Contains(string(content), "AllowedIPs = 10.8.0.2/32") {
		t.Errorf("Expected config to contain AllowedIPs = 10.8.0.2/32")
	}
}

func TestAddPeerToConfig_ExistingPeers(t *testing.T) {
	// Create a temporary config file
	f, err := os.CreateTemp("", "wg0_test_*.conf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	// Write initial server config with a peer
	initialConfig := `[Interface]
Address = 10.8.0.1/24
ListenPort = 51820
PrivateKey = server_private_key

[Peer]
PublicKey = existing_peer_1
AllowedIPs = 10.8.0.2/32

[Peer]
PublicKey = existing_peer_2
AllowedIPs = 10.8.0.3/32
`
	if _, err := f.WriteString(initialConfig); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	f.Close()

	// Add new peer
	assignedIP, err := AddPeerToConfig(f.Name(), "dummy_public_key_3")
	if err != nil {
		t.Fatalf("AddPeerToConfig failed: %v", err)
	}

	if assignedIP != "10.8.0.4" {
		t.Errorf("Expected assigned IP to be 10.8.0.4, got %s", assignedIP)
	}
}
