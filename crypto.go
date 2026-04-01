package main

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// GenerateKeys makes pub/priv keypair
func GenerateKeys() (string, string, error) {
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}
	
	// makes pubKey from privKey
	pubKey := privKey.PublicKey()
	
	return privKey.String(), pubKey.String(), nil
}
