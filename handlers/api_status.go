package handlers

import "net/http"

// ── GET /api/status ──────────────────────────────────────────

type apiStatusOverall struct {
	State    string `json:"state"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Updated  string `json:"updated_at"`
}

type apiStatusComponent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	State       string `json:"state"`
	Uptime      string `json:"uptime"`
}

type apiStatusMetric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
	Trend string `json:"trend"`
}

type apiStatusIncident struct {
	When     string `json:"timestamp"`
	Title    string `json:"title"`
	Severity string `json:"severity"`
	Body     string `json:"body"`
}

type apiStatusDay struct {
	Date  string `json:"date"`
	State string `json:"state"`
}

type apiStatusResponse struct {
	Overall    apiStatusOverall     `json:"overall"`
	Metrics    []apiStatusMetric    `json:"metrics"`
	Components []apiStatusComponent `json:"components"`
	Incidents  []apiStatusIncident  `json:"incidents"`
	UptimeDays []apiStatusDay       `json:"uptime_days"`
}

func apiStatusHandler(w http.ResponseWriter, r *http.Request) {
	src := buildPublicStatusData()

	resp := apiStatusResponse{
		Overall: apiStatusOverall{
			State:    src.Overall.State,
			Title:    src.Overall.Title,
			Subtitle: src.Overall.Subtitle,
			Updated:  src.Overall.Updated,
		},
	}

	for _, m := range src.Metrics {
		resp.Metrics = append(resp.Metrics, apiStatusMetric{
			Label: m.Label, Value: m.Value, Unit: m.Unit, Trend: m.Trend,
		})
	}

	for _, c := range src.Components {
		resp.Components = append(resp.Components, apiStatusComponent{
			Name: c.Name, Description: c.Description, State: c.State, Uptime: c.Uptime,
		})
	}

	for _, i := range src.Incidents {
		resp.Incidents = append(resp.Incidents, apiStatusIncident{
			When: i.When, Title: i.Title, Severity: i.Severity, Body: i.Body,
		})
	}

	for _, d := range src.Days {
		resp.UptimeDays = append(resp.UptimeDays, apiStatusDay{
			Date: d.Date, State: d.State,
		})
	}

	apiOK(w, resp)
}
