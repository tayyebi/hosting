package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"platform/db"
	"platform/providers"
)

// ── View Models ──────────────────────────────────────────────────────────────

type infraHostView struct {
	db.SupplyHostRecord
	MonthlyCostDisplay string
	RAMGBDisplay       string
	IPs                []infraIPView
}

type infraIPView struct {
	db.SupplyHostIP
	CostDisplay   string
	RetailDisplay string
}

type infraPageData struct {
	Hosts               []infraHostView
	SelectedHost        *infraHostView
	TotalHosts          int
	ReadyOnlineHosts    int
	TotalVCPUs          int
	TotalRAMGB          int
	TotalDiskGB         int
	AvailableIPs        int
	AssignedIPs         int
	TotalIPs            int
	MonthlySupplyCost   string
	SupportedDrivers    []string
	FlashSuccess        string
	FlashError          string
}

func buildInfraPageData(ctx context.Context, selectedHostID int64, flashOK, flashErr string) infraPageData {
	d := infraPageData{
		FlashSuccess:      flashOK,
		FlashError:        flashErr,
		MonthlySupplyCost: "$0.00",
		SupportedDrivers: []string{
			string(providers.DriverLinuxNestedQEMU),
			string(providers.DriverQEMULibvirt),
			string(providers.DriverProxmox),
			string(providers.DriverVMwareVSphere),
			string(providers.DriverUpstreamReseller),
		},
	}
	if Pools == nil {
		return d
	}

	hosts, err := db.ListSupplyHosts(ctx, Pools.Admin)
	if err != nil {
		log.Printf("infrastructure: ListSupplyHosts: %v", err)
		d.FlashError = "Failed to load infrastructure hosts: " + err.Error()
		return d
	}
	allIPs, err := db.ListSupplyHostIPs(ctx, Pools.Admin, 0)
	if err != nil {
		log.Printf("infrastructure: ListSupplyHostIPs: %v", err)
	}

	ipsByHost := make(map[int64][]infraIPView)
	for _, ip := range allIPs {
		ipsByHost[ip.HostID] = append(ipsByHost[ip.HostID], infraIPView{
			SupplyHostIP:  ip,
			CostDisplay:   formatCentsUSD(ip.MonthlyCostCents),
			RetailDisplay: formatCentsUSD(ip.RetailPriceCents),
		})
	}

	var totalCostCents int64
	var totalRAMMB int
	for _, h := range hosts {
		hv := infraHostView{
			SupplyHostRecord:   h,
			MonthlyCostDisplay: formatCentsUSD(h.MonthlyCostCents),
			RAMGBDisplay:       fmt.Sprintf("%.1f GB", float64(h.RAMMB)/1024.0),
			IPs:                ipsByHost[h.ID],
		}
		d.Hosts = append(d.Hosts, hv)
		d.TotalHosts++
		if h.InventoryStatus == "online" && h.VirtStackStatus == "ready" {
			d.ReadyOnlineHosts++
		}
		d.TotalVCPUs += h.CPUCores
		totalRAMMB += h.RAMMB
		d.TotalDiskGB += h.DiskGB
		d.AvailableIPs += h.AvailableIPs
		d.AssignedIPs += h.AssignedIPs
		d.TotalIPs += h.TotalIPs
		totalCostCents += h.MonthlyCostCents
	}
	d.TotalRAMGB = totalRAMMB / 1024
	d.MonthlySupplyCost = formatCentsUSD(totalCostCents)

	for i := range d.Hosts {
		if selectedHostID > 0 && d.Hosts[i].ID == selectedHostID {
			d.SelectedHost = &d.Hosts[i]
			break
		}
	}
	if d.SelectedHost == nil && len(d.Hosts) > 0 {
		d.SelectedHost = &d.Hosts[0]
	}
	return d
}

