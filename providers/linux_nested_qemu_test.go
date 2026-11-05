package providers

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type recordedSSH struct {
	scripts []string
	stdout  string
	stderr  string
	err     error
}

func (r *recordedSSH) RunScript(_ context.Context, _ SupplyHost, script string) (string, string, error) {
	r.scripts = append(r.scripts, script)
	return r.stdout, r.stderr, r.err
}

func TestDefaultRegistryExposesConcreteLinuxNestedDriver(t *testing.T) {
	reg := NewDefaultRegistry()

	drv, err := reg.Compute(DriverLinuxNestedQEMU)
	if err != nil {
		t.Fatalf("expected DriverLinuxNestedQEMU registered, got %v", err)
	}
	if _, ok := drv.(*LinuxNestedQEMUDriver); !ok {
		t.Fatalf("expected *LinuxNestedQEMUDriver, got %T", drv)
	}

	// Verify placeholder drivers remain registered and return ErrNotImplemented.
	proxmox, err := reg.Compute(DriverProxmox)
	if err != nil {
		t.Fatalf("expected DriverProxmox registered, got %v", err)
	}
	if err := proxmox.ValidateConnection(context.Background(), SupplyHost{}); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented from Proxmox placeholder, got %v", err)
	}
}

func TestLinuxNestedQEMUDriver_BootstrapAndLifecycle(t *testing.T) {
	mockSSH := &recordedSSH{
		stdout: strings.Join([]string{
			"==> [1/5] Preflight & Nested Virtualization Inspection",
			"Hardware nested KVM device (/dev/kvm) is active.",
			"PROBE_CPU_CORES=8",
			"PROBE_RAM_MB=16384",
			"PROBE_DISK_GB=180",
			"PROBE_CPU_FLAGS_VIRT=true",
			"PROBE_KVM_AVAILABLE=true",
			"PROBE_OS_RELEASE=Ubuntu 24.04.1 LTS",
			"PROBE_KERNEL=6.8.0-45-generic",
			"PROBE_HV_VER=10.0.0",
			"BOOTSTRAP_COMPLETE",
		}, "\n"),
	}

	drv := &LinuxNestedQEMUDriver{SSH: mockSSH}
	host := SupplyHost{
		ID:              1,
		Name:            "fra-nested-01",
		HostnameOrIP:    "203.0.113.10",
		SSHPort:         22,
		SSHUser:         "root",
		AuthMethod:      AuthSSHPrivateKey,
		DriverKind:      DriverLinuxNestedQEMU,
		OwnershipType:   OwnershipLeasedNestedVPS,
		BridgeInterface: "virbr0",
		StoragePoolPath: "/var/lib/libvirt/images/platform",
	}

	report, err := drv.BootstrapHost(context.Background(), host)
	if err != nil {
		t.Fatalf("BootstrapHost failed: %v", err)
	}
	if !report.Success {
		t.Fatalf("expected report.Success = true")
	}
	if report.Capacity.CPUCores != 8 || report.Capacity.RAMMB != 16384 || !report.Capacity.KVMAvailable {
		t.Fatalf("unexpected parsed capacity: %+v", report.Capacity)
	}
	if len(mockSSH.scripts) != 1 || !strings.Contains(mockSSH.scripts[0], "qemu-system-x86") {
		t.Fatalf("expected virtualization stack install commands in bootstrap script")
	}

	// Provision customer VM with manually-assigned static IP from host pool.
	mockSSH.stdout = "VM_CREATED=cust-vm-42\n"
	res, err := drv.CreateInstance(context.Background(), host, InstanceSpec{
		ResourceID:   42,
		Name:         "customer-web-01",
		VCPUs:        2,
		MemoryMB:     4096,
		DiskGB:       40,
		IPv4Address:  "198.51.100.25",
		CIDRPrefix:   24,
		Gateway:      "198.51.100.1",
		MACAddress:   "52:54:00:ab:cd:ef",
		SSHPublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITestKey user@example",
	})
	if err != nil {
		t.Fatalf("CreateInstance failed: %v", err)
	}
	if res.ExternalID != "cust-vm-42" || res.AssignedIPv4 != "198.51.100.25" {
		t.Fatalf("unexpected InstanceResult: %+v", res)
	}
	createScript := mockSSH.scripts[1]
	for _, mustContain := range []string{
		"198.51.100.25/24",
		"gateway4: 198.51.100.1",
		",mac=52:54:00:ab:cd:ef",
		"ip route replace \"$IPV4/32\" dev \"$BRIDGE\"",
		"virt-install",
	} {
		if !strings.Contains(createScript, mustContain) {
			t.Fatalf("CreateInstance script missing %q", mustContain)
		}
	}

	// Delete VM
	if err := drv.DeleteInstance(context.Background(), host, "cust-vm-42"); err != nil {
		t.Fatalf("DeleteInstance failed: %v", err)
	}
}
