package providers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var safeUserRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
var safeHostRe = regexp.MustCompile(`^[a-zA-Z0-9.:_-]+$`)
var safeIfaceRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// SSHExecutor abstracts remote command execution over SSH so production runs
// OpenSSH (`ssh`) and automated unit tests can verify every provisioning script
// deterministically without a live remote hypervisor.
type SSHExecutor interface {
	RunScript(ctx context.Context, host SupplyHost, script string) (stdout string, stderr string, err error)
}

// OpenSSHExecutor executes scripts over the system `ssh` client using temporary,
// permission-locked credential files that are wiped immediately on return.
type OpenSSHExecutor struct {
	ConnectTimeoutSec int
}

func (e *OpenSSHExecutor) RunScript(ctx context.Context, host SupplyHost, script string) (string, string, error) {
	targetHost := strings.TrimSpace(host.HostnameOrIP)
	if targetHost == "" || strings.HasPrefix(targetHost, "-") || !safeHostRe.MatchString(targetHost) {
		return "", "", fmt.Errorf("invalid host address %q", host.HostnameOrIP)
	}
	sshUser := strings.TrimSpace(host.SSHUser)
	if sshUser == "" {
		sshUser = "root"
	}
	if !safeUserRe.MatchString(sshUser) {
		return "", "", fmt.Errorf("invalid ssh user %q", sshUser)
	}
	port := host.SSHPort
	if port <= 0 || port > 65535 {
		port = 22
	}
	connTimeout := e.ConnectTimeoutSec
	if connTimeout <= 0 {
		connTimeout = 12
	}

	tmpDir, err := os.MkdirTemp("", "platform-ssh-*")
	if err != nil {
		return "", "", fmt.Errorf("create temp ssh dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	args := []string{
		"-p", strconv.Itoa(port),
		"-o", fmt.Sprintf("ConnectTimeout=%d", connTimeout),
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "UserKnownHostsFile=" + filepath.Join(tmpDir, "known_hosts"),
		"-o", "ServerAliveInterval=10",
		"-o", "ServerAliveCountMax=3",
	}

	env := os.Environ()

	switch host.AuthMethod {
	case AuthSSHPassword:
		if host.Credential == "" {
			return "", "", errors.New("ssh password credential is empty")
		}
		askpassPath := filepath.Join(tmpDir, "askpass.sh")
		askpassScript := "#!/bin/sh\nprintf '%s\\n' \"$PLATFORM_SSH_PW\"\n"
		if err := os.WriteFile(askpassPath, []byte(askpassScript), 0700); err != nil {
			return "", "", fmt.Errorf("prepare askpass helper: %w", err)
		}
		args = append(args,
			"-o", "PubkeyAuthentication=no",
			"-o", "PreferredAuthentications=password,keyboard-interactive",
		)
		env = append(env,
			"PLATFORM_SSH_PW="+host.Credential,
			"SSH_ASKPASS="+askpassPath,
			"SSH_ASKPASS_REQUIRE=force",
			"DISPLAY=platform:0",
		)

	default: // AuthSSHPrivateKey
		args = append(args, "-o", "BatchMode=yes")
		if strings.TrimSpace(host.Credential) != "" {
			keyPath := filepath.Join(tmpDir, "id_key")
			pem := strings.TrimSpace(host.Credential) + "\n"
			if err := os.WriteFile(keyPath, []byte(pem), 0600); err != nil {
				return "", "", fmt.Errorf("write ssh private key: %w", err)
			}
			args = append(args, "-i", keyPath, "-o", "IdentitiesOnly=yes")
		}
	}

	args = append(args, fmt.Sprintf("%s@%s", sshUser, targetHost), "bash -s")

	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Env = env
	cmd.Stdin = strings.NewReader(script)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	return stdout.String(), stderr.String(), err
}

// LinuxNestedQEMUDriver is the concrete ComputeDriver for onboarding a Linux
// Nested VPS, installing the QEMU/KVM + libvirt virtualization stack via SSH,
// and provisioning customer VMs bound to manually-managed assignable IPs.
type LinuxNestedQEMUDriver struct {
	SSH SSHExecutor
}

// NewLinuxNestedQEMUDriver constructs a production LinuxNestedQEMUDriver backed
// by OpenSSHExecutor.
func NewLinuxNestedQEMUDriver() *LinuxNestedQEMUDriver {
	return &LinuxNestedQEMUDriver{
		SSH: &OpenSSHExecutor{ConnectTimeoutSec: 12},
	}
}

func (d *LinuxNestedQEMUDriver) sshExec() SSHExecutor {
	if d != nil && d.SSH != nil {
		return d.SSH
	}
	return &OpenSSHExecutor{ConnectTimeoutSec: 12}
}

func (d *LinuxNestedQEMUDriver) Kind() DriverKind {
	return DriverLinuxNestedQEMU
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func normalizeBridge(b string) string {
	b = strings.TrimSpace(b)
	if b == "" || !safeIfaceRe.MatchString(b) {
		return "virbr0"
	}
	return b
}

func normalizeStoragePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || !strings.HasPrefix(p, "/") || strings.ContainsAny(p, " \t\r\n;|&$`") {
		return "/var/lib/libvirt/images/platform"
	}
	return filepath.Clean(p)
}

// ValidateConnection verifies SSH reachability, Linux OS, and root/passwordless sudo privileges.
func (d *LinuxNestedQEMUDriver) ValidateConnection(ctx context.Context, host SupplyHost) error {
	script := `set -euo pipefail
OS_NAME="$(uname -s)"
if [ "$OS_NAME" != "Linux" ]; then
  echo "unsupported OS: $OS_NAME" >&2
  exit 1
fi
if [ "$(id -u)" -ne 0 ] && ! sudo -n true 2>/dev/null; then
  echo "root or passwordless sudo is required on target Linux Nested VPS" >&2
  exit 1
fi
echo "SSH_OK"
`
	stdout, stderr, err := d.sshExec().RunScript(ctx, host, script)
	if err != nil {
		msg := strings.TrimSpace(stderr)
		if msg == "" {
			msg = strings.TrimSpace(stdout)
		}
		return fmt.Errorf("ssh validation failed (%v): %s", err, msg)
	}
	if !strings.Contains(stdout, "SSH_OK") {
		return errors.New("unexpected response during ssh validation")
	}
	return nil
}

const capacityProbeSnippet = `
CPU_CORES="$(nproc 2>/dev/null || echo 1)"
RAM_MB="$(awk '/MemTotal:/ {printf "%d", $2/1024}' /proc/meminfo 2>/dev/null || echo 0)"
DISK_GB="$(df -BG --output=avail /var/lib 2>/dev/null | tail -n1 | tr -dc '0-9' || true)"
if [ -z "$DISK_GB" ]; then
  DISK_GB="$(df -Pk / | awk 'NR==2 {printf "%d", $4/1048576}' 2>/dev/null || echo 0)"
fi
CPU_FLAGS_VIRT="false"
if grep -Eq '(vmx|svm)' /proc/cpuinfo 2>/dev/null; then
  CPU_FLAGS_VIRT="true"
fi
KVM_AVAILABLE="false"
if [ -e /dev/kvm ]; then
  KVM_AVAILABLE="true"
fi
OS_RELEASE="Linux"
if [ -r /etc/os-release ]; then
  OS_RELEASE="$(. /etc/os-release && printf '%s' "${PRETTY_NAME:-Linux}")"
fi
KERNEL_VER="$(uname -r 2>/dev/null || echo unknown)"
HV_VER="$(virsh --version 2>/dev/null | head -n1 || qemu-system-x86_64 --version 2>/dev/null | head -n1 || echo 'uninstalled')"
printf 'PROBE_CPU_CORES=%s\n' "$CPU_CORES"
printf 'PROBE_RAM_MB=%s\n' "$RAM_MB"
printf 'PROBE_DISK_GB=%s\n' "$DISK_GB"
printf 'PROBE_CPU_FLAGS_VIRT=%s\n' "$CPU_FLAGS_VIRT"
printf 'PROBE_KVM_AVAILABLE=%s\n' "$KVM_AVAILABLE"
printf 'PROBE_OS_RELEASE=%s\n' "$OS_RELEASE"
printf 'PROBE_KERNEL=%s\n' "$KERNEL_VER"
printf 'PROBE_HV_VER=%s\n' "$HV_VER"
`

func parseCapacityProbe(output string) HostCapacity {
	var cap HostCapacity
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch k {
		case "PROBE_CPU_CORES":
			cap.CPUCores, _ = strconv.Atoi(v)
		case "PROBE_RAM_MB":
			cap.RAMMB, _ = strconv.Atoi(v)
		case "PROBE_DISK_GB":
			cap.DiskGB, _ = strconv.Atoi(v)
		case "PROBE_CPU_FLAGS_VIRT":
			cap.CPUFlagsVirt = v == "true"
		case "PROBE_KVM_AVAILABLE":
			cap.KVMAvailable = v == "true"
		case "PROBE_OS_RELEASE":
			cap.OSPrettyName = v
		case "PROBE_KERNEL":
			cap.KernelVersion = v
		case "PROBE_HV_VER":
			cap.HypervisorVersion = v
		}
	}
	return cap
}

