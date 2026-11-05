package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

// ── Pro feature data structures ──────────────────────────────

type ProStat struct {
	Label string
	Value string
	Unit  string
	Delta string
	Up    bool
	Good  bool
	Spark []int
}

type ProResource struct {
	Name        string
	Status      string // running, stopped, error, pending, healthy, active, inactive
	StatusLabel string
	Columns     []string // column values in order
	Region      string
}

type ProAction struct {
	Label   string
	Icon    string
	Href    string
	Primary bool
}

type ProActivity struct {
	Time   string
	Actor  string
	Event  string
	Target string
	Level  string // info, success, warn
}

type ProFeature struct {
	Slug         string // URL path fragment (no trailing slash)
	Title        string
	Subtitle     string
	Lead         string // one-line value statement
	Accent       string // hex base color (e.g. "#4f46e5")
	AccentSoft   string // soft bg rgba
	IconPath     string // SVG path d
	Stats        []ProStat
	Columns      []string // column headers for resource table
	Resources    []ProResource
	Actions      []ProAction
	Activities   []ProActivity
	DocsLink     string
	CallToAction string
}

type WorkspaceProPageData struct {
	WorkspaceShellData
	Feature ProFeature
}

// ── Feature registry ──────────────────────────────────────────

func spark(vals ...int) []int { return vals }

