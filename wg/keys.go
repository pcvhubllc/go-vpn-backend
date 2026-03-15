package wg

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// KeyPair holds a matched pair of WireGuard keys
type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

// GenerateKeyPair generates a new valid WireGuard private and public key pair.
// It uses Curve25519 internally as implemented by the standard zx2c4 wgctrl library.
func GenerateKeyPair() (*KeyPair, error) {
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	pubKey := privKey.PublicKey()

	return &KeyPair{
		PrivateKey: privKey.String(),
		PublicKey:  pubKey.String(),
	}, nil
}
