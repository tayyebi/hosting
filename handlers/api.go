package handlers

import (
	"encoding/json"
	"net/http"
)

// ── JSON response helpers ────────────────────────────────────

// apiOK writes a 200 JSON response.
func apiOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// apiCreated writes a 201 JSON response.
func apiCreated(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

// apiError writes a JSON error response.
func apiError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{
		"error":   http.StatusText(code),
		"message": message,
		"code":    code,
	})
}

// apiValidationError writes a 422 JSON response with per-field error details.
// fields maps field names to human-readable error strings.
func apiValidationError(w http.ResponseWriter, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error":   "Validation Error",
		"message": "One or more fields failed validation.",
		"code":    http.StatusUnprocessableEntity,
		"errors":  fields,
	})
}

// ── Swagger / API documentation endpoint ─────────────────────

type swaggerPath struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Tag         string `json:"tag"`
}

func apiDocsHandler(w http.ResponseWriter, r *http.Request) {
	doc := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Platform Mock API",
			"version":     "1.0.0",
			"description": "Mock REST API for the Platform UI. All data is synthetic.",
		},
		"servers": []map[string]string{
			{"url": "/api", "description": "Local mock server"},
		},
		"endpoints": []swaggerPath{
			{Method: "POST", Path: "/api/login", Summary: "Authenticate", Description: "Authenticate with email and password. Returns a session token.", Tag: "Authentication"},
			{Method: "POST", Path: "/api/logout", Summary: "Logout", Description: "Destroy the current session.", Tag: "Authentication"},
			{Method: "GET", Path: "/api/status", Summary: "Platform status", Description: "Returns public platform health, component status, metrics, and recent incidents.", Tag: "Public"},
			{Method: "GET", Path: "/api/sso-guide", Summary: "SSO guide", Description: "Returns SSO integration documentation and provider configuration details.", Tag: "Public"},
			{Method: "GET", Path: "/api/sysadmin/dashboard", Summary: "Sysadmin dashboard", Description: "Returns sysadmin overview: tenant count, user count, navigation, and health.", Tag: "Sysadmin"},
			{Method: "GET", Path: "/api/sysadmin/tenants", Summary: "List tenants", Description: "Returns all tenants with owner information.", Tag: "Sysadmin"},
			{Method: "POST", Path: "/api/sysadmin/tenants", Summary: "Create tenant", Description: "Create a new tenant with name, slug, and optional domain.", Tag: "Sysadmin"},
			{Method: "GET", Path: "/api/sysadmin/tenants/{id}", Summary: "Get tenant", Description: "Returns a single tenant by ID.", Tag: "Sysadmin"},
			{Method: "PATCH", Path: "/api/sysadmin/tenants/{id}", Summary: "Update tenant", Description: "Update a tenant's name or domain. Slug is immutable.", Tag: "Sysadmin"},
			{Method: "DELETE", Path: "/api/sysadmin/tenants/{id}", Summary: "Delete tenant", Description: "Delete a tenant by ID.", Tag: "Sysadmin"},
			{Method: "GET", Path: "/api/sysadmin/logs", Summary: "System logs", Description: "Returns system log entries with histogram buckets. Supports level, service, and search filters.", Tag: "Sysadmin"},
			{Method: "GET", Path: "/api/sysadmin/jobs", Summary: "Background jobs", Description: "Returns background jobs, queue health, workers, and KPIs. Supports status, queue, and search filters.", Tag: "Sysadmin"},
			{Method: "GET", Path: "/api/sysadmin/database", Summary: "Database schema", Description: "Returns database schemas, tables, and columns.", Tag: "Sysadmin"},
			{Method: "GET", Path: "/api/sysadmin/database/table", Summary: "Browse table", Description: "Returns table structure and row data. Requires ?name=schema.table query parameter.", Tag: "Sysadmin"},
			{Method: "POST", Path: "/api/sysadmin/database/query", Summary: "Run SQL query", Description: "Execute a mock SQL query and return results.", Tag: "Sysadmin"},
			{Method: "GET", Path: "/api/sysadmin/replicas", Summary: "List replicas", Description: "Returns all replicas with health status and merchant code.", Tag: "Sysadmin"},
			{Method: "GET", Path: "/api/sysadmin/accounting", Summary: "Accounting system", Description: "Returns billing KPIs, API logs, webhooks, and connection status.", Tag: "Sysadmin"},
			{Method: "GET", Path: "/api/sysadmin/current-replica", Summary: "Current replica", Description: "Returns the current primary replica identity, network, resources, sync, certificates, and failover config.", Tag: "Sysadmin"},
			{Method: "GET", Path: "/api/sysadmin/sso", Summary: "SSO configuration", Description: "Returns SSO general settings, OAuth2, OIDC, SAML, clients, tokens, and security policies.", Tag: "Sysadmin"},
			{Method: "GET", Path: "/api/user/settings", Summary: "User settings", Description: "Returns the authenticated user's profile, preferences, and security settings.", Tag: "User"},
			{Method: "GET", Path: "/api/user/workspaces", Summary: "User workspaces", Description: "Returns the authenticated user's workspaces, teams, invitations, and permissions.", Tag: "User"},
			{Method: "GET", Path: "/api/wizard", Summary: "Wizard app catalog", Description: "Returns the full wizard marketplace: categories, apps, and featured apps.", Tag: "Wizard"},
			{Method: "GET", Path: "/api/wizard/{slug}", Summary: "Wizard app detail", Description: "Returns a single wizard app by slug.", Tag: "Wizard"},
			{Method: "GET", Path: "/api/tenant/{slug}/dashboard", Summary: "Tenant dashboard", Description: "Returns tenant overview: KPIs, navigation, and health status.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/clients", Summary: "List clients", Description: "Returns all clients for a tenant. Supports q and status query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/clients", Summary: "Create client", Description: "Create a new client in a tenant with name, email, and optional company.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/clients/{id}", Summary: "Get client", Description: "Returns a single client by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/clients/{id}", Summary: "Delete client", Description: "Delete a client by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/employees", Summary: "List employees", Description: "Returns all employees for a tenant. Supports q, role, and status query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/employees", Summary: "Create employee", Description: "Create a new employee in a tenant with name, email, and optional role.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/employees/{id}", Summary: "Get employee", Description: "Returns a single employee by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/employees/{id}", Summary: "Delete employee", Description: "Delete an employee by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/tickets", Summary: "List tickets", Description: "Returns all tickets for a tenant. Supports q, status, and priority query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/tickets", Summary: "Create ticket", Description: "Create a new ticket in a tenant with subject, client, and optional priority.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/tickets/{id}", Summary: "Get ticket", Description: "Returns a single ticket by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/tickets/{id}", Summary: "Delete ticket", Description: "Delete a ticket by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/abuse-reports", Summary: "List abuse reports", Description: "Returns all abuse reports for a tenant. Supports q, status, and type query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/abuse-reports", Summary: "Create abuse report", Description: "Create a new abuse report in a tenant with type, source, target, and summary.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/abuse-reports/{id}", Summary: "Get abuse report", Description: "Returns a single abuse report by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/abuse-reports/{id}", Summary: "Delete abuse report", Description: "Delete an abuse report by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/refund-complaints", Summary: "List refund complaints", Description: "Returns all refund complaints for a tenant. Supports q and status query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/refund-complaints", Summary: "Create refund complaint", Description: "Create a new refund complaint with client, amount, currency, reason, and invoice ID.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/refund-complaints/{id}", Summary: "Get refund complaint", Description: "Returns a single refund complaint by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/refund-complaints/{id}", Summary: "Delete refund complaint", Description: "Delete a refund complaint by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/one-time-access", Summary: "List one-time access tokens", Description: "Returns all one-time access tokens for a tenant. Supports q and status query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/one-time-access", Summary: "Create one-time access token", Description: "Create a new one-time access token with label, target, granted_to, and expires_at.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/one-time-access/{id}", Summary: "Get one-time access token", Description: "Returns a single one-time access token by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/one-time-access/{id}", Summary: "Delete one-time access token", Description: "Delete a one-time access token by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/resources", Summary: "List resources", Description: "Returns all infrastructure resources for a tenant. Supports q, status, and type query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/resources", Summary: "Create resource", Description: "Create a new resource with name, type, and optional region.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/resources/{id}", Summary: "Get resource", Description: "Returns a single resource by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/resources/{id}", Summary: "Delete resource", Description: "Delete a resource by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/monitoring", Summary: "List monitoring checks", Description: "Returns all monitoring checks for a tenant. Supports q, status, and type query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/monitoring", Summary: "Create monitoring check", Description: "Create a new monitoring check with name, endpoint, type, interval, and region.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/monitoring/{id}", Summary: "Get monitoring check", Description: "Returns a single monitoring check by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/monitoring/{id}", Summary: "Delete monitoring check", Description: "Delete a monitoring check by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/ids", Summary: "List IDS events", Description: "Returns all intrusion detection events for a tenant. Supports q, severity, and status query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/ids", Summary: "Create IDS event", Description: "Create a new IDS event with event_type, severity, source_ip, target, and rule_id.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/ids/{id}", Summary: "Get IDS event", Description: "Returns a single IDS event by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/ids/{id}", Summary: "Delete IDS event", Description: "Delete an IDS event by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/incidents", Summary: "List incidents", Description: "Returns all incidents for a tenant. Supports q, severity, and status query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/incidents", Summary: "Create incident", Description: "Declare a new incident with title, severity, commander, and affected_services.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/incidents/{id}", Summary: "Get incident", Description: "Returns a single incident by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/incidents/{id}", Summary: "Delete incident", Description: "Delete an incident by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/status-page", Summary: "List status page components", Description: "Returns all status page components for a tenant. Supports q, status, and group query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/status-page", Summary: "Create status page component", Description: "Create a new status page component with name, group, and optional status.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/status-page/{id}", Summary: "Get status page component", Description: "Returns a single status page component by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/status-page/{id}", Summary: "Delete status page component", Description: "Delete a status page component by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/maintenance", Summary: "List maintenance windows", Description: "Returns all maintenance windows for a tenant. Supports q, status, and type query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/maintenance", Summary: "Create maintenance window", Description: "Schedule a new maintenance window with title, type, services, scheduled_at, and duration.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/maintenance/{id}", Summary: "Get maintenance window", Description: "Returns a single maintenance window by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/maintenance/{id}", Summary: "Delete maintenance window", Description: "Delete a maintenance window by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/transactions", Summary: "List transactions", Description: "Returns all transactions for a tenant. Supports q, type, and status query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/transactions", Summary: "Create transaction", Description: "Record a new transaction with type, amount, currency, description, client, and method.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/transactions/{id}", Summary: "Get transaction", Description: "Returns a single transaction by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/transactions/{id}", Summary: "Delete transaction", Description: "Delete a transaction by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/pricing", Summary: "List pricing plans", Description: "Returns all pricing plans for a tenant. Supports q, tier, and status query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/pricing", Summary: "Create pricing plan", Description: "Create a new pricing plan with name, tier, price, currency, interval, features, and max_users.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/pricing/{id}", Summary: "Get pricing plan", Description: "Returns a single pricing plan by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/pricing/{id}", Summary: "Delete pricing plan", Description: "Delete a pricing plan by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/tax", Summary: "List tax rules", Description: "Returns all tax rules for a tenant. Supports q, type, and status query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/tax", Summary: "Create tax rule", Description: "Create a new tax rule with name, region, type, rate, applies_to, valid_from, and valid_until.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/tax/{id}", Summary: "Get tax rule", Description: "Returns a single tax rule by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/tax/{id}", Summary: "Delete tax rule", Description: "Delete a tax rule by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/sla", Summary: "List SLA policies", Description: "Returns all SLA policies for a tenant. Supports q and status query filters.", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/sla", Summary: "Create SLA policy", Description: "Create a new SLA policy with name, service, uptime_target, response_time, credit_policy, penalty_rate, and review_period.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/sla/{id}", Summary: "Get SLA policy", Description: "Returns a single SLA policy by ID within a tenant.", Tag: "Tenant"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/sla/{id}", Summary: "Delete SLA policy", Description: "Delete an SLA policy by ID within a tenant.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/brand", Summary: "Get brand settings", Description: "Returns brand identity, colours, social links, developer portal config, DNS entries, and email templates.", Tag: "Tenant"},
			{Method: "PUT", Path: "/api/tenant/{slug}/brand", Summary: "Update brand settings", Description: "Update brand identity fields (company name, tagline, website, support email).", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/localization", Summary: "Get localization settings", Description: "Returns regional, format, currency, API locale settings, and CDN locale routing rules.", Tag: "Tenant"},
			{Method: "PUT", Path: "/api/tenant/{slug}/localization", Summary: "Update localization settings", Description: "Update regional, format, or currency localization settings.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/registration", Summary: "Get registration settings", Description: "Returns registration general settings, rules, form fields, verification, notifications, KPIs, and funnel.", Tag: "Tenant"},
			{Method: "PUT", Path: "/api/tenant/{slug}/registration", Summary: "Update registration settings", Description: "Update registration general settings (enabled, access mode, redirect, default role).", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/register", Summary: "Get register form config", Description: "Returns the public registration form configuration for a tenant (fields, captcha, TOS/privacy URLs).", Tag: "Tenant"},
			{Method: "POST", Path: "/api/tenant/{slug}/register", Summary: "Submit registration", Description: "Submit a new user registration with email, password, and optional full name.", Tag: "Tenant"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}", Summary: "Workspace overview", Description: "Returns workspace details, owning client, sibling workspaces, and the full navigation tree.", Tag: "Workspace"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/general", Summary: "List general VMs", Description: "Returns all general-purpose VMs for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/general", Summary: "Create general VM", Description: "Provision a new general-purpose VM with name, type, and region.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/general/{vmid}", Summary: "Get general VM", Description: "Returns a single general-purpose VM by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/general/{vmid}", Summary: "Delete general VM", Description: "Terminate a general-purpose VM by ID.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/gpu", Summary: "List GPU VMs", Description: "Returns all GPU / accelerated VMs for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/gpu", Summary: "Create GPU VM", Description: "Provision a new GPU VM with name, gpu, type, and region.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/gpu/{vmid}", Summary: "Get GPU VM", Description: "Returns a single GPU VM by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/gpu/{vmid}", Summary: "Delete GPU VM", Description: "Terminate a GPU VM by ID.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/spot", Summary: "List spot VMs", Description: "Returns all spot / preemptible VMs for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/spot", Summary: "Create spot VM", Description: "Request a new spot VM with name, pool, bid, and region.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/spot/{vmid}", Summary: "Get spot VM", Description: "Returns a single spot VM by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/vms/spot/{vmid}", Summary: "Delete spot VM", Description: "Terminate a spot VM by ID.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/kubernetes", Summary: "List Kubernetes clusters", Description: "Returns all managed Kubernetes clusters for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/kubernetes", Summary: "Create Kubernetes cluster", Description: "Provision a new managed Kubernetes cluster with name, version, nodes, and region.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/kubernetes/{cid}", Summary: "Get Kubernetes cluster", Description: "Returns a single Kubernetes cluster by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/kubernetes/{cid}", Summary: "Delete Kubernetes cluster", Description: "Delete a Kubernetes cluster by ID.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/registry", Summary: "List container repos", Description: "Returns all container registry repositories for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/registry", Summary: "Create container repo", Description: "Create a new container repository with a name.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/registry/{rid}", Summary: "Get container repo", Description: "Returns a single container repository by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/registry/{rid}", Summary: "Delete container repo", Description: "Delete a container repository by ID.", Tag: "Workspace — Compute"},
			// Workspace — Compute: Container Instances
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/instances", Summary: "List container instances", Description: "Returns all container instances for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/instances", Summary: "Create container instance", Description: "Launch a new container instance with name, image, cpus, memory, and region.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/instances/{ciid}", Summary: "Get container instance", Description: "Returns a single container instance by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/containers/instances/{ciid}", Summary: "Delete container instance", Description: "Terminate a container instance by ID.", Tag: "Workspace — Compute"},
			// Workspace — Compute: Web App Hosting
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/hosting/web", Summary: "List web apps", Description: "Returns all web apps for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/hosting/web", Summary: "Create web app", Description: "Deploy a new web app with name, runtime, domain, and region.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/hosting/web/{waid}", Summary: "Get web app", Description: "Returns a single web app by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/hosting/web/{waid}", Summary: "Delete web app", Description: "Delete a web app by ID.", Tag: "Workspace — Compute"},
			// Workspace — Compute: Static Site Hosting
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/hosting/static", Summary: "List static sites", Description: "Returns all static sites for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/hosting/static", Summary: "Create static site", Description: "Deploy a new static site with name, framework, domain, and region.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/hosting/static/{ssid}", Summary: "Get static site", Description: "Returns a single static site by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/hosting/static/{ssid}", Summary: "Delete static site", Description: "Delete a static site by ID.", Tag: "Workspace — Compute"},
			// Workspace — Compute: Serverless Functions
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/serverless/functions", Summary: "List serverless functions", Description: "Returns all serverless functions for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/serverless/functions", Summary: "Create serverless function", Description: "Create a new serverless function with name, runtime, trigger, region, and memory.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/serverless/functions/{fid}", Summary: "Get serverless function", Description: "Returns a single serverless function by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/serverless/functions/{fid}", Summary: "Delete serverless function", Description: "Delete a serverless function by ID.", Tag: "Workspace — Compute"},
			// Workspace — Compute: Edge Functions
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/serverless/edge", Summary: "List edge functions", Description: "Returns all edge functions for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/serverless/edge", Summary: "Create edge function", Description: "Deploy a new edge function with name, route, and regions.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/serverless/edge/{eid}", Summary: "Get edge function", Description: "Returns a single edge function by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/serverless/edge/{eid}", Summary: "Delete edge function", Description: "Delete an edge function by ID.", Tag: "Workspace — Compute"},
			// Workspace — Compute: Batch Jobs
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/batch/jobs", Summary: "List batch jobs", Description: "Returns all scheduled batch jobs for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/batch/jobs", Summary: "Create batch job", Description: "Schedule a new batch job with name, queue, priority, schedule, and region.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/batch/jobs/{jid}", Summary: "Get batch job", Description: "Returns a single batch job by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/batch/jobs/{jid}", Summary: "Delete batch job", Description: "Delete a batch job by ID.", Tag: "Workspace — Compute"},
			// Workspace — Compute: HPC / Grid Computing
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/batch/hpc", Summary: "List HPC jobs", Description: "Returns all HPC / grid computing jobs for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/batch/hpc", Summary: "Create HPC job", Description: "Submit a new HPC job with name, cluster, nodes, cpus_per_node, partition, region, and wall_time.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/batch/hpc/{hid}", Summary: "Get HPC job", Description: "Returns a single HPC job by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/batch/hpc/{hid}", Summary: "Delete HPC job", Description: "Delete an HPC job by ID.", Tag: "Workspace — Compute"},
			// Workspace — Storage: Object Standard
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/standard", Summary: "List standard buckets", Description: "Returns all standard-tier object storage buckets for a workspace. Supports q and status query filters.", Tag: "Workspace — Storage"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/standard", Summary: "Create standard bucket", Description: "Create a new standard object storage bucket with name, region, versioning, and access.", Tag: "Workspace — Storage"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/standard/{bid}", Summary: "Get standard bucket", Description: "Returns a single standard bucket by ID.", Tag: "Workspace — Storage"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/standard/{bid}", Summary: "Delete standard bucket", Description: "Delete a standard bucket by ID.", Tag: "Workspace — Storage"},
			// Workspace — Storage: Object Archive
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/archive", Summary: "List archive vaults", Description: "Returns all archive / cold-tier storage vaults for a workspace. Supports q and status query filters.", Tag: "Workspace — Storage"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/archive", Summary: "Create archive vault", Description: "Create a new archive vault with name, region, and tier (glacier, deep-archive, glacier-ir).", Tag: "Workspace — Storage"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/archive/{vid}", Summary: "Get archive vault", Description: "Returns a single archive vault by ID.", Tag: "Workspace — Storage"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/archive/{vid}", Summary: "Delete archive vault", Description: "Delete an archive vault by ID.", Tag: "Workspace — Storage"},
			// Workspace — Storage: Data Lake
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/datalake", Summary: "List data lake zones", Description: "Returns all data lake zones for a workspace. Supports q and status query filters.", Tag: "Workspace — Storage"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/datalake", Summary: "Create data lake zone", Description: "Create a new data lake zone with name, region, format, and catalog.", Tag: "Workspace — Storage"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/datalake/{zid}", Summary: "Get data lake zone", Description: "Returns a single data lake zone by ID.", Tag: "Workspace — Storage"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/storage/object/datalake/{zid}", Summary: "Delete data lake zone", Description: "Delete a data lake zone by ID.", Tag: "Workspace — Storage"},
			// Workspace — Storage: Block SSD
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/block/ssd", Summary: "List SSD volumes", Description: "Returns all SSD high-performance block volumes for a workspace. Supports q and status query filters.", Tag: "Workspace — Storage"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/storage/block/ssd", Summary: "Create SSD volume", Description: "Create a new SSD block volume with name, region, size, iops, and encrypted flag.", Tag: "Workspace — Storage"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/block/ssd/{vid}", Summary: "Get SSD volume", Description: "Returns a single SSD volume by ID.", Tag: "Workspace — Storage"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/storage/block/ssd/{vid}", Summary: "Delete SSD volume", Description: "Delete an SSD volume by ID.", Tag: "Workspace — Storage"},
			// Workspace — Storage: Block HDD
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/block/hdd", Summary: "List HDD volumes", Description: "Returns all HDD standard block volumes for a workspace. Supports q and status query filters.", Tag: "Workspace — Storage"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/storage/block/hdd", Summary: "Create HDD volume", Description: "Create a new HDD block volume with name, region, size, and encrypted flag.", Tag: "Workspace — Storage"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/block/hdd/{vid}", Summary: "Get HDD volume", Description: "Returns a single HDD volume by ID.", Tag: "Workspace — Storage"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/storage/block/hdd/{vid}", Summary: "Delete HDD volume", Description: "Delete an HDD volume by ID.", Tag: "Workspace — Storage"},
			// Workspace — Storage: File NFS
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/file/nfs", Summary: "List NFS shares", Description: "Returns all NFS / SMB file shares for a workspace. Supports q and status query filters.", Tag: "Workspace — Storage"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/storage/file/nfs", Summary: "Create NFS share", Description: "Create a new NFS/SMB share with name, region, protocol, size, and access_cidr.", Tag: "Workspace — Storage"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/file/nfs/{sid}", Summary: "Get NFS share", Description: "Returns a single NFS share by ID.", Tag: "Workspace — Storage"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/storage/file/nfs/{sid}", Summary: "Delete NFS share", Description: "Delete an NFS share by ID.", Tag: "Workspace — Storage"},
			// Workspace — Storage: File Distributed
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/file/distributed", Summary: "List distributed filesystems", Description: "Returns all distributed file systems for a workspace. Supports q and status query filters.", Tag: "Workspace — Storage"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/storage/file/distributed", Summary: "Create distributed filesystem", Description: "Create a new distributed filesystem with name, region, type, capacity, nodes, and replication.", Tag: "Workspace — Storage"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/file/distributed/{did}", Summary: "Get distributed filesystem", Description: "Returns a single distributed filesystem by ID.", Tag: "Workspace — Storage"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/storage/file/distributed/{did}", Summary: "Delete distributed filesystem", Description: "Delete a distributed filesystem by ID.", Tag: "Workspace — Storage"},
			// Workspace — Storage: Backup Snapshots
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/backup/snapshots", Summary: "List snapshots", Description: "Returns all snapshots for a workspace. Supports q and status query filters.", Tag: "Workspace — Storage"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/storage/backup/snapshots", Summary: "Create snapshot", Description: "Create a new snapshot with name, region, source_type, source_name, retention, and encrypted flag.", Tag: "Workspace — Storage"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/backup/snapshots/{snid}", Summary: "Get snapshot", Description: "Returns a single snapshot by ID.", Tag: "Workspace — Storage"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/storage/backup/snapshots/{snid}", Summary: "Delete snapshot", Description: "Delete a snapshot by ID.", Tag: "Workspace — Storage"},
			// Workspace — Databases: NoSQL Document
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/document", Summary: "List document databases", Description: "Returns all document databases for a workspace. Supports q and status query filters.", Tag: "Workspace — Databases"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/document", Summary: "Create document database", Description: "Create a new document database with name, engine, version, region, size, and encrypted flag.", Tag: "Workspace — Databases"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/document/{dbid}", Summary: "Get document database", Description: "Returns a single document database by ID.", Tag: "Workspace — Databases"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/document/{dbid}", Summary: "Delete document database", Description: "Delete a document database by ID.", Tag: "Workspace — Databases"},
			// Workspace — Databases: NoSQL Key-Value
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/keyvalue", Summary: "List key-value stores", Description: "Returns all key-value stores for a workspace. Supports q and status query filters.", Tag: "Workspace — Databases"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/keyvalue", Summary: "Create key-value store", Description: "Create a new key-value store with name, engine, version, region, size, max_memory, and encrypted flag.", Tag: "Workspace — Databases"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/keyvalue/{dbid}", Summary: "Get key-value store", Description: "Returns a single key-value store by ID.", Tag: "Workspace — Databases"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/keyvalue/{dbid}", Summary: "Delete key-value store", Description: "Delete a key-value store by ID.", Tag: "Workspace — Databases"},
			// Workspace — Databases: NoSQL Wide-Column
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/wide-column", Summary: "List wide-column databases", Description: "Returns all wide-column databases for a workspace. Supports q and status query filters.", Tag: "Workspace — Databases"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/wide-column", Summary: "Create wide-column database", Description: "Create a new wide-column database with name, engine, version, region, nodes, and encrypted flag.", Tag: "Workspace — Databases"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/wide-column/{dbid}", Summary: "Get wide-column database", Description: "Returns a single wide-column database by ID.", Tag: "Workspace — Databases"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/wide-column/{dbid}", Summary: "Delete wide-column database", Description: "Delete a wide-column database by ID.", Tag: "Workspace — Databases"},
			// Workspace — Databases: NoSQL Graph
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/graph", Summary: "List graph databases", Description: "Returns all graph databases for a workspace. Supports q and status query filters.", Tag: "Workspace — Databases"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/graph", Summary: "Create graph database", Description: "Create a new graph database with name, engine, version, region, size, and encrypted flag.", Tag: "Workspace — Databases"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/graph/{dbid}", Summary: "Get graph database", Description: "Returns a single graph database by ID.", Tag: "Workspace — Databases"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/graph/{dbid}", Summary: "Delete graph database", Description: "Delete a graph database by ID.", Tag: "Workspace — Databases"},
			// Workspace — Databases: Memory Cache
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/memory/cache", Summary: "List cache instances", Description: "Returns all in-memory cache instances for a workspace. Supports q and status query filters.", Tag: "Workspace — Databases"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/databases/memory/cache", Summary: "Create cache instance", Description: "Create a new cache instance with name, engine, version, region, size, nodes, max_memory, eviction, and encrypted flag.", Tag: "Workspace — Databases"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/memory/cache/{cid}", Summary: "Get cache instance", Description: "Returns a single cache instance by ID.", Tag: "Workspace — Databases"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/databases/memory/cache/{cid}", Summary: "Delete cache instance", Description: "Delete a cache instance by ID.", Tag: "Workspace — Databases"},
			// Workspace — Databases: Memory Session Store
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/memory/session", Summary: "List session stores", Description: "Returns all in-memory session stores for a workspace. Supports q and status query filters.", Tag: "Workspace — Databases"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/databases/memory/session", Summary: "Create session store", Description: "Create a new session store with name, engine, version, region, size, nodes, max_memory, ttl, and encrypted flag.", Tag: "Workspace — Databases"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/memory/session/{sid}", Summary: "Get session store", Description: "Returns a single session store by ID.", Tag: "Workspace — Databases"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/databases/memory/session/{sid}", Summary: "Delete session store", Description: "Delete a session store by ID.", Tag: "Workspace — Databases"},
			// Workspace — Databases: Analytics Data Warehouse
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/warehouse", Summary: "List data warehouses", Description: "Returns all data warehouses for a workspace. Supports q and status query filters.", Tag: "Workspace — Databases"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/warehouse", Summary: "Create data warehouse", Description: "Create a new data warehouse with name, engine, version, region, size, and encrypted flag.", Tag: "Workspace — Databases"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/warehouse/{dwid}", Summary: "Get data warehouse", Description: "Returns a single data warehouse by ID.", Tag: "Workspace — Databases"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/warehouse/{dwid}", Summary: "Delete data warehouse", Description: "Delete a data warehouse by ID.", Tag: "Workspace — Databases"},
			// Workspace — Databases: Analytics Time-Series
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/timeseries", Summary: "List time-series databases", Description: "Returns all time-series databases for a workspace. Supports q and status query filters.", Tag: "Workspace — Databases"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/timeseries", Summary: "Create time-series database", Description: "Create a new time-series database with name, engine, version, region, size, retention, and encrypted flag.", Tag: "Workspace — Databases"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/timeseries/{tsid}", Summary: "Get time-series database", Description: "Returns a single time-series database by ID.", Tag: "Workspace — Databases"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/timeseries/{tsid}", Summary: "Delete time-series database", Description: "Delete a time-series database by ID.", Tag: "Workspace — Databases"},
			// Workspace — Databases: Analytics Search & Full-text
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/search", Summary: "List search engines", Description: "Returns all search & full-text engines for a workspace. Supports q and status query filters.", Tag: "Workspace — Databases"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/search", Summary: "Create search engine", Description: "Create a new search engine with name, engine, version, region, size, nodes, and encrypted flag.", Tag: "Workspace — Databases"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/search/{seid}", Summary: "Get search engine", Description: "Returns a single search engine by ID.", Tag: "Workspace — Databases"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/databases/analytics/search/{seid}", Summary: "Delete search engine", Description: "Delete a search engine by ID.", Tag: "Workspace — Databases"},
			// Workspace — Networking: VPC / Subnets
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/vpc", Summary: "List VPCs", Description: "Returns all VPCs / subnets for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/vpc", Summary: "Create VPC", Description: "Create a new VPC with name, cidr, region, dns, and encrypted flag.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/vpc/{vid}", Summary: "Get VPC", Description: "Returns a single VPC by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/vpc/{vid}", Summary: "Delete VPC", Description: "Delete a VPC by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: Peering & Transit
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/peering", Summary: "List peerings", Description: "Returns all VPC peerings & transit connections for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/peering", Summary: "Create peering", Description: "Create a new VPC peering with name, local_vpc, remote_vpc, remote_owner, direction, bandwidth, region, and encrypted flag.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/peering/{pid}", Summary: "Get peering", Description: "Returns a single VPC peering by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/peering/{pid}", Summary: "Delete peering", Description: "Delete a VPC peering by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: Network Security Groups
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/nsg", Summary: "List NSGs", Description: "Returns all network security groups for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/nsg", Summary: "Create NSG", Description: "Create a new network security group with name, vpc, region, default_action, and logging flag.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/nsg/{nid}", Summary: "Get NSG", Description: "Returns a single network security group by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/vnet/nsg/{nid}", Summary: "Delete NSG", Description: "Delete a network security group by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: Application Load Balancer
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/app", Summary: "List app load balancers", Description: "Returns all application (L7) load balancers for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/app", Summary: "Create app load balancer", Description: "Create a new application load balancer with name, scheme, protocol, port, region, health_check, ssl_arn, and waf flag.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/app/{lid}", Summary: "Get app load balancer", Description: "Returns a single application load balancer by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/app/{lid}", Summary: "Delete app load balancer", Description: "Delete an application load balancer by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: Network Load Balancer
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/network", Summary: "List network load balancers", Description: "Returns all network (L4) load balancers for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/network", Summary: "Create network load balancer", Description: "Create a new network load balancer with name, scheme, protocol, port, region, and cross_zone flag.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/network/{lid}", Summary: "Get network load balancer", Description: "Returns a single network load balancer by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/network/{lid}", Summary: "Delete network load balancer", Description: "Delete a network load balancer by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: Global Load Balancing
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/global", Summary: "List global load balancers", Description: "Returns all global load balancers for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/global", Summary: "Create global load balancer", Description: "Create a new global load balancer with name, protocol, routing, failover, and health_check.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/global/{gid}", Summary: "Get global load balancer", Description: "Returns a single global load balancer by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/lb/global/{gid}", Summary: "Delete global load balancer", Description: "Delete a global load balancer by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: DNS Management
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/manage", Summary: "List DNS zones", Description: "Returns all DNS zones for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/manage", Summary: "Create DNS zone", Description: "Create a new DNS zone with name, type, region, ttl, and dnssec.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/manage/{zid}", Summary: "Get DNS zone", Description: "Returns a single DNS zone by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/manage/{zid}", Summary: "Delete DNS zone", Description: "Delete a DNS zone by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: Domain Registration
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/domains", Summary: "List domains", Description: "Returns all registered domains for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/domains", Summary: "Register domain", Description: "Register a new domain with domain name, auto_renew, and privacy.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/domains/{did}", Summary: "Get domain", Description: "Returns a single domain registration by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/domains/{did}", Summary: "Delete domain", Description: "Delete a domain registration by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: CDN Static Asset Delivery
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/static", Summary: "List CDN distributions", Description: "Returns all static asset CDN distributions for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/static", Summary: "Create CDN distribution", Description: "Create a new CDN distribution with name, domain, origin, region, cache_policy, and ssl.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/static/{did}", Summary: "Get CDN distribution", Description: "Returns a single CDN distribution by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/static/{did}", Summary: "Delete CDN distribution", Description: "Delete a CDN distribution by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: CDN Video Streaming
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/video", Summary: "List video streams", Description: "Returns all video streaming endpoints for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/video", Summary: "Create video stream", Description: "Create a new video streaming endpoint with name, domain, origin, region, protocol, transcoding, and drm.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/video/{vid}", Summary: "Get video stream", Description: "Returns a single video streaming endpoint by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/video/{vid}", Summary: "Delete video stream", Description: "Delete a video streaming endpoint by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: CDN Edge Caching
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/edge", Summary: "List edge caches", Description: "Returns all edge caching configurations for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/edge", Summary: "Create edge cache", Description: "Create a new edge cache config with name, region, and waf.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/edge/{eid}", Summary: "Get edge cache", Description: "Returns a single edge cache config by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/cdn/edge/{eid}", Summary: "Delete edge cache", Description: "Delete an edge cache config by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: VPN Gateway
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/vpn", Summary: "List VPN gateways", Description: "Returns all VPN gateways for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/vpn", Summary: "Create VPN gateway", Description: "Create a new VPN gateway with name, type, region, remote_ip, encryption, and ike_version.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/vpn/{gid}", Summary: "Get VPN gateway", Description: "Returns a single VPN gateway by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/vpn/{gid}", Summary: "Delete VPN gateway", Description: "Delete a VPN gateway by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: Direct Connect
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/direct", Summary: "List direct connections", Description: "Returns all direct connections for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/direct", Summary: "Create direct connection", Description: "Create a new direct connection with name, region, location, port, bandwidth, bgp_asn, redundancy, and macsec.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/direct/{cid}", Summary: "Get direct connection", Description: "Returns a single direct connection by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/direct/{cid}", Summary: "Delete direct connection", Description: "Delete a direct connection by ID.", Tag: "Workspace — Networking"},
			// Workspace — Networking: SD-WAN
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/sdwan", Summary: "List SD-WAN networks", Description: "Returns all SD-WAN networks for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/sdwan", Summary: "Create SD-WAN network", Description: "Create a new SD-WAN network with name, topology, protocol, encryption, region, and qos.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/sdwan/{nid}", Summary: "Get SD-WAN network", Description: "Returns a single SD-WAN network by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/private/sdwan/{nid}", Summary: "Delete SD-WAN network", Description: "Delete an SD-WAN network by ID.", Tag: "Workspace — Networking"},
			// Workspace — AI: Distributed Training
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/distributed", Summary: "List distributed training jobs", Description: "Returns all distributed training jobs for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/distributed", Summary: "Create distributed training job", Description: "Create a new distributed training job with name, framework, nodes, gpus_per_node, gpu_type, dataset, epochs, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/distributed/{jid}", Summary: "Get distributed training job", Description: "Returns a single distributed training job by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/distributed/{jid}", Summary: "Delete distributed training job", Description: "Delete a distributed training job by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: AutoML
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/automl", Summary: "List AutoML runs", Description: "Returns all AutoML runs for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/automl", Summary: "Create AutoML run", Description: "Create a new AutoML run with name, task_type, dataset, target, trials, metric, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/automl/{rid}", Summary: "Get AutoML run", Description: "Returns a single AutoML run by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/automl/{rid}", Summary: "Delete AutoML run", Description: "Delete an AutoML run by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Fine-tuning
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/finetuning", Summary: "List fine-tuning jobs", Description: "Returns all fine-tuning jobs for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/finetuning", Summary: "Create fine-tuning job", Description: "Create a new fine-tuning job with name, base_model, method, dataset, epochs, learning_rate, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/finetuning/{ftid}", Summary: "Get fine-tuning job", Description: "Returns a single fine-tuning job by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/training/finetuning/{ftid}", Summary: "Delete fine-tuning job", Description: "Delete a fine-tuning job by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Inference Endpoints
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/inference", Summary: "List inference endpoints", Description: "Returns all inference endpoints for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/inference", Summary: "Create inference endpoint", Description: "Create a new inference endpoint with name, model, runtime, instance_type, replicas, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/inference/{eid}", Summary: "Get inference endpoint", Description: "Returns a single inference endpoint by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/inference/{eid}", Summary: "Delete inference endpoint", Description: "Delete an inference endpoint by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Batch Inference
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/batch", Summary: "List batch inference jobs", Description: "Returns all batch inference jobs for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/batch", Summary: "Create batch inference job", Description: "Create a new batch inference job with name, model, input_path, output_path, instance_type, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/batch/{bid}", Summary: "Get batch inference job", Description: "Returns a single batch inference job by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/batch/{bid}", Summary: "Delete batch inference job", Description: "Delete a batch inference job by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Edge Deployment
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/edge", Summary: "List edge deployments", Description: "Returns all edge deployments for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/edge", Summary: "Create edge deployment", Description: "Create a new edge deployment with name, model, runtime, target_device, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/edge/{did}", Summary: "Get edge deployment", Description: "Returns a single edge deployment by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/hosting/edge/{did}", Summary: "Delete edge deployment", Description: "Delete an edge deployment by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Vision API
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/vision", Summary: "List vision APIs", Description: "Returns all vision APIs for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/vision", Summary: "Create vision API", Description: "Create a new vision API with name, model, capability, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/vision/{aid}", Summary: "Get vision API", Description: "Returns a single vision API by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/vision/{aid}", Summary: "Delete vision API", Description: "Delete a vision API by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Language / NLP API
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/language", Summary: "List language APIs", Description: "Returns all language/NLP APIs for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/language", Summary: "Create language API", Description: "Create a new language API with name, model, capability, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/language/{aid}", Summary: "Get language API", Description: "Returns a single language API by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/language/{aid}", Summary: "Delete language API", Description: "Delete a language API by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Speech API
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/speech", Summary: "List speech APIs", Description: "Returns all speech APIs for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/speech", Summary: "Create speech API", Description: "Create a new speech API with name, model, capability, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/speech/{aid}", Summary: "Get speech API", Description: "Returns a single speech API by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/speech/{aid}", Summary: "Delete speech API", Description: "Delete a speech API by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Recommendation API
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/recommendation", Summary: "List recommendation APIs", Description: "Returns all recommendation APIs for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/recommendation", Summary: "Create recommendation API", Description: "Create a new recommendation API with name, model, algorithm, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/recommendation/{aid}", Summary: "Get recommendation API", Description: "Returns a single recommendation API by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/apis/recommendation/{aid}", Summary: "Delete recommendation API", Description: "Delete a recommendation API by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: ETL / ELT Pipelines
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/etl", Summary: "List ETL pipelines", Description: "Returns all ETL/ELT pipelines for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/etl", Summary: "Create ETL pipeline", Description: "Create a new ETL/ELT pipeline with name, type, source, destination, schedule, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/etl/{pid}", Summary: "Get ETL pipeline", Description: "Returns a single ETL/ELT pipeline by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/etl/{pid}", Summary: "Delete ETL pipeline", Description: "Delete an ETL/ELT pipeline by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Streaming Ingestion
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/streaming", Summary: "List streaming ingestions", Description: "Returns all streaming ingestion streams for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/streaming", Summary: "Create streaming ingestion", Description: "Create a new streaming ingestion with name, engine, source, destination, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/streaming/{sid}", Summary: "Get streaming ingestion", Description: "Returns a single streaming ingestion by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/streaming/{sid}", Summary: "Delete streaming ingestion", Description: "Delete a streaming ingestion by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Data Catalog
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/catalog", Summary: "List catalog assets", Description: "Returns all data catalog assets for a workspace. Supports q and pii query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/catalog", Summary: "Create catalog asset", Description: "Create a new catalog asset with name, source, owner, type, pii, tags, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/catalog/{aid}", Summary: "Get catalog asset", Description: "Returns a single catalog asset by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/data/catalog/{aid}", Summary: "Delete catalog asset", Description: "Delete a catalog asset by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Experiment Tracking
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/experiments", Summary: "List experiments", Description: "Returns all ML experiments for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/experiments", Summary: "Create experiment", Description: "Create a new ML experiment with name, framework, owner, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/experiments/{eid}", Summary: "Get experiment", Description: "Returns a single ML experiment by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/experiments/{eid}", Summary: "Delete experiment", Description: "Delete an ML experiment by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Model Registry
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/registry", Summary: "List registered models", Description: "Returns all registered models for a workspace. Supports q and stage query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/registry", Summary: "Register model", Description: "Register a new model with name, version, format, owner, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/registry/{mid}", Summary: "Get registered model", Description: "Returns a single registered model by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/registry/{mid}", Summary: "Delete registered model", Description: "Delete a registered model by ID.", Tag: "Workspace — AI"},
			// Workspace — AI: Monitoring & Drift Detection
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/monitoring", Summary: "List ML monitors", Description: "Returns all ML monitors for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/monitoring", Summary: "Create ML monitor", Description: "Create a new ML monitor with model name, owner, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/monitoring/{mid}", Summary: "Get ML monitor", Description: "Returns a single ML monitor by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/mlops/monitoring/{mid}", Summary: "Delete ML monitor", Description: "Delete an ML monitor by ID.", Tag: "Workspace — AI"},
			// Workspace — Security: SSO / Federation
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/sso", Summary: "List SSO providers", Description: "Returns all SSO providers for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/sso", Summary: "Create SSO provider", Description: "Create a new SSO provider with name, protocol, domains, and region.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/sso/{pid}", Summary: "Get SSO provider", Description: "Returns a single SSO provider by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/sso/{pid}", Summary: "Delete SSO provider", Description: "Delete an SSO provider by ID.", Tag: "Workspace — Security"},
			// Workspace — Security: Multi-factor Authentication
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/mfa", Summary: "List MFA enrollments", Description: "Returns all MFA enrollments for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/mfa", Summary: "Create MFA enrollment", Description: "Create a new MFA enrollment with user, method, roles, and region.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/mfa/{eid}", Summary: "Get MFA enrollment", Description: "Returns a single MFA enrollment by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/mfa/{eid}", Summary: "Delete MFA enrollment", Description: "Delete an MFA enrollment by ID.", Tag: "Workspace — Security"},
			// Workspace — Security: Role-based Access Control
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/rbac", Summary: "List RBAC roles", Description: "Returns all RBAC roles for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/rbac", Summary: "Create RBAC role", Description: "Create a new RBAC role with name, type, permissions, and region.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/rbac/{rid}", Summary: "Get RBAC role", Description: "Returns a single RBAC role by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/iam/rbac/{rid}", Summary: "Delete RBAC role", Description: "Delete an RBAC role by ID.", Tag: "Workspace — Security"},
			// Workspace — Security: Encryption at Rest
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/data/enc-rest", Summary: "List encrypted resources", Description: "Returns all encryption-at-rest resources for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/data/enc-rest", Summary: "Create encrypted resource", Description: "Create a new encryption-at-rest resource with name, type, algorithm, key, and region.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/data/enc-rest/{rid}", Summary: "Get encrypted resource", Description: "Returns a single encryption-at-rest resource by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/data/enc-rest/{rid}", Summary: "Delete encrypted resource", Description: "Delete an encryption-at-rest resource by ID.", Tag: "Workspace — Security"},
			// Workspace — Security: Encryption in Transit
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/data/enc-transit", Summary: "List TLS endpoints", Description: "Returns all encryption-in-transit endpoints for a workspace. Supports q and protocol query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/data/enc-transit", Summary: "Create TLS endpoint", Description: "Create a new TLS endpoint with name, protocol, cipher, and region.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/data/enc-transit/{eid}", Summary: "Get TLS endpoint", Description: "Returns a single TLS endpoint by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/data/enc-transit/{eid}", Summary: "Delete TLS endpoint", Description: "Delete a TLS endpoint by ID.", Tag: "Workspace — Security"},
			// Workspace — Security: Key Management
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/data/kms", Summary: "List KMS keys", Description: "Returns all KMS keys for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/data/kms", Summary: "Create KMS key", Description: "Create a new KMS key with alias, algorithm, rotation, and region.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/data/kms/{kid}", Summary: "Get KMS key", Description: "Returns a single KMS key by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/data/kms/{kid}", Summary: "Delete KMS key", Description: "Delete a KMS key by ID.", Tag: "Workspace — Security"},
			// Workspace — Security: Web Application Firewall
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/network/waf", Summary: "List WAF rules", Description: "Returns all WAF rules for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/network/waf", Summary: "Create WAF rule", Description: "Create a new WAF rule with name, rule set, action, and priority.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/network/waf/{rid}", Summary: "Get WAF rule", Description: "Returns a single WAF rule by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/network/waf/{rid}", Summary: "Delete WAF rule", Description: "Delete a WAF rule by ID.", Tag: "Workspace — Security"},
			// Workspace — Security: DDoS Protection
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/network/ddos", Summary: "List DDoS profiles", Description: "Returns all DDoS protection profiles for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/network/ddos", Summary: "Create DDoS profile", Description: "Create a new DDoS protection profile with name, tier, mode, and threshold.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/network/ddos/{did}", Summary: "Get DDoS profile", Description: "Returns a single DDoS protection profile by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/network/ddos/{did}", Summary: "Delete DDoS profile", Description: "Delete a DDoS protection profile by ID.", Tag: "Workspace — Security"},
			// Workspace — Security: Intrusion Detection
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/network/ids", Summary: "List IDS rules", Description: "Returns all intrusion detection rules for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/network/ids", Summary: "Create IDS rule", Description: "Create a new IDS rule with name, category, severity, and action.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/network/ids/{iid}", Summary: "Get IDS rule", Description: "Returns a single IDS rule by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/network/ids/{iid}", Summary: "Delete IDS rule", Description: "Delete an IDS rule by ID.", Tag: "Workspace — Security"},
			// Workspace — Security: Policy Management
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/policy", Summary: "List compliance policies", Description: "Returns all compliance policies for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/policy", Summary: "Create compliance policy", Description: "Create a new compliance policy with name, framework, scope, and enforcement.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/policy/{pid}", Summary: "Get compliance policy", Description: "Returns a single compliance policy by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/policy/{pid}", Summary: "Delete compliance policy", Description: "Delete a compliance policy by ID.", Tag: "Workspace — Security"},
			// Workspace — Security: Audit Logging
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/audit", Summary: "List audit logs", Description: "Returns all audit log configurations for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/audit", Summary: "Create audit log", Description: "Create a new audit log configuration with name, source, retention, and format.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/audit/{aid}", Summary: "Get audit log", Description: "Returns a single audit log configuration by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/audit/{aid}", Summary: "Delete audit log", Description: "Delete an audit log configuration by ID.", Tag: "Workspace — Security"},
			// Workspace — Security: Regulatory Frameworks
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/regulatory", Summary: "List regulatory frameworks", Description: "Returns all regulatory frameworks for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/regulatory", Summary: "Create regulatory framework", Description: "Create a new regulatory framework with name, standard, region, and scope.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/regulatory/{fid}", Summary: "Get regulatory framework", Description: "Returns a single regulatory framework by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/compliance/regulatory/{fid}", Summary: "Delete regulatory framework", Description: "Delete a regulatory framework by ID.", Tag: "Workspace — Security"},
			// Workspace — DevOps: Source Control
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/source", Summary: "List source repos", Description: "Returns all source control repositories for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/source", Summary: "Create source repo", Description: "Create a new source repository with name, provider, URL, and branch.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/source/{rid}", Summary: "Get source repo", Description: "Returns a single source repository by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/source/{rid}", Summary: "Delete source repo", Description: "Delete a source repository by ID.", Tag: "Workspace — DevOps"},
			// Workspace — DevOps: Build Automation
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/build", Summary: "List build pipelines", Description: "Returns all build pipelines for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/build", Summary: "Create build pipeline", Description: "Create a new build pipeline with name, repo, trigger, and runtime.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/build/{bid}", Summary: "Get build pipeline", Description: "Returns a single build pipeline by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/build/{bid}", Summary: "Delete build pipeline", Description: "Delete a build pipeline by ID.", Tag: "Workspace — DevOps"},
			// Workspace — DevOps: Deployment Automation
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/deploy", Summary: "List deploy pipelines", Description: "Returns all deployment pipelines for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/deploy", Summary: "Create deploy pipeline", Description: "Create a new deployment pipeline with name, environment, strategy, and target.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/deploy/{did}", Summary: "Get deploy pipeline", Description: "Returns a single deployment pipeline by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/cicd/deploy/{did}", Summary: "Delete deploy pipeline", Description: "Delete a deployment pipeline by ID.", Tag: "Workspace — DevOps"},
			// Workspace — DevOps: IaC Templating
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/iac/templates", Summary: "List IaC templates", Description: "Returns all IaC templates for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/iac/templates", Summary: "Create IaC template", Description: "Create a new IaC template with name, engine, provider, and type.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/iac/templates/{tid}", Summary: "Get IaC template", Description: "Returns a single IaC template by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/iac/templates/{tid}", Summary: "Delete IaC template", Description: "Delete an IaC template by ID.", Tag: "Workspace — DevOps"},
			// Workspace — DevOps: IaC Config Management
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/iac/config", Summary: "List config management entries", Description: "Returns all config management entries for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/iac/config", Summary: "Create config management entry", Description: "Create a new config management entry with name, tool, and scope.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/iac/config/{cid}", Summary: "Get config management entry", Description: "Returns a single config management entry by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/iac/config/{cid}", Summary: "Delete config management entry", Description: "Delete a config management entry by ID.", Tag: "Workspace — DevOps"},
			// Workspace — DevOps: API Gateway
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/gateway", Summary: "List API gateways", Description: "Returns all API gateways for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/gateway", Summary: "Create API gateway", Description: "Create a new API gateway with name, protocol, upstream, rate_limit, and auth.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/gateway/{gid}", Summary: "Get API gateway", Description: "Returns a single API gateway by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/gateway/{gid}", Summary: "Delete API gateway", Description: "Delete an API gateway by ID.", Tag: "Workspace — DevOps"},
			// Workspace — DevOps: Developer Portal
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/portal", Summary: "List developer portals", Description: "Returns all developer portals for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/portal", Summary: "Create developer portal", Description: "Create a new developer portal with name, url, and theme.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/portal/{pid}", Summary: "Get developer portal", Description: "Returns a single developer portal by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/portal/{pid}", Summary: "Delete developer portal", Description: "Delete a developer portal by ID.", Tag: "Workspace — DevOps"},
			// Workspace — DevOps: API Analytics
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/analytics", Summary: "List API analytics", Description: "Returns all API analytics entries for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/analytics", Summary: "Create API analytics entry", Description: "Create a new API analytics entry with name, endpoint, and method.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/analytics/{aid}", Summary: "Get API analytics entry", Description: "Returns a single API analytics entry by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/api/analytics/{aid}", Summary: "Delete API analytics entry", Description: "Delete an API analytics entry by ID.", Tag: "Workspace — DevOps"},
			// Workspace — DevOps: Observability Logging
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/logging", Summary: "List log pipelines", Description: "Returns all log pipelines for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/logging", Summary: "Create log pipeline", Description: "Create a new log pipeline with name, source, format, storage, and retention.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/logging/{lid}", Summary: "Get log pipeline", Description: "Returns a single log pipeline by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/logging/{lid}", Summary: "Delete log pipeline", Description: "Delete a log pipeline by ID.", Tag: "Workspace — DevOps"},
			// Workspace — DevOps: Observability Metrics
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/metrics", Summary: "List metric dashboards", Description: "Returns all metric dashboards for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/metrics", Summary: "Create metric dashboard", Description: "Create a new metric dashboard with name, type, source, refresh rate, and retention.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/metrics/{mid}", Summary: "Get metric dashboard", Description: "Returns a single metric dashboard by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/metrics/{mid}", Summary: "Delete metric dashboard", Description: "Delete a metric dashboard by ID.", Tag: "Workspace — DevOps"},
			// Workspace — DevOps: Observability Tracing
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/tracing", Summary: "List tracing endpoints", Description: "Returns all distributed tracing endpoints for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/tracing", Summary: "Create tracing endpoint", Description: "Create a new tracing endpoint with name, backend, protocol, sample rate, and retention.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/tracing/{tid}", Summary: "Get tracing endpoint", Description: "Returns a single tracing endpoint by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/obs/tracing/{tid}", Summary: "Delete tracing endpoint", Description: "Delete a tracing endpoint by ID.", Tag: "Workspace — DevOps"},
			// Workspace — Apps: Messaging Queues
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/queues", Summary: "List message queues", Description: "Returns all message queues for a workspace. Supports q and status query filters.", Tag: "Workspace — Apps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/queues", Summary: "Create message queue", Description: "Create a new message queue with name, engine, type, and retention.", Tag: "Workspace — Apps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/queues/{qid}", Summary: "Get message queue", Description: "Returns a single message queue by ID.", Tag: "Workspace — Apps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/queues/{qid}", Summary: "Delete message queue", Description: "Delete a message queue by ID.", Tag: "Workspace — Apps"},
			// Workspace — Apps: Messaging Streaming
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/streaming", Summary: "List event streams", Description: "Returns all event streams for a workspace. Supports q and status query filters.", Tag: "Workspace — Apps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/streaming", Summary: "Create event stream", Description: "Create a new event stream with name, engine, partitions, replication, and retention.", Tag: "Workspace — Apps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/streaming/{sid}", Summary: "Get event stream", Description: "Returns a single event stream by ID.", Tag: "Workspace — Apps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/streaming/{sid}", Summary: "Delete event stream", Description: "Delete an event stream by ID.", Tag: "Workspace — Apps"},
			// Workspace — Apps: Messaging Service Bus
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/servicebus", Summary: "List service buses", Description: "Returns all service buses for a workspace. Supports q and status query filters.", Tag: "Workspace — Apps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/servicebus", Summary: "Create service bus", Description: "Create a new service bus with name and tier.", Tag: "Workspace — Apps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/servicebus/{bid}", Summary: "Get service bus", Description: "Returns a single service bus by ID.", Tag: "Workspace — Apps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/apps/messaging/servicebus/{bid}", Summary: "Delete service bus", Description: "Delete a service bus by ID.", Tag: "Workspace — Apps"},
			// Workspace — Apps: Workflow Automation
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/workflow/automation", Summary: "List workflow automations", Description: "Returns all workflow automations for a workspace. Supports q and status query filters.", Tag: "Workspace — Apps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/apps/workflow/automation", Summary: "Create workflow automation", Description: "Create a new workflow automation with name, trigger, and actions count.", Tag: "Workspace — Apps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/workflow/automation/{wfid}", Summary: "Get workflow automation", Description: "Returns a single workflow automation by ID.", Tag: "Workspace — Apps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/apps/workflow/automation/{wfid}", Summary: "Delete workflow automation", Description: "Delete a workflow automation by ID.", Tag: "Workspace — Apps"},
			// Workspace — Apps: Workflow Scheduler
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/workflow/scheduler", Summary: "List scheduled jobs", Description: "Returns all scheduled/cron jobs for a workspace. Supports q and status query filters.", Tag: "Workspace — Apps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/apps/workflow/scheduler", Summary: "Create scheduled job", Description: "Create a new scheduled job with name, cron schedule, command, and runtime.", Tag: "Workspace — Apps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/workflow/scheduler/{jid}", Summary: "Get scheduled job", Description: "Returns a single scheduled job by ID.", Tag: "Workspace — Apps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/apps/workflow/scheduler/{jid}", Summary: "Delete scheduled job", Description: "Delete a scheduled job by ID.", Tag: "Workspace — Apps"},
			// Workspace — Apps: Media Encoding
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/encoding", Summary: "List encoding jobs", Description: "Returns all video encoding jobs for a workspace. Supports q and status query filters.", Tag: "Workspace — Apps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/encoding", Summary: "Create encoding job", Description: "Create a new encoding job with name, input file, output format, resolution, and bitrate.", Tag: "Workspace — Apps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/encoding/{eid}", Summary: "Get encoding job", Description: "Returns a single encoding job by ID.", Tag: "Workspace — Apps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/encoding/{eid}", Summary: "Delete encoding job", Description: "Delete an encoding job by ID.", Tag: "Workspace — Apps"},
			// Workspace — Apps: Media Live Streaming
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/livestream", Summary: "List live streams", Description: "Returns all live streams for a workspace. Supports q and status query filters.", Tag: "Workspace — Apps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/livestream", Summary: "Create live stream", Description: "Create a new live stream with name, protocol, resolution, and bitrate.", Tag: "Workspace — Apps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/livestream/{lid}", Summary: "Get live stream", Description: "Returns a single live stream by ID.", Tag: "Workspace — Apps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/livestream/{lid}", Summary: "Delete live stream", Description: "Delete a live stream by ID.", Tag: "Workspace — Apps"},
			// Workspace — Apps: Media Storage
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/storage", Summary: "List media assets", Description: "Returns all media assets for a workspace. Supports q and status query filters.", Tag: "Workspace — Apps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/storage", Summary: "Create media asset", Description: "Create a new media asset with name, type, size, format, and bucket.", Tag: "Workspace — Apps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/storage/{mid}", Summary: "Get media asset", Description: "Returns a single media asset by ID.", Tag: "Workspace — Apps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/apps/media/storage/{mid}", Summary: "Delete media asset", Description: "Delete a media asset by ID.", Tag: "Workspace — Apps"},
			// Workspace Pro — Security: Certificates SSL/TLS
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/certificates/ssl-tls", Summary: "List SSL certificates", Description: "Returns all SSL/TLS certificates for a workspace. Supports q and status query filters.", Tag: "Workspace — Security"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/security/certificates/ssl-tls", Summary: "Issue SSL certificate", Description: "Issue a new SSL/TLS certificate with domain, type, key_algorithm, and auto_renew.", Tag: "Workspace — Security"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/security/certificates/ssl-tls/{cid}", Summary: "Get SSL certificate", Description: "Returns a single SSL certificate by ID.", Tag: "Workspace — Security"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/security/certificates/ssl-tls/{cid}", Summary: "Delete SSL certificate", Description: "Revoke and delete an SSL certificate by ID.", Tag: "Workspace — Security"},
			// Workspace Pro — DevOps: Monitoring Metrics
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/monitoring/metrics", Summary: "List monitoring alerts", Description: "Returns all monitoring alert rules for a workspace. Supports q and status query filters.", Tag: "Workspace — DevOps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/devops/monitoring/metrics", Summary: "Create monitoring alert", Description: "Create a new monitoring alert rule with name, metric, condition, threshold, window, severity, and notify_channels.", Tag: "Workspace — DevOps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/devops/monitoring/metrics/{aid}", Summary: "Get monitoring alert", Description: "Returns a single monitoring alert rule by ID.", Tag: "Workspace — DevOps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/devops/monitoring/metrics/{aid}", Summary: "Delete monitoring alert", Description: "Delete a monitoring alert rule by ID.", Tag: "Workspace — DevOps"},
			// Workspace Pro — AI: Inference LLM Endpoints
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/inference/llm-endpoints", Summary: "List LLM endpoints", Description: "Returns all LLM inference endpoints for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/inference/llm-endpoints", Summary: "Deploy LLM endpoint", Description: "Deploy a new LLM endpoint with name, model, model_format, gpu_type, gpu_count, region, and scale_to_zero.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/inference/llm-endpoints/{eid}", Summary: "Get LLM endpoint", Description: "Returns a single LLM endpoint by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/inference/llm-endpoints/{eid}", Summary: "Delete LLM endpoint", Description: "Delete an LLM endpoint by ID.", Tag: "Workspace — AI"},
			// Workspace Pro — Apps: Serverless Functions
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/serverless/functions", Summary: "List serverless functions (pro)", Description: "Returns all serverless functions for a workspace. Supports q and status query filters.", Tag: "Workspace — Apps"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/apps/serverless/functions", Summary: "Deploy serverless function", Description: "Deploy a new serverless function with name, runtime, trigger, route, memory_mb, timeout_sec, and region.", Tag: "Workspace — Apps"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/apps/serverless/functions/{fid}", Summary: "Get serverless function (pro)", Description: "Returns a single serverless function by ID.", Tag: "Workspace — Apps"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/apps/serverless/functions/{fid}", Summary: "Delete serverless function (pro)", Description: "Delete a serverless function by ID.", Tag: "Workspace — Apps"},
			// Workspace Pro — Networking: DNS Zones
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/zones", Summary: "List DNS zones (pro)", Description: "Returns all DNS zones for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/zones", Summary: "Create DNS zone", Description: "Create a new DNS zone with domain, zone_type, dnssec, and ttl.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/zones/{zid}", Summary: "Get DNS zone (pro)", Description: "Returns a single DNS zone by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/dns/zones/{zid}", Summary: "Delete DNS zone", Description: "Delete a DNS zone by ID.", Tag: "Workspace — Networking"},
			// Workspace Pro — Compute: Managed Kubernetes
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/kubernetes/managed-k8s", Summary: "List K8s clusters", Description: "Returns all managed Kubernetes clusters for a workspace. Supports q and status query filters.", Tag: "Workspace — Compute"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/compute/kubernetes/managed-k8s", Summary: "Create K8s cluster", Description: "Create a new managed Kubernetes cluster with name, k8s_version, node_pools, total_nodes, cni_plugin, region, auto_scale, and control_plane_ha.", Tag: "Workspace — Compute"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/compute/kubernetes/managed-k8s/{cid}", Summary: "Get K8s cluster", Description: "Returns a single managed Kubernetes cluster by ID.", Tag: "Workspace — Compute"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/compute/kubernetes/managed-k8s/{cid}", Summary: "Delete K8s cluster", Description: "Delete a managed Kubernetes cluster by ID.", Tag: "Workspace — Compute"},
			// Workspace Pro — Storage: Block Storage
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/volumes/block-storage", Summary: "List block volumes", Description: "Returns all block storage volumes for a workspace. Supports q and status query filters.", Tag: "Workspace — Storage"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/storage/volumes/block-storage", Summary: "Create block volume", Description: "Create a new block storage volume with name, size_gb, volume_type, iops, encrypted, region, and availability_zone.", Tag: "Workspace — Storage"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/storage/volumes/block-storage/{vid}", Summary: "Get block volume", Description: "Returns a single block storage volume by ID.", Tag: "Workspace — Storage"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/storage/volumes/block-storage/{vid}", Summary: "Delete block volume", Description: "Delete a block storage volume by ID.", Tag: "Workspace — Storage"},
			// Workspace Pro — Databases: Document Database
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/document-db", Summary: "List document databases", Description: "Returns all document databases for a workspace. Supports q and status query filters.", Tag: "Workspace — Databases"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/document-db", Summary: "Create document database", Description: "Create a new document database with name, engine, engine_version, cluster_nodes, storage_gb, encrypted, region, and multi_az.", Tag: "Workspace — Databases"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/document-db/{did}", Summary: "Get document database", Description: "Returns a single document database by ID.", Tag: "Workspace — Databases"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/databases/nosql/document-db/{did}", Summary: "Delete document database", Description: "Delete a document database by ID.", Tag: "Workspace — Databases"},
			// Workspace Pro — Networking: Application Load Balancer
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/load-balancer/application-lb", Summary: "List app load balancers", Description: "Returns all application load balancers for a workspace. Supports q and status query filters.", Tag: "Workspace — Networking"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/networking/load-balancer/application-lb", Summary: "Create app load balancer", Description: "Create a new application load balancer with name, protocol, listener_ports, waf_enabled, access_logs, sticky_session, and region.", Tag: "Workspace — Networking"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/networking/load-balancer/application-lb/{lid}", Summary: "Get app load balancer", Description: "Returns a single application load balancer by ID.", Tag: "Workspace — Networking"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/networking/load-balancer/application-lb/{lid}", Summary: "Delete app load balancer", Description: "Delete an application load balancer by ID.", Tag: "Workspace — Networking"},
			// Workspace Pro — AI: Vector Search / Embeddings
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/vector-search/embeddings", Summary: "List vector collections", Description: "Returns all vector search collections for a workspace. Supports q and status query filters.", Tag: "Workspace — AI"},
			{Method: "POST", Path: "/api/tenant/{slug}/workspace/{id}/ai/vector-search/embeddings", Summary: "Create vector collection", Description: "Create a new vector collection with name, dimensions, index_type, metric, embedding_model, and region.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/tenant/{slug}/workspace/{id}/ai/vector-search/embeddings/{cid}", Summary: "Get vector collection", Description: "Returns a single vector search collection by ID.", Tag: "Workspace — AI"},
			{Method: "DELETE", Path: "/api/tenant/{slug}/workspace/{id}/ai/vector-search/embeddings/{cid}", Summary: "Delete vector collection", Description: "Delete a vector search collection by ID.", Tag: "Workspace — AI"},
			{Method: "GET", Path: "/api/docs", Summary: "API documentation", Description: "This endpoint. Returns the API schema.", Tag: "Meta"},
		},
	}
	apiOK(w, doc)
}

