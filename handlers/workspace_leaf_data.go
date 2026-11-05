package handlers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"platform/db"
)

// WSLeafKPI is a single KPI tile shown above the table.
type WSLeafKPI struct {
	Label string
	Value template.HTML // may include <span> for delta
	Sub   string
}

// WSLeafData is the per-leaf rendering payload.
type WSLeafData struct {
	KPIs    []WSLeafKPI
	Headers []string
	Rows    [][]template.HTML // rows of cells (HTML allowed for badges)
	Note    string
}

// helpers
func k(label, val, sub string) WSLeafKPI {
	return WSLeafKPI{Label: label, Value: template.HTML(val), Sub: sub}
}
func r(cells ...string) []template.HTML {
	out := make([]template.HTML, len(cells))
	for i, c := range cells {
		out[i] = template.HTML(c)
	}
	return out
}

// status badges
const (
	bGreen = `<span class="lf-badge lf-badge--green"><span class="dot"></span>%s</span>`
	bAmber = `<span class="lf-badge lf-badge--amber"><span class="dot"></span>%s</span>`
	bRed   = `<span class="lf-badge lf-badge--red"><span class="dot"></span>%s</span>`
	bBlue  = `<span class="lf-badge lf-badge--blue"><span class="dot"></span>%s</span>`
	bGray  = `<span class="lf-badge lf-badge--gray"><span class="dot"></span>%s</span>`
)

func g(s string) string {
	return `<span class="lf-badge lf-badge--green"><span class="dot"></span>` + s + `</span>`
}
func y(s string) string {
	return `<span class="lf-badge lf-badge--amber"><span class="dot"></span>` + s + `</span>`
}
func b(s string) string {
	return `<span class="lf-badge lf-badge--blue"><span class="dot"></span>` + s + `</span>`
}
func e(s string) string {
	return `<span class="lf-badge lf-badge--red"><span class="dot"></span>` + s + `</span>`
}
func gr(s string) string {
	return `<span class="lf-badge lf-badge--gray"><span class="dot"></span>` + s + `</span>`
}

// progress bar
func bar(pct int, kind string) string {
	cls := ""
	switch kind {
	case "amber":
		cls = " lf-prog__bar--amber"
	case "red":
		cls = " lf-prog__bar--red"
	}
	return `<div class="lf-prog"><div class="lf-prog__bar` + cls + `" style="width:` + intStr(pct) + `%"></div></div>`
}