func renderSysadminInfrastructure(w http.ResponseWriter, r *http.Request, selectedHostID int64, flashOK, flashErr string) {
	user, _ := getSession(r)
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	pageData := buildInfraPageData(ctx, selectedHostID, flashOK, flashErr)

	innerTmpl, err := template.ParseFiles(
		filepath.Join("templates", "sysadmin", "infrastructure_content.html"),
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "sysadmin-infrastructure-content", pageData); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shellTmpl, err := template.ParseFiles(
		filepath.Join("templates", "sysadmin", "base.html"),
		filepath.Join("templates", "sysadmin", "shell.html"),
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shellData := SysadminData{
		User:      user,
		Nav:       sysadminNav,
		ActiveNav: "infrastructure",
		PageTitle: "Infrastructure & IP Pools",
		Content:   template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "sysadmin-base", shellData); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Auto-Provisioning Orchestrator ───────────────────────────────────────────

func runHostAutoProvisioning(ctx context.Context, host db.SupplyHostRecord, actorID int64) (providers.BootstrapReport, error) {
	if Pools == nil {
		return providers.BootstrapReport{}, errors.New("database unavailable")
	}
	drv, err := Drivers.Compute(providers.DriverKind(host.DriverKind))
	if err != nil {
		_ = db.MarkHostProvisioningFailed(ctx, Pools.Admin, host.ID, err.Error(), "")
		return providers.BootstrapReport{}, err
	}

	_ = db.MarkHostProvisioningStart(ctx, Pools.Admin, host.ID)
	_ = db.WriteSystemLog(ctx, Pools.Admin, "info", "infra-provisioner",
		fmt.Sprintf("Starting virtualization stack auto-provisioning on host %s (%s)", host.Name, host.HostnameOrIP),
		map[string]any{"host_id": host.ID, "driver": host.DriverKind})

	report, bootErr := drv.BootstrapHost(ctx, host.ToProviderHost())
	if bootErr != nil {
		_ = db.MarkHostProvisioningFailed(ctx, Pools.Admin, host.ID, bootErr.Error(), report.LogOutput)
		_ = db.WriteSystemLog(ctx, Pools.Admin, "error", "infra-provisioner",
			fmt.Sprintf("Auto-provisioning failed on host %s: %v", host.Name, bootErr),
			map[string]any{"host_id": host.ID})
		db.InsertAuditEvent(ctx, Pools.Admin, actorID, "sysadmin", "infra.host.bootstrap.failed", "warn",
			map[string]any{"host_id": host.ID, "name": host.Name}, bootErr.Error(), "", "")
		return report, bootErr
	}

	if err := db.MarkHostProvisioningComplete(ctx, Pools.Admin, host.ID, report.Capacity, report.LogOutput); err != nil {
		return report, err
	}
	_ = db.WriteSystemLog(ctx, Pools.Admin, "info", "infra-provisioner",
		fmt.Sprintf("Host %s provisioned and added to online inventory (%d vCPUs, %d MB RAM, KVM=%v)",
			host.Name, report.Capacity.CPUCores, report.Capacity.RAMMB, report.Capacity.KVMAvailable),
		map[string]any{"host_id": host.ID})
	db.InsertAuditEvent(ctx, Pools.Admin, actorID, "sysadmin", "infra.host.bootstrap.ready", "info",
		map[string]any{
			"host_id":       host.ID,
			"name":          host.Name,
			"cpu_cores":     report.Capacity.CPUCores,
			"ram_mb":        report.Capacity.RAMMB,
			"disk_gb":       report.Capacity.DiskGB,
			"kvm_available": report.Capacity.KVMAvailable,
		}, "", "", "")
	return report, nil
}

func parseDollarToCents(s string) int64 {
	s = strings.TrimSpace(strings.TrimPrefix(s, "$"))
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f < 0 {
		return 0
	}
	return int64(math.Round(f * 100))
}

// ── Sysadmin HTML Handlers (`/s/infrastructure/*`) ───────────────────────────

func sysadminInfraPageHandler(w http.ResponseWriter, r *http.Request) {
	selID, _ := strconv.ParseInt(r.URL.Query().Get("host"), 10, 64)
	renderSysadminInfrastructure(w, r, selID, r.URL.Query().Get("ok"), r.URL.Query().Get("err"))
}

func sysadminInfraCreateHostFormHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	if err := r.ParseForm(); err != nil {
		renderSysadminInfrastructure(w, r, 0, "", "Invalid form payload.")
		return
	}
	sshPort, _ := strconv.Atoi(r.FormValue("ssh_port"))
	cpuCores, _ := strconv.Atoi(r.FormValue("cpu_cores"))
	ramGB, _ := strconv.Atoi(r.FormValue("ram_gb"))
	diskGB, _ := strconv.Atoi(r.FormValue("disk_gb"))
	costCents := parseDollarToCents(r.FormValue("monthly_cost_usd"))

	in := db.CreateSupplyHostInput{
		Name:             r.FormValue("name"),
		HostnameOrIP:     r.FormValue("hostname_or_ip"),
		SSHPort:          sshPort,
		SSHUser:          r.FormValue("ssh_user"),
		AuthMethod:       r.FormValue("auth_method"),
		SSHCredential:    r.FormValue("ssh_credential"),
		DriverKind:       r.FormValue("driver_kind"),
		OwnershipType:    r.FormValue("ownership_type"),
		RegionSlug:       r.FormValue("region_slug"),
		BridgeInterface:  r.FormValue("bridge_interface"),
		StoragePoolPath:  r.FormValue("storage_pool_path"),
		CPUCores:         cpuCores,
		RAMMB:            ramGB * 1024,
		DiskGB:           diskGB,
		MonthlyCostCents: costCents,
		CreatedBy:        user.ID,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	host, err := db.CreateSupplyHost(ctx, Pools.Admin, in)
	if err != nil {
		renderSysadminInfrastructure(w, r, 0, "", "Failed to add host: "+err.Error())
		return
	}

	db.InsertAuditEvent(ctx, Pools.Admin, user.ID, string(user.Role), "infra.host.created", "info",
		map[string]any{"host_id": host.ID, "name": host.Name, "driver": host.DriverKind}, "", "", clientIP(r))

	if r.FormValue("auto_provision") == "on" {
		go func(h db.SupplyHostRecord, uid int64) {
			bgCtx, bgCancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer bgCancel()
			_, _ = runHostAutoProvisioning(bgCtx, h, uid)
		}(host, user.ID)
		http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&ok=Host+added+and+auto-provisioning+started", host.ID), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&ok=Linux+Nested+VPS+registered+in+inventory", host.ID), http.StatusSeeOther)
}

func sysadminInfraProbeHostFormHandler(w http.ResponseWriter, r *http.Request) {
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	host, ok, err := db.GetSupplyHostByID(ctx, Pools.Admin, hostID)
	if err != nil || !ok {
		http.Redirect(w, r, "/s/infrastructure/?err=Host+not+found", http.StatusSeeOther)
		return
	}
	drv, err := Drivers.Compute(providers.DriverKind(host.DriverKind))
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&err=%s", hostID, err.Error()), http.StatusSeeOther)
		return
	}
	if err := drv.ValidateConnection(ctx, host.ToProviderHost()); err != nil {
		_ = db.MarkHostProvisioningFailed(ctx, Pools.Admin, hostID, err.Error(), "")
		http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&err=SSH+validation+failed", hostID), http.StatusSeeOther)
		return
	}
	capInfo, err := drv.InspectCapacity(ctx, host.ToProviderHost())
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&err=Capacity+probe+failed", hostID), http.StatusSeeOther)
		return
	}
	_ = db.UpdateHostProbeCapacity(ctx, Pools.Admin, hostID, capInfo)
	http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&ok=SSH+verified+and+hardware+capacity+updated", hostID), http.StatusSeeOther)
}

func sysadminInfraProvisionHostFormHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	host, ok, err := db.GetSupplyHostByID(ctx, Pools.Admin, hostID)
	if err != nil || !ok {
		http.Redirect(w, r, "/s/infrastructure/?err=Host+not+found", http.StatusSeeOther)
		return
	}
	_ = db.MarkHostProvisioningStart(ctx, Pools.Admin, hostID)

	go func(h db.SupplyHostRecord, uid int64) {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer bgCancel()
		_, _ = runHostAutoProvisioning(bgCtx, h, uid)
	}(host, user.ID)

	http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&ok=Auto-provisioning+started+in+background", hostID), http.StatusSeeOther)
}

func sysadminInfraStatusFormHandler(w http.ResponseWriter, r *http.Request) {
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	_ = r.ParseForm()
	newStatus := strings.TrimSpace(r.FormValue("inventory_status"))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.SetHostInventoryStatus(ctx, Pools.Admin, hostID, newStatus); err != nil {
		http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&err=Invalid+inventory+status", hostID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&ok=Inventory+status+updated", hostID), http.StatusSeeOther)
}

func sysadminInfraDeleteHostFormHandler(w http.ResponseWriter, r *http.Request) {
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.DeleteSupplyHost(ctx, Pools.Admin, hostID); err != nil {
		http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&err=Cannot+delete+host+with+active+workloads", hostID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/s/infrastructure/?ok=Host+removed+from+inventory", http.StatusSeeOther)
}

func sysadminInfraAddIPsFormHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&err=Invalid+form", hostID), http.StatusSeeOther)
		return
	}
	cidr, _ := strconv.Atoi(r.FormValue("cidr_prefix"))
	costCents := parseDollarToCents(r.FormValue("monthly_cost_usd"))
	retailCents := parseDollarToCents(r.FormValue("retail_price_usd"))
	if r.FormValue("retail_price_usd") == "" {
		retailCents = 200 // $2.00/mo default retail per routed IPv4
	}

	entries, err := db.ParseManualIPList(
		r.FormValue("ip_list"),
		cidr,
		r.FormValue("gateway"),
		r.FormValue("mac_address"),
		costCents,
		retailCents,
		r.FormValue("notes"),
	)
	if err != nil {
		renderSysadminInfrastructure(w, r, hostID, "", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	if _, err := db.AddSupplyHostIPs(ctx, Pools.Admin, hostID, entries); err != nil {
		renderSysadminInfrastructure(w, r, hostID, "", "Failed to add IPs: "+err.Error())
		return
	}
	db.InsertAuditEvent(ctx, Pools.Admin, user.ID, string(user.Role), "infra.ips.added", "info",
		map[string]any{"host_id": hostID, "added_count": len(entries)}, "", "", clientIP(r))

	http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&ok=Added+%d+assignable+IP(s)+to+host+pool", hostID, len(entries)), http.StatusSeeOther)
}

func sysadminInfraDeleteIPFormHandler(w http.ResponseWriter, r *http.Request) {
	ipID, _ := strconv.ParseInt(r.PathValue("ip_id"), 10, 64)
	hostID, _ := strconv.ParseInt(r.URL.Query().Get("host"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.DeleteSupplyHostIP(ctx, Pools.Admin, ipID); err != nil {
		http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&err=Cannot+remove+assigned+IP", hostID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/s/infrastructure/?host=%d&ok=IP+removed+from+pool", hostID), http.StatusSeeOther)
}

func RegisterSysadminInfrastructureRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /s/infrastructure/", requireRole(RoleSysAdmin, sysadminInfraPageHandler))
	mux.HandleFunc("POST /s/infrastructure/hosts/", requireRole(RoleSysAdmin, sysadminInfraCreateHostFormHandler))
	mux.HandleFunc("POST /s/infrastructure/hosts/{id}/probe/", requireRole(RoleSysAdmin, sysadminInfraProbeHostFormHandler))
	mux.HandleFunc("POST /s/infrastructure/hosts/{id}/provision/", requireRole(RoleSysAdmin, sysadminInfraProvisionHostFormHandler))
	mux.HandleFunc("POST /s/infrastructure/hosts/{id}/status/", requireRole(RoleSysAdmin, sysadminInfraStatusFormHandler))
	mux.HandleFunc("POST /s/infrastructure/hosts/{id}/delete/", requireRole(RoleSysAdmin, sysadminInfraDeleteHostFormHandler))
	mux.HandleFunc("POST /s/infrastructure/hosts/{id}/ips/", requireRole(RoleSysAdmin, sysadminInfraAddIPsFormHandler))
	mux.HandleFunc("POST /s/infrastructure/ips/{ip_id}/delete/", requireRole(RoleSysAdmin, sysadminInfraDeleteIPFormHandler))
}

// ── Sysadmin JSON REST API (`/api/sysadmin/infrastructure/*`) ────────────────

type apiSupplyHost struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	HostnameOrIP      string    `json:"hostname_or_ip"`
	SSHPort           int       `json:"ssh_port"`
	SSHUser           string    `json:"ssh_user"`
	AuthMethod        string    `json:"auth_method"`
	DriverKind        string    `json:"driver_kind"`
	OwnershipType     string    `json:"ownership_type"`
	RegionSlug        string    `json:"region_slug"`
	BridgeInterface   string    `json:"bridge_interface"`
	StoragePoolPath   string    `json:"storage_pool_path"`
	CPUCores          int       `json:"cpu_cores"`
	RAMMB             int       `json:"ram_mb"`
	DiskGB            int       `json:"disk_gb"`
	KVMAvailable      bool      `json:"kvm_available"`
	OSInfo            string    `json:"os_info"`
	HypervisorVersion string    `json:"hypervisor_version"`
	MonthlyCostCents  int64     `json:"monthly_cost_cents"`
	VirtStackStatus   string    `json:"virt_stack_status"`
	InventoryStatus   string    `json:"inventory_status"`
	ProvisionLog      string    `json:"provision_log,omitempty"`
	LastError         string    `json:"last_error,omitempty"`
	TotalIPs          int       `json:"total_ips"`
	AvailableIPs      int       `json:"available_ips"`
	AssignedIPs       int       `json:"assigned_ips"`
	ActiveVMs         int       `json:"active_vms"`
	CreatedAt         time.Time `json:"created_at"`
}