// InspectCapacity probes CPU, RAM, disk, nested virtualization flags, /dev/kvm,
// OS release, and installed hypervisor version on the remote Linux Nested VPS.
func (d *LinuxNestedQEMUDriver) InspectCapacity(ctx context.Context, host SupplyHost) (HostCapacity, error) {
	script := "set -euo pipefail\n" + capacityProbeSnippet
	stdout, stderr, err := d.sshExec().RunScript(ctx, host, script)
	if err != nil {
		return HostCapacity{}, fmt.Errorf("inspect capacity failed (%v): %s", err, strings.TrimSpace(stderr))
	}
	return parseCapacityProbe(stdout), nil
}

// BootstrapHost runs the idempotent auto-provisioning tool on a newly added
// Linux Nested VPS to install QEMU/KVM + libvirt + cloud-init tooling, configure
// IPv4 forwarding / proxy-ARP and the storage pool, and bring the server ready
// for customer VM scheduling.
func (d *LinuxNestedQEMUDriver) BootstrapHost(ctx context.Context, host SupplyHost) (BootstrapReport, error) {
	started := time.Now().UTC()
	storagePath := normalizeStoragePath(host.StoragePoolPath)
	bridge := normalizeBridge(host.BridgeInterface)

	script := fmt.Sprintf(`set -euo pipefail
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  SUDO="sudo -n"
fi

echo "==> [1/5] Preflight & Nested Virtualization Inspection"
uname -a
$SUDO modprobe kvm 2>/dev/null || true
$SUDO modprobe kvm_intel 2>/dev/null || $SUDO modprobe kvm_amd 2>/dev/null || true
if [ -e /dev/kvm ]; then
  echo "Hardware nested KVM device (/dev/kvm) is active."
else
  echo "Notice: /dev/kvm not exposed by upstream VPS; QEMU TCG emulation fallback will be used."
fi

echo "==> [2/5] Installing QEMU/KVM + libvirt + cloud-init Virtualization Stack"
if command -v apt-get >/dev/null 2>&1; then
  export DEBIAN_FRONTEND=noninteractive
  $SUDO apt-get update -qq
  $SUDO apt-get install -y -qq \
    qemu-system-x86 qemu-utils libvirt-daemon-system libvirt-clients \
    virtinst bridge-utils cloud-image-utils genisoimage curl iproute2
elif command -v dnf >/dev/null 2>&1; then
  $SUDO dnf install -y \
    qemu-kvm qemu-img libvirt libvirt-client virt-install \
    genisoimage curl iproute
elif command -v yum >/dev/null 2>&1; then
  $SUDO yum install -y \
    qemu-kvm qemu-img libvirt libvirt-client virt-install \
    genisoimage curl iproute
else
  echo "Unsupported Linux package manager (expected apt-get, dnf, or yum)" >&2
  exit 1
fi

echo "==> [3/5] Enabling libvirtd & Configuring Routed Bridge Networking (%s)"
$SUDO systemctl enable --now libvirtd 2>/dev/null || $SUDO systemctl start libvirtd 2>/dev/null || true
$SUDO virsh net-autostart default 2>/dev/null || true
$SUDO virsh net-start default 2>/dev/null || true
$SUDO sysctl -w net.ipv4.ip_forward=1 >/dev/null 2>&1 || true
$SUDO sysctl -w net.ipv4.conf.all.proxy_arp=1 >/dev/null 2>&1 || true
if [ -d /etc/sysctl.d ]; then
  printf 'net.ipv4.ip_forward = 1\nnet.ipv4.conf.all.proxy_arp = 1\n' | $SUDO tee /etc/sysctl.d/99-platform-virt.conf >/dev/null
fi

echo "==> [4/5] Initializing VM Image Storage Pool (%s)"
$SUDO mkdir -p %s %s/base
$SUDO chmod 755 %s

echo "==> [5/5] Verifying Hypervisor Readiness & Probing Host Capacity"
virsh --version || qemu-system-x86_64 --version | head -n1
%s
echo "BOOTSTRAP_COMPLETE"
`, bridge, storagePath, shellQuote(storagePath), shellQuote(storagePath), shellQuote(storagePath), capacityProbeSnippet)

	stdout, stderr, err := d.sshExec().RunScript(ctx, host, script)
	combinedLog := strings.TrimSpace(stdout)
	if strings.TrimSpace(stderr) != "" {
		combinedLog += "\n[stderr]\n" + strings.TrimSpace(stderr)
	}

	report := BootstrapReport{
		DriverKind: DriverLinuxNestedQEMU,
		StartedAt:  started,
		FinishedAt: time.Now().UTC(),
		LogOutput:  combinedLog,
	}

	if err != nil {
		return report, fmt.Errorf("auto-provisioning failed (%v): %s", err, strings.TrimSpace(stderr))
	}
	if !strings.Contains(stdout, "BOOTSTRAP_COMPLETE") {
		return report, errors.New("auto-provisioning ended without completion marker")
	}

	report.Success = true
	report.StepsCompleted = []string{
		"preflight_nested_kvm_check",
		"install_qemu_libvirt_stack",
		"enable_libvirtd_and_routed_networking",
		"init_storage_pool",
		"verify_hypervisor_and_probe_capacity",
	}
	report.Capacity = parseCapacityProbe(stdout)
	return report, nil
}

