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
	r.HandleFunc("/api/signup", handlers.Signup).Methods("POST")
	r.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"ha": "ha"})
	}).Methods("GET")
}