type apiSupplyIP struct {
	ID                   int64  `json:"id"`
	HostID               int64  `json:"host_id"`
	HostName             string `json:"host_name"`
	IPAddress            string `json:"ip_address"`
	CIDRPrefix           int    `json:"cidr_prefix"`
	Gateway              string `json:"gateway"`
	MACAddress           string `json:"mac_address,omitempty"`
	MonthlyCostCents     int64  `json:"monthly_cost_cents"`
	RetailPriceCents     int64  `json:"retail_price_cents"`
	Status               string `json:"status"`
	AssignedVMResourceID *int64 `json:"assigned_vm_resource_id,omitempty"`
	AssignedVMName       string `json:"assigned_vm_name,omitempty"`
	Notes                string `json:"notes,omitempty"`
}

func toAPISupplyHost(h db.SupplyHostRecord, includeLog bool) apiSupplyHost {
	out := apiSupplyHost{
		ID: h.ID, Name: h.Name, HostnameOrIP: h.HostnameOrIP,
		SSHPort: h.SSHPort, SSHUser: h.SSHUser, AuthMethod: h.AuthMethod,
		DriverKind: h.DriverKind, OwnershipType: h.OwnershipType, RegionSlug: h.RegionSlug,
		BridgeInterface: h.BridgeInterface, StoragePoolPath: h.StoragePoolPath,
		CPUCores: h.CPUCores, RAMMB: h.RAMMB, DiskGB: h.DiskGB,
		KVMAvailable: h.KVMAvailable, OSInfo: h.OSInfo, HypervisorVersion: h.HypervisorVersion,
		MonthlyCostCents: h.MonthlyCostCents,
		VirtStackStatus:  h.VirtStackStatus, InventoryStatus: h.InventoryStatus,
		LastError:    h.LastError,
		TotalIPs:     h.TotalIPs, AvailableIPs: h.AvailableIPs, AssignedIPs: h.AssignedIPs,
		ActiveVMs: h.ActiveVMs, CreatedAt: h.CreatedAt,
	}
	if includeLog {
		out.ProvisionLog = h.ProvisionLog
	}
	return out
}

