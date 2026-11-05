package handlers

import (
	"net/http"
	"time"
)

// StatusOverall is the high-level health pill at the top of the page.
type StatusOverall struct {
	State    string // operational | degraded | partial | major
	Title    string
	Subtitle string
	Updated  string
}

// StatusComponent is a row in the components list.
type StatusComponent struct {
	Name        string
	Description string
	State       string // operational | degraded | partial | major | maintenance
	Uptime      string // e.g. "99.998%"
}

// StatusMetric is a small KPI tile.
type StatusMetric struct {
	Label string
	Value string
	Unit  string
	Trend string // up | down | flat
}

// StatusIncident is an entry in the incident timeline.
type StatusIncident struct {
	When     string
	Title    string
	Severity string // resolved | monitoring | investigating | major
	Body     string
}

// StatusPagePublicData drives the public /status page.
type StatusPagePublicData struct {
	Overall    StatusOverall
	Metrics    []StatusMetric
	Components []StatusComponent
	Incidents  []StatusIncident
	Days       []DayBar // 90-day uptime chart
}

// DayBar is one bar in the 90-day uptime visualisation.
type DayBar struct {
	State string // operational | degraded | partial | major | maintenance
	Date  string
}

func buildPublicStatusData() StatusPagePublicData {
	now := time.Now().UTC()
	days := make([]DayBar, 90)
	for i := 0; i < 90; i++ {
		d := now.AddDate(0, 0, -89+i)
		state := "operational"
		// Sprinkle in a few incidents to make the bar interesting.
		switch i {
		case 12:
			state = "degraded"
		case 24, 25:
			state = "partial"
		case 47:
			state = "maintenance"
		case 71:
			state = "degraded"
		}
		days[i] = DayBar{State: state, Date: d.Format("Jan 2")}
	}

	return StatusPagePublicData{
		Overall: StatusOverall{
			State:    "operational",
			Title:    "All systems operational",
			Subtitle: "All services running smoothly across every region.",
			Updated:  now.Format("Mon Jan 2, 15:04 MST"),
		},
		Metrics: []StatusMetric{
			{Label: "Uptime (90d)", Value: "99.991", Unit: "%", Trend: "up"},
			{Label: "Avg. response", Value: "142", Unit: "ms", Trend: "down"},
			{Label: "Incidents (30d)", Value: "2", Unit: "", Trend: "flat"},
			{Label: "Active regions", Value: "8", Unit: "", Trend: "up"},
		},
		Components: []StatusComponent{
			{Name: "API · platform-api", Description: "REST and GraphQL endpoints", State: "operational", Uptime: "99.998%"},
			{Name: "Workspace dashboard", Description: "Web client & realtime updates", State: "operational", Uptime: "99.997%"},
			{Name: "Authentication", Description: "Login, SSO, and OAuth flows", State: "operational", Uptime: "99.999%"},
			{Name: "Compute · provisioning", Description: "VM, container, and serverless creation", State: "operational", Uptime: "99.992%"},
			{Name: "Storage · object", Description: "S3-compatible buckets", State: "operational", Uptime: "99.996%"},
			{Name: "Storage · block", Description: "SSD and HDD volumes", State: "operational", Uptime: "99.995%"},
			{Name: "Databases · managed", Description: "Postgres, MySQL, Mongo, Redis", State: "operational", Uptime: "99.994%"},
			{Name: "Networking · load balancers", Description: "L4 & L7 traffic distribution", State: "operational", Uptime: "99.997%"},
			{Name: "CDN · global edge", Description: "180+ edge locations", State: "operational", Uptime: "99.999%"},
			{Name: "Billing & invoicing", Description: "Subscriptions and metering", State: "operational", Uptime: "99.998%"},
			{Name: "Support portal", Description: "Tickets and help center", State: "operational", Uptime: "99.992%"},
			{Name: "Marketplace", Description: "One-click app deployment", State: "operational", Uptime: "99.991%"},
		},
		Incidents: []StatusIncident{
			{When: "Apr 14, 11:42 UTC", Title: "Elevated latency in eu-central-1", Severity: "resolved",
				Body: "Resolved — A noisy upstream peer caused a brief spike in latency for ~12 minutes. Traffic was rerouted and metrics are back to baseline."},
			{When: "Apr 02, 03:18 UTC", Title: "Scheduled maintenance — block storage", Severity: "resolved",
				Body: "Completed — Rolling firmware upgrade on block storage hosts in us-east-1. No customer impact reported."},
			{When: "Mar 21, 16:05 UTC", Title: "Brief authentication degradation", Severity: "resolved",
				Body: "Resolved — A subset of SSO logins experienced 2-3s delay for ~9 minutes. Root cause: cache warm-up after a deploy."},
			{When: "Mar 04, 22:30 UTC", Title: "Status page maintenance", Severity: "resolved",
				Body: "Completed — The status page itself was upgraded with finer-grained component breakdown."},
		},
		Days: days,
	}
}

// StatusPublicHandler serves the public /status page.
func StatusPublicHandler(w http.ResponseWriter, r *http.Request) {
	render(w, buildPublicStatusData(), "base.html", "status.html")
}
