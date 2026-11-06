package handlers

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
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

func TestTenantResourcesTemplateRendersSupplyHostsIPPoolsAndZeroDummyData(t *testing.T) {
	resTmpl, err := template.ParseFiles(filepath.Join("..", "templates", "tenant", "resources_content.html"))
	if err != nil {
		t.Fatalf("parse tenant resources_content.html: %v", err)
	}

	// 1. Empty state: zero dummy rows, full Resource Tree taxonomy + IP Address Pools present
	var emptyBuf bytes.Buffer
	emptyData := resourcesInnerData{
		TenantSlug:  "cloud-one",
		ActiveView:  "all",
		GroupCounts: map[string]int{},
		CatCounts:   map[string]int{},
	}
	if err := resTmpl.ExecuteTemplate(&emptyBuf, "tenant-page-content", emptyData); err != nil {
		t.Fatalf("render empty tenant-page-content: %v", err)
	}
	emptyHTML := emptyBuf.String()
	for _, forbidden := range []string{"DummyProvider", "summyprovider", "acme.io", "prx-cluster-prod", "2,048 cores"} {
		if strings.Contains(emptyHTML, forbidden) {
			t.Fatalf("expected zero hardcoded dummy strings in resources_content.html, found %q", forbidden)
		}
	}
	for _, required := range []string{"Linux Nested VPS (QEMU/KVM)", "IP Address Pools", "On-board New Resource"} {
		if !strings.Contains(emptyHTML, required) {
			t.Fatalf("expected %q in tenant resources_content.html", required)
		}
	}

	// 2. Populated state: supply host + slide-in inspector + manual assignable IP pool
	host := TenantHostView{
		SupplyHostRecord: db.SupplyHostRecord{
			ID:              1,
			TenantID:        10,
			Name:            "fra-nested-vps-01",
			HostnameOrIP:    "203.0.113.10",
			SSHPort:         22,
			SSHUser:         "root",
			DriverKind:      "linux_nested_qemu",
			OwnershipType:   "rented",
			RegionSlug:      "eu-central-1",
			BridgeInterface: "br0",
			StoragePoolPath: "/var/lib/libvirt/images",
			CPUCores:        8,
			RAMMB:           16384,
			DiskGB:          160,
			KVMAvailable:    true,
			VirtStackStatus: "ready",
			InventoryStatus: "online",
			AvailableIPs:    2,
			TotalIPs:        2,
		},
		CategoryKey:        "linux-vps",
		BadgeLabel:         "Linux VPS",
		BadgeClass:         "res-type-badge--linux",
		MonthlyCostDisplay: "$24.00",
		RAMGBDisplay:       "16 GB",
		IPs: []TenantIPView{
			{
				SupplyHostIP:  db.SupplyHostIP{ID: 10, TenantID: 10, HostID: 1, HostName: "fra-nested-vps-01", IPAddress: "198.51.100.10", CIDRPrefix: 24, Gateway: "198.51.100.1", Status: "available"},
				CostDisplay:   "$0.50",
				RetailDisplay: "$2.00",
			},
		},
	}

	var buf bytes.Buffer
	err = resTmpl.ExecuteTemplate(&buf, "tenant-page-content", resourcesInnerData{
		TenantSlug:   "cloud-one",
		ActiveView:   "linux-vps",
		TotalCount:   1,
		OnlineCount:  1,
		GroupCounts:  map[string]int{"linux-vps": 1, "ip-pool": 1},
		CatCounts:    map[string]int{"servers": 1, "networking": 1},
		Hosts:        []TenantHostView{host},
		IPs:          host.IPs,
		SelectedHost: &host,
		AvailableIPs: 2,
		TotalIPs:     2,
		Rows: []TenantInventoryRow{
			{
				RowKind:        "supply_host",
				NumericID:      1,
				ID:             "host-1",
				CategoryKey:    "linux-vps",
				Name:           "fra-nested-vps-01",
				SubLabel:       "Linux VPS · eu-central-1",
				BadgeLabel:     "Linux VPS",
				BadgeClass:     "res-type-badge--linux",
				Status:         "online",
				StatusLabel:    "Online",
				StatusDotClass: "res-dot--ok",
				HostOrIP:       "203.0.113.10:22",
				Region:         "eu-central-1",
				CPUDisplay:     "8 vCPU",
				RAMDisplay:     "16 GB",
				Tags:           []string{"rented", "KVM"},
				DetailJSON:     "{}",
			},
		},
	})
	if err != nil {
		t.Fatalf("render populated tenant-page-content: %v", err)
	}
	rendered := buf.String()
	for _, needle := range []string{"fra-nested-vps-01", "198.51.100.10/24", "Auto-Provision QEMU/KVM", "Assignable IP Pool"} {
		if !strings.Contains(rendered, needle) {
			t.Fatalf("expected rendered tenant resources page to contain %q", needle)
		}
	}
}

