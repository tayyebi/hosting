package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"platform/accounting"
	"platform/db"
	"platform/providers"
)

// CustomerVMPurchaseInput carries normalized VM sizing & customer context
// from workspace compute endpoints.
type CustomerVMPurchaseInput struct {
	TenantSlug      string
	WorkspaceIDStr  string
	ActorUserID     int64
	Name            string
	InstanceType    string
	Distro          string
	Version         string
	SSHPublicKey    string
	VCPUs           int
	MemoryGB        int
	DiskGB          int
	PreferredRegion string
}

// CustomerVMPurchaseResult contains the provisioned VM details, its allocated
// public IPv4 from the server's manual IP pool, and computed pricing.
type CustomerVMPurchaseResult struct {
	VMResourceID       int64
	IPResourceID       int64
	ExternalID         string
	SupplyHostID       int64
	SupplyHostName     string
	RegionSlug         string
	AssignedIPv4       string
	Status             string
	MonthlyCostDisplay string
}

func resolveTenantAndWorkspaceDBIDs(ctx context.Context, slug, wsIDStr string) (int64, int64, error) {
	t, ok := store.getBySlug(slug)
	if !ok {
		return 0, 0, fmt.Errorf("tenant %q not found", slug)
	}
	tenantDBID := t.DBID
	if tenantDBID == 0 && Pools != nil {
		id, err := db.LookupTenantIDBySlug(ctx, Pools.Admin, slug)
		if err != nil {
			return 0, 0, err
		}
		tenantDBID = id
	}
	_, ws, ok := clientStore.getWorkspace(slug, wsIDStr)
	if !ok {
		return 0, 0, fmt.Errorf("workspace %q not found", wsIDStr)
	}
	wsDBID := ws.DBID
	if wsDBID == 0 && Pools != nil && tenantDBID > 0 {
		id, err := db.EnsureWorkspaceRow(ctx, Pools.Admin, tenantDBID, "ws-"+wsIDStr, ws.Name)
		if err != nil {
			return 0, 0, err
		}
		wsDBID = id
	}
	return tenantDBID, wsDBID, nil
}

