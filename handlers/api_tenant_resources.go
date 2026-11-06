package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"platform/db"
	"platform/providers"
)

// ── In-memory fallback store (used only when DB pool is nil) ─────────────────

type Resource struct {
	ID      string
	Name    string
	Type    string
	Status  string
	Region  string
	CPU     string
	Memory  string
	Created string
	Updated string
}

type resourceStoreType struct {
	mu        sync.RWMutex
	resources map[string][]Resource
}

var resourceStore = &resourceStoreType{
	resources: map[string][]Resource{},
}

func (s *resourceStoreType) list(slug string) []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.resources[slug]
	out := make([]Resource, len(src))
	copy(out, src)
	return out
}

func (s *resourceStoreType) getByID(slug, id string) (Resource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.resources[slug] {
		if r.ID == id {
			return r, true
		}
	}
	return Resource{}, false
}

func (s *resourceStoreType) add(slug string, r Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[slug] = append(s.resources[slug], r)
}

func (s *resourceStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.resources[slug]
	for i, r := range list {
		if r.ID == id {
			s.resources[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterResources(all []Resource, search, status, rtype string) []Resource {
	var out []Resource
	q := strings.ToLower(search)
	for _, r := range all {
		if status != "" && r.Status != status {
			continue
		}
		if rtype != "" && r.Type != rtype {
			continue
		}
		if q != "" {
			hay := strings.ToLower(r.Name + " " + r.Type + " " + r.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

// ── Tenant Auth + DB Resolution Helper ───────────────────────────────────────

func requireTenantAdminAPI(w http.ResponseWriter, r *http.Request) (User, Tenant, bool) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return User{}, Tenant{}, false
	}
	slug := r.PathValue("slug")
	t, found := store.getBySlug(slug)
	if !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return User{}, Tenant{}, false
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return User{}, Tenant{}, false
	}
	if t.DBID == 0 && Pools != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if dbID, err := db.LookupTenantIDBySlug(ctx, Pools.Admin, slug); err == nil {
			t.DBID = dbID
		}
	}
	return user, t, true
}

// ── API DTOs ─────────────────────────────────────────────────────────────────

type apiResource struct {
	ID      string `json:"id"`
	Kind    string `json:"kind,omitempty"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	HostIP  string `json:"host_or_ip,omitempty"`
	Region  string `json:"region"`
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

type apiSupplyHost struct {
	ID                int64     `json:"id"`
	TenantID          int64     `json:"tenant_id"`
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
	TenantID             int64  `json:"tenant_id"`
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
		ID: h.ID, TenantID: h.TenantID, Name: h.Name, HostnameOrIP: h.HostnameOrIP,
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
		ID: ip.ID, TenantID: ip.TenantID, HostID: ip.HostID, HostName: ip.HostName,
		IPAddress: ip.IPAddress, CIDRPrefix: ip.CIDRPrefix,
		Gateway: ip.Gateway, MACAddress: ip.MACAddress,
		MonthlyCostCents: ip.MonthlyCostCents, RetailPriceCents: ip.RetailPriceCents,
		Status:               ip.Status,
		AssignedVMResourceID: ip.AssignedVMResourceID,
		AssignedVMName:       ip.AssignedVMName,
		Notes:                ip.Notes,
	}
}

// ── GET /api/tenant/{slug}/resources ─────────────────────────────────────────

func apiTenantResourcesListHandler(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}

	search := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	rtype := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type")))

	if Pools == nil || tenant.DBID <= 0 {
		resources := filterResources(resourceStore.list(tenant.Slug), search, status, rtype)
		out := make([]apiResource, len(resources))
		for i, res := range resources {
			out[i] = apiResource{
				ID: res.ID, Name: res.Name, Type: res.Type,
				Status: res.Status, Region: res.Region,
				CPU: res.CPU, Memory: res.Memory,
				Created: res.Created, Updated: res.Updated,
			}
		}
		apiOK(w, map[string]any{"resources": out, "total": len(out)})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	d := buildTenantResourcesData(ctx, tenant, "all", search, status, 0, "", "")
	var out []apiResource
	for _, row := range d.Rows {
		if rtype != "" && row.CategoryKey != rtype {
			continue
		}
		out = append(out, apiResource{
			ID:     row.ID,
			Kind:   row.RowKind,
			Name:   row.Name,
			Type:   row.CategoryKey,
			Status: row.Status,
			HostIP: row.HostOrIP,
			Region: row.Region,
			CPU:    row.CPUDisplay,
			Memory: row.RAMDisplay,
		})
	}
	if out == nil {
		out = []apiResource{}
	}
	apiOK(w, map[string]any{
		"resources":     out,
		"total":         len(out),
		"available_ips": d.AvailableIPs,
		"total_ips":     d.TotalIPs,
	})
}

// ── POST /api/tenant/{slug}/resources ────────────────────────────────────────

type apiCreateResourceRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Region     string `json:"region"`
	HostOrIP   string `json:"hostname_or_ip"`
	APIURL     string `json:"api_url"`
	AuthMethod string `json:"auth_method"`
	Credential string `json:"credential"`
	Tags       string `json:"tags"`
}

func apiTenantResourcesCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}

	var req apiCreateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	rtype := strings.ToLower(strings.TrimSpace(req.Type))
	region := strings.TrimSpace(req.Region)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	if rtype == "" {
		rtype = "linux-vm"
	}
	if region == "" {
		region = "eu-central-1"
	}

	if Pools == nil || tenant.DBID <= 0 {
		now := time.Now().UTC().Format(time.RFC3339)
		res := Resource{
			ID: newID(), Name: name, Type: rtype,
			Status: "online", Region: region,
			CPU: "—", Memory: "—",
			Created: now, Updated: now,
		}
		resourceStore.add(tenant.Slug, res)
		apiCreated(w, apiResource{
			ID: res.ID, Name: res.Name, Type: res.Type,
			Status: res.Status, Region: res.Region,
			CPU: res.CPU, Memory: res.Memory,
			Created: res.Created, Updated: res.Updated,
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	meta, known := resourceCategoryMap[rtype]
	if !known {
		meta = categoryForResourceType(rtype)
	}
	if meta.DriverKind != "" && strings.TrimSpace(req.HostOrIP) != "" {
		host, err := db.CreateSupplyHost(ctx, Pools.Admin, db.CreateSupplyHostInput{
			TenantID:      tenant.DBID,
			Name:          name,
			HostnameOrIP:  req.HostOrIP,
			AuthMethod:    req.AuthMethod,
			SSHCredential: req.Credential,
			DriverKind:    meta.DriverKind,
			RegionSlug:    region,
			CreatedBy:     user.ID,
		})
		if err != nil {
			apiError(w, http.StatusBadRequest, err.Error())
			return
		}
		apiCreated(w, apiResource{
			ID:      fmt.Sprintf("host-%d", host.ID),
			Kind:    "supply_host",
			Name:    host.Name,
			Type:    meta.Key,
			Status:  host.InventoryStatus,
			HostIP:  host.HostnameOrIP,
			Region:  host.RegionSlug,
			Created: host.CreatedAt.UTC().Format(time.RFC3339),
			Updated: host.UpdatedAt.UTC().Format(time.RFC3339),
		})
		return
	}

	created, err := db.CreateTenantResourceFull(ctx, Pools.Admin, db.CreateTenantResourceInput{
		TenantID:     tenant.DBID,
		Name:         name,
		ResourceType: meta.Key,
		Provider:     meta.Label,
		Region:       region,
		Status:       "online",
		IPAddress:    req.HostOrIP,
		APIURL:       req.APIURL,
		AuthMethod:   req.AuthMethod,
		Credential:   req.Credential,
		Tags:         req.Tags,
	})
	if err != nil {
		apiError(w, http.StatusBadRequest, err.Error())
		return
	}
	apiCreated(w, apiResource{
		ID:      strconv.FormatInt(created.ID, 10),
		Kind:    "tenant_resource",
		Name:    created.Name,
		Type:    created.ResourceType,
		Status:  created.Status,
		HostIP:  created.IPAddress,
		Region:  created.Region,
		Created: created.CreatedAt.UTC().Format(time.RFC3339),
		Updated: created.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ── GET / DELETE /api/tenant/{slug}/resources/{id} ───────────────────────────

func apiTenantResourcesGetHandler(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if Pools == nil || tenant.DBID <= 0 {
		res, ok := resourceStore.getByID(tenant.Slug, id)
		if !ok {
			apiError(w, http.StatusNotFound, "Resource not found.")
			return
		}
		apiOK(w, apiResource{
			ID: res.ID, Name: res.Name, Type: res.Type,
			Status: res.Status, Region: res.Region,
			CPU: res.CPU, Memory: res.Memory,
			Created: res.Created, Updated: res.Updated,
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if strings.HasPrefix(id, "host-") {
		hid, _ := strconv.ParseInt(strings.TrimPrefix(id, "host-"), 10, 64)
		h, found, err := db.GetTenantSupplyHostByID(ctx, Pools.Admin, tenant.DBID, hid)
		if err != nil || !found {
			apiError(w, http.StatusNotFound, "Resource not found.")
			return
		}
		apiOK(w, toAPISupplyHost(h, true))
		return
	}

	list, err := db.ListTenantResources(ctx, Pools.Admin, tenant.DBID, "", "")
	if err == nil {
		for _, item := range list {
			if strconv.FormatInt(item.ID, 10) == id {
				apiOK(w, apiResource{
					ID:      id,
					Kind:    "tenant_resource",
					Name:    item.Name,
					Type:    item.ResourceType,
					Status:  item.Status,
					HostIP:  item.IPAddress,
					Region:  item.Region,
					Created: item.CreatedAt.UTC().Format(time.RFC3339),
					Updated: item.UpdatedAt.UTC().Format(time.RFC3339),
				})
				return
			}
		}
	}
	apiError(w, http.StatusNotFound, "Resource not found.")
}

func apiTenantResourcesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if Pools == nil || tenant.DBID <= 0 {
		if _, ok := resourceStore.getByID(tenant.Slug, id); !ok {
			apiError(w, http.StatusNotFound, "Resource not found.")
			return
		}
		resourceStore.delete(tenant.Slug, id)
		apiOK(w, map[string]any{"ok": true, "message": "Resource deleted."})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if strings.HasPrefix(id, "host-") {
		hid, _ := strconv.ParseInt(strings.TrimPrefix(id, "host-"), 10, 64)
		if err := db.DeleteTenantSupplyHost(ctx, Pools.Admin, tenant.DBID, hid); err != nil {
			if errors.Is(err, db.ErrHostHasActiveWorkloads) {
				apiError(w, http.StatusConflict, err.Error())
				return
			}
			apiError(w, http.StatusNotFound, "Resource not found.")
			return
		}
		apiOK(w, map[string]any{"ok": true, "message": "Supply host deleted."})
		return
	}

	numID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		apiError(w, http.StatusNotFound, "Resource not found.")
		return
	}
	if err := db.DeleteTenantResource(ctx, Pools.Admin, tenant.DBID, numID); err != nil {
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiOK(w, map[string]any{"ok": true, "message": "Resource deleted."})
}

// ── Tenant Supply Hosts & Manual Assignable IP Pool JSON Endpoints ───────────

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

func apiTenantHostsListHandler(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	if Pools == nil || tenant.DBID <= 0 {
		apiError(w, http.StatusServiceUnavailable, "Database unavailable.")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	hosts, err := db.ListSupplyHostsByTenant(ctx, Pools.Admin, tenant.DBID)
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

func apiTenantHostsCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	if Pools == nil || tenant.DBID <= 0 {
		apiError(w, http.StatusServiceUnavailable, "Database unavailable.")
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
		TenantID:         tenant.DBID,
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

func apiTenantHostsGetHandler(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	if Pools == nil || tenant.DBID <= 0 {
		apiError(w, http.StatusServiceUnavailable, "Database unavailable.")
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	host, found, err := db.GetTenantSupplyHostByID(ctx, Pools.Admin, tenant.DBID, hostID)
	if err != nil {
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !found {
		apiError(w, http.StatusNotFound, "Supply host not found.")
		return
	}
	ips, _ := db.ListTenantSupplyHostIPs(ctx, Pools.Admin, tenant.DBID, hostID)
	apiIPs := make([]apiSupplyIP, len(ips))
	for i, ip := range ips {
		apiIPs[i] = toAPISupplyIP(ip)
	}
	apiOK(w, map[string]any{
		"host": toAPISupplyHost(host, true),
		"ips":  apiIPs,
	})
}

func apiTenantHostsProbeHandler(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	if Pools == nil || tenant.DBID <= 0 {
		apiError(w, http.StatusServiceUnavailable, "Database unavailable.")
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	host, found, err := db.GetTenantSupplyHostByID(ctx, Pools.Admin, tenant.DBID, hostID)
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
	updated, _, _ := db.GetTenantSupplyHostByID(ctx, Pools.Admin, tenant.DBID, hostID)
	apiOK(w, toAPISupplyHost(updated, false))
}

func apiTenantHostsProvisionHandler(w http.ResponseWriter, r *http.Request) {
	user, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	if Pools == nil || tenant.DBID <= 0 {
		apiError(w, http.StatusServiceUnavailable, "Database unavailable.")
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	host, found, err := db.GetTenantSupplyHostByID(ctx, Pools.Admin, tenant.DBID, hostID)
	if err != nil || !found {
		apiError(w, http.StatusNotFound, "Supply host not found.")
		return
	}

	if r.URL.Query().Get("sync") == "1" {
		report, bootErr := runHostAutoProvisioning(ctx, host, user.ID)
		updated, _, _ := db.GetTenantSupplyHostByID(ctx, Pools.Admin, tenant.DBID, hostID)
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

func apiTenantHostsPatchHandler(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	if Pools == nil || tenant.DBID <= 0 {
		apiError(w, http.StatusServiceUnavailable, "Database unavailable.")
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

	if err := db.SetTenantHostInventoryStatus(ctx, Pools.Admin, tenant.DBID, hostID, body.InventoryStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			apiError(w, http.StatusNotFound, "Supply host not found.")
			return
		}
		apiError(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, _, _ := db.GetTenantSupplyHostByID(ctx, Pools.Admin, tenant.DBID, hostID)
	apiOK(w, toAPISupplyHost(updated, false))
}

func apiTenantHostsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	if Pools == nil || tenant.DBID <= 0 {
		apiError(w, http.StatusServiceUnavailable, "Database unavailable.")
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.DeleteTenantSupplyHost(ctx, Pools.Admin, tenant.DBID, hostID); err != nil {
		if errors.Is(err, db.ErrHostHasActiveWorkloads) {
			apiError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			apiError(w, http.StatusNotFound, "Supply host not found.")
			return
		}
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiOK(w, map[string]any{"ok": true, "message": "Supply host deleted."})
}

func apiTenantIPsListHandler(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	if Pools == nil || tenant.DBID <= 0 {
		apiError(w, http.StatusServiceUnavailable, "Database unavailable.")
		return
	}
	hostID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if hostID <= 0 {
		hostID, _ = strconv.ParseInt(r.URL.Query().Get("host_id"), 10, 64)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	ips, err := db.ListTenantSupplyHostIPs(ctx, Pools.Admin, tenant.DBID, hostID)
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

func apiTenantIPsCreateHandler(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	if Pools == nil || tenant.DBID <= 0 {
		apiError(w, http.StatusServiceUnavailable, "Database unavailable.")
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

	ips, err := db.AddTenantSupplyHostIPs(ctx, Pools.Admin, tenant.DBID, hostID, entries)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			apiError(w, http.StatusNotFound, "Supply host not found.")
			return
		}
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]apiSupplyIP, len(ips))
	for i, ip := range ips {
		out[i] = toAPISupplyIP(ip)
	}
	apiCreated(w, map[string]any{"ips": out, "total": len(out)})
}

func apiTenantIPDeleteHandler(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := requireTenantAdminAPI(w, r)
	if !ok {
		return
	}
	if Pools == nil || tenant.DBID <= 0 {
		apiError(w, http.StatusServiceUnavailable, "Database unavailable.")
		return
	}
	ipID, _ := strconv.ParseInt(r.PathValue("ip_id"), 10, 64)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.DeleteTenantSupplyHostIP(ctx, Pools.Admin, tenant.DBID, ipID); err != nil {
		if errors.Is(err, db.ErrIPCurrentlyAssigned) {
			apiError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			apiError(w, http.StatusNotFound, "IP not found.")
			return
		}
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiOK(w, map[string]any{"ok": true, "message": "IP removed from pool."})
}

func RegisterTenantInfrastructureAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/tenant/{slug}/resources/hosts", apiTenantHostsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/resources/hosts", apiTenantHostsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/resources/hosts/{id}", apiTenantHostsGetHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/resources/hosts/{id}/probe", apiTenantHostsProbeHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/resources/hosts/{id}/provision", apiTenantHostsProvisionHandler)
	mux.HandleFunc("PATCH /api/tenant/{slug}/resources/hosts/{id}", apiTenantHostsPatchHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/resources/hosts/{id}", apiTenantHostsDeleteHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/resources/ips", apiTenantIPsListHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/resources/hosts/{id}/ips", apiTenantIPsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/resources/hosts/{id}/ips", apiTenantIPsCreateHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/resources/ips/{ip_id}", apiTenantIPDeleteHandler)
}