// CreateInstance provisions a guest Linux VM on the Linux Nested VPS and binds
// the allocated static IP from the server's manual IP pool via NoCloud cloud-init
// and host `/32` bridge routing.
func (d *LinuxNestedQEMUDriver) CreateInstance(ctx context.Context, host SupplyHost, spec InstanceSpec) (InstanceResult, error) {
	if spec.ResourceID <= 0 {
		return InstanceResult{}, errors.New("CreateInstance requires positive ResourceID")
	}
	if spec.VCPUs <= 0 {
		spec.VCPUs = 1
	}
	if spec.MemoryMB <= 0 {
		spec.MemoryMB = 1024
	}
	if spec.DiskGB <= 0 {
		spec.DiskGB = 20
	}

	addr, err := netip.ParseAddr(strings.TrimSpace(spec.IPv4Address))
	if err != nil || !addr.Is4() {
		return InstanceResult{}, fmt.Errorf("valid IPv4 address required for nested VPS guest, got %q", spec.IPv4Address)
	}
	cidr := spec.CIDRPrefix
	if cidr <= 0 || cidr > 32 {
		cidr = 24
	}
	gw := strings.TrimSpace(spec.Gateway)
	if gw != "" {
		if gwAddr, gwErr := netip.ParseAddr(gw); gwErr != nil || !gwAddr.Is4() {
			return InstanceResult{}, fmt.Errorf("invalid gateway IPv4 %q", gw)
		}
	}

	macOpt := ""
	if strings.TrimSpace(spec.MACAddress) != "" {
		hw, macErr := net.ParseMAC(strings.TrimSpace(spec.MACAddress))
		if macErr != nil {
			return InstanceResult{}, fmt.Errorf("invalid MAC address %q: %w", spec.MACAddress, macErr)
		}
		macOpt = ",mac=" + hw.String()
	}

	domainName := fmt.Sprintf("cust-vm-%d", spec.ResourceID)
	storagePath := normalizeStoragePath(host.StoragePoolPath)
	vmDir := storagePath + "/" + domainName
	bridge := normalizeBridge(host.BridgeInterface)

	guestHostname := sanitizeHostname(spec.Name)
	if guestHostname == "" {
		guestHostname = domainName
	}

	sshKeyLine := ""
	if k := strings.TrimSpace(spec.SSHPublicKey); k != "" && !strings.ContainsAny(k, "\r\n") {
		sshKeyLine = fmt.Sprintf("ssh_authorized_keys:\n  - %s\n", shellSafeYAMLString(k))
	}

	gwYAML := ""
	if gw != "" {
		gwYAML = fmt.Sprintf("      gateway4: %s\n", gw)
	}

	userData := spec.CloudInitUserData
	if strings.TrimSpace(userData) == "" {
		userData = fmt.Sprintf("#cloud-config\nhostname: %s\nmanage_etc_hosts: true\n%s", guestHostname, sshKeyLine)
	}
	metaData := fmt.Sprintf("instance-id: %s\nlocal-hostname: %s\n", domainName, guestHostname)
	netConfig := fmt.Sprintf("version: 2\nethernets:\n  id0:\n    match:\n      name: \"e*\"\n    dhcp4: false\n    addresses:\n      - %s/%d\n%s    nameservers:\n      addresses: [1.1.1.1, 8.8.8.8]\n",
		addr.String(), cidr, gwYAML)

	script := fmt.Sprintf(`set -euo pipefail
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  SUDO="sudo -n"
fi

VM_DIR=%s
DOMAIN=%s
BRIDGE=%s
IPV4=%s

$SUDO mkdir -p "$VM_DIR"

BASE_IMG="$(ls %s/base/*.qcow2 %s/base/*.img 2>/dev/null | head -n1 || true)"
if [ -n "$BASE_IMG" ]; then
  $SUDO qemu-img create -f qcow2 -F qcow2 -b "$BASE_IMG" "$VM_DIR/disk.qcow2" %dG
else
  $SUDO qemu-img create -f qcow2 "$VM_DIR/disk.qcow2" %dG
fi

cat <<'EOF_USERDATA' | $SUDO tee "$VM_DIR/user-data" >/dev/null
%s
EOF_USERDATA

cat <<'EOF_METADATA' | $SUDO tee "$VM_DIR/meta-data" >/dev/null
%s
EOF_METADATA

cat <<'EOF_NETCFG' | $SUDO tee "$VM_DIR/network-config" >/dev/null
%s
EOF_NETCFG

if command -v cloud-localds >/dev/null 2>&1; then
  $SUDO cloud-localds -N "$VM_DIR/network-config" "$VM_DIR/seed.iso" "$VM_DIR/user-data" "$VM_DIR/meta-data"
elif command -v genisoimage >/dev/null 2>&1; then
  $SUDO genisoimage -output "$VM_DIR/seed.iso" -volid cidata -joliet -rock "$VM_DIR/user-data" "$VM_DIR/meta-data" "$VM_DIR/network-config" >/dev/null 2>&1
fi

VIRT_TYPE="qemu"
if [ -e /dev/kvm ]; then
  VIRT_TYPE="kvm"
fi

$SUDO ip route replace "$IPV4/32" dev "$BRIDGE" 2>/dev/null || true

$SUDO virt-install \
  --name "$DOMAIN" \
  --virt-type "$VIRT_TYPE" \
  --memory %d \
  --vcpus %d \
  --disk path="$VM_DIR/disk.qcow2",format=qcow2,bus=virtio \
  --disk path="$VM_DIR/seed.iso",device=cdrom \
  --os-variant generic \
  --network bridge="$BRIDGE",model=virtio%s \
  --graphics vnc,listen=127.0.0.1 \
  --import \
  --noautoconsole

echo "VM_CREATED=$DOMAIN"
`,
		shellQuote(vmDir),
		shellQuote(domainName),
		shellQuote(bridge),
		shellQuote(addr.String()),
		shellQuote(storagePath),
		shellQuote(storagePath),
		spec.DiskGB,
		spec.DiskGB,
		userData,
		metaData,
		netConfig,
		spec.MemoryMB,
		spec.VCPUs,
		macOpt,
	)

	stdout, stderr, err := d.sshExec().RunScript(ctx, host, script)
	combined := strings.TrimSpace(stdout)
	if strings.TrimSpace(stderr) != "" {
		combined += "\n" + strings.TrimSpace(stderr)
	}
	if err != nil {
		return InstanceResult{CommandLog: combined}, fmt.Errorf("CreateInstance %s failed (%v): %s", domainName, err, strings.TrimSpace(stderr))
	}

	return InstanceResult{
		ExternalID:   domainName,
		Status:       "active",
		AssignedIPv4: addr.String(),
		CommandLog:   combined,
	}, nil
}