func TestSysadminInfrastructureRemovedAndAccountingRendersCleanly(t *testing.T) {
	for _, item := range sysadminNav {
		if item.ID == "infrastructure" || strings.Contains(item.Path, "/s/infrastructure") {
			t.Fatalf("expected /s/infrastructure to be removed from sysadminNav, found %+v", item)
		}
	}

	acctTmpl, err := template.ParseFiles(filepath.Join("..", "templates", "sysadmin", "accounting_content.html"))
	if err != nil {
		t.Fatalf("parse accounting_content.html: %v", err)
	}
	var buf bytes.Buffer
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
		ActiveNav: "tenants",
		PageTitle: "Tenants",
		Content:   template.HTML("<p>ok</p>"),
	}); err != nil {
		t.Fatalf("render sysadmin-base: %v", err)
	}
	if strings.Contains(buf.String(), "/s/infrastructure") {
		t.Fatalf("expected sysadmin shell not to reference /s/infrastructure")
	}
}

func TestTenantResourcesAPICrossTenantIsolation(t *testing.T) {
	alpha := store.add(Tenant{ID: "t-alpha", Name: "Alpha Cloud", Slug: "alpha-test"})
	beta := store.add(Tenant{ID: "t-beta", Name: "Beta Cloud", Slug: "beta-test"})
	defer func() {
		store.delete(alpha.ID)
		store.delete(beta.ID)
	}()

	mux := http.NewServeMux()
	RegisterAPIRoutes(mux)

	// Unauthenticated request -> 401
	unauthReq := httptest.NewRequest(http.MethodGet, "/api/tenant/alpha-test/resources", nil)
	unauthRec := httptest.NewRecorder()
	mux.ServeHTTP(unauthRec, unauthReq)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated request, got %d", unauthRec.Code)
	}

	// Mint session cookie for Alpha tenant admin
	seedRec := httptest.NewRecorder()
	seedReq := httptest.NewRequest(http.MethodGet, "/", nil)
	createSession(seedRec, seedReq, User{
		ID:         101,
		Email:      "admin@alpha.io",
		Name:       "Alpha Admin",
		Role:       RoleTenant,
		TenantSlug: "alpha-test",
	}, false)
	cookies := seedRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie to be issued")
	}

	// Own tenant resources -> 200 OK
	ownReq := httptest.NewRequest(http.MethodGet, "/api/tenant/alpha-test/resources", nil)
	ownReq.AddCookie(cookies[0])
	ownRec := httptest.NewRecorder()
	mux.ServeHTTP(ownRec, ownReq)
	if ownRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for own tenant resources, got %d (%s)", ownRec.Code, ownRec.Body.String())
	}

	// Cross-tenant resources & supply hosts -> 403 Forbidden
	for _, crossPath := range []string{
		"/api/tenant/beta-test/resources",
		"/api/tenant/beta-test/resources/hosts",
		"/api/tenant/beta-test/resources/ips",
	} {
		req := httptest.NewRequest(http.MethodGet, crossPath, nil)
		req.AddCookie(cookies[0])
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for cross-tenant %s, got %d", crossPath, rec.Code)
		}
	}
}
