package providers

import (
	"fmt"
	"sync"
)

// Registry maps DriverKind identifiers to their ComputeDriver, HostingDriver,
// and NetworkStorageDriver implementations.
type Registry struct {
	mu         sync.RWMutex
	compute    map[DriverKind]ComputeDriver
	hosting    map[DriverKind]HostingDriver
	netStorage map[DriverKind]NetworkStorageDriver
}

// NewRegistry returns an empty driver registry.
func NewRegistry() *Registry {
	return &Registry{
		compute:    make(map[DriverKind]ComputeDriver),
		hosting:    make(map[DriverKind]HostingDriver),
		netStorage: make(map[DriverKind]NetworkStorageDriver),
	}
}

func (r *Registry) RegisterCompute(d ComputeDriver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.compute[d.Kind()] = d
}

func (r *Registry) Compute(kind DriverKind) (ComputeDriver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.compute[kind]
	if !ok {
		return nil, fmt.Errorf("%w: compute driver %q", ErrDriverNotFound, kind)
	}
	return d, nil
}

func (r *Registry) ListComputeKinds() []DriverKind {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]DriverKind, 0, len(r.compute))
	for k := range r.compute {
		out = append(out, k)
	}
	return out
}

func (r *Registry) RegisterHosting(d HostingDriver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hosting[d.Kind()] = d
}

func (r *Registry) Hosting(kind DriverKind) (HostingDriver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.hosting[kind]
	if !ok {
		return nil, fmt.Errorf("%w: hosting driver %q", ErrDriverNotFound, kind)
	}
	return d, nil
}

func (r *Registry) RegisterNetworkStorage(d NetworkStorageDriver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.netStorage[d.Kind()] = d
}

func (r *Registry) NetworkStorage(kind DriverKind) (NetworkStorageDriver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.netStorage[kind]
	if !ok {
		return nil, fmt.Errorf("%w: network/storage driver %q", ErrDriverNotFound, kind)
	}
	return d, nil
}

// NewDefaultRegistry registers all built-in drivers (including placeholders).
func NewDefaultRegistry() *Registry {
	reg := NewRegistry()
	reg.RegisterCompute(&LinuxNestedQEMUPlaceholderDriver{})
	reg.RegisterCompute(&QEMULibvirtDriver{})
	reg.RegisterCompute(&ProxmoxDriver{})
	reg.RegisterCompute(&VMwareVSphereDriver{})
	reg.RegisterCompute(&UpstreamResellerDriver{})

	reg.RegisterHosting(&SharedWebDriver{})
	reg.RegisterHosting(&PaaSAppDriver{})

	reg.RegisterNetworkStorage(&IPAMDriver{})
	reg.RegisterNetworkStorage(&BlockStorageDriver{})
	return reg
}