func proFeatureRegistry() map[string]ProFeature {
	return map[string]ProFeature{
		"apps/app-hosting/static-sites": {
			Slug:     "apps/app-hosting/static-sites",
			Title:    "Websites & Apps",
			Subtitle: "Git-connected deployments · Edge CDN · Preview branches",
			Lead:     "Push to Git, deploy everywhere. Each branch gets its own preview URL with instant rollbacks, atomic deploys to 180+ PoPs, and built-in Core Web Vitals monitoring.",
			Accent:   "#4f46e5", AccentSoft: "rgba(79,70,229,0.1)",
			IconPath: "M2 4a2 2 0 012-2h12a2 2 0 012 2v8a2 2 0 01-2 2H4a2 2 0 01-2-2V4zm0 2h16M6 2v2",
			Stats: []ProStat{
				{Label: "Active Sites", Value: "0"},
				{Label: "Deploys (7d)", Value: "0"},
				{Label: "LCP p75", Value: "0", Unit: "s"},
				{Label: "Avg build", Value: "0", Unit: "s"},
			},
			Columns:   []string{"Site", "Framework", "Branch / Commit", "Visits (24h)", "Build", "Region"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "New Site", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Connect Git Repo", Icon: "git", Href: "#"},
				{Label: "Import from URL", Icon: "link", Href: "#"},
				{Label: "Environment Variables", Icon: "key", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Deploy a new site",
		},
		"compute/virtual-machines/linux-vms": {
			Slug:     "compute/virtual-machines/linux-vms",
			Title:    "Cloud Servers",
			Subtitle: "Bare-metal performance · Live migration · Per-second billing",
			Lead:     "Launch VMs in under 30 seconds across 12 regions. Live-migrate without downtime, attach NVMe volumes on the fly, and schedule snapshots with configurable retention.",
			Accent:   "#15803d", AccentSoft: "rgba(22,163,74,0.1)",
			IconPath: "M2 4h16v12H2V4zm2 2v8h12V6H4zm3 3h6M7 11h4",
			Stats: []ProStat{
				{Label: "Running VMs", Value: "0"},
				{Label: "vCPU allocated", Value: "0", Unit: "/128"},
				{Label: "Volumes attached", Value: "0", Unit: "TB"},
				{Label: "Cost efficiency", Value: "0", Unit: "%"},
			},
			Columns:   []string{"Server", "Image", "Plan", "vCPU / Mem / Disk", "Region", "Uptime"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Launch VM", Primary: true, Icon: "plus", Href: "#"},
				{Label: "SSH Keys", Icon: "key", Href: "#"},
				{Label: "Snapshots", Icon: "camera", Href: "#"},
				{Label: "Firewall Rules", Icon: "shield", Href: "#"},
				{Label: "Resize Plan", Icon: "scale", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Launch your first server",
		},
		"databases/relational/postgresql": {
			Slug:     "databases/relational/postgresql",
			Title:    "Managed Databases",
			Subtitle: "Automatic failover · Point-in-time recovery · Connection pooling",
			Lead:     "Production-grade PostgreSQL, MySQL, and Redis with automatic daily backups (30-day retention), transparent failover under 15 seconds, built-in connection pooling, and one-click read replica promotion.",
			Accent:   "#1d4ed8", AccentSoft: "rgba(59,130,246,0.1)",
			IconPath: "M10 2C6 2 3 3.2 3 4.5v11C3 16.8 6 18 10 18s7-1.2 7-2.5v-11C17 3.2 14 2 10 2zM3 4.5v11M17 4.5v11",
			Stats: []ProStat{
				{Label: "Clusters", Value: "0"},
				{Label: "Connections", Value: "0", Unit: "/200"},
				{Label: "QPS (avg)", Value: "0", Unit: "K"},
				{Label: "Replica lag", Value: "0", Unit: "ms"},
			},
			Columns:   []string{"Cluster", "Engine", "Plan", "Replicas", "Storage", "Region"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Create Cluster", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Restore Backup", Icon: "backup", Href: "#"},
				{Label: "Connection Pooler", Icon: "link", Href: "#"},
				{Label: "Query Console", Icon: "terminal", Href: "#"},
				{Label: "Slow Query Log", Icon: "logs", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Create a database cluster",
		},
		"storage/object-storage/s3-compatible": {
			Slug:     "storage/object-storage/s3-compatible",
			Title:    "Storage & Backups",
			Subtitle: "S3-compatible API · Lifecycle tiering · Immutable object lock",
			Lead:     "Drop-in S3-compatible object storage at 80% less cost. Automatic hot → warm → archive tiering, cross-region replication with conflict resolution, and WORM compliance locks for regulatory retention.",
			Accent:   "#b45309", AccentSoft: "rgba(245,158,11,0.1)",
			IconPath: "M2 5l8-3 8 3-8 3-8-3zm0 0v10l8 3 8-3V5",
			Stats: []ProStat{
				{Label: "Buckets", Value: "0"},
				{Label: "PUT/GET (24h)", Value: "0", Unit: "M"},
				{Label: "Storage", Value: "0", Unit: "TB"},
				{Label: "Saved (vs S3)", Value: "0", Unit: "/mo"},
			},
			Columns:   []string{"Bucket", "Region", "Objects", "Size", "Tier", "Repl / Lock"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "New Bucket", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Lifecycle Rules", Icon: "clock", Href: "#"},
				{Label: "Access Keys (HMAC)", Icon: "key", Href: "#"},
				{Label: "Replication Rules", Icon: "sync", Href: "#"},
				{Label: "Storage Calculator", Icon: "calculator", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Create your first bucket",
		},
		"apps/communication/transactional-email": {
			Slug:     "apps/communication/transactional-email",
			Title:    "Email & Messaging",
			Subtitle: "SMTP & API relay · DKIM/SPF/DMARC · Real-time webhooks",
			Lead:     "Send transactional email via SMTP or REST API with automatic DKIM signing, dedicated IP warm-up, per-recipient engagement tracking, and real-time webhooks for delivered / opened / clicked / bounced events.",
			Accent:   "#9d174d", AccentSoft: "rgba(236,72,153,0.1)",
			IconPath: "M2 5a1 1 0 011-1h14a1 1 0 011 1v10a1 1 0 01-1 1H3a1 1 0 01-1-1V5zm0 0l8 6 8-6",
			Stats: []ProStat{
				{Label: "Sent (24h)", Value: "0", Unit: "K"},
				{Label: "Inbox rate", Value: "0", Unit: "%"},
				{Label: "Open rate", Value: "0", Unit: "%"},
				{Label: "Bounce rate", Value: "0", Unit: "%"},
			},
			Columns:   []string{"Sending Domain", "Senders", "Volume (mo)", "Reputation", "Auth", "Dedicated IP"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Add Domain", Primary: true, Icon: "plus", Href: "#"},
				{Label: "API Keys & SMTP", Icon: "key", Href: "#"},
				{Label: "Email Templates", Icon: "template", Href: "#"},
				{Label: "Suppression List", Icon: "block", Href: "#"},
				{Label: "Webhook Events", Icon: "link", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Verify your sending domain",
		},
		"networking/cdn/pop-management": {
			Slug:     "networking/cdn/pop-management",
			Title:    "Content Delivery",
			Subtitle: "180+ PoPs · Instant purge · Image optimization · Edge compute",
			Lead:     "Serve content from 180+ edge locations with sub-20ms global latency. On-the-fly image resizing (WebP/AVIF), Brotli compression, instant cache invalidation (<500ms worldwide), and programmable edge rules via JavaScript workers.",
			Accent:   "#0e7490", AccentSoft: "rgba(6,182,212,0.1)",
			IconPath: "M10 2a8 8 0 100 16 8 8 0 000-16zm0 0v16M2 10h16M4 5c2 2 4 3 6 3s4-1 6-3M4 15c2-2 4-3 6-3s4 1 6 3",
			Stats: []ProStat{
				{Label: "Cache hit ratio", Value: "0", Unit: "%"},
				{Label: "Requests (24h)", Value: "0", Unit: "M"},
				{Label: "P50 TTFB", Value: "0", Unit: "ms"},
				{Label: "Bandwidth saved", Value: "0", Unit: "%"},
			},
			Columns:   []string{"Distribution", "Origin", "Hit Rate", "Reqs (24h)", "Transfer", "TLS / HTTP"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "New Distribution", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Purge Cache", Icon: "refresh", Href: "#"},
				{Label: "Edge Rules (JS)", Icon: "rules", Href: "#"},
				{Label: "Image Optimization", Icon: "image", Href: "#"},
				{Label: "Origin Shield", Icon: "shield", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Create a distribution",
		},
		"security/certificates/ssl-tls": {
			Slug:     "security/certificates/ssl-tls",
			Title:    "Security & SSL",
			Subtitle: "Auto-renewing certs · Managed WAF · Secrets vault · DDoS shield",
			Lead:     "Zero-touch TLS: certificates are issued, deployed to edge, and renewed automatically. Layer-7 WAF with OWASP Core Rule Set, real-time DDoS mitigation up to 10 Tbps, and a KMS-backed secrets vault with audit trail.",
			Accent:   "#059669", AccentSoft: "rgba(16,185,129,0.1)",
			IconPath: "M10 1L3 4v5c0 4 3 7.5 7 9 4-1.5 7-5 7-9V4l-7-3zm0 6v4m0 2v.5",
			Stats: []ProStat{
				{Label: "Active certs", Value: "0"},
				{Label: "WAF blocks (24h)", Value: "0"},
				{Label: "DDoS mitigated", Value: "0", Unit: "Gbps"},
				{Label: "Vault secrets", Value: "0"},
			},
			Columns:   []string{"Certificate", "Type / Grade", "Issuer", "Expires", "Auto-renew", "Endpoints"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Issue Certificate", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Upload Certificate", Icon: "upload", Href: "#"},
				{Label: "WAF Rules (OWASP)", Icon: "shield", Href: "#"},
				{Label: "Secrets Vault", Icon: "key", Href: "#"},
				{Label: "DDoS Settings", Icon: "shield", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Issue your first certificate",
		},
		"devops/monitoring/metrics": {
			Slug:     "devops/monitoring/metrics",
			Title:    "Monitoring & Alerts",
			Subtitle: "OpenTelemetry-native · PromQL queries · PagerDuty / Slack routing",
			Lead:     "Unified observability: ingest metrics via OpenTelemetry or Prometheus remote-write, query with PromQL, correlate with logs and traces in one timeline, and route alerts through escalation policies to PagerDuty, Slack, or webhooks.",
			Accent:   "#7c3aed", AccentSoft: "rgba(168,85,247,0.1)",
			IconPath: "M2 14l4-4 3 3 4-6 5 7",
			Stats: []ProStat{
				{Label: "Series ingested", Value: "0", Unit: "M/min"},
				{Label: "Active alerts", Value: "0"},
				{Label: "Log volume (24h)", Value: "0", Unit: "GB"},
				{Label: "MTTR (7d avg)", Value: "0", Unit: "min"},
			},
			Columns:   []string{"Alert", "Severity", "PromQL / Condition", "Firing", "Scope", "Escalation"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "New Alert Rule", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Dashboards", Icon: "dashboard", Href: "#"},
				{Label: "Log Explorer", Icon: "logs", Href: "#"},
				{Label: "Trace Waterfall", Icon: "trace", Href: "#"},
				{Label: "Escalation Policies", Icon: "route", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Build your first dashboard",
		},
		"ai/inference/llm-endpoints": {
			Slug:     "ai/inference/llm-endpoints",
			Title:    "AI & Automation",
			Subtitle: "vLLM-powered inference · GGUF & SafeTensors · Auto-scaling to zero",
			Lead:     "Deploy open-weight LLMs on dedicated GPUs with vLLM continuous batching, automatic KV-cache offloading, quantized GGUF support, and scale-to-zero when idle. Includes managed Qdrant vector DB and a visual prompt playground.",
			Accent:   "#c2410c", AccentSoft: "rgba(249,115,22,0.1)",
			IconPath: "M10 2a8 8 0 018 8c0 3-2 6-4 7l-1 2h-6l-1-2c-2-1-4-4-4-7a8 8 0 018-8zm-2 7a1 1 0 100 2 1 1 0 000-2zm4 0a1 1 0 100 2 1 1 0 000-2z",
			Stats: []ProStat{
				{Label: "Endpoints", Value: "0"},
				{Label: "Tokens (24h)", Value: "0", Unit: "M"},
				{Label: "P95 TTFT", Value: "0", Unit: "ms"},
				{Label: "GPU utilization", Value: "0", Unit: "%"},
			},
			Columns:   []string{"Endpoint", "Model / Quant", "GPU", "Throughput", "P95 TTFT / TPS", "Status"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Deploy Model", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Vector DB (Qdrant)", Icon: "vector", Href: "#"},
				{Label: "Prompt Playground", Icon: "prompt", Href: "#"},
				{Label: "Eval Suites", Icon: "check", Href: "#"},
				{Label: "Fine-tune (LoRA)", Icon: "tune", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Deploy your first model",
		},
		"apps/serverless/functions": {
			Slug:     "apps/serverless/functions",
			Title:    "Serverless Functions",
			Subtitle: "V8 isolate cold starts · HTTP / cron / queue triggers · Pay-per-invocation",
			Lead:     "Deploy functions in Node, Python, Go, or Rust with V8-isolate cold starts under 10ms. HTTP routes, cron schedules, and queue consumers as triggers — billed per millisecond of execution with no idle cost.",
			Accent:   "#4338ca", AccentSoft: "rgba(99,102,241,0.1)",
			IconPath: "M5 2l5 7-3 1 5 6-2-7 3-1-5-6",
			Stats: []ProStat{
				{Label: "Functions", Value: "0"},
				{Label: "Invocations (24h)", Value: "0", Unit: "M"},
				{Label: "P99 cold start", Value: "0", Unit: "ms"},
				{Label: "Error budget", Value: "0", Unit: "% OK"},
			},
			Columns:   []string{"Function", "Runtime", "Trigger", "Invocations (24h)", "P50 / P99", "Errors (24h)"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "New Function", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Cron Schedules", Icon: "clock", Href: "#"},
				{Label: "Environment & Secrets", Icon: "key", Href: "#"},
				{Label: "Deploy via CLI", Icon: "terminal", Href: "#"},
				{Label: "Logs & Tracing", Icon: "logs", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Write your first function",
		},
		"networking/dns/zones": {
			Slug:     "networking/dns/zones",
			Title:    "Domains & DNS",
			Subtitle: "Registrar · DNS · Domains",
			Lead:     "Register, transfer, and manage domains with DNSSEC, ANAME flattening, geolocation routing, and sub-millisecond resolution.",
			Accent:   "#374151", AccentSoft: "rgba(107,114,128,0.1)",
			IconPath: "M10 2a8 8 0 00-8 8 8 8 0 008 8 8 8 0 008-8 8 8 0 00-8-8zm0 0c-2.5 2-4 5-4 8s1.5 6 4 8c2.5-2 4-5 4-8s-1.5-6-4-8zM2 10h16",
			Stats: []ProStat{
				{Label: "Domains", Value: "0"},
				{Label: "DNS zones", Value: "0"},
				{Label: "Queries (24h)", Value: "0", Unit: "M"},
				{Label: "Resolution p50", Value: "0", Unit: "ms"},
			},
			Columns:   []string{"Domain", "Registrar", "DNS", "Records", "DNSSEC", "Expires"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Register Domain", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Transfer Domain", Icon: "arrow", Href: "#"},
				{Label: "Import Zone", Icon: "upload", Href: "#"},
				{Label: "Auto-renew Settings", Icon: "refresh", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Register a domain",
		},
		"compute/kubernetes/managed-k8s": {
			Slug:     "compute/kubernetes/managed-k8s",
			Title:    "Managed Kubernetes",
			Subtitle: "Multi-cluster orchestration · Auto-scaling node pools · GitOps-ready",
			Lead:     "Run production-grade Kubernetes without managing control planes. Auto-scaling node pools, one-click version upgrades, integrated GitOps workflows, and built-in observability across every cluster.",
			Accent:   "#2563eb", AccentSoft: "rgba(37,99,235,0.1)",
			IconPath: "M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5",
			Stats: []ProStat{
				{Label: "Clusters", Value: "0"},
				{Label: "Pods running", Value: "0"},
				{Label: "Node utilization", Value: "0", Unit: "%"},
				{Label: "Deploy frequency", Value: "0", Unit: "/d"},
			},
			Columns:   []string{"Cluster", "Version", "Node Pools", "Pods", "Region", "API Endpoint"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Create Cluster", Primary: true, Icon: "plus", Href: "#"},
				{Label: "kubectl Access", Icon: "terminal", Href: "#"},
				{Label: "Helm Charts", Icon: "package", Href: "#"},
				{Label: "Upgrade Version", Icon: "refresh", Href: "#"},
				{Label: "Node Pool", Icon: "server", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Create a cluster",
		},
		"storage/volumes/block-storage": {
			Slug:     "storage/volumes/block-storage",
			Title:    "Block Storage",
			Subtitle: "NVMe-backed volumes · Instant snapshots · AES-256 encryption at rest",
			Lead:     "Attach high-performance NVMe block volumes to any compute instance. Provision up to 100k IOPS per volume, schedule automatic snapshots, and encrypt everything at rest with customer-managed keys.",
			Accent:   "#7c3aed", AccentSoft: "rgba(124,92,255,0.1)",
			IconPath: "M4 4h16v4H4V4zm0 6h16v4H4v-4zm0 6h16v4H4v-4zM7 6h2M7 12h2M7 18h2",
			Stats: []ProStat{
				{Label: "Volumes", Value: "0"},
				{Label: "Total IOPS", Value: "0"},
				{Label: "Throughput", Value: "0", Unit: "GB/s"},
				{Label: "Snapshot count", Value: "0"},
			},
			Columns:   []string{"Volume", "Type", "Size", "IOPS", "Attached To", "Status"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Create Volume", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Take Snapshot", Icon: "camera", Href: "#"},
				{Label: "Resize Volume", Icon: "scale", Href: "#"},
				{Label: "Encryption Keys", Icon: "key", Href: "#"},
				{Label: "Restore Snapshot", Icon: "refresh", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Create a volume",
		},
		"databases/nosql/document-db": {
			Slug:     "databases/nosql/document-db",
			Title:    "Document Database",
			Subtitle: "MongoDB-compatible API · Auto-sharding · Real-time change streams",
			Lead:     "Fully managed document database with MongoDB wire-protocol compatibility. Automatic sharding distributes data across zones, change streams power real-time event pipelines, and point-in-time recovery protects every write.",
			Accent:   "#047857", AccentSoft: "rgba(16,185,129,0.1)",
			IconPath: "M4 6c0-1.1 3.6-2 8-2s8 .9 8 2M4 6v4c0 1.1 3.6 2 8 2s8-.9 8-2V6M4 10v4c0 1.1 3.6 2 8 2s8-.9 8-2v-4M4 14v4c0 1.1 3.6 2 8 2s8-.9 8-2v-4",
			Stats: []ProStat{
				{Label: "Clusters", Value: "0"},
				{Label: "Collections", Value: "0"},
				{Label: "Documents", Value: "0"},
				{Label: "Ops/sec", Value: "0"},
			},
			Columns:   []string{"Cluster", "Tier", "Storage", "Connections", "Region", "Replica Set"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Create Cluster", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Connection Strings", Icon: "link", Href: "#"},
				{Label: "Data Explorer", Icon: "search", Href: "#"},
				{Label: "Backup & Restore", Icon: "download", Href: "#"},
				{Label: "Index Advisor", Icon: "chart", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Launch a cluster",
		},
		"networking/load-balancer/application-lb": {
			Slug:     "networking/load-balancer/application-lb",
			Title:    "Application Load Balancer",
			Subtitle: "Layer-7 routing · Automatic TLS termination · Path-based rules · Active health checks",
			Lead:     "Distribute HTTP/HTTPS traffic across backends with content-based routing, automatic certificate management, WebSocket support, and real-time health checks that remove unhealthy targets in under 10 seconds.",
			Accent:   "#0284c7", AccentSoft: "rgba(14,165,233,0.1)",
			IconPath: "M12 2v6m0 8v6M2 12h6m8 0h6M8 8l-4-4M16 8l4-4M8 16l-4 4M16 16l4 4",
			Stats: []ProStat{
				{Label: "Load Balancers", Value: "0"},
				{Label: "Active connections", Value: "0"},
				{Label: "RPS", Value: "0"},
				{Label: "Healthy targets", Value: "0", Unit: "%"},
			},
			Columns:   []string{"Load Balancer", "Protocol", "Targets", "RPS", "Cert Status", "Region"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Create Load Balancer", Primary: true, Icon: "plus", Href: "#"},
				{Label: "TLS Certificates", Icon: "shield", Href: "#"},
				{Label: "Routing Rules", Icon: "git", Href: "#"},
				{Label: "Health Checks", Icon: "heart", Href: "#"},
				{Label: "Access Logs", Icon: "file", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Create a load balancer",
		},
		"ai/vector-search/embeddings": {
			Slug:     "ai/vector-search/embeddings",
			Title:    "Vector Search",
			Subtitle: "Purpose-built vector DB · Semantic search · RAG pipeline integration",
			Lead:     "Index and query billions of high-dimensional vectors with sub-millisecond latency. Build semantic search, recommendation engines, and retrieval-augmented generation pipelines with native embedding model support and hybrid keyword+vector queries.",
			Accent:   "#db2777", AccentSoft: "rgba(236,72,153,0.1)",
			IconPath: "M12 2a4 4 0 014 4c0 1.5-.8 2.8-2 3.5M12 2a4 4 0 00-4 4c0 1.5.8 2.8 2 3.5M6 14l6 6 6-6M3 18h18M8 10h8",
			Stats: []ProStat{
				{Label: "Collections", Value: "0"},
				{Label: "Vectors indexed", Value: "0"},
				{Label: "Queries/sec", Value: "0"},
				{Label: "Recall@10", Value: "0", Unit: "%"},
			},
			Columns:   []string{"Collection", "Dimensions", "Vectors", "Index Type", "Queries/sec", "Status"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Create Collection", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Query Playground", Icon: "search", Href: "#"},
				{Label: "Embedding Models", Icon: "cpu", Href: "#"},
				{Label: "Import Vectors", Icon: "upload", Href: "#"},
				{Label: "RAG Templates", Icon: "template", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Create a collection",
		},
		"security/iam/identity-platform": {
			Slug:     "security/iam/identity-platform",
			Title:    "Identity Platform",
			Subtitle: "Enterprise SSO · SCIM provisioning · Adaptive MFA · Fine-grained RBAC",
			Lead:     "Centralize authentication and authorization across your organization. Connect enterprise identity providers via SAML/OIDC, automate user lifecycle with SCIM, enforce adaptive MFA policies, and define fine-grained roles with attribute-based access control.",
			Accent:   "#0f766e", AccentSoft: "rgba(20,184,166,0.1)",
			IconPath: "M12 2a5 5 0 015 5c0 2.76-2.24 5-5 5s-5-2.24-5-5a5 5 0 015-5zM2 20c0-3.31 4.03-6 9-6h2c4.97 0 9 2.69 9 6M16 4l2 2-2 2",
			Stats: []ProStat{
				{Label: "Users", Value: "0"},
				{Label: "SSO connections", Value: "0"},
				{Label: "MFA adoption", Value: "0", Unit: "%"},
				{Label: "Login success", Value: "0", Unit: "%"},
			},
			Columns:   []string{"Provider", "Protocol", "Users", "MFA", "Status", "Last Sync"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Add Provider", Primary: true, Icon: "plus", Href: "#"},
				{Label: "User Directory", Icon: "users", Href: "#"},
				{Label: "MFA Policies", Icon: "shield", Href: "#"},
				{Label: "Role Manager", Icon: "key", Href: "#"},
				{Label: "Audit Log", Icon: "file", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Add an identity provider",
		},
		"security/firewall/cloud-waf": {
			Slug:     "security/firewall/cloud-waf",
			Title:    "Cloud WAF",
			Subtitle: "OWASP Core Ruleset · Bot management · Rate limiting · Geo-blocking",
			Lead:     "Protect web applications with a globally distributed web application firewall. Block OWASP Top-10 attacks, manage bot traffic with JavaScript challenges, enforce granular rate limits, and restrict access by geography or IP reputation.",
			Accent:   "#dc2626", AccentSoft: "rgba(239,68,68,0.1)",
			IconPath: "M10 2L17 5v6c0 3.5-3 6.5-7 7.5C6 17.5 3 14.5 3 11V5L10 2zm-2.5 9l2 2 4-4",
			Stats: []ProStat{
				{Label: "Rules active", Value: "0"},
				{Label: "Requests analyzed", Value: "0", Unit: "/h"},
				{Label: "Threats blocked", Value: "0", Unit: "/d"},
				{Label: "False positive rate", Value: "0", Unit: "%"},
			},
			Columns:   []string{"Policy", "Scope", "Rules", "Blocked (24h)", "Mode", "Status"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Create Policy", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Rule Editor", Icon: "edit", Href: "#"},
				{Label: "IP Allowlist", Icon: "list", Href: "#"},
				{Label: "Threat Analytics", Icon: "chart", Href: "#"},
				{Label: "Bot Scores", Icon: "shield", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Create a WAF policy",
		},
		"devops/cicd/pipelines": {
			Slug:     "devops/cicd/pipelines",
			Title:    "CI/CD Pipelines",
			Subtitle: "Build automation · Artifact registry · Canary & blue-green deploys",
			Lead:     "Automate every step from commit to production. Parallel build stages, built-in artifact registry, canary and blue-green deployment strategies, and automatic rollback on failed health checks keep your releases safe and fast.",
			Accent:   "#ea580c", AccentSoft: "rgba(249,115,22,0.1)",
			IconPath: "M2 3h5v5H2V3zm11 0h5v5h-5V3zM2 12h5v5H2v-5zm11 0h5v5h-5v-5zM7 5.5h6M7 14.5h6M4.5 8v4m11-4v4",
			Stats: []ProStat{
				{Label: "Pipelines", Value: "0"},
				{Label: "Builds (7d)", Value: "0"},
				{Label: "Deploy frequency", Value: "0", Unit: "/d"},
				{Label: "Success rate", Value: "0", Unit: "%"},
			},
			Columns:   []string{"Pipeline", "Repository", "Last Run", "Duration", "Status", "Trigger"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Create Pipeline", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Run Pipeline", Icon: "play", Href: "#"},
				{Label: "Artifact Registry", Icon: "package", Href: "#"},
				{Label: "Environment Secrets", Icon: "key", Href: "#"},
				{Label: "Deploy Strategies", Icon: "git", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Create a pipeline",
		},
		"apps/queues/message-broker": {
			Slug:     "apps/queues/message-broker",
			Title:    "Message Broker",
			Subtitle: "AMQP & MQTT protocols · Topic routing · Dead-letter queues · Exactly-once delivery",
			Lead:     "Decouple services with a fully managed message broker supporting AMQP 0.9.1 and MQTT 5.0. Topic-based routing, dead-letter queues for failed messages, exactly-once delivery semantics, and horizontal scaling to millions of messages per second.",
			Accent:   "#7e22ce", AccentSoft: "rgba(168,85,247,0.1)",
			IconPath: "M4 4h4v4H4V4zm12 0h4v4h-4V4zM4 16h4v4H4v-4zm12 0h4v4h-4v-4zM8 6h8M8 18h8M6 8v8m12-8v8",
			Stats: []ProStat{
				{Label: "Queues", Value: "0"},
				{Label: "Messages/sec", Value: "0"},
				{Label: "Consumers", Value: "0"},
				{Label: "Avg latency", Value: "0", Unit: "ms"},
			},
			Columns:   []string{"Queue / Topic", "Protocol", "Messages/sec", "Consumers", "Depth", "Status"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Create Queue", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Publish Message", Icon: "send", Href: "#"},
				{Label: "Exchange Config", Icon: "git", Href: "#"},
				{Label: "DLQ Inspector", Icon: "search", Href: "#"},
				{Label: "Consumer Groups", Icon: "users", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Create a queue",
		},
		"multicloud/cost/cloud-optimizer": {
			Slug:     "multicloud/cost/cloud-optimizer",
			Title:    "Cost Optimizer",
			Subtitle: "FinOps dashboard · Rightsizing engine · Reserved instance planner · Anomaly detection",
			Lead:     "Take control of cloud spending with real-time cost visibility, automated rightsizing recommendations, reserved instance purchase planning, and anomaly detection that alerts you before budget overruns happen.",
			Accent:   "#16a34a", AccentSoft: "rgba(22,163,74,0.1)",
			IconPath: "M10 2v3m0 12v3M2 10h3m12 0h3M4.22 4.22l2.12 2.12m7.32 7.32l2.12 2.12M15.78 4.22l-2.12 2.12m-7.32 7.32l-2.12 2.12M10 6.5a3.5 3.5 0 100 7 3.5 3.5 0 000-7z",
			Stats: []ProStat{
				{Label: "Monthly spend", Value: "0"},
				{Label: "Savings identified", Value: "0", Unit: "/mo"},
				{Label: "Waste detected", Value: "0", Unit: "%"},
				{Label: "Unit cost trend", Value: "0", Unit: "/req"},
			},
			Columns:   []string{"Recommendation", "Category", "Resource", "Current Cost", "Potential Savings", "Status"},
			Resources: []ProResource{},
			Actions: []ProAction{
				{Label: "Run Analysis", Primary: true, Icon: "chart", Href: "#"},
				{Label: "Set Budgets", Icon: "dollar", Href: "#"},
				{Label: "RI Planner", Icon: "calendar", Href: "#"},
				{Label: "Tag Policies", Icon: "tag", Href: "#"},
				{Label: "Export Report", Icon: "download", Href: "#"},
			},
			Activities: []ProActivity{},
			DocsLink:   "#", CallToAction: "Run a cost analysis",
		},
	}
}

// ── Handler ───────────────────────────────────────────────────

func proFeatureHandler(w http.ResponseWriter, r *http.Request, d WorkspaceShellData) {
	rest := strings.TrimSuffix(strings.TrimPrefix(r.PathValue("rest"), "/"), "/")
	feat, ok := proFeatureRegistry()[rest]
	if !ok {
		return // caller falls through
	}

	d.ActiveNav = "pro:" + feat.Slug
	d.PageTitle = feat.Title
	d.PageSub = feat.Subtitle

	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"actionIcon": func(name string) template.HTML {
			paths := map[string]string{
				"plus":       "M7 1v12M1 7h12",
				"key":        "M9 5a4 4 0 11-1.3 7.9L6 14H4v-2H2v-2l1.1-1.1A4 4 0 019 5zm0 0V3m2 2h2",
				"backup":     "M8 12V4M4 8l4-4 4 4M2 13h12",
				"link":       "M5 7h4m-2-4a3 3 0 000 6h2a3 3 0 000-6H7M7 7a3 3 0 000 6H5a3 3 0 000-6h2",
				"terminal":   "M1 3h12v10H1V3zm2 3.5l2 2-2 2M7 11h4",
				"logs":       "M2 4h10M2 7h10M2 10h6",
				"clock":      "M7 1a6 6 0 100 12A6 6 0 007 1zm0 3v3.5l2.5 1.5",
				"sync":       "M13 4.5A6 6 0 001 7M1 9.5A6 6 0 0013 7M1 2v3h3M13 12v-3h-3",
				"calculator": "M2 1h10v12H2V1zm0 5h10M7 1v12M4 4h.5M4 9h.5M9 9h.5M9 11h.5",
				"template":   "M1 2h12v10H1V2zm0 5h12M6 7v5",
				"block":      "M7 1a6 6 0 100 12A6 6 0 007 1zm3.5-3.5l-7 7",
				"refresh":    "M12.5 4A5.5 5.5 0 002 7M1.5 10a5.5 5.5 0 0011 0M13 2l-.5 3-3-.5",
				"rules":      "M1 3h12M1 7h8M1 11h5",
				"image":      "M1 2h12v10H1V2zm0 7l3-3 2.5 2.5L9 6l4 6M10 5a1 1 0 100 2 1 1 0 000-2z",
				"shield":     "M7 1l5 2v4c0 2.8-2.2 5-5 6-2.8-1-5-3.2-5-6V3l5-2z",
				"upload":     "M7 10V2M3 5l4-3 4 3M1 13h12",
				"camera":     "M2 4a2 2 0 00-2 2v5a2 2 0 002 2h10a2 2 0 002-2V6a2 2 0 00-2-2h-.8L10 2H4L2.8 4H2zm5 6a2 2 0 100-4 2 2 0 000 4z",
				"scale":      "M7 2v10M2 12h10M4 12l-2-6h6L6 12M7 4a2 2 0 000-4 2 2 0 000 4z",
				"git":        "M3 3a2 2 0 100 4 2 2 0 000-4zm8 0a2 2 0 100 4 2 2 0 000-4zM5 5h4M3 7v2a2 2 0 002 2h2a2 2 0 002-2V7",
				"dashboard":  "M1 1h5v5H1V1zm7 0h5v5H8V1zm-7 7h5v5H1V8zm7 0h5v5H8V8z",
				"trace":      "M1 13l3-6 3 3 3-5 3 4",
				"route":      "M1 7h4a2 2 0 012 2v2a2 2 0 002 2h4M10 5l3-3-3-3M4 9l-3 3 3 3",
				"vector":     "M7 2a5 5 0 100 10A5 5 0 007 2zm0 0V0m0 14v-2M0 7H2m10 0h2M3.1 3.1l1.4 1.4m4.9-1.4L7.9 4.5M3.1 10.9l1.4-1.4m4.9 1.4l-1.4-1.4",
				"prompt":     "M1 2h12v8H8l-3 2v-2H1V2zm2 3h8M3 7h5",
				"check":      "M1 7l4 4L13 3",
				"tune":       "M1 4h12M4 1v6M1 10h12M10 7v6",
				"arrow":      "M5 1l6 6-6 6",
				"package":    "M7 1l6 3v6l-6 3L1 10V4L7 1zm0 0v12M1 4l6 3 6-3",
				"server":     "M1 2h12v4H1V2zm0 6h12v4H1V8zm2 1.5h2M3 3.5h2",
				"search":     "M5.5 1a4.5 4.5 0 100 9 4.5 4.5 0 000-9zm0 0L13 13",
				"download":   "M7 2v9M3 8l4 3 4-3M1 13h12",
				"chart":      "M1 13l3-4 3 2 3-5 3 7",
				"heart":      "M7 12S1 8 1 4.5A3 3 0 017 3a3 3 0 016 1.5C13 8 7 12 7 12z",
				"file":       "M3 1h6l4 4v8a1 1 0 01-1 1H3a1 1 0 01-1-1V2a1 1 0 011-1zm6 0v4h4M5 8h4M5 11h4",
				"cpu":        "M3 3h8v8H3V3zm2-2v2m4-2v2M1 5h2M1 9h2M11 5h2M11 9h2M5 11v2M9 11v2",
				"users":      "M5 6a3 3 0 100-6 3 3 0 000 6zM1 13c0-2.2 1.8-4 4-4s4 1.8 4 4M10 4a2 2 0 110 4M13 13c0-2.2-1.35-4-3-4.47",
				"play":       "M3 2l10 5.5L3 13V2z",
				"send":       "M2 2l10 5-10 5V9l7-2-7-2V2z",
				"dollar":     "M7 1v12M4.5 3.5H9a2 2 0 010 4H5.5a2 2 0 000 4H10",
				"calendar":   "M1 4.5h12v8H1v-8zm3-3v3m6-3v3M1 8.5h12",
				"tag":        "M1.5 6V1.5H6L13 8.5l-4.5 4.5L1.5 6zm3-2.5a.5.5 0 110 1 .5.5 0 010-1z",
				"edit":       "M9 2.5l2.5 2.5-7 7H2V9.5l7-7z",
				"list":       "M1 3.5h12M1 7h8M1 10.5h10",
			}
			p, ok := paths[name]
			if !ok {
				p = paths["plus"]
			}
			return template.HTML(`<svg width="14" height="14" viewBox="0 0 14 14" fill="none"><path d="` + p + `" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/></svg>`)
		},
		"sparkPoints": func(vals []int, w, h int) string {
			if len(vals) == 0 {
				return ""
			}
			max := vals[0]
			min := vals[0]
			for _, v := range vals {
				if v > max {
					max = v
				}
				if v < min {
					min = v
				}
			}
			if max == min {
				max = min + 1
			}
			var b strings.Builder
			step := float64(w) / float64(len(vals)-1)
			for i, v := range vals {
				x := float64(i) * step
				y := float64(h) - (float64(v-min)/float64(max-min))*float64(h)
				if i > 0 {
					b.WriteString(" ")
				}
				b.WriteString(fmtFloat(x) + "," + fmtFloat(y))
			}
			return b.String()
		},
	}

	pageData := WorkspaceProPageData{
		WorkspaceShellData: d,
		Feature:            feat,
	}

	tmpl, err := template.New("pro_feature.html").Funcs(funcMap).ParseFiles(
		filepath.Join("templates", "workspace", "base.html"),
		filepath.Join("templates", "workspace", "shell.html"),
		filepath.Join("templates", "workspace", "pro_feature.html"),
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "workspace-base", pageData); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func fmtFloat(f float64) string {
	// tiny helper: round to 1 decimal
	i := int(f*10 + 0.5)
	whole := i / 10
	frac := i % 10
	if frac < 0 {
		frac = -frac
	}
	if whole == 0 && i < 0 {
		return "-0." + string(rune('0'+frac))
	}
	return itoa(whole) + "." + string(rune('0'+frac))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// isProFeature reports whether the path matches a registered pro feature.
func isProFeature(rest string) bool {
	rest = strings.TrimSuffix(strings.TrimPrefix(rest, "/"), "/")
	_, ok := proFeatureRegistry()[rest]
	return ok
}
