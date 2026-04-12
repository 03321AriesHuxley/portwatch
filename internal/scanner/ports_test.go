package scanner

import (
	"os"
	"testing"
)

func TestParseProcNet_MissingFile(t *testing.T) {
	entries, err := parseProcNet("/nonexistent/path", "tcp")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if entries != nil {
		t.Fatalf("expected nil entries, got %v", entries)
	}
}

func TestParseProcNet_ValidData(t *testing.T) {
	// Write a minimal /proc/net/tcp-style fixture.
	content := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 12345 1 0000000000000000 100 0 0 10 0
`
	tmp, err := os.CreateTemp("", "procnet-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	entries, err := parseProcNet(tmp.Name(), "tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Port != 0x1F90 {
		t.Errorf("expected port 0x1F90 (8080), got %d", entries[0].Port)
	}
	if entries[0].Protocol != "tcp" {
		t.Errorf("expected protocol tcp, got %s", entries[0].Protocol)
	}
}

func TestSnapshotReturnsNoError(t *testing.T) {
	// Snapshot reads /proc — on Linux this should succeed; elsewhere it returns empty.
	_, err := Snapshot()
	if err != nil {
		t.Fatalf("Snapshot returned unexpected error: %v", err)
	}
}
