package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"go-vpn-backend/supabase"
	"go-vpn-backend/wg"
)

// CreatePeerRequest is the expected JSON payload for creating a new peer.
type CreatePeerRequest struct {
	UserID string `json:"user_id"`
}

// CreatePeerResponse is the JSON payload returned upon successful creation.
type CreatePeerResponse struct {
	UserID     string `json:"user_id"`
	IPAddress  string `json:"ip_address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

var (
	configFile    = "/etc/wireguard/wg0.conf"
	interfaceName = "wg0"
)

func main() {
	// Initialize Supabase Client
	supaClient, err := supabase.NewClient()
	if err != nil {
		log.Printf("Warning: Failed to initialize Supabase client: %v. Supabase integration will fail.", err)
		// Proceed anyway, we might just be testing key gen/config
	}

	// Override config file path for testing if provided via env var
	if envConf := os.Getenv("WG_CONFIG_FILE"); envConf != "" {
		configFile = envConf
	}

	http.HandleFunc("/api/peers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// The Citadel Security Lock
		apiSecret := os.Getenv("CITADEL_SECRET")
		if r.Header.Get("Authorization") != "Bearer "+apiSecret {
			http.Error(w, "Unauthorized: Citadel doors are locked.", http.StatusUnauthorized)
			return
		}

		var req CreatePeerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request: invalid JSON payload", http.StatusBadRequest)
			return
		}

		if req.UserID == "" {
			http.Error(w, "Bad request: user_id is required", http.StatusBadRequest)
			return
		}

		// 1. Generate WireGuard Keys
		privKey, pubKey, err := GenerateKeys()
		if err != nil {
		    log.Fatalf("Critical Error: Failed to generate WireGuard keys: %v", err)
		}

		log.Printf("Server successfully generated keys. Public Key: %s", pubKey)

		// 2. Add Peer to Config and Allocate IP
		assignedIP, err := wg.AddPeerToConfig(configFile, keys.PublicKey)
		if err != nil {
			log.Printf("Failed to add peer to config: %v", err)
			http.Error(w, "Internal server error: failed to update configuration", http.StatusInternalServerError)
			return
		}

		// 3. Reload WireGuard Interface
		if err := wg.ReloadInterface(interfaceName); err != nil {
			log.Printf("Failed to reload interface: %v", err)
			http.Error(w, "Internal server error: failed to reload network interface", http.StatusInternalServerError)
			return
		}

		// 4. Update Supabase
		if supaClient != nil {
			if err := supaClient.InsertPeer(req.UserID, assignedIP, pubKey); err != nil {
				log.Printf("Failed to insert peer into Supabase: %v", err)
				http.Error(w, "Internal server error: failed to update database", http.StatusInternalServerError)
				return
			}
		}

		// Success Response
		resp := CreatePeerResponse{
			UserID:     req.UserID,
			IPAddress:  assignedIP,
			PublicKey:  kpubKey,
			PrivateKey: privKey, // Give private key ONLY to the client once
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s...", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
