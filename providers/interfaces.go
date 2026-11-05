package providers

import "context"

// ComputeDriver abstracts VPS / virtual-machine lifecycle and host bootstrap
// across bare-metal QEMU/libvirt, nested Linux VPS hosts, Proxmox VE clusters,
// VMware vSphere/ESXi, and upstream cloud reseller APIs.
type ComputeDriver interface {
	Kind() DriverKind

	// ValidateConnection checks credentials and reachability of a supply host.
	ValidateConnection(ctx context.Context, host SupplyHost) error

	// InspectCapacity probes CPU, memory, disk, and virtualization capabilities.
	InspectCapacity(ctx context.Context, host SupplyHost) (HostCapacity, error)

	// BootstrapHost installs and configures the virtualization stack on a raw host
	// so it is ready to accept customer instances in inventory.
	BootstrapHost(ctx context.Context, host SupplyHost) (BootstrapReport, error)

	// Instance lifecycle operations:
	CreateInstance(ctx context.Context, host SupplyHost, spec InstanceSpec) (InstanceResult, error)
	StartInstance(ctx context.Context, host SupplyHost, externalID string) error
	StopInstance(ctx context.Context, host SupplyHost, externalID string) error
	RebootInstance(ctx context.Context, host SupplyHost, externalID string) error
	ResizeInstance(ctx context.Context, host SupplyHost, externalID string, vcpus, memoryMB, diskGB int) error
	SnapshotInstance(ctx context.Context, host SupplyHost, externalID, snapshotName string) (string, error)
	DeleteInstance(ctx context.Context, host SupplyHost, externalID string) error
	ConsoleInfo(ctx context.Context, host SupplyHost, externalID string) (string, error)
}

// HostingDriver abstracts non-VM compute products such as shared web hosting
// panels (cPanel/DirectAdmin/FPM-jail) and PaaS application runtimes.
type HostingDriver interface {
	Kind() DriverKind
	CreateAccount(ctx context.Context, host SupplyHost, spec HostingAccountSpec) (HostingAccountResult, error)
	BindDomain(ctx context.Context, host SupplyHost, externalID, domain string) error
	IssueSSL(ctx context.Context, host SupplyHost, externalID, domain string) error
	SetQuota(ctx context.Context, host SupplyHost, externalID string, diskMB, bandwidthGB int) error
	SuspendAccount(ctx context.Context, host SupplyHost, externalID string) error
	DeleteAccount(ctx context.Context, host SupplyHost, externalID string) error
}

// NetworkStorageDriver abstracts routable IP allocation/binding and attachable
// block volumes across local ZFS/LVM/Ceph pools or upstream cloud APIs.
type NetworkStorageDriver interface {
	Kind() DriverKind
	AllocateIP(ctx context.Context, host SupplyHost, spec IPAllocationSpec) (AllocatedIP, error)
	ReleaseIP(ctx context.Context, host SupplyHost, ipAddress string) error
	BindIP(ctx context.Context, host SupplyHost, ipAddress, instanceExternalID string) error
	CreateVolume(ctx context.Context, host SupplyHost, spec VolumeSpec) (VolumeResult, error)
	AttachVolume(ctx context.Context, host SupplyHost, volumeExternalID, instanceExternalID string) (string, error)
	DetachVolume(ctx context.Context, host SupplyHost, volumeExternalID, instanceExternalID string) error
	DeleteVolume(ctx context.Context, host SupplyHost, volumeExternalID string) error
}
