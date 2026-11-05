package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// WizardApp is a one-click deployable app entry on the public /wizard page.
type WizardApp struct {
	Slug        string
	Name        string
	Tagline     string
	Category    string
	IconColor   string
	IconFG      string
	IconLetter  string
	Version     string
	License     string
	Publisher   string
	Rating      float64
	Installs    string
	Tags        []string
	Featured    bool
	Verified    bool
	MinRAM      string
	MinCPU      string
	MinDisk     string
	Description string
	Hero        string
}

type WizardCategory struct {
	ID    string
	Label string
	Icon  string
	Count int
}

type WizardData struct {
	Categories []WizardCategory
	Apps       []WizardApp
	Featured   []WizardApp
	TotalApps  int
}

func wizardApps() []WizardApp {
	return []WizardApp{
		{Slug: "wordpress", Name: "WordPress", Tagline: "The world's most popular CMS", Category: "cms",
			IconColor: "#21759b", IconFG: "#fff", IconLetter: "W", Version: "6.5", License: "GPL-2.0",
			Publisher: "WordPress Foundation", Rating: 4.8, Installs: "284k", Featured: true, Verified: true,
			Tags: []string{"PHP", "MySQL", "Blog"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Build any kind of website — blog, portfolio, storefront. Comes preconfigured with LAMP stack, automatic SSL, and daily backups.",
			Hero:        "Over 43% of the web runs on WordPress. Launch a production-ready site in under 60 seconds."},

		{Slug: "nextcloud", Name: "Nextcloud", Tagline: "Self-hosted productivity suite", Category: "productivity",
			IconColor: "#0082c9", IconFG: "#fff", IconLetter: "N", Version: "28.0", License: "AGPL-3.0",
			Publisher: "Nextcloud GmbH", Rating: 4.6, Installs: "78.2k", Featured: true, Verified: true,
			Tags: []string{"Files", "Collaboration", "E2EE"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "40 GB",
			Description: "Files, calendar, contacts, video calls, collaborative documents — your own private cloud with end-to-end encryption.",
			Hero:        "The complete open-source Dropbox + Google Workspace alternative. Own your data."},

		{Slug: "ghost", Name: "Ghost", Tagline: "Modern publishing platform", Category: "cms",
			IconColor: "#15171a", IconFG: "#fff", IconLetter: "G", Version: "5.80", License: "MIT",
			Publisher: "Ghost Foundation", Rating: 4.7, Installs: "41.2k", Featured: true, Verified: true,
			Tags: []string{"Node.js", "Newsletter"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "5 GB",
			Description: "Professional publishing for creators — built-in memberships, paid newsletters, and lightning-fast SEO.",
			Hero:        "Turn your audience into a business — newsletters, subscriptions, and a beautiful reading experience."},

		{Slug: "gitlab", Name: "GitLab CE", Tagline: "The complete DevOps platform", Category: "devops",
			IconColor: "#fc6d26", IconFG: "#fff", IconLetter: "GL", Version: "16.10", License: "MIT",
			Publisher: "GitLab Inc", Rating: 4.5, Installs: "62.1k", Featured: true, Verified: true,
			Tags: []string{"Git", "CI/CD", "Registry"}, MinRAM: "4 GB", MinCPU: "2 vCPU", MinDisk: "50 GB",
			Description: "Self-hosted Git repo, CI/CD runners, container registry, issue tracking, wiki — the entire DevOps lifecycle in one app.",
			Hero:        "Replace GitHub, Jira, Jenkins, and Docker Hub — all with one open-source install."},

		{Slug: "grafana", Name: "Grafana", Tagline: "Observability & dashboards", Category: "monitoring",
			IconColor: "#f46800", IconFG: "#fff", IconLetter: "G", Version: "10.4", License: "AGPL-3.0",
			Publisher: "Grafana Labs", Rating: 4.9, Installs: "156k", Featured: true, Verified: true,
			Tags: []string{"Metrics", "Alerts"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Visualize metrics, logs, and traces from any data source. 150+ data source plugins, unified alerting, and gorgeous dashboards.",
			Hero:        "The industry-standard observability stack — beautiful out of the box."},

		{Slug: "ollama", Name: "Ollama", Tagline: "Run LLMs locally", Category: "ai",
			IconColor: "#000000", IconFG: "#fff", IconLetter: "O", Version: "0.1.32", License: "MIT",
			Publisher: "Ollama", Rating: 4.9, Installs: "142k", Featured: true, Verified: true,
			Tags: []string{"LLM", "GPU", "Llama3"}, MinRAM: "8 GB", MinCPU: "2 vCPU", MinDisk: "40 GB",
			Description: "Run Llama 3, Mistral, Gemma, Phi-3, and hundreds of other LLMs locally with a single command. GPU acceleration included.",
			Hero:        "Private, offline AI on your own hardware — no API keys, no data leaving your cloud."},

		{Slug: "postgresql", Name: "PostgreSQL", Tagline: "Advanced relational database", Category: "databases",
			IconColor: "#336791", IconFG: "#fff", IconLetter: "P", Version: "16.2", License: "PostgreSQL",
			Publisher: "PostgreSQL GDG", Rating: 4.9, Installs: "512k", Featured: true, Verified: true,
			Tags: []string{"SQL", "ACID", "JSON"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "The world's most advanced open source relational database — extensible, standards-compliant, battle-tested.",
			Hero:        "35+ years of active development. The database used by Apple, Reddit, Instagram."},

		{Slug: "redis", Name: "Redis", Tagline: "In-memory data structure store", Category: "databases",
			IconColor: "#dc382d", IconFG: "#fff", IconLetter: "R", Version: "7.2", License: "BSD-3",
			Publisher: "Redis Ltd", Rating: 4.9, Installs: "441k", Featured: true, Verified: true,
			Tags: []string{"Cache", "Pub/Sub"}, MinRAM: "512 MB", MinCPU: "1 vCPU", MinDisk: "5 GB",
			Description: "Lightning-fast in-memory store for caching, session storage, queues, real-time analytics, and more.",
			Hero:        "Sub-millisecond responses. The world's favorite cache and message broker."},

		{Slug: "plausible", Name: "Plausible", Tagline: "Privacy-friendly analytics", Category: "analytics",
			IconColor: "#5850ec", IconFG: "#fff", IconLetter: "P", Version: "2.1", License: "AGPL-3.0",
			Publisher: "Plausible Insights", Rating: 4.9, Installs: "37.4k", Featured: true, Verified: true,
			Tags: []string{"GDPR", "Cookieless"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Simple, GDPR-compliant alternative to Google Analytics — no cookies, no personal data tracked.",
			Hero:        "A lightweight 1 KB script, no cookie banners, and you own 100% of the data."},

		{Slug: "vaultwarden", Name: "Vaultwarden", Tagline: "Bitwarden-compatible password server", Category: "security",
			IconColor: "#175ddc", IconFG: "#fff", IconLetter: "V", Version: "1.30", License: "GPL-3.0",
			Publisher: "dani-garcia", Rating: 4.9, Installs: "64.8k", Featured: true, Verified: true,
			Tags: []string{"Passwords", "Rust"}, MinRAM: "256 MB", MinCPU: "1 vCPU", MinDisk: "2 GB",
			Description: "Lightweight password manager server — fully Bitwarden API compatible, works with all official clients.",
			Hero:        "The same security, a fraction of the resources. Runs on a Raspberry Pi."},

		{Slug: "n8n", Name: "n8n", Tagline: "Workflow automation", Category: "automation",
			IconColor: "#ea4b71", IconFG: "#fff", IconLetter: "n", Version: "1.39", License: "SustainUse",
			Publisher: "n8n.io", Rating: 4.8, Installs: "91.7k", Featured: true, Verified: true,
			Tags: []string{"Workflows"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Zapier-style workflow automation with 400+ integrations. Fair-code licensed — self-host at no cost.",
			Hero:        "Automate anything. Connect 400+ apps with visual workflows, or drop in custom JavaScript."},

		{Slug: "jellyfin", Name: "Jellyfin", Tagline: "Free media streaming", Category: "media",
			IconColor: "#00a4dc", IconFG: "#fff", IconLetter: "J", Version: "10.8", License: "GPL-2.0",
			Publisher: "Jellyfin", Rating: 4.8, Installs: "67.3k", Featured: true, Verified: true,
			Tags: []string{"Media", "No-account"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "100 GB",
			Description: "Free alternative to Plex and Emby — 100% FOSS media server. No accounts, no subscriptions, no tracking.",
			Hero:        "Your movies, your music, your shows — streamed to any device without cloud lock-in."},

		// Non-featured apps
		{Slug: "strapi", Name: "Strapi", Tagline: "Headless CMS for developers", Category: "cms",
			IconColor: "#4945ff", IconFG: "#fff", IconLetter: "S", Version: "4.25", License: "MIT",
			Publisher: "Strapi Solutions", Rating: 4.6, Installs: "28.9k", Verified: true,
			Tags: []string{"GraphQL", "REST"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "8 GB",
			Description: "Headless CMS with auto-generated REST and GraphQL APIs.",
			Hero:        "Design your content model in minutes, ship APIs in seconds."},

		{Slug: "gitea", Name: "Gitea", Tagline: "Lightweight Git service", Category: "devops",
			IconColor: "#609926", IconFG: "#fff", IconLetter: "G", Version: "1.21", License: "MIT",
			Publisher: "Gitea Ltd", Rating: 4.7, Installs: "38.5k", Verified: true,
			Tags: []string{"Git", "Self-hosted"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Painless self-hosted Git — a lightweight alternative to GitLab.",
			Hero:        "Runs on almost anything — 100 MB RAM minimum."},

		{Slug: "mongodb", Name: "MongoDB", Tagline: "Document database", Category: "databases",
			IconColor: "#00684a", IconFG: "#fff", IconLetter: "M", Version: "7.0", License: "SSPL",
			Publisher: "MongoDB Inc", Rating: 4.6, Installs: "198k", Verified: true,
			Tags: []string{"NoSQL", "JSON"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "General-purpose document database. Rich query language with flexible schema.",
			Hero:        "Ship fast with flexible schemas that evolve with your product."},

		{Slug: "mysql", Name: "MySQL", Tagline: "The world's most popular RDBMS", Category: "databases",
			IconColor: "#00758f", IconFG: "#fff", IconLetter: "M", Version: "8.3", License: "GPL-2.0",
			Publisher: "Oracle", Rating: 4.5, Installs: "389k", Verified: true,
			Tags: []string{"SQL", "Replication"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Widely adopted SQL database, tuned defaults and automatic backups enabled.",
			Hero:        "Battle-tested reliability trusted by Facebook, Twitter, YouTube, and WordPress."},

		{Slug: "mariadb", Name: "MariaDB", Tagline: "Community MySQL fork", Category: "databases",
			IconColor: "#003545", IconFG: "#fff", IconLetter: "M", Version: "11.3", License: "GPL-2.0",
			Publisher: "MariaDB Foundation", Rating: 4.6, Installs: "87.2k", Verified: true,
			Tags: []string{"SQL"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Drop-in replacement for MySQL with additional storage engines.",
			Hero:        "A truly open-source fork — run by the original MySQL developers."},

		{Slug: "clickhouse", Name: "ClickHouse", Tagline: "Columnar OLAP database", Category: "databases",
			IconColor: "#ffcc01", IconFG: "#000", IconLetter: "C", Version: "24.3", License: "Apache-2.0",
			Publisher: "ClickHouse Inc", Rating: 4.8, Installs: "24.7k", Verified: true,
			Tags: []string{"Analytics"}, MinRAM: "4 GB", MinCPU: "2 vCPU", MinDisk: "40 GB",
			Description: "Blazing-fast columnar database for real-time analytics at scale.",
			Hero:        "Crunch billions of rows per second on a single machine."},

		{Slug: "prometheus", Name: "Prometheus", Tagline: "Metrics & alerting", Category: "monitoring",
			IconColor: "#e6522c", IconFG: "#fff", IconLetter: "P", Version: "2.51", License: "Apache-2.0",
			Publisher: "CNCF", Rating: 4.8, Installs: "124k", Verified: true,
			Tags: []string{"TSDB", "PromQL"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "40 GB",
			Description: "Pull-based metrics collection with PromQL query language and alert manager.",
			Hero:        "The Kubernetes-native metrics standard. Open, powerful, pull-based."},

		{Slug: "uptime-kuma", Name: "Uptime Kuma", Tagline: "Self-hosted uptime monitor", Category: "monitoring",
			IconColor: "#5cdd8b", IconFG: "#111", IconLetter: "U", Version: "1.23", License: "MIT",
			Publisher: "louislam", Rating: 4.9, Installs: "54.3k", Verified: true,
			Tags: []string{"Status", "Alerts"}, MinRAM: "512 MB", MinCPU: "1 vCPU", MinDisk: "5 GB",
			Description: "Beautiful status pages and HTTP/TCP/Ping monitoring — all self-hosted.",
			Hero:        "The prettiest status page you can self-host. 90+ notification integrations."},

		{Slug: "portainer", Name: "Portainer", Tagline: "Container management UI", Category: "devops",
			IconColor: "#13bef9", IconFG: "#fff", IconLetter: "P", Version: "2.19", License: "Zlib",
			Publisher: "Portainer.io", Rating: 4.7, Installs: "89.4k", Verified: true,
			Tags: []string{"Docker", "Kubernetes"}, MinRAM: "512 MB", MinCPU: "1 vCPU", MinDisk: "5 GB",
			Description: "Manage Docker, Kubernetes, Swarm and Nomad from a single pane of glass.",
			Hero:        "Orchestrate containers without the CLI — point, click, deploy."},

		{Slug: "mattermost", Name: "Mattermost", Tagline: "Self-hosted team messaging", Category: "communication",
			IconColor: "#1e325c", IconFG: "#fff", IconLetter: "M", Version: "9.6", License: "MIT",
			Publisher: "Mattermost Inc", Rating: 4.5, Installs: "32.7k", Verified: true,
			Tags: []string{"Chat"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Open-source alternative to Slack — secure team collaboration for enterprises.",
			Hero:        "The Slack your CISO will actually approve."},

		{Slug: "rocketchat", Name: "Rocket.Chat", Tagline: "Team collaboration", Category: "communication",
			IconColor: "#f5455c", IconFG: "#fff", IconLetter: "R", Version: "6.7", License: "MIT",
			Publisher: "Rocket.Chat Tech", Rating: 4.4, Installs: "29.6k", Verified: true,
			Tags: []string{"Chat", "Omnichannel"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "15 GB",
			Description: "Extensible team chat with video, voice, file sharing and omnichannel support.",
			Hero:        "Chat + live chat widget for your website + omnichannel inbox, unified."},

		{Slug: "jitsi", Name: "Jitsi Meet", Tagline: "Secure video conferencing", Category: "communication",
			IconColor: "#1e6fff", IconFG: "#fff", IconLetter: "J", Version: "2.0", License: "Apache-2.0",
			Publisher: "8x8", Rating: 4.6, Installs: "41.9k", Verified: true,
			Tags: []string{"Video", "WebRTC"}, MinRAM: "4 GB", MinCPU: "2 vCPU", MinDisk: "10 GB",
			Description: "Self-hosted video conferencing — end-to-end encryption, no account needed.",
			Hero:        "Share a link, start a call. No downloads, no accounts, no limits."},

		{Slug: "discourse", Name: "Discourse", Tagline: "Modern community forum", Category: "communication",
			IconColor: "#000000", IconFG: "#fff", IconLetter: "D", Version: "3.2", License: "GPL-2.0",
			Publisher: "Discourse", Rating: 4.7, Installs: "23.8k", Verified: true,
			Tags: []string{"Forum"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Next-generation forum software — civilized discussion for your community.",
			Hero:        "Built by the co-founder of Stack Overflow. The forum software of the decade."},

		{Slug: "minio", Name: "MinIO", Tagline: "S3-compatible object storage", Category: "productivity",
			IconColor: "#c72c48", IconFG: "#fff", IconLetter: "M", Version: "RELEASE", License: "AGPL-3.0",
			Publisher: "MinIO Inc", Rating: 4.8, Installs: "58.6k", Verified: true,
			Tags: []string{"S3", "Object"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "50 GB",
			Description: "High-performance object storage compatible with the S3 API.",
			Hero:        "Your own S3 — often faster than AWS. Petabyte-scale, single-binary."},

		{Slug: "seafile", Name: "Seafile", Tagline: "File sync & share", Category: "productivity",
			IconColor: "#ff9800", IconFG: "#fff", IconLetter: "S", Version: "11.0", License: "AGPL-3.0",
			Publisher: "Seafile Ltd", Rating: 4.4, Installs: "18.9k", Verified: true,
			Tags: []string{"Sync"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "30 GB",
			Description: "Fast and reliable file sync with client-side encryption option.",
			Hero:        "The fastest self-hosted Dropbox clone. Block-level deduplication."},

		{Slug: "matomo", Name: "Matomo", Tagline: "Open analytics platform", Category: "analytics",
			IconColor: "#3152a0", IconFG: "#fff", IconLetter: "M", Version: "5.0", License: "GPL-3.0",
			Publisher: "Matomo", Rating: 4.5, Installs: "27.1k", Verified: true,
			Tags: []string{"Heatmaps"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "15 GB",
			Description: "Google Analytics alternative — you own 100% of your data.",
			Hero:        "Everything GA4 does, plus heatmaps, session recordings, and A/B testing."},

		{Slug: "metabase", Name: "Metabase", Tagline: "BI for everyone", Category: "analytics",
			IconColor: "#509ee3", IconFG: "#fff", IconLetter: "M", Version: "0.49", License: "AGPL-3.0",
			Publisher: "Metabase Inc", Rating: 4.7, Installs: "44.6k", Verified: true,
			Tags: []string{"BI", "SQL"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "The easiest way to let everyone in your company ask questions and learn from data.",
			Hero:        "Self-serve BI without having to teach the whole company SQL."},

		{Slug: "authelia", Name: "Authelia", Tagline: "SSO & 2FA portal", Category: "security",
			IconColor: "#10439f", IconFG: "#fff", IconLetter: "A", Version: "4.38", License: "Apache-2.0",
			Publisher: "Authelia", Rating: 4.6, Installs: "19.4k", Verified: true,
			Tags: []string{"SSO", "2FA"}, MinRAM: "512 MB", MinCPU: "1 vCPU", MinDisk: "2 GB",
			Description: "Self-hosted authentication portal — single sign-on and 2FA for your services.",
			Hero:        "Put every internal app behind SSO — no more password fatigue."},

		{Slug: "wireguard", Name: "WireGuard", Tagline: "Fast modern VPN", Category: "security",
			IconColor: "#88171a", IconFG: "#fff", IconLetter: "W", Version: "1.0", License: "GPL-2.0",
			Publisher: "Jason A. Donenfeld", Rating: 4.9, Installs: "103k", Verified: true,
			Tags: []string{"VPN"}, MinRAM: "256 MB", MinCPU: "1 vCPU", MinDisk: "1 GB",
			Description: "Simple, fast, secure VPN using state-of-the-art cryptography.",
			Hero:        "A 4000-line codebase vs. OpenVPN's 100,000 — and it's faster."},

		{Slug: "woocommerce", Name: "WooCommerce", Tagline: "E-commerce for WordPress", Category: "ecommerce",
			IconColor: "#7f54b3", IconFG: "#fff", IconLetter: "W", Version: "8.7", License: "GPL-3.0",
			Publisher: "Automattic", Rating: 4.5, Installs: "112k", Verified: true,
			Tags: []string{"WordPress", "Store"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "20 GB",
			Description: "Turn any WordPress site into a powerful online store.",
			Hero:        "Powers 28% of all online stores. Extensible and completely customizable."},

		{Slug: "medusa", Name: "Medusa", Tagline: "Open-source Shopify", Category: "ecommerce",
			IconColor: "#000000", IconFG: "#fff", IconLetter: "M", Version: "1.20", License: "MIT",
			Publisher: "Medusa", Rating: 4.5, Installs: "11.3k", Verified: true,
			Tags: []string{"Commerce"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "10 GB",
			Description: "Headless commerce engine — fully customizable digital storefronts.",
			Hero:        "Commerce primitives for developers who demand flexibility."},

		{Slug: "openwebui", Name: "Open WebUI", Tagline: "ChatGPT-like UI for LLMs", Category: "ai",
			IconColor: "#10a37f", IconFG: "#fff", IconLetter: "O", Version: "0.1", License: "MIT",
			Publisher: "Open WebUI", Rating: 4.8, Installs: "48.2k", Verified: true,
			Tags: []string{"Ollama"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "5 GB",
			Description: "Beautiful web UI for Ollama and OpenAI API — multi-user, multi-model.",
			Hero:        "Your own private ChatGPT — with RAG, vision, and team workspaces."},

		{Slug: "jupyterhub", Name: "JupyterHub", Tagline: "Multi-user Jupyter", Category: "ai",
			IconColor: "#f37626", IconFG: "#fff", IconLetter: "J", Version: "4.1", License: "BSD-3",
			Publisher: "Jupyter", Rating: 4.7, Installs: "33.5k", Verified: true,
			Tags: []string{"Python"}, MinRAM: "4 GB", MinCPU: "2 vCPU", MinDisk: "30 GB",
			Description: "Serve Jupyter notebooks to multiple users — the standard for teams & classrooms.",
			Hero:        "The backbone of data science education and research worldwide."},

		{Slug: "stable-diffusion", Name: "Stable Diffusion", Tagline: "Text-to-image AI", Category: "ai",
			IconColor: "#ff6b35", IconFG: "#fff", IconLetter: "S", Version: "1.9", License: "AGPL-3.0",
			Publisher: "AUTOMATIC1111", Rating: 4.7, Installs: "78.3k", Verified: true,
			Tags: []string{"GPU", "LoRA"}, MinRAM: "16 GB", MinCPU: "4 vCPU", MinDisk: "60 GB",
			Description: "Industry-standard interface for Stable Diffusion — text-to-image with full control.",
			Hero:        "From prompt to 4K art in seconds — on your own GPU."},

		{Slug: "nocodb", Name: "NocoDB", Tagline: "Airtable alternative", Category: "automation",
			IconColor: "#0088ff", IconFG: "#fff", IconLetter: "N", Version: "0.207", License: "AGPL-3.0",
			Publisher: "NocoDB", Rating: 4.6, Installs: "36.2k", Verified: true,
			Tags: []string{"No-code"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Turn any database into a smart spreadsheet — open-source Airtable.",
			Hero:        "Give anyone in your company a UI on top of your production database."},

		{Slug: "appsmith", Name: "Appsmith", Tagline: "Internal tools fast", Category: "automation",
			IconColor: "#2f71fe", IconFG: "#fff", IconLetter: "A", Version: "1.19", License: "Apache-2.0",
			Publisher: "Appsmith", Rating: 4.5, Installs: "27.8k", Verified: true,
			Tags: []string{"Low-code"}, MinRAM: "2 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Drag-and-drop internal tool builder — connect any database or API.",
			Hero:        "Build admin panels in an afternoon, not a sprint."},

		{Slug: "code-server", Name: "code-server", Tagline: "VS Code in the browser", Category: "devops",
			IconColor: "#007acc", IconFG: "#fff", IconLetter: "C", Version: "4.22", License: "MIT",
			Publisher: "Coder", Rating: 4.8, Installs: "84.3k", Verified: true,
			Tags: []string{"IDE", "Remote"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "20 GB",
			Description: "VS Code on a remote server, accessible through the browser. Full extension support.",
			Hero:        "Code from your iPad. Persistent dev environments that never sleep."},

		{Slug: "jenkins", Name: "Jenkins", Tagline: "The leading automation server", Category: "devops",
			IconColor: "#d33833", IconFG: "#fff", IconLetter: "J", Version: "2.440", License: "MIT",
			Publisher: "CD Foundation", Rating: 4.3, Installs: "72.8k", Verified: true,
			Tags: []string{"CI/CD"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "30 GB",
			Description: "Build, test, and deploy with 1800+ plugins. Industry-standard CI server.",
			Hero:        "Still the most flexible CI/CD server ever built."},

		{Slug: "mastodon", Name: "Mastodon", Tagline: "Decentralized social", Category: "communication",
			IconColor: "#6364ff", IconFG: "#fff", IconLetter: "M", Version: "4.2", License: "AGPL-3.0",
			Publisher: "Mastodon gGmbH", Rating: 4.5, Installs: "22.1k", Verified: true,
			Tags: []string{"Fediverse"}, MinRAM: "4 GB", MinCPU: "2 vCPU", MinDisk: "50 GB",
			Description: "Your own Mastodon instance — federated microblogging, part of the Fediverse.",
			Hero:        "Your server, your rules. No algorithmic feed, no ads, no lock-in."},

		{Slug: "bookstack", Name: "BookStack", Tagline: "Documentation wiki", Category: "cms",
			IconColor: "#0288d1", IconFG: "#fff", IconLetter: "B", Version: "24.02", License: "MIT",
			Publisher: "BookStack", Rating: 4.8, Installs: "21.4k", Verified: true,
			Tags: []string{"Docs", "Wiki"}, MinRAM: "1 GB", MinCPU: "1 vCPU", MinDisk: "10 GB",
			Description: "Beautiful, intuitive documentation platform organized into books, chapters, and pages.",
			Hero:        "A wiki that actually looks good. Bring your docs out of the intranet."},

		{Slug: "plex", Name: "Plex Media Server", Tagline: "Stream your media", Category: "media",
			IconColor: "#e5a00d", IconFG: "#000", IconLetter: "P", Version: "1.40", License: "Proprietary",
			Publisher: "Plex Inc", Rating: 4.6, Installs: "89.5k", Verified: true,
			Tags: []string{"Media"}, MinRAM: "2 GB", MinCPU: "2 vCPU", MinDisk: "100 GB",
			Description: "Organize and stream your personal media library to any device, anywhere.",
			Hero:        "The polished commercial option — slick apps on every platform imaginable."},
	}
}

func wizardCategories(apps []WizardApp) []WizardCategory {
	order := []WizardCategory{
		{ID: "all", Label: "All apps", Icon: "grid"},
		{ID: "cms", Label: "CMS & Content", Icon: "doc"},
		{ID: "databases", Label: "Databases", Icon: "db"},
		{ID: "devops", Label: "DevOps & Tools", Icon: "code"},
		{ID: "monitoring", Label: "Monitoring", Icon: "chart"},
		{ID: "communication", Label: "Communication", Icon: "chat"},
		{ID: "productivity", Label: "Productivity", Icon: "folder"},
		{ID: "analytics", Label: "Analytics", Icon: "stats"},
		{ID: "security", Label: "Security", Icon: "shield"},
		{ID: "ecommerce", Label: "E-Commerce", Icon: "cart"},
		{ID: "ai", Label: "AI / ML", Icon: "spark"},
		{ID: "automation", Label: "Automation", Icon: "zap"},
		{ID: "media", Label: "Media", Icon: "play"},
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

func wizardFeatured(apps []WizardApp) []WizardApp {
	out := make([]WizardApp, 0, 12)
	for _, a := range apps {
		if a.Featured {
			out = append(out, a)
		}
	}
	return out
}

// WizardHandler renders the fancy standalone marketplace at /wizard.
func WizardHandler(w http.ResponseWriter, r *http.Request) {
	apps := wizardApps()
	data := WizardData{
		Categories: wizardCategories(apps),
		Apps:       apps,
		Featured:   wizardFeatured(apps),
		TotalApps:  len(apps),
	}

	tmpl, err := template.ParseFiles(filepath.Join("templates", "wizard.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// WizardAppHandler renders an individual app page at /wizard/{slug}.
func WizardAppHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var found *WizardApp
	for i, a := range wizardApps() {
		if a.Slug == slug {
			apps := wizardApps()
			found = &apps[i]
			break
		}
	}
	if found == nil {
		ServeError(w, r, http.StatusNotFound, "App not found.")
		return
	}
	tmpl, err := template.ParseFiles(filepath.Join("templates", "wizard_app.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, found); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func RegisterWizardRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /wizard", WizardHandler)
	mux.HandleFunc("GET /wizard/", WizardHandler)
	mux.HandleFunc("GET /wizard/{slug}", WizardAppHandler)
}