func validateExternalDomain(externalID string) error {
	if externalID == "" || !safeUserRe.MatchString(externalID) {
		return fmt.Errorf("invalid libvirt domain identifier %q", externalID)
	}
	return nil
}

func (d *LinuxNestedQEMUDriver) StartInstance(ctx context.Context, host SupplyHost, externalID string) error {
	if err := validateExternalDomain(externalID); err != nil {
		return err
	}
	script := fmt.Sprintf("set -euo pipefail\n$([ \"$(id -u)\" -eq 0 ] || echo 'sudo -n') virsh start %s\n", shellQuote(externalID))
	_, stderr, err := d.sshExec().RunScript(ctx, host, script)
	if err != nil {
		return fmt.Errorf("virsh start %s: %v (%s)", externalID, err, strings.TrimSpace(stderr))
	}
	return nil
}

func (d *LinuxNestedQEMUDriver) StopInstance(ctx context.Context, host SupplyHost, externalID string) error {
	if err := validateExternalDomain(externalID); err != nil {
		return err
	}
	script := fmt.Sprintf("set -euo pipefail\nSUDO=\"\"\nif [ \"$(id -u)\" -ne 0 ]; then SUDO=\"sudo -n\"; fi\n$SUDO virsh shutdown %s 2>/dev/null || $SUDO virsh destroy %s\n",
		shellQuote(externalID), shellQuote(externalID))
	_, stderr, err := d.sshExec().RunScript(ctx, host, script)
	if err != nil {
		return fmt.Errorf("virsh stop %s: %v (%s)", externalID, err, strings.TrimSpace(stderr))
	}
	return nil
}