func intStr(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// monospace ID
func m(s string) string { return `<span class="lf-mono">` + s + `</span>` }

// strong text
func s(t string) string { return `<strong>` + t + `</strong>` }

// strong + monospace ID combo (like a name + ID)
func ni(name, id string) string { return s(name) + `<div class="lf-mono">` + id + `</div>` }

// kpi delta
func up(delta string) string {
	return ` <span class="lf-kpi__delta lf-kpi__delta--up">` + delta + `</span>`
}
func down(delta string) string {
	return ` <span class="lf-kpi__delta lf-kpi__delta--down">` + delta + `</span>`
}

// Common KPI sets per category — empty/zeroed baseline when no live DB rows exist.
func kpiCompute() []WSLeafKPI {
	return []WSLeafKPI{
		k("Active resources", "0", "0 provisioned"),
		k("vCPUs in use", "0", "0% utilisation"),
		k("Memory", `0 <span style="font-size:13px;color:var(--lf-dim)">GB</span>`, "0% utilisation"),
		k("Monthly spend", "$0.00", "no active resources"),
	}
}
func kpiStorage() []WSLeafKPI {
	return []WSLeafKPI{
		k("Total stored", `0 <span style="font-size:13px;color:var(--lf-dim)">GB</span>`, "across 0 objects"),
		k("Monthly egress", `0 <span style="font-size:13px;color:var(--lf-dim)">GB</span>`, "$0.00"),
		k("Avg. latency", `—`, "no traffic yet"),
		k("Backup coverage", "—", "no volumes"),
	}
}
func kpiDB() []WSLeafKPI {
	return []WSLeafKPI{
		k("Active databases", "0", "0 primaries · 0 replicas"),
		k("Connections", "0", "no active pools"),
		k("QPS", "0", "reads + writes"),
		k("Storage", `0 <span style="font-size:13px;color:var(--lf-dim)">GB</span>`, "0% of allocated"),
	}
}
func kpiNet() []WSLeafKPI {
	return []WSLeafKPI{
		k("Resources", "0", "across 0 regions"),
		k("Traffic (24h)", `0 <span style="font-size:13px;color:var(--lf-dim)">GB</span>`, "ingress + egress"),
		k("Requests / s", "0", "no active traffic"),
		k("Packets dropped", "0%", "no active traffic"),
	}
}
func kpiAI() []WSLeafKPI {
	return []WSLeafKPI{
		k("Active jobs", "0", "0 running now"),
		k("GPU hours", "0", "this month"),
		k("Models deployed", "0", "in production"),
		k("Cost (MTD)", "$0.00", "no active usage"),
	}
}
func kpiSec() []WSLeafKPI {
	return []WSLeafKPI{
		k("Security score", `—`, "no resources evaluated"),
		k("Active alerts", "0", "0 critical · 0 medium"),
		k("Identities", "0", "managed"),
		k("Last audit", "—", "none recorded"),
	}
}
func kpiDevOps() []WSLeafKPI {
	return []WSLeafKPI{
		k("Builds today", "0", "no builds"),
		k("Success rate", "—", "last 7 days"),
		k("Mean duration", "—", "no runs"),
		k("Deployments (7d)", "0", "to production"),
	}
}
func kpiApps() []WSLeafKPI {
	return []WSLeafKPI{
		k("Throughput", "0 /s", "last hour"),
		k("Queues / Workflows", "0", "0 running"),
		k("Success rate", "—", "last 24h"),
		k("Lag", `0 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`, "idle"),
	}
}
func kpiMC() []WSLeafKPI {
	return []WSLeafKPI{
		k("Clouds connected", "0", "none configured"),
		k("Monthly spend", "$0.00", "no active resources"),
		k("Policy violations", "0", "0 critical"),
		k("Migrations", "0", "in flight"),
	}
}

// hdr is a shorthand for headers
func hdr(cells ...string) []string { return cells }

// leafFor returns the per-leaf schema (headers + zeroed KPIs) when a workspace has
// zero database rows for that leaf.
func leafFor(navID string) WSLeafData {
	all := leafRegistry()
	if d, ok := all[navID]; ok {
		d.Rows = nil
		for i := range d.KPIs {
			d.KPIs[i].Value = template.HTML("0")
			d.KPIs[i].Sub = "no active resources"
		}
		return d
	}
	return WSLeafData{}
}

// leafFromDB attempts to build WSLeafData from the live database for leaf pages
// that have a DB-backed resource table. It returns (data, true) when DB rows
// were found, or (WSLeafData{}, false) when there is no DB data (so the caller
// can fall back to the static registry).
func leafFromDB(ctx context.Context, navID string, workspaceDBID int64) (WSLeafData, bool) {
	if Pools == nil || workspaceDBID == 0 {
		return WSLeafData{}, false
	}

	qCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	switch navID {
	case "compute/vms/general":
		return computeVMLeafFromDB(qCtx, workspaceDBID, "general")
	case "compute/vms/gpu":
		return computeVMLeafFromDB(qCtx, workspaceDBID, "gpu")
	case "compute/vms/spot":
		return computeVMLeafFromDB(qCtx, workspaceDBID, "spot")
	case "compute/containers/kubernetes":
		return computeK8sLeafFromDB(qCtx, workspaceDBID)
	case "compute/containers/registry":
		return computeRegistryLeafFromDB(qCtx, workspaceDBID)
	case "compute/containers/instances":
		return computeContainerInstancesLeafFromDB(qCtx, workspaceDBID)
	case "compute/hosting/web":
		return hostingWebLeafFromDB(qCtx, workspaceDBID)
	case "compute/hosting/static":
		return hostingStaticLeafFromDB(qCtx, workspaceDBID)
	case "compute/serverless/functions":
		return serverlessFunctionsLeafFromDB(qCtx, workspaceDBID)
	case "compute/serverless/edge":
		return serverlessEdgeLeafFromDB(qCtx, workspaceDBID)
	case "compute/batch/jobs":
		return batchJobsLeafFromDB(qCtx, workspaceDBID)
	case "compute/batch/hpc":
		return batchHPCLeafFromDB(qCtx, workspaceDBID)
	case "storage/object/standard":
		return storageObjectLeafFromDB(qCtx, workspaceDBID, "standard")
	case "storage/object/archive":
		return storageObjectLeafFromDB(qCtx, workspaceDBID, "archive")
	case "storage/object/datalake":
		return storageObjectLeafFromDB(qCtx, workspaceDBID, "datalake")
	case "storage/block/ssd":
		return storageBlockLeafFromDB(qCtx, workspaceDBID, "ssd")
	case "storage/block/hdd":
		return storageBlockLeafFromDB(qCtx, workspaceDBID, "hdd")
	case "storage/file/nfs":
		return storageFileNFSLeafFromDB(qCtx, workspaceDBID)
	case "storage/file/distributed":
		return storageFileDistributedLeafFromDB(qCtx, workspaceDBID)
	case "storage/backup/snapshots":
		return storageBackupSnapshotsLeafFromDB(qCtx, workspaceDBID)
	case "storage/backup/dr":
		return storageBackupDRLeafFromDB(qCtx, workspaceDBID)
	case "storage/backup/archive":
		return storageBackupArchiveLeafFromDB(qCtx, workspaceDBID)
	case "databases/relational/mysql-pg":
		return dbRelationalMySQLPGLeafFromDB(qCtx, workspaceDBID)
	case "databases/relational/sql-server":
		return dbRelationalSQLServerLeafFromDB(qCtx, workspaceDBID)
	case "databases/relational/cloud-native":
		return dbRelationalCloudNativeLeafFromDB(qCtx, workspaceDBID)
	case "databases/nosql/document":
		return dbNoSQLDocumentLeafFromDB(qCtx, workspaceDBID)
	case "databases/nosql/keyvalue":
		return dbNoSQLKeyValueLeafFromDB(qCtx, workspaceDBID)
	case "databases/nosql/wide-column":
		return dbNoSQLWideColumnLeafFromDB(qCtx, workspaceDBID)
	case "databases/nosql/graph":
		return dbNoSQLGraphLeafFromDB(qCtx, workspaceDBID)
	case "databases/memory/cache":
		return dbMemoryCacheLeafFromDB(qCtx, workspaceDBID)
	case "databases/memory/session":
		return dbMemorySessionLeafFromDB(qCtx, workspaceDBID)
	case "databases/analytics/warehouse":
		return dbAnalyticsWarehouseLeafFromDB(qCtx, workspaceDBID)
	case "databases/analytics/timeseries":
		return dbAnalyticsTimeSeriesLeafFromDB(qCtx, workspaceDBID)
	case "databases/analytics/search":
		return dbAnalyticsSearchLeafFromDB(qCtx, workspaceDBID)
	case "networking/vnet/vpc":
		return dbNetVPCLeafFromDB(qCtx, workspaceDBID)
	case "networking/vnet/peering":
		return dbNetPeeringLeafFromDB(qCtx, workspaceDBID)
	case "networking/vnet/nsg":
		return dbNetNSGLeafFromDB(qCtx, workspaceDBID)
	case "networking/lb/app":
		return dbNetLBAppLeafFromDB(qCtx, workspaceDBID)
	case "networking/lb/network":
		return dbNetLBNetworkLeafFromDB(qCtx, workspaceDBID)
	case "networking/lb/global":
		return dbNetLBGlobalLeafFromDB(qCtx, workspaceDBID)
	case "networking/dns/manage":
		return dbNetDNSZoneLeafFromDB(qCtx, workspaceDBID)
	case "networking/dns/domains":
		return dbNetDomainLeafFromDB(qCtx, workspaceDBID)
	case "networking/cdn/static":
		return dbNetCDNStaticLeafFromDB(qCtx, workspaceDBID)
	case "networking/cdn/video":
		return dbNetCDNVideoLeafFromDB(qCtx, workspaceDBID)
	case "networking/cdn/edge":
		return dbNetCDNEdgeLeafFromDB(qCtx, workspaceDBID)
	case "networking/private/vpn":
		return dbNetVPNGatewayLeafFromDB(qCtx, workspaceDBID)
	case "networking/private/direct":
		return dbNetDirectConnectLeafFromDB(qCtx, workspaceDBID)
	case "networking/private/sdwan":
		return dbNetSDWANLeafFromDB(qCtx, workspaceDBID)
	case "ai/training/distributed":
		return dbAIDistributedTrainingLeafFromDB(qCtx, workspaceDBID)
	case "ai/training/automl":
		return dbAIAutoMLLeafFromDB(qCtx, workspaceDBID)
	case "ai/training/finetuning":
		return dbAIFineTuningLeafFromDB(qCtx, workspaceDBID)
	case "ai/hosting/inference":
		return dbAIInferenceEndpointLeafFromDB(qCtx, workspaceDBID)
	case "ai/hosting/batch":
		return dbAIBatchJobLeafFromDB(qCtx, workspaceDBID)
	case "ai/hosting/edge":
		return dbAIEdgeDeploymentLeafFromDB(qCtx, workspaceDBID)
	case "ai/apis/vision":
		return dbAIVisionAPILeafFromDB(qCtx, workspaceDBID)
	case "ai/apis/language":
		return dbAILanguageAPILeafFromDB(qCtx, workspaceDBID)
	case "ai/apis/speech":
		return dbAISpeechAPILeafFromDB(qCtx, workspaceDBID)
	case "ai/apis/recommendation":
		return dbAIRecommendationAPILeafFromDB(qCtx, workspaceDBID)
	case "ai/data/etl":
		return dbAIETLPipelineLeafFromDB(qCtx, workspaceDBID)
	case "ai/data/streaming":
		return dbAIDataStreamLeafFromDB(qCtx, workspaceDBID)
	case "ai/data/catalog":
		return dbAIDataCatalogLeafFromDB(qCtx, workspaceDBID)
	case "ai/mlops/experiments":
		return dbAIMLExperimentLeafFromDB(qCtx, workspaceDBID)
	case "ai/mlops/registry":
		return dbAIModelRegistryLeafFromDB(qCtx, workspaceDBID)
	case "ai/mlops/monitoring":
		return dbAIMLMonitorLeafFromDB(qCtx, workspaceDBID)
	case "security/iam/sso":
		return dbSecuritySSOLeafFromDB(qCtx, workspaceDBID)
	case "security/iam/mfa":
		return dbSecurityMFALeafFromDB(qCtx, workspaceDBID)
	case "security/iam/rbac":
		return dbSecurityRBACLeafFromDB(qCtx, workspaceDBID)
	case "security/data/enc-rest":
		return dbSecurityEncRestLeafFromDB(qCtx, workspaceDBID)
	case "security/data/enc-transit":
		return dbSecurityEncTransitLeafFromDB(qCtx, workspaceDBID)
	case "security/data/kms":
		return dbSecurityKMSLeafFromDB(qCtx, workspaceDBID)
	case "security/network/waf":
		return dbSecurityWAFLeafFromDB(qCtx, workspaceDBID)
	case "security/network/ddos":
		return dbSecurityDDoSLeafFromDB(qCtx, workspaceDBID)
	case "security/network/ids":
		return dbSecurityIDSLeafFromDB(qCtx, workspaceDBID)
	case "security/compliance/policy":
		return dbSecurityCompliancePolicyLeafFromDB(qCtx, workspaceDBID)
	case "security/compliance/audit":
		return dbSecurityComplianceAuditLeafFromDB(qCtx, workspaceDBID)
	case "security/compliance/regulatory":
		return dbSecurityRegulatoryLeafFromDB(qCtx, workspaceDBID)
	case "devops/cicd/source":
		return dbDevOpsCICDSourceLeafFromDB(qCtx, workspaceDBID)
	case "devops/cicd/build":
		return dbDevOpsCICDBuildLeafFromDB(qCtx, workspaceDBID)
	case "devops/cicd/deploy":
		return dbDevOpsCICDDeployLeafFromDB(qCtx, workspaceDBID)
	case "devops/iac/templates":
		return dbDevOpsIaCTemplatesLeafFromDB(qCtx, workspaceDBID)
	case "devops/iac/config":
		return dbDevOpsIaCConfigLeafFromDB(qCtx, workspaceDBID)
	case "devops/api/gateway":
		return dbDevOpsAPIGatewayLeafFromDB(qCtx, workspaceDBID)
	case "devops/api/portal":
		return dbDevOpsAPIPortalLeafFromDB(qCtx, workspaceDBID)
	case "devops/api/analytics":
		return dbDevOpsAPIAnalyticsLeafFromDB(qCtx, workspaceDBID)
	case "devops/obs/logging":
		return dbDevOpsObsLoggingLeafFromDB(qCtx, workspaceDBID)
	case "devops/obs/metrics":
		return dbDevOpsObsMetricsLeafFromDB(qCtx, workspaceDBID)
	case "devops/obs/tracing":
		return dbDevOpsObsTracingLeafFromDB(qCtx, workspaceDBID)
	case "apps/messaging/queues":
		return dbAppsMessagingQueuesLeafFromDB(qCtx, workspaceDBID)
	case "apps/messaging/streaming":
		return dbAppsMessagingStreamingLeafFromDB(qCtx, workspaceDBID)
	case "apps/messaging/servicebus":
		return dbAppsMessagingServiceBusLeafFromDB(qCtx, workspaceDBID)
	case "apps/workflow/automation":
		return dbAppsWorkflowAutomationLeafFromDB(qCtx, workspaceDBID)
	case "apps/workflow/scheduler":
		return dbAppsWorkflowSchedulerLeafFromDB(qCtx, workspaceDBID)
	case "apps/media/encoding":
		return dbAppsMediaEncodingLeafFromDB(qCtx, workspaceDBID)
	case "apps/media/livestream":
		return dbAppsMediaLivestreamLeafFromDB(qCtx, workspaceDBID)
	case "apps/media/storage":
		return dbAppsMediaStorageLeafFromDB(qCtx, workspaceDBID)
	case "multicloud/hybrid/onprem":
		return dbMulticloudOnpremLeafFromDB(qCtx, workspaceDBID)
	case "multicloud/hybrid/control":
		return dbMulticloudControlLeafFromDB(qCtx, workspaceDBID)
	case "multicloud/hybrid/edge":
		return dbMulticloudEdgeLeafFromDB(qCtx, workspaceDBID)
	case "multicloud/mgmt/cost":
		return dbMulticloudCostLeafFromDB(qCtx, workspaceDBID)
	case "multicloud/mgmt/policy":
		return dbMulticloudPolicyLeafFromDB(qCtx, workspaceDBID)
	case "multicloud/mgmt/workload":
		return dbMulticloudWorkloadLeafFromDB(qCtx, workspaceDBID)
	case "multicloud/migration/assessment":
		return dbMulticloudMigrationAssessmentLeafFromDB(qCtx, workspaceDBID)
	case "multicloud/migration/data":
		return dbMulticloudDataMigrationLeafFromDB(qCtx, workspaceDBID)
	case "multicloud/migration/app":
		return dbMulticloudAppMigrationLeafFromDB(qCtx, workspaceDBID)
	case "apps/app-hosting/static-sites":
		return dbAppsStaticSitesLeafFromDB(qCtx, workspaceDBID)
	case "compute/virtual-machines/linux-vms":
		return dbComputeLinuxVMsLeafFromDB(qCtx, workspaceDBID)
	case "databases/relational/postgresql":
		return dbManagedPostgresLeafFromDB(qCtx, workspaceDBID)
	case "storage/object-storage/s3-compatible":
		return dbProS3BucketsLeafFromDB(qCtx, workspaceDBID)
	case "apps/communication/transactional-email":
		return dbProTransactionalEmailLeafFromDB(qCtx, workspaceDBID)
	case "networking/cdn/pop-management":
		return dbProCDNPoPLeafFromDB(qCtx, workspaceDBID)
	case "security/certificates/ssl-tls":
		return dbProSSLCertsLeafFromDB(qCtx, workspaceDBID)
	case "devops/monitoring/metrics":
		return dbProMonitoringAlertsLeafFromDB(qCtx, workspaceDBID)
	}
	return WSLeafData{}, false
}

func computeVMLeafFromDB(ctx context.Context, workspaceDBID int64, vmClass string) (WSLeafData, bool) {
	vms, err := db.ListComputeVMs(ctx, Pools.Admin, workspaceDBID, vmClass)
	if err != nil {
		log.Printf("computeVMLeafFromDB(%s): %v", vmClass, err)
		return WSLeafData{}, false
	}
	if len(vms) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(vms))
	for i, v := range vms {
		var statusBadge string
		switch v.Status {
		case "active":
			statusBadge = g("Running")
		case "stopped", "deleted":
			statusBadge = gr(v.Status)
		case "degraded", "failed":
			statusBadge = e(v.Status)
		default:
			statusBadge = y(v.Status)
		}

		if vmClass == "gpu" {
			gpuInfo := v.GPUModel
			if v.GPUCount > 0 {
				gpuInfo = fmt.Sprintf("%d× %s", v.GPUCount, v.GPUModel)
			}
			rows[i] = r(
				ni(v.Name, fmt.Sprintf("res-%d", v.ResourceID)),
				gpuInfo,
				v.InstanceType,
				v.Region,
				statusBadge,
				fmt.Sprintf("%.0f GB", v.MemoryGB),
				`<button class="lf-btn lf-btn--ghost">Open →</button>`,
			)
		} else {
			rows[i] = r(
				ni(v.Name, fmt.Sprintf("res-%d", v.ResourceID)),
				v.InstanceType,
				v.Region,
				statusBadge,
				fmt.Sprintf("%d vCPU", v.VCPUs),
				fmt.Sprintf("%.1f GB RAM", v.MemoryGB),
				`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
			)
		}
	}

	var headers []string
	if vmClass == "gpu" {
		headers = hdr("Name", "GPU", "Type", "Region", "Status", "VRAM", "")
	} else {
		headers = hdr("Name", "Type", "Region", "Status", "vCPU", "Memory", "")
	}

	return WSLeafData{
		KPIs:    kpiCompute(),
		Headers: headers,
		Rows:    rows,
	}, true
}

func computeK8sLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	clusters, err := db.ListComputeK8sClusters(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("computeK8sLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(clusters) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(clusters))
	for i, c := range clusters {
		var statusBadge string
		switch c.Status {
		case "active":
			statusBadge = g("Healthy")
		case "provisioning":
			statusBadge = b("Provisioning")
		case "degraded":
			statusBadge = y("Degraded")
		case "failed":
			statusBadge = e("Failed")
		default:
			statusBadge = gr(c.Status)
		}
		autoscale := ""
		if c.Autoscale {
			autoscale = "✓"
		}
		rows[i] = r(
			ni(c.Name, c.ClusterID),
			c.K8sVersion,
			fmt.Sprintf("%d", c.NodeCount),
			c.Region,
			statusBadge,
			fmt.Sprintf("%d", c.PodsRunning),
			autoscale,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}

	return WSLeafData{
		KPIs:    kpiCompute(),
		Headers: hdr("Cluster", "Version", "Nodes", "Region", "Status", "Pods", "Autoscale", ""),
		Rows:    rows,
	}, true
}

func computeRegistryLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	repos, err := db.ListContainerRegistries(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("computeRegistryLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(repos) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(repos))
	for i, repo := range repos {
		var vulnBadge string
		switch {
		case repo.VulnCritical > 0:
			vulnBadge = e(fmt.Sprintf("%d critical", repo.VulnCritical))
		case repo.VulnHigh > 0:
			vulnBadge = y(fmt.Sprintf("%d high", repo.VulnHigh))
		case repo.VulnMedium > 0:
			vulnBadge = y(fmt.Sprintf("%d medium", repo.VulnMedium))
		default:
			vulnBadge = g("0 critical")
		}
		scanStr := "never"
		if repo.LastScanAt != nil {
			scanStr = repo.LastScanAt.Format("2006-01-02")
		}
		sizeMB := fmt.Sprintf("%.0f MB", float64(repo.SizeBytes)/1024/1024)
		if repo.SizeBytes >= 1024*1024*1024 {
			sizeMB = fmt.Sprintf("%.1f GB", float64(repo.SizeBytes)/1024/1024/1024)
		}
		rows[i] = r(
			s(repo.RepoName),
			fmt.Sprintf("%d", repo.TagCount),
			sizeMB,
			scanStr,
			vulnBadge,
			`<button class="lf-btn lf-btn--ghost">View tags →</button>`,
		)
	}

	return WSLeafData{
		KPIs:    kpiCompute(),
		Headers: hdr("Repository", "Tags", "Size", "Last scan", "Vulns", ""),
		Rows:    rows,
	}, true
}

func computeContainerInstancesLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListContainerInstances(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("computeContainerInstancesLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(instances))
	for i, ci := range instances {
		var statusBadge string
		switch ci.Status {
		case "active":
			statusBadge = g("Running")
		case "pending", "provisioning":
			statusBadge = b("Pending")
		case "failed":
			statusBadge = e("Failed")
		case "stopped", "deleted":
			statusBadge = gr(ci.Status)
		default:
			statusBadge = y(ci.Status)
		}
		cpuStr := fmt.Sprintf("%.2f", float64(ci.CPUMillicores)/1000)
		specStr := cpuStr + " / " + fmt.Sprintf("%d MB", ci.MemoryMB)
		rows[i] = r(
			s(ci.Name),
			m(ci.Image),
			specStr,
			ci.Region,
			statusBadge,
			fmt.Sprintf("%d", ci.RestartCount),
			`<button class="lf-btn lf-btn--ghost">Logs →</button>`,
		)
	}

	return WSLeafData{
		KPIs:    kpiCompute(),
		Headers: hdr("Name", "Image", "CPU/Mem", "Region", "Status", "Restarts", ""),
		Rows:    rows,
	}, true
}

func hostingWebLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	apps, err := db.ListHostingWebApps(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("hostingWebLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(apps) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(apps))
	for i, app := range apps {
		var statusBadge string
		switch app.Status {
		case "active":
			statusBadge = g("Online")
		case "stopped":
			statusBadge = gr("Stopped")
		case "degraded", "failed":
			statusBadge = e(app.Status)
		case "provisioning":
			statusBadge = b("Deploying")
		default:
			statusBadge = y(app.Status)
		}
		stackLabel := app.Runtime
		if app.Stack != "" {
			stackLabel = app.Stack + " · " + app.Runtime
		}
		rows[i] = r(
			ni(app.Domain, app.Name),
			stackLabel,
			app.Region,
			statusBadge,
			bar(app.CPUPct, cpuBarKind(app.CPUPct)),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}

	return WSLeafData{
		KPIs:    kpiCompute(),
		Headers: hdr("Domain", "Stack", "Region", "Status", "CPU", ""),
		Rows:    rows,
	}, true
}

func hostingStaticLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	sites, err := db.ListStaticSites(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("hostingStaticLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(sites) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(sites))
	for i, site := range sites {
		var statusBadge string
		switch site.Status {
		case "active":
			statusBadge = g("Live")
		case "building":
			statusBadge = b("Building")
		case "failed":
			statusBadge = e("Build failed")
		case "stopped":
			statusBadge = gr("Stopped")
		default:
			statusBadge = y(site.Status)
		}
		buildStr := "—"
		if site.BuildTimeSec > 0 {
			buildStr = fmt.Sprintf("%ds", site.BuildTimeSec)
		}
		deployStr := "never"
		if site.LastDeployAt != nil {
			deployStr = site.LastDeployAt.Format("2006-01-02 15:04")
		}
		rows[i] = r(
			ni(site.Name, fmt.Sprintf("res-%d", site.ResourceID)),
			site.Framework,
			m(site.Branch),
			deployStr,
			buildStr,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Deploys →</button>`,
		)
	}

	return WSLeafData{
		KPIs:    kpiCompute(),
		Headers: hdr("Site", "Framework", "Branch", "Last deploy", "Build time", "Status", ""),
		Rows:    rows,
	}, true
}

func serverlessFunctionsLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	funcs, err := db.ListServerlessFunctions(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("serverlessFunctionsLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(funcs) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(funcs))
	for i, fn := range funcs {
		var statusBadge string
		switch fn.Status {
		case "active":
			statusBadge = g("Active")
		case "disabled":
			statusBadge = gr("Disabled")
		case "failed":
			statusBadge = e("Failed")
		default:
			statusBadge = y(fn.Status)
		}
		runtimeMem := fmt.Sprintf("%s · %d MB", fn.Runtime, fn.MemoryMB)
		invStr := intStr(fn.Invocations1D)
		durStr := fmt.Sprintf("%d ms", fn.AvgDurationMS)
		rows[i] = r(
			ni(fn.Name, fmt.Sprintf("res-%d", fn.ResourceID)),
			m(runtimeMem),
			fn.TriggerType,
			invStr,
			durStr,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Logs →</button>`,
		)
	}

	return WSLeafData{
		KPIs:    kpiCompute(),
		Headers: hdr("Function", "Runtime", "Trigger", "Invocations/day", "Avg duration", "Status", ""),
		Rows:    rows,
	}, true
}

func serverlessEdgeLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	fns, err := db.ListEdgeFunctions(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("serverlessEdgeLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(fns) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(fns))
	for i, fn := range fns {
		var statusBadge string
		switch fn.Status {
		case "active":
			statusBadge = g("Active")
		case "disabled":
			statusBadge = gr("Disabled")
		case "failed":
			statusBadge = e("Failed")
		default:
			statusBadge = y(fn.Status)
		}
		latStr := fmt.Sprintf("%d ms", fn.P99LatencyMS)
		rpsStr := intStr(fn.RequestsPerSec)
		rows[i] = r(
			ni(fn.Name, fmt.Sprintf("res-%d", fn.ResourceID)),
			m(fn.PathPattern),
			intStr(fn.PoPs),
			rpsStr,
			latStr,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}

	return WSLeafData{
		KPIs:    kpiCompute(),
		Headers: hdr("Function", "Path", "PoPs", "Requests/s", "p99 latency", "Status", ""),
		Rows:    rows,
	}, true
}

func batchJobsLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	jobs, err := db.ListBatchJobs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("batchJobsLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(jobs) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(jobs))
	for i, job := range jobs {
		var statusBadge string
		switch job.Status {
		case "active":
			statusBadge = g("Last ok")
		case "running":
			statusBadge = b("Running")
		case "failed":
			statusBadge = e("Last failed")
		case "disabled":
			statusBadge = gr("Disabled")
		default:
			statusBadge = y(job.Status)
		}
		lastRunStr := "never"
		if job.LastRunAt != nil {
			lastRunStr = job.LastRunAt.Format("2006-01-02 15:04")
		}
		durStr := "—"
		if job.LastDurationSec > 0 {
			mins := job.LastDurationSec / 60
			secs := job.LastDurationSec % 60
			durStr = fmt.Sprintf("%dm %ds", mins, secs)
		}
		nextRunStr := "—"
		if job.NextRunAt != nil {
			nextRunStr = job.NextRunAt.Format("2006-01-02 15:04")
		}
		rows[i] = r(
			ni(job.Name, fmt.Sprintf("res-%d", job.ResourceID)),
			m(job.Schedule),
			lastRunStr,
			durStr,
			nextRunStr,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Runs →</button>`,
		)
	}

	return WSLeafData{
		KPIs:    kpiCompute(),
		Headers: hdr("Job", "Schedule", "Last run", "Duration", "Next run", "Status", ""),
		Rows:    rows,
	}, true
}

func batchHPCLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	jobs, err := db.ListHPCJobs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("batchHPCLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(jobs) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(jobs))
	for i, job := range jobs {
		var statusBadge string
		switch job.Status {
		case "active", "running":
			statusBadge = g("Running")
		case "pending":
			statusBadge = b("Queued")
		case "failed":
			statusBadge = e("Failed")
		case "completed":
			statusBadge = gr("Done")
		default:
			statusBadge = y(job.Status)
		}
		coreStr := fmt.Sprintf("%d cores · %d nodes", job.Cores, job.NodeCount)
		elapsedStr := "—"
		walltimeStr := "—"
		if job.WalltimeSec > 0 {
			walltimeStr = fmt.Sprintf("%dh", job.WalltimeSec/3600)
		}
		if job.ElapsedSec > 0 {
			eh := job.ElapsedSec / 3600
			em := (job.ElapsedSec % 3600) / 60
			elapsedStr = fmt.Sprintf("%dh %dm / %s", eh, em, walltimeStr)
		}
		rows[i] = r(
			ni(job.Name, fmt.Sprintf("res-%d", job.ResourceID)),
			job.UserName,
			m(job.Framework),
			coreStr,
			elapsedStr,
			job.QueueName,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Top →</button>`,
		)
	}

	return WSLeafData{
		KPIs:    kpiCompute(),
		Headers: hdr("Job", "User", "Framework", "Cores", "Walltime", "Queue", "Status", ""),
		Rows:    rows,
	}, true
}

func storageObjectLeafFromDB(ctx context.Context, workspaceDBID int64, storageClass string) (WSLeafData, bool) {
	buckets, err := db.ListObjectBuckets(ctx, Pools.Admin, workspaceDBID, storageClass)
	if err != nil {
		log.Printf("storageObjectLeafFromDB(%s): %v", storageClass, err)
		return WSLeafData{}, false
	}
	if len(buckets) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(buckets))
	switch storageClass {
	case "standard":
		for i, bkt := range buckets {
			versionBadge := gr("Disabled")
			if bkt.VersioningEnabled {
				versionBadge = g("Enabled")
			}
			accessBadge := gr("Private")
			if bkt.AccessLevel == "public" {
				accessBadge = b("Read public")
			}
			rows[i] = r(
				ni(bkt.Name, "s3://"+bkt.Name),
				bkt.Region,
				intStr(int(bkt.ObjectCount)),
				fmt.Sprintf("%.0f GB", bkt.SizeGB),
				versionBadge,
				accessBadge,
				`<button class="lf-btn lf-btn--ghost">Browse →</button>`,
			)
		}
		return WSLeafData{
			KPIs:    kpiStorage(),
			Headers: hdr("Bucket", "Region", "Objects", "Size", "Versioning", "Access", ""),
			Rows:    rows,
		}, true

	case "archive":
		for i, bkt := range buckets {
			tierBadge := b(bkt.ArchiveTier)
			retStr := bkt.RetentionPolicy
			if retStr == "" {
				retStr = "—"
			}
			lastAccStr := "never"
			if bkt.LastAccessedAt != nil {
				days := int(time.Since(*bkt.LastAccessedAt).Hours() / 24)
				lastAccStr = fmt.Sprintf("%d days ago", days)
			}
			rows[i] = r(
				ni(bkt.Name, "s3://"+bkt.Name),
				bkt.Region,
				fmt.Sprintf("%.1f TB", bkt.SizeGB/1024),
				tierBadge,
				retStr,
				lastAccStr,
				`<button class="lf-btn lf-btn--ghost">Restore →</button>`,
			)
		}
		return WSLeafData{
			KPIs:    kpiStorage(),
			Headers: hdr("Bucket", "Region", "Size", "Tier", "Retention", "Last accessed", ""),
			Rows:    rows,
		}, true

	case "datalake":
		for i, bkt := range buckets {
			partStr := bkt.PartitionKey
			if partStr == "" {
				partStr = gr("none")
			}
			lastWriteStr := "never"
			if bkt.LastWriteAt != nil {
				lastWriteStr = bkt.LastWriteAt.Format("2006-01-02 15:04")
			}
			rows[i] = r(
				s(bkt.Name),
				m(bkt.DataFormat),
				partStr,
				fmt.Sprintf("%.0f GB", bkt.SizeGB),
				intStr(int(bkt.RecordCount)),
				lastWriteStr,
				`<button class="lf-btn lf-btn--ghost">Query →</button>`,
			)
		}
		return WSLeafData{
			KPIs:    kpiStorage(),
			Headers: hdr("Path", "Format", "Partitions", "Size", "Records", "Last write", ""),
			Rows:    rows,
		}, true
	}
	return WSLeafData{}, false
}

func storageBlockLeafFromDB(ctx context.Context, workspaceDBID int64, category string) (WSLeafData, bool) {
	vols, err := db.ListBlockVolumes(ctx, Pools.Admin, workspaceDBID, category)
	if err != nil {
		log.Printf("storageBlockLeafFromDB(%s): %v", category, err)
		return WSLeafData{}, false
	}
	if len(vols) == 0 {
		return WSLeafData{}, false
	}

	rows := make([][]template.HTML, len(vols))
	if category == "ssd" {
		for i, vol := range vols {
			encBadge := y("None")
			if vol.Encrypted {
				encBadge = g("AES-256")
			}
			attachStr := gr("unattached")
			if vol.AttachedTo != "" {
				attachStr = vol.AttachedTo
			}
			rows[i] = r(
				ni(vol.Name, fmt.Sprintf("res-%d", vol.ResourceID)),
				b(vol.VolumeType),
				fmt.Sprintf("%d GB", vol.SizeGB),
				intStr(vol.IOPS),
				fmt.Sprintf("%d MB/s", vol.ThroughputMBPS),
				attachStr,
				encBadge,
				`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
			)
		}
		return WSLeafData{
			KPIs:    kpiStorage(),
			Headers: hdr("Volume", "Type", "Size", "IOPS", "Throughput", "Attached to", "Encrypted", ""),
			Rows:    rows,
		}, true
	}

	// hdd
	for i, vol := range vols {
		encBadge := y("None")
		if vol.Encrypted {
			encBadge = g("AES-256")
		}
		attachStr := vol.AttachedTo
		if attachStr == "" {
			attachStr = gr("unattached")
		}
		rows[i] = r(
			ni(vol.Name, fmt.Sprintf("res-%d", vol.ResourceID)),
			b(vol.VolumeType),
			fmt.Sprintf("%d TB", vol.SizeGB/1024),
			fmt.Sprintf("%d MB/s", vol.ThroughputMBPS),
			bar(vol.UtilizationPct, diskBarKind(vol.UtilizationPct)),
			attachStr,
			encBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiStorage(),
		Headers: hdr("Volume", "Type", "Size", "Throughput", "Utilization", "Attached to", "Encrypted", ""),
		Rows:    rows,
	}, true
}

func storageFileNFSLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	shares, err := db.ListFileShares(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("storageFileNFSLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(shares) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(shares))
	for i, sh := range shares {
		utilPct := 0
		if sh.SizeGB > 0 {
			utilPct = int(float64(sh.UsedGB) / float64(sh.SizeGB) * 100)
		}
		statusBadge := statusBadgeStorage(sh.Status)
		rows[i] = r(
			ni(sh.Name, fmt.Sprintf("res-%d", sh.ResourceID)),
			b(sh.Protocol),
			fmt.Sprintf("%d", sh.ActiveMounts),
			fmt.Sprintf("%d / %d GB", sh.UsedGB, sh.SizeGB),
			bar(utilPct, diskBarKind(utilPct)),
			sh.Region,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiStorage(),
		Headers: hdr("Share", "Protocol", "Mounts", "Used / Quota", "Utilization", "Region", "Status", ""),
		Rows:    rows,
	}, true
}

func storageFileDistributedLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	fsList, err := db.ListDistributedFilesystems(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("storageFileDistributedLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(fsList) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(fsList))
	for i, fs := range fsList {
		utilPct := 0
		if fs.CapacityTB > 0 {
			utilPct = int(fs.UsedTB / fs.CapacityTB * 100)
		}
		statusBadge := statusBadgeStorage(fs.Status)
		rows[i] = r(
			ni(fs.Name, fmt.Sprintf("res-%d", fs.ResourceID)),
			b(fs.FSType),
			fmt.Sprintf("%d nodes", fs.NodeCount),
			fmt.Sprintf("%.1f TB", fs.CapacityTB),
			bar(utilPct, diskBarKind(utilPct)),
			fmt.Sprintf("%d", fs.ActiveClients),
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiStorage(),
		Headers: hdr("Filesystem", "Type", "Nodes", "Capacity", "Utilization", "Clients", "Status", ""),
		Rows:    rows,
	}, true
}

func storageBackupSnapshotsLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	snaps, err := db.ListBackupSnapshots(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("storageBackupSnapshotsLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(snaps) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(snaps))
	for i, sn := range snaps {
		typeBadge := b("Auto")
		if sn.SnapshotType == "manual" {
			typeBadge = y("Manual")
		}
		verifiedBadge := y("Not tested")
		if sn.VerifiedAt != nil {
			verifiedBadge = g("Verified")
		}
		retention := fmt.Sprintf("%d days", sn.RetentionDays)
		sizeStr := fmt.Sprintf("%.0f GB", sn.SizeGB)
		if sn.Status == "in-progress" {
			sizeStr = "—"
		}
		rows[i] = r(
			ni(sn.Name, fmt.Sprintf("res-%d", sn.ResourceID)),
			sn.SourceName,
			sizeStr,
			typeBadge,
			retention,
			sn.CreatedAt.Format("2 Jan 15:04"),
			verifiedBadge,
			`<button class="lf-btn lf-btn--ghost">Restore →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiStorage(),
		Headers: hdr("Snapshot", "Source", "Size", "Type", "Retention", "Created", "Status", ""),
		Rows:    rows,
	}, true
}

func storageBackupDRLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	plans, err := db.ListDRPlans(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("storageBackupDRLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(plans) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(plans))
	for i, p := range plans {
		statusBadge := statusBadgeStorage(p.Status)
		lagStr := fmt.Sprintf("%ds", p.ReplicationLagSec)
		lagBadge := lagStr
		if p.ReplicationLagSec > 120 {
			lagBadge = y(lagStr)
		}
		lastTested := "—"
		if p.LastTestedAt != nil {
			lastTested = p.LastTestedAt.Format("2 Jan 2006")
		}
		rows[i] = r(
			ni(p.Name, fmt.Sprintf("res-%d", p.ResourceID)),
			fmt.Sprintf("%s → %s", p.SourceRegion, p.TargetRegion),
			fmt.Sprintf("%dm / %dm", p.RPOMinutes, p.RTOMinutes),
			lagBadge,
			lastTested,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Failover →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiStorage(),
		Headers: hdr("Plan", "Region pair", "RPO / RTO", "Replication lag", "Last tested", "Status", ""),
		Rows:    rows,
	}, true
}

func storageBackupArchiveLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	archives, err := db.ListArchiveJobs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("storageBackupArchiveLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(archives) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(archives))
	for i, a := range archives {
		verifiedBadge := y("Not verified")
		if a.VerifiedAt != nil {
			verifiedBadge = g("SHA-256 ✓")
		}
		retYears := a.RetentionDays / 365
		retStr := fmt.Sprintf("%d yrs", retYears)
		if retYears == 0 {
			retStr = fmt.Sprintf("%d days", a.RetentionDays)
		}
		rows[i] = r(
			ni(a.Name, fmt.Sprintf("res-%d", a.ResourceID)),
			a.SourceName,
			fmt.Sprintf("%.0f GB", a.SizeGB),
			b(a.ArchiveClass),
			retStr,
			a.CreatedAt.Format("2006-01-02"),
			verifiedBadge,
			`<button class="lf-btn lf-btn--ghost">Restore →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiStorage(),
		Headers: hdr("Archive", "Source", "Size", "Class", "Retention", "Created", "Verified", ""),
		Rows:    rows,
	}, true
}

func statusBadgeStorage(status string) string {
	switch status {
	case "active", "healthy", "in sync", "snapshotting", "archiving":
		return g(status)
	case "degraded", "failed", "error":
		return e(status)
	case "provisioning", "in-progress":
		return y(status)
	case "disabled", "paused", "stopped":
		return gr(status)
	default:
		return status
	}
}

func diskBarKind(pct int) string {
	if pct >= 90 {
		return "red"
	}
	if pct >= 75 {
		return "amber"
	}
	return ""
}

func cpuBarKind(pct int) string {
	if pct >= 85 {
		return "red"
	}
	if pct >= 65 {
		return "amber"
	}
	return ""
}

// ──────────────────────────────────────────────────────────────────────────
// REGISTRY: per-leaf KPIs, table headers and row data. Compact data-only.
// ──────────────────────────────────────────────────────────────────────────

func leafRegistry() map[string]WSLeafData {
	return map[string]WSLeafData{
		// ─── COMPUTE ────────────────────────────────────────────────
		"compute/vms/general": {
			KPIs:    kpiCompute(),
			Headers: hdr("Name", "Type", "Region", "Status", "CPU", "Memory", ""),
			Rows:    [][]template.HTML{},
		},
		"compute/vms/gpu": {
			KPIs: []WSLeafKPI{
				k("GPU instances", "4", "A100 + H100 + L40S"),
				k("GPU hours (MTD)", "1,284", "$3,421 spent"),
				k("VRAM in use", "242 / 320 GB", "75%"),
				k("Avg utilisation", "82%", "high — efficient"),
			},
			Headers: hdr("Name", "GPU", "Type", "Region", "Status", "GPU util", "VRAM", ""),
			Rows:    [][]template.HTML{},
			Note:    "💡 Idle GPU instances cost $1.20/hr — consider stopping them.",
		},
		"compute/vms/spot": {
			KPIs: []WSLeafKPI{
				k("Spot instances", "18", "vs. 24 on-demand"),
				k("Savings (MTD)", "$2,840"+up("+18%"), "vs. on-demand"),
				k("Reclaim rate", "0.4 / day", "very low"),
				k("Avg discount", "78%", "off list price"),
			},
			Headers: hdr("Name", "Pool", "Bid", "Region", "Status", "Up since", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚡ 1 instance reclaimed today. Enable capacity rebalancing to auto-replace interrupted workloads.",
		},
		"compute/containers/kubernetes": {
			KPIs: []WSLeafKPI{
				k("Clusters", "3", "1 prod · 1 stg · 1 ml"),
				k("Nodes", "19", "across 3 clusters"),
				k("Pods running", "183"+up("+12"), "of 220 capacity"),
				k("Deployments", "29", "all healthy"),
			},
			Headers: hdr("Cluster", "Version", "Nodes", "Region", "Status", "Pods", "CPU", ""),
			Rows:    [][]template.HTML{},
			Note:    "🔄 ml-cluster-gpu is upgrading to v1.29 — rolling restart in progress, no pod disruption expected.",
		},
		"compute/containers/registry": {
			KPIs: []WSLeafKPI{
				k("Repositories", "18", "across 3 namespaces"),
				k("Images", "284", "tags total"),
				k("Storage", `12.4 <span style="font-size:13px;color:var(--lf-dim)">GB</span>`, "compressed"),
				k("Pulls (24h)", "8,421"+up("+18%"), "vs yesterday"),
			},
			Headers: hdr("Repository", "Tags", "Size", "Last scan", "Vulns", ""),
			Rows:    [][]template.HTML{},
			Note:    "🛡️ 1 critical vulnerability in legacy-monolith — update base image to resolve CVE-2024-38816.",
		},
		"compute/containers/instances": {
			KPIs: []WSLeafKPI{
				k("Containers", "84", "running now"),
				k("Cold starts", "12 / hr", "p99 1.2s"),
				k("Memory", "142 / 320 GB", "44% utilisation"),
				k("Spend (MTD)", "$284"+up("+12%"), "vs. last month"),
			},
			Headers: hdr("Name", "Image", "CPU/Mem", "Region", "Status", "Restarts", "Up since", ""),
			Rows:    [][]template.HTML{},
		},
		"compute/hosting/web": {
			KPIs: []WSLeafKPI{
				k("Apps hosted", "9", "PHP · Node · Python · Ruby"),
				k("Visits (24h)", "182k"+up("+8%"), "vs. previous day"),
				k("SSL certificates", "9/9", "all auto-renewed"),
				k("Avg TTFB", `92 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("-11ms"), "p50 global"),
			},
			Headers: hdr("Domain", "Stack", "Region", "Status", "CPU", "Visits (24h)", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ legacy.acme.io runs Drupal 9 (EOL Nov 2026) — begin migration to avoid unsupported runtime.",
		},
		"compute/hosting/static": {
			KPIs: []WSLeafKPI{
				k("Sites", "12", "all Git-connected"),
				k("Builds (24h)", "48"+up("+14%"), "98% success rate"),
				k("CDN cache hit", "97%", "across 142 PoPs"),
				k("Bandwidth (MTD)", `412 <span style="font-size:13px;color:var(--lf-dim)">GB</span>`+down("-6%"), "vs. last month"),
			},
			Headers: hdr("Site", "Framework", "Branch", "Last deploy", "Build time", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "🔴 landing-campaign build failed — dependency conflict in gatsby-plugin-image. Check build logs.",
		},
		"compute/serverless/functions": {
			KPIs: []WSLeafKPI{
				k("Functions", "28", "4 runtimes · 3 regions"),
				k("Invocations (24h)", "284k"+up("+12%"), "vs. previous day"),
				k("Avg duration", `142 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`, "p95 · budget 200ms"),
				k("Error rate", "0.04%"+down("-0.01%"), "well within SLO"),
			},
			Headers: hdr("Function", "Runtime", "Trigger", "Invocations/day", "p95 latency", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ webhook-stripe p95 latency exceeds 200ms target — consider increasing memory or optimising cold start.",
		},
		"compute/serverless/edge": {
			KPIs: []WSLeafKPI{
				k("Edge functions", "8", "deployed to 142 PoPs"),
				k("Requests (24h)", "4.8M"+up("+22%"), "across all functions"),
				k("p50 latency", `12 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`, "measured at PoP"),
				k("Error rate", "0.001%", "0 user-facing errors"),
			},
			Headers: hdr("Function", "Path", "PoPs", "Requests/s", "p99 latency", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚡ img-transform p99 at 42ms — consider pre-generating common sizes to reduce compute at the edge.",
		},
		"compute/batch/jobs": {
			KPIs: []WSLeafKPI{
				k("Scheduled jobs", "18", "12 daily · 6 weekly"),
				k("Runs (24h)", "47", "44 passed · 3 failed"),
				k("Avg duration", "12m 18s"+down("-2m"), "improving trend"),
				k("Spend (MTD)", "$142"+down("-8%"), "spot-backed"),
			},
			Headers: hdr("Job", "Schedule", "Last run", "Duration", "Next run", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "🔴 report-roll-up failed — exit code 1: timeout connecting to analytics-db. Retry or check db status.",
		},
		"compute/batch/hpc": {
			KPIs: []WSLeafKPI{
				k("Active jobs", "6"+up("+2"), "MPI + Slurm workloads"),
				k("Cores in use", "1,248 / 2,048", "61% fleet utilisation"),
				k("Queue wait", `2.4 <span style="font-size:13px;color:var(--lf-dim)">min</span>`+down("-0.8m"), "avg across queues"),
				k("Spend (MTD)", "$8,420"+down("-6%"), "$0.012 / core-hr"),
			},
			Headers: hdr("Job", "User", "Framework", "Cores", "Walltime", "Queue", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "🔵 2 jobs queued — estimated start in ~3 min based on current completion rates. Fleet at 61% utilisation.",
		},

		// ─── STORAGE ────────────────────────────────────────────────
		"storage/object/standard": {
			KPIs: []WSLeafKPI{
				k("Buckets", "12", "across 2 regions"),
				k("Objects", "1.2M"+up("+42k"), "vs. last week"),
				k("Size", `1.4 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`, "$32.40 / month"),
				k("Requests (24h)", "284k"+up("+8%"), "GET 78% · PUT 22%"),
			},
			Headers: hdr("Bucket", "Region", "Objects", "Size", "Versioning", "Access", ""),
			Rows:    [][]template.HTML{},
			Note:    "💡 acme-temp has no lifecycle policy — 1,284 objects with no expiry. Consider adding a 30-day TTL to reduce costs.",
		},
		"storage/object/archive": {
			KPIs: []WSLeafKPI{
				k("Archive buckets", "4", "Glacier + Deep Archive"),
				k("Size", `12.8 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+up("+1.2 TB"), "$32 / TB / mo"),
				k("Restore SLA", "1–5 hr", "expedited 1–5 min"),
				k("Savings vs std", "−82%"+down("−$186/mo"), "vs. standard tier"),
			},
			Headers: hdr("Bucket", "Region", "Size", "Tier", "Retention", "Last accessed", ""),
			Rows:    [][]template.HTML{},
			Note:    "🔒 acme-legal-hold is under legal hold — deletion and lifecycle transitions are blocked until the hold is released.",
		},
		"storage/object/datalake": {
			KPIs: []WSLeafKPI{
				k("Lake size", `8.4 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+up("+320 GB"), "vs. last month"),
				k("Tables registered", "284"+up("+9"), "via Glue catalog"),
				k("Queries (24h)", "1,842"+up("+14%"), "via Athena / Trino"),
				k("Scan cost (24h)", "$240"+down("-18%"), "partitioning savings"),
			},
			Headers: hdr("Path", "Format", "Partitions", "Size", "Records", "Last write", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ logs/unpartitioned/ has no partition key — full scans cost ~$20 each. Add a date partition to cut scan costs by 95%.",
		},
		"storage/block/ssd": {
			KPIs: []WSLeafKPI{
				k("SSD volumes", "32", "NVMe-backed · gp3 + io2"),
				k("IOPS provisioned", "184k / 200k", "92% of account limit"),
				k("Throughput", `4.2 <span style="font-size:13px;color:var(--lf-dim)">GB/s</span>`+up("+0.4"), "aggregate read+write"),
				k("Avg latency", `0.42 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("-0.08"), "p99 read · improving"),
			},
			Headers: hdr("Volume", "Type", "Size", "IOPS", "Throughput", "Attached to", "Encrypted", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ vol-snapshot-stage is unattached and unencrypted — costs $16/mo idle. Attach, encrypt, or delete to reduce spend and risk.",
		},
		"storage/block/hdd": {
			KPIs: []WSLeafKPI{
				k("HDD volumes", "8", "4 st1 · 4 sc1"),
				k("Total size", `48 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+up("+8 TB"), "grew from archive ingest"),
				k("Throughput", `1.8 <span style="font-size:13px;color:var(--lf-dim)">GB/s</span>`+up("+0.2"), "aggregate sequential read"),
				k("Cost savings", `$2,640 <span style="font-size:13px;color:var(--lf-dim)">/mo</span>`+down("−72%"), "vs. equivalent SSD"),
			},
			Headers: hdr("Volume", "Type", "Size", "Throughput", "Utilization", "Attached to", "Encrypted", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ vol-archive-2024 is at 91% capacity — expand or migrate cold data to object archive. vol-media-raw lacks encryption; enable AES-256 to meet compliance.",
		},
		"storage/file/nfs": {
			KPIs: []WSLeafKPI{
				k("File shares", "5", "3 NFSv4 · 2 SMB 3.0"),
				k("Total mounts", "84"+up("+6"), "across 18 hosts"),
				k("Used / Allocated", `1.2 / 4 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`, "30% utilization"),
				k("Throughput", `420 <span style="font-size:13px;color:var(--lf-dim)">MB/s</span>`+up("+40"), "p95 aggregate"),
			},
			Headers: hdr("Share", "Protocol", "Mounts", "Used / Quota", "Utilization", "Region", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ dev-shared is at 82% quota — expand quota or archive stale files to avoid mount failures. legacy-home has 0 active mounts; consider decommissioning to save $24/mo.",
		},
		"storage/file/distributed": {
			KPIs: []WSLeafKPI{
				k("Filesystems", "3", "2 Lustre · 1 GlusterFS"),
				k("Capacity", `64 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+up("+16 TB"), "expanded for HPC workload"),
				k("Throughput", `14.8 <span style="font-size:13px;color:var(--lf-dim)">GB/s</span>`+up("+2.8"), "parallel read aggregate"),
				k("Active clients", "82"+up("+18"), "across HPC + research"),
			},
			Headers: hdr("Filesystem", "Type", "OSTs / Bricks", "Capacity", "Utilization", "Clients", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ hpc-burst is at 94% capacity — purge completed job scratch data or expand OSTs. Sustained usage above 90% degrades metadata performance.",
		},
		"storage/backup/snapshots": {
			KPIs: []WSLeafKPI{
				k("Active snapshots", "284"+up("+12"), "218 auto · 66 manual"),
				k("Total size", `4.2 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+down("−180 GB"), "dedup ratio 3.4:1"),
				k("Cost (MTD)", `$68 <span style="font-size:13px;color:var(--lf-dim)">/mo</span>`+down("−8%"), "incremental storage only"),
				k("Restore tests", "94%"+up("+2%"), "last tested 3d ago"),
			},
			Headers: hdr("Snapshot", "Source", "Size", "Type", "Retention", "Created", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ snap-staging-stale is 84 days old with no expiry and has never been restore-tested — verify or delete to avoid relying on an untested backup. Costs $3.20/mo.",
		},
		"storage/backup/dr": {
			KPIs: []WSLeafKPI{
				k("DR plans", "4", "2 cross-region · 1 multi-AZ · 1 cold"),
				k("RPO achieved", `8 <span style="font-size:13px;color:var(--lf-dim)">min</span>`+down("−7 min"), "better than 15m target"),
				k("RTO achieved", `42 <span style="font-size:13px;color:var(--lf-dim)">min</span>`+down("−18 min"), "better than 1h target"),
				k("Last drill", "12d ago", g("4/4 passed")),
			},
			Headers: hdr("Plan", "Region pair", "RPO / RTO", "Replication lag", "Last tested", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ dr-files-cross-region lag is 4m 12s (target < 2m) — check WAN throughput and consider enabling parallel replication. dr-cold-yearly hasn't been tested in 38 days; schedule a drill.",
		},
		"storage/backup/archive": {
			KPIs: []WSLeafKPI{
				k("Archives", "248"+up("+14"), "182 scheduled · 66 manual"),
				k("Archived size", `12.4 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+up("+1.2 TB"), "$96/mo · $0.0077/GB"),
				k("Compliance retention", "7 yrs", "SEC 17a-4(f) · HIPAA"),
				k("Integrity checks", "2d ago", g("248/248 passed")),
			},
			Headers: hdr("Archive", "Source", "Size", "Class", "Retention", "Created", "Verified", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ db-export-legacy is on Standard-IA ($0.0125/GB) instead of Glacier Deep ($0.00099/GB) — moving it would save ~$1.50/mo. It also hasn't been integrity-verified; schedule a hash check.",
		},

		// ─── DATABASES ──────────────────────────────────────────────
		"databases/relational/mysql-pg": {
			KPIs: []WSLeafKPI{
				k("PG + MySQL clusters", "5"+up("+1"), "3 PG · 2 MySQL"),
				k("Connections", "284 / 800", "35% pool utilization"),
				k("QPS", "8.2k"+up("+1.4k"), "6.8k reads · 1.4k writes"),
				k("Replication lag", `12 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("−3 ms"), g("Healthy")),
			},
			Headers: hdr("Cluster", "Engine", "Plan", "Storage", "Connections", "QPS", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ legacy-mysql-57 is running MySQL 5.7 which reaches EOL Oct 2026 — plan migration to MySQL 8.0 or PostgreSQL. Only 2 active connections; good candidate for consolidation into billing-mysql.",
		},
		"databases/relational/sql-server": {
			KPIs: []WSLeafKPI{
				k("SQL Server instances", "3", "1 AG group · 1 standalone"),
				k("Connections", "112 / 500"+down("−18"), "peak yesterday 148"),
				k("Wait stats", g("Healthy"), "top wait: ASYNC_NETWORK_IO 0.4 ms"),
				k("Backup chain", "PITR 7d", g("Unbroken")+" · weekly full Sun 02:00"),
			},
			Headers: hdr("Instance", "Edition", "Plan", "Storage", "Utilization", "Connections", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ reporting-standalone is on Standard edition with 6 connections — evaluate whether it can be consolidated into a read replica of the AG to reduce licensing costs (~$1,800/yr savings).",
		},
		"databases/relational/cloud-native": {
			KPIs: []WSLeafKPI{
				k("Aurora clusters", "3", "2 PG-compat · 1 Serverless v2"),
				k("ACUs in use", "12.4"+up("+2.1"), "auto-scales 1–32 range"),
				k("IOPS", "8.4k"+up("+1.2k"), "burst credits available"),
				k("Cost savings", `$680 <span style="font-size:13px;color:var(--lf-dim)">/mo</span>`+down("−32%"), "vs. self-managed RDS"),
			},
			Headers: hdr("Cluster", "Engine", "ACUs (cur / max)", "Storage", "Region", "Connections", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ aurora-serverless-v2 has 0 connections and is consuming idle ACUs ($0.12/ACU-hr) — consider pausing or setting min ACU to 0 if the workload is truly dormant. Saves ~$43/mo.",
		},
		"databases/nosql/document": {
			KPIs: []WSLeafKPI{
				k("Document DBs", "3", "2 Atlas · 1 self-managed"),
				k("Collections", "84"+up("+6"), "across 12 databases"),
				k("Documents", "412M"+up("+28M"), `8.4 <span style="font-size:13px;color:var(--lf-dim)">TB</span> on disk`),
				k("Ops/s", "12.4k"+up("+1.8k"), "p99 18 ms · p50 4 ms"),
			},
			Headers: hdr("Cluster", "Version", "Topology", "Disk", "Ops/s", "Region", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ analytics-events is running a background compaction that may spike latency for 2–4 hours — avoid bulk writes during this window. session-store is on MongoDB 6.0 (EOL Aug 2026); plan upgrade to 7.0.",
		},
		"databases/nosql/keyvalue": {
			KPIs: []WSLeafKPI{
				k("Tables", "8"+up("+1"), "DynamoDB · 4 on-demand · 4 provisioned"),
				k("Items", "284M"+up("+12M"), `92 <span style="font-size:13px;color:var(--lf-dim)">GB</span> total storage`),
				k("Throughput", "12k / 4k"+up("+1.2k"), "peak RCU / WCU (auto-scaled)"),
				k("Monthly cost", "$342"+down("−8%"), "vs. last month · on-demand savings"),
			},
			Headers: hdr("Table", "Items", "Size", "Indexes", "RCU / WCU", "Mode", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ rate_limits is on provisioned mode with consistent 80%+ utilization — consider switching to on-demand to avoid throttling during traffic spikes. event_log has 3 GSIs contributing $128/mo; evaluate whether the oldest GSI (created 14 months ago) is still queried.",
		},
		"databases/nosql/wide-column": {
			KPIs: []WSLeafKPI{
				k("Cluster nodes", "12"+up("+2"), "1 ring · 3 DCs · Cassandra 4.1"),
				k("Keyspaces", "8", "18 tables · RF ≥ 3 everywhere"),
				k("Throughput", "48k / 12k"+up("+6k"), "reads / writes per sec"),
				k("Latency", `4 <span style="font-size:13px;color:var(--lf-dim)">ms</span> p99`+down("−0.8"), "reads · p99 writes 2 ms"),
			},
			Headers: hdr("Keyspace", "Tables", "Replication", "Size", "Read/s", "Write/s", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ logs_ingest uses SimpleStrategy (single-DC aware) while the cluster spans 3 DCs — switch to NetworkTopologyStrategy to guarantee cross-DC redundancy. Compaction is running; expect elevated disk I/O for ~1 hour.",
		},
		"databases/nosql/graph": {
			KPIs: []WSLeafKPI{
				k("Graph databases", "2", "Neo4j 5.17 · Neptune (Gremlin)"),
				k("Nodes / edges", "12.4M / 84M"+up("+4.2M"), "6.8× edge density"),
				k("Queries (24 h)", "184k"+up("+12%"), "p50 4 ms · p95 18 ms"),
				k("Storage", `22 <span style="font-size:13px;color:var(--lf-dim)">GB</span>`+down("−3%"), "after last compaction"),
			},
			Headers: hdr("Graph", "Engine", "Nodes", "Edges", "Queries/s", "Latency p95", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ fraud-network p95 latency (22 ms) exceeds the 15 ms SLO — the instance is still warming after last night's snapshot restore. If latency doesn't drop below 15 ms within 2 hours, consider scaling the Neptune instance from db.r6g.large to db.r6g.xlarge.",
		},
		"databases/memory/cache": {
			KPIs: []WSLeafKPI{
				k("Cache clusters", "3", "2 Redis 7.2 · 1 Memcached 1.6"),
				k("Memory", "84 / 128 GB"+down("−4"), bar(66, "")+" 66% used"),
				k("Hit rate", "97.4%"+up("+0.6"), g("Excellent")+" · miss rate 2.6%"),
				k("Evictions/s", "0", g("No pressure")+" · peak 12/s last week"),
			},
			Headers: hdr("Cluster", "Engine", "Topology", "Memory", "Hit rate", "Ops/s", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ legacy-memcached hit rate (92.8%) is below the 95% target — a recent deployment changed serialization format, causing cold misses. Consider flushing and warming the cache. Also evaluate migrating from Memcached 1.6 (no TLS) to Redis 7.2 for encryption-in-transit compliance.",
		},
		"databases/memory/session": {
			KPIs: []WSLeafKPI{
				k("Active sessions", "48,210"+up("+3,412"), "across 3 stores · web + admin + OAuth"),
				k("Avg TTL", "30 min", "rolling extension on activity"),
				k("Memory", "497 MB / 5.3 GB"+down("−22 MB"), bar(9, "")+" 9.4% used"),
				k("Lookups/s", "2,841"+up("+180"), "p99 0.4 ms · p50 0.1 ms"),
			},
			Headers: hdr("Store", "Engine", "Sessions", "Memory", "TTL", "Lookups/s", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ admin-session uses an 8-hour TTL with no activity-based extension — stale sessions remain in memory long after admins leave. Consider adding sliding expiry or reducing TTL to 2 hours to match the admin inactivity timeout policy.",
		},
		"databases/analytics/warehouse": {
			KPIs: []WSLeafKPI{
				k("Warehouses", "2"+up("+0"), "Snowflake + Redshift · 2 regions"),
				k("Queries (24h)", "12,841"+up("+2,048"), "avg duration 4.2 s · peak 142 concurrent"),
				k("Bytes scanned", `48 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+down("−3 TB"), "est. cost $240 · $18.70/TB"),
				k("Slot utilisation", "68%"+up("+4%"), bar(68, "")+" avg over 24 h"),
			},
			Headers: hdr("Warehouse", "Engine", "Size", "Storage", "Queries/24h", "Avg duration", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ ml-features-redshift query queue depth averaged 14 over the past 6 hours — consider adding concurrency scaling or moving exploratory workloads to Snowflake where auto-scale absorbs burst demand without queuing.",
		},
		"databases/analytics/timeseries": {
			KPIs: []WSLeafKPI{
				k("TS databases", "3"+up("+1"), "InfluxDB × 2 + TimescaleDB · 2 regions"),
				k("Ingest rate", "412k pts/s"+up("+38k"), "combined write throughput"),
				k("Active series", "8.4M"+up("+620k"), "within retention window"),
				k("Compression", "94×"+up("+2×"), "disk savings vs. raw · 14 TB → 149 GB"),
			},
			Headers: hdr("Database", "Engine", "Series", "Retention", "Ingest/s", "Storage", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ financial-ticks is at 87% of its storage quota (39 / 45 GB) — current ingest rate will exhaust it in ~11 days. Consider enabling downsampling for data older than 90 days or upgrading the storage tier to avoid write rejections.",
		},
		"databases/analytics/search": {
			KPIs: []WSLeafKPI{
				k("Search clusters", "2"+up("+0"), "OpenSearch + Elasticsearch · 8 data nodes"),
				k("Indices", "284"+up("+12"), "218 active · 66 read-only"),
				k("Documents", "1.2B"+up("+48M"), "indexed across all clusters"),
				k("Queries/s", "8.4k"+up("+640"), "p99 18 ms · p50 3.2 ms"),
			},
			Headers: hdr("Cluster", "Engine", "Nodes", "Indices", "Documents", "Storage", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ product-search cluster is yellow — 2 replica shards are unassigned after a node replacement 45 minutes ago. Shard relocation is in progress (ETA ~20 min). No impact on search availability but writes to affected indices may be slower until rebalancing completes.",
		},

		// ─── NETWORKING ─────────────────────────────────────────────
		"networking/vnet/vpc": {
			KPIs: []WSLeafKPI{
				k("VPCs", "4"+up("+0"), "across 4 regions · fully meshed via peering"),
				k("Subnets", "16"+up("+0"), "10 private · 6 public · /24 standard"),
				k("IPs available", "2,814 / 4,096"+down("−128"), bar(31, "")+" 31% utilised"),
				k("Flow logs", g("Enabled"), "all VPCs · 14 d retention · S3 export"),
			},
			Headers: hdr("VPC", "CIDR", "Region", "Subnets", "IPs used", "Peerings", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ prod-vpc is at 55% IP utilisation in the eu-central-1a private subnet (842 of 1,536 IPs) — auto-scaling events during traffic spikes could exhaust available addresses. Consider adding a secondary CIDR block (e.g., 10.4.0.0/16) or converting to IPv6 dual-stack.",
		},
		"networking/vnet/peering": {
			KPIs: []WSLeafKPI{
				k("Peerings", "6"+up("+0"), "4 VPC peering + 2 transit gateway attachments"),
				k("Cross-region", "4", "eu-central ↔ us-east ↔ ap-southeast"),
				k("Throughput", `1.2 <span style="font-size:13px;color:var(--lf-dim)">Gbps</span>`+up("+0.3"), "p95 last 24 h · no throttle events"),
				k("Data transfer", `412 <span style="font-size:13px;color:var(--lf-dim)">GB</span>`+up("+48 GB"), "this month · est. $36.90 cross-region"),
			},
			Headers: hdr("Peering", "From", "To", "CIDR exchange", "Bandwidth", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ ml-to-staging peering is configured one-way only (ml-vpc → staging-vpc) — staging cannot initiate connections back to ml-vpc. If the ML pipeline requires bidirectional communication (e.g., callback webhooks), enable the return route in the staging-vpc route table.",
		},
		"networking/vnet/nsg": {
			KPIs: []WSLeafKPI{
				k("Security groups", "24"+up("+2"), "across 4 VPCs · 18 prod · 4 staging · 2 ML"),
				k("Total rules", "412"+up("+14"), "286 ingress · 126 egress · avg 17 per SG"),
				k("Wide-open rules", g("0"), "no 0.0.0.0/0 → :22 · last scanned 4 h ago"),
				k("Last audit", g("2 d ago"), "all 24 SGs compliant · next audit in 5 d"),
			},
			Headers: hdr("Security group", "VPC", "Ingress rules", "Egress rules", "Attached to", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ sg-staging-loose allows all TCP traffic from the entire staging VPC CIDR (10.1.0.0/16) — this passes audit because staging is non-production, but 12 instances are attached. Consider scoping ingress to specific ports (80, 443, 8080) and source security groups to limit lateral movement if a staging host is compromised.",
		},
		"networking/lb/app": {
			KPIs: []WSLeafKPI{
				k("Application LBs", "5"+up("+0"), "L7 HTTP/HTTPS · 3 eu-central · 1 us-east · 1 ap-southeast"),
				k("Targets", "84"+down("−1"), "76 healthy · "+e("8 unhealthy")+" · 4 draining"),
				k("Requests/s", `12.8k`+up("+1.2k"), "across all ALBs · peak 18.4k today at 14:22"),
				k("p99 latency", `42 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("−3"), "SLO < 100 ms · "+bar(42, "")+" 42% of budget"),
			},
			Headers: hdr("Load balancer", "Region", "Listeners", "Targets", "Healthy", "RPS", "p99", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ ml-inference-alb has 7 unhealthy targets in ap-southeast-1 — health checks on :8501/health are timing out after 5 s. The affected instances (i-7a…, i-7b…, i-7c…) were launched 12 min ago by an auto-scaling event and may still be loading model weights. If targets don't recover within 10 min, check instance logs for OOM errors.",
		},
		"networking/lb/network": {
			KPIs: []WSLeafKPI{
				k("Network LBs", "3"+up("+0"), "L4 TCP/UDP · 2 eu-central · 1 us-east"),
				k("New conn/s", `84k`+up("+12k"), "TCP 71k · UDP 13k · peak 102k today"),
				k("Active connections", `412k`+up("+28k"), "long-lived persistent · avg duration 4.2 min"),
				k("p50 latency", `<1 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`, "wire-speed passthrough · no TLS termination"),
			},
			Headers: hdr("Load balancer", "Region", "Protocol", "Targets", "Healthy", "Active conns", "Bandwidth", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ grpc-nlb is handling 284k concurrent connections on 6 targets (~47k each) — the per-instance connection limit is 65k. If a single target fails, the remaining 5 would each absorb ~57k connections, approaching the limit. Consider adding 2 targets as headroom or enabling cross-zone load balancing to distribute more evenly.",
		},
		"networking/lb/global": {
			KPIs: []WSLeafKPI{
				k("Global LBs", "2"+up("+0"), "anycast endpoints · premium network tier"),
				k("Backend regions", "4", "eu-central · us-east · us-west · ap-southeast"),
				k("Failovers (30 d)", g("0"), "ultra-stable · last failover 94 d ago"),
				k("Avg geo-latency", `18 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("−2"), "users routed to nearest PoP · 142 edge locations"),
			},
			Headers: hdr("Endpoint", "Anycast IP", "Backends", "Geo-policy", "RPS", "Health", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ rt.acme.io only has backends in 3 of 4 regions — ap-southeast users are routed to us-west (140 ms round-trip). Adding a backend in ap-southeast-1 would cut latency for APAC users to ~20 ms and improve failover coverage if the us-west region degrades.",
		},
		"networking/dns/manage": {
			KPIs: []WSLeafKPI{
				k("DNS zones", "8"+up("+0"), "5 public · 3 private · 4 nameservers per zone"),
				k("Records", "284"+up("+6"), "142 A · 48 CNAME · 32 MX · 62 other"),
				k("Queries (24 h)", `4.2M`+up("+12%"), "3.8M answered from cache · "+bar(10, "")+" 10% miss rate"),
				k("Avg resolution", `12 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("−1"), "p95 · anycast from 4 global PoPs"),
			},
			Headers: hdr("Zone", "Type", "Records", "Queries (24 h)", "TTL range", "DNSSEC", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ legacy.acme.com has DNSSEC disabled — this zone serves public DNS for 8k queries/day and is vulnerable to spoofing. Enable DNSSEC signing and publish DS records with your registrar. Also review the 12 records for stale entries — 3 A records point to decommissioned IPs (203.0.113.x range).",
		},
		"networking/dns/domains": {
			KPIs: []WSLeafKPI{
				k("Domains owned", "12"+up("+2"), "6 TLDs · .io .com .dev .net .org .co"),
				k("Auto-renew", "11 / 12"+down("−1"), "acme-legacy.net set to manual — "+bar(92, "")+" 92% covered"),
				k("Expiring < 90 d", "1"+up("+1"), y("acme-spare.io")+" expires 2026-07-04"),
				k("Annual cost", "$284"+down("−$18"), "avg $23.67 per domain · 2 premium TLDs"),
			},
			Headers: hdr("Domain", "Registrar", "Registered", "Expires", "Auto-renew", "WHOIS privacy", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ acme-legacy.net has auto-renew disabled — it expires 2027-03-22 and hosts MX records for legacy email forwarding. Enable auto-renew or set a calendar reminder. Also, acme.dev is registered with Google Domains and not yet transferred to Platform Registrar — transfer to consolidate billing and enable unified WHOIS privacy.",
		},
		"networking/cdn/static": {
			KPIs: []WSLeafKPI{
				k("Distributions", "5"+up("+1"), "3 S3 origins · 1 GCS · 1 custom server"),
				k("Cache hit ratio", "97.2%"+up("+0.8%"), bar(97, "")+" 3.8M hits / 3.9M requests (24 h)"),
				k("Bandwidth (24 h)", `1.2 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+up("+8%"), "$8.40 egress · 142 PoPs across 6 continents"),
				k("Avg TTFB", `18 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("−3"), "p95 from edge · Brotli + gzip on 92% of responses"),
			},
			Headers: hdr("Distribution", "Origin", "Path", "Hit ratio", "Bandwidth / mo", "SSL", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ api-static.acme.io uses a custom origin server with a manually managed SSL certificate expiring in 14 days. Switch to ACM auto-renewal or upload a new cert before 2026-05-03. Also, its cache hit ratio (82%) is below the 90% target — review Cache-Control headers on the origin to ensure static assets set max-age ≥ 86400.",
		},
		"networking/cdn/video": {
			KPIs: []WSLeafKPI{
				k("Live channels", "3"+up("+1"), "2 HLS · 1 DASH · adaptive bitrate 360p–4K"),
				k("VOD assets", "1,284"+up("+42"), "1,180 transcoded · 104 queued · "+bar(92, "")+" 92% ready"),
				k("Concurrent viewers", "8,421"+up("+18%"), "peak 12.4k this week · avg session 14 min"),
				k("Bandwidth (24 h)", `4.8 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+up("+22%"), "$33.60 egress · 62% from NA · 28% EU · 10% APAC"),
			},
			Headers: hdr("Channel / Library", "Protocol", "Resolution", "Viewers", "Bitrate", "Uptime", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ live-gaming channel is experiencing elevated buffering (1.6% rebuffer rate vs 0.3% target) — the DASH origin in eu-west-1 is at 94% CPU. Consider scaling the transcoder fleet or reducing the 1080p60 ladder to 1080p30 during peak hours. 104 VOD assets are queued for transcoding; estimated completion in ~3 h at current throughput.",
		},
		"networking/cdn/edge": {
			KPIs: []WSLeafKPI{
				k("Edge rules", "24"+up("+3"), "18 cache policies · 4 transforms · 2 redirects"),
				k("Edge functions", "8"+up("+2"), "12 KB total · avg 2.4 ms exec · "+bar(6, "")+" 6% CPU budget"),
				k("p50 edge latency", `8 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("−2"), "p95 22 ms · p99 48 ms · 142 PoPs"),
				k("Cache efficiency", "94.1%"+up("+1.2%"), bar(94, "")+" stale-while-revalidate on 82% of rules"),
			},
			Headers: hdr("Rule", "Pattern", "Action", "Invocations (24 h)", "Avg latency", "Errors", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ legacy-rewrite rule is producing 38 errors in the last 24 h — the target path /docs/api/v1 returns 404 for 3 rewritten URLs. Update the rewrite map or add a fallback rule. Also, img-resize-otf is averaging 4.8 ms execution time (budget is 5 ms) — consider pre-generating common resize variants to reduce edge compute pressure.",
		},
		"networking/private/vpn": {
			KPIs: []WSLeafKPI{
				k("VPN gateways", "3"+up("+1"), "2 IPsec IKEv2 · 1 WireGuard · HA pair in eu-west-1"),
				k("Active tunnels", "6 / 8"+up("+1"), "6 established · 1 standby · 1 down · "+bar(75, "")+" 75%"),
				k("Throughput", `1.2 <span style="font-size:13px;color:var(--lf-dim)">Gbps</span>`+up("+8%"), "p95 aggregate · 892 Mbps IPsec · 412 Mbps WireGuard"),
				k("Connected clients", "84"+up("+12"), "via WG userspace · 3 admin · 81 developer"),
			},
			Headers: hdr("Gateway", "Type", "Peer / Clients", "Throughput", "Uptime", "IKE SA", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ branch-vpn-nyc tunnel is DOWN — IKE SA negotiation failed 2 h ago after the peer rotated its pre-shared key. Coordinate with the NYC office to update the PSK on both sides, then run a manual re-key. The backup BGP path via corp-vpn-backup is routing NYC traffic through Frankfurt, adding ~82 ms latency to US-East services.",
		},
		"networking/private/direct": {
			KPIs: []WSLeafKPI{
				k("Direct Connect", "3"+up("+1"), "2× 10 Gbps LAG (FRA-6) · 1× 10 Gbps single (DC11)"),
				k("Cross-connects", "4"+up("+0"), "2 colocations · Equinix FRA-6 + DC11 · redundant paths"),
				k("Throughput", `8.4 <span style="font-size:13px;color:var(--lf-dim)">Gbps</span>`+up("+6%"), "p95 aggregate · 6.1 Gbps FRA-6 LAG · 2.3 Gbps DC11 · "+bar(42, "")+" 42% capacity"),
				k("Avg latency", `1.2 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("−0.1"), "to nearest AZ · 0.8 ms FRA · 1.8 ms DC11 · sub-2 ms SLA met"),
			},
			Headers: hdr("Circuit", "Location", "Bandwidth", "VLAN / BGP", "Provider", "Throughput", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ dx-ash-dc11-bk is still provisioning (ETA 3 business days) — until then DC11 has no redundancy. If dxcon-2c3d4e fails, US-East traffic will reroute through FRA-6, adding ~85 ms latency. The FRA-6 LAG pair is running at 61% combined capacity during peak hours (6.1 Gbps on 10+10 Gbps) — monitor for growth before the new DC11 circuit is live.",
		},
		"networking/private/sdwan": {
			KPIs: []WSLeafKPI{
				k("SD-WAN sites", "12"+up("+2"), "5 hub · 7 branch · full mesh topology · 3 regions"),
				k("Underlay paths", "3"+up("+0"), "1× MPLS (100 Mbps) · 2× broadband ISP · "+bar(75, "")+" 75% path diversity"),
				k("Apps steered", "84"+up("+8"), "42 voice/video (DSCP EF) · 28 business-critical · 14 bulk transfer"),
				k("Path failover", `<1 <span style="font-size:13px;color:var(--lf-dim)">sec</span>`, "sub-second convergence · 0 outages (30 d) · BFD-assisted"),
			},
			Headers: hdr("Site", "Role", "Underlays", "Active paths", "Bandwidth", "QoS class", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ factory-stuttgart MPLS underlay is degraded — packet loss 4.2% (threshold 1%) since 08:14 UTC. LTE backup is active but limited to 12 Mbps, so voice/video QoS classes have been demoted to best-effort at this site. Contact MPLS provider (ticket #SDW-1842) for ETA. Also, warehouse-munich runs on LTE only with no path redundancy — consider adding a broadband ISP link to meet the 2-path minimum SLA requirement.",
		},

		// ─── AI / DATA INTELLIGENCE ─────────────────────────────────
		"ai/training/distributed": {
			KPIs: []WSLeafKPI{
				k("Active jobs", "3"+up("+1"), "1 multi-node (8× H100) · 1 single-node (4× A100) · 1 single-GPU (V100)"),
				k("GPU hours (MTD)", "1,284"+up("+218"), "$3,421 spend · $2.66/GPU-hr avg · "+bar(64, "")+" 64% of $5k budget"),
				k("Largest run", "8× H100", "recommender-v2 · DeepSpeed ZeRO-3 · 4 h elapsed / 5.7 h est."),
				k("Avg utilisation", "82%"+up("+4%"), "target ≥ 75% · A100 at 88% · H100 at 79% · V100 at 91%"),
			},
			Headers: hdr("Job", "Framework", "Model", "GPUs", "Strategy", "Progress", "Cost", "ETA", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ recommender-v2 is running at 79% GPU utilisation on 8× H100 — gradient accumulation steps may be too large (currently 64). Reducing to 32 could improve throughput by ~15% and cut the remaining ETA by ~20 min. Also, embed-retrain-q2 is queued because all A100 capacity is allocated — it will start once llm-finetune-v4 completes in ~38 min. Current MTD spend is $3,421 of the $5,000 budget with 11 days remaining.",
		},
		"ai/training/automl": {
			KPIs: []WSLeafKPI{
				k("AutoML runs", "12"+up("+3"), "4 active · 6 completed · 2 failed · this month"),
				k("Best model", "0.94 AUC"+up("+0.02"), "churn-pred-aml · XGBoost ensemble · tabular classification"),
				k("Trials per run", "248"+up("+32"), "avg across active runs · early stopping at plateau · "+bar(50, "")+" 50% of budget"),
				k("Compute (MTD)", "$842"+up("+$124"), "auto-scaled 0–8 vCPU · $70/run avg · $1.2k budget remaining"),
			},
			Headers: hdr("Experiment", "Task", "Trials", "Best score", "Best algo", "Duration", "Cost", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ anomaly-detect failed after 42 trials — the dataset has severe class imbalance (0.3% positive rate) and the default search space doesn't include oversampling strategies. Re-run with SMOTE or focal-loss enabled in the pipeline config. Also, lead-scoring-v2 has been running for 5 h 44 m with diminishing returns — the best F1 score hasn't improved in the last 80 trials. Consider early-stopping to save the remaining ~$60 in compute.",
		},
		"ai/training/finetuning": {
			KPIs: []WSLeafKPI{
				k("Fine-tunes", "8"+up("+2"), "3 active · 4 completed · 1 failed · across 4 base models"),
				k("Adapters created", "42"+up("+6"), "28 LoRA · 14 QLoRA · best deployed: ft-llama3-acme-v4"),
				k("Datasets", "12"+up("+1"), "in feature store · 2.8M total rows · avg 234k rows/dataset"),
				k("Avg training time", `1h 42m`+down("−12m"), "for 8B model on 4× A100 · 3h 18m for 13B · "+bar(68, "")+" 68% GPU util"),
			},
			Headers: hdr("Run", "Base model", "Method", "Dataset", "Loss", "Epochs", "GPU", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ ft-gemma-qa failed with OOM at epoch 2 — QLoRA r=8 on Gemma 2 9B requires ~38 GB VRAM with the current batch size (16). Reduce batch size to 8 or switch to 4× A100 to provide 320 GB aggregate memory. Also, ft-llama3-acme-v5 is training on a superset of v4's data (+28k new pairs) — if the final loss beats 0.42, consider promoting it as the new production adapter and retiring v4.",
		},
		"ai/hosting/inference": {
			KPIs: []WSLeafKPI{
				k("Live endpoints", "9"+up("+2"), "5 GPU · 3 CPU · 1 serverless · auto-scaling 1–6 replicas"),
				k("Requests / s", "248"+up("+34"), "peak 412/s at 14:22 UTC · /chat 57% · /embed 35% · other 8%"),
				k("p99 latency", `842 ms`+down("−58ms"), "SLO target < 1 s · /chat 842 ms · /embed 88 ms · "+bar(84, "")+" 84% of SLO budget"),
				k("GPU utilisation", "62%"+up("+6%"), "A100 at 71% · T4 at 48% · L40S at 82% · A10 at 54% · headroom OK"),
			},
			Headers: hdr("Endpoint", "Model", "Replicas", "RPS", "p99", "GPU util", "Cost / hr", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ /v1/image/gen on L40S is at 82% GPU utilisation with only 1 replica — during peak hours (14:00–16:00 UTC) p99 spikes to 3.4 s. Consider adding a second replica or enabling auto-scale with a 70% GPU threshold. Also, /v1/chat p99 (842 ms) is consuming 84% of the 1 s SLO budget — if traffic grows another 20%, latency will likely breach the SLO. Pre-scaling to 4 replicas during business hours would add ~$8.40/hr but keep p99 under 600 ms.",
		},
		"ai/hosting/batch": {
			KPIs: []WSLeafKPI{
				k("Batch jobs", "7"+up("+3"), "2 running · 2 queued · 3 completed today · 0 failed"),
				k("Items processed (24h)", "412k"+up("+84k"), "embed 68% · classify 18% · summarise 14% · "+bar(72, "")+" 72% of daily target"),
				k("Avg cost / 1k items", "$0.04"+down("−$0.01"), "3× cheaper than real-time ($0.12) · $16.48 spent today · $500 budget"),
				k("Throughput", "8,420 items/min"+up("+1.2k"), "auto-scaled to 4 workers · peaks at 12k/min · scales to 0 when idle"),
			},
			Headers: hdr("Job", "Model", "Items", "Progress", "Throughput", "Cost", "Started", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ monthly-summarise is processing at 420 items/min on Llama-3-8B — significantly slower than the embed jobs because summarisation requires full sequence generation. At this rate, the remaining 7,960 items will take ~19 min. The 2 queued jobs (review-classifier + weekly-sentiment-scan) are waiting for GPU capacity — they'll start once nightly-embed-articles completes in ~12 min. Today's spend is $16.48 of the $500 monthly budget with 11 days remaining.",
		},
		"ai/hosting/edge": {
			KPIs: []WSLeafKPI{
				k("Edge devices", "248"+up("+18"), "84 Jetson Orin · 42 Coral TPU · 122 RPi 5 · 6 offline · "+bar(98, "green")+" 98% online"),
				k("Models deployed", "12"+up("+2"), "8 ONNX INT8 · 3 TFLite · 1 CoreML · avg model size 14 MB"),
				k("Inferences / s", "8.4k"+up("+1.1k"), "cameras 1.3k · phones 0.4k · IoT 6.7k · peak 12.8k at 10:00 UTC"),
				k("Avg latency", "12 ms"+down("−3ms"), "Jetson 8 ms · Coral 14 ms · RPi 18 ms · target < 25 ms all devices"),
			},
			Headers: hdr("Model", "Format", "Devices", "Inf / s", "Latency", "Memory", "Sync", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ ocr-label-reader has 6 offline scanners that failed the v2.0 OTA update — the new model (64 MB) exceeds the 48 MB flash limit on the older RPi 4 units. Either re-quantise to INT4 (~38 MB) or schedule a hardware refresh for those 6 devices. Also, speech-wakeword v3.0 hasn't synced to 4 devices in 48 hours — these are behind a restricted network segment. A manual push via the local gateway at 10.0.14.1 is needed to complete the rollout.",
		},
		"ai/apis/vision": {
			KPIs: []WSLeafKPI{
				k("API calls today", "84,210"+up("+18%"), "OCR 34% · object detect 15% · face 10% · NSFW 41% · peak 142/s at 11:30 UTC"),
				k("p95 latency", "142 ms"+down("−18ms"), "SLO < 500 ms · OCR 142 ms · detect 312 ms · face 82 ms · "+bar(28, "green")+" 28% of SLO"),
				k("Error rate", "0.02%"+down("−0.01%"), g("SLO < 0.5%")+" · 17 errors / 84k calls · 14 timeout · 3 invalid format"),
				k("Cost today", "$160.10"+up("+$22"), "$0.0019/call avg · MTD $2,842 · "+bar(57, "")+" 57% of $5k budget · 11 days left"),
			},
			Headers: hdr("Operation", "Model", "Calls today", "p95", "Errors", "Quota", "Cost", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ NSFW classification is at 69% of its daily quota (50k calls) with 12 hours remaining — at the current rate it will hit the limit by ~18:00 UTC. Consider raising the quota to 75k or enabling overflow to a secondary endpoint. Also, object detection p95 (312 ms) is the highest across all vision APIs — the YOLOv8x model processes high-resolution images (avg 2048×1536) without downscaling. Adding a pre-resize step to 1024px max dimension would cut latency to ~180 ms with < 1% accuracy loss.",
		},
		"ai/apis/language": {
			KPIs: []WSLeafKPI{
				k("API calls today", "142,810"+up("+24%"), "sentiment 34% · summarise 9% · translate 58% · embed 66% · peak 284/s"),
				k("p95 latency", "82 ms"+down("−14ms"), "SLO < 200 ms · embed 24 ms · sentiment 48 ms · translate 124 ms · "+bar(41, "green")+" 41% of SLO"),
				k("Tokens today", "24.8M"+up("+4.2M"), "input 18.2M · output 6.6M · summarise uses 68% of output tokens"),
				k("Cost today", "$84.20"+up("+$12"), "$0.0034/call avg · MTD $1,480 · "+bar(37, "")+" 37% of $4k budget · 11 days left"),
			},
			Headers: hdr("Operation", "Model", "Calls today", "Tokens", "p95", "Errors", "Cost", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ Summarise p95 latency (184 ms) is approaching the 200 ms SLO — Llama 3 8B generates ~120 output tokens per call averaging 1.5 ms/token. Long documents (> 4k input tokens) push p99 to 340 ms. Consider enabling speculative decoding or switching to a distilled 3B variant for inputs under 2k tokens to cut latency by ~40%. Also, named entity recognition error rate (0.12%) spiked from 0.03% yesterday — 22 failures are on inputs with mixed CJK + Latin scripts. The spaCy-acme-lg model needs the multilingual tokeniser enabled for these cases.",
		},
		"ai/apis/speech": {
			KPIs: []WSLeafKPI{
				k("API calls today", "12,841"+up("+8%"), "ASR 66% · TTS 22% · diarize 12% · peak 42/s at 09:15 UTC"),
				k("Audio processed", "412 hr"+up("+38hr"), "ASR 284 hr · TTS 82 hr · diarize 46 hr · "+bar(41, "green")+" 41% of 1k hr quota"),
				k("p95 latency", "1.8s"+down("−0.3s"), "SLO < 5s · ASR 2.1s · TTS 0.8s · diarize 4.2s · "+bar(36, "green")+" 36% of SLO"),
				k("Cost today", "$142.80"+up("+$18"), "$0.011/min avg · MTD $2,180 · "+bar(44, "")+" 44% of $5k budget · 11 days left"),
			},
			Headers: hdr("Operation", "Model", "Calls today", "Audio hrs", "p95", "Errors", "Quota", "Cost", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ Speaker diarization p95 (4.2s) is the highest across speech APIs — long recordings (> 30 min) with overlapping speakers push p99 to 8.4s. Consider enabling chunked diarization (5-min windows with 10s overlap) to cap latency at ~2s per chunk. Also, ASR quota is at 57% (284 / 500 hr) with 12 hours remaining — call-centre transcription jobs spike to ~60 hr/hr between 14:00–18:00 UTC. If the afternoon burst matches yesterday's pattern, the daily quota will hit at ~16:30 UTC. Pre-approve a 200 hr overflow or stagger the batch jobs to off-peak hours.",
		},
		"ai/apis/recommendation": {
			KPIs: []WSLeafKPI{
				k("API calls today", "412k"+up("+12%"), "related 60% · feed 20% · cold-start 15% · re-rank 5% · peak 1,840/s at 12:00 UTC"),
				k("p95 latency", "12 ms"+down("−2ms"), "SLO < 50 ms · related 14 ms · feed 18 ms · cold-start 8 ms · "+bar(24, "green")+" 24% of SLO"),
				k("Users served", "84,210"+up("+6,400"), "unique today · returning 72% · new 28% · "+bar(84, "green")+" 84% coverage of DAU"),
				k("Click-through lift", "+18%"+up("+2pp"), "vs. random baseline · related +18% · feed +24% · cold-start +8% · revenue uplift $4.2k/day"),
			},
			Headers: hdr("Endpoint", "Model", "Calls today", "p95", "CTR lift", "Coverage", "Cache hit", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ Re-rank overlay p95 latency (42 ms) is approaching the 50 ms SLO — LambdaMART-v2 runs a 2-stage sort over the top 200 candidates, and calls with > 500 candidates push p99 to 68 ms. Consider capping the candidate set at 200 or pre-filtering with a cheaper model. Also, the for-you feed A/B test (variant B — multi-objective with diversity boost) has reached statistical significance (p < 0.01, +24% CTR, +6% revenue) after 14 days. Ready to promote variant B to 100% traffic — current 50/50 split is leaving ~$2.1k/day of incremental revenue on the table.",
		},
		"ai/data/etl": {
			KPIs: []WSLeafKPI{
				k("Pipelines", "18"+up("+2"), "12 daily · 6 hourly · 14 active · 2 paused · 2 new this week"),
				k("Records (24h)", "412M"+up("+48M"), "ingested 412M · transformed 408M · loaded 408M · "+bar(99, "green")+" 99.0% success"),
				k("Avg duration", "12m 34s"+down("−1m 12s"), "fastest: events-rollup 2m 18s · slowest: nightly-user-features 48m · "+bar(25, "green")+" 25% of 1h timeout"),
				k("Failed runs (7d)", "2"+down("−3"), "1 timeout · 1 schema mismatch · both retried ok · "+bar(1, "green")+" 0.8% failure rate"),
			},
			Headers: hdr("Pipeline", "Source → Sink", "Schedule", "Last run", "Records", "Duration", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ nightly-user-features is taking 48m — up from 32m last week. The Redis SCAN for session features now iterates 12M keys (was 8M before the session-store migration). Adding a cursor-based bloom filter would cut scan time by ~60%. Also, events-rollup-daily processed 284M records in 2m 18s but only because ClickHouse materialized views handle the aggregation — if the MV is dropped during a schema migration, the pipeline will fall back to client-side rollup and take ~45 min. Add a pre-flight check to verify the MV exists before starting the run.",
		},
		"ai/data/streaming": {
			KPIs: []WSLeafKPI{
				k("Streams", "8"+up("+1"), "6 Kafka topics · 2 Kinesis streams · all replicated 3× · 1 new this week"),
				k("Throughput", "12.4k/s"+up("+1.8k"), "ingress 12.4k/s · egress 48.2k/s (fan-out) · peak 28.4k/s at 11:00 UTC · "+bar(44, "green")+" 44% of capacity"),
				k("Consumer lag", "420 ms"+down("−80ms"), "SLO < 2s · orders 120 ms · clickstream 420 ms · iot 84 ms · payments 210 ms"),
				k("Retention", "7 days", "orders 7d · clickstream 3d · iot 14d · payments 30d · total disk 1.8 TB / 4 TB · "+bar(45, "")+" 45%"),
			},
			Headers: hdr("Stream", "Type", "Throughput", "Partitions", "Consumers", "Lag", "Disk", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ events.clickstream consumer lag (420 ms) is the highest — the analytics-enrichment consumer group (12 consumers) is processing at 380k/s against 412k ingress. Adding 4 more consumers to that group would close the gap. Also, iot.sensors is using 840 GB with 14-day retention on 64 shards — at the current growth rate (+8%/week) it will hit the 1 TB shard limit in ~3 weeks. Plan a shard split to 96 shards or reduce retention to 10 days to buy another month. The payments.events topic has TLS + ACL enabled but the dead-letter queue (payments.events.dlq) has 142 unprocessed messages from a schema-v3 migration — these need manual replay after the consumer is patched.",
		},
		"ai/data/catalog": {
			KPIs: []WSLeafKPI{
				k("Catalogued assets", "284"+up("+12"), "142 tables · 84 views · 42 topics · 16 feature sets · 6 sources connected"),
				k("Column coverage", "8,420 / 9,180"+up("+340"), bar(92, "green")+" 92% have descriptions · 760 undocumented columns across 18 tables"),
				k("Lineage edges", "1,284"+up("+48"), "auto-discovered 1,102 · manual 182 · avg depth 4.2 hops · longest chain 11 hops"),
				k("Steward coverage", "84%"+up("+4%"), "238 / 284 assets have owners · 46 orphaned · "+bar(84, "green")+" goal 95% by Q3"),
			},
			Headers: hdr("Asset", "Source", "Owner", "Columns", "Updated", "Freshness", "PII", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ stg_erp_inventory hasn't been updated in 3 days — the upstream ERP SFTP export failed on Friday (disk full on the ERP side) and the ETL pipeline silently skipped the empty file. The table is now serving stale data to 4 downstream dashboards. Also, 46 assets have no assigned steward — 18 of these are in the prod PostgreSQL source and contain PII columns (emails, phone numbers). Assign owners before the Q3 compliance audit. The ml_features.user_v3 table has 12 undocumented columns added in the last sprint — the ML team should add descriptions before the feature set is promoted to production.",
		},
		"ai/mlops/experiments": {
			KPIs: []WSLeafKPI{
				k("Experiments", "284"+up("+18"), "142 classification · 86 NLP · 32 vision · 24 recommendation — 48 active this week"),
				k("Total runs", "8,421"+up("+312"), bar(72, "green")+" 6,042 completed · 1,841 failed · 538 running — avg 4.2 runs/experiment"),
				k("GPU hours (MTD)", "3,842"+down("-8%"), "2,140h A100 · 1,420h V100 · 282h T4 — $11,842 compute cost this month"),
				k("Best metric", "0.946 AUC"+up("+0.012"), "churn-pred-v3 run-8421 · beat previous best (0.934) from 6 days ago"),
			},
			Headers: hdr("Experiment", "Best run", "Metric", "Runs", "GPU hrs", "Framework", "Duration", "Owner", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ lora-finetune-r32 (exp-0415) has consumed 842 GPU hours over 3 days on 4×A100 — the loss curve has flattened since run-8440 (loss 0.43 vs. current 0.42). Consider early-stopping to save ~$2,400 in compute. The sentiment-bert experiment (exp-0342) failed with OOM on run-7984 — the batch size was doubled to 64 without adjusting gradient accumulation. Reduce batch size to 32 or enable gradient checkpointing before restarting. Also, recommender-v2 has 86 runs but MAP@10 has plateaued at 0.62 for the last 12 runs — the team should try a different embedding architecture or add user-session features from the clickstream data.",
		},
		"ai/mlops/registry": {
			KPIs: []WSLeafKPI{
				k("Registered models", "42"+up("+3"), "18 classification · 12 NLP · 8 vision · 4 recommendation — 6 teams contributing"),
				k("Total versions", "284"+up("+14"), bar(68, "green")+" 9 production · 4 staging · 12 stale (90d+) — avg 6.8 versions/model"),
				k("Serving requests", "84.2k/s"+up("+12%"), "across 9 production models — $4,210/day inference cost · p99 latency 42 ms"),
				k("Stale models", "12"+down("-2"), bar(29, "amber")+" 12 / 42 unused >90d — auto-archive scheduled for 8 · 4 need manual review"),
			},
			Headers: hdr("Model", "Version", "Size", "Stage", "Requests/s", "p99 latency", "Framework", "Owner", "Last used", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ churn-pred v3.0 has been in staging for 14 days — the A/B test (experiment-0412) shows +3.2% improvement in retention prediction vs. v2.0 in production. It's ready for promotion but needs sign-off from data-team@. The llm-acme-v4 model is consuming $2,840/day (68% of total inference cost) at 842 ms p99 — consider quantizing from FP16 to INT8 to cut latency and cost by ~40%. Also, 4 stale models (legacy-rules-v1, sentiment-v1, entity-extract-v1, image-tag-v1) haven't been called in 280+ days but still hold 18 GB of registry storage — approve the auto-archive to free space.",
		},
		"ai/mlops/monitoring": {
			KPIs: []WSLeafKPI{
				k("Active monitors", "18"+up("+2"), "9 production models × 2 checks (data drift + prediction drift) — all endpoints covered"),
				k("Drift alerts (24h)", "3"+up("+2"), y("churn-pred PSI 0.18")+" · "+e("recommender feature drift")+" · "+y("vision input shift")),
				k("Inferences logged", "12.8M"+up("+1.2M"), bar(64, "green")+" 64% of 20M monthly budget — 842k today · avg confidence 0.88"),
				k("Model health", "7 / 9 healthy", bar(78, "green")+" 2 models drifting — retraining recommended within 48h"),
			},
			Headers: hdr("Model", "Drift (PSI)", "Confidence", "Requests/s", "Latency p99", "Error rate", "Last retrain", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ churn-pred (mon-1b2c3d) PSI has risen from 0.08 to 0.18 over the past 10 days — the feature distribution shift is concentrated in the user_tenure and monthly_spend features, likely caused by the April billing migration. Confidence has dropped from 0.88 to 0.84 and the error rate doubled. Trigger retraining on the April dataset before drift worsens. Vision-defect (mon-4e5f6a) shows early drift at PSI 0.09 (threshold 0.10) — the factory lighting was changed last week, causing subtle input distribution shifts. Adding brightness normalization to the preprocessing pipeline would address this without retraining. The LLM model (llm-acme-v4) doesn't have PSI monitoring — consider adding embedding-space drift detection using the reference distribution from the evaluation set.",
		},

		// ─── SECURITY ───────────────────────────────────────────────
		"security/iam/sso": {
			KPIs: []WSLeafKPI{
				k("SSO providers", "3"+up("+0"), "Okta (SAML) · Google Workspace (OIDC) · Microsoft Entra ID (OIDC) — all with SCIM provisioning"),
				k("SSO coverage", "214 / 214"+up("+4"), bar(100, "green")+" 100% workforce on SSO — 0 local-only accounts · auto-provisioning enabled"),
				k("Logins (24h)", "1,284"+up("+8%"), "Okta 842 (66%) · Google 312 (24%) · Entra 130 (10%) — peak hour 08:00–09:00 UTC"),
				k("Cert / key health", "2 / 3 OK", y("Okta cert expires in 18d")+" — SCIM sync healthy · MFA enforced on all providers"),
			},
			Headers: hdr("Provider", "Protocol", "Domain", "Users", "Logins (24h)", "SCIM", "Cert expires", "Last rotation", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ The Okta SAML signing certificate (SHA-256, 2048-bit RSA) expires on 2026-05-07 — 18 days from now. Rotate it immediately in Okta Admin → Applications → Platform SSO → SAML Settings → Certificate, then upload the new metadata XML here. If the cert expires, all 142 Okta users will be locked out at next login. The LDAP bridge (sso-3d4e5f) has been disabled since the migration to Okta in January — it still has 4 service accounts mapped. Confirm these accounts are decommissioned and delete the provider to reduce attack surface. Google Workspace SCIM sync last ran 2 minutes ago and provisioned 4 new users from the engineering group — verify they received the correct RBAC roles.",
		},
		"security/iam/mfa": {
			KPIs: []WSLeafKPI{
				k("MFA coverage", "212 / 214"+up("+4"), bar(99, "green")+" 99.1% enrolled — 2 users pending · policy requires 100% by end of month"),
				k("Passkey / WebAuthn", "184"+up("+12"), bar(87, "green")+" 87% using phishing-resistant — 142 platform keys · 42 cross-platform (YubiKey)"),
				k("TOTP fallback", "28"+down("-4"), bar(13, "amber")+" 13% still on TOTP — migration campaign active · 8 converted this week"),
				k("SMS fallback", "2"+down("-1"), e("2 users on SMS only")+" — compliance risk: NIST 800-63B deprecates SMS · upgrade deadline Apr 30"),
			},
			Headers: hdr("User", "Primary", "Backup", "Roles", "Recovery codes", "Devices", "Last MFA", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ 2 users are still on SMS-only MFA (temp.contractor@, ext.auditor@) — SMS is vulnerable to SIM-swap attacks and NIST SP 800-63B recommends phishing-resistant authenticators. The compliance deadline is April 30; after that, the policy engine will block SMS-only logins. Send these users a YubiKey or enroll them in platform passkeys. Also, sam.patel@ has exhausted all 10 recovery codes and has no backup method — if they lose their TOTP device, they'll be locked out. Force recovery code regeneration at next login. The TOTP→Passkey migration campaign converted 8 users this week (from 36 to 28 TOTP-only) — at this rate the org will reach 95% passkey adoption by mid-May.",
		},
		"security/iam/rbac": {
			KPIs: []WSLeafKPI{
				k("Roles defined", "12"+up("+1"), "4 built-in · 8 custom — last custom role created 6d ago (\"Incident Responder\")"),
				k("Permissions", "284", bar(100, "green")+" all 284 fine-grained permissions mapped — 0 orphaned · 0 unused in 90d"),
				k("Role assignments", "412"+up("+8"), "214 users · avg 1.9 roles/user — 4 admins · 12 elevated-privilege accounts"),
				k("Access review", "12d ago", g("On track")+" quarterly cycle — next review due May 12 · 0 overdue findings"),
			},
			Headers: hdr("Role", "Type", "Users", "Permissions", "Scope", "Last modified", "Review status", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ The \"Contractor (temp)\" role (role-7b8c9d) has not been reviewed in 84 days — well past the 90-day quarterly review cycle. It grants read access to compute and storage resources for 5 contractor accounts. Verify these contractors are still active and their access is still justified. The new \"Incident Responder\" role (role-5f6a7b) was created 6 days ago and has not yet been reviewed — it grants write access to incidents and read access to monitoring/logs for 3 SRE team members. Schedule a review before the May 12 deadline. 4 users hold the Admin role with full 284-permission access — confirm all 4 require global admin privileges and consider scoping down to least-privilege where possible.",
		},
		"security/data/enc-rest": {
			KPIs: []WSLeafKPI{
				k("Volumes encrypted", "18 / 18"+up("+2"), bar(100, "green")+" 100% coverage — AES-256-XTS · all using CMK via KMS · last unencrypted volume migrated 14d ago"),
				k("Buckets encrypted", "24 / 24", bar(100, "green")+" 100% SSE-KMS — default encryption policy enforced · 0 public buckets · versioning enabled on 22/24"),
				k("Databases encrypted", "12 / 12", bar(100, "green")+" 100% AES-256 — TDE enabled on all instances · backup encryption verified · point-in-time recovery encrypted"),
				k("Avg key age", "84d", y("6d until rotation")+" auto-rotation set to 90d cycle · oldest key: alias/backup-key at 88d · all CMK healthy"),
			},
			Headers: hdr("Resource", "Type", "Algorithm", "KMS key", "Key age", "Last verified", "Compliance", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ The KMS key alias/rds-default (used by prod-main database) is 88 days old — only 2 days from the 90-day auto-rotation trigger. Verify the rotation event will proceed cleanly by checking that no application is pinned to a specific key version. The backup encryption key alias/backup-key is at 88 days and will also rotate soon. After rotation, validate that existing backups remain decryptable with the previous key version (KMS retains old versions automatically). All 18 block volumes were migrated to CMK-based encryption 14 days ago — the 2 volumes that were previously using platform-managed keys (vol-0e1f, vol-0f1g) are now confirmed encrypted with alias/db-master.",
		},
		"security/data/enc-transit": {
			KPIs: []WSLeafKPI{
				k("TLS endpoints", "84"+up("+2"), "82 public · 2 internal — all TLS 1.3+ · 0 legacy TLS 1.0/1.1 endpoints remaining"),
				k("TLS 1.3 coverage", "100%", bar(100, "green")+" full modern crypto — TLS 1.2 sunset completed 28d ago · 0 fallback negotiation"),
				k("HSTS coverage", "82 / 82 public", g("max-age 31536000")+" includeSubDomains + preload on all public endpoints · HSTS preload list submitted"),
				k("mTLS internal", "48 services", bar(100, "green")+" 100% service-mesh coverage — Istio auto-rotation · cert lifetime 24h · SPIFFE identity"),
			},
			Headers: hdr("Endpoint", "Protocol", "Cipher suite", "Cert issuer", "Grade", "Expires", "HSTS", "Last scan", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ TLS 1.2 sunset was completed 28 days ago — 2 legacy API clients (partner-feed-v1, legacy-webhook-handler) were migrated to TLS 1.3 during the transition. Monitor the connection error logs for any stragglers still attempting TLS 1.2 handshakes (current count: 0 in the last 7 days). The admin.acme.io certificate is issued by DigiCert (EV) and expires in 82 days — all other public certs use Let's Encrypt with auto-renewal. Verify the DigiCert cert is in the auto-renewal pipeline or set a calendar reminder for manual renewal. All 48 internal services use mTLS with SPIFFE-based identity and 24-hour certificate rotation via Istio — no manual cert management required.",
		},
		"security/data/kms": {
			KPIs: []WSLeafKPI{
				k("Customer-managed keys", "12"+up("+1"), "8 symmetric (AES-256) · 3 asymmetric (RSA-3072) · 1 ECDSA P-256 — all hardware-backed HSM"),
				k("Envelope encryption", "54 resources", bar(100, "green")+" 100% data-at-rest covered — 18 volumes · 24 buckets · 12 databases wrapped with CMKs"),
				k("Auto-rotation", "11 / 12", y("1 key on manual rotation")+" 11 keys on 90d auto-cycle · alias/asymmetric-signing requires manual rotation (RSA constraint)"),
				k("Key operations (24h)", "284k"+up("+6%"), "encrypt 184k (65%) · decrypt 96k (34%) · sign 4k (1%) — peak at 14:00 UTC · no throttling"),
			},
			Headers: hdr("Key alias", "Algorithm", "Type", "Region", "Age", "Rotation", "Operations (24h)", "Grants", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ Three keys (alias/db-master, alias/rds-default, alias/backup-key) are at 88 days and will auto-rotate in 2 days. Auto-rotation creates a new key version while keeping old versions available for decryption — no downtime expected, but verify that no application is hardcoded to a specific key version ID. The asymmetric signing key (alias/asymmetric-signing, RSA-3072) is 284 days old and requires manual rotation since KMS cannot auto-rotate asymmetric keys. Schedule the rotation and coordinate with downstream services that validate signatures against the public key — they will need the new public key material. The ECDSA P-256 key (alias/jwt-signing) was added 42 days ago for JWT token signing — it replaced the previous HMAC-based approach for better key isolation.",
		},
		"security/network/waf": {
			KPIs: []WSLeafKPI{
				k("WAF rules", "84"+up("+3"), "68 managed (OWASP CRS 3.3) · 16 custom — last rule added 2d ago (rate-limit /api/auth)"),
				k("Requests blocked (24h)", "12,841"+up("+14%"), bar(2, "green")+" 2.1% block rate — 12,584 bots (98%) · 241 SQLi · 16 XSS · 0 false-positive escalations"),
				k("False positive rate", "0.01%"+down("-0.02%"), g("Very low")+" 1 in 10,000 — tuned via 3 custom exclusion rules · last FP incident 14d ago (resolved)"),
				k("Top attack vector", "SQLi 41%", y("High")+" 5,266 SQLi attempts (24h) — 38% from TOR exits · 22% from VPS ranges · geo: US 34%, CN 28%, RU 18%"),
			},
			Headers: hdr("Rule", "Priority", "Action", "Source", "Matches (24h)", "False positives", "Severity", "Last triggered", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ SQLi attempts spiked 14% in the last 24 hours — 38% originating from TOR exit nodes. The OWASP CRS 3.3 ruleset is blocking all attempts with zero false positives. The new rate-limit rule on /api/auth (waf-2c3d4e, added 2 days ago) has triggered 3,420 times and produced 2 false positives — both were legitimate API integrations hitting the 60 req/min threshold. Consider adding those integration IPs to the allowlist or raising the threshold to 120 req/min for authenticated clients. The bot UA challenge rule (waf-4e5f6a) has 4 false positives — all from a monitoring service using a non-standard User-Agent. Add an exclusion for the monitoring service's UA string to eliminate these.",
		},
		"security/network/ddos": {
			KPIs: []WSLeafKPI{
				k("Attacks mitigated (30d)", "8"+down("-3"), "5 L3 SYN floods · 2 L7 HTTP floods · 1 DNS amplification — all auto-mitigated, zero manual intervention"),
				k("Largest attack", `412 <span style="font-size:13px;color:var(--lf-dim)">Gbps</span>`+up("+18%"), "L3 SYN flood from 14k source IPs on Apr 11 — scrubbed in 1.8s via anycast PoPs (AMS, FRA, IAD)"),
				k("Time to mitigate (p95)", "<3s", g("Excellent")+" avg 1.4s — fastest: 0.6s (DNS amp) · slowest: 2.8s (multi-vector L3+L7) · SLA target: <10s"),
				k("Customer impact", "0 incidents", g("No degradation")+" 100% uptime maintained during all 8 attacks · no latency spikes >5ms observed on origin"),
			},
			Headers: hdr("Protection layer", "Method", "Capacity / Threshold", "Attacks handled (30d)", "Last triggered", "False positives", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "The largest attack on Apr 11 (412 Gbps SYN flood from 14k sources) was fully mitigated in 1.8s with zero customer impact. Bot behaviour scoring (ddos-bot) has 12 false positives this month — all from legitimate crawlers with atypical JS execution profiles. Consider adding Googlebot and Bingbot IP ranges to the scoring allowlist to eliminate these. API per-token quota (ddos-api) had 3 false positives from a partner integration exceeding the 1,000 req/min limit during batch syncs — raise that specific token's quota to 3,000 req/min or switch to a bulk endpoint.",
		},
		"security/network/ids": {
			KPIs: []WSLeafKPI{
				k("Sensors deployed", "12"+up("+2"), "across 4 VPCs · 3 production · 1 staging — last deployed: ids-sensor-12 in payments-vpc (3d ago)"),
				k("Events analysed (24h)", "8,421"+up("+6%"), bar(2, "green")+" 8,252 benign (98%) · 142 suspicious · 27 confirmed threats — Suricata 7.0 ruleset (ET Pro, updated 4h ago)"),
				k("Alerts (24h)", "2"+down("-5"), g("0 critical")+" 0 critical · 2 medium (port scan, lateral movement) — MTTD avg 12s · MTTR avg 4m · 0 escalations to SOC"),
				k("Threat intel feeds", "4 active", g("All synced")+" ET Pro · Abuse.ch · AlienVault OTX · internal IOC — last sync 4h ago · 284k indicators loaded"),
			},
			Headers: hdr("Rule / signature", "Severity", "Source → Destination", "Matches (24h)", "False positives", "Action", "Last triggered", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "The port scan from 10.0.4.42 (ids-rule-001) correlates with the lateral movement alert (ids-rule-002) — same source IP probed 1,024 ports then attempted SMB access to file-srv. Investigate whether 10.0.4.42 is compromised or if this is an authorized vulnerability scan (no scan window was scheduled). SSH brute-force (ids-rule-004) is at 412 attempts from unique external IPs in 24h — all blocked, but consider tightening the geo-fence or switching to certificate-only auth for the bastion host. Tor exit node traffic (ids-rule-006) has 2 false positives from a legitimate privacy-conscious vendor — add their exit IP to the allowlist.",
		},
		"security/compliance/policy": {
			KPIs: []WSLeafKPI{
				k("Active policies", "84"+up("+6"), "68 managed (CIS Benchmark v1.5) · 16 custom — 6 added this month (all encryption-related, triggered by Q1 audit findings)"),
				k("Compliance score", "98.4%"+up("+1.2%"), bar(98, "green")+" 1,842 / 1,872 resources compliant — 22 non-compliant (8 IAM keys, 6 open SGs, 4 unencrypted volumes, 4 public buckets remediated)"),
				k("Auto-remediated (24h)", "4"+down("-2"), "policy-bot closed 4 drift events — 2 public bucket ACLs reset · 1 SG port 22 removed · 1 EBS encryption enabled on new volume"),
				k("Evaluation cadence", "continuous", g("Real-time")+" last full scan: 12s ago · avg evaluation latency: 340ms · 1.2M resource checks/day · 0 evaluation failures"),
			},
			Headers: hdr("Policy", "Severity", "Framework", "Compliant / Total", "Drift (7d)", "Auto-remediate", "Last evaluated", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "Two IAM access keys (pol-iam-rot) are overdue for rotation: key AKIA…3F7Q for user ci-deploy (age: 112d) and key AKIA…9K2M for user legacy-sync (age: 98d). The ci-deploy key is used by the CI/CD pipeline — coordinate rotation with the DevOps team and update the pipeline secrets before revoking. The legacy-sync key powers a deprecated data sync; confirm with the data team whether it can be decommissioned entirely. Auto-remediation resolved 4 drift events in the last 24h with zero false positives — the policy-bot is operating within expected parameters.",
		},
		"security/compliance/audit": {
			KPIs: []WSLeafKPI{
				k("Events logged (24h)", "1.2M"+up("+8%"), "842k API calls · 312k console actions · 46k system events — peak hour: 14:00 UTC (84k events) · avg event size: 1.4 KB"),
				k("Retention policy", `7 <span style="font-size:13px;color:var(--lf-dim)">years</span>`, g("WORM-locked")+" immutable storage · SHA-256 hash chain verified every 1h · total archive: 2.8 TB (compressed) across 3 regions"),
				k("Integrity status", "0 tamper events", g("Verified")+" last hash-chain validation: 42 min ago · 0 anomalies since inception · FIPS 140-2 compliant storage backend"),
				k("SIEM export", "1h ago"+down("-12m lag"), g("Healthy")+" streaming to Splunk via Kinesis · avg latency: 3.2s · 0 dropped events (24h) · next batch: 18 min"),
			},
			Headers: hdr("Event", "Actor", "Source IP", "Resource", "Region", "When", "Severity", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "The break-glass role assumption by jane.doe@acme.io (evt-d4e5f6, 2h ago) is flagged — verify the incident ticket that justified emergency access and ensure the session was time-bounded. The same user also created an IAM access key for a contractor (evt-a1b2c3, 2 min ago) — confirm this follows the contractor onboarding process and that the key has an expiry policy attached. Console login from a new geography (evt-a7b8c9, alex.kumar from Singapore) triggered an anomaly — verify with the user whether this is legitimate travel or a credential compromise. All 1.2M events in the last 24h were delivered to Splunk with zero drops and average latency of 3.2 seconds.",
		},
		"security/compliance/regulatory": {
			KPIs: []WSLeafKPI{
				k("Active frameworks", "6", "SOC 2 · ISO 27001 · HIPAA · PCI-DSS v4 · GDPR · FedRAMP — all mapped to unified control library (412 unique controls)"),
				k("Compliance score", "5 / 6"+down("-1 finding"), bar(83, "green")+" 786 / 787 controls passing — 1 open HIPAA finding (BAA amendment pending for new sub-processor, due May 1)"),
				k("Last external audit", "Q1 2026", g("Pass")+" Deloitte completed SOC 2 Type II examination on Mar 28 — zero exceptions · report available in the evidence vault"),
				k("Next renewal", "Q3 2026", y("84 days")+" ISO 27001 surveillance audit scheduled Jul 12 — pre-audit readiness review planned for Jun 15 · 0 open pre-audit items"),
			},
			Headers: hdr("Framework", "Scope", "Controls passed", "Evidence items", "Last assessment", "Assessor", "Renewal", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "HIPAA has 1 open finding (reg-hipaa, control §164.502): a new sub-processor (analytics vendor) requires a Business Associate Agreement amendment — legal is drafting, expected completion by May 1. No PHI exposure occurred. The ISO 27001 surveillance audit is in 84 days (Jul 12) — schedule the internal pre-audit readiness review for Jun 15 and assign control owners to refresh evidence for Annex A controls A.8 (asset management) and A.12 (operations security), which had the most observations in the last cycle. SOC 2 Type II report for FY2025–2026 is available in the evidence vault with zero exceptions — share with customers via the trust portal.",
		},

		// ─── DEVOPS ─────────────────────────────────────────────────
		"devops/cicd/source": {
			KPIs: []WSLeafKPI{
				k("Repositories", "84"+up("+3"), "42 Go · 28 TypeScript · 14 Python — git over SSH (74%) and HTTPS (26%) · signed commits enforced on 80/84"),
				k("Pull requests open", "32"+down("-4"), bar(38, "amber")+" 12 ready to merge · 8 awaiting review > 48h · avg review cycle 6h — auto-merge enabled on 18 repos"),
				k("Pushes (24h)", "248"+up("+12%"), "main 84 (34%) · feature branches 142 (57%) · release 22 (9%) — peak hour 14:00–15:00 UTC"),
				k("Avg PR age", "1d 4h"+down("-8h"), g("Improving")+" merged-PR median: 18h · p95: 4d 2h — 3 PRs older than 7d need attention"),
			},
			Headers: hdr("Repository", "Default branch", "Language", "Last push", "PRs open", "Coverage", "CI status", "Branch protection", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ platform/worker (repo-2c3d4e) has 2 failing tests on the feat/tracing branch — the OpenTelemetry span propagation test broke after the gRPC interceptor refactor merged 4h ago. The 12 open PRs on this repo are blocked by the red CI. Fix the tracing tests before merging any of them. platform/docs (repo-4e5f6a) does not have branch protection enforced — anyone with write access can push directly to main without review. Enable branch protection rules to match the org-wide policy. 3 PRs across the org are older than 7 days: #482 on platform/api (waiting on security review), #214 on platform/web (draft — stale), #38 on shared/design-tokens (author on PTO). Assign reviewers or close stale drafts.",
		},
		"devops/cicd/build": {
			KPIs: []WSLeafKPI{
				k("Builds today", "248"+up("+18%"), bar(98, "green")+" 98.4% success rate — 244 passed · 4 failed · 0 cancelled — best 7d average since March"),
				k("Avg duration", "4m 12s"+down("-22s"), g("Trending faster")+" p50: 3m 04s · p95: 8m 42s · p99: 14m 18s — Docker layer caching saving ~2m per build"),
				k("Cache hit rate", "84%"+up("+6%"), bar(84, "green")+" Docker layers 92% · Go modules 88% · npm 74% — cold builds avg 11m vs cached 4m"),
				k("Concurrent builders", "12 / 16"+up("+2"), bar(75, "green")+" auto-scaled from 8 → 12 at 10:00 UTC · peak 14 at 15:00 — headroom OK, no queue wait"),
			},
			Headers: hdr("Pipeline", "Repo", "Branch", "Trigger", "Duration", "Runner", "Cache", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ The security-scan pipeline (pl-2c3d4e) failed on platform/worker feat/tracing branch — Trivy found 1 critical CVE (CVE-2024-3094 in xz-utils 5.6.0) introduced by the new tracing dependency. Upgrade xz-utils to 5.6.1+ or pin the base image to alpine:3.19.1 which includes the patched version. This is blocking 12 PRs on this repo. The integration-db-tests pipeline (pl-5f6a7b) has a 68% cache hit rate because the test database migrations changed — the cache key includes the migration checksum, so this is expected after schema changes and will recover on the next run. Runner auto-scaling peaked at 14/16 today at 15:00 UTC — if the team grows, consider increasing the runner pool ceiling to 24 to avoid queue wait times during peak hours.",
		},
		"devops/cicd/deploy": {
			KPIs: []WSLeafKPI{
				k("Deployments (24h)", "32"+up("+6"), "prod 8 · staging 14 · dev 10 — all via automated pipelines · 0 manual deploys"),
				k("Production deploys", "8"+up("+2"), bar(100, "green")+" 100% canary-gated — avg canary duration 18 min · error-rate gate < 0.1% · latency gate p99 < 500ms"),
				k("Rollbacks (7d)", "1"+down("-2"), g("Low")+" auto-detected via error-rate spike on worker v0.8.9 — rolled back in 42s · MTTR: 3m 12s including detection"),
				k("Avg lead time", "12 min"+down("-3 min"), g("PR → prod")+" commit-to-deploy p50: 8 min · p95: 28 min — DORA Elite classification · change failure rate 1.2%"),
			},
			Headers: hdr("Deploy", "App", "Env", "Strategy", "Traffic", "Error rate", "Latency p99", "Duration", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ The api-canary deployment (dpl-3d4e5f, v2.1.5-rc) has been running at 25% traffic for 12 minutes — error rate is 0.08% (below the 0.1% auto-rollback gate but elevated vs baseline 0.02%) and p99 latency is 284ms (up from 142ms baseline). The canary will auto-promote to 50% at the 20-minute mark if metrics stay within gates. Monitor closely — the latency increase correlates with the new request-validation middleware added in this release. The worker v0.8.9 rollback (dpl-5f6a7b) on April 17 was caused by a nil-pointer panic in the job-retry handler when processing malformed payloads. The fix is in v0.9.1 (now in staging). Schedule the post-mortem and update the runbook to include payload-validation checks before job retry.",
		},
		"devops/iac/templates": {
			KPIs: []WSLeafKPI{
				k("IaC stacks", "18"+up("+1"), "14 Terraform 1.7.4 · 3 Pulumi (TypeScript) · 1 Pulumi (Python) — state stored in S3 + DynamoDB locking"),
				k("Resources managed", "1,284"+up("+42"), bar(100, "green")+" 100% infrastructure as code — 0 manually created resources · last import: 14d ago (3 legacy SGs)"),
				k("Drift detected", "2", y("staging-cluster")+" security group rules modified outside Terraform — auto-drift scan runs every 4h · last clean plan on prod: 12m ago"),
				k("Last plan", "12m ago", g("Clean on prod")+" prod-network: 0 changes · prod-compute: 0 changes · staging-cluster: 2 drift items require reconciliation"),
			},
			Headers: hdr("Stack", "Tool", "Resources", "State backend", "Last applied", "Plan output", "Drift", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ staging-cluster (stack-4e5f6a) has 2 drift items: sg-0x4a2b had 2 ingress rules added manually (port 8080 from 10.0.0.0/8 and port 9090 from 0.0.0.0/0). The 0.0.0.0/0 rule on port 9090 is an open-to-internet security risk — remove it immediately via the AWS console, then run terraform apply to reconcile state. The dev-sandbox stack (stack-5f6a7b) has not been planned in 14 days — its state may have diverged significantly. Run pulumi preview before applying any changes. The 3 legacy security groups imported 14 days ago (sg-legacy-web, sg-legacy-api, sg-legacy-db) were successfully imported into prod-network and are now fully managed by Terraform — no further manual management required.",
		},
		"devops/iac/config": {
			KPIs: []WSLeafKPI{
				k("Config bundles", "84"+up("+2"), "48 Ansible playbooks · 24 Chef cookbooks · 12 Puppet manifests — across 84 hosts in 3 environments"),
				k("Drift detected", "12"+down("-4"), bar(86, "green")+" 86% auto-corrected in < 5 min — 10 corrected automatically · 2 require manual review (Puppet hosts)"),
				k("Avg apply time", `82 <span style="font-size:13px;color:var(--lf-dim)">s</span>`+down("-12s"), g("Improving")+" p50: 48s · p95: 184s · p99: 312s — Ansible 42s avg · Chef 128s avg · Puppet 214s avg"),
				k("Compliance", "98.4%"+up("+0.8%"), bar(98, "green")+" CIS Level 2 benchmark — 82/84 hosts fully compliant · 2 Puppet hosts with sshd_config drift"),
			},
			Headers: hdr("Bundle", "Tool", "Hosts", "Inventory", "Last apply", "Changed", "Compliance", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ legacy-puppet-pp (cfg-4e5f6a) has sshd_config drift on 2 hosts (node-legacy-07 and node-legacy-08): PermitRootLogin was changed from 'no' to 'yes' outside of Puppet, likely during manual troubleshooting on April 16. This is a CIS Level 1 violation — root SSH login must be disabled. Fix by running puppet agent -t on both nodes, or if intentional, update the manifest and document the exception. The monitoring-agent bundle updated the Datadog agent from 7.52 to 7.53 on all 84 hosts 1 hour ago — the change was auto-applied via the scheduled Ansible cron job. Verify agent health on the Datadog dashboard. Consider migrating the 8 Puppet-managed hosts to Ansible to reduce tooling fragmentation — the Puppet master is a single point of failure and has not been upgraded since the Puppet 7→8 migration 6 months ago.",
		},
		"devops/api/gateway": {
			KPIs: []WSLeafKPI{
				k("APIs published", "12"+up("+1"), "8 REST · 3 GraphQL · 1 gRPC — all behind Kong 3.6 gateway with rate-limiting and JWT auth enabled"),
				k("Calls (24h)", "412k"+up("+18%"), bar(72, "green")+" 72% of 600k daily capacity — peak hour 14:00 UTC at 32k req/min · auto-scaling threshold at 40k req/min"),
				k("p95 latency", `48 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("-6 ms"), g("Within SLO")+" edge-to-origin · p50: 12 ms · p99: 142 ms — CDN cache hit rate 84% reducing origin load"),
				k("Error rate", "0.04%"+down("-0.01%"), g("Well within SLO")+" 0.5% budget — 168 errors/412k calls · 142 were 429 rate-limits · 26 genuine 5xx on billing-api"),
			},
			Headers: hdr("API", "Version", "Protocol", "Calls (24h)", "p95", "Rate limit", "Auth", "Error rate", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ billing-api (gw-2c3d4e) p95 latency spiked from 82 ms to 142 ms after yesterday's v1.4.2 deploy — the 26 errors (5xx) are upstream timeouts from the payment processor when card-validation calls exceed 200 ms. The issue is intermittent and correlates with payment processor maintenance windows (14:00–16:00 UTC). Consider adding a circuit breaker with 500 ms timeout and fallback to queued processing. admin-api (gw-4e5f6a) v3.0-beta is in canary at 10% of internal traffic — the elevated latency (184 ms vs 82 ms baseline) is caused by the new request-validation middleware; profiling shows 60 ms spent on JSON schema validation per request. Optimize by caching compiled schemas. The public-events-api rate limit of 50k/min was raised from 20k/min on April 15 to accommodate a partner integration launch — review in 30 days and consider per-consumer limits instead of a global ceiling.",
		},
		"devops/api/portal": {
			KPIs: []WSLeafKPI{
				k("Registered developers", "1,842"+up("+124"), "1,284 active (70%) · 558 dormant (>90d) — self-service registration via OAuth + email verification"),
				k("Active apps", "284"+up("+18"), bar(78, "green")+" 78% calling APIs in last 7d — 142 internal · 84 partner · 58 public tier"),
				k("Docs visitors (24h)", "8,421"+up("+22%"), "top pages: /v1/listings (2.4k) · /auth/guide (1.8k) · /webhooks (1.2k) — avg session 4.2 min"),
				k("API key usage", "412k calls/day", g("Healthy")+" top consumer: ci-bot (41%) · 2nd: acme-mobile-ios (18%) — 0 revoked keys in last 30d"),
			},
			Headers: hdr("App", "Owner", "Plan", "API keys", "Calls (7d)", "Last used", "Quota usage", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ ci-bot (app-3d4e5f) is consuming 91% of its Internal-tier quota (2M calls/week) — at current growth (+8%/week) it will hit the hard limit within 6 days. Either raise the quota ceiling or split ci-bot into per-pipeline tokens with separate limits. partner-zapier-integration (app-2c3d4e) is at 82% of its Partner quota; Zapier's onboarding campaign goes live April 25, expected to double traffic — proactively raise from 500k to 1M/week. legacy-sync-tool (app-5f6a7b) has a single API key expiring May 1 and hasn't been used in 14 days — confirm with migrations-team whether the tool is decommissioned, then revoke the key to reduce attack surface. The /auth/guide docs page had 1.8k views yesterday — most visitors are new partner developers onboarding; consider adding a quick-start tutorial to reduce support tickets (currently 12/week tagged 'api-onboarding').",
		},
		"devops/api/analytics": {
			KPIs: []WSLeafKPI{
				k("Endpoints monitored", "84"+up("+4"), "across 12 APIs — 62 REST · 18 GraphQL fields · 4 gRPC methods — auto-discovered via OpenAPI + reflection"),
				k("Error budget", "0.46% consumed"+down("-0.12%"), bar(4, "green")+" 0.5% SLO · remaining budget: 1,684 errors this month — on track for 99.5% target"),
				k("Slowest endpoint", "/v1/search · p99 412 ms", e("Above SLO")+" p95: 218 ms · p50: 42 ms — root cause: full-text scan on unindexed description field in Elasticsearch"),
				k("Top consumer", "ci-bot · 41% of traffic", "1.8M calls/week — next: acme-mobile-ios (18%) · partner-zapier (12%) — no abuse patterns detected"),
			},
			Headers: hdr("Endpoint", "API", "Method", "Calls (24h)", "p50", "p95", "p99", "Error rate", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ /v1/search (ep-2c3d4e) has a p99 latency of 412 ms, breaching the 300 ms SLO since April 17. Root cause: the new full-text search on the 'description' field bypasses the Elasticsearch index and falls back to a scan query when descriptions exceed 4 KB. Fix: add a keyword sub-field to the mapping and reindex — estimated reindex time 40 min on the current cluster. The 0.4% error rate on /v1/listings (ep-0a1b2c) is entirely 429 (rate limit) responses from 3 partner apps hitting the 10k/min ceiling; no server errors. Consider per-partner rate limits instead of the global ceiling. /v1/webhooks/deliver (ep-5f6a7b) has 0.12% errors — all are timeouts from downstream endpoints that take >5s to respond; the retry queue is processing normally but consider adding a dead-letter queue for endpoints that fail 3+ consecutive deliveries.",
		},
		"devops/obs/logging": {
			KPIs: []WSLeafKPI{
				k("Log volume (24h)", `48 <span style="font-size:13px;color:var(--lf-dim)">GB</span>`+up("+12%"), "structured JSON via Fluent Bit 3.0 → OpenSearch 2.12 — compression ratio 8:1 · raw ingress 384 GB/day"),
				k("Indexed lines", "412M"+up("+48M"), bar(68, "green")+" 68% of index capacity — shard count: 24 primary · 24 replica · avg shard size 8.2 GB within best-practice range"),
				k("Error lines (24h)", "1,842"+down("-284"), g("0.0004%")+" of total — 1,284 from billing-api upstream timeouts · 412 from worker OOM kills · 146 nginx 502s"),
				k("Retention", "30d hot · 90d warm · 1y cold", "hot: NVMe SSD · warm: S3 Standard · cold: S3 Glacier — total stored: 4.2 TB across all tiers"),
			},
			Headers: hdr("Pipeline", "Source", "Format", "Volume (24h)", "Lines (24h)", "Retention", "Lag", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ k8s-events pipeline (log-3d4e5f) is experiencing 18s ingestion lag due to backpressure — the kube-system namespace on the staging cluster is emitting 4x normal event volume after a CrashLoopBackOff storm on the metrics-collector DaemonSet (42 pods restarting every 30s). The collector was fixed at 16:00 UTC but the log queue is still draining — expected to clear within 2 hours. Consider adding a dead-letter queue for k8s-events to prevent lag from affecting other pipelines. The api-prod pipeline (log-0a1b2c) accounts for 52% of total volume — the 24.8 GB/day is driven by DEBUG-level request/response body logging enabled on April 14 for the billing-api investigation. This should be turned off now (set LOG_LEVEL=INFO on the api deployment) to reduce volume by ~40% and save ~$180/month in storage costs. The audit-log pipeline (log-4e5f6a) is configured for 7-year immutable retention per SOC 2 Type II requirements — verify that the S3 Object Lock policy is still active on the compliance-logs-archive bucket.",
		},
		"devops/obs/metrics": {
			KPIs: []WSLeafKPI{
				k("Metrics in flight", "1.2M /min"+up("+8%"), "Prometheus 2.51 · scrape interval 15s · 84 targets — federation across 3 clusters via Thanos sidecar"),
				k("Active alerts", "2"+down("-3"), g("0 critical")+" · 1 warning (billing-api p95) · 1 info (disk 72%) — Alertmanager routing to PagerDuty + Slack #sre-alerts"),
				k("Cardinality", "8.4M series", bar(56, "green")+" 56% of 15M budget — top contributor: kube-state-metrics (1.8M) · 2nd: api-server (1.2M) · label churn: 0.4%/day"),
				k("Dashboards", "284"+up("+6"), "142 team-owned · 84 auto-generated · 58 personal — avg load time 1.2s · most-viewed: Platform Overview (84 views/day)"),
			},
			Headers: hdr("Dashboard", "Owner", "Panels", "Data source", "Refresh", "Last viewed", "Alerts linked", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ The API SLOs dashboard (dash-1b2c3d) shows billing-api p95 latency at 142 ms, breaching the 100 ms warning threshold — this correlates with the gateway latency spike (see devops/api/gateway). The alert has been firing for 18 hours; if it continues past 24h, it will auto-escalate to critical and page the on-call. Legacy infra dashboard (dash-5f6a7b) uses the Graphite data source which is scheduled for decommission on May 15 — the 6 panels need to be migrated to Prometheus queries or the dashboard should be archived. The ops-team hasn't viewed it in 8 days, suggesting it may be unused. Cardinality is at 56% of the 15M budget — kube-state-metrics is the largest contributor at 1.8M series; consider enabling metric relabeling to drop unused kube_pod_labels dimensions (estimated reduction: 400k series). Thanos compaction is healthy — the most recent compaction completed 2h ago with 0 errors and the store-gateway query latency is p99 84 ms.",
		},
		"devops/obs/tracing": {
			KPIs: []WSLeafKPI{
				k("Spans (24h)", "8.4M"+up("+1.2M"), "OTLP via Grafana Tempo 2.4 · ingestion rate 97 spans/s — 42% from api, 18% worker, 14% payments, rest long-tail"),
				k("Sample rate", "5% head + 100% errors", bar(68, "green")+" 68% storage budget used — tail-based on latency >500 ms and all 5xx responses"),
				k("Slowest trace (24h)", "8.4 s"+up("+2.1 s"), e("search-svc /v1/search")+" — spans: 142 · root cause: unindexed full-text query on description field"),
				k("Services traced", "84"+up("+3"), "auto-instrumented via OpenTelemetry SDK · 62 Go · 18 Node · 4 Python — 100% of prod services covered"),
			},
			Headers: hdr("Service", "Spans (24h)", "p50", "p95", "p99", "Error rate", "Sampling", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ search-svc (svc-3d4e5f) p99 latency has increased from 6.2 s to 8.4 s over the last 48 hours — the slowest traces show 142 spans with the root span blocked on an Elasticsearch full-text scan against an unindexed 'description' field. Adding a keyword sub-field and reindexing (estimated 40 min) would reduce p99 to <500 ms. The 0.18% error rate is caused by client-side timeouts (5s default) on the /v1/search endpoint — consider increasing the client timeout to 10s as a short-term mitigation. payments-svc and auth-svc are configured for 100% sampling (no head-based) per PCI compliance requirements — this accounts for 554k spans/day and 6.2% of storage; verify with compliance-team that tail-based sampling on non-error traces would satisfy audit requirements before switching. Tempo compaction is healthy — last compaction 45 min ago, 0 errors, query p99 is 84 ms.",
		},

		// ─── APPS ───────────────────────────────────────────────────
		"apps/messaging/queues": {
			KPIs: []WSLeafKPI{
				k("Queues", "12"+up("+2"), "8 SQS · 4 RabbitMQ 3.13 — across prod (8) · staging (4) · total capacity 50k msg/s"),
				k("Throughput", "8,421 msg/s"+up("+12%"), bar(42, "green")+" 42% of provisioned capacity — peak today: 14.2k msg/s at 09:14 UTC"),
				k("Consumer lag", `12 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("-4 ms"), g("Healthy")+" p99 lag 84 ms — all consumer groups within SLO (<200 ms)"),
				k("Dead-letter queue", "4 msgs"+down("-8"), y("Needs review")+" — 3 from events.orders (schema validation) · 1 from notifications.email (template render error)"),
			},
			Headers: hdr("Queue", "Engine", "Depth", "Throughput", "Consumers", "Oldest msg", "DLQ", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ deadletter.global (q-5f6a7b) has 4 messages — 3 are from events.orders and failed schema validation after the v2.1 order-model migration on April 17 (the producer sends a new 'tax_breakdown' field that the consumer's JSON schema rejects). Fix: update the consumer schema to accept the new field and replay the 3 messages. The 1 remaining message is a notifications.email render failure caused by a missing template variable ('user.preferred_name') for locale 'ja-JP' — the Japanese localization bundle was deployed without this key. analytics.clickstream (q-4e5f6a) depth of 1,284 is normal for this queue — it processes in micro-batches of 500 every 200 ms and depth fluctuates between 800–2,000 during peak hours. RabbitMQ cluster health: 3 nodes, mirrored queues, disk alarm threshold at 82% — consider expanding storage before it hits 90%.",
		},
		"apps/messaging/streaming": {
			KPIs: []WSLeafKPI{
				k("Topics", "18"+up("+2"), "across 2 Kafka 3.7 clusters — prod-events (12 topics) · prod-analytics (6 topics) · total partitions: 184"),
				k("Event throughput", "412k /s"+up("+18%"), bar(62, "green")+" 62% of cluster capacity — peak today: 684k /s at 09:14 UTC during morning ingest"),
				k("Consumer lag", `<200 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("-42 ms"), g("All groups within SLO")+" — slowest: ml.feature.updates at 184 ms (single consumer group)"),
				k("Retention", "7d default · 30d audit", "hot storage: 2.4 TB NVMe · tiered to S3 after 7d — total retained: 18 TB across all tiers"),
			},
			Headers: hdr("Topic", "Cluster", "Partitions", "Throughput", "Consumer groups", "Lag", "Retention", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ ml.feature.updates (topic-3d4e5f) has a single consumer group (ml-pipeline-prod) with 184 ms lag — the consumer is bottlenecked on feature-store writes (batch upserts to DynamoDB). Adding a second consumer instance and increasing the DynamoDB WCU from 1,000 to 2,000 would reduce lag to <50 ms. The clickstream.raw topic (topic-0a1b2c) is the largest by volume at 248k events/s — the 24-partition layout was sized for 300k/s; if traffic grows past 280k/s, add 8 partitions to maintain headroom. Kafka cluster health: prod-events has 5 brokers, ISR count stable at 100%, controller on broker-0. prod-analytics has 3 brokers — broker-2 disk is at 78%, consider expanding or reducing clickstream retention from 7d to 5d to free 400 GB.",
		},
		"apps/messaging/servicebus": {
			KPIs: []WSLeafKPI{
				k("Service buses", "2"+up("+0"), "acme-events-bus (Premium · 4 MU) · acme-control-bus (Standard · 1 MU) — both in eu-west-1 with geo-DR paired to eu-central-1"),
				k("Subscriptions", "84"+up("+6"), "across 18 topics — fan-out ratio avg 4.7 subs/topic · max 12 on order.placed"),
				k("Messages (24h)", "12.4M"+up("+1.8M"), bar(48, "green")+" 48% of throughput capacity — peak: 2.1k msg/s at 09:22 UTC"),
				k("Publish → deliver", `12 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("-3 ms"), g("Within SLO")+" p99: 42 ms · includes filter evaluation + delivery — 0 dead-letter msgs in 7d"),
			},
			Headers: hdr("Bus / Topic", "Subscriptions", "Throughput", "Dead-letter", "Avg delivery", "Msg size", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ acme-events-bus is running at Premium tier with 4 messaging units — the order.placed topic peaks at 1,284 msg/s which is 64% of a single MU's throughput. If the upcoming Black Friday campaign drives >2x traffic, scale to 6 MU proactively (takes ~2 min with no downtime). The geo-DR pairing to eu-central-1 was last tested on April 1 with a successful failover drill (RTO: 42s, RPO: 0 messages lost). acme-control-bus on Standard tier handles only internal DevOps events at low volume (12/h) — no need to upgrade. All 84 subscriptions have SQL filters; 12 of them on order.placed use correlation filters which are more efficient — consider migrating the remaining 72 SQL filters to correlation filters for a ~20% throughput improvement.",
		},
		"apps/workflow/automation": {
			KPIs: []WSLeafKPI{
				k("Active workflows", "32"+up("+4"), "12 webhook-triggered · 8 event-driven · 8 cron · 4 manual — owned by 6 teams"),
				k("Runs (24h)", "8,421"+up("+842"), bar(72, "green")+" 98.6% success · 1.2% retried · 0.2% failed — avg queue wait: 1.2s"),
				k("Avg duration", `42 <span style="font-size:13px;color:var(--lf-dim)">s</span>`+down("-8s"), "p50: 12s · p95: 2.4 min · p99: 8 min — longest: daily-report-builder at 12 min"),
				k("Failed runs (7d)", "14"+down("-8"), y("12 retried successfully")+" · 2 permanent failures on daily-summary-email — auto-retry policy: 3 attempts, exponential backoff"),
			},
			Headers: hdr("Workflow", "Trigger", "Steps", "Last run", "Duration", "Runs (7d)", "Success rate", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ daily-summary-email (wf-4e5f6a) failed on April 18 at step 4 (render-template) — the Handlebars template references {{user.preferred_name}} which is null for 3 users created via SSO who bypassed the profile-completion flow. Fix: add a fallback to {{user.email}} in the template, or make preferred_name required during SSO onboarding. The workflow retried 3 times and exhausted its retry budget. The previous failure on April 12 was the same root cause. invoice-generator (wf-5f6a7b) had 1 failure this week — a timeout on the PDF rendering step when a 142-page invoice exceeded the 5-minute step timeout. Consider splitting large invoices into chunked renders or increasing the step timeout to 10 min for this workflow only.",
		},
		"apps/workflow/scheduler": {
			KPIs: []WSLeafKPI{
				k("Scheduled jobs", "84"+up("+4"), "62 cron · 18 interval · 4 one-shot — distributed across 3 scheduler nodes with leader election"),
				k("Runs (24h)", "412"+up("+28"), bar(82, "green")+" 98% success · 2% skipped (overlap guard) — 0 permanent failures"),
				k("Misfires (7d)", "0", g("Zero misfires")+" — jitter window: ±30s · overlap policy: skip · max concurrent: 3 per job"),
				k("Next scheduled", "in 4 min", "report-builder → then hourly-cache-warm in 22 min · 8 jobs in next hour"),
			},
			Headers: hdr("Job", "Schedule", "Timezone", "Last run", "Duration", "Next run", "Streak", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ All 84 scheduled jobs are healthy with zero misfires in the last 7 days. The scheduler cluster runs 3 nodes with Raft-based leader election — last leader change was 12 days ago (planned rolling restart). db-vacuum-analyze (job-4e5f6a) took 18 min last run, up from the usual 8 min — the orders table grew by 2M rows last week due to the partner integration launch. If vacuum duration exceeds 30 min, consider splitting into per-table jobs. cert-renewal-check (job-5f6a7b) found 2 certificates expiring within 30 days on its last run — both are managed by cert-manager and will auto-renew, but verify the ACME challenge DNS records are still valid. The overlap guard skipped 8 runs this week — all were hourly-cache-warm instances that overlapped when the cache backend (Redis) had a 45s latency spike on April 16.",
		},
		"apps/media/encoding": {
			KPIs: []WSLeafKPI{
				k("Encoding queue", "12 jobs"+up("+4"), "ETA 18 min — 8 video · 4 audio · fleet: 6 workers (c5.4xlarge) · GPU: 2 workers (g5.2xlarge for H.265/AV1)"),
				k("Throughput", "1,842 min/hr"+up("+12%"), bar(68, "green")+" 68% fleet utilization — GPU workers at 92% (H.265 4K backlog) · CPU workers at 54%"),
				k("Codecs supported", "H.264 · H.265 · AV1 · VP9", "audio: AAC · Opus · MP3 — HLS/DASH adaptive bitrate packaging included"),
				k("Output (24h)", `412 <span style="font-size:13px;color:var(--lf-dim)">GB</span>`+up("+84 GB"), "from 842 GB input — avg compression ratio 2.04:1 · cost: $18.40 compute + $4.12 storage"),
			},
			Headers: hdr("Job", "Input", "Profile", "Duration", "Progress", "Output size", "ETA", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ GPU workers are at 92% utilization due to a backlog of H.265 and AV1 encoding jobs — product-demo-av1 (enc-4e5f6a) is an AV1 encode that runs 6x slower than H.264 at the same resolution, consuming one GPU worker for ~42 min. Consider adding a 3rd g5.2xlarge worker during peak hours (estimated cost: $3.80/hr) or offering AV1 as an async/overnight-only option. training-series-batch (enc-5f6a7b) is a 24-file batch job that will take ~2h 14m at current queue depth — it was submitted as low-priority and will yield GPU slots to higher-priority jobs. The H.264 codec accounts for 72% of all encodes, H.265 for 18%, and AV1 for 10% — AV1 adoption is growing 4%/month as browser support improves.",
		},
		"apps/media/livestream": {
			KPIs: []WSLeafKPI{
				k("Live channels", "3"+up("+1"), "RTMP/SRT ingest → HLS/DASH packaging — edge PoPs: 42 worldwide · origin: eu-west-1 + us-east-1 failover"),
				k("Concurrent viewers", "10,675"+up("+2,254"), bar(54, "green")+" 54% of licensed capacity (20k) — peak today: 12.4k during keynote opening"),
				k("Peak bitrate", `12 <span style="font-size:13px;color:var(--lf-dim)">Mbps</span>`, "1080p60 source · ABR ladder: 360p/720p/1080p — total egress: 84 Gbps across CDN"),
				k("Glass-to-glass latency", `4 <span style="font-size:13px;color:var(--lf-dim)">s</span>`+down("-0.8s"), g("Low-latency HLS")+" LL-HLS with CMAF chunks · p99: 6.2s — DVR window: 4h per channel"),
			},
			Headers: hdr("Channel", "Source", "Protocol", "Profile", "Viewers", "Bitrate", "Latency", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ live-event-keynote (ch-0a1b2c) has 8,421 concurrent viewers at 4K — total CDN egress is 64 Gbps, which is 76% of the contracted CDN capacity (84 Gbps). If viewership exceeds 12k concurrent (expected during the Q&A segment), the CDN will auto-scale but cost increases from $0.02/GB to $0.035/GB for burst capacity. Consider pre-warming additional PoPs in APAC where 28% of viewers are located. The LL-HLS latency of 3.8s is excellent — the SRT ingest path adds 0.4s less latency than RTMP. live-internal-allhands (ch-2c3d4e) is using standard HLS (not low-latency) which is fine for internal use but adds ~8s latency if interactive Q&A is needed. live-product-launch (ch-4e5f6a) is configured for 4K60 which requires the g5.2xlarge transcoder — ensure the GPU worker is reserved before the event.",
		},
		"apps/media/storage": {
			KPIs: []WSLeafKPI{
				k("Media stored", `8.4 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+up("+420 GB"), "6.2 TB video · 1.8 TB audio · 0.4 TB images — hot tier: 2.1 TB (NVMe) · warm: 4.3 TB (S3 Standard) · cold: 2.0 TB (Glacier)"),
				k("Assets", "1,284"+up("+42"), "842 video · 384 audio · 58 image sets — 92% have CDN-ready transcodes · 8% raw-only (pending encoding)"),
				k("CDN bandwidth (24h)", `4.8 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+up("+0.6 TB"), "$33.60 — cache hit ratio: 94.2% · origin pulls: 5.8% · top asset: intro-video.mp4 (18% of bandwidth)"),
				k("Playback start p50", `1.2 <span style="font-size:13px;color:var(--lf-dim)">s</span>`+down("-0.2s"), g("Within SLO")+" p95: 2.8s · buffering ratio: 0.4% — preload enabled on top 50 assets"),
			},
			Headers: hdr("Asset", "Type", "Size", "Codec", "Plays (7d)", "Bandwidth", "CDN hit", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ legacy-archive-2024 (ast-5f6a7b) accounts for 24% of total storage (2 TB) but only 0.6% of plays — it was moved to Glacier cold tier on March 1, saving $42/month. Access latency from cold tier is 3–5 hours; if any of these assets are needed for the upcoming anniversary campaign, restore them to warm tier proactively (estimated restore cost: $8.40 for 2 TB). training-library (ast-4e5f6a) has a CDN hit ratio of only 78% because the 84 HLS packages generate thousands of segment files with low individual popularity — consider enabling origin shield to reduce origin pulls by ~40%. The intro-video.mp4 (ast-0a1b2c) is the most-played asset at 12,841 plays/week and is served from hot tier with 96% CDN cache hit — no action needed. Total storage cost: $124/month (hot: $48 · warm: $52 · cold: $24).",
		},

		// ─── MULTICLOUD ─────────────────────────────────────────────
		"multicloud/hybrid/onprem": {
			KPIs: []WSLeafKPI{
				k("On-prem sites", "3", "Frankfurt HQ · Manhattan DR · Tokyo office — total rack units: 42 · power: 128 kW · PUE: 1.4 avg"),
				k("Cloud connectivity", g("All 3 up"), "FRA + NYC: 10 Gbps AWS Direct Connect (redundant) · TYO: 1 Gbps IPsec VPN (single) — failover tested April 1"),
				k("Workloads", "84"+up("+6"), "42 on FRA (legacy ERP, databases) · 12 on NYC (DR replicas) · 30 on TYO (edge inference, local compliance)"),
				k("Sync latency", `1.2 <span style="font-size:13px;color:var(--lf-dim)">ms</span>`+down("-0.3 ms"), g("FRA → eu-west-1")+" 1.2 ms · NYC → us-east-1 0.8 ms · TYO → ap-northeast-1 18 ms (VPN overhead)"),
			},
			Headers: hdr("Site", "Location", "Hosts", "vCPUs", "Workloads", "Link", "Bandwidth", "Latency", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ Tokyo office (site-2c3d4e) is connected via a single 1 Gbps IPsec VPN tunnel — there is no redundant link, so a tunnel failure would isolate 30 workloads including the edge inference pipeline and Japan-region compliance data. Recommendation: provision a second VPN tunnel through a different ISP (estimated cost: $840/month) or upgrade to AWS Direct Connect 1 Gbps (estimated cost: $1,200/month, reduces latency from 18 ms to ~3 ms). The 18 ms latency is acceptable for current workloads but would become a bottleneck if real-time inference traffic (currently 200 req/s) exceeds 1,000 req/s. Frankfurt HQ hardware refresh is scheduled for Q3 — 12 servers reaching EOL in August will be replaced with next-gen AMD EPYC nodes (estimated 40% vCPU density improvement). DR failover drill from FRA to NYC completed successfully on April 1 with RTO of 8 minutes and RPO of 0 (synchronous replication).",
		},
		"multicloud/hybrid/control": {
			KPIs: []WSLeafKPI{
				k("Managed clusters", "12"+up("+1"), "4 AWS EKS · 3 Azure AKS · 2 GCP GKE · 2 on-prem k3s · 1 edge — all managed via unified control plane v2.8"),
				k("Total pods", "842"+up("+62"), bar(58, "green")+" 58% of cluster capacity — prod: 412 pods · staging: 284 pods · dev: 146 pods"),
				k("Policies enforced", "84 via OPA"+up("+4"), g("98.4% compliant")+" across all clusters — 8 resources drifted, auto-remediated within 2 min"),
				k("Control plane API (24h)", "248k calls"+up("+12%"), "p99 latency: 42 ms · 0 errors — Raft consensus healthy across 3 control-plane nodes"),
			},
			Headers: hdr("Cluster", "Provider", "Region", "Nodes", "Pods", "Policies", "Config drift", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ on-prem-fra (cls-4e5f6a) has 2 config drift items: (1) the resource-quota policy was manually overridden on April 16 to allow the ERP batch job to use 64 GB RAM instead of 32 GB — this should be formalized as a policy exception rather than a manual override; (2) a network policy allowing port 5432 ingress from 0.0.0.0/0 was added for database debugging and not removed — this is a security risk and should be scoped to the admin CIDR only. 4 policies are marked N/A on on-prem-fra because they reference cloud-specific resources (IAM roles, cloud KMS) that don't apply to bare metal. The edge-retail cluster (cls-5f6a7b) manages 24 pods across 12 retail sites — each site runs a 2-pod stack (POS sync + inventory cache) with local persistence. Control plane sync latency to edge sites averages 2.4s, well within the 10s SLO.",
		},
		"multicloud/hybrid/edge": {
			KPIs: []WSLeafKPI{
				k("Edge sites", "248"+up("+12"), "across 12 countries — 142 retail · 84 industrial · 22 outdoor/smart-city — managed via fleet orchestrator v3.2"),
				k("Managed devices", "8,421"+up("+284"), "4,842 RPi 5 · 2,184 Jetson Orin · 1,395 industrial PCs — OTA update success rate: 99.4%"),
				k("Edge workloads", "42 apps"+up("+3"), "POS sync · ML inference · video analytics · sensor ingestion — avg container image size: 84 MB (optimized)"),
				k("Fleet uptime", "99.7%"+up("+0.1%"), bar(82, "green")+" — 24 devices offline (0.3%): 18 scheduled maintenance · 4 network issues · 2 hardware failures"),
			},
			Headers: hdr("Site group", "Type", "Country", "Devices", "Workload", "Connectivity", "Uptime (30d)", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ Factory — Stuttgart (edge-2c3d4e) has 98.2% uptime, below the 99.5% SLO — the LTE fallback link drops for 2–4 minutes during shift changes (06:00, 14:00, 22:00) when 200+ workers' phones saturate the cell tower. The primary Ethernet link is stable but the factory's managed switch (HP ProCurve 2920) has a known firmware bug causing port flaps under broadcast storms from legacy SCADA equipment. Fix: update switch firmware to v16.10.0023 (vendor-validated for this environment) and add broadcast storm control. The 2 hardware failures are both in the London cluster — RPi units rpi-lon-042 and rpi-lon-087 have SD card corruption; replacement units shipped April 18, ETA April 21. OTA update v3.2.1 is staged for rollout to all retail sites on April 22 — it includes a fix for the intermittent POS sync timeout that affected 0.02% of transactions.",
		},
		"multicloud/mgmt/cost": {
			KPIs: []WSLeafKPI{
				k("Total spend (MTD)", "$48,210"+down("-3.2%"), "vs. $49,820 last month — budget: $65k · "+bar(74, "green")+" 74% of monthly budget consumed at 63% of month"),
				k("Forecast (EOM)", "$62,400"+down("-$2.8k"), g("Within budget")+" — linear forecast based on 19d actual + historical seasonality · confidence: ±$1.8k"),
				k("Savings identified", "$2,420/mo"+up("+$840"), "RI coverage gaps: $1,200 · idle resources: $820 · right-sizing: $400 — action required on 8 recommendations"),
				k("Unit economics", "$0.042 / 1k requests"+down("-8%"), "improved by 8% via spot instances and reserved capacity — cost per active user: $2.14/mo"),
			},
			Headers: hdr("Cloud / Account", "Spend (MTD)", "% of total", "Trend", "Budget", "Savings avail.", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ GCP ml-training (acct-1b2c3d) is trending 8% above last month — the increase is driven by 4 A100 GPU instances (g2-standard-96) that ran for 72 hours during a fine-tuning experiment and were not terminated after completion. Those idle GPUs cost $1,200 over the weekend. Recommendation: enable auto-shutdown after 2 hours of <5% GPU utilization on all ml-* projects. Azure analytics (acct-2c3d4e) is trending 12% up due to the new Synapse Serverless SQL pool processing 40% more data after the partner integration launch — this is expected growth, but consider switching to reserved capacity if the trend holds for another month (estimated savings: $840/month). AWS staging (acct-4e5f6a) dropped 18% after implementing the night/weekend shutdown policy on April 8 — savings are tracking to $400/month as projected.",
		},
		"multicloud/mgmt/policy": {
			KPIs: []WSLeafKPI{
				k("Total policies", "84"+up("+4"), "42 OPA Rego · 28 Sentinel · 14 custom — enforced across AWS, Azure, GCP, and on-prem environments"),
				k("Compliance rate", "98.4%"+up("+0.2%"), bar(98, "green")+" — 1,264 compliant / 1,284 total resources · 20 violations across 3 bundles · 8 auto-remediated"),
				k("Policy bundles", "6 active", "ACME-BASE v3 (global) · PROD-HARD-v2 · DATA-PRIVACY · ML-TRAINING v2 · ON-PREM-BASE · EDGE-MINIMAL"),
				k("Evaluation cycle", "continuous — 12s avg", g("Real-time")+" OPA decision log: 248k evaluations/24h · avg decision time: 2.1 ms · 0 evaluation errors"),
			},
			Headers: hdr("Policy bundle", "Scope", "Engine", "Resources", "Compliant %", "Violations", "Last eval", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ ON-PREM-BASE (pol-4e5f6a) has 5 violations on the Frankfurt on-prem cluster: (1-2) two network policies allow ingress from 0.0.0.0/0 on ports 5432 and 8080 — these were added during the April 16 database debugging session and not reverted. (3) resource-quota override on the ERP namespace exceeds the 32 GB RAM limit (currently set to 64 GB). (4-5) two pods running as root in the legacy namespace, which violates the no-root-containers policy. Auto-remediation is disabled for ON-PREM-BASE because manual approval is required per the change management process — create change tickets for items 1-3 (security risk) and schedule items 4-5 for the next maintenance window. DATA-PRIVACY (pol-2c3d4e) has 2 violations: an S3 bucket (analytics-raw-exports) lacks server-side encryption, and an Azure Blob container (staging-data-lake) has public access enabled. Both were flagged 48 hours ago — escalate if not resolved by April 21.",
		},
		"multicloud/mgmt/workload": {
			KPIs: []WSLeafKPI{
				k("Portable workloads", "84"+up("+6"), "OCI images + Terraform/Pulumi IaC — 72 containerized (K8s) · 8 VM-based · 4 serverless — all with cloud-agnostic config"),
				k("Cross-cloud moves (YTD)", "12"+up("+3"), "8 completed · 3 in progress · 1 planned — avg move time: 4.2 hours including data sync and DNS cutover"),
				k("Avg migration time", `4.2 <span style="font-size:13px;color:var(--lf-dim)">hr</span>`+down("-1.8 hr"), "includes IaC provision (20 min) + data sync (2–3 hr) + DNS cutover (10 min) + validation (30 min)"),
				k("Cost saved by re-homing", "$8,420"+up("+$2.4k"), "since Jan — ml-batch-training to GCP saved $4.2k (GPU pricing) · cdn-static-eu to Cloudflare saved $1.8k (bandwidth)"),
			},
			Headers: hdr("Workload", "Source", "Target", "Reason", "Duration", "Data moved", "Cost impact", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ dev-cluster (wl-3d4e5f) migration from AWS EKS to DigitalOcean K8s is 62% complete — IaC provisioning and application deployment are done, currently syncing persistent volumes (120 GB, ETA 2 hours). The cost savings estimate of $840/month is based on equivalent node specs (4x s-4vcpu-8gb at $48/mo vs 4x t3.large at $60/mo + EKS control plane fee $74/mo). DNS cutover is scheduled for April 21 during the maintenance window. analytics-pipeline (wl-4e5f6a) move from Azure Synapse to GCP BigQuery is planned for Q2 — the primary driver is 3x faster query performance on the ML feature-generation workload, but the 4.8 TB data transfer will take ~12 hours over a dedicated 10 Gbps cross-cloud link. The legacy-erp move to on-prem (wl-1b2c3d) completed on March 28 — post-move performance is 15% better due to lower network latency to the Frankfurt database cluster.",
		},
		"multicloud/migration/assessment": {
			KPIs: []WSLeafKPI{
				k("Apps assessed", "84"+up("+8"), "full portfolio scan completed April 12 — agent-based discovery on 84 apps · 12 databases · 42 middleware components"),
				k("Recommendation split", "62 lift+shift · 18 refactor · 4 rebuild", bar(74, "green")+" 74% can move with minimal changes — 6R classification: rehost 62 · re-platform 8 · refactor 10 · rebuild 4"),
				k("Estimated savings", "$24k/mo"+up("+$3.2k"), "vs. current on-prem TCO of $68k/mo — payback period: 8 months including migration costs of $142k one-time"),
				k("Confidence score", "72 / 84 scored 8+", g("High confidence")+" — 12 apps scored 5–7 (need manual review for licensing, compliance, or custom hardware dependencies)"),
			},
			Headers: hdr("Application", "Stack", "Recommendation", "Effort", "Savings/mo", "Confidence", "Dependencies", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ custom-stack-app (app-3d4e5f) is blocked on licensing — the application uses a USB hardware dongle for DRM that cannot be virtualized in the cloud. Options: (1) negotiate a cloud-compatible license with the vendor (estimated 6–8 weeks, unknown cost), (2) rebuild the DRM layer using a cloud-native licensing service (20 weeks, $48k dev cost), or (3) keep this app on-prem permanently and expose via API gateway. monolith-erp (app-0a1b2c) is the highest-value migration at $2,400/mo savings but requires careful planning — it has 4 downstream API consumers and an Oracle DB dependency that needs to be migrated to Aurora PostgreSQL (separate data migration workstream). The 12 apps with confidence scores 5–7 should be reviewed by the architecture team before migration planning begins — schedule a review session by May 1.",
		},
		"multicloud/migration/data": {
			KPIs: []WSLeafKPI{
				k("Active migrations", "8"+up("+2"), "5 database · 2 file storage · 1 data warehouse — methods: DMS/CDC (5), Snowball (1), ETL bulk (2)"),
				k("Data moved (24h)", `4.2 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`+up("+0.8 TB"), bar(78, "green")+" 78% of total 48 TB migrated — transfer rate: 180 MB/s avg via Direct Connect"),
				k("Avg cutover time", `12 <span style="font-size:13px;color:var(--lf-dim)">min</span>`+down("-4 min"), g("CDC-based cutovers")+" — zero-downtime for 5 databases using change data capture · max cutover: 28 min (warehouse)"),
				k("Remaining to migrate", `10.6 <span style="font-size:13px;color:var(--lf-dim)">TB</span>`, "ETA: 12 days at current rate — on track for May 1 completion target · $840 remaining transfer cost"),
			},
			Headers: hdr("Migration", "Source", "Target", "Method", "Data size", "Transferred", "Lag", "Progress", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ warehouse-to-snowflake (mig-3d4e5f) is progressing slower than expected — the ETL bulk load is at 42% after 8 days (target was 60%). The bottleneck is the delta sync phase: Vertica's change tracking requires a full table scan on 3 of the 12 largest tables (partitioned by date but lacking change timestamps). Options: (1) add a last_modified column and backfill (2-day effort, then delta sync drops from 4h to 15 min), (2) accept the 4h lag and do the cutover during the weekend maintenance window when lag can drain to zero. users-pg-to-aurora (mig-1b2c3d), mongo-to-documentdb (mig-4e5f6a), and redis-to-elasticache (mig-5f6a7b) are all at 100% sync and ready for cutover — schedule the cutover during the next maintenance window (April 22, 02:00–06:00 UTC). orders-pg-to-aurora (mig-0a1b2c) is at 94% initial load with CDC active — replication lag is < 1s, ready for cutover once the dependent application migration (monolith-erp) is also prepared.",
		},
		"multicloud/migration/app": {
			KPIs: []WSLeafKPI{
				k("In-flight migrations", "5"+up("+1"), "3 lift+shift · 1 refactor · 1 strangler pattern — tracked via migration runbooks with automated health checks"),
				k("Completed (12 mo)", "142"+up("+18"), g("100% successful")+" — 0 failed cutovers · 0 rollbacks — avg migration: 4.2 weeks planning + 2.1 weeks execution"),
				k("Avg cutover downtime", `<10 <span style="font-size:13px;color:var(--lf-dim)">min</span>`+down("-3 min"), g("Blue/green deploys")+" — DNS-weighted cutover with automated rollback trigger at >1% error rate"),
				k("Failed cutovers", "0", g("Zero rollbacks in 12 mo")+" — pre-cutover validation: synthetic tests + canary traffic at 5% for 30 min before full switch"),
			},
			Headers: hdr("Migration", "Source", "Target", "Type", "Start date", "Cutover plan", "Progress", "Status", ""),
			Rows:    [][]template.HTML{},
			Note:    "⚠️ erp-refactor-microservices (mig-2c3d4e) is 1 week behind schedule — the payment-service extraction (service 3 of 8) is blocked on the Oracle-to-Aurora database migration (see multicloud/migration/data). The team cannot test the refactored payment-service against Aurora until the CDC replication reaches steady state with <1s lag. Estimated slip: cutover moves from Jun 1 to Jun 8. monolith-decompose-strangler (mig-3d4e5f) is using the strangler fig pattern to incrementally extract services from a .NET monolith — 4 of 14 services have been extracted so far. Risk: the remaining 10 services share a single database with tightly coupled stored procedures. The architecture team recommends extracting the 6 read-heavy services first (weeks 12–20) and deferring the 4 write-heavy services to Phase 2 (Q3). legacy-app-containerise (mig-0a1b2c) has completed Dockerfile creation, CI pipeline integration, and staging validation — production cutover is scheduled for April 22 during the maintenance window.",
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Database leaf page helpers
// ─────────────────────────────────────────────────────────────────────────────

func statusBadgeDB(status string) string {
	switch status {
	case "running", "active":
		return g(status)
	case "creating", "provisioning":
		return y(status)
	case "maintenance":
		return y("maintenance")
	case "degraded", "failed", "error":
		return e(status)
	case "stopped", "paused", "sleeping":
		return gr(status)
	default:
		return status
	}
}

// databases/relational/mysql-pg
func dbRelationalMySQLPGLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListRelationalDbs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbRelationalMySQLPGLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		replicaStr := fmt.Sprintf("%d", inst.Replicas)
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("TLS+enc")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.Version,
			inst.Size,
			fmt.Sprintf("%d GB", inst.StorageGB),
			replicaStr,
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Instance", "Engine", "Version", "Size", "Storage", "Replicas", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/relational/sql-server
func dbRelationalSQLServerLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSQLServerDbs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbRelationalSQLServerLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("TLS+enc")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Edition),
			inst.Version,
			inst.Size,
			fmt.Sprintf("%d GB", inst.StorageGB),
			fmt.Sprintf("%d", inst.Replicas),
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Instance", "Edition", "Version", "Size", "Storage", "Replicas", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/relational/cloud-native
func dbRelationalCloudNativeLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListCloudNativeDbs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbRelationalCloudNativeLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		autoScale := gr("off")
		if inst.AutoScaling {
			autoScale = g("on")
		}
		suStr := inst.ServerlessUnits
		if suStr == "" {
			suStr = "—"
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.Version,
			inst.Size,
			fmt.Sprintf("%d GB used", inst.StorageUsedGB),
			autoScale,
			suStr,
			inst.Region,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Instance", "Engine", "Version", "Size", "Storage", "Auto-scale", "Serverless units", "Region", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/nosql/document
func dbNoSQLDocumentLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDocumentDbs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNoSQLDocumentLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("encrypted")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.Version,
			inst.Size,
			fmt.Sprintf("%d GB", inst.StorageGB),
			fmt.Sprintf("%d", inst.Collections),
			fmt.Sprintf("%d", inst.DocumentCount),
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Instance", "Engine", "Version", "Size", "Storage", "Collections", "Documents", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/nosql/keyvalue
func dbNoSQLKeyValueLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListKeyValueStores(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNoSQLKeyValueLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("TLS+enc")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.Version,
			inst.Size,
			fmt.Sprintf("%d GB", inst.MaxMemoryGB),
			fmt.Sprintf("%d", inst.KeyCount),
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Instance", "Engine", "Version", "Size", "Max memory", "Keys", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/nosql/wide-column
func dbNoSQLWideColumnLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListWideColumnDbs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNoSQLWideColumnLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("TLS+enc")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.Version,
			fmt.Sprintf("%d nodes", inst.Nodes),
			fmt.Sprintf("%d", inst.Keyspaces),
			fmt.Sprintf("%d GB", inst.StorageGB),
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Cluster", "Engine", "Version", "Nodes", "Keyspaces", "Storage", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/nosql/graph
func dbNoSQLGraphLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListGraphDbs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNoSQLGraphLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("TLS+enc")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.Version,
			inst.Size,
			fmt.Sprintf("%d", inst.NodeCount),
			fmt.Sprintf("%d", inst.EdgeCount),
			fmt.Sprintf("%d GB", inst.StorageGB),
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Graph", "Engine", "Version", "Size", "Nodes", "Edges", "Storage", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/memory/cache
func dbMemoryCacheLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListMemoryCaches(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbMemoryCacheLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("TLS+enc")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.Version,
			inst.Size,
			fmt.Sprintf("%d", inst.Nodes),
			fmt.Sprintf("%d GB", inst.MaxMemoryGB),
			inst.EvictionPolicy,
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Instance", "Engine", "Version", "Size", "Nodes", "Max memory", "Eviction", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/memory/session
func dbMemorySessionLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListMemorySessions(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbMemorySessionLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("TLS+enc")
		}
		ttlStr := fmt.Sprintf("%d s", inst.TTLSeconds)
		if inst.TTLSeconds >= 3600 {
			ttlStr = fmt.Sprintf("%d h", inst.TTLSeconds/3600)
		} else if inst.TTLSeconds >= 60 {
			ttlStr = fmt.Sprintf("%d min", inst.TTLSeconds/60)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.Version,
			inst.Size,
			fmt.Sprintf("%d", inst.Nodes),
			fmt.Sprintf("%d GB", inst.MaxMemoryGB),
			ttlStr,
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Store", "Engine", "Version", "Size", "Nodes", "Max memory", "TTL", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/analytics/warehouse
func dbAnalyticsWarehouseLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDataWarehouses(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAnalyticsWarehouseLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("encrypted")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.Version,
			inst.Size,
			fmt.Sprintf("%d TB", inst.StorageTB),
			fmt.Sprintf("%d", inst.Tables),
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Warehouse", "Engine", "Version", "Size", "Storage", "Tables", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/analytics/timeseries
func dbAnalyticsTimeSeriesLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListTimeSeriesDbs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAnalyticsTimeSeriesLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("TLS+enc")
		}
		retStr := fmt.Sprintf("%d d", inst.RetentionDays)
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.Version,
			inst.Size,
			retStr,
			fmt.Sprintf("%d GB", inst.StorageGB),
			fmt.Sprintf("%d", inst.MetricsCount),
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Instance", "Engine", "Version", "Size", "Retention", "Storage", "Metrics", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/analytics/search
func dbAnalyticsSearchLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSearchClusters(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAnalyticsSearchLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("TLS+enc")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.Version,
			inst.Size,
			fmt.Sprintf("%d", inst.Nodes),
			fmt.Sprintf("%d GB", inst.StorageGB),
			fmt.Sprintf("%d", inst.Indices),
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Cluster", "Engine", "Version", "Size", "Nodes", "Storage", "Indices", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/vnet/vpc
func dbNetVPCLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListVPCs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetVPCLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("encrypted")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			m(inst.CIDR),
			inst.Region,
			fmt.Sprintf("%d", inst.Subnets),
			inst.DNS,
			fmt.Sprintf("%d", inst.Peerings),
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("VPC", "CIDR", "Region", "Subnets", "DNS", "Peerings", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/vnet/peering
func dbNetPeeringLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListVPCPeerings(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetPeeringLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		encBadge := gr("plain")
		if inst.Encrypted {
			encBadge = g("encrypted")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.LocalVPC,
			inst.RemoteVPC,
			inst.Direction,
			inst.Bandwidth,
			inst.Region,
			encBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Peering", "Local VPC", "Remote VPC", "Direction", "Bandwidth", "Region", "Encryption", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/vnet/nsg
func dbNetNSGLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListNetworkSecurityGroups(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetNSGLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		loggingBadge := gr("disabled")
		if inst.Logging {
			loggingBadge = g("enabled")
		}
		actionBadge := g(inst.DefaultAction)
		if inst.DefaultAction == "allow" {
			actionBadge = y("allow")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.VPCName,
			inst.Region,
			fmt.Sprintf("%d", inst.InboundRules),
			fmt.Sprintf("%d", inst.OutboundRules),
			fmt.Sprintf("%d", inst.AssociatedNICs),
			actionBadge,
			loggingBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Edit →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("NSG", "VPC", "Region", "Inbound", "Outbound", "NICs", "Default action", "Logging", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/lb/app
func dbNetLBAppLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListLBApps(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetLBAppLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		wafBadge := gr("off")
		if inst.WAF {
			wafBadge = g("on")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Scheme,
			inst.Protocol,
			fmt.Sprintf("%d", inst.Port),
			inst.Region,
			fmt.Sprintf("%d", inst.Targets),
			inst.HealthCheck,
			wafBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Load Balancer", "Scheme", "Protocol", "Port", "Region", "Targets", "Health check", "WAF", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/lb/network
func dbNetLBNetworkLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListLBNetworks(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetLBNetworkLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		czBadge := gr("off")
		if inst.CrossZone {
			czBadge = g("on")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Scheme,
			inst.Protocol,
			fmt.Sprintf("%d", inst.Port),
			inst.Region,
			fmt.Sprintf("%d", inst.Targets),
			inst.Throughput,
			czBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Load Balancer", "Scheme", "Protocol", "Port", "Region", "Targets", "Throughput", "Cross-zone", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/lb/global
func dbNetLBGlobalLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListLBGlobals(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetLBGlobalLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Protocol,
			inst.AnycastIP,
			fmt.Sprintf("%d", inst.Regions),
			fmt.Sprintf("%d", inst.Backends),
			inst.Routing,
			inst.Failover,
			inst.Latency,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Load Balancer", "Protocol", "Anycast IP", "Regions", "Backends", "Routing", "Failover", "Latency", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/dns/manage
func dbNetDNSZoneLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDNSZones(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetDNSZoneLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		dnssecBadge := gr("disabled")
		if inst.DNSSEC == "enabled" {
			dnssecBadge = g("enabled")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.ZoneType,
			inst.Region,
			fmt.Sprintf("%d", inst.Records),
			inst.TTL,
			dnssecBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Zone", "Type", "Region", "Records", "TTL", "DNSSEC", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/dns/domains
func dbNetDomainLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDomainRegistrations(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetDomainLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		renewBadge := gr("manual")
		if inst.AutoRenew {
			renewBadge = g("auto")
		}
		privBadge := gr("off")
		if inst.Privacy {
			privBadge = g("on")
		}
		expires := "—"
		if inst.ExpiresAt != nil {
			expires = inst.ExpiresAt.Format("2006-01-02")
		}
		rows[i] = r(
			ni(inst.Domain, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.TLD,
			inst.Registrar,
			expires,
			renewBadge,
			privBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Domain", "TLD", "Registrar", "Expires", "Auto-renew", "Privacy", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/cdn/static
func dbNetCDNStaticLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListCDNStatic(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetCDNStaticLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		sslBadge := gr("off")
		if inst.SSL {
			sslBadge = g("on")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Domain,
			inst.Origin,
			inst.CachePolicy,
			inst.Bandwidth,
			inst.HitRate,
			sslBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Distribution", "Domain", "Origin", "Cache policy", "Bandwidth", "Hit rate", "SSL", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/cdn/video
func dbNetCDNVideoLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListCDNVideo(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetCDNVideoLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		drmBadge := gr("off")
		if inst.DRM {
			drmBadge = g("on")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Domain,
			inst.Protocol,
			inst.Transcoding,
			inst.Viewers,
			inst.Bandwidth,
			drmBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Stream", "Domain", "Protocol", "Transcoding", "Viewers", "Bandwidth", "DRM", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/cdn/edge
func dbNetCDNEdgeLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListCDNEdge(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetCDNEdgeLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		wafBadge := gr("off")
		if inst.WAF {
			wafBadge = g("on")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Region,
			fmt.Sprintf("%d", inst.POPs),
			fmt.Sprintf("%d", inst.Rules),
			inst.CacheSize,
			inst.HitRate,
			wafBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Edge config", "Region", "PoPs", "Rules", "Cache size", "Hit rate", "WAF", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/private/vpn
func dbNetVPNGatewayLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListVPNGateways(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetVPNGatewayLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.VPNType,
			inst.Region,
			inst.PublicIP,
			inst.RemoteIP,
			fmt.Sprintf("%d", inst.Tunnels),
			inst.Encryption,
			inst.IKEVersion,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Gateway", "Type", "Region", "Public IP", "Remote IP", "Tunnels", "Encryption", "IKE", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/private/direct
func dbNetDirectConnectLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDirectConnects(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetDirectConnectLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		macsecBadge := gr("off")
		if inst.MACsec {
			macsecBadge = g("on")
		}
		vlanBGP := fmt.Sprintf("VLAN %d · AS %s", inst.VLAN, inst.BGPASN)
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Location,
			inst.Bandwidth,
			vlanBGP,
			inst.Redundancy,
			inst.Latency,
			macsecBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Circuit", "Location", "Bandwidth", "VLAN / BGP ASN", "Redundancy", "Latency", "MACsec", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/private/sdwan
func dbNetSDWANLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSDWANNetworks(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbNetSDWANLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		qosBadge := gr("off")
		if inst.QoS {
			qosBadge = g("on")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Topology,
			inst.Region,
			fmt.Sprintf("%d", inst.Sites),
			fmt.Sprintf("%d", inst.Hubs),
			inst.Bandwidth,
			inst.Protocol,
			qosBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("Network", "Topology", "Region", "Sites", "Hubs", "Bandwidth", "Protocol", "QoS", "Status", ""),
		Rows:    rows,
	}, true
}

// ai/training/distributed
func dbAIDistributedTrainingLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIDistributedTraining(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIDistributedTrainingLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		gpus := fmt.Sprintf("%d× %s", inst.Nodes*inst.GPUsPerNode, inst.GPUType)
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Framework,
			inst.Dataset,
			gpus,
			fmt.Sprintf("%d", inst.Epochs),
			inst.Progress,
			inst.Runtime,
			inst.Cost,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Logs →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Job", "Framework", "Dataset", "GPUs", "Epochs", "Progress", "Runtime", "Cost", "Status", ""),
		Rows:    rows,
	}, true
}

// ai/training/automl
func dbAIAutoMLLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIAutoMLRuns(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIAutoMLLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.TaskType,
			inst.Dataset,
			inst.TargetCol,
			fmt.Sprintf("%d", inst.Trials),
			inst.BestModel,
			inst.BestScore,
			inst.Runtime,
			inst.Cost,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Run", "Task", "Dataset", "Target", "Trials", "Best model", "Score", "Runtime", "Cost", "Status", ""),
		Rows:    rows,
	}, true
}

// ai/training/finetuning
func dbAIFineTuningLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIFineTuningJobs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIFineTuningLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		outputModel := inst.OutputModel
		if outputModel == "" {
			outputModel = "—"
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.BaseModel,
			inst.Method,
			inst.Dataset,
			fmt.Sprintf("%d", inst.Epochs),
			inst.LearningRate,
			inst.Progress,
			inst.Runtime,
			inst.Cost,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Logs →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Job", "Base model", "Method", "Dataset", "Epochs", "LR", "Progress", "Runtime", "Cost", "Status", ""),
		Rows:    rows,
	}, true
}

// ai/hosting/inference
func dbAIInferenceEndpointLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIInferenceEndpoints(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIInferenceEndpointLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		autoScaleBadge := gr("off")
		if inst.AutoScale {
			autoScaleBadge = g("on")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Model,
			inst.Framework,
			inst.InstanceType,
			fmt.Sprintf("%d", inst.Replicas),
			inst.Latency,
			inst.RPS,
			autoScaleBadge,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Endpoint", "Model", "Framework", "Instance", "Replicas", "Latency", "RPS", "Auto-scale", "Status", ""),
		Rows:    rows,
	}, true
}

// ai/hosting/batch
func dbAIBatchJobLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIBatchJobs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIBatchJobLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Model,
			inst.InputPath,
			inst.Records,
			inst.Processed,
			inst.Progress,
			inst.Runtime,
			inst.Cost,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Logs →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Job", "Model", "Input", "Records", "Processed", "Progress", "Runtime", "Cost", "Status", ""),
		Rows:    rows,
	}, true
}

// ai/hosting/edge
func dbAIEdgeDeploymentLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIEdgeDeployments(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIEdgeDeploymentLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Model,
			inst.Runtime,
			inst.Target,
			fmt.Sprintf("%d", inst.Devices),
			inst.Version,
			inst.Latency,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Deployment", "Model", "Runtime", "Target", "Devices", "Version", "Latency", "Status", ""),
		Rows:    rows,
	}, true
}

// ai/apis/vision
func dbAIVisionAPILeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIVisionAPIs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIVisionAPILeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Model,
			inst.Capability,
			inst.Version,
			inst.Requests,
			inst.Latency,
			inst.Accuracy,
			inst.Region,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Docs →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("API", "Model", "Capability", "Version", "Requests", "Latency", "Accuracy", "Region", "Status", ""),
		Rows:    rows,
	}, true
}

// ai/apis/language
func dbAILanguageAPILeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAILanguageAPIs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAILanguageAPILeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Model,
			inst.Capability,
			inst.Version,
			inst.Requests,
			inst.Latency,
			inst.Languages,
			inst.Region,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Docs →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("API", "Model", "Capability", "Version", "Requests", "Latency", "Languages", "Region", "Status", ""),
		Rows:    rows,
	}, true
}

// ai/apis/speech
func dbAISpeechAPILeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAISpeechAPIs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAISpeechAPILeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Model,
			inst.Capability,
			inst.Version,
			inst.Requests,
			inst.Latency,
			inst.Languages,
			inst.SampleRate,
			inst.Region,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Docs →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("API", "Model", "Capability", "Version", "Requests", "Latency", "Languages", "Sample Rate", "Region", "Status", ""),
		Rows:    rows,
	}, true
}

// ai/apis/recommendation
func dbAIRecommendationAPILeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIRecommendationAPIs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIRecommendationAPILeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Model,
			inst.Algorithm,
			inst.Version,
			inst.Requests,
			inst.Latency,
			inst.Catalog,
			inst.Region,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Docs →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("API", "Model", "Algorithm", "Version", "Requests", "Latency", "Catalog Size", "Region", "Status", ""),
		Rows:    rows,
	}, true
}

// ai/data/etl
func dbAIETLPipelineLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIETLPipelines(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIETLPipelineLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			strings.ToUpper(inst.PipelineType),
			inst.Source,
			inst.Destination,
			inst.Schedule,
			inst.LastRun,
			inst.Duration,
			inst.Records,
			inst.Region,
			statusBadgeDB(inst.Status),
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Pipeline", "Type", "Source", "Destination", "Schedule", "Last Run", "Duration", "Records", "Region", "Status"),
		Rows:    rows,
	}, true
}

// ai/data/streaming
func dbAIDataStreamLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIDataStreams(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIDataStreamLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Engine,
			inst.Source,
			inst.Destination,
			inst.Throughput,
			fmt.Sprintf("%d", inst.Partitions),
			inst.Retention,
			inst.Lag,
			inst.Region,
			statusBadgeDB(inst.Status),
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Stream", "Engine", "Source", "Destination", "Throughput", "Partitions", "Retention", "Lag", "Region", "Status"),
		Rows:    rows,
	}, true
}

// ai/data/catalog
func dbAIDataCatalogLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIDataCatalogAssets(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIDataCatalogLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Source,
			inst.Owner,
			inst.AssetType,
			inst.PII,
			inst.Tags,
			fmt.Sprintf("%d", inst.Columns),
			inst.AssetUpdated,
			inst.Region,
			statusBadgeDB(inst.Status),
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Asset", "Source", "Owner", "Type", "PII", "Tags", "Columns", "Updated", "Region", "Status"),
		Rows:    rows,
	}, true
}

// ai/mlops/experiments
func dbAIMLExperimentLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIMLExperiments(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIMLExperimentLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Framework,
			inst.Owner,
			inst.BestRun,
			inst.Metric,
			fmt.Sprintf("%d", inst.Runs),
			inst.Region,
			statusBadgeDB(inst.Status),
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Experiment", "Framework", "Owner", "Best Run", "Best Metric", "Runs", "Region", "Status"),
		Rows:    rows,
	}, true
}

// ai/mlops/registry
func dbAIModelRegistryLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIModelRegistry(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIModelRegistryLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var stageBadge string
		switch inst.Stage {
		case "production":
			stageBadge = g("Production")
		case "staging":
			stageBadge = b("Staging")
		case "archived":
			stageBadge = gr("Archived")
		default:
			stageBadge = y(inst.Stage)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Version,
			inst.SizeBytes,
			stageBadge,
			inst.RequestsS,
			inst.P99Latency,
			inst.Format,
			inst.Owner,
			inst.LastUsed,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Model", "Version", "Size", "Stage", "Requests/s", "p99 latency", "Format", "Owner", "Last used", ""),
		Rows:    rows,
	}, true
}

// ai/mlops/monitoring
func dbAIMLMonitorLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAIMLMonitors(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAIMLMonitorLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.DriftPSI,
			inst.Confidence,
			inst.RequestsS,
			inst.P99Latency,
			inst.ErrorRate,
			inst.LastRetrain,
			statusBadgeDB(inst.MonitorStatus),
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Model", "Drift (PSI)", "Confidence", "Requests/s", "Latency p99", "Error rate", "Last retrain", "Status", ""),
		Rows:    rows,
	}, true
}

// security/iam/sso
func dbSecuritySSOLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecuritySSOProviders(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecuritySSOLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Protocol),
			inst.Domain,
			fmt.Sprintf("%d", inst.UsersCount),
			fmt.Sprintf("%d", inst.Logins24h),
			inst.SCIM,
			inst.CertExpires,
			inst.LastRotation,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Provider", "Protocol", "Domain", "Users", "Logins (24h)", "SCIM", "Cert expires", "Last rotation", "Status", ""),
		Rows:    rows,
	}, true
}

// security/iam/mfa
func dbSecurityMFALeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecurityMFAUsers(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecurityMFALeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.PrimaryMethod,
			inst.BackupMethod,
			inst.Roles,
			inst.RecoveryCodes,
			fmt.Sprintf("%d devices", inst.Devices),
			inst.LastMFA,
			statusBadgeDB(inst.MFAStatus),
			`<button class="lf-btn lf-btn--ghost">Reset →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("User", "Primary", "Backup", "Roles", "Recovery codes", "Devices", "Last MFA", "Status", ""),
		Rows:    rows,
	}, true
}

// security/iam/rbac
func dbSecurityRBACLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecurityRBACRoles(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecurityRBACLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.RoleType),
			fmt.Sprintf("%d users", inst.UsersCount),
			inst.Permissions,
			inst.Scope,
			inst.LastModified,
			inst.ReviewStatus,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Edit →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiAI(),
		Headers: hdr("Role", "Type", "Users", "Permissions", "Scope", "Last modified", "Review status", "Status", ""),
		Rows:    rows,
	}, true
}

// security/data/enc-rest
func dbSecurityEncRestLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecurityEncRest(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecurityEncRestLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		compliance := inst.Compliance
		if compliance != "" {
			compliance = g(compliance)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.ResourceType),
			inst.Algorithm,
			m(inst.KMSKey),
			inst.KeyAge,
			inst.LastVerified,
			compliance,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Inspect →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiSec(),
		Headers: hdr("Resource", "Type", "Algorithm", "KMS key", "Key age", "Last verified", "Compliance", "Status", ""),
		Rows:    rows,
	}, true
}

// security/data/enc-transit
func dbSecurityEncTransitLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecurityEncTransit(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecurityEncTransitLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var hstsBadge string
		switch inst.HSTS {
		case "Preloaded":
			hstsBadge = g("Preloaded")
		case "Enabled":
			hstsBadge = g("Enabled")
		default:
			hstsBadge = gr(inst.HSTS)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Protocol),
			inst.Cipher,
			inst.CertIssuer,
			g(inst.Grade),
			inst.Expires,
			hstsBadge,
			inst.LastScan,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Test →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiSec(),
		Headers: hdr("Endpoint", "Protocol", "Cipher suite", "Cert issuer", "Grade", "Expires", "HSTS", "Last scan", "Status", ""),
		Rows:    rows,
	}, true
}

// security/data/kms
func dbSecurityKMSLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecurityKMSKeys(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecurityKMSLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Algorithm,
			b(inst.KeyType),
			inst.Region,
			inst.AgeDays,
			g(inst.Rotation),
			inst.Operations24h,
			fmt.Sprintf("%d grants", inst.GrantsCount),
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Rotate now →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiSec(),
		Headers: hdr("Key alias", "Algorithm", "Type", "Region", "Age", "Rotation", "Operations (24h)", "Grants", "Status", ""),
		Rows:    rows,
	}, true
}

// security/network/waf
func dbSecurityWAFLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecurityWAFRules(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecurityWAFLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var actionBadge string
		switch inst.Action {
		case "block":
			actionBadge = e("Block")
		case "challenge":
			actionBadge = y("Challenge")
		case "allow":
			actionBadge = g("Allow")
		default:
			actionBadge = y(inst.Action)
		}
		var severityBadge string
		switch inst.Severity {
		case "Critical":
			severityBadge = e("Critical")
		case "High":
			severityBadge = y("High")
		case "Info":
			severityBadge = b("Info")
		default:
			severityBadge = y(inst.Severity)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			fmt.Sprintf("%d", inst.Priority),
			actionBadge,
			inst.Source,
			inst.Matches24h,
			fmt.Sprintf("%d", inst.FalsePositives),
			severityBadge,
			inst.LastTriggered,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Edit →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiSec(),
		Headers: hdr("Rule", "Priority", "Action", "Source", "Matches (24h)", "False positives", "Severity", "Last triggered", "Status", ""),
		Rows:    rows,
	}, true
}

// security/network/ddos
func dbSecurityDDoSLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecurityDDoSProfiles(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecurityDDoSLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Method),
			inst.Threshold,
			inst.Attacks30d,
			inst.LastTriggered,
			fmt.Sprintf("%d", inst.FalsePositives),
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Configure →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiSec(),
		Headers: hdr("Protection layer", "Method", "Capacity / Threshold", "Attacks handled (30d)", "Last triggered", "False positives", "Status", ""),
		Rows:    rows,
	}, true
}

// security/network/ids
func dbSecurityIDSLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecurityIDSRules(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecurityIDSLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var sevBadge string
		switch inst.Severity {
		case "Critical":
			sevBadge = e("Critical")
		case "High":
			sevBadge = e("High")
		case "Medium":
			sevBadge = y("Medium")
		default:
			sevBadge = b(inst.Severity)
		}
		var actionBadge string
		switch inst.Action {
		case "Block + alert", "Block + log", "Block + isolate":
			actionBadge = e(inst.Action)
		case "Log + alert":
			actionBadge = b(inst.Action)
		default:
			actionBadge = b(inst.Action)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			sevBadge,
			inst.Direction,
			inst.Matches24h,
			fmt.Sprintf("%d", inst.FalsePositives),
			actionBadge,
			inst.LastTriggered,
			statusBadgeDB(inst.Status),
			`<button class="lf-btn lf-btn--ghost">Investigate →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiSec(),
		Headers: hdr("Rule / signature", "Severity", "Direction", "Matches (24h)", "False positives", "Action", "Last triggered", "Status", ""),
		Rows:    rows,
	}, true
}

// security/compliance/policy
func dbSecurityCompliancePolicyLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecurityCompliancePolicies(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecurityCompliancePolicyLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var sevBadge string
		switch inst.Severity {
		case "Critical":
			sevBadge = e("Critical")
		case "High":
			sevBadge = y("High")
		default:
			sevBadge = b(inst.Severity)
		}
		compliantCell := fmt.Sprintf("%d / %d", inst.CompliantCount, inst.TotalCount)
		var statusBadge string
		if inst.CompliantCount < inst.TotalCount {
			statusBadge = y(fmt.Sprintf("%d violations", inst.TotalCount-inst.CompliantCount))
		} else {
			statusBadge = g("Compliant")
		}
		var remediateBadge string
		switch inst.AutoRemediate {
		case "No":
			remediateBadge = gr("No")
		case "N/A — manual", "N/A — immutable":
			remediateBadge = gr(inst.AutoRemediate)
		default:
			remediateBadge = g(inst.AutoRemediate)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			sevBadge,
			inst.Framework,
			compliantCell,
			inst.Drift7d,
			remediateBadge,
			inst.LastEvaluated,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Edit →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiSec(),
		Headers: hdr("Policy", "Severity", "Framework", "Compliant / Total", "Drift (7d)", "Auto-remediate", "Last evaluated", "Status", ""),
		Rows:    rows,
	}, true
}

// security/compliance/audit
func dbSecurityComplianceAuditLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecurityAuditEvents(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecurityComplianceAuditLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var sevBadge string
		switch inst.Severity {
		case "High", "Critical":
			sevBadge = e(inst.Severity)
		case "Medium":
			sevBadge = y("Medium")
		default:
			sevBadge = b(inst.Severity)
		}
		var auditStatusBadge string
		switch inst.AuditStatus {
		case "Flagged", "Anomaly", "Under review":
			auditStatusBadge = y(inst.AuditStatus)
		case "Normal", "Expected", "Scheduled":
			auditStatusBadge = g(inst.AuditStatus)
		default:
			auditStatusBadge = b(inst.AuditStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			m(inst.Actor),
			inst.SourceIP,
			inst.Resource,
			inst.AuditRegion,
			inst.CreatedAt.Format("2 Jan 15:04"),
			sevBadge,
			auditStatusBadge,
			`<button class="lf-btn lf-btn--ghost">View →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiSec(),
		Headers: hdr("Event", "Actor", "Source IP", "Resource", "Region", "When", "Severity", "Status", ""),
		Rows:    rows,
	}, true
}

// security/compliance/regulatory
func dbSecurityRegulatoryLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListSecurityRegulatoryFrameworks(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbSecurityRegulatoryLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		controlsCell := fmt.Sprintf("%d / %d", inst.ControlsPassed, inst.ControlsTotal)
		var statusBadge string
		if inst.ControlsPassed < inst.ControlsTotal {
			statusBadge = y(fmt.Sprintf("%d findings", inst.ControlsTotal-inst.ControlsPassed))
		} else {
			statusBadge = g("Compliant")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Scope,
			controlsCell,
			inst.EvidenceItems,
			inst.LastAssessment,
			inst.Assessor,
			inst.Renewal,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Report →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiSec(),
		Headers: hdr("Framework", "Scope", "Controls passed", "Evidence items", "Last assessment", "Assessor", "Renewal", "Status", ""),
		Rows:    rows,
	}, true
}

// devops/cicd/build
func dbDevOpsCICDBuildLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDevOpsCICDBuildPipelines(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbDevOpsCICDBuildLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.BuildStatus {
		case "passing", "Success":
			statusBadge = g(inst.BuildStatus)
		case "failing", "Failed":
			statusBadge = e(inst.BuildStatus)
		default:
			statusBadge = y(inst.BuildStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			m(inst.Repo),
			b(inst.Trigger),
			inst.Duration,
			inst.Runtime,
			inst.CacheHit,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Logs →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("Pipeline", "Repo", "Trigger", "Duration", "Runner", "Cache", "Status", ""),
		Rows:    rows,
	}, true
}

// devops/cicd/deploy
func dbDevOpsCICDDeployLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDevOpsCICDDeployPipelines(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbDevOpsCICDDeployLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.DeployStatus {
		case "active", "Healthy":
			statusBadge = g("Healthy")
		case "inactive", "Rolled back":
			statusBadge = e(inst.DeployStatus)
		default:
			statusBadge = y(inst.DeployStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			m(inst.App),
			b(inst.Environment),
			inst.Strategy,
			inst.Traffic,
			inst.ErrorRate,
			inst.LatencyP99,
			inst.Duration,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("Deploy", "App", "Env", "Strategy", "Traffic", "Error rate", "Latency p99", "Duration", "Status", ""),
		Rows:    rows,
	}, true
}

// devops/iac/templates
func dbDevOpsIaCTemplatesLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDevOpsIaCTemplates(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbDevOpsIaCTemplatesLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var driftBadge string
		switch inst.DriftStatus {
		case "Clean":
			driftBadge = g("Clean")
		case "In sync":
			driftBadge = g("In sync")
		default:
			driftBadge = y(inst.DriftStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Tool),
			fmt.Sprintf("%d res", inst.ResourceCount),
			m(inst.StateBackend),
			inst.LastApplied,
			inst.PlanOutput,
			driftBadge,
			driftBadge,
			`<button class="lf-btn lf-btn--ghost">Plan →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("Stack", "Tool", "Resources", "State backend", "Last applied", "Plan output", "Drift", "Status", ""),
		Rows:    rows,
	}, true
}

// devops/iac/config
func dbDevOpsIaCConfigLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDevOpsIaCConfigs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbDevOpsIaCConfigLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.CfgStatus {
		case "In sync":
			statusBadge = g("In sync")
		case "Drift":
			statusBadge = y("Drift")
		default:
			statusBadge = y(inst.CfgStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Tool),
			fmt.Sprintf("%d hosts", inst.HostCount),
			m(inst.Inventory),
			inst.LastApply,
			inst.Changed,
			inst.Compliance,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Run →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("Bundle", "Tool", "Hosts", "Inventory", "Last apply", "Changed", "Compliance", "Status", ""),
		Rows:    rows,
	}, true
}

// devops/api/gateway
func dbDevOpsAPIGatewayLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDevOpsAPIGateways(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbDevOpsAPIGatewayLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.GwStatus {
		case "Live":
			statusBadge = g("Live")
		case "Degraded":
			statusBadge = y("Degraded")
		case "Down":
			statusBadge = e("Down")
		default:
			statusBadge = b(inst.GwStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.APIVersion,
			b(inst.Protocol),
			inst.Calls24h,
			inst.LatencyP95,
			inst.RateLimit,
			m(inst.AuthType),
			inst.ErrorRate,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("API", "Version", "Protocol", "Calls (24h)", "p95", "Rate limit", "Auth", "Error rate", "Status", ""),
		Rows:    rows,
	}, true
}

// devops/cicd/source
func dbDevOpsCICDSourceLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDevOpsCICDSourceRepos(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbDevOpsCICDSourceLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var ciStatusBadge string
		switch inst.CIStatus {
		case "Passing":
			ciStatusBadge = g("Passing")
		case "Failing", "Failed":
			ciStatusBadge = e(inst.CIStatus)
		default:
			ciStatusBadge = y(inst.CIStatus)
		}
		var protectionBadge string
		switch inst.BranchProtection {
		case "Enforced":
			protectionBadge = g("Enforced")
		default:
			protectionBadge = y(inst.BranchProtection)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			m(inst.DefaultBranch),
			b(inst.Language),
			inst.LastPush,
			fmt.Sprintf("%d", inst.PRsOpen),
			inst.Coverage,
			ciStatusBadge,
			protectionBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("Repository", "Default branch", "Language", "Last push", "PRs open", "Coverage", "CI status", "Branch protection", ""),
		Rows:    rows,
	}, true
}

// devops/api/portal
func dbDevOpsAPIPortalLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDevOpsAPIPortals(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbDevOpsAPIPortalLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.PortalStatus {
		case "Active":
			statusBadge = g("Active")
		case "Throttle risk":
			statusBadge = y("Throttle risk")
		case "Dormant — review":
			statusBadge = gr("Dormant — review")
		default:
			statusBadge = b(inst.PortalStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Owner,
			b(inst.Plan),
			inst.APIKeys,
			inst.Calls7d,
			inst.LastUsed,
			inst.QuotaUsage,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("App", "Owner", "Plan", "API keys", "Calls (7d)", "Last used", "Quota usage", "Status", ""),
		Rows:    rows,
	}, true
}

// devops/api/analytics
func dbDevOpsAPIAnalyticsLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDevOpsAPIAnalytics(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbDevOpsAPIAnalyticsLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.EpStatus {
		case "Healthy":
			statusBadge = g("Healthy")
		case "SLO breach":
			statusBadge = e("SLO breach")
		case "Watch":
			statusBadge = y("Watch")
		default:
			statusBadge = b(inst.EpStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.APIName,
			b(inst.Method),
			inst.Calls24h,
			inst.LatencyP50,
			inst.LatencyP95,
			inst.LatencyP99,
			inst.ErrorRate,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("Endpoint", "API", "Method", "Calls (24h)", "p50", "p95", "p99", "Error rate", "Status", ""),
		Rows:    rows,
	}, true
}

// devops/obs/logging
func dbDevOpsObsLoggingLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDevOpsObsLogging(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbDevOpsObsLoggingLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.LogStatus {
		case "Ingesting":
			statusBadge = g("Ingesting")
		case "Lagging":
			statusBadge = y("Lagging")
		case "Error":
			statusBadge = e("Error")
		default:
			statusBadge = b(inst.LogStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Source,
			b(inst.Format),
			inst.Volume24h,
			inst.Lines24h,
			inst.Retention,
			inst.Lag,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Tail →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("Pipeline", "Source", "Format", "Volume (24h)", "Lines (24h)", "Retention", "Lag", "Status", ""),
		Rows:    rows,
	}, true
}

// devops/obs/metrics
func dbDevOpsObsMetricsLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDevOpsObsMetrics(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbDevOpsObsMetricsLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.DashStatus {
		case "Live":
			statusBadge = g("Live")
		case "Warning":
			statusBadge = y("Warning")
		case "Stale — review":
			statusBadge = gr("Stale — review")
		default:
			statusBadge = b(inst.DashStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Owner,
			fmt.Sprintf("%d panels", inst.Panels),
			m(inst.DataSource),
			inst.RefreshRate,
			inst.LastViewed,
			inst.AlertsLinked,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("Dashboard", "Owner", "Panels", "Data source", "Refresh", "Last viewed", "Alerts linked", "Status", ""),
		Rows:    rows,
	}, true
}

// devops/obs/tracing
func dbDevOpsObsTracingLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListDevOpsObsTracing(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbDevOpsObsTracingLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.SvcStatus {
		case "Healthy":
			statusBadge = g("Healthy")
		case "Slow tail":
			statusBadge = e("Slow tail")
		case "Degraded":
			statusBadge = y("Degraded")
		default:
			statusBadge = b(inst.SvcStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Spans24h,
			inst.LatencyP50,
			inst.LatencyP95,
			inst.LatencyP99,
			inst.ErrorRate,
			inst.Sampling,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Traces →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("Service", "Spans (24h)", "p50", "p95", "p99", "Error rate", "Sampling", "Status", ""),
		Rows:    rows,
	}, true
}

// apps/messaging/queues
func dbAppsMessagingQueuesLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAppsMessagingQueues(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAppsMessagingQueuesLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.QueueStatus {
		case "active":
			statusBadge = g("Healthy")
		case "degraded":
			statusBadge = y("Degraded")
		case "error":
			statusBadge = e("Error")
		case "deprecated":
			statusBadge = gr("Deprecated")
		default:
			statusBadge = b(inst.QueueStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Engine),
			inst.QueueType,
			inst.Messages,
			inst.Throughput,
			fmt.Sprintf("%d groups", inst.Consumers),
			inst.Retention,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiApps(),
		Headers: hdr("Queue", "Engine", "Type", "Depth", "Throughput", "Consumers", "Retention", "Status", ""),
		Rows:    rows,
	}, true
}

// apps/messaging/streaming
func dbAppsMessagingStreamingLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAppsMessagingStreams(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAppsMessagingStreamingLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.StreamStatus {
		case "active":
			statusBadge = g("Healthy")
		case "lagging":
			statusBadge = y("Lag — watch")
		case "error":
			statusBadge = e("Error")
		case "deprecated":
			statusBadge = gr("Deprecated")
		default:
			statusBadge = b(inst.StreamStatus)
		}
		var lagBadge string
		if inst.StreamStatus == "lagging" {
			lagBadge = y(inst.Lag)
		} else {
			lagBadge = g(inst.Lag)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			m(inst.Engine),
			fmt.Sprintf("%d", inst.Partitions),
			inst.Throughput,
			fmt.Sprintf("%d groups", inst.ConsumerGroups),
			lagBadge,
			inst.Retention,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiApps(),
		Headers: hdr("Topic", "Engine", "Partitions", "Throughput", "Consumer groups", "Lag", "Retention", "Status", ""),
		Rows:    rows,
	}, true
}

// apps/messaging/servicebus
func dbAppsMessagingServiceBusLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAppsMessagingServiceBuses(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAppsMessagingServiceBusLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.BusStatus {
		case "active":
			statusBadge = g("Healthy")
		case "degraded":
			statusBadge = y("Degraded")
		case "error":
			statusBadge = e("Error")
		default:
			statusBadge = b(inst.BusStatus)
		}
		var dlBadge string
		if inst.DeadLetter > 0 {
			dlBadge = y(fmt.Sprintf("%d msgs", inst.DeadLetter))
		} else {
			dlBadge = g("0")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			fmt.Sprintf("%d subs", inst.Subscriptions),
			inst.Throughput,
			dlBadge,
			inst.AvgDelivery,
			inst.MsgSize,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiApps(),
		Headers: hdr("Bus / Topic", "Subscriptions", "Throughput", "Dead-letter", "Avg delivery", "Msg size", "Status", ""),
		Rows:    rows,
	}, true
}

// apps/workflow/automation
func dbAppsWorkflowAutomationLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAppsWorkflowAutomations(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAppsWorkflowAutomationLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.WfStatus {
		case "active":
			statusBadge = g("Healthy")
		case "degraded":
			statusBadge = y("Degraded")
		case "error":
			statusBadge = e("Error")
		case "deprecated":
			statusBadge = gr("Deprecated")
		default:
			statusBadge = b(inst.WfStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Trigger),
			fmt.Sprintf("%d steps", inst.Actions),
			inst.LastRun,
			"—",
			inst.Runs,
			inst.SuccessRate,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">History →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiApps(),
		Headers: hdr("Workflow", "Trigger", "Steps", "Last run", "Duration", "Runs (7d)", "Success rate", "Status", ""),
		Rows:    rows,
	}, true
}

// apps/workflow/scheduler
func dbAppsWorkflowSchedulerLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAppsWorkflowSchedulers(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAppsWorkflowSchedulerLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.JobStatus {
		case "active":
			statusBadge = g("Last ok")
		case "failed":
			statusBadge = e("Failed")
		case "paused":
			statusBadge = y("Paused")
		default:
			statusBadge = b(inst.JobStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			m(inst.Schedule),
			inst.Timezone,
			inst.LastRun,
			inst.Duration,
			inst.NextRun,
			g(inst.Streak),
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">History →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiApps(),
		Headers: hdr("Job", "Schedule", "Timezone", "Last run", "Duration", "Next run", "Streak", "Status", ""),
		Rows:    rows,
	}, true
}

// apps/media/encoding
func dbAppsMediaEncodingLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAppsMediaEncodings(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAppsMediaEncodingLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.EncodeStatus {
		case "ready":
			statusBadge = g("Ready")
		case "encoding":
			statusBadge = b("Encoding")
		case "queued":
			statusBadge = y("Queued")
		case "failed":
			statusBadge = e("Failed")
		default:
			statusBadge = gr(inst.EncodeStatus)
		}
		var progressBar string
		if inst.EncodeStatus == "queued" {
			progressBar = bar(inst.Progress, "amber")
		} else {
			progressBar = bar(inst.Progress, "")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.InputFormat,
			b(inst.Profile),
			inst.Duration,
			progressBar,
			inst.OutputSize,
			inst.ETA,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiApps(),
		Headers: hdr("Job", "Input", "Profile", "Duration", "Progress", "Output size", "ETA", "Status", ""),
		Rows:    rows,
	}, true
}

// apps/media/livestream
func dbAppsMediaLivestreamLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAppsMediaLivestreams(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAppsMediaLivestreamLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.StreamStatus {
		case "live":
			statusBadge = g("Live")
		case "scheduled":
			statusBadge = b("Scheduled")
		case "draft":
			statusBadge = gr("Draft")
		case "error":
			statusBadge = e("Error")
		default:
			statusBadge = y(inst.StreamStatus)
		}
		viewersStr := "—"
		if inst.Viewers > 0 {
			viewersStr = fmt.Sprintf("%d", inst.Viewers)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.SourceProtocol),
			inst.OutputProtocol,
			inst.Profile,
			viewersStr,
			inst.Bitrate,
			inst.Latency,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Monitor →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiApps(),
		Headers: hdr("Channel", "Source", "Protocol", "Profile", "Viewers", "Bitrate", "Latency", "Status", ""),
		Rows:    rows,
	}, true
}

// apps/media/storage
func dbAppsMediaStorageLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListAppsMediaStorageAssets(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAppsMediaStorageLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.AssetStatus {
		case "live_hot":
			statusBadge = g("Live — hot")
		case "live_warm":
			statusBadge = g("Live — warm")
		case "cold":
			statusBadge = b("Cold tier")
		case "pending":
			statusBadge = y("Pending encode")
		default:
			statusBadge = gr(inst.AssetStatus)
		}
		playsStr := fmt.Sprintf("%d", inst.Plays7d)
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.AssetType),
			inst.Size,
			inst.Codec,
			playsStr,
			inst.Bandwidth,
			inst.CDNHit,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiApps(),
		Headers: hdr("Asset", "Type", "Size", "Codec", "Plays (7d)", "Bandwidth", "CDN hit", "Status", ""),
		Rows:    rows,
	}, true
}

// multicloud/hybrid/onprem
func dbMulticloudOnpremLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListMulticloudOnpremSites(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbMulticloudOnpremLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.SiteStatus {
		case "connected":
			statusBadge = g("Connected")
		case "degraded":
			statusBadge = y("Single link")
		case "offline":
			statusBadge = e("Offline")
		default:
			statusBadge = b(inst.SiteStatus)
		}
		var linkBadge string
		if inst.SiteStatus == "degraded" {
			linkBadge = y(inst.LinkType)
		} else {
			linkBadge = b(inst.LinkType)
		}
		var latencyBadge string
		if inst.SiteStatus == "degraded" {
			latencyBadge = y(inst.Latency)
		} else {
			latencyBadge = g(inst.Latency)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Location,
			fmt.Sprintf("%d servers", inst.Hosts),
			fmt.Sprintf("%d vCPU", inst.VCPUs),
			fmt.Sprintf("%d workloads", inst.Workloads),
			linkBadge,
			inst.Bandwidth,
			latencyBadge,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiMC(),
		Headers: hdr("Site", "Location", "Hosts", "vCPUs", "Workloads", "Link", "Bandwidth", "Latency", "Status", ""),
		Rows:    rows,
	}, true
}

// multicloud/hybrid/control
func dbMulticloudControlLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListMulticloudControlClusters(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbMulticloudControlLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.ClusterStatus {
		case "healthy":
			statusBadge = g("Healthy")
		case "drift":
			statusBadge = y("Drift detected")
		case "degraded":
			statusBadge = y("Degraded")
		case "offline":
			statusBadge = e("Offline")
		default:
			statusBadge = b(inst.ClusterStatus)
		}
		var driftBadge string
		if inst.ClusterStatus == "drift" {
			driftBadge = y(inst.ConfigDrift)
		} else {
			driftBadge = g(inst.ConfigDrift)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Provider),
			inst.ClusterRegion,
			fmt.Sprintf("%d nodes", inst.Nodes),
			fmt.Sprintf("%d pods", inst.Pods),
			inst.Policies,
			driftBadge,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiMC(),
		Headers: hdr("Cluster", "Provider", "Region", "Nodes", "Pods", "Policies", "Config drift", "Status", ""),
		Rows:    rows,
	}, true
}

// multicloud/hybrid/edge
func dbMulticloudEdgeLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListMulticloudEdgeNodes(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbMulticloudEdgeLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.NodeStatus {
		case "active":
			statusBadge = g("Active")
		case "degraded":
			statusBadge = y("Degraded")
		case "offline":
			statusBadge = e("Offline")
		default:
			statusBadge = b(inst.NodeStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Location,
			b(inst.Provider),
			fmt.Sprintf("%d cores", inst.CPUCores),
			fmt.Sprintf("%d GB", inst.MemoryGB),
			fmt.Sprintf("%d", inst.Workloads),
			g(inst.Uptime),
			b(inst.Latency),
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiMC(),
		Headers: hdr("Node", "Location", "Provider", "CPU", "Memory", "Workloads", "Uptime", "Latency", "Status", ""),
		Rows:    rows,
	}, true
}

// multicloud/mgmt/cost
func dbMulticloudCostLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListMulticloudCostProfiles(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbMulticloudCostLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.ProfileStatus {
		case "active":
			statusBadge = g("Active")
		case "warning":
			statusBadge = y("Warning")
		default:
			statusBadge = b(inst.ProfileStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Provider),
			inst.Account,
			inst.MonthlyBudget,
			inst.CurrentSpend,
			inst.Forecast,
			g(inst.SavingsFound),
			fmt.Sprintf("%d", inst.Alerts),
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiMC(),
		Headers: hdr("Profile", "Provider", "Account", "Budget", "Spend", "Forecast", "Savings", "Alerts", "Status", ""),
		Rows:    rows,
	}, true
}

// multicloud/mgmt/policy
func dbMulticloudPolicyLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListMulticloudGovernancePolicies(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbMulticloudPolicyLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.PolicyStatus {
		case "active":
			statusBadge = g("Active")
		case "warning":
			statusBadge = y("Warning")
		default:
			statusBadge = b(inst.PolicyStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Category),
			inst.Scope,
			b(inst.Enforcement),
			g(fmt.Sprintf("%d", inst.Compliant)),
			e(fmt.Sprintf("%d", inst.NonCompliant)),
			fmt.Sprintf("%d", inst.Exemptions),
			inst.LastEval,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">View →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiMC(),
		Headers: hdr("Policy", "Category", "Scope", "Enforcement", "Compliant", "Non-compliant", "Exempt", "Last eval", "Status", ""),
		Rows:    rows,
	}, true
}

// multicloud/mgmt/workload
func dbMulticloudWorkloadLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListMulticloudWorkloadProfiles(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbMulticloudWorkloadLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.WorkloadStatus {
		case "active":
			statusBadge = g("Active")
		case "migrating":
			statusBadge = y("Migrating")
		default:
			statusBadge = b(inst.WorkloadStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.SourceProvider),
			b(inst.TargetProvider),
			inst.WorkloadType,
			fmt.Sprintf("%d", inst.Components),
			inst.Portability,
			inst.DriftStatus,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiMC(),
		Headers: hdr("Workload", "Source", "Target", "Type", "Components", "Portability", "Drift", "Status", ""),
		Rows:    rows,
	}, true
}

// multicloud/migration/assessment
func dbMulticloudMigrationAssessmentLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListMulticloudMigrationAssessments(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbMulticloudMigrationAssessmentLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.AssessmentStatus {
		case "active":
			statusBadge = g("Active")
		case "in_progress":
			statusBadge = y("In progress")
		case "completed":
			statusBadge = b("Completed")
		default:
			statusBadge = b(inst.AssessmentStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.SourceEnv,
			inst.TargetEnv,
			fmt.Sprintf("%d", inst.DiscoveredAssets),
			g(fmt.Sprintf("%d", inst.ReadyToMigrate)),
			y(fmt.Sprintf("%d", inst.NeedsRemediation)),
			inst.EstimatedCost,
			b(inst.Complexity),
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiMC(),
		Headers: hdr("Assessment", "Source", "Target", "Discovered", "Ready", "Remediate", "Est. cost", "Complexity", "Status", ""),
		Rows:    rows,
	}, true
}

// multicloud/migration/data
func dbMulticloudDataMigrationLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListMulticloudDataMigrationJobs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbMulticloudDataMigrationLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.JobStatus {
		case "active":
			statusBadge = g("Active")
		case "completed":
			statusBadge = b("Completed")
		case "pending":
			statusBadge = y("Pending")
		case "error":
			statusBadge = e("Error")
		default:
			statusBadge = b(inst.JobStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.SourceStore,
			inst.TargetStore,
			b(inst.DataType),
			inst.TotalSize,
			inst.Transferred,
			inst.ProgressPct,
			inst.Throughput,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiMC(),
		Headers: hdr("Job", "Source", "Target", "Type", "Total", "Transferred", "Progress", "Throughput", "Status", ""),
		Rows:    rows,
	}, true
}

// multicloud/migration/app
func dbMulticloudAppMigrationLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListMulticloudAppMigrations(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbMulticloudAppMigrationLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.MigrationStatus {
		case "active":
			statusBadge = g("Active")
		case "completed":
			statusBadge = b("Completed")
		case "pending":
			statusBadge = y("Pending")
		default:
			statusBadge = b(inst.MigrationStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.AppType),
			inst.SourcePlatform,
			inst.TargetPlatform,
			b(inst.Strategy),
			fmt.Sprintf("%d", inst.Services),
			inst.MigratedPct,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Open →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiMC(),
		Headers: hdr("Migration", "App type", "Source", "Target", "Strategy", "Services", "Progress", "Status", ""),
		Rows:    rows,
	}, true
}

// apps/app-hosting/static-sites
func dbAppsStaticSitesLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListProStaticSites(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbAppsStaticSitesLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var deployBadge string
		switch inst.DeployStatus {
		case "Live":
			deployBadge = g("Live")
		case "Not deployed":
			deployBadge = gr("Not deployed")
		default:
			deployBadge = y(inst.DeployStatus)
		}
		cdnStr := fmt.Sprintf("%d regions", inst.CDNRegions)
		if inst.CDNRegions == 0 {
			cdnStr = "—"
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Domain,
			b(inst.Framework),
			inst.Branch,
			cdnStr,
			inst.LastDeploy,
			b(inst.SSL),
			deployBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiApps(),
		Headers: hdr("Site", "Domain", "Framework", "Branch", "CDN", "Last deploy", "SSL", "Status", ""),
		Rows:    rows,
	}, true
}

// compute/virtual-machines/linux-vms
func dbComputeLinuxVMsLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListLinuxVMs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbComputeLinuxVMsLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.VMStatus {
		case "running":
			statusBadge = g("Running")
		case "stopped":
			statusBadge = gr("Stopped")
		case "error", "failed":
			statusBadge = e(inst.VMStatus)
		default:
			statusBadge = y(inst.VMStatus)
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.Distro),
			inst.Version,
			fmt.Sprintf("%d vCPU", inst.VCPUs),
			fmt.Sprintf("%d GB RAM", inst.MemoryGB),
			fmt.Sprintf("%d GB", inst.DiskGB),
			inst.Region,
			inst.PublicIP,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiCompute(),
		Headers: hdr("Name", "Distro", "Version", "vCPU", "Memory", "Disk", "Region", "IP", "Status", ""),
		Rows:    rows,
	}, true
}

// databases/relational/postgresql
func dbManagedPostgresLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListManagedPostgres(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbManagedPostgresLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		sslBadge := gr("Off")
		if inst.SSLEnforced {
			sslBadge = g("Enforced")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b("PostgreSQL "+inst.PGVersion),
			inst.Plan,
			fmt.Sprintf("%d vCPU", inst.VCPUs),
			fmt.Sprintf("%d GB RAM", inst.MemoryGB),
			fmt.Sprintf("%d GB", inst.StorageGB),
			fmt.Sprintf("%d", inst.Replicas),
			inst.BackupSchedule,
			sslBadge,
			statusBadgeDB(inst.DBStatus),
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDB(),
		Headers: hdr("Instance", "Engine", "Plan", "vCPU", "Memory", "Storage", "Replicas", "Backup", "SSL", "Status", ""),
		Rows:    rows,
	}, true
}

// storage/object-storage/s3-compatible
func dbProS3BucketsLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListProS3Buckets(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbProS3BucketsLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.BucketStatus {
		case "active":
			statusBadge = g("Active")
		case "locked":
			statusBadge = b("Locked")
		default:
			statusBadge = y(inst.BucketStatus)
		}
		cdnBadge := gr("Off")
		if inst.CDNEnabled {
			cdnBadge = g("On")
		}
		verBadge := gr("Off")
		if inst.Versioning {
			verBadge = b("On")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Region,
			b(inst.StorageClass),
			fmt.Sprintf("%.1f GB", inst.SizeGB),
			fmt.Sprintf("%d", inst.ObjectCount),
			inst.Encryption,
			inst.AccessPolicy,
			verBadge,
			cdnBadge,
			inst.LifecycleRule,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiStorage(),
		Headers: hdr("Bucket", "Region", "Class", "Size", "Objects", "Encryption", "Access", "Versioning", "CDN", "Lifecycle", "Status", ""),
		Rows:    rows,
	}, true
}

// apps/communication/transactional-email
func dbProTransactionalEmailLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListProTransactionalEmails(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbProTransactionalEmailLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.EmailStatus {
		case "active":
			statusBadge = g("Active")
		case "sandbox":
			statusBadge = b("Sandbox")
		case "pending_verification":
			statusBadge = y("Pending")
		default:
			statusBadge = gr(inst.EmailStatus)
		}
		dkimBadge := y(inst.DKIMStatus)
		if inst.DKIMStatus == "Verified" {
			dkimBadge = g("Verified")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Domain,
			inst.SenderAddress,
			dkimBadge,
			fmt.Sprintf("%d / %d", inst.DailySent, inst.DailyLimit),
			fmt.Sprintf("%.1f%%", inst.DeliveryRate),
			fmt.Sprintf("%.1f%%", inst.BounceRate),
			fmt.Sprintf("%.1f%%", inst.OpenRate),
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiApps(),
		Headers: hdr("Service", "Domain", "Sender", "DKIM", "Sent / Limit", "Delivery", "Bounce", "Open rate", "Status", ""),
		Rows:    rows,
	}, true
}

// networking/cdn/pop-management
func dbProCDNPoPLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListProCDNPoPs(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbProCDNPoPLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.PoPStatus {
		case "active":
			statusBadge = g("Active")
		case "maintenance":
			statusBadge = y("Maintenance")
		default:
			statusBadge = gr(inst.PoPStatus)
		}
		http3Badge := gr("No")
		if inst.HTTP3 {
			http3Badge = g("Yes")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			inst.Location,
			inst.Provider,
			fmt.Sprintf("%.1f%%", inst.CacheHitRate),
			fmt.Sprintf("%.1f TB", inst.BandwidthTB),
			fmt.Sprintf("%.0fM", inst.RequestsM),
			fmt.Sprintf("%d ms", inst.AvgLatencyMs),
			inst.Capacity,
			b("TLS "+inst.TLSVersion),
			http3Badge,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiNet(),
		Headers: hdr("PoP", "Location", "Provider", "Cache hit", "Bandwidth", "Requests", "Latency", "Capacity", "TLS", "HTTP/3", "Status", ""),
		Rows:    rows,
	}, true
}

// security/certificates/ssl-tls
func dbProSSLCertsLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListProSSLCerts(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbProSSLCertsLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var statusBadge string
		switch inst.CertStatus {
		case "active":
			statusBadge = g("Active")
		case "expiring_soon":
			statusBadge = y("Expiring soon")
		case "expired":
			statusBadge = e("Expired")
		case "pending_validation":
			statusBadge = b("Pending")
		default:
			statusBadge = gr(inst.CertStatus)
		}
		renewBadge := gr("Manual")
		if inst.AutoRenew {
			renewBadge = g("Auto")
		}
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			b(inst.CertType),
			inst.Issuer,
			inst.IssuedAt,
			inst.ExpiresAt,
			b(inst.Protocol),
			fmt.Sprintf("%s %d", inst.KeyAlgorithm, inst.KeySize),
			renewBadge,
			statusBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiSec(),
		Headers: hdr("Domain", "Type", "Issuer", "Issued", "Expires", "Protocol", "Key", "Auto-renew", "Status", ""),
		Rows:    rows,
	}, true
}

// devops/monitoring/metrics
func dbProMonitoringAlertsLeafFromDB(ctx context.Context, workspaceDBID int64) (WSLeafData, bool) {
	instances, err := db.ListProMonitoringAlerts(ctx, Pools.Admin, workspaceDBID)
	if err != nil {
		log.Printf("dbProMonitoringAlertsLeafFromDB: %v", err)
		return WSLeafData{}, false
	}
	if len(instances) == 0 {
		return WSLeafData{}, false
	}
	rows := make([][]template.HTML, len(instances))
	for i, inst := range instances {
		var stateBadge string
		switch inst.State {
		case "ok":
			stateBadge = g("OK")
		case "alerting":
			stateBadge = e("Alerting")
		case "pending":
			stateBadge = y("Pending")
		default:
			stateBadge = gr(inst.State)
		}
		var sevBadge string
		switch inst.Severity {
		case "critical":
			sevBadge = e("Critical")
		case "warning":
			sevBadge = y("Warning")
		default:
			sevBadge = b(inst.Severity)
		}
		silencedStr := ""
		if inst.Silenced {
			silencedStr = gr("Silenced")
		}
		condStr := fmt.Sprintf("%s %s %.0f", inst.Metric, inst.Condition, inst.Threshold)
		rows[i] = r(
			ni(inst.Name, fmt.Sprintf("res-%d", inst.ResourceID)),
			sevBadge,
			m(condStr),
			inst.Window,
			inst.EvalInterval,
			inst.NotifyChannels,
			inst.LastTriggered,
			silencedStr,
			stateBadge,
			`<button class="lf-btn lf-btn--ghost">Manage →</button>`,
		)
	}
	return WSLeafData{
		KPIs:    kpiDevOps(),
		Headers: hdr("Alert", "Severity", "Condition", "Window", "Eval", "Notify", "Last triggered", "Silenced", "State", ""),
		Rows:    rows,
	}, true
}
