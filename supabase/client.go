package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// PeerRecord represents the data we insert into Supabase
type PeerRecord struct {
	UserID    string `json:"user_id"`
	IPAddress string `json:"ip_address"`
	PublicKey string `json:"public_key"`
}

// Client handles communication with the Supabase REST API
type Client struct {
	URL        string
	AnonKey    string
	TableName  string
	HTTPClient *http.Client
}

// NewClient creates a new Supabase client reading configuration from environment variables.
func NewClient() (*Client, error) {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")
	tableName := os.Getenv("SUPABASE_TABLE_NAME") // e.g., "vpn_peers"

	if url == "" || key == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_KEY environment variables must be set")
	}

	if tableName == "" {
		tableName = "vpn_peers" // Default table name
	}

	return &Client{
		URL:        url,
		AnonKey:    key,
		TableName:  tableName,
		HTTPClient: &http.Client{},
	}, nil
}

// InsertPeer performs a POST request to Supabase to insert a new peer record.
func (c *Client) InsertPeer(userID, ipAddress, publicKey string) error {
	record := PeerRecord{
		UserID:    userID,
		IPAddress: ipAddress,
		PublicKey: publicKey,
	}

	jsonData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal peer record: %w", err)
	}

	endpoint := fmt.Sprintf("%s/rest/v1/%s", c.URL, c.TableName)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Supabase specific headers
	req.Header.Set("apikey", c.AnonKey)
	req.Header.Set("Authorization", "Bearer "+c.AnonKey)
	req.Header.Set("Content-Type", "application/json")
	// Prefer: return=representation to return the inserted row (optional, but good for confirmation)
	// req.Header.Set("Prefer", "return=representation")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("supabase API returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
