// Package providers defines the infrastructure supply-layer abstractions:
// host inventory models, ownership classifications, and pluggable hypervisor,
// shared-hosting, PaaS, IPAM, and block-storage driver contracts.
package providers

import (
	"errors"
	"time"
)

// ErrNotImplemented is returned by placeholder drivers whose wire protocol
// has not yet been implemented.
var ErrNotImplemented = errors.New("providers: driver operation not yet implemented")

// ErrDriverNotFound is returned when no driver is registered for a DriverKind.
var ErrDriverNotFound = errors.New("providers: driver not registered")

// OwnershipType classifies how the platform sources an underlying supply node.
type OwnershipType string

const (
	OwnershipOwnedBareMetal   OwnershipType = "owned_bare_metal"
	OwnershipColocated        OwnershipType = "colocated"
	OwnershipLeasedNestedVPS  OwnershipType = "leased_nested_vps"
	OwnershipCloudResellerAPI OwnershipType = "cloud_reseller_api"
	OwnershipNestedCluster    OwnershipType = "nested_cluster"
)

// DriverKind identifies which concrete driver speaks to a SupplyHost.
type DriverKind string

const (
	DriverLinuxNestedQEMU  DriverKind = "linux_nested_qemu"
	DriverQEMULibvirt      DriverKind = "qemu_libvirt"
	DriverProxmox          DriverKind = "proxmox"
	DriverVMwareVSphere    DriverKind = "vmware_vsphere"
	DriverUpstreamReseller DriverKind = "upstream_reseller"
	DriverSharedWeb        DriverKind = "shared_web"
	DriverPaaSApp          DriverKind = "paas_app"
	DriverIPAM             DriverKind = "ipam"
	DriverBlockStorage     DriverKind = "block_storage"
)

// VirtStackStatus tracks whether a raw host has had its virtualization stack
// installed and verified by the platform's auto-provisioning tool.
type VirtStackStatus string

const (
	VirtStackUnprovisioned VirtStackStatus = "unprovisioned"
	VirtStackProvisioning  VirtStackStatus = "provisioning"
	VirtStackReady         VirtStackStatus = "ready"
	VirtStackFailed        VirtStackStatus = "failed"
)

// InventoryStatus tracks whether a supply host can receive new customer workloads.
type InventoryStatus string

const (
	InventoryPending     InventoryStatus = "pending"
	InventoryOnline      InventoryStatus = "online"
	InventoryMaintenance InventoryStatus = "maintenance"
	InventoryOffline     InventoryStatus = "offline"
)

// HostAuthMethod specifies how the driver authenticates to the supply host.
type HostAuthMethod string

const (
	AuthSSHPrivateKey HostAuthMethod = "ssh_private_key"
	AuthSSHPassword   HostAuthMethod = "ssh_password"
	AuthAPIToken      HostAuthMethod = "api_token"
)

// SupplyHost represents one physical, virtual/nested, or API-backed supply node
// registered in the platform's infrastructure inventory.
type SupplyHost struct {
	ID               int64
	Name             string
	HostnameOrIP     string
	SSHPort          int
	SSHUser          string
	AuthMethod       HostAuthMethod
	Credential       string // SSH private key PEM, SSH password, or upstream API token
	APIEndpoint      string // Optional REST/SOAP URL for Proxmox / vSphere / Reseller APIs
	DriverKind       DriverKind
	OwnershipType    OwnershipType
	RegionSlug       string
	CPUCores         int
	RAMMB            int
	DiskGB           int
	MonthlyCostCents int64
	Currency         string
	VirtStackStatus  VirtStackStatus
	InventoryStatus  InventoryStatus
	BridgeInterface  string
	StoragePoolPath  string
	Metadata         map[string]string
}

// HostCapacity captures hardware & virtualization capabilities discovered by probing a host.
type HostCapacity struct {
	CPUCores          int
	RAMMB             int
	DiskGB            int
	KVMAvailable      bool
	CPUFlagsVirt      bool
	OSPrettyName      string
	KernelVersion     string
	HypervisorVersion string
}

// BootstrapReport summarizes an auto-provisioning run that installs and verifies
// the virtualization stack on a newly added SupplyHost.
type BootstrapReport struct {
	Success        bool
	DriverKind     DriverKind
	StepsCompleted []string
	LogOutput      string
	Capacity       HostCapacity
	StartedAt      time.Time
	FinishedAt     time.Time
}

// InstanceSpec carries everything a ComputeDriver needs to create a customer VPS/VM.
type InstanceSpec struct {
	ResourceID        int64
	Name              string
	VCPUs             int
	MemoryMB          int
	DiskGB            int
	OSImage           string
	IPv4Address       string
	CIDRPrefix        int
	Gateway           string
	DNSServers        []string
	MACAddress        string
	SSHPublicKey      string
	RootPassword      string
	CloudInitUserData string
}

// InstanceResult describes the provisioned instance on the underlying provider.
type InstanceResult struct {
	ExternalID      string
	Status          string
	AssignedIPv4    string
	ConsoleEndpoint string
	CommandLog      string
}

// HostingAccountSpec describes a shared web hosting or PaaS application target.
type HostingAccountSpec struct {
	ResourceID   int64
	Domain       string
	Runtime      string
	DiskQuotaMB  int
	BandwidthGB  int
	MemoryLimitM int
	GitRepoURL   string
	GitBranch    string
}

// HostingAccountResult describes a provisioned shared-hosting or PaaS account.
type HostingAccountResult struct {
	ExternalID  string
	Status      string
	PreviewURL  string
	ServerIPv4  string
	DetailsJSON string
}

// IPAllocationSpec requests an assignable IP from an IPAM or upstream pool.
type IPAllocationSpec struct {
	SupplyHostID int64
	RegionSlug   string
	IPVersion    int // 4 or 6
	ResourceID   int64
}

// AllocatedIP represents a routable IP leased/assigned to a customer resource.
type AllocatedIP struct {
	ID         int64
	Address    string
	CIDRPrefix int
	Gateway    string
	MACAddress string
	IPVersion  int
}

// VolumeSpec describes a block storage volume request.
type VolumeSpec struct {
	ResourceID int64
	Name       string
	SizeGB     int
	Tier       string // "nvme", "ssd", "hdd"
}

// VolumeResult describes a provisioned block volume.
type VolumeResult struct {
	ExternalID string
	DevicePath string
	Status     string
}
