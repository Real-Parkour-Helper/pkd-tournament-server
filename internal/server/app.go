package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func StartServer() {
	r := http.NewServeMux()

	port := os.Getenv("PORT")
	if len(port) == 0 {
		slog.Warn("PORT environment variable not set. Using 8080 as a default port")
		port = "8080"
	}
	port = fmt.Sprintf(":%s", port)

	// TODO: add some middleware to log requests and handle panics and stuff

	slog.Error(http.ListenAndServe(port, r).Error())
}
