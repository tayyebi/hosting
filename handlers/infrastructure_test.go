package handlers

import (
	"bytes"
	"context"
	"html/template"
	"path/filepath"
	"strings"
	"testing"

	"platform/db"
)

func TestOnlySysadminBootstrapCredentialsAcceptedWithoutDB(t *testing.T) {
	u, ok := authenticateCredentials(context.Background(), "tayyebimohammadreza@gmail.com", "sysadmin123")
	if !ok || u.Role != RoleSysAdmin {
		t.Fatalf("expected bootstrap sysadmin login to succeed, got ok=%v role=%s", ok, u.Role)
	}
	if _, ok := authenticateCredentials(context.Background(), "tenant@platform.io", "tenant123"); ok {
		t.Fatalf("expected removed demo tenant account to be rejected")
	}
	if _, ok := authenticateCredentials(context.Background(), "user@platform.io", "user123"); ok {
		t.Fatalf("expected removed demo user account to be rejected")
	}
	if got := len(store.list()); got != 0 {
		t.Fatalf("expected 0 hardcoded dummy tenants, got %d", got)
	}
}

func TestSysadminInfrastructureAndAccountingTemplatesRenderCleanly(t *testing.T) {
	infraTmpl, err := template.ParseFiles(filepath.Join("..", "templates", "sysadmin", "infrastructure_content.html"))
	if err != nil {
		t.Fatalf("parse infrastructure_content.html: %v", err)
	}
	host := infraHostView{
		SupplyHostRecord: db.SupplyHostRecord{
			ID:              1,
			Name:            "fra-nested-vps-01",
			HostnameOrIP:    "203.0.113.10",
			SSHPort:         22,
			SSHUser:         "root",
			DriverKind:      "linux_nested_qemu",
			RegionSlug:      "eu-central-1",
			CPUCores:        8,
			RAMMB:           16384,
			DiskGB:          160,
			KVMAvailable:    true,
			VirtStackStatus: "ready",
			InventoryStatus: "online",
			AvailableIPs:    2,
			TotalIPs:        2,
		},
		MonthlyCostDisplay: "$24.00",
		RAMGBDisplay:       "16.0 GB",
		IPs: []infraIPView{
			{
				SupplyHostIP:  db.SupplyHostIP{ID: 10, HostID: 1, IPAddress: "198.51.100.10", CIDRPrefix: 24, Gateway: "198.51.100.1", Status: "available"},
				CostDisplay:   "$0.50",
				RetailDisplay: "$2.00",
			},
		},
	}
	var buf bytes.Buffer
	err = infraTmpl.ExecuteTemplate(&buf, "sysadmin-infrastructure-content", infraPageData{
		Hosts:             []infraHostView{host},
		SelectedHost:      &host,
		TotalHosts:        1,
		ReadyOnlineHosts:  1,
		TotalVCPUs:        8,
		TotalRAMGB:        16,
		TotalDiskGB:       160,
		AvailableIPs:      2,
		TotalIPs:          2,
		MonthlySupplyCost: "$24.00",
	})
	if err != nil {
		t.Fatalf("render sysadmin-infrastructure-content: %v", err)
	}
	if !strings.Contains(buf.String(), "198.51.100.10") || !strings.Contains(buf.String(), "fra-nested-vps-01") {
		t.Fatalf("expected rendered infrastructure page to contain host and manual IP pool")
	}

	acctTmpl, err := template.ParseFiles(filepath.Join("..", "templates", "sysadmin", "accounting_content.html"))
	if err != nil {
		t.Fatalf("parse accounting_content.html: %v", err)
	}
	buf.Reset()
	if err := acctTmpl.ExecuteTemplate(&buf, "sysadmin-accounting-content", accountingPageData{
		MRRDisplay:         "$0.00",
		SupplyCostDisplay:  "$0.00",
		GrossMarginDisplay: "$0.00",
	}); err != nil {
		t.Fatalf("render sysadmin-accounting-content: %v", err)
	}

	shellTmpl, err := template.ParseFiles(
		filepath.Join("..", "templates", "sysadmin", "base.html"),
		filepath.Join("..", "templates", "sysadmin", "shell.html"),
	)
	if err != nil {
		t.Fatalf("parse sysadmin shell: %v", err)
	}
	buf.Reset()
	if err := shellTmpl.ExecuteTemplate(&buf, "sysadmin-base", SysadminData{
		User:      User{Name: "System Administrator", Email: "tayyebimohammadreza@gmail.com", Role: RoleSysAdmin},
		Nav:       sysadminNav,
		ActiveNav: "infrastructure",
		PageTitle: "Infrastructure & IP Pools",
		Content:   template.HTML("<p>ok</p>"),
	}); err != nil {
		t.Fatalf("render sysadmin-base: %v", err)
	}
}
