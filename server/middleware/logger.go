package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httplog/v2"
	"github.com/prior-it/apollo/config"
)

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
