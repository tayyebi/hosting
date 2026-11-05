package handlers

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"platform/db"
)

// ── Nav data structures ───────────────────────────────────────

type WSNavLeaf struct {
	ID    string
	Label string
	Path  string
}

type WSNavSection struct {
	ID    string
	Label string
	Items []WSNavLeaf
}

type WSNavCategory struct {
	ID       string
	Label    string
	Icon     string
	Sections []WSNavSection
}

type WorkspaceShellData struct {
	User           User
	Tenant         Tenant
	Client         Client
	Workspace      Workspace
	Workspaces     []Workspace
	Nav            []WSNavCategory
	ActiveNav      string
	ActiveCat      string
	ActiveSec      string
	PageTitle      string
	PageSub        string
	Leaf           WSLeafData
	ResourceCount  int
	ActiveCount    int
	MemberCount    int
}

// ── Nav tree ─────────────────────────────────────────────────

func wsBase(slug, wsID string) string {
	return "/t/" + slug + "/w/" + wsID
}

func wsNavCategories(base string) []WSNavCategory {
	leaf := func(id, label, path string) WSNavLeaf {
		return WSNavLeaf{ID: id, Label: label, Path: base + "/" + path + "/"}
	}

	return []WSNavCategory{
		{
			ID: "compute", Label: "Compute", Icon: "compute",
			Sections: []WSNavSection{
				{ID: "compute-vms", Label: "Virtual Machines", Items: []WSNavLeaf{
					leaf("compute/vms/general", "General purpose", "compute/vms/general"),
					leaf("compute/vms/gpu", "GPU / accelerated", "compute/vms/gpu"),
					leaf("compute/vms/spot", "Spot / preemptible", "compute/vms/spot"),
				}},
				{ID: "compute-containers", Label: "Containers", Items: []WSNavLeaf{
					leaf("compute/containers/kubernetes", "Managed Kubernetes", "compute/containers/kubernetes"),
					leaf("compute/containers/registry", "Container registry", "compute/containers/registry"),
					leaf("compute/containers/instances", "Container instances", "compute/containers/instances"),
				}},
				{ID: "compute-hosting", Label: "Shared Hosting", Items: []WSNavLeaf{
					leaf("compute/hosting/web", "Web app hosting", "compute/hosting/web"),
					leaf("compute/hosting/static", "Static site hosting", "compute/hosting/static"),
				}},
				{ID: "compute-serverless", Label: "Serverless Compute", Items: []WSNavLeaf{
					leaf("compute/serverless/functions", "Functions as a service", "compute/serverless/functions"),
					leaf("compute/serverless/edge", "Edge functions", "compute/serverless/edge"),
				}},
				{ID: "compute-batch", Label: "Batch Processing", Items: []WSNavLeaf{
					leaf("compute/batch/jobs", "Job scheduling", "compute/batch/jobs"),
					leaf("compute/batch/hpc", "HPC / grid computing", "compute/batch/hpc"),
				}},
			},
		},
		{
			ID: "storage", Label: "Storage", Icon: "storage",
			Sections: []WSNavSection{
				{ID: "storage-object", Label: "Object Storage", Items: []WSNavLeaf{
					leaf("storage/object/standard", "Standard tier", "storage/object/standard"),
					leaf("storage/object/archive", "Archive / cold tier", "storage/object/archive"),
					leaf("storage/object/datalake", "Data lake", "storage/object/datalake"),
				}},
				{ID: "storage-block", Label: "Block Storage", Items: []WSNavLeaf{
					leaf("storage/block/ssd", "SSD – high performance", "storage/block/ssd"),
					leaf("storage/block/hdd", "HDD – standard", "storage/block/hdd"),
				}},
				{ID: "storage-file", Label: "File Storage", Items: []WSNavLeaf{
					leaf("storage/file/nfs", "NFS / SMB shares", "storage/file/nfs"),
					leaf("storage/file/distributed", "Distributed file systems", "storage/file/distributed"),
				}},
				{ID: "storage-backup", Label: "Backup & Recovery", Items: []WSNavLeaf{
					leaf("storage/backup/snapshots", "Snapshot management", "storage/backup/snapshots"),
					leaf("storage/backup/dr", "Disaster recovery", "storage/backup/dr"),
					leaf("storage/backup/archive", "Long-term archiving", "storage/backup/archive"),
				}},
			},
		},
		{
			ID: "databases", Label: "Databases", Icon: "databases",
			Sections: []WSNavSection{
				{ID: "db-relational", Label: "Relational Databases", Items: []WSNavLeaf{
					leaf("databases/relational/mysql-pg", "MySQL / PostgreSQL", "databases/relational/mysql-pg"),
					leaf("databases/relational/sql-server", "SQL Server", "databases/relational/sql-server"),
					leaf("databases/relational/cloud-native", "Cloud-native", "databases/relational/cloud-native"),
				}},
				{ID: "db-nosql", Label: "NoSQL Databases", Items: []WSNavLeaf{
					leaf("databases/nosql/document", "Document", "databases/nosql/document"),
					leaf("databases/nosql/keyvalue", "Key-value", "databases/nosql/keyvalue"),
					leaf("databases/nosql/wide-column", "Wide-column", "databases/nosql/wide-column"),
					leaf("databases/nosql/graph", "Graph", "databases/nosql/graph"),
				}},
				{ID: "db-memory", Label: "In-Memory Databases", Items: []WSNavLeaf{
					leaf("databases/memory/cache", "Cache", "databases/memory/cache"),
					leaf("databases/memory/session", "Session store", "databases/memory/session"),
				}},
				{ID: "db-analytics", Label: "Analytics Databases", Items: []WSNavLeaf{
					leaf("databases/analytics/warehouse", "Data warehouse", "databases/analytics/warehouse"),
					leaf("databases/analytics/timeseries", "Time-series", "databases/analytics/timeseries"),
					leaf("databases/analytics/search", "Search & full-text", "databases/analytics/search"),
				}},
			},
		},
		{
			ID: "networking", Label: "Networking", Icon: "networking",
			Sections: []WSNavSection{
				{ID: "net-vnet", Label: "Virtual Networks", Items: []WSNavLeaf{
					leaf("networking/vnet/vpc", "VPC / subnets", "networking/vnet/vpc"),
					leaf("networking/vnet/peering", "Peering & transit", "networking/vnet/peering"),
					leaf("networking/vnet/nsg", "Network security groups", "networking/vnet/nsg"),
				}},
				{ID: "net-lb", Label: "Load Balancing", Items: []WSNavLeaf{
					leaf("networking/lb/app", "Application load balancer", "networking/lb/app"),
					leaf("networking/lb/network", "Network load balancer", "networking/lb/network"),
					leaf("networking/lb/global", "Global load balancing", "networking/lb/global"),
				}},
				{ID: "net-dns", Label: "DNS & Domains", Items: []WSNavLeaf{
					leaf("networking/dns/manage", "DNS management", "networking/dns/manage"),
					leaf("networking/dns/domains", "Domain registration", "networking/dns/domains"),
				}},
				{ID: "net-cdn", Label: "CDN", Items: []WSNavLeaf{
					leaf("networking/cdn/static", "Static asset delivery", "networking/cdn/static"),
					leaf("networking/cdn/video", "Video streaming", "networking/cdn/video"),
					leaf("networking/cdn/edge", "Edge caching", "networking/cdn/edge"),
				}},
				{ID: "net-private", Label: "Private Connectivity", Items: []WSNavLeaf{
					leaf("networking/private/vpn", "VPN gateway", "networking/private/vpn"),
					leaf("networking/private/direct", "Direct connect", "networking/private/direct"),
					leaf("networking/private/sdwan", "SD-WAN", "networking/private/sdwan"),
				}},
			},
		},
		{
			ID: "ai", Label: "AI & Data Intelligence", Icon: "ai",
			Sections: []WSNavSection{
				{ID: "ai-training", Label: "Model Training", Items: []WSNavLeaf{
					leaf("ai/training/distributed", "Distributed training", "ai/training/distributed"),
					leaf("ai/training/automl", "AutoML", "ai/training/automl"),
					leaf("ai/training/finetuning", "Fine-tuning", "ai/training/finetuning"),
				}},
				{ID: "ai-hosting", Label: "Model Hosting", Items: []WSNavLeaf{
					leaf("ai/hosting/inference", "Inference endpoints", "ai/hosting/inference"),
					leaf("ai/hosting/batch", "Batch inference", "ai/hosting/batch"),
					leaf("ai/hosting/edge", "Edge deployment", "ai/hosting/edge"),
				}},
				{ID: "ai-apis", Label: "Prebuilt AI APIs", Items: []WSNavLeaf{
					leaf("ai/apis/vision", "Vision", "ai/apis/vision"),
					leaf("ai/apis/language", "Language / NLP", "ai/apis/language"),
					leaf("ai/apis/speech", "Speech", "ai/apis/speech"),
					leaf("ai/apis/recommendation", "Recommendation", "ai/apis/recommendation"),
				}},
				{ID: "ai-data", Label: "Data Engineering", Items: []WSNavLeaf{
					leaf("ai/data/etl", "ETL / ELT pipelines", "ai/data/etl"),
					leaf("ai/data/streaming", "Streaming ingestion", "ai/data/streaming"),
					leaf("ai/data/catalog", "Data catalog", "ai/data/catalog"),
				}},
				{ID: "ai-mlops", Label: "ML Ops", Items: []WSNavLeaf{
					leaf("ai/mlops/experiments", "Experiment tracking", "ai/mlops/experiments"),
					leaf("ai/mlops/registry", "Model registry", "ai/mlops/registry"),
					leaf("ai/mlops/monitoring", "Monitoring & drift detection", "ai/mlops/monitoring"),
				}},
			},
		},
		{
			ID: "security", Label: "Security & Identity", Icon: "security",
			Sections: []WSNavSection{
				{ID: "sec-iam", Label: "Identity & Access Management", Items: []WSNavLeaf{
					leaf("security/iam/sso", "SSO / federation", "security/iam/sso"),
					leaf("security/iam/mfa", "Multi-factor authentication", "security/iam/mfa"),
					leaf("security/iam/rbac", "Role-based access control", "security/iam/rbac"),
				}},
				{ID: "sec-data", Label: "Data Protection", Items: []WSNavLeaf{
					leaf("security/data/enc-rest", "Encryption at rest", "security/data/enc-rest"),
					leaf("security/data/enc-transit", "Encryption in transit", "security/data/enc-transit"),
					leaf("security/data/kms", "Key management", "security/data/kms"),
				}},
				{ID: "sec-network", Label: "Network Security", Items: []WSNavLeaf{
					leaf("security/network/waf", "Web application firewall", "security/network/waf"),
					leaf("security/network/ddos", "DDoS protection", "security/network/ddos"),
					leaf("security/network/ids", "Intrusion detection", "security/network/ids"),
				}},
				{ID: "sec-compliance", Label: "Compliance & Audit", Items: []WSNavLeaf{
					leaf("security/compliance/policy", "Policy management", "security/compliance/policy"),
					leaf("security/compliance/audit", "Audit logging", "security/compliance/audit"),
					leaf("security/compliance/regulatory", "Regulatory frameworks", "security/compliance/regulatory"),
				}},
			},
		},
		{
			ID: "devops", Label: "Developer Tools & DevOps", Icon: "devops",
			Sections: []WSNavSection{
				{ID: "devops-cicd", Label: "CI/CD Pipelines", Items: []WSNavLeaf{
					leaf("devops/cicd/source", "Source control", "devops/cicd/source"),
					leaf("devops/cicd/build", "Build automation", "devops/cicd/build"),
					leaf("devops/cicd/deploy", "Deployment automation", "devops/cicd/deploy"),
				}},
				{ID: "devops-iac", Label: "Infrastructure as Code", Items: []WSNavLeaf{
					leaf("devops/iac/templates", "Templating", "devops/iac/templates"),
					leaf("devops/iac/config", "Config management", "devops/iac/config"),
				}},
				{ID: "devops-api", Label: "API Management", Items: []WSNavLeaf{
					leaf("devops/api/gateway", "API gateway", "devops/api/gateway"),
					leaf("devops/api/portal", "Developer portal", "devops/api/portal"),
					leaf("devops/api/analytics", "API analytics", "devops/api/analytics"),
				}},
				{ID: "devops-obs", Label: "Observability", Items: []WSNavLeaf{
					leaf("devops/obs/logging", "Logging", "devops/obs/logging"),
					leaf("devops/obs/metrics", "Metrics & dashboards", "devops/obs/metrics"),
					leaf("devops/obs/tracing", "Distributed tracing", "devops/obs/tracing"),
				}},
			},
		},
		{
			ID: "apps", Label: "Application Services", Icon: "apps",
			Sections: []WSNavSection{
				{ID: "apps-messaging", Label: "Messaging & Integration", Items: []WSNavLeaf{
					leaf("apps/messaging/queues", "Message queues", "apps/messaging/queues"),
					leaf("apps/messaging/streaming", "Event streaming", "apps/messaging/streaming"),
					leaf("apps/messaging/servicebus", "Service bus", "apps/messaging/servicebus"),
				}},
				{ID: "apps-workflow", Label: "Workflow & Orchestration", Items: []WSNavLeaf{
					leaf("apps/workflow/automation", "Workflow automation", "apps/workflow/automation"),
					leaf("apps/workflow/scheduler", "Scheduler / cron jobs", "apps/workflow/scheduler"),
				}},
				{ID: "apps-media", Label: "Media Services", Items: []WSNavLeaf{
					leaf("apps/media/encoding", "Video encoding", "apps/media/encoding"),
					leaf("apps/media/livestream", "Live streaming", "apps/media/livestream"),
					leaf("apps/media/storage", "Media storage", "apps/media/storage"),
				}},
			},
		},
		{
			ID: "multicloud", Label: "Multi-Cloud & Hybrid", Icon: "multicloud",
			Sections: []WSNavSection{
				{ID: "mc-hybrid", Label: "Hybrid Cloud", Items: []WSNavLeaf{
					leaf("multicloud/hybrid/onprem", "On-premises extension", "multicloud/hybrid/onprem"),
					leaf("multicloud/hybrid/control", "Unified control plane", "multicloud/hybrid/control"),
					leaf("multicloud/hybrid/edge", "Edge computing", "multicloud/hybrid/edge"),
				}},
				{ID: "mc-mgmt", Label: "Multi-Cloud Management", Items: []WSNavLeaf{
					leaf("multicloud/mgmt/cost", "Cost management", "multicloud/mgmt/cost"),
					leaf("multicloud/mgmt/policy", "Policy governance", "multicloud/mgmt/policy"),
					leaf("multicloud/mgmt/workload", "Workload portability", "multicloud/mgmt/workload"),
				}},
				{ID: "mc-migration", Label: "Migration Services", Items: []WSNavLeaf{
					leaf("multicloud/migration/assessment", "Assessment & discovery", "multicloud/migration/assessment"),
					leaf("multicloud/migration/data", "Data migration", "multicloud/migration/data"),
					leaf("multicloud/migration/app", "Application migration", "multicloud/migration/app"),
				}},
			},
		},
	}
}

