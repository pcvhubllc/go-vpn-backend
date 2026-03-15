package wg

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

var (
	// Default base IP if no peers exist.  This assumes 10.8.0.0/24 subnet and server is .1
	DefaultBaseIP = "10.8.0.2"
	subnetPrefix  = "10.8.0."

	// regex to extract IP from AllowedIPs = 10.8.0.x/32
	allowedIPsRegex = regexp.MustCompile(`AllowedIPs\s*=\s*([0-9\.]+)\/`)
)

// AddPeerToConfig safely appends a new peer to the WireGuard configuration file
// and returns the newly assigned IP address.
func AddPeerToConfig(configPath, publicKey string) (string, error) {
	// 1. Read existing config
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If config doesn't exist, we can't append to it securely.
			return "", fmt.Errorf("config file %s does not exist", configPath)
		}
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	// 2. Parse existing IPs
	usedIPs := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "AllowedIPs") {
			matches := allowedIPsRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				ipStr := matches[1]
				usedIPs[ipStr] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error parsing config file: %w", err)
	}

	// 3. Determine next available IP
	nextIP, err := getNextAvailableIP(usedIPs)
	if err != nil {
		return "", fmt.Errorf("failed to allocate IP: %w", err)
	}

	// 4. Create new peer block
	newPeerBlock := fmt.Sprintf("\n[Peer]\nPublicKey = %s\nAllowedIPs = %s/32\n", publicKey, nextIP)

	// 5. Append to file
	// Open file for appending, create if not exist (though we checked earlier), write-only.
	// We use 0600 for permissions to keep it secure.
	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to open config for appending: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(newPeerBlock); err != nil {
		return "", fmt.Errorf("failed to write new peer block: %w", err)
	}

	return nextIP, nil
}

// getNextAvailableIP finds the next available IP address in the 10.8.0.x subnet.
func getNextAvailableIP(usedIPs map[string]bool) (string, error) {
	// Start from 10.8.0.2 up to 10.8.0.254 (assuming .1 is server and .255 is broadcast)
	for i := 2; i <= 254; i++ {
		ipCandidate := fmt.Sprintf("%s%d", subnetPrefix, i)
		if !usedIPs[ipCandidate] {
			// Basic validation
			if net.ParseIP(ipCandidate) == nil {
				continue
			}
			return ipCandidate, nil
		}
	}
	return "", fmt.Errorf("no available IPs in subnet %s0/24", subnetPrefix)
}