// ProvisionCustomerVM executes the end-to-end purchase & provisioning flow:
//  1. Resolves tenant & workspace DB IDs.
//  2. Estimates COGS and quotes retail VM price via the Cost-Centric Accounting engine.
//  3. Atomically selects an `online` + `ready` SupplyHost and locks a free IP
//     from that server's manual assignable IP pool (`FOR UPDATE SKIP LOCKED`),
//     creating both the parent VM `resources` row and the child payable IPv4
//     `resources` row (`parent_resource_id`).
//  4. Invokes the concrete `ComputeDriver` (e.g. `LinuxNestedQEMUDriver`) over
//     SSH to create the guest VM with cloud-init static IP configuration.
//  5. Records Cost & Revenue events in `MarginLedger`.
func ProvisionCustomerVM(ctx context.Context, in CustomerVMPurchaseInput) (CustomerVMPurchaseResult, error) {
	if Pools == nil {
		return CustomerVMPurchaseResult{}, fmt.Errorf("database pool is not initialized")
	}

	tenantDBID, wsDBID, err := resolveTenantAndWorkspaceDBIDs(ctx, in.TenantSlug, in.WorkspaceIDStr)
	if err != nil {
		return CustomerVMPurchaseResult{}, err
	}

	vcpus := in.VCPUs
	if vcpus <= 0 {
		vcpus = 1
	}
	memGB := in.MemoryGB
	if memGB <= 0 {
		memGB = 2
	}
	diskGB := in.DiskGB
	if diskGB <= 0 {
		diskGB = 20
	}

	// Compute baseline COGS & Retail Quote via CostCalc + PricingEngine
	cogs, _ := CostCalc.CalculateSupplyCost(ctx, accounting.SupplyCostInput{
		Model:                accounting.CostModelAmortizedHardware,
		HostMonthlyCostCents: 2400, // refined once host is chosen; floor baseline
		HostTotalVCPUs:       8,
		HostTotalRAMMB:       16384,
		HostTotalDiskGB:      160,
		AllocatedVCPUs:       vcpus,
		AllocatedRAMMB:       memGB * 1024,
		AllocatedDiskGB:      diskGB,
		Currency:             "USD",
	})
	quote, _ := Pricing.Quote(ctx, cogs, accounting.PricingContext{
		TenantID:              tenantDBID,
		TenantTier:            "pro",
		DriverKind:            providers.DriverLinuxNestedQEMU,
		OwnershipType:         providers.OwnershipLeasedNestedVPS,
		RegionSlug:            in.PreferredRegion,
		ProviderMarkupBps:     3500, // +35% platform wholesale margin
		TenantRetailMarkupBps: 2500, // +25% tenant retail margin
		EnforceCostFloor:      true,
	})
	vmSupplyCents := cogs.TotalSupplyCostCents
	vmRetailCents := quote.EndUserTotalCents
	if vmRetailCents <= 0 {
		vmRetailCents = int64(vcpus*400 + memGB*200 + diskGB*10)
	}

	// Atomically allocate an online+ready SupplyHost + free assignable IP
	alloc, err := db.AllocateHostAndIPForVM(ctx, Pools.Admin, db.CustomerVMOrderRequest{
		TenantID:           tenantDBID,
		WorkspaceID:        wsDBID,
		CreatedBy:          in.ActorUserID,
		Name:               in.Name,
		InstanceType:       in.InstanceType,
		Distro:             in.Distro,
		Version:            in.Version,
		SSHKey:             in.SSHPublicKey,
		VCPUs:              vcpus,
		MemoryGB:           memGB,
		DiskGB:             diskGB,
		PreferredRegion:    in.PreferredRegion,
		VMSupplyCostCents:  vmSupplyCents,
		VMRetailPriceCents: vmRetailCents,
	})
	if err != nil {
		return CustomerVMPurchaseResult{}, err
	}

	drv, drvErr := Drivers.Compute(providers.DriverKind(alloc.Host.DriverKind))
	if drvErr != nil {
		_ = db.FinalizeCustomerVMProvisioning(ctx, Pools.Admin, alloc.VMResourceID, "", drvErr)
		return CustomerVMPurchaseResult{}, drvErr
	}

	instRes, createErr := drv.CreateInstance(ctx, alloc.Host.ToProviderHost(), providers.InstanceSpec{
		ResourceID:   alloc.VMResourceID,
		Name:         in.Name,
		VCPUs:        vcpus,
		MemoryMB:     memGB * 1024,
		DiskGB:       diskGB,
		OSImage:      strings.ToLower(in.Distro) + "-" + in.Version,
		IPv4Address:  alloc.IP.IPAddress,
		CIDRPrefix:   alloc.IP.CIDRPrefix,
		Gateway:      alloc.IP.Gateway,
		MACAddress:   alloc.IP.MACAddress,
		SSHPublicKey: in.SSHPublicKey,
	})
	if createErr != nil {
		_ = db.FinalizeCustomerVMProvisioning(ctx, Pools.Admin, alloc.VMResourceID, "", createErr)
		_ = db.WriteSystemLog(ctx, Pools.Admin, "error", "vm-provisioner",
			fmt.Sprintf("Failed to provision VM %s on host %s: %v", in.Name, alloc.Host.Name, createErr),
			map[string]any{"vm_resource_id": alloc.VMResourceID, "host_id": alloc.Host.ID})
		return CustomerVMPurchaseResult{}, createErr
	}

	if err := db.FinalizeCustomerVMProvisioning(ctx, Pools.Admin, alloc.VMResourceID, instRes.ExternalID, nil); err != nil {
		log.Printf("vm-provisioner: finalize %d: %v", alloc.VMResourceID, err)
	}

	// Record Cost & Revenue events in the Cost-Centric Accounting MarginLedger
	now := time.Now().UTC()
	vmResID := alloc.VMResourceID
	hostID := alloc.Host.ID
	tID := tenantDBID
	wID := wsDBID
	_ = Ledger.RecordCost(ctx, accounting.CostEvent{
		ResourceID:   &vmResID,
		SupplyHostID: &hostID,
		TenantID:     &tID,
		WorkspaceID:  &wID,
		AmountCents:  vmSupplyCents + alloc.IP.MonthlyCostCents,
		Currency:     "USD",
		Description:  fmt.Sprintf("VM %s + IPv4 %s COGS on %s", in.Name, alloc.IP.IPAddress, alloc.Host.Name),
		OccurredAt:   now,
	})
	_ = Ledger.RecordRevenue(ctx, accounting.RevenueEvent{
		ResourceID:       &vmResID,
		TenantID:         tenantDBID,
		WorkspaceID:      &wID,
		GrossAmountCents: alloc.TotalRetailCents,
		Currency:         "USD",
		OccurredAt:       now,
	})

	_ = db.WriteSystemLog(ctx, Pools.Admin, "info", "vm-provisioner",
		fmt.Sprintf("Provisioned customer VM %s (%s) on %s with assigned IP %s",
			in.Name, instRes.ExternalID, alloc.Host.Name, alloc.IP.IPAddress),
		map[string]any{"vm_resource_id": alloc.VMResourceID, "ip": alloc.IP.IPAddress, "host": alloc.Host.Name})

	return CustomerVMPurchaseResult{
		VMResourceID:       alloc.VMResourceID,
		IPResourceID:       alloc.IPResourceID,
		ExternalID:         instRes.ExternalID,
		SupplyHostID:       alloc.Host.ID,
		SupplyHostName:     alloc.Host.Name,
		RegionSlug:         alloc.RegionSlug,
		AssignedIPv4:       alloc.IP.IPAddress,
		Status:             "running",
		MonthlyCostDisplay: alloc.MonthlyCostDisplay,
	}, nil
}

// TerminateCustomerVM destroys the guest instance via the host's ComputeDriver,
// soft-deletes the parent VM + child IP `resources`, and returns the IP back to
// `status = 'available'` in `supply_host_ips`.
func TerminateCustomerVM(ctx context.Context, vmResourceID int64) error {
	if Pools == nil {
		return nil
	}
	host, externalRef, err := db.ReleaseCustomerVMAndIP(ctx, Pools.Admin, vmResourceID)
	if err != nil {
		return err
	}
	if host != nil && externalRef != "" {
		if drv, dErr := Drivers.Compute(providers.DriverKind(host.DriverKind)); dErr == nil {
			if delErr := drv.DeleteInstance(ctx, host.ToProviderHost(), externalRef); delErr != nil {
				log.Printf("vm-provisioner: driver DeleteInstance(%s on %s): %v", externalRef, host.Name, delErr)
			}
		}
	}
	return nil
}