func toAPISupplyIP(ip db.SupplyHostIP) apiSupplyIP {
	return apiSupplyIP{
		ID: ip.ID, HostID: ip.HostID, HostName: ip.HostName,
		IPAddress: ip.IPAddress, CIDRPrefix: ip.CIDRPrefix,
		Gateway: ip.Gateway, MACAddress: ip.MACAddress,
		MonthlyCostCents: ip.MonthlyCostCents, RetailPriceCents: ip.RetailPriceCents,
		Status:               ip.Status,
		AssignedVMResourceID: ip.AssignedVMResourceID,
		AssignedVMName:       ip.AssignedVMName,
		Notes:                ip.Notes,
	}
}

func requireSysadminAPI(w http.ResponseWriter, r *http.Request) (User, bool) {
	u, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return User{}, false
	}
	if u.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return User{}, false
	}
	if Pools == nil {
		apiError(w, http.StatusServiceUnavailable, "Database unavailable.")
		return User{}, false
	}
	return u, true
}

func apiSysadminInfraHostsListHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireSysadminAPI(w, r); !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	hosts, err := db.ListSupplyHosts(ctx, Pools.Admin)
	if err != nil {
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]apiSupplyHost, len(hosts))
	for i, h := range hosts {
		out[i] = toAPISupplyHost(h, false)
	}
	apiOK(w, map[string]any{
		"hosts":   out,
		"total":   len(out),
		"drivers": Drivers.ListComputeKinds(),
	})
}

type apiCreateSupplyHostRequest struct {
	Name             string `json:"name"`
	HostnameOrIP     string `json:"hostname_or_ip"`
	SSHPort          int    `json:"ssh_port"`
	SSHUser          string `json:"ssh_user"`
	AuthMethod       string `json:"auth_method"`
	SSHCredential    string `json:"ssh_credential"`
	DriverKind       string `json:"driver_kind"`
	OwnershipType    string `json:"ownership_type"`
	RegionSlug       string `json:"region_slug"`
	BridgeInterface  string `json:"bridge_interface"`
	StoragePoolPath  string `json:"storage_pool_path"`
	CPUCores         int    `json:"cpu_cores"`
	RAMMB            int    `json:"ram_mb"`
	DiskGB           int    `json:"disk_gb"`
	MonthlyCostCents int64  `json:"monthly_cost_cents"`
	AutoProvision    bool   `json:"auto_provision"`
}

func apiSysadminInfraHostsCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := requireSysadminAPI(w, r)
	if !ok {
		return
	}
	var req apiCreateSupplyHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	host, err := db.CreateSupplyHost(ctx, Pools.Admin, db.CreateSupplyHostInput{
		Name:             req.Name,
		HostnameOrIP:     req.HostnameOrIP,
		SSHPort:          req.SSHPort,
		SSHUser:          req.SSHUser,
		AuthMethod:       req.AuthMethod,
		SSHCredential:    req.SSHCredential,
		DriverKind:       req.DriverKind,
		OwnershipType:    req.OwnershipType,
		RegionSlug:       req.RegionSlug,
		BridgeInterface:  req.BridgeInterface,
		StoragePoolPath:  req.StoragePoolPath,
		CPUCores:         req.CPUCores,
		RAMMB:            req.RAMMB,
		DiskGB:           req.DiskGB,
		MonthlyCostCents: req.MonthlyCostCents,
		CreatedBy:        user.ID,
	})
	if err != nil {
		apiError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.AutoProvision {
		go func(h db.SupplyHostRecord, uid int64) {
			bgCtx, bgCancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer bgCancel()
			_, _ = runHostAutoProvisioning(bgCtx, h, uid)
		}(host, user.ID)
	}
	apiCreated(w, toAPISupplyHost(host, true))
}

func apiSysadminInfraHostsGetHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireSysadminAPI(w, r); !ok {
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	host, found, err := db.GetSupplyHostByID(ctx, Pools.Admin, hostID)
	if err != nil {
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !found {
		apiError(w, http.StatusNotFound, "Supply host not found.")
		return
	}
	ips, _ := db.ListSupplyHostIPs(ctx, Pools.Admin, hostID)
	apiIPs := make([]apiSupplyIP, len(ips))
	for i, ip := range ips {
		apiIPs[i] = toAPISupplyIP(ip)
	}
	apiOK(w, map[string]any{
		"host": toAPISupplyHost(host, true),
		"ips":  apiIPs,
	})
}

func apiSysadminInfraHostsProbeHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireSysadminAPI(w, r); !ok {
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	host, found, err := db.GetSupplyHostByID(ctx, Pools.Admin, hostID)
	if err != nil || !found {
		apiError(w, http.StatusNotFound, "Supply host not found.")
		return
	}
	drv, err := Drivers.Compute(providers.DriverKind(host.DriverKind))
	if err != nil {
		apiError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := drv.ValidateConnection(ctx, host.ToProviderHost()); err != nil {
		apiError(w, http.StatusBadGateway, err.Error())
		return
	}
	capInfo, err := drv.InspectCapacity(ctx, host.ToProviderHost())
	if err != nil {
		apiError(w, http.StatusBadGateway, err.Error())
		return
	}
	_ = db.UpdateHostProbeCapacity(ctx, Pools.Admin, hostID, capInfo)
	updated, _, _ := db.GetSupplyHostByID(ctx, Pools.Admin, hostID)
	apiOK(w, toAPISupplyHost(updated, false))
}

func apiSysadminInfraHostsProvisionHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := requireSysadminAPI(w, r)
	if !ok {
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	host, found, err := db.GetSupplyHostByID(ctx, Pools.Admin, hostID)
	if err != nil || !found {
		apiError(w, http.StatusNotFound, "Supply host not found.")
		return
	}

	if r.URL.Query().Get("sync") == "1" {
		report, bootErr := runHostAutoProvisioning(ctx, host, user.ID)
		updated, _, _ := db.GetSupplyHostByID(ctx, Pools.Admin, hostID)
		if bootErr != nil {
			apiError(w, http.StatusBadGateway, bootErr.Error())
			return
		}
		apiOK(w, map[string]any{
			"ok":     true,
			"host":   toAPISupplyHost(updated, true),
			"report": report,
		})
		return
	}

	_ = db.MarkHostProvisioningStart(ctx, Pools.Admin, hostID)
	go func(h db.SupplyHostRecord, uid int64) {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer bgCancel()
		_, _ = runHostAutoProvisioning(bgCtx, h, uid)
	}(host, user.ID)

	apiOK(w, map[string]any{
		"ok":                true,
		"message":           "Virtualization stack auto-provisioning started.",
		"virt_stack_status": "provisioning",
	})
}

func apiSysadminInfraHostsPatchHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireSysadminAPI(w, r); !ok {
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var body struct {
		InventoryStatus string `json:"inventory_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.SetHostInventoryStatus(ctx, Pools.Admin, hostID, body.InventoryStatus); err != nil {
		apiError(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, found, _ := db.GetSupplyHostByID(ctx, Pools.Admin, hostID)
	if !found {
		apiError(w, http.StatusNotFound, "Supply host not found.")
		return
	}
	apiOK(w, toAPISupplyHost(updated, false))
}

func apiSysadminInfraHostsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireSysadminAPI(w, r); !ok {
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.DeleteSupplyHost(ctx, Pools.Admin, hostID); err != nil {
		if errors.Is(err, db.ErrHostHasActiveWorkloads) {
			apiError(w, http.StatusConflict, err.Error())
			return
		}
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiOK(w, map[string]any{"ok": true, "message": "Supply host deleted."})
}

func apiSysadminInfraIPsListHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireSysadminAPI(w, r); !ok {
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	ips, err := db.ListSupplyHostIPs(ctx, Pools.Admin, hostID)
	if err != nil {
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]apiSupplyIP, len(ips))
	for i, ip := range ips {
		out[i] = toAPISupplyIP(ip)
	}
	apiOK(w, map[string]any{"ips": out, "total": len(out)})
}

type apiAddHostIPsRequest struct {
	IPs              []string `json:"ips"`
	IPList           string   `json:"ip_list"`
	CIDRPrefix       int      `json:"cidr_prefix"`
	Gateway          string   `json:"gateway"`
	MACAddress       string   `json:"mac_address"`
	MonthlyCostCents int64    `json:"monthly_cost_cents"`
	RetailPriceCents *int64   `json:"retail_price_cents"`
	Notes            string   `json:"notes"`
}

func apiSysadminInfraIPsCreateHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireSysadminAPI(w, r); !ok {
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req apiAddHostIPsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	combined := req.IPList
	if len(req.IPs) > 0 {
		combined = combined + "\n" + strings.Join(req.IPs, "\n")
	}
	retail := int64(200)
	if req.RetailPriceCents != nil {
		retail = *req.RetailPriceCents
	}

	entries, err := db.ParseManualIPList(combined, req.CIDRPrefix, req.Gateway, req.MACAddress, req.MonthlyCostCents, retail, req.Notes)
	if err != nil {
		apiError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	if _, found, _ := db.GetSupplyHostByID(ctx, Pools.Admin, hostID); !found {
		apiError(w, http.StatusNotFound, "Supply host not found.")
		return
	}

	ips, err := db.AddSupplyHostIPs(ctx, Pools.Admin, hostID, entries)
	if err != nil {
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]apiSupplyIP, len(ips))
	for i, ip := range ips {
		out[i] = toAPISupplyIP(ip)
	}
	apiCreated(w, map[string]any{"ips": out, "total": len(out)})
}

func apiSysadminInfraIPDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireSysadminAPI(w, r); !ok {
		return
	}
	ipID, _ := strconv.ParseInt(r.PathValue("ip_id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.DeleteSupplyHostIP(ctx, Pools.Admin, ipID); err != nil {
		if errors.Is(err, db.ErrIPCurrentlyAssigned) {
			apiError(w, http.StatusConflict, err.Error())
			return
		}
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiOK(w, map[string]any{"ok": true, "message": "IP removed from pool."})
}

func RegisterSysadminInfrastructureAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/sysadmin/infrastructure/hosts", apiSysadminInfraHostsListHandler)
	mux.HandleFunc("POST /api/sysadmin/infrastructure/hosts", apiSysadminInfraHostsCreateHandler)
	mux.HandleFunc("GET /api/sysadmin/infrastructure/hosts/{id}", apiSysadminInfraHostsGetHandler)
	mux.HandleFunc("POST /api/sysadmin/infrastructure/hosts/{id}/probe", apiSysadminInfraHostsProbeHandler)
	mux.HandleFunc("POST /api/sysadmin/infrastructure/hosts/{id}/provision", apiSysadminInfraHostsProvisionHandler)
	mux.HandleFunc("PATCH /api/sysadmin/infrastructure/hosts/{id}", apiSysadminInfraHostsPatchHandler)
	mux.HandleFunc("DELETE /api/sysadmin/infrastructure/hosts/{id}", apiSysadminInfraHostsDeleteHandler)
	mux.HandleFunc("GET /api/sysadmin/infrastructure/hosts/{id}/ips", apiSysadminInfraIPsListHandler)
	mux.HandleFunc("POST /api/sysadmin/infrastructure/hosts/{id}/ips", apiSysadminInfraIPsCreateHandler)
	mux.HandleFunc("DELETE /api/sysadmin/infrastructure/ips/{ip_id}", apiSysadminInfraIPDeleteHandler)
}
