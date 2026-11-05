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
	Name     string
	Status   string // running, stopped, error, pending, healthy, active, inactive
	StatusLabel string
	Columns  []string // column values in order
	Region   string
}

type ProAction struct {
	Label string
	Icon  string
	Href  string
	Primary bool
}

type ProActivity struct {
	Time    string
	Actor   string
	Event   string
	Target  string
	Level   string // info, success, warn
}

type ProFeature struct {
	Slug        string   // URL path fragment (no trailing slash)
	Title       string
	Subtitle    string
	Lead        string   // one-line value statement
	Accent      string   // hex base color (e.g. "#4f46e5")
	AccentSoft  string   // soft bg rgba
	IconPath    string   // SVG path d
	Stats       []ProStat
	Columns     []string // column headers for resource table
	Resources   []ProResource
	Actions     []ProAction
	Activities  []ProActivity
	DocsLink    string
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
				{Label: "Active Sites", Value: "12", Delta: "+2", Up: true, Good: true, Spark: spark(4, 5, 6, 6, 8, 10, 11, 12)},
				{Label: "Deploys (7d)", Value: "94", Delta: "+31%", Up: true, Good: true, Spark: spark(48, 56, 62, 68, 72, 80, 88, 94)},
				{Label: "LCP p75", Value: "1.1", Unit: "s", Delta: "-0.2s", Up: false, Good: true, Spark: spark(18, 16, 15, 14, 14, 13, 12, 11)},
				{Label: "Avg build", Value: "28", Unit: "s", Delta: "-6s", Up: false, Good: true, Spark: spark(42, 40, 38, 36, 34, 32, 30, 28)},
			},
			Columns: []string{"Site", "Framework", "Branch / Commit", "Visits (24h)", "Build", "Region"},
			Resources: []ProResource{
				{Name: "marketing-site", Status: "healthy", StatusLabel: "Live", Columns: []string{"marketing-site", "Next.js 14.2 (App Router)", "main · a3f9c21", "42,108", "18s", "Global CDN (180 PoPs)"}},
				{Name: "docs-portal", Status: "healthy", StatusLabel: "Live", Columns: []string{"docs-portal", "Astro 4.8 + Starlight", "main · f1b08e2", "18,340", "12s", "Global CDN (180 PoPs)"}},
				{Name: "app-dashboard", Status: "healthy", StatusLabel: "Live", Columns: []string{"app-dashboard", "React 18 + Vite 5", "main · c7e4210", "87,621", "22s", "Global CDN (180 PoPs)"}},
				{Name: "blog", Status: "pending", StatusLabel: "Deploying", Columns: []string{"blog", "Hugo 0.125", "feat/seo · d92f01a", "—", "building…", "Global CDN (180 PoPs)"}},
				{Name: "landing-q2", Status: "error", StatusLabel: "Build failed", Columns: []string{"landing-q2", "Vite 5 + Vue 3.4", "main · b81c3f0", "—", "failed (42s)", "—"}},
				{Name: "checkout-widget", Status: "healthy", StatusLabel: "Live", Columns: []string{"checkout-widget", "Svelte 5 + SvelteKit", "main · e4a12b8", "124,880", "9s", "Global CDN (180 PoPs)"}},
			},
			Actions: []ProAction{
				{Label: "New Site", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Connect Git Repo", Icon: "git", Href: "#"},
				{Label: "Import from URL", Icon: "link", Href: "#"},
				{Label: "Environment Variables", Icon: "key", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "2m ago", Actor: "CI", Event: "deployed to production", Target: "app-dashboard (c7e4210 → 180 PoPs, 1.8s)", Level: "success"},
				{Time: "8m ago", Actor: "alice", Event: "enabled password protection on preview", Target: "staging.marketing-site (feat/redesign)", Level: "info"},
				{Time: "14m ago", Actor: "CI", Event: "build failed — missing env NEXT_PUBLIC_API_URL in", Target: "landing-q2 (b81c3f0)", Level: "warn"},
				{Time: "1h ago", Actor: "bob", Event: "custom domain verified via CNAME", Target: "docs.company.io → docs-portal", Level: "info"},
				{Time: "3h ago", Actor: "system", Event: "invalidated edge cache across 180 PoPs for", Target: "checkout-widget (cache-control update)", Level: "success"},
			},
			DocsLink:     "#", CallToAction: "Deploy a new site",
		},
		"compute/virtual-machines/linux-vms": {
			Slug:     "compute/virtual-machines/linux-vms",
			Title:    "Cloud Servers",
			Subtitle: "Bare-metal performance · Live migration · Per-second billing",
			Lead:     "Launch VMs in under 30 seconds across 12 regions. Live-migrate without downtime, attach NVMe volumes on the fly, and schedule snapshots with configurable retention.",
			Accent:   "#15803d", AccentSoft: "rgba(22,163,74,0.1)",
			IconPath: "M2 4h16v12H2V4zm2 2v8h12V6H4zm3 3h6M7 11h4",
			Stats: []ProStat{
				{Label: "Running VMs", Value: "18", Delta: "+3", Up: true, Good: true, Spark: spark(12, 14, 13, 15, 16, 15, 17, 18)},
				{Label: "vCPU allocated", Value: "84", Unit: "/128", Delta: "+12%", Up: true, Good: false, Spark: spark(50, 58, 62, 60, 68, 72, 78, 84)},
				{Label: "Volumes attached", Value: "24", Unit: "TB", Delta: "+2 TB", Up: true, Good: true, Spark: spark(16, 17, 18, 19, 20, 21, 22, 24)},
				{Label: "Cost efficiency", Value: "74", Unit: "%", Delta: "+6pp", Up: true, Good: true, Spark: spark(58, 60, 62, 64, 66, 68, 72, 74)},
			},
			Columns: []string{"Server", "Image", "Plan", "vCPU / Mem / Disk", "Region", "Uptime"},
			Resources: []ProResource{
				{Name: "web-prod-01", Status: "running", StatusLabel: "Running", Columns: []string{"web-prod-01", "Ubuntu 24.04 LTS", "c3.xlarge", "4 vCPU / 16 GB / 100 GB NVMe", "eu-central-1", "47d 12h"}},
				{Name: "web-prod-02", Status: "running", StatusLabel: "Running", Columns: []string{"web-prod-02", "Ubuntu 24.04 LTS", "c3.xlarge", "4 vCPU / 16 GB / 100 GB NVMe", "eu-central-1", "47d 12h"}},
				{Name: "db-primary", Status: "running", StatusLabel: "Running", Columns: []string{"db-primary", "Debian 12.5", "m3.2xlarge", "8 vCPU / 64 GB / 500 GB NVMe", "eu-central-1", "93d 06h"}},
				{Name: "worker-gpu-1", Status: "running", StatusLabel: "Running", Columns: []string{"worker-gpu-1", "Ubuntu 22.04 + CUDA 12.4", "g4.xlarge (A10G)", "8 vCPU / 32 GB / 24 GB VRAM", "eu-west-1", "3d 18h"}},
				{Name: "staging-01", Status: "pending", StatusLabel: "Provisioning", Columns: []string{"staging-01", "Ubuntu 24.04 LTS", "c2.medium", "2 vCPU / 4 GB / 50 GB SSD", "eu-central-1", "—"}},
				{Name: "legacy-win", Status: "stopped", StatusLabel: "Stopped", Columns: []string{"legacy-win", "Windows Server 2022", "c2.large", "4 vCPU / 8 GB / 120 GB SSD", "us-west-2", "— (stopped 4d)"}},
				{Name: "ci-runner-01", Status: "running", StatusLabel: "Running", Columns: []string{"ci-runner-01", "Fedora 40", "c3.medium", "4 vCPU / 8 GB / 80 GB NVMe", "us-east-1", "12d 04h"}},
			},
			Actions: []ProAction{
				{Label: "Launch VM", Primary: true, Icon: "plus", Href: "#"},
				{Label: "SSH Keys", Icon: "key", Href: "#"},
				{Label: "Snapshots", Icon: "camera", Href: "#"},
				{Label: "Firewall Rules", Icon: "shield", Href: "#"},
				{Label: "Resize Plan", Icon: "scale", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "5m ago", Actor: "you", Event: "provisioning VM", Target: "staging-01 (c2.medium, eu-central-1)", Level: "info"},
				{Time: "22m ago", Actor: "auto-scale", Event: "scaled web pool from 1 → 2 instances based on CPU rule (>70%)", Target: "", Level: "success"},
				{Time: "2h ago", Actor: "system", Event: "daily snapshot completed", Target: "db-primary → snap-0416-0200 (44 GB, retained 7d)", Level: "success"},
				{Time: "6h ago", Actor: "system", Event: "live-migrated for host maintenance", Target: "web-prod-01 (zero downtime, 1.2s pause)", Level: "info"},
				{Time: "yesterday", Actor: "carol", Event: "stopped to save cost — $0.18/hr saved", Target: "legacy-win", Level: "warn"},
			},
			DocsLink: "#", CallToAction: "Launch your first server",
		},
		"databases/relational/postgresql": {
			Slug:     "databases/relational/postgresql",
			Title:    "Managed Databases",
			Subtitle: "Automatic failover · Point-in-time recovery · Connection pooling",
			Lead:     "Production-grade PostgreSQL, MySQL, and Redis with automatic daily backups (30-day retention), transparent failover under 15 seconds, built-in connection pooling, and one-click read replica promotion.",
			Accent:   "#1d4ed8", AccentSoft: "rgba(59,130,246,0.1)",
			IconPath: "M10 2C6 2 3 3.2 3 4.5v11C3 16.8 6 18 10 18s7-1.2 7-2.5v-11C17 3.2 14 2 10 2zM3 4.5v11M17 4.5v11",
			Stats: []ProStat{
				{Label: "Clusters", Value: "7", Delta: "+1", Up: true, Good: true, Spark: spark(5, 5, 6, 6, 6, 7, 7, 7)},
				{Label: "Connections", Value: "142", Unit: "/200", Delta: "+8", Up: true, Good: false, Spark: spark(110, 120, 118, 125, 132, 138, 140, 142)},
				{Label: "QPS (avg)", Value: "4.8", Unit: "K", Delta: "+12%", Up: true, Good: true, Spark: spark(3200, 3600, 3800, 4000, 4200, 4400, 4600, 4800)},
				{Label: "Replica lag", Value: "12", Unit: "ms", Delta: "-3ms", Up: false, Good: true, Spark: spark(22, 20, 18, 16, 15, 14, 13, 12)},
			},
			Columns: []string{"Cluster", "Engine", "Plan", "Replicas", "Storage", "Region"},
			Resources: []ProResource{
				{Name: "billing-prod", Status: "healthy", StatusLabel: "Primary + HA", Columns: []string{"billing-prod", "PostgreSQL 16.3", "Pro 8 vCPU / 32 GB", "2 read · 1 standby", "184 / 500 GB", "eu-central-1"}},
				{Name: "analytics", Status: "healthy", StatusLabel: "Primary + HA", Columns: []string{"analytics", "PostgreSQL 16.3", "Pro 4 vCPU / 16 GB", "1 read", "220 / 500 GB", "eu-central-1"}},
				{Name: "sessions-cache", Status: "healthy", StatusLabel: "Cluster", Columns: []string{"sessions-cache", "Redis 7.2 (Cluster mode)", "Standard 4 GB × 3 shards", "1 replica/shard", "1.8 / 12 GB", "eu-central-1"}},
				{Name: "legacy-mysql", Status: "running", StatusLabel: "Single node", Columns: []string{"legacy-mysql", "MySQL 8.0.36", "Standard 2 vCPU / 8 GB", "—", "42 / 100 GB", "us-west-2"}},
				{Name: "staging-pg", Status: "pending", StatusLabel: "Upgrading", Columns: []string{"staging-pg", "PG 15.5 → 16.3", "Standard 2 vCPU / 8 GB", "—", "18 / 50 GB", "eu-west-1"}},
				{Name: "search-elastic", Status: "healthy", StatusLabel: "Healthy", Columns: []string{"search-elastic", "OpenSearch 2.13", "Standard 4 vCPU / 16 GB × 3", "built-in", "82 / 300 GB", "eu-central-1"}},
			},
			Actions: []ProAction{
				{Label: "Create Cluster", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Restore Backup", Icon: "backup", Href: "#"},
				{Label: "Connection Pooler", Icon: "link", Href: "#"},
				{Label: "Query Console", Icon: "terminal", Href: "#"},
				{Label: "Slow Query Log", Icon: "logs", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "1m ago", Actor: "system", Event: "daily backup completed", Target: "billing-prod (412 MB compressed, retained 30d)", Level: "success"},
				{Time: "14m ago", Actor: "you", Event: "started major-version upgrade", Target: "staging-pg (15.5 → 16.3, ETA 8 min)", Level: "info"},
				{Time: "1h ago", Actor: "system", Event: "auto-vacuum reclaimed 1.2 GB on", Target: "analytics (8 bloated tables)", Level: "info"},
				{Time: "3h ago", Actor: "alice", Event: "promoted read replica to primary (planned failover)", Target: "billing-prod (zero-downtime, 11s)", Level: "success"},
				{Time: "yesterday", Actor: "system", Event: "connection pool expanded 50 → 80 for", Target: "billing-prod (pgbouncer)", Level: "info"},
			},
			DocsLink: "#", CallToAction: "Create a database cluster",
		},
		"storage/object-storage/s3-compatible": {
			Slug:     "storage/object-storage/s3-compatible",
			Title:    "Storage & Backups",
			Subtitle: "S3-compatible API · Lifecycle tiering · Immutable object lock",
			Lead:     "Drop-in S3-compatible object storage at 80% less cost. Automatic hot → warm → archive tiering, cross-region replication with conflict resolution, and WORM compliance locks for regulatory retention.",
			Accent:   "#b45309", AccentSoft: "rgba(245,158,11,0.1)",
			IconPath: "M2 5l8-3 8 3-8 3-8-3zm0 0v10l8 3 8-3V5",
			Stats: []ProStat{
				{Label: "Buckets", Value: "9", Delta: "+1", Up: true, Good: true, Spark: spark(6, 7, 7, 8, 8, 8, 9, 9)},
				{Label: "PUT/GET (24h)", Value: "1.8", Unit: "M", Delta: "+14%", Up: true, Good: true, Spark: spark(120, 130, 140, 148, 155, 162, 170, 180)},
				{Label: "Storage", Value: "1.2", Unit: "TB", Delta: "+3.8%", Up: true, Good: false, Spark: spark(1100, 1115, 1130, 1145, 1160, 1175, 1190, 1200)},
				{Label: "Saved (vs S3)", Value: "$842", Unit: "/mo", Delta: "+$68", Up: true, Good: true, Spark: spark(620, 660, 700, 740, 770, 800, 820, 842)},
			},
			Columns: []string{"Bucket", "Region", "Objects", "Size", "Tier", "Repl / Lock"},
			Resources: []ProResource{
				{Name: "assets-public", Status: "healthy", StatusLabel: "Public-read", Columns: []string{"assets-public", "eu-central-1", "842,120", "480 GB", "Standard", "Repl → us-east-1"}},
				{Name: "backups-prod", Status: "healthy", StatusLabel: "Private", Columns: []string{"backups-prod", "eu-central-1", "1,204", "312 GB", "Archive", "WORM 90d · Repl → eu-west-1"}},
				{Name: "user-uploads", Status: "healthy", StatusLabel: "Private", Columns: []string{"user-uploads", "eu-central-1", "1,402,310", "280 GB", "Standard", "Versioned"}},
				{Name: "reports-finance", Status: "healthy", StatusLabel: "Private", Columns: []string{"reports-finance", "eu-west-1", "18,004", "64 GB", "Infrequent Access", "WORM 7yr (SOX)"}},
				{Name: "cdn-origin", Status: "healthy", StatusLabel: "Public-read", Columns: []string{"cdn-origin", "eu-central-1", "112,400", "86 GB", "Standard", "Cache-Control: 1yr"}},
				{Name: "ml-datasets", Status: "healthy", StatusLabel: "Private", Columns: []string{"ml-datasets", "us-east-1", "48,210", "420 GB", "Standard → IA (60d)", "Versioned"}},
			},
			Actions: []ProAction{
				{Label: "New Bucket", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Lifecycle Rules", Icon: "clock", Href: "#"},
				{Label: "Access Keys (HMAC)", Icon: "key", Href: "#"},
				{Label: "Replication Rules", Icon: "sync", Href: "#"},
				{Label: "Storage Calculator", Icon: "calculator", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "Just now", Actor: "lifecycle-rule", Event: "tiered 412 objects Standard → Archive in", Target: "backups-prod (saved $14/mo)", Level: "info"},
				{Time: "20m ago", Actor: "ci-bot", Event: "uploaded via presigned URL", Target: "reports-finance/q1-2026.pdf (24.8 MB, WORM locked)", Level: "success"},
				{Time: "2h ago", Actor: "alice", Event: "enabled WORM compliance lock (90d retention) on", Target: "backups-prod", Level: "success"},
				{Time: "4h ago", Actor: "replication", Event: "cross-region sync completed (delta: 842 objects) for", Target: "assets-public → us-east-1", Level: "info"},
				{Time: "yesterday", Actor: "system", Event: "incomplete multipart uploads cleaned (saved 2.1 GB) in", Target: "user-uploads", Level: "info"},
			},
			DocsLink: "#", CallToAction: "Create your first bucket",
		},
		"apps/communication/transactional-email": {
			Slug:     "apps/communication/transactional-email",
			Title:    "Email & Messaging",
			Subtitle: "SMTP & API relay · DKIM/SPF/DMARC · Real-time webhooks",
			Lead:     "Send transactional email via SMTP or REST API with automatic DKIM signing, dedicated IP warm-up, per-recipient engagement tracking, and real-time webhooks for delivered / opened / clicked / bounced events.",
			Accent:   "#9d174d", AccentSoft: "rgba(236,72,153,0.1)",
			IconPath: "M2 5a1 1 0 011-1h14a1 1 0 011 1v10a1 1 0 01-1 1H3a1 1 0 01-1-1V5zm0 0l8 6 8-6",
			Stats: []ProStat{
				{Label: "Sent (24h)", Value: "14.8", Unit: "K", Delta: "+12%", Up: true, Good: true, Spark: spark(8, 10, 9, 11, 12, 13, 14, 15)},
				{Label: "Inbox rate", Value: "99.2", Unit: "%", Delta: "+0.1pp", Up: true, Good: true, Spark: spark(98, 98, 99, 99, 99, 99, 99, 99)},
				{Label: "Open rate", Value: "42.8", Unit: "%", Delta: "+2pp", Up: true, Good: true, Spark: spark(36, 38, 37, 39, 40, 41, 42, 43)},
				{Label: "Bounce rate", Value: "0.4", Unit: "%", Delta: "-0.1pp", Up: false, Good: true, Spark: spark(6, 5, 5, 4, 4, 4, 4, 4)},
			},
			Columns: []string{"Sending Domain", "Senders", "Volume (mo)", "Reputation", "Auth", "Dedicated IP"},
			Resources: []ProResource{
				{Name: "company.io", Status: "healthy", StatusLabel: "Verified", Columns: []string{"company.io", "8 from-addresses", "247,812", "98 / 100", "DKIM + SPF + DMARC", "198.51.100.12 (warm)"}},
				{Name: "mail.company.io", Status: "healthy", StatusLabel: "Verified", Columns: []string{"mail.company.io", "3 from-addresses", "184,112", "96 / 100", "DKIM + SPF + DMARC", "198.51.100.13 (warm)"}},
				{Name: "tx.acme-corp.com", Status: "healthy", StatusLabel: "Verified", Columns: []string{"tx.acme-corp.com", "2 from-addresses", "42,108", "94 / 100", "DKIM + SPF", "Shared pool"}},
				{Name: "staging.company.io", Status: "pending", StatusLabel: "Awaiting DNS", Columns: []string{"staging.company.io", "0 from-addresses", "—", "—", "CNAME pending", "—"}},
			},
			Actions: []ProAction{
				{Label: "Add Domain", Primary: true, Icon: "plus", Href: "#"},
				{Label: "API Keys & SMTP", Icon: "key", Href: "#"},
				{Label: "Email Templates", Icon: "template", Href: "#"},
				{Label: "Suppression List", Icon: "block", Href: "#"},
				{Label: "Webhook Events", Icon: "link", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "Just now", Actor: "webhook", Event: "delivered (inbox, 280ms) for", Target: "invoice-34212 → jane@customer.com", Level: "success"},
				{Time: "3m ago", Actor: "webhook", Event: "soft-bounced (mailbox full), retry 1/3 scheduled for", Target: "user-7821@gmail.com", Level: "warn"},
				{Time: "1h ago", Actor: "alice", Event: "published template version", Target: "welcome-v3 (A/B: subject line test)", Level: "info"},
				{Time: "6h ago", Actor: "system", Event: "dedicated IP warm-up advanced to phase 4 (2K/hr) for", Target: "198.51.100.13", Level: "info"},
				{Time: "yesterday", Actor: "system", Event: "reputation score improved after DMARC alignment on", Target: "company.io (96 → 98)", Level: "success"},
			},
			DocsLink: "#", CallToAction: "Verify your sending domain",
		},
		"networking/cdn/pop-management": {
			Slug:     "networking/cdn/pop-management",
			Title:    "Content Delivery",
			Subtitle: "180+ PoPs · Instant purge · Image optimization · Edge compute",
			Lead:     "Serve content from 180+ edge locations with sub-20ms global latency. On-the-fly image resizing (WebP/AVIF), Brotli compression, instant cache invalidation (<500ms worldwide), and programmable edge rules via JavaScript workers.",
			Accent:   "#0e7490", AccentSoft: "rgba(6,182,212,0.1)",
			IconPath: "M10 2a8 8 0 100 16 8 8 0 000-16zm0 0v16M2 10h16M4 5c2 2 4 3 6 3s4-1 6-3M4 15c2-2 4-3 6-3s4 1 6 3",
			Stats: []ProStat{
				{Label: "Cache hit ratio", Value: "94.7", Unit: "%", Delta: "+0.8pp", Up: true, Good: true, Spark: spark(92, 93, 93, 94, 94, 94, 95, 95)},
				{Label: "Requests (24h)", Value: "2.4", Unit: "M", Delta: "+18%", Up: true, Good: true, Spark: spark(1200, 1400, 1600, 1800, 2000, 2100, 2300, 2400)},
				{Label: "P50 TTFB", Value: "14", Unit: "ms", Delta: "-2ms", Up: false, Good: true, Spark: spark(22, 20, 18, 17, 16, 15, 15, 14)},
				{Label: "Bandwidth saved", Value: "68", Unit: "%", Delta: "+3pp", Up: true, Good: true, Spark: spark(58, 60, 62, 63, 64, 65, 66, 68)},
			},
			Columns: []string{"Distribution", "Origin", "Hit Rate", "Reqs (24h)", "Transfer", "TLS / HTTP"},
			Resources: []ProResource{
				{Name: "cdn-main", Status: "healthy", StatusLabel: "Active", Columns: []string{"cdn-main", "assets-public.s3 (eu-central-1)", "95.1%", "1.4M", "1.2 GB (Brotli)", "TLS 1.3 / HTTP/3"}},
				{Name: "cdn-images", Status: "healthy", StatusLabel: "Active", Columns: []string{"cdn-images", "user-uploads.s3 (eu-central-1)", "97.4%", "820K", "820 MB (WebP auto)", "TLS 1.3 / HTTP/3"}},
				{Name: "cdn-api", Status: "healthy", StatusLabel: "Active", Columns: []string{"cdn-api", "api.company.io (origin shield)", "86.2%", "168K", "40 MB", "TLS 1.3 / HTTP/2"}},
				{Name: "cdn-docs", Status: "healthy", StatusLabel: "Active", Columns: []string{"cdn-docs", "docs-portal (static)", "98.8%", "12K", "18 MB (Brotli)", "TLS 1.3 / HTTP/3"}},
				{Name: "cdn-video", Status: "healthy", StatusLabel: "Active", Columns: []string{"cdn-video", "media-origin.s3 (us-east-1)", "92.0%", "4.2K", "48 GB (HLS chunks)", "TLS 1.3 / HTTP/2"}},
			},
			Actions: []ProAction{
				{Label: "New Distribution", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Purge Cache", Icon: "refresh", Href: "#"},
				{Label: "Edge Rules (JS)", Icon: "rules", Href: "#"},
				{Label: "Image Optimization", Icon: "image", Href: "#"},
				{Label: "Origin Shield", Icon: "shield", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "4m ago", Actor: "you", Event: "purged /assets/*.css — propagated to 180 PoPs in 420ms on", Target: "cdn-main", Level: "info"},
				{Time: "1h ago", Actor: "system", Event: "auto-rotated TLS certificate (Let's Encrypt) on", Target: "cdn-api", Level: "success"},
				{Time: "3h ago", Actor: "auto", Event: "traffic surge +140% absorbed at edge (0 origin overload) for", Target: "cdn-images", Level: "info"},
				{Time: "8h ago", Actor: "edge-worker", Event: "resized 42K images to WebP (saved 62% bandwidth) on", Target: "cdn-images", Level: "success"},
				{Time: "yesterday", Actor: "bob", Event: "added geo-block rule (CN, RU) on", Target: "cdn-main (edge rule #14)", Level: "info"},
			},
			DocsLink: "#", CallToAction: "Create a distribution",
		},
		"security/certificates/ssl-tls": {
			Slug:     "security/certificates/ssl-tls",
			Title:    "Security & SSL",
			Subtitle: "Auto-renewing certs · Managed WAF · Secrets vault · DDoS shield",
			Lead:     "Zero-touch TLS: certificates are issued, deployed to edge, and renewed automatically. Layer-7 WAF with OWASP Core Rule Set, real-time DDoS mitigation up to 10 Tbps, and a KMS-backed secrets vault with audit trail.",
			Accent:   "#059669", AccentSoft: "rgba(16,185,129,0.1)",
			IconPath: "M10 1L3 4v5c0 4 3 7.5 7 9 4-1.5 7-5 7-9V4l-7-3zm0 6v4m0 2v.5",
			Stats: []ProStat{
				{Label: "Active certs", Value: "22", Delta: "+2", Up: true, Good: true, Spark: spark(18, 19, 20, 20, 21, 22, 22, 22)},
				{Label: "WAF blocks (24h)", Value: "1,248", Delta: "+6%", Up: true, Good: true, Spark: spark(800, 950, 1000, 1050, 1100, 1150, 1200, 1250)},
				{Label: "DDoS mitigated", Value: "3.2", Unit: "Gbps", Delta: "peak", Up: false, Good: true, Spark: spark(1, 1, 2, 1, 3, 2, 1, 1)},
				{Label: "Vault secrets", Value: "148", Delta: "+4", Up: true, Good: true, Spark: spark(130, 134, 138, 140, 142, 144, 146, 148)},
			},
			Columns: []string{"Certificate", "Type / Grade", "Issuer", "Expires", "Auto-renew", "Endpoints"},
			Resources: []ProResource{
				{Name: "*.company.io", Status: "healthy", StatusLabel: "A+ (SSL Labs)", Columns: []string{"*.company.io", "Wildcard DV · TLS 1.3", "Let's Encrypt R3", "2026-07-12 (88d)", "On (ACME DNS-01)", "CDN + LB (4 origins)"}},
				{Name: "api.company.io", Status: "healthy", StatusLabel: "A+ (SSL Labs)", Columns: []string{"api.company.io", "Single DV · TLS 1.3", "Let's Encrypt R3", "2026-06-04 (50d)", "On (ACME HTTP-01)", "API gateway"}},
				{Name: "docs.company.io", Status: "pending", StatusLabel: "Renewing", Columns: []string{"docs.company.io", "Single DV · TLS 1.3", "Let's Encrypt R3", "2026-04-26 (11d)", "On (renewal queued)", "CDN"}},
				{Name: "legacy.acme-corp.com", Status: "error", StatusLabel: "Expired -5d", Columns: []string{"legacy.acme-corp.com", "Single OV · TLS 1.2", "Sectigo", "2026-04-10 (-5d)", "Off (manual)", "Direct origin"}},
				{Name: "admin.company.io", Status: "healthy", StatusLabel: "A+ (SSL Labs)", Columns: []string{"admin.company.io", "Single DV · TLS 1.3", "Let's Encrypt R3", "2026-09-22 (160d)", "On (ACME DNS-01)", "Internal LB + mTLS"}},
			},
			Actions: []ProAction{
				{Label: "Issue Certificate", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Upload Certificate", Icon: "upload", Href: "#"},
				{Label: "WAF Rules (OWASP)", Icon: "shield", Href: "#"},
				{Label: "Secrets Vault", Icon: "key", Href: "#"},
				{Label: "DDoS Settings", Icon: "shield", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "2m ago", Actor: "WAF", Event: "blocked SQLi attempt (OWASP CRS 942100) from", Target: "103.21.x.x → api.company.io/login", Level: "warn"},
				{Time: "34m ago", Actor: "DDoS shield", Event: "mitigated 3.2 Gbps volumetric attack (UDP flood, 48s) on", Target: "cdn-main (zero user impact)", Level: "warn"},
				{Time: "1h ago", Actor: "ACME", Event: "auto-renewed certificate (DNS-01 challenge, 12s) for", Target: "admin.company.io (+90 days)", Level: "success"},
				{Time: "5h ago", Actor: "alice", Event: "rotated secret in vault (version 4 → 5, audit logged)", Target: "prod/stripe_api_key", Level: "success"},
				{Time: "yesterday", Actor: "system", Event: "expiry warning — auto-renewal scheduled in 3 days for", Target: "docs.company.io (11 days remaining)", Level: "warn"},
			},
			DocsLink: "#", CallToAction: "Issue your first certificate",
		},
		"devops/monitoring/metrics": {
			Slug:     "devops/monitoring/metrics",
			Title:    "Monitoring & Alerts",
			Subtitle: "OpenTelemetry-native · PromQL queries · PagerDuty / Slack routing",
			Lead:     "Unified observability: ingest metrics via OpenTelemetry or Prometheus remote-write, query with PromQL, correlate with logs and traces in one timeline, and route alerts through escalation policies to PagerDuty, Slack, or webhooks.",
			Accent:   "#7c3aed", AccentSoft: "rgba(168,85,247,0.1)",
			IconPath: "M2 14l4-4 3 3 4-6 5 7",
			Stats: []ProStat{
				{Label: "Series ingested", Value: "14.2", Unit: "M/min", Delta: "+8%", Up: true, Good: true, Spark: spark(10, 11, 11, 12, 13, 13, 14, 14)},
				{Label: "Active alerts", Value: "4", Delta: "+1", Up: true, Good: false, Spark: spark(2, 2, 3, 3, 3, 4, 4, 4)},
				{Label: "Log volume (24h)", Value: "28", Unit: "GB", Delta: "+4%", Up: true, Good: false, Spark: spark(22, 23, 24, 25, 25, 26, 27, 28)},
				{Label: "MTTR (7d avg)", Value: "8", Unit: "min", Delta: "-2m", Up: false, Good: true, Spark: spark(14, 12, 11, 10, 10, 9, 9, 8)},
			},
			Columns: []string{"Alert", "Severity", "PromQL / Condition", "Firing", "Scope", "Escalation"},
			Resources: []ProResource{
				{Name: "High CPU — web-prod", Status: "error", StatusLabel: "Firing 14m", Columns: []string{"High CPU — web-prod", "P1 Critical", "avg(rate(cpu_usage[5m])) > 0.85", "14m", "web-prod-01, web-prod-02", "→ PagerDuty (#oncall)"}},
				{Name: "Queue backlog", Status: "error", StatusLabel: "Firing 4m", Columns: []string{"Queue backlog", "P2 Warning", "jobs_queue_depth > 500", "4m", "jobs.critical", "→ Slack #alerts"}},
				{Name: "DB replica lag", Status: "pending", StatusLabel: "Pending", Columns: []string{"DB replica lag", "P2 Warning", "pg_replication_lag_ms > 200", "just now", "billing-prod-replica-1", "→ Slack #db-alerts"}},
				{Name: "5xx error rate", Status: "healthy", StatusLabel: "Silenced 2h", Columns: []string{"5xx error rate", "P1 Critical", "rate(http_5xx[5m]) / rate(http_total[5m]) > 0.01", "—", "api-gateway", "Silenced (deploy)"}},
				{Name: "Disk usage", Status: "healthy", StatusLabel: "OK", Columns: []string{"Disk usage", "P3 Info", "node_filesystem_avail_pct < 15", "—", "all VMs (18)", "→ Slack #infra"}},
				{Name: "SSL cert expiry", Status: "healthy", StatusLabel: "OK", Columns: []string{"SSL cert expiry", "P2 Warning", "ssl_cert_expires_in_days < 14", "—", "all endpoints (22)", "→ Email ops@"}},
			},
			Actions: []ProAction{
				{Label: "New Alert Rule", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Dashboards", Icon: "dashboard", Href: "#"},
				{Label: "Log Explorer", Icon: "logs", Href: "#"},
				{Label: "Trace Waterfall", Icon: "trace", Href: "#"},
				{Label: "Escalation Policies", Icon: "route", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "Just now", Actor: "alert-manager", Event: "fired — paged oncall via PagerDuty for", Target: "DB replica lag (pg_replication_lag_ms = 248)", Level: "warn"},
				{Time: "14m ago", Actor: "alert-manager", Event: "fired — escalated to P1 after 10m for", Target: "High CPU — web-prod (auto-scale triggered)", Level: "warn"},
				{Time: "1h ago", Actor: "bob", Event: "created PromQL dashboard", Target: "payments-latency-v2 (6 panels, P50/P95/P99)", Level: "info"},
				{Time: "3h ago", Actor: "auto-remediation", Event: "scaled web pool 2 → 4 in response to", Target: "High CPU — web-prod (CPU dropped to 52%)", Level: "success"},
				{Time: "yesterday", Actor: "alice", Event: "updated escalation policy — added 15m timeout to", Target: "P1 Critical (PagerDuty → Slack #incidents → email)", Level: "info"},
			},
			DocsLink: "#", CallToAction: "Build your first dashboard",
		},
		"ai/inference/llm-endpoints": {
			Slug:     "ai/inference/llm-endpoints",
			Title:    "AI & Automation",
			Subtitle: "vLLM-powered inference · GGUF & SafeTensors · Auto-scaling to zero",
			Lead:     "Deploy open-weight LLMs on dedicated GPUs with vLLM continuous batching, automatic KV-cache offloading, quantized GGUF support, and scale-to-zero when idle. Includes managed Qdrant vector DB and a visual prompt playground.",
			Accent:   "#c2410c", AccentSoft: "rgba(249,115,22,0.1)",
			IconPath: "M10 2a8 8 0 018 8c0 3-2 6-4 7l-1 2h-6l-1-2c-2-1-4-4-4-7a8 8 0 018-8zm-2 7a1 1 0 100 2 1 1 0 000-2zm4 0a1 1 0 100 2 1 1 0 000-2z",
			Stats: []ProStat{
				{Label: "Endpoints", Value: "5", Delta: "+1", Up: true, Good: true, Spark: spark(2, 3, 3, 4, 4, 4, 5, 5)},
				{Label: "Tokens (24h)", Value: "18.4", Unit: "M", Delta: "+15%", Up: true, Good: true, Spark: spark(12, 13, 14, 15, 16, 17, 18, 18)},
				{Label: "P95 TTFT", Value: "142", Unit: "ms", Delta: "-28ms", Up: false, Good: true, Spark: spark(240, 220, 200, 190, 180, 170, 155, 142)},
				{Label: "GPU utilization", Value: "78", Unit: "%", Delta: "+4pp", Up: true, Good: true, Spark: spark(62, 64, 68, 70, 72, 74, 76, 78)},
			},
			Columns: []string{"Endpoint", "Model / Quant", "GPU", "Throughput", "P95 TTFT / TPS", "Status"},
			Resources: []ProResource{
				{Name: "chat-assistant", Status: "running", StatusLabel: "Serving", Columns: []string{"chat-assistant", "Llama-3.1-70B-Instruct (AWQ 4-bit)", "2× A100 80 GB", "12.4 req/s", "142ms / 84 tok/s", "Serving (vLLM)"}},
				{Name: "code-gen", Status: "running", StatusLabel: "Serving", Columns: []string{"code-gen", "DeepSeek-Coder-V2-Lite (FP16)", "1× A100 40 GB", "4.2 req/s", "280ms / 62 tok/s", "Serving (vLLM)"}},
				{Name: "embed-prod", Status: "running", StatusLabel: "Serving", Columns: []string{"embed-prod", "bge-large-en-v1.5 (ONNX)", "1× T4 16 GB", "48.6 req/s", "42ms / —", "Serving (batch)"}},
				{Name: "classifier", Status: "running", StatusLabel: "Serving", Columns: []string{"classifier", "DistilBERT-base (fine-tuned, ONNX)", "CPU (4 vCPU)", "102 req/s", "18ms / —", "Serving (CPU)"}},
				{Name: "vision-ocr", Status: "pending", StatusLabel: "Cold start", Columns: []string{"vision-ocr", "Qwen2-VL-7B (AWQ 4-bit)", "1× A10G 24 GB", "—", "—", "Scaling 0 → 1"}},
			},
			Actions: []ProAction{
				{Label: "Deploy Model", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Vector DB (Qdrant)", Icon: "vector", Href: "#"},
				{Label: "Prompt Playground", Icon: "prompt", Href: "#"},
				{Label: "Eval Suites", Icon: "check", Href: "#"},
				{Label: "Fine-tune (LoRA)", Icon: "tune", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "Just now", Actor: "auto-scale", Event: "cold-starting from zero (GPU provisioning, ETA 45s)", Target: "vision-ocr (1 req queued)", Level: "info"},
				{Time: "32m ago", Actor: "you", Event: "deployed v5 — swapped AWQ quant, TTFT dropped 28ms on", Target: "code-gen", Level: "success"},
				{Time: "2h ago", Actor: "system", Event: "reindexed Qdrant collection (HNSW rebuild, 4.2M vectors, 18s)", Target: "docs-embeddings", Level: "success"},
				{Time: "6h ago", Actor: "auto-scale", Event: "scaled to zero after 30m idle (saved $2.40/hr)", Target: "vision-ocr", Level: "info"},
				{Time: "yesterday", Actor: "alice", Event: "eval suite passed on v4 checkpoint", Target: "chat-assistant (92.4% on internal bench, +1.2pp vs v3)", Level: "info"},
			},
			DocsLink: "#", CallToAction: "Deploy your first model",
		},
		"apps/serverless/functions": {
			Slug:     "apps/serverless/functions",
			Title:    "Serverless Functions",
			Subtitle: "V8 isolate cold starts · HTTP / cron / queue triggers · Pay-per-invocation",
			Lead:     "Deploy functions in Node, Python, Go, or Rust with V8-isolate cold starts under 10ms. HTTP routes, cron schedules, and queue consumers as triggers — billed per millisecond of execution with no idle cost.",
			Accent:   "#4338ca", AccentSoft: "rgba(99,102,241,0.1)",
			IconPath: "M5 2l5 7-3 1 5 6-2-7 3-1-5-6",
			Stats: []ProStat{
				{Label: "Functions", Value: "24", Delta: "+3", Up: true, Good: true, Spark: spark(18, 19, 20, 21, 22, 23, 24, 24)},
				{Label: "Invocations (24h)", Value: "2.1", Unit: "M", Delta: "+9%", Up: true, Good: true, Spark: spark(1500, 1600, 1700, 1800, 1900, 2000, 2050, 2100)},
				{Label: "P99 cold start", Value: "8", Unit: "ms", Delta: "-2ms", Up: false, Good: true, Spark: spark(14, 12, 11, 10, 9, 9, 8, 8)},
				{Label: "Error budget", Value: "99.88", Unit: "% OK", Delta: "+0.03pp", Up: true, Good: true, Spark: spark(9970, 9975, 9978, 9980, 9982, 9984, 9986, 9988)},
			},
			Columns: []string{"Function", "Runtime", "Trigger", "Invocations (24h)", "P50 / P99", "Errors (24h)"},
			Resources: []ProResource{
				{Name: "webhook-stripe", Status: "healthy", StatusLabel: "Active", Columns: []string{"webhook-stripe", "Node 22 (128 MB)", "HTTP POST /hooks/stripe", "42,108", "12ms / 28ms", "0"}},
				{Name: "image-resize", Status: "healthy", StatusLabel: "Active", Columns: []string{"image-resize", "Node 22 (256 MB)", "HTTP GET (edge, 180 PoPs)", "820,400", "38ms / 92ms", "4"}},
				{Name: "daily-report", Status: "healthy", StatusLabel: "Scheduled", Columns: []string{"daily-report", "Python 3.12 (512 MB)", "Cron: 0 6 * * * (UTC)", "1", "4.2s / 4.2s", "0"}},
				{Name: "queue-worker", Status: "healthy", StatusLabel: "Active", Columns: []string{"queue-worker", "Go 1.22 (128 MB)", "Queue: jobs.default (concurrency: 10)", "184,120", "82ms / 240ms", "12"}},
				{Name: "auth-verify", Status: "healthy", StatusLabel: "Active", Columns: []string{"auth-verify", "Rust (WASM, 64 MB)", "HTTP GET /auth/verify", "1,042,800", "1.4ms / 3.8ms", "0"}},
				{Name: "legacy-cb", Status: "error", StatusLabel: "Degraded", Columns: []string{"legacy-cb", "Node 16 (EOL ⚠)", "HTTP POST /callbacks", "248", "180ms / 1.2s", "18 (7.3%)"}},
			},
			Actions: []ProAction{
				{Label: "New Function", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Cron Schedules", Icon: "clock", Href: "#"},
				{Label: "Environment & Secrets", Icon: "key", Href: "#"},
				{Label: "Deploy via CLI", Icon: "terminal", Href: "#"},
				{Label: "Logs & Tracing", Icon: "logs", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "3m ago", Actor: "CI", Event: "deployed v42 (Node 22, 256 MB, SHA a91c4f2) to", Target: "image-resize (zero-downtime canary: 10% → 100%)", Level: "success"},
				{Time: "14m ago", Actor: "cron", Event: "completed daily run", Target: "daily-report (4.2s, 2,841 rows exported to S3)", Level: "success"},
				{Time: "1h ago", Actor: "monitor", Event: "error rate exceeded 5% threshold on", Target: "legacy-cb (18 errors / 248 invocations — Node 16 EOL)", Level: "warn"},
				{Time: "2h ago", Actor: "alice", Event: "rotated secret (version 3 → 4)", Target: "STRIPE_WEBHOOK_SECRET (webhook-stripe, auth-verify)", Level: "info"},
				{Time: "yesterday", Actor: "system", Event: "auto-upgraded runtime", Target: "queue-worker (Go 1.21 → 1.22, cold start -1ms)", Level: "success"},
			},
			DocsLink: "#", CallToAction: "Write your first function",
		},
		"networking/dns/zones": {
			Slug:     "networking/dns/zones",
			Title:    "Domains & DNS",
			Subtitle: "Registrar · DNS · Domains",
			Lead:     "Register, transfer, and manage domains with DNSSEC, ANAME flattening, geolocation routing, and sub-millisecond resolution.",
			Accent:   "#374151", AccentSoft: "rgba(107,114,128,0.1)",
			IconPath: "M10 2a8 8 0 00-8 8 8 8 0 008 8 8 8 0 008-8 8 8 0 00-8-8zm0 0c-2.5 2-4 5-4 8s1.5 6 4 8c2.5-2 4-5 4-8s-1.5-6-4-8zM2 10h16",
			Stats: []ProStat{
				{Label: "Domains", Value: "14", Delta: "+1", Up: true, Good: true, Spark: spark(10, 11, 12, 12, 13, 13, 14, 14)},
				{Label: "DNS zones", Value: "22", Delta: "+2", Up: true, Good: true, Spark: spark(16, 17, 18, 19, 20, 21, 22, 22)},
				{Label: "Queries (24h)", Value: "8.4", Unit: "M", Delta: "+4%", Up: true, Good: true, Spark: spark(6.0 * 100, 620, 660, 700, 740, 780, 820, 840)},
				{Label: "Resolution p50", Value: "12", Unit: "ms", Delta: "-1ms", Up: false, Good: true, Spark: spark(16, 15, 15, 14, 13, 13, 12, 12)},
			},
			Columns: []string{"Domain", "Registrar", "DNS", "Records", "DNSSEC", "Expires"},
			Resources: []ProResource{
				{Name: "company.io", Status: "healthy", StatusLabel: "Active", Columns: []string{"company.io", "Platform Registrar", "Platform DNS", "48", "Enabled", "2027-01-14 (274d)"}},
				{Name: "company.com", Status: "healthy", StatusLabel: "Active", Columns: []string{"company.com", "Platform Registrar", "Platform DNS", "18", "Enabled", "2026-12-02 (231d)"}},
				{Name: "acme-corp.com", Status: "healthy", StatusLabel: "Active", Columns: []string{"acme-corp.com", "External", "Platform DNS", "24", "Enabled", "2026-08-20 (127d)"}},
				{Name: "beta.company.io", Status: "healthy", StatusLabel: "Active", Columns: []string{"beta.company.io", "—", "Platform DNS", "12", "Inherited", "—"}},
				{Name: "newproduct.dev", Status: "pending", StatusLabel: "Transferring", Columns: []string{"newproduct.dev", "Platform Registrar", "Platform DNS", "—", "Pending", "Transferring"}},
			},
			Actions: []ProAction{
				{Label: "Register Domain", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Transfer Domain", Icon: "arrow", Href: "#"},
				{Label: "Import Zone", Icon: "upload", Href: "#"},
				{Label: "Auto-renew Settings", Icon: "refresh", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "Just now", Actor: "you", Event: "added A record", Target: "api.company.io → 203.0.113.42", Level: "info"},
				{Time: "1h ago", Actor: "transfer", Event: "waiting on losing-registrar auth for", Target: "newproduct.dev", Level: "warn"},
				{Time: "yesterday", Actor: "system", Event: "auto-renewed", Target: "company.io (+1 year)", Level: "success"},
				{Time: "2d ago", Actor: "alice", Event: "enabled DNSSEC on", Target: "acme-corp.com", Level: "success"},
			},
			DocsLink: "#", CallToAction: "Register a domain",
		},
		"compute/kubernetes/managed-k8s": {
			Slug:     "compute/kubernetes/managed-k8s",
			Title:    "Managed Kubernetes",
			Subtitle: "Multi-cluster orchestration · Auto-scaling node pools · GitOps-ready",
			Lead:     "Run production-grade Kubernetes without managing control planes. Auto-scaling node pools, one-click version upgrades, integrated GitOps workflows, and built-in observability across every cluster.",
			Accent:   "#2563eb", AccentSoft: "rgba(37,99,235,0.1)",
			IconPath: "M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5",
			Stats: []ProStat{
				{Label: "Clusters", Value: "6", Delta: "+1", Up: true, Good: true, Spark: spark(3, 3, 4, 4, 4, 5, 5, 6)},
				{Label: "Pods running", Value: "847", Delta: "+12%", Up: true, Good: true, Spark: spark(520, 580, 640, 690, 720, 760, 810, 847)},
				{Label: "Node utilization", Value: "72", Unit: "%", Delta: "+4pp", Up: true, Good: true, Spark: spark(58, 60, 63, 65, 66, 68, 70, 72)},
				{Label: "Deploy frequency", Value: "38", Unit: "/d", Delta: "+9", Up: true, Good: true, Spark: spark(18, 22, 24, 26, 29, 32, 35, 38)},
			},
			Columns: []string{"Cluster", "Version", "Node Pools", "Pods", "Region", "API Endpoint"},
			Resources: []ProResource{
				{Name: "prod-us-east", Status: "healthy", StatusLabel: "Healthy", Columns: []string{"prod-us-east", "1.30.2", "3 (12 nodes)", "312 / 360", "us-east-1", "https://k8s-prod-use1.platform.io"}},
				{Name: "prod-eu-west", Status: "healthy", StatusLabel: "Healthy", Columns: []string{"prod-eu-west", "1.30.2", "2 (8 nodes)", "246 / 280", "eu-west-1", "https://k8s-prod-euw1.platform.io"}},
				{Name: "staging", Status: "healthy", StatusLabel: "Healthy", Columns: []string{"staging", "1.31.0-rc.1", "1 (4 nodes)", "128 / 160", "us-east-1", "https://k8s-stg.platform.io"}},
				{Name: "ml-workloads", Status: "running", StatusLabel: "Scaling", Columns: []string{"ml-workloads", "1.30.2", "2 (6 nodes)", "94 / 120", "us-west-2", "https://k8s-ml.platform.io"}},
				{Name: "dev-sandbox", Status: "healthy", StatusLabel: "Healthy", Columns: []string{"dev-sandbox", "1.30.1", "1 (3 nodes)", "42 / 80", "eu-central-1", "https://k8s-dev.platform.io"}},
				{Name: "edge-apac", Status: "pending", StatusLabel: "Provisioning", Columns: []string{"edge-apac", "1.30.2", "— (pending)", "—", "ap-southeast-1", "—"}},
			},
			Actions: []ProAction{
				{Label: "Create Cluster", Primary: true, Icon: "plus", Href: "#"},
				{Label: "kubectl Access", Icon: "terminal", Href: "#"},
				{Label: "Helm Charts", Icon: "package", Href: "#"},
				{Label: "Upgrade Version", Icon: "refresh", Href: "#"},
				{Label: "Node Pool", Icon: "server", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "4m ago", Actor: "CI", Event: "deployed v2.14.0 via Helm to", Target: "prod-us-east (canary 10%)", Level: "success"},
				{Time: "18m ago", Actor: "autoscaler", Event: "scaled node pool gpu-pool from 4 to 6 in", Target: "ml-workloads", Level: "info"},
				{Time: "1h ago", Actor: "bob", Event: "upgraded control plane 1.29.4 → 1.30.2 on", Target: "prod-eu-west", Level: "success"},
				{Time: "3h ago", Actor: "system", Event: "pod CrashLoopBackOff detected in", Target: "staging/payments-svc (OOMKilled)", Level: "warn"},
				{Time: "5h ago", Actor: "alice", Event: "applied NetworkPolicy restricting egress in", Target: "prod-us-east/default", Level: "info"},
			},
			DocsLink: "#", CallToAction: "Create a cluster",
		},
		"storage/volumes/block-storage": {
			Slug:     "storage/volumes/block-storage",
			Title:    "Block Storage",
			Subtitle: "NVMe-backed volumes · Instant snapshots · AES-256 encryption at rest",
			Lead:     "Attach high-performance NVMe block volumes to any compute instance. Provision up to 100k IOPS per volume, schedule automatic snapshots, and encrypt everything at rest with customer-managed keys.",
			Accent:   "#7c3aed", AccentSoft: "rgba(124,92,255,0.1)",
			IconPath: "M4 4h16v4H4V4zm0 6h16v4H4v-4zm0 6h16v4H4v-4zM7 6h2M7 12h2M7 18h2",
			Stats: []ProStat{
				{Label: "Volumes", Value: "34", Delta: "+4", Up: true, Good: true, Spark: spark(22, 24, 26, 27, 28, 30, 32, 34)},
				{Label: "Total IOPS", Value: "248k", Delta: "+18%", Up: true, Good: true, Spark: spark(140, 155, 170, 185, 200, 215, 232, 248)},
				{Label: "Throughput", Value: "4.2", Unit: "GB/s", Delta: "+0.6", Up: true, Good: true, Spark: spark(28, 30, 32, 34, 35, 37, 40, 42)},
				{Label: "Snapshot count", Value: "186", Delta: "+12", Up: true, Good: true, Spark: spark(120, 130, 140, 148, 156, 164, 174, 186)},
			},
			Columns: []string{"Volume", "Type", "Size", "IOPS", "Attached To", "Status"},
			Resources: []ProResource{
				{Name: "db-primary-data", Status: "healthy", StatusLabel: "In-use", Columns: []string{"db-primary-data", "NVMe io2", "500 GB", "64,000", "db-primary", "Attached"}},
				{Name: "db-replica-data", Status: "healthy", StatusLabel: "In-use", Columns: []string{"db-replica-data", "NVMe io2", "500 GB", "64,000", "db-replica-01", "Attached"}},
				{Name: "app-data-01", Status: "healthy", StatusLabel: "In-use", Columns: []string{"app-data-01", "NVMe gp3", "200 GB", "16,000", "web-prod-01", "Attached"}},
				{Name: "ml-scratch", Status: "healthy", StatusLabel: "In-use", Columns: []string{"ml-scratch", "NVMe io2", "1 TB", "100,000", "worker-gpu-1", "Attached"}},
				{Name: "archive-q1", Status: "stopped", StatusLabel: "Available", Columns: []string{"archive-q1", "Standard HDD", "2 TB", "500", "—", "Detached"}},
				{Name: "staging-boot", Status: "pending", StatusLabel: "Creating", Columns: []string{"staging-boot", "NVMe gp3", "50 GB", "3,000", "staging-01", "Attaching"}},
			},
			Actions: []ProAction{
				{Label: "Create Volume", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Take Snapshot", Icon: "camera", Href: "#"},
				{Label: "Resize Volume", Icon: "scale", Href: "#"},
				{Label: "Encryption Keys", Icon: "key", Href: "#"},
				{Label: "Restore Snapshot", Icon: "refresh", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "6m ago", Actor: "system", Event: "snapshot completed for", Target: "db-primary-data (snap-20260416-0910, 22 GB)", Level: "success"},
				{Time: "22m ago", Actor: "alice", Event: "resized volume from 100 GB to 200 GB on", Target: "app-data-01 (online, zero-downtime)", Level: "info"},
				{Time: "2h ago", Actor: "system", Event: "IOPS throttle warning on", Target: "ml-scratch (sustained 98k / 100k limit)", Level: "warn"},
				{Time: "6h ago", Actor: "bob", Event: "detached and archived", Target: "archive-q1 from legacy-win", Level: "info"},
				{Time: "1d ago", Actor: "system", Event: "rotated encryption key for", Target: "db-primary-data (CMK rotation policy)", Level: "success"},
			},
			DocsLink: "#", CallToAction: "Create a volume",
		},
		"databases/nosql/document-db": {
			Slug:     "databases/nosql/document-db",
			Title:    "Document Database",
			Subtitle: "MongoDB-compatible API · Auto-sharding · Real-time change streams",
			Lead:     "Fully managed document database with MongoDB wire-protocol compatibility. Automatic sharding distributes data across zones, change streams power real-time event pipelines, and point-in-time recovery protects every write.",
			Accent:   "#047857", AccentSoft: "rgba(16,185,129,0.1)",
			IconPath: "M4 6c0-1.1 3.6-2 8-2s8 .9 8 2M4 6v4c0 1.1 3.6 2 8 2s8-.9 8-2V6M4 10v4c0 1.1 3.6 2 8 2s8-.9 8-2v-4M4 14v4c0 1.1 3.6 2 8 2s8-.9 8-2v-4",
			Stats: []ProStat{
				{Label: "Clusters", Value: "5", Delta: "+1", Up: true, Good: true, Spark: spark(3, 3, 3, 4, 4, 4, 4, 5)},
				{Label: "Collections", Value: "142", Delta: "+18", Up: true, Good: true, Spark: spark(80, 90, 100, 108, 116, 124, 132, 142)},
				{Label: "Documents", Value: "48M", Delta: "+6M", Up: true, Good: true, Spark: spark(20, 25, 28, 32, 36, 40, 44, 48)},
				{Label: "Ops/sec", Value: "12.4k", Delta: "+22%", Up: true, Good: true, Spark: spark(62, 70, 78, 84, 90, 100, 110, 124)},
			},
			Columns: []string{"Cluster", "Tier", "Storage", "Connections", "Region", "Replica Set"},
			Resources: []ProResource{
				{Name: "prod-atlas-main", Status: "healthy", StatusLabel: "Primary", Columns: []string{"prod-atlas-main", "M40 (Dedicated)", "128 GB / 512 GB", "340 / 500", "eu-central-1", "rs0 (3-node)"}},
				{Name: "prod-atlas-analytics", Status: "healthy", StatusLabel: "Healthy", Columns: []string{"prod-atlas-analytics", "M30 (Dedicated)", "86 GB / 256 GB", "120 / 300", "eu-central-1", "rs1 (3-node)"}},
				{Name: "sessions-store", Status: "healthy", StatusLabel: "Healthy", Columns: []string{"sessions-store", "M20 (Shared)", "14 GB / 64 GB", "82 / 200", "us-east-1", "rs2 (3-node)"}},
				{Name: "search-catalog", Status: "running", StatusLabel: "Rebalancing", Columns: []string{"search-catalog", "M40 (Dedicated)", "210 GB / 512 GB", "190 / 500", "us-east-1", "rs3 (3-node, shard x2)"}},
				{Name: "dev-sandbox", Status: "pending", StatusLabel: "Restoring", Columns: []string{"dev-sandbox", "M10 (Shared)", "— / 32 GB", "—", "eu-west-1", "rs4 (restoring PITR)"}},
			},
			Actions: []ProAction{
				{Label: "Create Cluster", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Connection Strings", Icon: "link", Href: "#"},
				{Label: "Data Explorer", Icon: "search", Href: "#"},
				{Label: "Backup & Restore", Icon: "download", Href: "#"},
				{Label: "Index Advisor", Icon: "chart", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "3m ago", Actor: "system", Event: "shard rebalance in progress on", Target: "search-catalog (chunk migration 68%)", Level: "info"},
				{Time: "30m ago", Actor: "alice", Event: "created compound index on", Target: "prod-atlas-main.orders (userId + createdAt)", Level: "success"},
				{Time: "2h ago", Actor: "system", Event: "slow query detected (>500ms) in", Target: "prod-atlas-analytics.events (full collection scan)", Level: "warn"},
				{Time: "4h ago", Actor: "bob", Event: "initiated point-in-time restore to", Target: "dev-sandbox (target: 2026-04-15T14:00Z)", Level: "info"},
				{Time: "1d ago", Actor: "system", Event: "automated backup completed for", Target: "prod-atlas-main (snapshot: 42 GB, retained 7d)", Level: "success"},
			},
			DocsLink: "#", CallToAction: "Launch a cluster",
		},
		"networking/load-balancer/application-lb": {
			Slug:     "networking/load-balancer/application-lb",
			Title:    "Application Load Balancer",
			Subtitle: "Layer-7 routing · Automatic TLS termination · Path-based rules · Active health checks",
			Lead:     "Distribute HTTP/HTTPS traffic across backends with content-based routing, automatic certificate management, WebSocket support, and real-time health checks that remove unhealthy targets in under 10 seconds.",
			Accent:   "#0284c7", AccentSoft: "rgba(14,165,233,0.1)",
			IconPath: "M12 2v6m0 8v6M2 12h6m8 0h6M8 8l-4-4M16 8l4-4M8 16l-4 4M16 16l4 4",
			Stats: []ProStat{
				{Label: "Load Balancers", Value: "8", Delta: "+1", Up: true, Good: true, Spark: spark(5, 5, 6, 6, 6, 7, 7, 8)},
				{Label: "Active connections", Value: "24.6k", Delta: "+18%", Up: true, Good: true, Spark: spark(140, 152, 168, 180, 194, 210, 228, 246)},
				{Label: "RPS", Value: "18.2k", Delta: "+2.4k", Up: true, Good: true, Spark: spark(98, 110, 120, 132, 142, 156, 170, 182)},
				{Label: "Healthy targets", Value: "98", Unit: "%", Delta: "+1pp", Up: true, Good: true, Spark: spark(94, 95, 95, 96, 96, 97, 97, 98)},
			},
			Columns: []string{"Load Balancer", "Protocol", "Targets", "RPS", "Cert Status", "Region"},
			Resources: []ProResource{
				{Name: "api-gateway-prod", Status: "healthy", StatusLabel: "Active", Columns: []string{"api-gateway-prod", "HTTPS (TLS 1.3)", "6/6 healthy", "8,420", "Valid (247d)", "eu-central-1"}},
				{Name: "web-frontend-prod", Status: "healthy", StatusLabel: "Active", Columns: []string{"web-frontend-prod", "HTTPS (TLS 1.3)", "4/4 healthy", "6,100", "Valid (182d)", "eu-central-1"}},
				{Name: "grpc-services", Status: "healthy", StatusLabel: "Active", Columns: []string{"grpc-services", "gRPC + TLS", "8/8 healthy", "2,800", "Valid (310d)", "us-east-1"}},
				{Name: "websocket-lb", Status: "healthy", StatusLabel: "Active", Columns: []string{"websocket-lb", "WSS", "3/3 healthy", "640", "Valid (88d)", "us-east-1"}},
				{Name: "staging-lb", Status: "error", StatusLabel: "Degraded", Columns: []string{"staging-lb", "HTTPS", "2/4 healthy", "120", "Expired", "eu-west-1"}},
			},
			Actions: []ProAction{
				{Label: "Create Load Balancer", Primary: true, Icon: "plus", Href: "#"},
				{Label: "TLS Certificates", Icon: "shield", Href: "#"},
				{Label: "Routing Rules", Icon: "git", Href: "#"},
				{Label: "Health Checks", Icon: "heart", Href: "#"},
				{Label: "Access Logs", Icon: "file", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "1m ago", Actor: "system", Event: "target failed health check in", Target: "staging-lb (staging-03: HTTP 503, 3 consecutive)", Level: "warn"},
				{Time: "12m ago", Actor: "certbot", Event: "auto-renewed TLS certificate for", Target: "api-gateway-prod (*.api.company.io, Let's Encrypt)", Level: "success"},
				{Time: "45m ago", Actor: "alice", Event: "added path rule /v2/* forwarding to", Target: "api-gateway-prod → api-v2-target-group", Level: "info"},
				{Time: "3h ago", Actor: "system", Event: "connection spike detected on", Target: "web-frontend-prod (24.6k active, 92% capacity)", Level: "warn"},
				{Time: "8h ago", Actor: "bob", Event: "enabled sticky sessions on", Target: "websocket-lb (cookie: WSLB, TTL: 3600s)", Level: "info"},
			},
			DocsLink: "#", CallToAction: "Create a load balancer",
		},
		"ai/vector-search/embeddings": {
			Slug:     "ai/vector-search/embeddings",
			Title:    "Vector Search",
			Subtitle: "Purpose-built vector DB · Semantic search · RAG pipeline integration",
			Lead:     "Index and query billions of high-dimensional vectors with sub-millisecond latency. Build semantic search, recommendation engines, and retrieval-augmented generation pipelines with native embedding model support and hybrid keyword+vector queries.",
			Accent:   "#db2777", AccentSoft: "rgba(236,72,153,0.1)",
			IconPath: "M12 2a4 4 0 014 4c0 1.5-.8 2.8-2 3.5M12 2a4 4 0 00-4 4c0 1.5.8 2.8 2 3.5M6 14l6 6 6-6M3 18h18M8 10h8",
			Stats: []ProStat{
				{Label: "Collections", Value: "14", Delta: "+3", Up: true, Good: true, Spark: spark(6, 7, 8, 9, 10, 11, 12, 14)},
				{Label: "Vectors indexed", Value: "128M", Delta: "+34M", Up: true, Good: true, Spark: spark(40, 52, 64, 74, 84, 98, 112, 128)},
				{Label: "Queries/sec", Value: "4.8k", Delta: "+41%", Up: true, Good: true, Spark: spark(18, 22, 26, 30, 34, 38, 42, 48)},
				{Label: "Recall@10", Value: "97.2", Unit: "%", Delta: "+0.4pp", Up: true, Good: true, Spark: spark(952, 958, 960, 964, 966, 968, 970, 972)},
			},
			Columns: []string{"Collection", "Dimensions", "Vectors", "Index Type", "Queries/sec", "Status"},
			Resources: []ProResource{
				{Name: "product-embeddings", Status: "healthy", StatusLabel: "Serving", Columns: []string{"product-embeddings", "1536", "42M", "HNSW (M=32)", "1,840", "Indexed"}},
				{Name: "support-articles", Status: "healthy", StatusLabel: "Serving", Columns: []string{"support-articles", "768", "2.4M", "HNSW (M=16)", "620", "Indexed"}},
				{Name: "user-profiles", Status: "healthy", StatusLabel: "Serving", Columns: []string{"user-profiles", "1536", "18M", "IVF-PQ (nlist=4096)", "1,200", "Indexed"}},
				{Name: "code-search", Status: "running", StatusLabel: "Indexing", Columns: []string{"code-search", "1024", "64M (82%)", "HNSW (M=48)", "—", "Building index"}},
				{Name: "image-features", Status: "healthy", StatusLabel: "Serving", Columns: []string{"image-features", "512", "1.8M", "HNSW (M=16)", "340", "Indexed"}},
			},
			Actions: []ProAction{
				{Label: "Create Collection", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Query Playground", Icon: "search", Href: "#"},
				{Label: "Embedding Models", Icon: "cpu", Href: "#"},
				{Label: "Import Vectors", Icon: "upload", Href: "#"},
				{Label: "RAG Templates", Icon: "template", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "2m ago", Actor: "indexer", Event: "indexing progress 82% on", Target: "code-search (52M / 64M vectors, ETA 18m)", Level: "info"},
				{Time: "20m ago", Actor: "alice", Event: "created collection with HNSW index for", Target: "image-features (512d, cosine similarity)", Level: "success"},
				{Time: "1h ago", Actor: "system", Event: "recall degradation detected on", Target: "user-profiles (Recall@10 dropped to 94.1%, reindexing)", Level: "warn"},
				{Time: "4h ago", Actor: "CI", Event: "deployed RAG pipeline v3.2 using", Target: "support-articles + GPT-4o (hybrid search)", Level: "success"},
				{Time: "1d ago", Actor: "bob", Event: "migrated embedding model from ada-002 to", Target: "product-embeddings (text-embedding-3-large)", Level: "info"},
			},
			DocsLink: "#", CallToAction: "Create a collection",
		},
		"security/iam/identity-platform": {
			Slug:     "security/iam/identity-platform",
			Title:    "Identity Platform",
			Subtitle: "Enterprise SSO · SCIM provisioning · Adaptive MFA · Fine-grained RBAC",
			Lead:     "Centralize authentication and authorization across your organization. Connect enterprise identity providers via SAML/OIDC, automate user lifecycle with SCIM, enforce adaptive MFA policies, and define fine-grained roles with attribute-based access control.",
			Accent:   "#0f766e", AccentSoft: "rgba(20,184,166,0.1)",
			IconPath: "M12 2a5 5 0 015 5c0 2.76-2.24 5-5 5s-5-2.24-5-5a5 5 0 015-5zM2 20c0-3.31 4.03-6 9-6h2c4.97 0 9 2.69 9 6M16 4l2 2-2 2",
			Stats: []ProStat{
				{Label: "Users", Value: "4,218", Delta: "+186", Up: true, Good: true, Spark: spark(3200, 3400, 3550, 3680, 3820, 3940, 4080, 4218)},
				{Label: "SSO connections", Value: "12", Delta: "+2", Up: true, Good: true, Spark: spark(6, 7, 8, 8, 9, 10, 11, 12)},
				{Label: "MFA adoption", Value: "94", Unit: "%", Delta: "+3pp", Up: true, Good: true, Spark: spark(78, 82, 84, 86, 88, 90, 92, 94)},
				{Label: "Login success", Value: "99.4", Unit: "%", Delta: "+0.2pp", Up: true, Good: true, Spark: spark(988, 990, 990, 992, 992, 993, 993, 994)},
			},
			Columns: []string{"Provider", "Protocol", "Users", "MFA", "Status", "Last Sync"},
			Resources: []ProResource{
				{Name: "Okta (Corporate)", Status: "healthy", StatusLabel: "Connected", Columns: []string{"Okta (Corporate)", "SAML 2.0", "2,840", "Enforced", "Synced", "2m ago"}},
				{Name: "Azure AD (Partners)", Status: "healthy", StatusLabel: "Connected", Columns: []string{"Azure AD (Partners)", "OIDC", "680", "Enforced", "Synced", "5m ago"}},
				{Name: "Google Workspace", Status: "healthy", StatusLabel: "Connected", Columns: []string{"Google Workspace", "OIDC", "420", "Optional", "Synced", "8m ago"}},
				{Name: "LDAP (Legacy)", Status: "running", StatusLabel: "Migrating", Columns: []string{"LDAP (Legacy)", "LDAPS", "210", "Not supported", "Migrating", "—"}},
				{Name: "GitHub (Dev SSO)", Status: "healthy", StatusLabel: "Connected", Columns: []string{"GitHub (Dev SSO)", "OAuth 2.0", "68", "Org-enforced", "Synced", "12m ago"}},
				{Name: "SCIM Provisioner", Status: "error", StatusLabel: "Sync Error", Columns: []string{"SCIM Provisioner", "SCIM 2.0", "—", "—", "Error", "Failed 22m ago"}},
			},
			Actions: []ProAction{
				{Label: "Add Provider", Primary: true, Icon: "plus", Href: "#"},
				{Label: "User Directory", Icon: "users", Href: "#"},
				{Label: "MFA Policies", Icon: "shield", Href: "#"},
				{Label: "Role Manager", Icon: "key", Href: "#"},
				{Label: "Audit Log", Icon: "file", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "5m ago", Actor: "system", Event: "SCIM sync failed for", Target: "SCIM Provisioner (HTTP 401 — token expired)", Level: "warn"},
				{Time: "22m ago", Actor: "alice", Event: "enforced MFA for all users in", Target: "Azure AD (Partners) — 680 users affected", Level: "success"},
				{Time: "1h ago", Actor: "system", Event: "provisioned 12 new users via SCIM from", Target: "Okta (Corporate) — engineering group", Level: "info"},
				{Time: "6h ago", Actor: "bob", Event: "created custom role api-read-only with", Target: "3 permissions across 2 resource types", Level: "info"},
				{Time: "1d ago", Actor: "security", Event: "blocked suspicious login attempt for", Target: "admin@company.io (impossible travel: US → CN in 20m)", Level: "warn"},
			},
			DocsLink: "#", CallToAction: "Add an identity provider",
		},
		"security/firewall/cloud-waf": {
			Slug:     "security/firewall/cloud-waf",
			Title:    "Cloud WAF",
			Subtitle: "OWASP Core Ruleset · Bot management · Rate limiting · Geo-blocking",
			Lead:     "Protect web applications with a globally distributed web application firewall. Block OWASP Top-10 attacks, manage bot traffic with JavaScript challenges, enforce granular rate limits, and restrict access by geography or IP reputation.",
			Accent:   "#dc2626", AccentSoft: "rgba(239,68,68,0.1)",
			IconPath: "M10 2L17 5v6c0 3.5-3 6.5-7 7.5C6 17.5 3 14.5 3 11V5L10 2zm-2.5 9l2 2 4-4",
			Stats: []ProStat{
				{Label: "Rules active", Value: "842", Delta: "+36", Up: true, Good: true, Spark: spark(680, 710, 740, 760, 780, 800, 820, 842)},
				{Label: "Requests analyzed", Value: "2.1M", Unit: "/h", Delta: "+28%", Up: true, Good: true, Spark: spark(12, 14, 15, 16, 17, 18, 20, 21)},
				{Label: "Threats blocked", Value: "14.8k", Unit: "/d", Delta: "+1.2k", Up: true, Good: false, Spark: spark(88, 96, 102, 110, 118, 128, 138, 148)},
				{Label: "False positive rate", Value: "0.02", Unit: "%", Delta: "-0.01pp", Up: false, Good: true, Spark: spark(8, 7, 6, 6, 5, 4, 3, 2)},
			},
			Columns: []string{"Policy", "Scope", "Rules", "Blocked (24h)", "Mode", "Status"},
			Resources: []ProResource{
				{Name: "prod-api-shield", Status: "healthy", StatusLabel: "Enforcing", Columns: []string{"prod-api-shield", "api.company.io/*", "342", "6,240", "Block", "Active"}},
				{Name: "web-app-protect", Status: "healthy", StatusLabel: "Enforcing", Columns: []string{"web-app-protect", "app.company.io/*", "286", "4,180", "Block", "Active"}},
				{Name: "bot-management", Status: "healthy", StatusLabel: "Enforcing", Columns: []string{"bot-management", "*.company.io", "124", "3,820", "Challenge", "Active"}},
				{Name: "rate-limiter-global", Status: "healthy", StatusLabel: "Enforcing", Columns: []string{"rate-limiter-global", "*.company.io", "18", "480", "Block (429)", "Active"}},
				{Name: "staging-monitor", Status: "running", StatusLabel: "Log-only", Columns: []string{"staging-monitor", "staging.company.io/*", "342", "— (log)", "Detect", "Monitoring"}},
				{Name: "geo-block-ofac", Status: "healthy", StatusLabel: "Enforcing", Columns: []string{"geo-block-ofac", "*.company.io", "12", "92", "Block (403)", "Active"}},
			},
			Actions: []ProAction{
				{Label: "Create Policy", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Rule Editor", Icon: "edit", Href: "#"},
				{Label: "IP Allowlist", Icon: "list", Href: "#"},
				{Label: "Threat Analytics", Icon: "chart", Href: "#"},
				{Label: "Bot Scores", Icon: "shield", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "Just now", Actor: "waf", Event: "blocked SQLi attempt on", Target: "prod-api-shield (POST /api/v1/search — UNION SELECT)", Level: "warn"},
				{Time: "8m ago", Actor: "system", Event: "rate limit triggered for", Target: "rate-limiter-global (IP 198.51.100.44 — 520 req/min)", Level: "info"},
				{Time: "35m ago", Actor: "alice", Event: "promoted staging-monitor rules to block mode on", Target: "prod-api-shield (18 new rules)", Level: "success"},
				{Time: "2h ago", Actor: "system", Event: "bot challenge pass rate dropped to 12% for", Target: "bot-management (credential-stuffing cluster)", Level: "warn"},
				{Time: "6h ago", Actor: "bob", Event: "added OFAC-sanctioned regions to", Target: "geo-block-ofac (5 country codes added)", Level: "info"},
			},
			DocsLink: "#", CallToAction: "Create a WAF policy",
		},
		"devops/cicd/pipelines": {
			Slug:     "devops/cicd/pipelines",
			Title:    "CI/CD Pipelines",
			Subtitle: "Build automation · Artifact registry · Canary & blue-green deploys",
			Lead:     "Automate every step from commit to production. Parallel build stages, built-in artifact registry, canary and blue-green deployment strategies, and automatic rollback on failed health checks keep your releases safe and fast.",
			Accent:   "#ea580c", AccentSoft: "rgba(249,115,22,0.1)",
			IconPath: "M2 3h5v5H2V3zm11 0h5v5h-5V3zM2 12h5v5H2v-5zm11 0h5v5h-5v-5zM7 5.5h6M7 14.5h6M4.5 8v4m11-4v4",
			Stats: []ProStat{
				{Label: "Pipelines", Value: "24", Delta: "+3", Up: true, Good: true, Spark: spark(14, 16, 17, 18, 19, 20, 22, 24)},
				{Label: "Builds (7d)", Value: "312", Delta: "+18%", Up: true, Good: true, Spark: spark(180, 200, 220, 240, 256, 272, 294, 312)},
				{Label: "Deploy frequency", Value: "8.4", Unit: "/d", Delta: "+1.6", Up: true, Good: true, Spark: spark(42, 50, 56, 60, 66, 72, 78, 84)},
				{Label: "Success rate", Value: "96.2", Unit: "%", Delta: "+1.1pp", Up: true, Good: true, Spark: spark(920, 928, 936, 940, 944, 950, 956, 962)},
			},
			Columns: []string{"Pipeline", "Repository", "Last Run", "Duration", "Status", "Trigger"},
			Resources: []ProResource{
				{Name: "api-service-build", Status: "healthy", StatusLabel: "Passed", Columns: []string{"api-service-build", "company/api-service", "4m ago (v2.14.0)", "3m 42s", "Passed", "push → main"}},
				{Name: "frontend-deploy", Status: "running", StatusLabel: "Running", Columns: []string{"frontend-deploy", "company/web-app", "Running (2m 10s)", "~4m est.", "Stage: test", "push → main"}},
				{Name: "ml-model-train", Status: "healthy", StatusLabel: "Passed", Columns: []string{"ml-model-train", "company/ml-pipeline", "2h ago", "28m 14s", "Passed", "schedule (daily)"}},
				{Name: "infra-terraform", Status: "healthy", StatusLabel: "Passed", Columns: []string{"infra-terraform", "company/infra", "6h ago", "5m 08s", "Passed", "PR merge → main"}},
				{Name: "mobile-ios-build", Status: "error", StatusLabel: "Failed", Columns: []string{"mobile-ios-build", "company/ios-app", "1h ago", "12m 30s", "Failed (test)", "push → develop"}},
				{Name: "security-scan", Status: "healthy", StatusLabel: "Passed", Columns: []string{"security-scan", "company/* (monorepo)", "30m ago", "8m 22s", "Passed", "schedule (hourly)"}},
			},
			Actions: []ProAction{
				{Label: "Create Pipeline", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Run Pipeline", Icon: "play", Href: "#"},
				{Label: "Artifact Registry", Icon: "package", Href: "#"},
				{Label: "Environment Secrets", Icon: "key", Href: "#"},
				{Label: "Deploy Strategies", Icon: "git", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "2m ago", Actor: "CI", Event: "started canary deploy (10%) for", Target: "api-service-build v2.14.0 → prod-us-east", Level: "info"},
				{Time: "18m ago", Actor: "system", Event: "build failed — test suite error in", Target: "mobile-ios-build (UITests/LoginSpec: timeout)", Level: "warn"},
				{Time: "1h ago", Actor: "alice", Event: "promoted canary to 100% for", Target: "api-service-build v2.13.8 (all health checks green)", Level: "success"},
				{Time: "4h ago", Actor: "bob", Event: "added secret STRIPE_KEY to", Target: "api-service-build (prod environment)", Level: "info"},
				{Time: "8h ago", Actor: "CI", Event: "auto-rolled back failed deploy for", Target: "frontend-deploy v4.2.0 (error rate > 5%)", Level: "warn"},
			},
			DocsLink: "#", CallToAction: "Create a pipeline",
		},
		"apps/queues/message-broker": {
			Slug:     "apps/queues/message-broker",
			Title:    "Message Broker",
			Subtitle: "AMQP & MQTT protocols · Topic routing · Dead-letter queues · Exactly-once delivery",
			Lead:     "Decouple services with a fully managed message broker supporting AMQP 0.9.1 and MQTT 5.0. Topic-based routing, dead-letter queues for failed messages, exactly-once delivery semantics, and horizontal scaling to millions of messages per second.",
			Accent:   "#7e22ce", AccentSoft: "rgba(168,85,247,0.1)",
			IconPath: "M4 4h4v4H4V4zm12 0h4v4h-4V4zM4 16h4v4H4v-4zm12 0h4v4h-4v-4zM8 6h8M8 18h8M6 8v8m12-8v8",
			Stats: []ProStat{
				{Label: "Queues", Value: "42", Delta: "+6", Up: true, Good: true, Spark: spark(24, 28, 30, 32, 34, 36, 38, 42)},
				{Label: "Messages/sec", Value: "86k", Delta: "+24%", Up: true, Good: true, Spark: spark(44, 50, 56, 62, 68, 74, 80, 86)},
				{Label: "Consumers", Value: "128", Delta: "+14", Up: true, Good: true, Spark: spark(80, 88, 94, 100, 106, 112, 120, 128)},
				{Label: "Avg latency", Value: "2.4", Unit: "ms", Delta: "-0.8ms", Up: false, Good: true, Spark: spark(52, 48, 44, 40, 36, 32, 28, 24)},
			},
			Columns: []string{"Queue / Topic", "Protocol", "Messages/sec", "Consumers", "Depth", "Status"},
			Resources: []ProResource{
				{Name: "orders.created", Status: "healthy", StatusLabel: "Active", Columns: []string{"orders.created", "AMQP", "12,400", "8", "240", "Draining"}},
				{Name: "payments.process", Status: "healthy", StatusLabel: "Active", Columns: []string{"payments.process", "AMQP", "3,200", "4", "18", "Healthy"}},
				{Name: "notifications.email", Status: "healthy", StatusLabel: "Active", Columns: []string{"notifications.email", "AMQP", "1,800", "6", "42", "Healthy"}},
				{Name: "iot/sensors/telemetry", Status: "healthy", StatusLabel: "Active", Columns: []string{"iot/sensors/telemetry", "MQTT 5.0", "64,000", "32", "1,200", "High volume"}},
				{Name: "analytics.events", Status: "running", StatusLabel: "Backpressure", Columns: []string{"analytics.events", "AMQP", "4,600", "2", "84,200", "Backpressure"}},
				{Name: "orders.created.dlq", Status: "error", StatusLabel: "DLQ Alert", Columns: []string{"orders.created.dlq", "AMQP (DLQ)", "—", "0", "1,842", "Needs attention"}},
			},
			Actions: []ProAction{
				{Label: "Create Queue", Primary: true, Icon: "plus", Href: "#"},
				{Label: "Publish Message", Icon: "send", Href: "#"},
				{Label: "Exchange Config", Icon: "git", Href: "#"},
				{Label: "DLQ Inspector", Icon: "search", Href: "#"},
				{Label: "Consumer Groups", Icon: "users", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "Just now", Actor: "system", Event: "backpressure detected on", Target: "analytics.events (depth 84k, 2 consumers — add capacity)", Level: "warn"},
				{Time: "10m ago", Actor: "system", Event: "dead-letter queue threshold breached on", Target: "orders.created.dlq (1,842 messages, investigate failures)", Level: "warn"},
				{Time: "40m ago", Actor: "alice", Event: "scaled consumers from 4 to 8 on", Target: "orders.created (Black Friday traffic spike)", Level: "success"},
				{Time: "3h ago", Actor: "bob", Event: "created MQTT topic with QoS 2 for", Target: "iot/sensors/telemetry (exactly-once delivery)", Level: "info"},
				{Time: "12h ago", Actor: "CI", Event: "deployed consumer v3.1.0 with retry policy to", Target: "payments.process (max 3 retries, exp. backoff)", Level: "success"},
			},
			DocsLink: "#", CallToAction: "Create a queue",
		},
		"multicloud/cost/cloud-optimizer": {
			Slug:     "multicloud/cost/cloud-optimizer",
			Title:    "Cost Optimizer",
			Subtitle: "FinOps dashboard · Rightsizing engine · Reserved instance planner · Anomaly detection",
			Lead:     "Take control of cloud spending with real-time cost visibility, automated rightsizing recommendations, reserved instance purchase planning, and anomaly detection that alerts you before budget overruns happen.",
			Accent:   "#16a34a", AccentSoft: "rgba(22,163,74,0.1)",
			IconPath: "M10 2v3m0 12v3M2 10h3m12 0h3M4.22 4.22l2.12 2.12m7.32 7.32l2.12 2.12M15.78 4.22l-2.12 2.12m-7.32 7.32l-2.12 2.12M10 6.5a3.5 3.5 0 100 7 3.5 3.5 0 000-7z",
			Stats: []ProStat{
				{Label: "Monthly spend", Value: "$48.2k", Delta: "-8%", Up: false, Good: true, Spark: spark(580, 560, 540, 530, 520, 510, 496, 482)},
				{Label: "Savings identified", Value: "$12.6k", Unit: "/mo", Delta: "+$2.1k", Up: true, Good: true, Spark: spark(62, 72, 80, 88, 96, 104, 116, 126)},
				{Label: "Waste detected", Value: "18", Unit: "%", Delta: "-4pp", Up: false, Good: true, Spark: spark(32, 30, 28, 26, 24, 22, 20, 18)},
				{Label: "Unit cost trend", Value: "$0.024", Unit: "/req", Delta: "-12%", Up: false, Good: true, Spark: spark(38, 36, 34, 32, 30, 28, 26, 24)},
			},
			Columns: []string{"Recommendation", "Category", "Resource", "Current Cost", "Potential Savings", "Status"},
			Resources: []ProResource{
				{Name: "Rightsize db-analytics", Status: "active", StatusLabel: "Action needed", Columns: []string{"Downsize instance", "Rightsizing", "db-analytics (m5.4xl → m5.2xl)", "$1,248/mo", "$624/mo (50%)", "Recommended"}},
				{Name: "Delete unused volumes", Status: "active", StatusLabel: "Action needed", Columns: []string{"Remove unattached", "Waste", "14 volumes (2.8 TB, 90d unused)", "$392/mo", "$392/mo (100%)", "Recommended"}},
				{Name: "RI purchase plan", Status: "healthy", StatusLabel: "Optimized", Columns: []string{"Reserve capacity", "Commitment", "12x c5.xlarge (1yr no-upfront)", "$4,320/mo", "$1,728/mo (40%)", "Purchased"}},
				{Name: "Spot fleet migration", Status: "running", StatusLabel: "In progress", Columns: []string{"Use spot instances", "Compute", "ci-runners (8 instances)", "$864/mo", "$605/mo (70%)", "Migrating"}},
				{Name: "S3 lifecycle policies", Status: "healthy", StatusLabel: "Optimized", Columns: []string{"Tiered storage", "Storage", "logs-* buckets (12 TB → Glacier)", "$276/mo", "$248/mo (90%)", "Applied"}},
				{Name: "Budget anomaly Q2", Status: "error", StatusLabel: "Alert", Columns: []string{"Spending anomaly", "Anomaly", "GPU cluster (320% above forecast)", "+$4,200/mo", "— (investigate)", "Alerting"}},
			},
			Actions: []ProAction{
				{Label: "Run Analysis", Primary: true, Icon: "chart", Href: "#"},
				{Label: "Set Budgets", Icon: "dollar", Href: "#"},
				{Label: "RI Planner", Icon: "calendar", Href: "#"},
				{Label: "Tag Policies", Icon: "tag", Href: "#"},
				{Label: "Export Report", Icon: "download", Href: "#"},
			},
			Activities: []ProActivity{
				{Time: "12m ago", Actor: "optimizer", Event: "identified $624/mo savings by rightsizing", Target: "db-analytics (avg CPU 18%, peak 42% over 30d)", Level: "info"},
				{Time: "1h ago", Actor: "system", Event: "budget anomaly detected on", Target: "GPU cluster — $4,200/mo over forecast (model training spike)", Level: "warn"},
				{Time: "3h ago", Actor: "alice", Event: "purchased 1-year reserved instances for", Target: "12x c5.xlarge — projected savings $20,736/yr", Level: "success"},
				{Time: "8h ago", Actor: "system", Event: "applied S3 Intelligent-Tiering to", Target: "logs-* buckets (12 TB, estimated 90% cost reduction)", Level: "success"},
				{Time: "1d ago", Actor: "bob", Event: "tagged 42 untagged resources in", Target: "us-east-1 (cost allocation: team, env, project)", Level: "info"},
			},
			DocsLink: "#", CallToAction: "Run a cost analysis",
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
