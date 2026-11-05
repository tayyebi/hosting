package handlers

import "net/http"

func SSOGuideHandler(w http.ResponseWriter, r *http.Request) {
	render(w, nil, "base.html", "sso_guide.html")
}