func (d *LinuxNestedQEMUDriver) RebootInstance(ctx context.Context, host SupplyHost, externalID string) error {
	if err := validateExternalDomain(externalID); err != nil {
		return err
	}
	script := fmt.Sprintf("set -euo pipefail\n$([ \"$(id -u)\" -eq 0 ] || echo 'sudo -n') virsh reboot %s\n", shellQuote(externalID))
	_, stderr, err := d.sshExec().RunScript(ctx, host, script)
	if err != nil {
		return fmt.Errorf("virsh reboot %s: %v (%s)", externalID, err, strings.TrimSpace(stderr))
	}
	return nil
}

func (d *LinuxNestedQEMUDriver) ResizeInstance(ctx context.Context, host SupplyHost, externalID string, vcpus, memoryMB, diskGB int) error {
	if err := validateExternalDomain(externalID); err != nil {
		return err
	}
	script := fmt.Sprintf(`set -euo pipefail
SUDO=""
if [ "$(id -u)" -ne 0 ]; then SUDO="sudo -n"; fi
$SUDO virsh setvcpus %s %d --config --maximum 2>/dev/null || true
$SUDO virsh setvcpus %s %d --config 2>/dev/null || true
$SUDO virsh setmaxmem %s %dM --config 2>/dev/null || true
$SUDO virsh setmem %s %dM --config 2>/dev/null || true
`, shellQuote(externalID), vcpus, shellQuote(externalID), vcpus, shellQuote(externalID), memoryMB, shellQuote(externalID), memoryMB)
	_, stderr, err := d.sshExec().RunScript(ctx, host, script)
	if err != nil {
		return fmt.Errorf("virsh resize %s: %v (%s)", externalID, err, strings.TrimSpace(stderr))
	}
	return nil
}

