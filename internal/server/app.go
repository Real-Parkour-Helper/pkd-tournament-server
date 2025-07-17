package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"tournament-manager/internal/server/handlers"

	"github.com/gorilla/mux"
)

func StartServer() {
	r := mux.NewRouter()
	registerRoutes(r)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		slog.Warn("PORT environment variable not set. Using 8080 as a default port")
		port = "8080"
	}
	port = fmt.Sprintf(":%s", port)

	// TODO: add some middleware to log requests and handle panics and stuff
	slog.Error(http.ListenAndServe(port, r).Error())
}

func registerRoutes(r *mux.Router) {
	r.HandleFunc("/api/tournament", handlers.CreateTournament).Methods("POST")

	r.HandleFunc("/api/tournament/{id}/start", handlers.StartTournament).Methods("POST")
	r.HandleFunc("/api/tournament/{id}/result", handlers.SubmitGameResult).Methods("POST")
	r.HandleFunc("/api/tournament/{id}/status", handlers.GetTournamentStatus).Methods("GET")
	r.HandleFunc("/api/tournament/{id}/matches", handlers.GetNextMatches).Methods("GET")
	r.HandleFunc("/api/tournament/{id}/bracket", handlers.GetTournamentBracket).Methods("GET")
	r.HandleFunc("/api/tournament/{id}/stop", handlers.StopTournament).Methods("DELETE")

	r.HandleFunc("/api/tournaments/active", handlers.ListActiveTournaments).Methods("GET")

	r.HandleFunc("/api/signup", handlers.Signup).Methods("POST")
	r.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"ha": "ha"})
	}).Methods("GET")

	r.Use(authMiddleware)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != os.Getenv("AUTH_USER") || password != os.Getenv("AUTH_PASSWORD") {
			slog.Error("Attempted unauthorized access from %v (%v)", username, r.RemoteAddr)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