// RegisterAPIRoutes wires all /api/* routes onto the mux.
func RegisterAPIRoutes(mux *http.ServeMux) {
	// Meta
	mux.HandleFunc("GET /api/docs", apiDocsHandler)

	// Auth
	mux.HandleFunc("POST /api/login", apiLoginHandler)
	mux.HandleFunc("POST /api/logout", apiLogoutHandler)

	// Public
	mux.HandleFunc("GET /api/status", apiStatusHandler)
	mux.HandleFunc("GET /api/sso-guide", apiSSOGuideHandler)

	// Sysadmin
	mux.HandleFunc("GET /api/sysadmin/dashboard", apiSysadminDashboardHandler)
	mux.HandleFunc("GET /api/sysadmin/tenants", apiSysadminTenantsListHandler)
	mux.HandleFunc("POST /api/sysadmin/tenants", apiSysadminTenantsCreateHandler)
	mux.HandleFunc("GET /api/sysadmin/tenants/{id}", apiSysadminTenantsGetHandler)
	mux.HandleFunc("PATCH /api/sysadmin/tenants/{id}", apiSysadminTenantsUpdateHandler)
	mux.HandleFunc("DELETE /api/sysadmin/tenants/{id}", apiSysadminTenantsDeleteHandler)
	mux.HandleFunc("GET /api/sysadmin/logs", apiSysadminLogsHandler)
	mux.HandleFunc("GET /api/sysadmin/jobs", apiSysadminJobsHandler)
	mux.HandleFunc("GET /api/sysadmin/database", apiSysadminDatabaseHandler)
	mux.HandleFunc("GET /api/sysadmin/database/table", apiSysadminDatabaseTableHandler)
	mux.HandleFunc("GET /api/sysadmin/database/erd", apiSysadminDatabaseERDHandler)
	mux.HandleFunc("POST /api/sysadmin/database/query", apiSysadminDatabaseQueryHandler)
	mux.HandleFunc("GET /api/sysadmin/replicas", apiSysadminReplicasHandler)
	mux.HandleFunc("GET /api/sysadmin/accounting", apiSysadminAccountingHandler)
	mux.HandleFunc("GET /api/sysadmin/current-replica", apiSysadminCurrentReplicaHandler)
	mux.HandleFunc("GET /api/sysadmin/sso", apiSysadminSSOHandler)
	mux.HandleFunc("PUT /api/sysadmin/sso", apiSysadminSSOUpdateHandler)
	mux.HandleFunc("POST /api/sysadmin/sso/clients", apiSysadminSSOCreateClientHandler)
	mux.HandleFunc("DELETE /api/sysadmin/sso/clients/{id}", apiSysadminSSODeleteClientHandler)
	mux.HandleFunc("POST /api/sysadmin/sso/cors", apiSysadminSSOAddCORSHandler)
	mux.HandleFunc("DELETE /api/sysadmin/sso/cors/{id}", apiSysadminSSODeleteCORSHandler)

	// User
	mux.HandleFunc("GET /api/user/settings", apiUserSettingsHandler)
	mux.HandleFunc("PATCH /api/user/settings", apiUserSettingsPatchHandler)
	mux.HandleFunc("POST /api/user/password", apiChangePasswordHandler)
	mux.HandleFunc("GET /api/user/sessions", apiUserSessionsHandler)
	mux.HandleFunc("DELETE /api/user/sessions/{id}", apiUserSessionRevokeHandler)
	mux.HandleFunc("POST /api/user/sessions/revoke-all", apiUserSessionsRevokeAllHandler)
	mux.HandleFunc("GET /api/user/api-keys", apiUserAPIKeysListHandler)
	mux.HandleFunc("POST /api/user/api-keys", apiUserAPIKeysCreateHandler)
	mux.HandleFunc("DELETE /api/user/api-keys/{id}", apiUserAPIKeyRevokeHandler)
	mux.HandleFunc("GET /api/user/workspaces", apiUserWorkspacesHandler)

	// Wizard (public)
	mux.HandleFunc("GET /api/wizard", apiWizardHandler)
	mux.HandleFunc("GET /api/wizard/{slug}", apiWizardAppHandler)
	mux.HandleFunc("POST /api/wizard/{slug}/deploy", apiWizardDeployHandler)

	// Tenant
	mux.HandleFunc("GET /api/tenant/{slug}/dashboard", apiTenantDashboardHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/clients", apiTenantClientsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/clients", apiTenantClientsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/clients/{id}", apiTenantClientsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/clients/{id}", apiTenantClientsDeleteHandler)

	// Tenant — Employees
	mux.HandleFunc("GET /api/tenant/{slug}/employees", apiTenantEmployeesListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/employees", apiTenantEmployeesCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/employees/{id}", apiTenantEmployeesGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/employees/{id}", apiTenantEmployeesDeleteHandler)

	// Tenant — Tickets
	mux.HandleFunc("GET /api/tenant/{slug}/tickets", apiTenantTicketsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/tickets", apiTenantTicketsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/tickets/{id}", apiTenantTicketsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/tickets/{id}", apiTenantTicketsDeleteHandler)

	// Tenant — Abuse Reports
	mux.HandleFunc("GET /api/tenant/{slug}/abuse-reports", apiTenantAbuseReportsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/abuse-reports", apiTenantAbuseReportsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/abuse-reports/{id}", apiTenantAbuseReportsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/abuse-reports/{id}", apiTenantAbuseReportsDeleteHandler)

	// Tenant — Refund Complaints
	mux.HandleFunc("GET /api/tenant/{slug}/refund-complaints", apiTenantRefundComplaintsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/refund-complaints", apiTenantRefundComplaintsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/refund-complaints/{id}", apiTenantRefundComplaintsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/refund-complaints/{id}", apiTenantRefundComplaintsDeleteHandler)

	// Tenant — One-Time Access
	mux.HandleFunc("GET /api/tenant/{slug}/one-time-access", apiTenantOneTimeAccessListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/one-time-access", apiTenantOneTimeAccessCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/one-time-access/{id}", apiTenantOneTimeAccessGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/one-time-access/{id}", apiTenantOneTimeAccessDeleteHandler)

	// Tenant — Resources
	mux.HandleFunc("GET /api/tenant/{slug}/resources", apiTenantResourcesListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/resources", apiTenantResourcesCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/resources/{id}", apiTenantResourcesGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/resources/{id}", apiTenantResourcesDeleteHandler)

	// Tenant — Monitoring
	mux.HandleFunc("GET /api/tenant/{slug}/monitoring", apiTenantMonitoringListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/monitoring", apiTenantMonitoringCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/monitoring/{id}", apiTenantMonitoringGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/monitoring/{id}", apiTenantMonitoringDeleteHandler)

	// Tenant — IDS (Intrusion Detection)
	mux.HandleFunc("GET /api/tenant/{slug}/ids", apiTenantIDSListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/ids", apiTenantIDSCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/ids/{id}", apiTenantIDSGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/ids/{id}", apiTenantIDSDeleteHandler)

	// Tenant — Incidents
	mux.HandleFunc("GET /api/tenant/{slug}/incidents", apiTenantIncidentsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/incidents", apiTenantIncidentsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/incidents/{id}", apiTenantIncidentsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/incidents/{id}", apiTenantIncidentsDeleteHandler)

	// Tenant — Status Page
	mux.HandleFunc("GET /api/tenant/{slug}/status-page", apiTenantStatusPageListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/status-page", apiTenantStatusPageCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/status-page/{id}", apiTenantStatusPageGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/status-page/{id}", apiTenantStatusPageDeleteHandler)

	// Tenant — Maintenance
	mux.HandleFunc("GET /api/tenant/{slug}/maintenance", apiTenantMaintenanceListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/maintenance", apiTenantMaintenanceCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/maintenance/{id}", apiTenantMaintenanceGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/maintenance/{id}", apiTenantMaintenanceDeleteHandler)

	// Tenant — Transactions
	mux.HandleFunc("GET /api/tenant/{slug}/transactions", apiTenantTransactionsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/transactions", apiTenantTransactionsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/transactions/{id}", apiTenantTransactionsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/transactions/{id}", apiTenantTransactionsDeleteHandler)

	// Tenant — Pricing
	mux.HandleFunc("GET /api/tenant/{slug}/pricing", apiTenantPricingListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/pricing", apiTenantPricingCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/pricing/{id}", apiTenantPricingGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/pricing/{id}", apiTenantPricingDeleteHandler)

	// Tenant — Tax
	mux.HandleFunc("GET /api/tenant/{slug}/tax", apiTenantTaxListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/tax", apiTenantTaxCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/tax/{id}", apiTenantTaxGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/tax/{id}", apiTenantTaxDeleteHandler)

	// Tenant — SLA
	mux.HandleFunc("GET /api/tenant/{slug}/sla", apiTenantSLAListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/sla", apiTenantSLACreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/sla/{id}", apiTenantSLAGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/sla/{id}", apiTenantSLADeleteHandler)

	// Tenant — Brand
	mux.HandleFunc("GET /api/tenant/{slug}/brand", apiTenantBrandGetHandler)
	mux.HandleFunc("PUT /api/tenant/{slug}/brand", apiTenantBrandUpdateHandler)

	// Tenant — Localization
	mux.HandleFunc("GET /api/tenant/{slug}/localization", apiTenantLocalizationGetHandler)
	mux.HandleFunc("PUT /api/tenant/{slug}/localization", apiTenantLocalizationUpdateHandler)

	// Tenant — Registration Settings
	mux.HandleFunc("GET /api/tenant/{slug}/registration", apiTenantRegistrationGetHandler)
	mux.HandleFunc("PUT /api/tenant/{slug}/registration", apiTenantRegistrationUpdateHandler)

	// Tenant — Register (public form)
	mux.HandleFunc("GET /api/tenant/{slug}/register", apiTenantRegisterGetHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/register", apiTenantRegisterPostHandler)

	// Workspace — Overview
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}", apiTenantWorkspaceOverviewHandler)

	// Workspace — Compute: General VMs
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/vms/general", apiWSComputeVMsGeneralListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/vms/general", apiWSComputeVMsGeneralCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/vms/general/{vmid}", apiWSComputeVMsGeneralGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/vms/general/{vmid}", apiWSComputeVMsGeneralDeleteHandler)

	// Workspace — Compute: GPU VMs
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/vms/gpu", apiWSComputeVMsGPUListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/vms/gpu", apiWSComputeVMsGPUCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/vms/gpu/{vmid}", apiWSComputeVMsGPUGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/vms/gpu/{vmid}", apiWSComputeVMsGPUDeleteHandler)

	// Workspace — Compute: Spot VMs
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/vms/spot", apiWSComputeVMsSpotListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/vms/spot", apiWSComputeVMsSpotCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/vms/spot/{vmid}", apiWSComputeVMsSpotGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/vms/spot/{vmid}", apiWSComputeVMsSpotDeleteHandler)

	// Workspace — Compute: Managed Kubernetes
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/containers/kubernetes", apiWSComputeK8sListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/containers/kubernetes", apiWSComputeK8sCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/containers/kubernetes/{cid}", apiWSComputeK8sGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/containers/kubernetes/{cid}", apiWSComputeK8sDeleteHandler)

	// Workspace — Compute: Container Registry
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/containers/registry", apiWSComputeRegistryListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/containers/registry", apiWSComputeRegistryCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/containers/registry/{rid}", apiWSComputeRegistryGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/containers/registry/{rid}", apiWSComputeRegistryDeleteHandler)

	// Workspace — Compute: Container Instances
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/containers/instances", apiWSComputeContainerInstancesListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/containers/instances", apiWSComputeContainerInstancesCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/containers/instances/{ciid}", apiWSComputeContainerInstancesGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/containers/instances/{ciid}", apiWSComputeContainerInstancesDeleteHandler)

	// Workspace — Compute: Web App Hosting
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/hosting/web", apiWSComputeHostingWebListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/hosting/web", apiWSComputeHostingWebCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/hosting/web/{waid}", apiWSComputeHostingWebGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/hosting/web/{waid}", apiWSComputeHostingWebDeleteHandler)

	// Workspace — Compute: Static Site Hosting
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/hosting/static", apiWSComputeHostingStaticListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/hosting/static", apiWSComputeHostingStaticCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/hosting/static/{ssid}", apiWSComputeHostingStaticGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/hosting/static/{ssid}", apiWSComputeHostingStaticDeleteHandler)

	// Workspace — Compute: Serverless Functions
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/serverless/functions", apiWSComputeServerlessFunctionsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/serverless/functions", apiWSComputeServerlessFunctionsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/serverless/functions/{fid}", apiWSComputeServerlessFunctionsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/serverless/functions/{fid}", apiWSComputeServerlessFunctionsDeleteHandler)

	// Workspace — Compute: Edge Functions
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/serverless/edge", apiWSComputeServerlessEdgeListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/serverless/edge", apiWSComputeServerlessEdgeCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/serverless/edge/{eid}", apiWSComputeServerlessEdgeGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/serverless/edge/{eid}", apiWSComputeServerlessEdgeDeleteHandler)

	// Workspace — Compute: Batch Jobs
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/batch/jobs", apiWSComputeBatchJobsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/batch/jobs", apiWSComputeBatchJobsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/batch/jobs/{jid}", apiWSComputeBatchJobsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/batch/jobs/{jid}", apiWSComputeBatchJobsDeleteHandler)

	// Workspace — Compute: HPC / Grid Computing
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/batch/hpc", apiWSComputeBatchHPCListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/batch/hpc", apiWSComputeBatchHPCCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/batch/hpc/{hid}", apiWSComputeBatchHPCGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/batch/hpc/{hid}", apiWSComputeBatchHPCDeleteHandler)

	// Workspace — Storage: Object Standard
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/object/standard", apiWSStorageObjectStandardListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/object/standard", apiWSStorageObjectStandardCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/object/standard/{bid}", apiWSStorageObjectStandardGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/object/standard/{bid}", apiWSStorageObjectStandardDeleteHandler)

	// Workspace — Storage: Object Archive
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/object/archive", apiWSStorageObjectArchiveListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/object/archive", apiWSStorageObjectArchiveCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/object/archive/{vid}", apiWSStorageObjectArchiveGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/object/archive/{vid}", apiWSStorageObjectArchiveDeleteHandler)

	// Workspace — Storage: Data Lake
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/object/datalake", apiWSStorageObjectDatalakeListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/object/datalake", apiWSStorageObjectDatalakeCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/object/datalake/{zid}", apiWSStorageObjectDatalakeGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/object/datalake/{zid}", apiWSStorageObjectDatalakeDeleteHandler)

	// Workspace — Storage: Block SSD
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/block/ssd", apiWSStorageBlockSSDListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/block/ssd", apiWSStorageBlockSSDCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/block/ssd/{vid}", apiWSStorageBlockSSDGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/block/ssd/{vid}", apiWSStorageBlockSSDDeleteHandler)

	// Workspace — Storage: Block HDD
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/block/hdd", apiWSStorageBlockHDDListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/block/hdd", apiWSStorageBlockHDDCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/block/hdd/{vid}", apiWSStorageBlockHDDGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/block/hdd/{vid}", apiWSStorageBlockHDDDeleteHandler)

	// Workspace — Storage: File NFS
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/file/nfs", apiWSStorageFileNFSListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/file/nfs", apiWSStorageFileNFSCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/file/nfs/{sid}", apiWSStorageFileNFSGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/file/nfs/{sid}", apiWSStorageFileNFSDeleteHandler)

	// Workspace — Storage: File Distributed
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/file/distributed", apiWSStorageFileDistributedListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/file/distributed", apiWSStorageFileDistributedCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/file/distributed/{did}", apiWSStorageFileDistributedGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/file/distributed/{did}", apiWSStorageFileDistributedDeleteHandler)

	// Workspace — Storage: Backup Snapshots
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/backup/snapshots", apiWSStorageBackupSnapshotsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/backup/snapshots", apiWSStorageBackupSnapshotsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/backup/snapshots/{snid}", apiWSStorageBackupSnapshotsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/backup/snapshots/{snid}", apiWSStorageBackupSnapshotsDeleteHandler)

	// Workspace — Storage: Backup Disaster Recovery
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/backup/dr", apiWSStorageBackupDRListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/backup/dr", apiWSStorageBackupDRCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/backup/dr/{drid}", apiWSStorageBackupDRGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/backup/dr/{drid}", apiWSStorageBackupDRDeleteHandler)

	// Workspace — Storage: Backup Long-term Archiving
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/backup/archive", apiWSStorageBackupArchiveListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/backup/archive", apiWSStorageBackupArchiveCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/backup/archive/{aid}", apiWSStorageBackupArchiveGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/backup/archive/{aid}", apiWSStorageBackupArchiveDeleteHandler)

	// Workspace — Databases: Relational MySQL / PostgreSQL
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/relational/mysql-pg", apiWSDBRelationalMySQLPGListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/relational/mysql-pg", apiWSDBRelationalMySQLPGCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/relational/mysql-pg/{dbid}", apiWSDBRelationalMySQLPGGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/relational/mysql-pg/{dbid}", apiWSDBRelationalMySQLPGDeleteHandler)

	// Workspace — Databases: Relational SQL Server
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/relational/sql-server", apiWSDBRelationalSQLServerListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/relational/sql-server", apiWSDBRelationalSQLServerCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/relational/sql-server/{dbid}", apiWSDBRelationalSQLServerGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/relational/sql-server/{dbid}", apiWSDBRelationalSQLServerDeleteHandler)

	// Workspace — Databases: Relational Cloud-native
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/relational/cloud-native", apiWSDBRelationalCloudNativeListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/relational/cloud-native", apiWSDBRelationalCloudNativeCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/relational/cloud-native/{dbid}", apiWSDBRelationalCloudNativeGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/relational/cloud-native/{dbid}", apiWSDBRelationalCloudNativeDeleteHandler)

	// Workspace — Databases: NoSQL Document
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/nosql/document", apiWSDBNoSQLDocumentListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/nosql/document", apiWSDBNoSQLDocumentCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/nosql/document/{dbid}", apiWSDBNoSQLDocumentGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/nosql/document/{dbid}", apiWSDBNoSQLDocumentDeleteHandler)

	// Workspace — Databases: NoSQL Key-Value
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/nosql/keyvalue", apiWSDBNoSQLKeyValueListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/nosql/keyvalue", apiWSDBNoSQLKeyValueCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/nosql/keyvalue/{dbid}", apiWSDBNoSQLKeyValueGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/nosql/keyvalue/{dbid}", apiWSDBNoSQLKeyValueDeleteHandler)

	// Workspace — Databases: NoSQL Wide-Column
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/nosql/wide-column", apiWSDBNoSQLWideColumnListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/nosql/wide-column", apiWSDBNoSQLWideColumnCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/nosql/wide-column/{dbid}", apiWSDBNoSQLWideColumnGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/nosql/wide-column/{dbid}", apiWSDBNoSQLWideColumnDeleteHandler)

	// Workspace — Databases: NoSQL Graph
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/nosql/graph", apiWSDBNoSQLGraphListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/nosql/graph", apiWSDBNoSQLGraphCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/nosql/graph/{dbid}", apiWSDBNoSQLGraphGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/nosql/graph/{dbid}", apiWSDBNoSQLGraphDeleteHandler)

	// Workspace — Databases: Memory Cache
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/memory/cache", apiWSDBMemoryCacheListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/memory/cache", apiWSDBMemoryCacheCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/memory/cache/{cid}", apiWSDBMemoryCacheGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/memory/cache/{cid}", apiWSDBMemoryCacheDeleteHandler)

	// Workspace — Databases: Memory Session Store
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/memory/session", apiWSDBMemorySessionListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/memory/session", apiWSDBMemorySessionCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/memory/session/{sid}", apiWSDBMemorySessionGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/memory/session/{sid}", apiWSDBMemorySessionDeleteHandler)

	// Workspace — Databases: Analytics Data Warehouse
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/analytics/warehouse", apiWSDBAnalyticsWarehouseListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/analytics/warehouse", apiWSDBAnalyticsWarehouseCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/analytics/warehouse/{dwid}", apiWSDBAnalyticsWarehouseGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/analytics/warehouse/{dwid}", apiWSDBAnalyticsWarehouseDeleteHandler)

	// Workspace — Databases: Analytics Time-Series
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/analytics/timeseries", apiWSDBAnalyticsTimeSeriesListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/analytics/timeseries", apiWSDBAnalyticsTimeSeriesCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/analytics/timeseries/{tsid}", apiWSDBAnalyticsTimeSeriesGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/analytics/timeseries/{tsid}", apiWSDBAnalyticsTimeSeriesDeleteHandler)

	// Workspace — Databases: Analytics Search & Full-text
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/analytics/search", apiWSDBAnalyticsSearchListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/analytics/search", apiWSDBAnalyticsSearchCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/analytics/search/{seid}", apiWSDBAnalyticsSearchGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/analytics/search/{seid}", apiWSDBAnalyticsSearchDeleteHandler)

	// Workspace — Networking: VPC / Subnets
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/vnet/vpc", apiWSNetVPCListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/vnet/vpc", apiWSNetVPCCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/vnet/vpc/{vid}", apiWSNetVPCGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/vnet/vpc/{vid}", apiWSNetVPCDeleteHandler)

	// Workspace — Networking: Peering & Transit
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/vnet/peering", apiWSNetPeeringListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/vnet/peering", apiWSNetPeeringCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/vnet/peering/{pid}", apiWSNetPeeringGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/vnet/peering/{pid}", apiWSNetPeeringDeleteHandler)

	// Workspace — Networking: Network Security Groups
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/vnet/nsg", apiWSNetNSGListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/vnet/nsg", apiWSNetNSGCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/vnet/nsg/{nid}", apiWSNetNSGGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/vnet/nsg/{nid}", apiWSNetNSGDeleteHandler)

	// Workspace — Networking: Application Load Balancer
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/lb/app", apiWSNetLBAppListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/lb/app", apiWSNetLBAppCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/lb/app/{lid}", apiWSNetLBAppGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/lb/app/{lid}", apiWSNetLBAppDeleteHandler)

	// Workspace — Networking: Network Load Balancer
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/lb/network", apiWSNetLBNetworkListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/lb/network", apiWSNetLBNetworkCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/lb/network/{lid}", apiWSNetLBNetworkGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/lb/network/{lid}", apiWSNetLBNetworkDeleteHandler)

	// Workspace — Networking: Global Load Balancing
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/lb/global", apiWSNetLBGlobalListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/lb/global", apiWSNetLBGlobalCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/lb/global/{gid}", apiWSNetLBGlobalGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/lb/global/{gid}", apiWSNetLBGlobalDeleteHandler)

	// Workspace — Networking: DNS Management
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/dns/manage", apiWSNetDNSManageListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/dns/manage", apiWSNetDNSManageCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/dns/manage/{zid}", apiWSNetDNSManageGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/dns/manage/{zid}", apiWSNetDNSManageDeleteHandler)

	// Workspace — Networking: Domain Registration
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/dns/domains", apiWSNetDNSDomainsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/dns/domains", apiWSNetDNSDomainsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/dns/domains/{did}", apiWSNetDNSDomainsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/dns/domains/{did}", apiWSNetDNSDomainsDeleteHandler)

	// Workspace — Networking: CDN Static Asset Delivery
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/cdn/static", apiWSNetCDNStaticListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/cdn/static", apiWSNetCDNStaticCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/cdn/static/{did}", apiWSNetCDNStaticGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/cdn/static/{did}", apiWSNetCDNStaticDeleteHandler)

	// Workspace — Networking: CDN Video Streaming
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/cdn/video", apiWSNetCDNVideoListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/cdn/video", apiWSNetCDNVideoCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/cdn/video/{vid}", apiWSNetCDNVideoGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/cdn/video/{vid}", apiWSNetCDNVideoDeleteHandler)

	// Workspace — Networking: CDN Edge Caching
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/cdn/edge", apiWSNetCDNEdgeListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/cdn/edge", apiWSNetCDNEdgeCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/cdn/edge/{eid}", apiWSNetCDNEdgeGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/cdn/edge/{eid}", apiWSNetCDNEdgeDeleteHandler)

	// Workspace — Networking: VPN Gateway
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/private/vpn", apiWSNetPrivVPNListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/private/vpn", apiWSNetPrivVPNCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/private/vpn/{gid}", apiWSNetPrivVPNGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/private/vpn/{gid}", apiWSNetPrivVPNDeleteHandler)

	// Workspace — Networking: Direct Connect
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/private/direct", apiWSNetPrivDirectListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/private/direct", apiWSNetPrivDirectCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/private/direct/{cid}", apiWSNetPrivDirectGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/private/direct/{cid}", apiWSNetPrivDirectDeleteHandler)

	// Workspace — Networking: SD-WAN
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/private/sdwan", apiWSNetPrivSDWANListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/private/sdwan", apiWSNetPrivSDWANCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/private/sdwan/{nid}", apiWSNetPrivSDWANGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/private/sdwan/{nid}", apiWSNetPrivSDWANDeleteHandler)

	// Workspace — AI: Distributed Training
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/training/distributed", apiWSAITrainDistListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/training/distributed", apiWSAITrainDistCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/training/distributed/{jid}", apiWSAITrainDistGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/training/distributed/{jid}", apiWSAITrainDistDeleteHandler)

	// Workspace — AI: AutoML
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/training/automl", apiWSAITrainAutoMLListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/training/automl", apiWSAITrainAutoMLCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/training/automl/{rid}", apiWSAITrainAutoMLGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/training/automl/{rid}", apiWSAITrainAutoMLDeleteHandler)

	// Workspace — AI: Fine-tuning
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/training/finetuning", apiWSAITrainFineTuneListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/training/finetuning", apiWSAITrainFineTuneCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/training/finetuning/{ftid}", apiWSAITrainFineTuneGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/training/finetuning/{ftid}", apiWSAITrainFineTuneDeleteHandler)

	// Workspace — AI: Inference Endpoints
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/hosting/inference", apiWSAIHostInferenceListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/hosting/inference", apiWSAIHostInferenceCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/hosting/inference/{eid}", apiWSAIHostInferenceGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/hosting/inference/{eid}", apiWSAIHostInferenceDeleteHandler)

	// Workspace — AI: Batch Inference
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/hosting/batch", apiWSAIHostBatchListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/hosting/batch", apiWSAIHostBatchCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/hosting/batch/{bid}", apiWSAIHostBatchGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/hosting/batch/{bid}", apiWSAIHostBatchDeleteHandler)

	// Workspace — AI: Edge Deployment
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/hosting/edge", apiWSAIHostEdgeListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/hosting/edge", apiWSAIHostEdgeCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/hosting/edge/{did}", apiWSAIHostEdgeGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/hosting/edge/{did}", apiWSAIHostEdgeDeleteHandler)

	// Workspace — AI: Vision API
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/apis/vision", apiWSAIAPIsVisionListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/apis/vision", apiWSAIAPIsVisionCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/apis/vision/{aid}", apiWSAIAPIsVisionGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/apis/vision/{aid}", apiWSAIAPIsVisionDeleteHandler)

	// Workspace — AI: Language / NLP API
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/apis/language", apiWSAIAPIsLanguageListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/apis/language", apiWSAIAPIsLanguageCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/apis/language/{aid}", apiWSAIAPIsLanguageGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/apis/language/{aid}", apiWSAIAPIsLanguageDeleteHandler)

	// Workspace — AI: Speech API
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/apis/speech", apiWSAIAPIsSpeechListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/apis/speech", apiWSAIAPIsSpeechCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/apis/speech/{aid}", apiWSAIAPIsSpeechGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/apis/speech/{aid}", apiWSAIAPIsSpeechDeleteHandler)

	// Workspace — AI: Recommendation API
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/apis/recommendation", apiWSAIAPIsRecommendListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/apis/recommendation", apiWSAIAPIsRecommendCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/apis/recommendation/{aid}", apiWSAIAPIsRecommendGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/apis/recommendation/{aid}", apiWSAIAPIsRecommendDeleteHandler)

	// Workspace — AI: ETL / ELT Pipelines
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/data/etl", apiWSAIDataETLListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/data/etl", apiWSAIDataETLCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/data/etl/{pid}", apiWSAIDataETLGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/data/etl/{pid}", apiWSAIDataETLDeleteHandler)

	// Workspace — AI: Streaming Ingestion
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/data/streaming", apiWSAIDataStreamListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/data/streaming", apiWSAIDataStreamCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/data/streaming/{sid}", apiWSAIDataStreamGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/data/streaming/{sid}", apiWSAIDataStreamDeleteHandler)

	// Workspace — AI: Data Catalog
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/data/catalog", apiWSAIDataCatalogListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/data/catalog", apiWSAIDataCatalogCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/data/catalog/{aid}", apiWSAIDataCatalogGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/data/catalog/{aid}", apiWSAIDataCatalogDeleteHandler)

	// Workspace — AI: Experiment Tracking
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/mlops/experiments", apiWSAIMLOpsExperimentsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/mlops/experiments", apiWSAIMLOpsExperimentsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/mlops/experiments/{eid}", apiWSAIMLOpsExperimentsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/mlops/experiments/{eid}", apiWSAIMLOpsExperimentsDeleteHandler)

	// Workspace — AI: Model Registry
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/mlops/registry", apiWSAIMLOpsRegistryListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/mlops/registry", apiWSAIMLOpsRegistryCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/mlops/registry/{mid}", apiWSAIMLOpsRegistryGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/mlops/registry/{mid}", apiWSAIMLOpsRegistryDeleteHandler)

	// Workspace — AI: Monitoring & Drift Detection
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/mlops/monitoring", apiWSAIMLOpsMonitoringListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/mlops/monitoring", apiWSAIMLOpsMonitoringCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/mlops/monitoring/{mid}", apiWSAIMLOpsMonitoringGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/mlops/monitoring/{mid}", apiWSAIMLOpsMonitoringDeleteHandler)

	// Workspace — Security: SSO / Federation
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/iam/sso", apiWSSecIAMSSOListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/iam/sso", apiWSSecIAMSSOCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/iam/sso/{pid}", apiWSSecIAMSSOGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/iam/sso/{pid}", apiWSSecIAMSSODeleteHandler)

	// Workspace — Security: Multi-factor Authentication
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/iam/mfa", apiWSSecIAMMFAListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/iam/mfa", apiWSSecIAMMFACreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/iam/mfa/{eid}", apiWSSecIAMMFAGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/iam/mfa/{eid}", apiWSSecIAMMFADeleteHandler)

	// Workspace — Security: Role-based Access Control
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/iam/rbac", apiWSSecIAMRBACListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/iam/rbac", apiWSSecIAMRBACCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/iam/rbac/{rid}", apiWSSecIAMRBACGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/iam/rbac/{rid}", apiWSSecIAMRBACDeleteHandler)

	// Workspace — Security: Encryption at Rest
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/data/enc-rest", apiWSSecDataEncRestListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/data/enc-rest", apiWSSecDataEncRestCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/data/enc-rest/{rid}", apiWSSecDataEncRestGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/data/enc-rest/{rid}", apiWSSecDataEncRestDeleteHandler)

	// Workspace — Security: Encryption in Transit
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/data/enc-transit", apiWSSecDataEncTransitListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/data/enc-transit", apiWSSecDataEncTransitCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/data/enc-transit/{eid}", apiWSSecDataEncTransitGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/data/enc-transit/{eid}", apiWSSecDataEncTransitDeleteHandler)

	// Workspace — Security: Key Management
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/data/kms", apiWSSecDataKMSListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/data/kms", apiWSSecDataKMSCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/data/kms/{kid}", apiWSSecDataKMSGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/data/kms/{kid}", apiWSSecDataKMSDeleteHandler)

	// Workspace — Security: Web Application Firewall
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/network/waf", apiWSSecNetWAFListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/network/waf", apiWSSecNetWAFCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/network/waf/{rid}", apiWSSecNetWAFGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/network/waf/{rid}", apiWSSecNetWAFDeleteHandler)

	// Workspace — Security: DDoS Protection
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/network/ddos", apiWSSecNetDDoSListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/network/ddos", apiWSSecNetDDoSCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/network/ddos/{did}", apiWSSecNetDDoSGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/network/ddos/{did}", apiWSSecNetDDoSDeleteHandler)

	// Workspace — Security: Intrusion Detection
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/network/ids", apiWSSecNetIDSListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/network/ids", apiWSSecNetIDSCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/network/ids/{iid}", apiWSSecNetIDSGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/network/ids/{iid}", apiWSSecNetIDSDeleteHandler)

	// Workspace — Security: Policy Management
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/compliance/policy", apiWSSecCompPolicyListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/compliance/policy", apiWSSecCompPolicyCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/compliance/policy/{pid}", apiWSSecCompPolicyGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/compliance/policy/{pid}", apiWSSecCompPolicyDeleteHandler)

	// Workspace — Security: Audit Logging
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/compliance/audit", apiWSSecCompAuditListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/compliance/audit", apiWSSecCompAuditCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/compliance/audit/{aid}", apiWSSecCompAuditGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/compliance/audit/{aid}", apiWSSecCompAuditDeleteHandler)

	// Workspace — Security: Regulatory Frameworks
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/compliance/regulatory", apiWSSecCompRegulatoryListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/compliance/regulatory", apiWSSecCompRegulatoryCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/compliance/regulatory/{fid}", apiWSSecCompRegulatoryGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/compliance/regulatory/{fid}", apiWSSecCompRegulatoryDeleteHandler)

	// Workspace — DevOps: Source Control
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/cicd/source", apiWSDevOpsCICDSourceListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/cicd/source", apiWSDevOpsCICDSourceCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/cicd/source/{rid}", apiWSDevOpsCICDSourceGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/cicd/source/{rid}", apiWSDevOpsCICDSourceDeleteHandler)

	// Workspace — DevOps: Build Automation
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/cicd/build", apiWSDevOpsCICDBuildListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/cicd/build", apiWSDevOpsCICDBuildCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/cicd/build/{bid}", apiWSDevOpsCICDBuildGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/cicd/build/{bid}", apiWSDevOpsCICDBuildDeleteHandler)

	// Workspace — DevOps: Deployment Automation
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/cicd/deploy", apiWSDevOpsCICDDeployListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/cicd/deploy", apiWSDevOpsCICDDeployCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/cicd/deploy/{did}", apiWSDevOpsCICDDeployGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/cicd/deploy/{did}", apiWSDevOpsCICDDeployDeleteHandler)

	// Workspace — DevOps: IaC Templating
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/iac/templates", apiWSDevOpsIaCTemplatesListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/iac/templates", apiWSDevOpsIaCTemplatesCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/iac/templates/{tid}", apiWSDevOpsIaCTemplatesGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/iac/templates/{tid}", apiWSDevOpsIaCTemplatesDeleteHandler)

	// Workspace — DevOps: IaC Config Management
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/iac/config", apiWSDevOpsIaCConfigListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/iac/config", apiWSDevOpsIaCConfigCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/iac/config/{cid}", apiWSDevOpsIaCConfigGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/iac/config/{cid}", apiWSDevOpsIaCConfigDeleteHandler)

	// Workspace — DevOps: API Gateway
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/api/gateway", apiWSDevOpsAPIGatewayListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/api/gateway", apiWSDevOpsAPIGatewayCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/api/gateway/{gid}", apiWSDevOpsAPIGatewayGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/api/gateway/{gid}", apiWSDevOpsAPIGatewayDeleteHandler)

	// Workspace — DevOps: Developer Portal
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/api/portal", apiWSDevOpsAPIPortalListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/api/portal", apiWSDevOpsAPIPortalCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/api/portal/{pid}", apiWSDevOpsAPIPortalGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/api/portal/{pid}", apiWSDevOpsAPIPortalDeleteHandler)

	// Workspace — DevOps: API Analytics
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/api/analytics", apiWSDevOpsAPIAnalyticsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/api/analytics", apiWSDevOpsAPIAnalyticsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/api/analytics/{aid}", apiWSDevOpsAPIAnalyticsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/api/analytics/{aid}", apiWSDevOpsAPIAnalyticsDeleteHandler)

	// Workspace — DevOps: Observability Logging
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/obs/logging", apiWSDevOpsObsLoggingListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/obs/logging", apiWSDevOpsObsLoggingCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/obs/logging/{lid}", apiWSDevOpsObsLoggingGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/obs/logging/{lid}", apiWSDevOpsObsLoggingDeleteHandler)

	// Workspace — DevOps: Observability Metrics
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/obs/metrics", apiWSDevOpsObsMetricsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/obs/metrics", apiWSDevOpsObsMetricsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/obs/metrics/{mid}", apiWSDevOpsObsMetricsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/obs/metrics/{mid}", apiWSDevOpsObsMetricsDeleteHandler)

	// Workspace — DevOps: Observability Tracing
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/obs/tracing", apiWSDevOpsObsTracingListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/obs/tracing", apiWSDevOpsObsTracingCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/obs/tracing/{tid}", apiWSDevOpsObsTracingGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/obs/tracing/{tid}", apiWSDevOpsObsTracingDeleteHandler)

	// Workspace — Apps: Messaging Queues
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/messaging/queues", apiWSAppsMessagingQueuesListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/messaging/queues", apiWSAppsMessagingQueuesCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/messaging/queues/{qid}", apiWSAppsMessagingQueuesGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/messaging/queues/{qid}", apiWSAppsMessagingQueuesDeleteHandler)

	// Workspace — Apps: Messaging Streaming
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/messaging/streaming", apiWSAppsMessagingStreamingListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/messaging/streaming", apiWSAppsMessagingStreamingCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/messaging/streaming/{sid}", apiWSAppsMessagingStreamingGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/messaging/streaming/{sid}", apiWSAppsMessagingStreamingDeleteHandler)

	// Workspace — Apps: Messaging Service Bus
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/messaging/servicebus", apiWSAppsMessagingServiceBusListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/messaging/servicebus", apiWSAppsMessagingServiceBusCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/messaging/servicebus/{bid}", apiWSAppsMessagingServiceBusGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/messaging/servicebus/{bid}", apiWSAppsMessagingServiceBusDeleteHandler)

	// Workspace — Apps: Workflow Automation
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/workflow/automation", apiWSAppsWorkflowAutomationListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/workflow/automation", apiWSAppsWorkflowAutomationCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/workflow/automation/{wfid}", apiWSAppsWorkflowAutomationGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/workflow/automation/{wfid}", apiWSAppsWorkflowAutomationDeleteHandler)

	// Workspace — Apps: Workflow Scheduler
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/workflow/scheduler", apiWSAppsWorkflowSchedulerListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/workflow/scheduler", apiWSAppsWorkflowSchedulerCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/workflow/scheduler/{jid}", apiWSAppsWorkflowSchedulerGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/workflow/scheduler/{jid}", apiWSAppsWorkflowSchedulerDeleteHandler)

	// Workspace — Apps: Media Encoding
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/media/encoding", apiWSAppsMediaEncodingListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/media/encoding", apiWSAppsMediaEncodingCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/media/encoding/{eid}", apiWSAppsMediaEncodingGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/media/encoding/{eid}", apiWSAppsMediaEncodingDeleteHandler)

	// Workspace — Apps: Media Live Streaming
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/media/livestream", apiWSAppsMediaLivestreamListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/media/livestream", apiWSAppsMediaLivestreamCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/media/livestream/{lid}", apiWSAppsMediaLivestreamGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/media/livestream/{lid}", apiWSAppsMediaLivestreamDeleteHandler)

	// Workspace — Apps: Media Storage
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/media/storage", apiWSAppsMediaStorageListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/media/storage", apiWSAppsMediaStorageCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/media/storage/{mid}", apiWSAppsMediaStorageGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/media/storage/{mid}", apiWSAppsMediaStorageDeleteHandler)

	// Workspace — Multi-Cloud: Hybrid On-Premises Extension
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/onprem", apiWSMulticloudHybridOnpremListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/onprem", apiWSMulticloudHybridOnpremCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/onprem/{oid}", apiWSMulticloudHybridOnpremGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/onprem/{oid}", apiWSMulticloudHybridOnpremDeleteHandler)

	// Workspace — Multi-Cloud: Hybrid Unified Control Plane
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/control", apiWSMulticloudHybridControlListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/control", apiWSMulticloudHybridControlCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/control/{cid}", apiWSMulticloudHybridControlGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/control/{cid}", apiWSMulticloudHybridControlDeleteHandler)

	// Workspace — Multi-Cloud: Hybrid Edge Computing
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/edge", apiWSMulticloudHybridEdgeListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/edge", apiWSMulticloudHybridEdgeCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/edge/{eid}", apiWSMulticloudHybridEdgeGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/multicloud/hybrid/edge/{eid}", apiWSMulticloudHybridEdgeDeleteHandler)

	// Workspace — Multi-Cloud: Management Cost
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/cost", apiWSMulticloudMgmtCostListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/cost", apiWSMulticloudMgmtCostCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/cost/{pid}", apiWSMulticloudMgmtCostGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/cost/{pid}", apiWSMulticloudMgmtCostDeleteHandler)

	// Workspace — Multi-Cloud: Management Policy Governance
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/policy", apiWSMulticloudMgmtPolicyListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/policy", apiWSMulticloudMgmtPolicyCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/policy/{pid}", apiWSMulticloudMgmtPolicyGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/policy/{pid}", apiWSMulticloudMgmtPolicyDeleteHandler)

	// Workspace — Multi-Cloud: Management Workload Portability
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/workload", apiWSMulticloudMgmtWorkloadListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/workload", apiWSMulticloudMgmtWorkloadCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/workload/{wid}", apiWSMulticloudMgmtWorkloadGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/multicloud/mgmt/workload/{wid}", apiWSMulticloudMgmtWorkloadDeleteHandler)

	// Workspace — Multi-Cloud: Migration Assessment & Discovery
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/migration/assessment", apiWSMulticloudMigrationAssessmentListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/multicloud/migration/assessment", apiWSMulticloudMigrationAssessmentCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/migration/assessment/{aid}", apiWSMulticloudMigrationAssessmentGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/multicloud/migration/assessment/{aid}", apiWSMulticloudMigrationAssessmentDeleteHandler)

	// Workspace — Multi-Cloud: Migration Data
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/migration/data", apiWSMulticloudMigrationDataListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/multicloud/migration/data", apiWSMulticloudMigrationDataCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/migration/data/{did}", apiWSMulticloudMigrationDataGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/multicloud/migration/data/{did}", apiWSMulticloudMigrationDataDeleteHandler)

	// Workspace — Multi-Cloud: Migration Application
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/migration/app", apiWSMulticloudMigrationAppListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/multicloud/migration/app", apiWSMulticloudMigrationAppCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/migration/app/{mid}", apiWSMulticloudMigrationAppGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/multicloud/migration/app/{mid}", apiWSMulticloudMigrationAppDeleteHandler)

	// Workspace — Apps: App Hosting Static Sites
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/app-hosting/static-sites", apiWSAppsAppHostingStaticSitesListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/app-hosting/static-sites", apiWSAppsAppHostingStaticSitesCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/app-hosting/static-sites/{sid}", apiWSAppsAppHostingStaticSitesGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/app-hosting/static-sites/{sid}", apiWSAppsAppHostingStaticSitesDeleteHandler)

	// Workspace — Compute: Linux VMs (Cloud Servers)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/virtual-machines/linux-vms", apiWSComputeVMsLinuxListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/virtual-machines/linux-vms", apiWSComputeVMsLinuxCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/virtual-machines/linux-vms/{vid}", apiWSComputeVMsLinuxGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/virtual-machines/linux-vms/{vid}", apiWSComputeVMsLinuxDeleteHandler)

	// Workspace — Databases: Relational PostgreSQL (Managed Databases)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/relational/postgresql", apiWSDBRelationalPostgresListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/relational/postgresql", apiWSDBRelationalPostgresCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/relational/postgresql/{dbid}", apiWSDBRelationalPostgresGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/relational/postgresql/{dbid}", apiWSDBRelationalPostgresDeleteHandler)

	// Workspace — Storage: Object Storage S3-Compatible (Storage & Backups)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/object-storage/s3-compatible", apiWSStorageObjectS3ListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/object-storage/s3-compatible", apiWSStorageObjectS3CreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/object-storage/s3-compatible/{bid}", apiWSStorageObjectS3GetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/object-storage/s3-compatible/{bid}", apiWSStorageObjectS3DeleteHandler)

	// Workspace — Apps: Communication Transactional Email (Email & Messaging)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/communication/transactional-email", apiWSAppsCommEmailListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/communication/transactional-email", apiWSAppsCommEmailCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/communication/transactional-email/{eid}", apiWSAppsCommEmailGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/communication/transactional-email/{eid}", apiWSAppsCommEmailDeleteHandler)

	// Workspace — Networking: CDN PoP Management (Content Delivery)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/cdn/pop-management", apiWSNetCDNPoPListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/cdn/pop-management", apiWSNetCDNPoPCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/cdn/pop-management/{pid}", apiWSNetCDNPoPGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/cdn/pop-management/{pid}", apiWSNetCDNPoPDeleteHandler)

	// Workspace Pro — Security: Certificates SSL/TLS (Security & SSL)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/certificates/ssl-tls", apiWSSecCertSSLListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/certificates/ssl-tls", apiWSSecCertSSLCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/certificates/ssl-tls/{cid}", apiWSSecCertSSLGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/certificates/ssl-tls/{cid}", apiWSSecCertSSLDeleteHandler)

	// Workspace Pro — DevOps: Monitoring Metrics (Monitoring & Alerts)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/monitoring/metrics", apiWSDevOpsMonMetricsListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/monitoring/metrics", apiWSDevOpsMonMetricsCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/monitoring/metrics/{aid}", apiWSDevOpsMonMetricsGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/monitoring/metrics/{aid}", apiWSDevOpsMonMetricsDeleteHandler)

	// Workspace Pro — AI: Inference LLM Endpoints (AI & Automation)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/inference/llm-endpoints", apiWSAIInferenceLLMListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/inference/llm-endpoints", apiWSAIInferenceLLMCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/inference/llm-endpoints/{eid}", apiWSAIInferenceLLMGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/inference/llm-endpoints/{eid}", apiWSAIInferenceLLMDeleteHandler)

	// Workspace Pro — Apps: Serverless Functions
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/serverless/functions", apiWSAppsServerlessFuncListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/serverless/functions", apiWSAppsServerlessFuncCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/serverless/functions/{fid}", apiWSAppsServerlessFuncGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/serverless/functions/{fid}", apiWSAppsServerlessFuncDeleteHandler)

	// Workspace Pro — Networking: DNS Zones (Domains & DNS)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/dns/zones", apiWSNetDNSZonesListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/dns/zones", apiWSNetDNSZonesCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/dns/zones/{zid}", apiWSNetDNSZonesGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/dns/zones/{zid}", apiWSNetDNSZonesDeleteHandler)

	// Workspace Pro — Compute: Managed Kubernetes
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/kubernetes/managed-k8s", apiWSComputeProK8sListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/compute/kubernetes/managed-k8s", apiWSComputeProK8sCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/compute/kubernetes/managed-k8s/{cid}", apiWSComputeProK8sGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/compute/kubernetes/managed-k8s/{cid}", apiWSComputeProK8sDeleteHandler)

	// Workspace Pro — Storage: Block Storage
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/volumes/block-storage", apiWSStorageBlockVolListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/storage/volumes/block-storage", apiWSStorageBlockVolCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/storage/volumes/block-storage/{vid}", apiWSStorageBlockVolGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/storage/volumes/block-storage/{vid}", apiWSStorageBlockVolDeleteHandler)

	// Workspace Pro — Databases: Document Database
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/nosql/document-db", apiWSDBNoSQLDocDBListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/databases/nosql/document-db", apiWSDBNoSQLDocDBCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/databases/nosql/document-db/{did}", apiWSDBNoSQLDocDBGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/databases/nosql/document-db/{did}", apiWSDBNoSQLDocDBDeleteHandler)

	// Workspace Pro — Networking: Application Load Balancer
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/load-balancer/application-lb", apiWSNetLBAppLBListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/networking/load-balancer/application-lb", apiWSNetLBAppLBCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/networking/load-balancer/application-lb/{lid}", apiWSNetLBAppLBGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/networking/load-balancer/application-lb/{lid}", apiWSNetLBAppLBDeleteHandler)

	// Workspace Pro — AI: Vector Search / Embeddings
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/vector-search/embeddings", apiWSAIVectorSearchListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/ai/vector-search/embeddings", apiWSAIVectorSearchCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/ai/vector-search/embeddings/{cid}", apiWSAIVectorSearchGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/ai/vector-search/embeddings/{cid}", apiWSAIVectorSearchDeleteHandler)

	// Workspace Pro — Security: IAM Identity Platform
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/iam/identity-platform", apiWSSecIAMIdentityListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/iam/identity-platform", apiWSSecIAMIdentityCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/iam/identity-platform/{pid}", apiWSSecIAMIdentityGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/iam/identity-platform/{pid}", apiWSSecIAMIdentityDeleteHandler)

	// Workspace Pro — Security: Firewall Cloud WAF
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/firewall/cloud-waf", apiWSSecFirewallWAFListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/security/firewall/cloud-waf", apiWSSecFirewallWAFCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/security/firewall/cloud-waf/{wid}", apiWSSecFirewallWAFGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/security/firewall/cloud-waf/{wid}", apiWSSecFirewallWAFDeleteHandler)

	// Workspace Pro — DevOps: CI/CD Pipelines
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/cicd/pipelines", apiWSDevOpsCICDPipelineListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/devops/cicd/pipelines", apiWSDevOpsCICDPipelineCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/devops/cicd/pipelines/{pid}", apiWSDevOpsCICDPipelineGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/devops/cicd/pipelines/{pid}", apiWSDevOpsCICDPipelineDeleteHandler)

	// Workspace Pro — Apps: Queues Message Broker
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/queues/message-broker", apiWSAppsQueuesBrokerListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/apps/queues/message-broker", apiWSAppsQueuesBrokerCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/apps/queues/message-broker/{bid}", apiWSAppsQueuesBrokerGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/apps/queues/message-broker/{bid}", apiWSAppsQueuesBrokerDeleteHandler)

	// Workspace Pro — Multi-Cloud: Cost Optimizer
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/cost/cloud-optimizer", apiWSMulticloudCostOptimizerListHandler)
	mux.HandleFunc("POST /api/tenant/{slug}/workspace/{id}/multicloud/cost/cloud-optimizer", apiWSMulticloudCostOptimizerCreateHandler)
	mux.HandleFunc("GET /api/tenant/{slug}/workspace/{id}/multicloud/cost/cloud-optimizer/{oid}", apiWSMulticloudCostOptimizerGetHandler)
	mux.HandleFunc("DELETE /api/tenant/{slug}/workspace/{id}/multicloud/cost/cloud-optimizer/{oid}", apiWSMulticloudCostOptimizerDeleteHandler)

	// Sysadmin — Infrastructure Supply Hosts & Manual Assignable IP Pools
	RegisterSysadminInfrastructureAPIRoutes(mux)
}