func (d *LinuxNestedQEMUDriver) SnapshotInstance(ctx context.Context, host SupplyHost, externalID, snapshotName string) (string, error) {
	if err := validateExternalDomain(externalID); err != nil {
		return "", err
	}
	snap := sanitizeHostname(snapshotName)
	if snap == "" {
		snap = fmt.Sprintf("snap-%d", time.Now().Unix())
	}
	script := fmt.Sprintf("set -euo pipefail\n$([ \"$(id -u)\" -eq 0 ] || echo 'sudo -n') virsh snapshot-create-as --domain %s --name %s\n",
		shellQuote(externalID), shellQuote(snap))
	_, stderr, err := d.sshExec().RunScript(ctx, host, script)
	if err != nil {
		return "", fmt.Errorf("virsh snapshot %s: %v (%s)", externalID, err, strings.TrimSpace(stderr))
	}
	return snap, nil
}

func (d *LinuxNestedQEMUDriver) DeleteInstance(ctx context.Context, host SupplyHost, externalID string) error {
	if err := validateExternalDomain(externalID); err != nil {
		return err
	}
	storagePath := normalizeStoragePath(host.StoragePoolPath)
	vmDir := storagePath + "/" + externalID
	script := fmt.Sprintf(`set -euo pipefail
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  SUDO="sudo -n"
fi
$SUDO virsh destroy %s 2>/dev/null || true
$SUDO virsh undefine %s --nvram 2>/dev/null || $SUDO virsh undefine %s 2>/dev/null || true
$SUDO rm -rf %s
`, shellQuote(externalID), shellQuote(externalID), shellQuote(externalID), shellQuote(vmDir))
	_, stderr, err := d.sshExec().RunScript(ctx, host, script)
	if err != nil {
		return fmt.Errorf("delete domain %s: %v (%s)", externalID, err, strings.TrimSpace(stderr))
	}
	return nil
}

func (d *LinuxNestedQEMUDriver) ConsoleInfo(ctx context.Context, host SupplyHost, externalID string) (string, error) {
	if err := validateExternalDomain(externalID); err != nil {
		return "", err
	}
	script := fmt.Sprintf("set -euo pipefail\n$([ \"$(id -u)\" -eq 0 ] || echo 'sudo -n') virsh domdisplay %s\n", shellQuote(externalID))
	stdout, stderr, err := d.sshExec().RunScript(ctx, host, script)
	if err != nil {
		return "", fmt.Errorf("virsh domdisplay %s: %v (%s)", externalID, err, strings.TrimSpace(stderr))
	}
	return strings.TrimSpace(stdout), nil
}

func sanitizeHostname(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else if r == '_' || r == ' ' || r == '.' {
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func shellSafeYAMLString(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
