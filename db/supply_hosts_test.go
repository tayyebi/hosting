package db

import "testing"

func TestParseManualIPList(t *testing.T) {
	raw := `
		# Primary customer pool on fra-nested-01
		198.51.100.10
		198.51.100.11, 198.51.100.12/28
		198.51.100.10
	`
	entries, err := ParseManualIPList(raw, 24, "198.51.100.1", "52:54:00:11:22:33", 50, 200, "FRA pool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 deduplicated IPs, got %d", len(entries))
	}
	if entries[0].IPAddress != "198.51.100.10" || entries[0].CIDRPrefix != 24 || entries[0].Gateway != "198.51.100.1" {
		t.Fatalf("unexpected entry[0]: %+v", entries[0])
	}
	if entries[2].IPAddress != "198.51.100.12" || entries[2].CIDRPrefix != 28 {
		t.Fatalf("expected explicit /28 override on entry[2], got %+v", entries[2])
	}

	if _, err := ParseManualIPList("not-an-ip", 24, "", "", 0, 200, ""); err == nil {
		t.Fatalf("expected error on invalid IP address")
	}
}