// ── Nav helper ────────────────────────────────────────────────

func wsNavFindLeaf(nav []WSNavCategory, leafID string) (cat WSNavCategory, sec WSNavSection, leaf WSNavLeaf, ok bool) {
	for _, c := range nav {
		for _, s := range c.Sections {
			for _, l := range s.Items {
				if l.ID == leafID {
					return c, s, l, true
				}
			}
		}
	}
	return
}

// ── Middleware ────────────────────────────────────────────────

func workspaceMiddleware(next func(http.ResponseWriter, *http.Request, WorkspaceShellData)) http.HandlerFunc {
	return tenantMiddleware(func(w http.ResponseWriter, r *http.Request) {
		user, _ := getSession(r)
		tenant, _ := tenantFromContext(r)
		wsID := r.PathValue("workspace_id")

		client, ws, ok := clientStore.getWorkspace(tenant.Slug, wsID)
		if !ok {
			ServeError(w, r, http.StatusNotFound, "Workspace not found.")
			return
		}

		// Resolve workspace DB ID and live stats when Postgres is available.
		var resourceCount, activeCount, memberCount int
		if Pools != nil && tenant.DBID > 0 {
			ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
			defer cancel()
			// Try to find the workspace in DB by name-slug (lower-case, hyphenated).
			slug := strings.ToLower(strings.ReplaceAll(ws.Name, " ", "-"))
			if wsd, err := db.GetWorkspaceBySlug(ctx, Pools.Admin, tenant.DBID, slug); err == nil {
				ws.DBID = wsd.ID
				if stats, serr := db.GetWorkspaceStats(ctx, Pools.Admin, wsd.ID); serr == nil {
					resourceCount = stats.ResourceCount
					activeCount = stats.ActiveCount
					memberCount = stats.MemberCount
				} else {
					log.Printf("workspaceMiddleware stats: %v", serr)
				}
			}
		}

		base := wsBase(tenant.Slug, wsID)
		d := WorkspaceShellData{
			User:          user,
			Tenant:        tenant,
			Client:        client,
			Workspace:     ws,
			Workspaces:    client.Workspaces,
			Nav:           wsNavCategories(base),
			ResourceCount: resourceCount,
			ActiveCount:   activeCount,
			MemberCount:   memberCount,
		}
		next(w, r, d)
	})
}

