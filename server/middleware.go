package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/httplog/v2"
	"github.com/prior-it/apollo/config"
)

// RequireLogin is middleware that requires that any user is logged in before continuing on.
func RequireLogin[state any](apollo *Apollo, _ state) (context.Context, error) {
	return apollo.Context(), apollo.RequiresLogin()
}

// Debug is middleware that can be inserted anywhere and will print some useful debug information about the current
// request.
func Debug(printFullRequest bool, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if printFullRequest {
			slog.Debug("Debug middleware", "request", r)
		} else {
			slog.Debug("Debug middleware", "path", r.URL.Path)
		}

		h.ServeHTTP(w, r)
	})
}

// HTTPLogger is middleware that will log HTTP requests, including context that might be added by the handler itself
// by calling apollo.LogField.
func HTTPLogger(cfg *config.Config) func(http.Handler) http.Handler {
	sourceFieldName := ""
	if cfg.Log.Verbose || cfg.App.Debug {
		sourceFieldName = "source"
	}
	logger := httplog.NewLogger(cfg.App.Name, httplog.Options{
		LogLevel: cfg.Log.Level.ToSlog(),
		JSON:     cfg.Log.Format == config.LogFormatJSON,
		Concise:  !cfg.Log.Verbose,
		Tags: map[string]string{
			"version": cfg.App.Version,
			"env":     string(cfg.App.Env),
		},
		RequestHeaders:  cfg.Log.Verbose,
		ResponseHeaders: cfg.Log.Verbose,
		QuietDownRoutes: []string{
			"/",
			"/favicon.ico",
			"/ping",
			"/static",
			"/apollo",
		},
		QuietDownPeriod: 10 * time.Second, //nolint:mnd
		SourceFieldName: sourceFieldName,
	})
	return httplog.RequestLogger(logger)
}
