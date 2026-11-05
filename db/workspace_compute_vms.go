package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ComputeVM is a row from workspace_compute_vms joined with resources.
type ComputeVM struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	VMClass      string
	InstanceType string
	VCPUs        int
	MemoryGB     float64
	GPUModel     string
	GPUCount     int
	IPAddress    string
	OSImage      string
	DiskGB       int
}

// ListComputeVMs returns all VM instances in a workspace filtered by class.
// vmClass is "general", "gpu", or "spot".
func ListComputeVMs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64, vmClass string) ([]ComputeVM, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       v.vm_class, v.instance_type, v.vcpus, v.memory_gb,
		       v.gpu_model, v.gpu_count, v.ip_address, v.os_image, v.disk_gb
		FROM workspace_compute_vms v
		JOIN resources r ON r.id = v.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND v.vm_class    = $2
		  AND r.deleted_at  IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID, vmClass)
	if err != nil {
		return nil, fmt.Errorf("db.ListComputeVMs: %w", err)
	}
	defer rows.Close()
	var out []ComputeVM
	for rows.Next() {
		var v ComputeVM
		if err := rows.Scan(&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.VMClass, &v.InstanceType, &v.VCPUs, &v.MemoryGB,
			&v.GPUModel, &v.GPUCount, &v.IPAddress, &v.OSImage, &v.DiskGB); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// CreateComputeVM inserts a new VM resource into both resources and
// workspace_compute_vms. Returns the new resource ID.
func CreateComputeVM(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, vmClass, instanceType, region, osImage string,
	vcpus, diskGB int, memoryGB float64,
	gpuModel string, gpuCount int,
) (int64, error) {
	// Resolve region
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'compute_vm',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateComputeVM resource: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_compute_vms
		    (resource_id, vm_class, instance_type, vcpus, memory_gb, gpu_model, gpu_count, os_image, disk_gb)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		resourceID, vmClass, instanceType, vcpus, memoryGB, gpuModel, gpuCount, osImage, diskGB)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateComputeVM vm row: %w", err)
	}

	return resourceID, nil
}
