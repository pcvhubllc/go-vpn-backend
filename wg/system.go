package wg

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ReloadInterface runs the system command to seamlessly reload the WireGuard interface
// without dropping existing client connections.
func ReloadInterface(interfaceName string) error {
	// The command: wg syncconf wg0 <(wg-quick strip wg0)
	// Requires bash process substitution
	cmdStr := fmt.Sprintf("wg syncconf %s <(wg-quick strip %s)", interfaceName, interfaceName)

	cmd := exec.Command("bash", "-c", cmdStr)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload interface: %w, stderr: %s", err, stderr.String())
	}

	return nil
}
