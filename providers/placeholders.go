package providers

import (
	"context"
	"fmt"
)

// placeholderComputeDriver implements ComputeDriver by returning descriptive
// ErrNotImplemented errors until each concrete hypervisor/API backend is wired.
type placeholderComputeDriver struct {
	kind  DriverKind
	label string
}

func (p *placeholderComputeDriver) Kind() DriverKind { return p.kind }

func (p *placeholderComputeDriver) notImpl(op string) error {
	return fmt.Errorf("%w: %s (%s).%s", ErrNotImplemented, p.label, p.kind, op)
}

func (p *placeholderComputeDriver) ValidateConnection(context.Context, SupplyHost) error {
	return p.notImpl("ValidateConnection")
}
func (p *placeholderComputeDriver) InspectCapacity(context.Context, SupplyHost) (HostCapacity, error) {
	return HostCapacity{}, p.notImpl("InspectCapacity")
}
func (p *placeholderComputeDriver) BootstrapHost(context.Context, SupplyHost) (BootstrapReport, error) {
	return BootstrapReport{}, p.notImpl("BootstrapHost")
}
func (p *placeholderComputeDriver) CreateInstance(context.Context, SupplyHost, InstanceSpec) (InstanceResult, error) {
	return InstanceResult{}, p.notImpl("CreateInstance")
}
func (p *placeholderComputeDriver) StartInstance(context.Context, SupplyHost, string) error {
	return p.notImpl("StartInstance")
}
func (p *placeholderComputeDriver) StopInstance(context.Context, SupplyHost, string) error {
	return p.notImpl("StopInstance")
}
func (p *placeholderComputeDriver) RebootInstance(context.Context, SupplyHost, string) error {
	return p.notImpl("RebootInstance")
}
func (p *placeholderComputeDriver) ResizeInstance(context.Context, SupplyHost, string, int, int, int) error {
	return p.notImpl("ResizeInstance")
}
func (p *placeholderComputeDriver) SnapshotInstance(context.Context, SupplyHost, string, string) (string, error) {
	return "", p.notImpl("SnapshotInstance")
}
func (p *placeholderComputeDriver) DeleteInstance(context.Context, SupplyHost, string) error {
	return p.notImpl("DeleteInstance")
}
func (p *placeholderComputeDriver) ConsoleInfo(context.Context, SupplyHost, string) (string, error) {
	return "", p.notImpl("ConsoleInfo")
}

// LinuxNestedQEMUPlaceholderDriver is the initial placeholder for a Linux nested VPS
// running QEMU/KVM + libvirt.
type LinuxNestedQEMUPlaceholderDriver struct {
	placeholderComputeDriver
}

func (d *LinuxNestedQEMUPlaceholderDriver) Kind() DriverKind { return DriverLinuxNestedQEMU }

// QEMULibvirtDriver is the placeholder driver for dedicated bare-metal KVM/QEMU hosts.
type QEMULibvirtDriver struct{ placeholderComputeDriver }

func (d *QEMULibvirtDriver) Kind() DriverKind { return DriverQEMULibvirt }

// ProxmoxDriver is the placeholder driver for Proxmox VE clusters (REST API).
type ProxmoxDriver struct{ placeholderComputeDriver }

func (d *ProxmoxDriver) Kind() DriverKind { return DriverProxmox }

// VMwareVSphereDriver is the placeholder driver for VMware vCenter / ESXi (including nested ESXi).
type VMwareVSphereDriver struct{ placeholderComputeDriver }

func (d *VMwareVSphereDriver) Kind() DriverKind { return DriverVMwareVSphere }

// UpstreamResellerDriver is the placeholder driver for reselling upstream cloud APIs
// (Hetzner, BuyVM, DigitalOcean, OpenStack, AWS, etc.).
type UpstreamResellerDriver struct{ placeholderComputeDriver }

func (d *UpstreamResellerDriver) Kind() DriverKind { return DriverUpstreamReseller }

// ── Hosting Drivers (Shared Web & PaaS) ──────────────────────────────────────