// ── Handlers ─────────────────────────────────────────────────

func workspaceOverviewHandler(w http.ResponseWriter, r *http.Request, d WorkspaceShellData) {
	d.ActiveNav = ""
	d.ActiveCat = ""
	d.ActiveSec = ""
	d.PageTitle = "Overview"

	tmpl, err := template.ParseFiles(
		filepath.Join("templates", "workspace", "base.html"),
		filepath.Join("templates", "workspace", "shell.html"),
		filepath.Join("templates", "workspace", "overview_content.html"),
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "workspace-base", d); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func workspaceLeafHandler(w http.ResponseWriter, r *http.Request, d WorkspaceShellData) {
	rest := strings.TrimSuffix(strings.TrimPrefix(r.PathValue("rest"), "/"), "/")
	nav := wsNavCategories(wsBase(d.Tenant.Slug, d.Workspace.ID))
	cat, sec, leaf, ok := wsNavFindLeaf(nav, rest)
	if !ok {
		ServeError(w, r, http.StatusNotFound, "Page not found.")
		return
	}
	d.Nav = nav
	d.ActiveNav = leaf.ID
	d.ActiveCat = cat.ID
	d.ActiveSec = sec.ID
	d.PageTitle = leaf.Label
	d.PageSub = sec.Label + " · " + cat.Label

	// Try live DB data first; fall back to static registry.
	if dbLeaf, ok := leafFromDB(r.Context(), leaf.ID, d.Workspace.DBID); ok {
		d.Leaf = dbLeaf
	} else {
		d.Leaf = leafFor(leaf.ID)
	}

	tmpl, err := template.ParseFiles(
		filepath.Join("templates", "workspace", "base.html"),
		filepath.Join("templates", "workspace", "shell.html"),
		filepath.Join("templates", "workspace", "leaf_content.html"),
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tmpl.ExecuteTemplate(w, "workspace-base", d)
}

// ── Routes ────────────────────────────────────────────────────

func workspaceRootOrLeaf(w http.ResponseWriter, r *http.Request, d WorkspaceShellData) {
	rest := strings.TrimPrefix(r.PathValue("rest"), "/")
	rest = strings.TrimSuffix(rest, "/")
	if rest == "" {
		workspaceOverviewHandler(w, r, d)
		return
	}
	if rest == "marketplace" {
		workspaceMarketplaceHandler(w, r, d)
		return
	}
	if strings.HasPrefix(rest, "marketplace/") {
		workspaceMarketplaceHandler(w, r, d)
		return
	}
	if isProFeature(rest) {
		proFeatureHandler(w, r, d)
		return
	}
	workspaceLeafHandler(w, r, d)
}

func RegisterWorkspaceRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/w/{workspace_id}/{rest...}", workspaceMiddleware(workspaceRootOrLeaf))
}
