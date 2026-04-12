package scanner

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PortEntry represents a single bound port on the host.
type PortEntry struct {
	Protocol string
	LocalAddr string
	Port     uint16
	PID      int
}

// Snapshot returns the current set of bound ports by reading /proc/net/tcp
// and /proc/net/tcp6. On non-Linux systems it returns an empty slice.
func Snapshot() ([]PortEntry, error) {
	var entries []PortEntry
	for _, proto := range []string{"tcp", "tcp6", "udp", "udp6"} {
		path := fmt.Sprintf("/proc/net/%s", proto)
		more, err := parseProcNet(path, proto)
		if err != nil {
			// file may not exist on some kernels — skip gracefully
			continue
		}
		entries = append(entries, more...)
	}
	return entries, nil
}

func parseProcNet(path, proto string) ([]PortEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []PortEntry
	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip header line
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}
		localHex := fields[1] // e.g. "0100007F:0050"
		parts := strings.SplitN(localHex, ":", 2)
		if len(parts) != 2 {
			continue
		}
		portVal, err := strconv.ParseUint(parts[1], 16, 16)
		if err != nil {
			continue
		}
		entries = append(entries, PortEntry{
			Protocol: proto,
			LocalAddr: parts[0],
			Port:     uint16(portVal),
		})
	}
	return entries, scanner.Err()
}
