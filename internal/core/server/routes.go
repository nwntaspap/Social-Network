package server

import (
	"net/http"
	"strings"
)

func RegisterRoutes(s *Server) {
	api := "/api/v1"

	s.mux.HandleFunc(api+"/health", healthHandler)

	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static"))))

	s.mux.HandleFunc("/", spaHandler("frontend/static/index.html"))
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func spaHandler(indexPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/static/") {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	}
}
