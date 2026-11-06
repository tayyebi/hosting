package db

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNormalizeDomain(t *testing.T) {
	valid := map[string]string{
		"system.example.com":          "system.example.com",
		"  HTTPS://Admin.Example.COM/": "admin.example.com",
		"http://cloud.acme.io:8443/":  "cloud.acme.io:8443",
		"sub.domain.co.uk":            "sub.domain.co.uk",
		"localhost":                   "localhost",
		"localhost:8021":              "localhost:8021",
		"203.0.113.10:8021":           "203.0.113.10:8021",
		"[::1]:8021":                  "[::1]:8021",
	}
	for input, want := range valid {
		got, err := NormalizeDomain(input)
		if err != nil {
			t.Fatalf("NormalizeDomain(%q) unexpected error: %v", input, err)
		}
		if got != want {
			t.Fatalf("NormalizeDomain(%q) = %q, want %q", input, got, want)
		}
	}

	invalid := []string{
		"",
		"   ",
		"https://",
		"example.com/path",
		"example.com?x=1",
		"user@example.com",
		"-bad.example.com",
		"bad-.example.com",
		"singlelabel",
		"example.com:99999",
		"example.com:0",
	}
	for _, input := range invalid {
		if got, err := NormalizeDomain(input); err == nil {
			t.Fatalf("NormalizeDomain(%q) succeeded with %q, expected error", input, got)
		}
	}
}

func TestHostMatchesConfiguredDomain(t *testing.T) {
	full, bare := SplitHostAndPort("Admin.Example.com:8021")
	if full != "admin.example.com:8021" || bare != "admin.example.com" {
		t.Fatalf("unexpected SplitHostAndPort result: full=%q bare=%q", full, bare)
	}

	// Portless configured domain matches regardless of request port.
	if !HostMatchesConfiguredDomain(full, bare, "admin.example.com") {
		t.Fatal("expected portless domain to match host with port")
	}

	// Port-qualified configured domain requires exact port match.
	if !HostMatchesConfiguredDomain(full, bare, "admin.example.com:8021") {
		t.Fatal("expected matching port-qualified domain to match")
	}
	if HostMatchesConfiguredDomain(full, bare, "admin.example.com:9000") {
		t.Fatal("expected different port-qualified domain not to match")
	}
}

func TestVerifyPasswordSupportsPBKDF2AndBcrypt(t *testing.T) {
	pbkdfHash, err := HashPassword("StrongPass123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !VerifyPassword(pbkdfHash, "StrongPass123") {
		t.Fatal("expected PBKDF2 hash to verify")
	}
	if VerifyPassword(pbkdfHash, "WrongPass123") {
		t.Fatal("expected wrong password against PBKDF2 hash to fail")
	}

	bcHash, err := bcrypt.GenerateFromPassword([]byte("TenantRegPass9"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}
	if !VerifyPassword(string(bcHash), "TenantRegPass9") {
		t.Fatal("expected bcrypt hash to verify")
	}
	if VerifyPassword(string(bcHash), "WrongPass9") {
		t.Fatal("expected wrong password against bcrypt hash to fail")
	}
}
