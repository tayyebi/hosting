package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ComputeK8sCluster is a row from workspace_compute_k8s joined with resources.
type ComputeK8sCluster struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	ClusterID   string
	K8sVersion  string
	NodeCount   int
	NodeConfig  string
	PodsRunning int
	CNIPlugin   string
	IngressCtrl string
	Autoscale   bool
	HAControl   bool
	CreatedAt   time.Time
}

// ListComputeK8sClusters returns all Kubernetes clusters in a workspace.
func ListComputeK8sClusters(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ComputeK8sCluster, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       k.cluster_id, k.k8s_version, k.node_count, k.node_config,
		       k.pods_running, k.cni_plugin, k.ingress_ctrl, k.autoscale, k.ha_control,
		       r.created_at
		FROM workspace_compute_k8s k
		JOIN resources r ON r.id = k.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListComputeK8sClusters: %w", err)
	}
	defer rows.Close()
	var out []ComputeK8sCluster
	for rows.Next() {
		var c ComputeK8sCluster
		if err := rows.Scan(&c.ResourceID, &c.Name, &c.Status, &c.Region,
			&c.ClusterID, &c.K8sVersion, &c.NodeCount, &c.NodeConfig,
			&c.PodsRunning, &c.CNIPlugin, &c.IngressCtrl, &c.Autoscale, &c.HAControl,
			&c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// CreateComputeK8sCluster inserts a new Kubernetes cluster resource.
func CreateComputeK8sCluster(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, region, k8sVersion, nodeConfig string,
	nodeCount int,
) (int64, error) {
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
		VALUES ($1,$2,'compute_k8s',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateComputeK8sCluster resource: %w", err)
	}

	clusterID := fmt.Sprintf("eks-%d", resourceID)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_compute_k8s
		    (resource_id, cluster_id, k8s_version, node_count, node_config)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, clusterID, k8sVersion, nodeCount, nodeConfig)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateComputeK8sCluster k8s row: %w", err)
	}

	return resourceID, nil
}
