package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// MarketplaceApp is a one-click deployable app entry.
type MarketplaceApp struct {
	Slug        string
	Name        string
	Tagline     string
	Category    string
	Icon        string // Inline SVG or letter mark
	IconColor   string // Brand color hex (background tint)
	IconFG      string // Foreground color
	Version     string
	License     string
	Publisher   string
	Rating      float64
	Installs    string // e.g. "12.4k"
	Tags        []string
	Featured    bool
	Verified    bool
	MinRAM      string // e.g. "1 GB"
	MinCPU      string // e.g. "1 vCPU"
	MinDisk     string // e.g. "10 GB"
	Description string
}

// MarketplaceCategory is a top-level filter category.
type MarketplaceCategory struct {
	ID    string
	Label string
	Count int
}

// MarketplaceData is passed to the marketplace template.
type MarketplaceData struct {
	WorkspaceShellData
	Categories []MarketplaceCategory
	Apps       []MarketplaceApp
	Featured   []MarketplaceApp
}

func marketplaceApps() []MarketplaceApp {
	return []MarketplaceApp{
		// ── CMS & Content ────────────────────────────────────────
		{Slug: "wordpress", Name: "WordPress", Tagline: "The world's most popular CMS", Category: "cms",
			IconColor: "#21759b", IconFG: "#ffffff", Version: "6.5", License: "GPL-2.0",
			Publisher: "WordPress Foundation", Rating: 4.8, Installs: "284k", Featured: true, Verified: true,
			Tags: []string{"PHP", "MySQL", "Blog"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Build any kind of website — blog, portfolio, storefront. Comes preconfigured with LAMP stack."},

		{Slug: "ghost", Name: "Ghost", Tagline: "Modern publishing platform", Category: "cms",
			IconColor: "#15171a", IconFG: "#ffffff", Version: "5.80", License: "MIT",
			Publisher: "Ghost Foundation", Rating: 4.7, Installs: "41.2k", Featured: true, Verified: true,
			Tags: []string{"Node.js", "Newsletter", "Blog"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "5 GB",
			Description: "Professional publishing for creators — built-in memberships, newsletters, and SEO."},

		{Slug: "strapi", Name: "Strapi", Tagline: "Headless CMS for developers", Category: "cms",
			IconColor: "#4945ff", IconFG: "#ffffff", Version: "4.25", License: "MIT",
			Publisher: "Strapi Solutions", Rating: 4.6, Installs: "28.9k", Verified: true,
			Tags: []string{"Node.js", "REST", "GraphQL"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "8 GB",
			Description: "Headless CMS with auto-generated REST and GraphQL APIs."},

		{Slug: "medusa", Name: "Medusa", Tagline: "Open-source Shopify alternative", Category: "ecommerce",
			IconColor: "#000000", IconFG: "#ffffff", Version: "1.20", License: "MIT",
			Publisher: "Medusa", Rating: 4.5, Installs: "11.3k", Verified: true,
			Tags: []string{"Node.js", "Commerce", "API-first"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "10 GB",
			Description: "Headless commerce engine — fully customizable digital storefronts."},

		// ── Databases ────────────────────────────────────────────
		{Slug: "postgresql", Name: "PostgreSQL", Tagline: "Advanced relational database", Category: "databases",
			IconColor: "#336791", IconFG: "#ffffff", Version: "16.2", License: "PostgreSQL",
			Publisher: "PostgreSQL Global Dev Group", Rating: 4.9, Installs: "512k", Featured: true, Verified: true,
			Tags: []string{"SQL", "ACID", "JSON"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "The world's most advanced open source relational database — extensible, standards-compliant."},

		{Slug: "mysql", Name: "MySQL", Tagline: "The world's most popular RDBMS", Category: "databases",
			IconColor: "#00758f", IconFG: "#ffffff", Version: "8.3", License: "GPL-2.0",
			Publisher: "Oracle", Rating: 4.5, Installs: "389k", Verified: true,
			Tags: []string{"SQL", "Replication"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Widely adopted SQL database, tuned defaults and automatic backups enabled."},

		{Slug: "mongodb", Name: "MongoDB", Tagline: "Document database for modern apps", Category: "databases",
			IconColor: "#00684a", IconFG: "#ffffff", Version: "7.0", License: "SSPL",
			Publisher: "MongoDB Inc", Rating: 4.6, Installs: "198k", Verified: true,
			Tags: []string{"NoSQL", "JSON", "Replica set"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "General-purpose document database. Rich query language with flexible schema."},

		{Slug: "redis", Name: "Redis", Tagline: "In-memory data structure store", Category: "databases",
			IconColor: "#dc382d", IconFG: "#ffffff", Version: "7.2", License: "BSD-3",
			Publisher: "Redis Ltd", Rating: 4.9, Installs: "441k", Featured: true, Verified: true,
			Tags: []string{"Cache", "Pub/Sub", "Streams"}, MinRAM: "512 MB", MinCPU: "1 vCPU", MinDisk: "5 GB",
			Description: "Lightning-fast in-memory store for caching, queues, real-time analytics, and more."},

		{Slug: "mariadb", Name: "MariaDB", Tagline: "Community-developed MySQL fork", Category: "databases",
			IconColor: "#003545", IconFG: "#ffffff", Version: "11.3", License: "GPL-2.0",
			Publisher: "MariaDB Foundation", Rating: 4.6, Installs: "87.2k", Verified: true,
			Tags: []string{"SQL", "MySQL-compatible"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Drop-in replacement for MySQL with additional storage engines and features."},

		{Slug: "clickhouse", Name: "ClickHouse", Tagline: "Columnar OLAP database", Category: "databases",
			IconColor: "#ffcc01", IconFG: "#000000", Version: "24.3", License: "Apache-2.0",
			Publisher: "ClickHouse Inc", Rating: 4.8, Installs: "24.7k", Verified: true,
			Tags: []string{"Analytics", "OLAP", "Columnar"}, MinRAM: "4 GB", MinCPU: "2 vCPU", MinDisk: "40 GB",
			Description: "Blazing-fast columnar database for real-time analytics at scale."},

		// ── DevOps & Monitoring ──────────────────────────────────
		{Slug: "gitlab", Name: "GitLab CE", Tagline: "The complete DevOps platform", Category: "devops",
			IconColor: "#fc6d26", IconFG: "#ffffff", Version: "16.10", License: "MIT",
			Publisher: "GitLab Inc", Rating: 4.5, Installs: "62.1k", Featured: true, Verified: true,
			Tags: []string{"Git", "CI/CD", "Registry"}, MinRAM: "4 GB", MinCPU: "2 vCPU", MinDisk: "50 GB",
			Description: "Self-hosted Git repo, CI/CD runners, container registry — all in one."},

		{Slug: "gitea", Name: "Gitea", Tagline: "Lightweight Git service", Category: "devops",
			IconColor: "#609926", IconFG: "#ffffff", Version: "1.21", License: "MIT",
			Publisher: "Gitea Ltd", Rating: 4.7, Installs: "38.5k", Verified: true,
			Tags: []string{"Git", "Self-hosted"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Painless self-hosted Git service — a lightweight alternative to GitLab."},

		{Slug: "jenkins", Name: "Jenkins", Tagline: "The leading automation server", Category: "devops",
			IconColor: "#d33833", IconFG: "#ffffff", Version: "2.440", License: "MIT",
			Publisher: "CD Foundation", Rating: 4.3, Installs: "72.8k", Verified: true,
			Tags: []string{"CI/CD", "Pipeline", "Plugins"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "30 GB",
			Description: "Build, test, and deploy with 1800+ plugins. Industry-standard CI server."},

		{Slug: "grafana", Name: "Grafana", Tagline: "Observability & dashboards", Category: "monitoring",
			IconColor: "#f46800", IconFG: "#ffffff", Version: "10.4", License: "AGPL-3.0",
			Publisher: "Grafana Labs", Rating: 4.9, Installs: "156k", Featured: true, Verified: true,
			Tags: []string{"Metrics", "Logs", "Alerting"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Visualize metrics, logs, and traces from any data source with gorgeous dashboards."},

		{Slug: "prometheus", Name: "Prometheus", Tagline: "Metrics & alerting toolkit", Category: "monitoring",
			IconColor: "#e6522c", IconFG: "#ffffff", Version: "2.51", License: "Apache-2.0",
			Publisher: "CNCF", Rating: 4.8, Installs: "124k", Verified: true,
			Tags: []string{"TSDB", "PromQL", "Scraping"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "40 GB",
			Description: "Pull-based metrics collection with PromQL query language and alert manager."},

		{Slug: "uptime-kuma", Name: "Uptime Kuma", Tagline: "Self-hosted uptime monitor", Category: "monitoring",
			IconColor: "#5cdd8b", IconFG: "#111111", Version: "1.23", License: "MIT",
			Publisher: "louislam", Rating: 4.9, Installs: "54.3k", Verified: true,
			Tags: []string{"Monitoring", "Status", "Alerts"}, MinRAM: "512 MB", MinCPU: "1 vCPU", MinDisk: "5 GB",
			Description: "Beautiful status pages and HTTP/TCP/Ping monitoring — all self-hosted."},

		{Slug: "portainer", Name: "Portainer", Tagline: "Container management UI", Category: "devops",
			IconColor: "#13bef9", IconFG: "#ffffff", Version: "2.19", License: "Zlib",
			Publisher: "Portainer.io", Rating: 4.7, Installs: "89.4k", Verified: true,
			Tags: []string{"Docker", "Kubernetes", "Swarm"}, MinRAM: "512 MB", MinCPU: "1 vCPU", MinDisk: "5 GB",
			Description: "Manage Docker, Kubernetes, Swarm and Nomad from a single pane of glass."},

		// ── Communication & Collaboration ────────────────────────
		{Slug: "mattermost", Name: "Mattermost", Tagline: "Self-hosted team messaging", Category: "communication",
			IconColor: "#1e325c", IconFG: "#ffffff", Version: "9.6", License: "MIT",
			Publisher: "Mattermost Inc", Rating: 4.5, Installs: "32.7k", Verified: true,
			Tags: []string{"Chat", "DM", "Slack-alt"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Open-source alternative to Slack — secure team collaboration for enterprises."},

		{Slug: "rocketchat", Name: "Rocket.Chat", Tagline: "Team collaboration platform", Category: "communication",
			IconColor: "#f5455c", IconFG: "#ffffff", Version: "6.7", License: "MIT",
			Publisher: "Rocket.Chat Tech", Rating: 4.4, Installs: "29.6k", Verified: true,
			Tags: []string{"Chat", "Omnichannel"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "15 GB",
			Description: "Extensible team chat with video, voice, file sharing and omnichannel support."},

		{Slug: "jitsi", Name: "Jitsi Meet", Tagline: "Secure video conferencing", Category: "communication",
			IconColor: "#1e6fff", IconFG: "#ffffff", Version: "2.0", License: "Apache-2.0",
			Publisher: "8x8", Rating: 4.6, Installs: "41.9k", Verified: true,
			Tags: []string{"Video", "WebRTC", "Meetings"}, MinRAM: "4 GB", MinCPU: "2 vCPU", MinDisk: "10 GB",
			Description: "Self-hosted video conferencing — end-to-end encryption, no account needed."},

		{Slug: "discourse", Name: "Discourse", Tagline: "Modern community forum", Category: "communication",
			IconColor: "#000000", IconFG: "#ffffff", Version: "3.2", License: "GPL-2.0",
			Publisher: "Discourse", Rating: 4.7, Installs: "23.8k", Verified: true,
			Tags: []string{"Forum", "Community", "Ruby"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Next-generation forum software — civilized discussion for your community."},

		// ── Storage & Files ──────────────────────────────────────
		{Slug: "nextcloud", Name: "Nextcloud", Tagline: "Self-hosted productivity suite", Category: "storage",
			IconColor: "#0082c9", IconFG: "#ffffff", Version: "28.0", License: "AGPL-3.0",
			Publisher: "Nextcloud GmbH", Rating: 4.6, Installs: "78.2k", Featured: true, Verified: true,
			Tags: []string{"Files", "Collaboration", "E2EE"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "40 GB",
			Description: "Files, calendar, contacts, video calls — your own private cloud."},

		{Slug: "minio", Name: "MinIO", Tagline: "High-performance object storage", Category: "storage",
			IconColor: "#c72c48", IconFG: "#ffffff", Version: "RELEASE.2024", License: "AGPL-3.0",
			Publisher: "MinIO Inc", Rating: 4.8, Installs: "58.6k", Verified: true,
			Tags: []string{"S3", "Object", "Storage"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "50 GB",
			Description: "S3-compatible object storage — high-performance, cloud-native."},

		{Slug: "seafile", Name: "Seafile", Tagline: "Enterprise file sync & share", Category: "storage",
			IconColor: "#ff9800", IconFG: "#ffffff", Version: "11.0", License: "AGPL-3.0",
			Publisher: "Seafile Ltd", Rating: 4.4, Installs: "18.9k", Verified: true,
			Tags: []string{"Sync", "Share", "E2EE"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "30 GB",
			Description: "Fast and reliable file sync with client-side encryption option."},

		// ── Analytics ────────────────────────────────────────────
		{Slug: "plausible", Name: "Plausible", Tagline: "Privacy-friendly analytics", Category: "analytics",
			IconColor: "#5850ec", IconFG: "#ffffff", Version: "2.1", License: "AGPL-3.0",
			Publisher: "Plausible Insights", Rating: 4.9, Installs: "37.4k", Featured: true, Verified: true,
			Tags: []string{"GDPR", "Cookieless", "Lightweight"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Simple, GDPR-compliant alternative to Google Analytics — no cookies."},

		{Slug: "matomo", Name: "Matomo", Tagline: "Open analytics platform", Category: "analytics",
			IconColor: "#3152a0", IconFG: "#ffffff", Version: "5.0", License: "GPL-3.0",
			Publisher: "Matomo", Rating: 4.5, Installs: "27.1k", Verified: true,
			Tags: []string{"PHP", "Heatmaps", "Funnels"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "15 GB",
			Description: "Google Analytics alternative — you own 100% of your data."},

		{Slug: "metabase", Name: "Metabase", Tagline: "Business intelligence for everyone", Category: "analytics",
			IconColor: "#509ee3", IconFG: "#ffffff", Version: "0.49", License: "AGPL-3.0",
			Publisher: "Metabase Inc", Rating: 4.7, Installs: "44.6k", Verified: true,
			Tags: []string{"BI", "Dashboards", "SQL"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "The easiest way to let everyone in your company ask questions and learn from data."},

		// ── Security ─────────────────────────────────────────────
		{Slug: "vaultwarden", Name: "Vaultwarden", Tagline: "Bitwarden-compatible password server", Category: "security",
			IconColor: "#175ddc", IconFG: "#ffffff", Version: "1.30", License: "GPL-3.0",
			Publisher: "dani-garcia", Rating: 4.9, Installs: "64.8k", Featured: true, Verified: true,
			Tags: []string{"Passwords", "Self-hosted", "Rust"}, MinRAM: "256 MB", MinCPU: "1 vCPU", MinDisk: "2 GB",
			Description: "Lightweight password manager server — Bitwarden API compatible."},

		{Slug: "authelia", Name: "Authelia", Tagline: "SSO & 2FA for your apps", Category: "security",
			IconColor: "#10439f", IconFG: "#ffffff", Version: "4.38", License: "Apache-2.0",
			Publisher: "Authelia", Rating: 4.6, Installs: "19.4k", Verified: true,
			Tags: []string{"SSO", "OIDC", "2FA"}, MinRAM: "512 MB", MinCPU: "1 vCPU", MinDisk: "2 GB",
			Description: "Self-hosted authentication portal — single sign-on and 2FA for your services."},

		{Slug: "wireguard", Name: "WireGuard", Tagline: "Fast modern VPN", Category: "security",
			IconColor: "#88171a", IconFG: "#ffffff", Version: "1.0", License: "GPL-2.0",
			Publisher: "Jason A. Donenfeld", Rating: 4.9, Installs: "103k", Verified: true,
			Tags: []string{"VPN", "Kernel", "Fast"}, MinRAM: "256 MB", MinCPU: "1 vCPU", MinDisk: "1 GB",
			Description: "Simple, fast, secure VPN using state-of-the-art cryptography."},

		// ── E-Commerce ───────────────────────────────────────────
		{Slug: "woocommerce", Name: "WooCommerce", Tagline: "Customizable e-commerce for WordPress", Category: "ecommerce",
			IconColor: "#7f54b3", IconFG: "#ffffff", Version: "8.7", License: "GPL-3.0",
			Publisher: "Automattic", Rating: 4.5, Installs: "112k", Verified: true,
			Tags: []string{"WordPress", "Store", "PHP"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Turn any WordPress site into a powerful online store — preinstalled and configured."},

		{Slug: "magento", Name: "Magento", Tagline: "Flexible commerce platform", Category: "ecommerce",
			IconColor: "#ee672f", IconFG: "#ffffff", Version: "2.4", License: "OSL-3.0",
			Publisher: "Adobe", Rating: 4.2, Installs: "24.5k", Verified: true,
			Tags: []string{"PHP", "Enterprise", "B2B"}, MinRAM: "4 GB", MinCPU: "2 vCPU", MinDisk: "40 GB",
			Description: "Enterprise-grade e-commerce platform for high-volume B2C and B2B stores."},

		// ── AI / ML ──────────────────────────────────────────────
		{Slug: "ollama", Name: "Ollama", Tagline: "Run LLMs locally", Category: "ai",
			IconColor: "#000000", IconFG: "#ffffff", Version: "0.1.32", License: "MIT",
			Publisher: "Ollama", Rating: 4.9, Installs: "142k", Featured: true, Verified: true,
			Tags: []string{"LLM", "GPU", "Llama3"}, MinRAM: "8 GB", MinCPU: "2 vCPU", MinDisk: "40 GB",
			Description: "Run Llama 3, Mistral, Gemma, and other LLMs locally with one command."},

		{Slug: "openwebui", Name: "Open WebUI", Tagline: "ChatGPT-like UI for local LLMs", Category: "ai",
			IconColor: "#10a37f", IconFG: "#ffffff", Version: "0.1", License: "MIT",
			Publisher: "Open WebUI", Rating: 4.8, Installs: "48.2k", Verified: true,
			Tags: []string{"ChatGPT-UI", "Ollama"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "5 GB",
			Description: "Beautiful web UI for Ollama and OpenAI API — multi-user, multi-model."},

		{Slug: "jupyterhub", Name: "JupyterHub", Tagline: "Multi-user Jupyter notebooks", Category: "ai",
			IconColor: "#f37626", IconFG: "#ffffff", Version: "4.1", License: "BSD-3",
			Publisher: "Jupyter", Rating: 4.7, Installs: "33.5k", Verified: true,
			Tags: []string{"Python", "Data Science", "Notebooks"}, MinRAM: "4 GB", MinCPU: "2 vCPU", MinDisk: "30 GB",
			Description: "Serve Jupyter notebooks to multiple users — the standard for teams & classrooms."},

		{Slug: "stable-diffusion", Name: "Stable Diffusion WebUI", Tagline: "Text-to-image generation", Category: "ai",
			IconColor: "#ff6b35", IconFG: "#ffffff", Version: "1.9", License: "AGPL-3.0",
			Publisher: "AUTOMATIC1111", Rating: 4.7, Installs: "78.3k", Verified: true,
			Tags: []string{"GPU", "Image gen", "LoRA"}, MinRAM: "16 GB", MinCPU: "4 vCPU", MinDisk: "60 GB",
			Description: "Industry-standard interface for Stable Diffusion — text-to-image with full control."},

		// ── Automation ───────────────────────────────────────────
		{Slug: "n8n", Name: "n8n", Tagline: "Workflow automation", Category: "automation",
			IconColor: "#ea4b71", IconFG: "#ffffff", Version: "1.39", License: "SustainUse",
			Publisher: "n8n.io", Rating: 4.8, Installs: "91.7k", Featured: true, Verified: true,
			Tags: []string{"Workflows", "Integration"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Zapier-style workflow automation — 400+ integrations, fair-code licensed."},

		{Slug: "nocodb", Name: "NocoDB", Tagline: "Airtable alternative", Category: "automation",
			IconColor: "#0088ff", IconFG: "#ffffff", Version: "0.207", License: "AGPL-3.0",
			Publisher: "NocoDB", Rating: 4.6, Installs: "36.2k", Verified: true,
			Tags: []string{"No-code", "DB-as-UI"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Turn any database into a smart spreadsheet — open-source Airtable."},

		{Slug: "appsmith", Name: "Appsmith", Tagline: "Build internal tools fast", Category: "automation",
			IconColor: "#2f71fe", IconFG: "#ffffff", Version: "1.19", License: "Apache-2.0",
			Publisher: "Appsmith", Rating: 4.5, Installs: "27.8k", Verified: true,
			Tags: []string{"Low-code", "Internal"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Drag-and-drop internal tool builder — connect any database or API."},

		// ── Dev Tools ────────────────────────────────────────────
		{Slug: "code-server", Name: "code-server", Tagline: "VS Code in the browser", Category: "devops",
			IconColor: "#007acc", IconFG: "#ffffff", Version: "4.22", License: "MIT",
			Publisher: "Coder", Rating: 4.8, Installs: "84.3k", Verified: true,
			Tags: []string{"IDE", "Remote", "VS Code"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "20 GB",
			Description: "Run VS Code on a remote server, accessible through the browser. Full extensions support."},

		{Slug: "gitpod", Name: "Gitpod", Tagline: "Dev environments in the cloud", Category: "devops",
			IconColor: "#ffae33", IconFG: "#000000", Version: "2024.04", License: "AGPL-3.0",
			Publisher: "Gitpod", Rating: 4.6, Installs: "14.9k", Verified: true,
			Tags: []string{"CDE", "Remote"}, MinRAM: "8 GB", MinCPU: "4 vCPU", MinDisk: "80 GB",
			Description: "Fully-featured cloud development environments — spin up a workspace in seconds."},

		// ── Forums / Social ──────────────────────────────────────
		{Slug: "mastodon", Name: "Mastodon", Tagline: "Decentralized social network", Category: "communication",
			IconColor: "#6364ff", IconFG: "#ffffff", Version: "4.2", License: "AGPL-3.0",
			Publisher: "Mastodon gGmbH", Rating: 4.5, Installs: "22.1k", Verified: true,
			Tags: []string{"Fediverse", "ActivityPub"}, MinRAM: "4 GB", MinCPU: "2 vCPU", MinDisk: "50 GB",
			Description: "Your own Mastodon instance — federated microblogging, part of the Fediverse."},

		{Slug: "mediawiki", Name: "MediaWiki", Tagline: "The software behind Wikipedia", Category: "cms",
			IconColor: "#006778", IconFG: "#ffffff", Version: "1.41", License: "GPL-2.0",
			Publisher: "Wikimedia", Rating: 4.5, Installs: "19.7k", Verified: true,
			Tags: []string{"Wiki", "PHP"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "15 GB",
			Description: "The collaborative encyclopedia software — scale from personal notes to Wikipedia."},

		{Slug: "bookstack", Name: "BookStack", Tagline: "Documentation & wiki platform", Category: "cms",
			IconColor: "#0288d1", IconFG: "#ffffff", Version: "24.02", License: "MIT",
			Publisher: "BookStack", Rating: 4.8, Installs: "21.4k", Verified: true,
			Tags: []string{"Docs", "Wiki", "PHP"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Beautiful, intuitive documentation platform organized into books, chapters, and pages."},

		// ── Game & Media ─────────────────────────────────────────
		{Slug: "plex", Name: "Plex Media Server", Tagline: "Stream your media anywhere", Category: "media",
			IconColor: "#e5a00d", IconFG: "#000000", Version: "1.40", License: "Proprietary",
			Publisher: "Plex Inc", Rating: 4.6, Installs: "89.5k", Verified: true,
			Tags: []string{"Media", "Streaming"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "100 GB",
			Description: "Organize and stream your personal media library to any device, anywhere."},

		{Slug: "jellyfin", Name: "Jellyfin", Tagline: "Free media streaming", Category: "media",
			IconColor: "#00a4dc", IconFG: "#ffffff", Version: "10.8", License: "GPL-2.0",
			Publisher: "Jellyfin", Rating: 4.8, Installs: "67.3k", Featured: true, Verified: true,
			Tags: []string{"Media", "Free", "No-account"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "100 GB",
			Description: "Free alternative to Plex and Emby — 100% FOSS media server."},

		// ── Email ────────────────────────────────────────────────
		{Slug: "mail-in-a-box", Name: "Mail-in-a-Box", Tagline: "Your own mail server", Category: "communication",
			IconColor: "#2c7ad2", IconFG: "#ffffff", Version: "65", License: "CC0-1.0",
			Publisher: "Josh Tauberer", Rating: 4.4, Installs: "15.8k", Verified: true,
			Tags: []string{"SMTP", "IMAP", "Webmail"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "A mail server in a box — the easy way to run your own domain's email."},
	}
}

func marketplaceCategories(apps []MarketplaceApp) []MarketplaceCategory {
	order := []MarketplaceCategory{
		{ID: "all", Label: "All apps"},
		{ID: "cms", Label: "CMS & Content"},
		{ID: "databases", Label: "Databases"},
		{ID: "devops", Label: "DevOps & Tools"},
		{ID: "monitoring", Label: "Monitoring"},
		{ID: "communication", Label: "Communication"},
		{ID: "storage", Label: "Storage & Files"},
		{ID: "analytics", Label: "Analytics & BI"},
		{ID: "security", Label: "Security"},
		{ID: "ecommerce", Label: "E-Commerce"},
		{ID: "ai", Label: "AI / ML"},
		{ID: "automation", Label: "Automation"},
		{ID: "media", Label: "Media"},
	}
	counts := map[string]int{"all": len(apps)}
	for _, a := range apps {
		counts[a.Category]++
	}
	for i := range order {
		order[i].Count = counts[order[i].ID]
	}
	return order
}

func marketplaceFeatured(apps []MarketplaceApp) []MarketplaceApp {
	out := make([]MarketplaceApp, 0, 4)
	for _, a := range apps {
		if a.Featured {
			out = append(out, a)
		}
	}
	return out
}

func workspaceMarketplaceHandler(w http.ResponseWriter, r *http.Request, d WorkspaceShellData) {
	d.ActiveNav = "marketplace"
	d.ActiveCat = ""
	d.ActiveSec = ""
	d.PageTitle = "Marketplace"
	d.PageSub = "One-click apps · Deploy in minutes"

	apps := marketplaceApps()

	data := MarketplaceData{
		WorkspaceShellData: d,
		Categories:         marketplaceCategories(apps),
		Apps:               apps,
		Featured:           marketplaceFeatured(apps),
	}

	tmpl, err := template.ParseFiles(
		filepath.Join("templates", "workspace", "base.html"),
		filepath.Join("templates", "workspace", "shell.html"),
		filepath.Join("templates", "workspace", "marketplace_content.html"),
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "workspace-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}