type placeholderHostingDriver struct {
	kind  DriverKind
	label string
}

func (p *placeholderHostingDriver) Kind() DriverKind { return p.kind }
func (p *placeholderHostingDriver) notImpl(op string) error {
	return fmt.Errorf("%w: %s (%s).%s", ErrNotImplemented, p.label, p.kind, op)
}
func (p *placeholderHostingDriver) CreateAccount(context.Context, SupplyHost, HostingAccountSpec) (HostingAccountResult, error) {
	return HostingAccountResult{}, p.notImpl("CreateAccount")
}
func (p *placeholderHostingDriver) BindDomain(context.Context, SupplyHost, string, string) error {
	return p.notImpl("BindDomain")
}
func (p *placeholderHostingDriver) IssueSSL(context.Context, SupplyHost, string, string) error {
	return p.notImpl("IssueSSL")
}
func (p *placeholderHostingDriver) SetQuota(context.Context, SupplyHost, string, int, int) error {
	return p.notImpl("SetQuota")
}
func (p *placeholderHostingDriver) SuspendAccount(context.Context, SupplyHost, string) error {
	return p.notImpl("SuspendAccount")
}
func (p *placeholderHostingDriver) DeleteAccount(context.Context, SupplyHost, string) error {
	return p.notImpl("DeleteAccount")
}

// SharedWebDriver is the placeholder driver for shared web hosting environments.
type SharedWebDriver struct{ placeholderHostingDriver }

func (d *SharedWebDriver) Kind() DriverKind { return DriverSharedWeb }

// PaaSAppDriver is the placeholder driver for git-driven PaaS / container apps.
type PaaSAppDriver struct{ placeholderHostingDriver }

func (d *PaaSAppDriver) Kind() DriverKind { return DriverPaaSApp }

// ── Network & Storage Drivers (IPAM & Block Volumes) ─────────────────────────

type placeholderNetStorageDriver struct {
	kind  DriverKind
	label string
}

func (p *placeholderNetStorageDriver) Kind() DriverKind { return p.kind }
func (p *placeholderNetStorageDriver) notImpl(op string) error {
	return fmt.Errorf("%w: %s (%s).%s", ErrNotImplemented, p.label, p.kind, op)
}
func (p *placeholderNetStorageDriver) AllocateIP(context.Context, SupplyHost, IPAllocationSpec) (AllocatedIP, error) {
	return AllocatedIP{}, p.notImpl("AllocateIP")
}
func (p *placeholderNetStorageDriver) ReleaseIP(context.Context, SupplyHost, string) error {
	return p.notImpl("ReleaseIP")
}
func (p *placeholderNetStorageDriver) BindIP(context.Context, SupplyHost, string, string) error {
	return p.notImpl("BindIP")
}
func (p *placeholderNetStorageDriver) CreateVolume(context.Context, SupplyHost, VolumeSpec) (VolumeResult, error) {
	return VolumeResult{}, p.notImpl("CreateVolume")
}
func (p *placeholderNetStorageDriver) AttachVolume(context.Context, SupplyHost, string, string) (string, error) {
	return "", p.notImpl("AttachVolume")
}
func (p *placeholderNetStorageDriver) DetachVolume(context.Context, SupplyHost, string, string) error {
	return p.notImpl("DetachVolume")
}
func (p *placeholderNetStorageDriver) DeleteVolume(context.Context, SupplyHost, string) error {
	return p.notImpl("DeleteVolume")
}

// IPAMDriver is the placeholder driver for automated subnet/VLAN/BGP IP management.
type IPAMDriver struct{ placeholderNetStorageDriver }

func (d *IPAMDriver) Kind() DriverKind { return DriverIPAM }

// BlockStorageDriver is the placeholder driver for Ceph/ZFS/LVM/cloud block volumes.
type BlockStorageDriver struct{ placeholderNetStorageDriver }

func (d *BlockStorageDriver) Kind() DriverKind { return DriverBlockStorage }
