package bootstrap

import (
	"context"
	"embed"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lmittmann/tint"
	"github.com/posthog/posthog-go"
	"github.com/prior-it/apollo/components"
	"github.com/prior-it/apollo/config"
	"github.com/prior-it/apollo/postgres"
	"github.com/prior-it/apollo/server"
)

// BootstrappedState is a state that is initialised through bootstrapping.
//
// The state is responsible for closing the database and the posthog client in its own Close function after it's done using it.
// The posthog client will be nil if posthog was disabled in config.
type BootstrappedState[state server.State] interface {
	server.State
	Init(server *server.Server[state], cfg *config.Config, db *postgres.DB, posthog posthog.Client)
}

// Minimal creates a new server and initializes all default systems.
// The Minimal bootstrapper is perfect for very lightweight applications or (almost) static sites.
//
// This will only initialise the server itself with some middleware but no other services.
//
// You can supply additional middleware if you want to.
//
// Note that this function will add routes before returning, which means it is not possible to add additional middleware after calling this function.
func Minimal[state server.State](
	stt state,
	cfg *config.Config,
	staticFS embed.FS,
	middlewares ...func(http.Handler) http.Handler,
) *server.Server[state] {
	if cfg == nil {
		panic("You need to supply a config.Config value to bootstrap a new server")
	}
	s := server.New(stt, cfg)
	s.AttachDefaultMiddleware()

	s.UseStd(middlewares...)

	s.StaticFiles("/static", "static", staticFS)
	s.StaticFiles(
		"/apollo",
		os.Getenv("APOLLO_STATIC_FILES"),
		components.EmbedStatic,
	)

	return s
}

// Full creates a new server and initializes all default systems.
// The Full bootstrapper is perfect for (complex) web applications.
//
// This will initialise the server itself as well as most middleware, Sentry and PostHog (if enabled in config), and a postgres database.
//
// You can supply additional middleware if you want to.
//
// Note that this function will add routes before returning, which means it is not possible to add additional global middleware after calling this function.
func Full[state BootstrappedState[state]](
	stt state,
	cfg *config.Config,
	staticFS embed.FS,
	middlewares ...server.Middleware[state],
) *server.Server[state] {
	if cfg == nil {
		panic("You need to supply a config.Config value to bootstrap a new server")
	}

	logger := createLogger(cfg)

	s := server.New(stt, cfg).
		WithLogger(logger)

	// Initialize Sentry
	if cfg.Sentry.Enabled {
		initSentry(logger, cfg)
	}

	// Initialize PostHog
	var posthog posthog.Client
	if cfg.PostHog.Enabled {
		posthog = initPostHog(logger, cfg)
	}

	// Connect to the database
	db, err := postgres.NewDB(context.Background(), cfg.Database.URL, cfg.Database.Schema)
	if err != nil {
		logger.Error("Could not initialize database", "error", err)
		os.Exit(1)
	}

	stt.Init(s, cfg, db, posthog)

	s.AttachDefaultMiddleware()

	// Enable sentry middleware
	if cfg.Sentry.Enabled {
		sentryHandler := sentryhttp.New(sentryhttp.Options{
			Repanic:         true,
			WaitForDelivery: true,
			Timeout:         5 * time.Second, //nolint:mnd
		})
		s.UseStd(sentryHandler.Handle)
	}

	// Fully disable caching in debug mode
	if cfg.App.Debug {
		s.UseStd(middleware.NoCache)
	}

	s.Use(middlewares...)

	s.StaticFiles("/static", "static", staticFS)
	s.StaticFiles(
		"/apollo",
		os.Getenv("APOLLO_STATIC_FILES"),
		components.EmbedStatic,
	)

	return s
}

func createLogger(cfg *config.Config) *slog.Logger {
	var logger *slog.Logger
	loggerOptions := &slog.HandlerOptions{
		Level:     cfg.Log.Level.ToSlog(),
		AddSource: cfg.Log.Verbose && cfg.App.Debug,
	}
	switch cfg.Log.Format {
	case config.LogFormatPlaintext:
		logger = slog.New(slog.NewTextHandler(os.Stdout, loggerOptions))
	case config.LogFormatJSON:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, loggerOptions))
	case config.LogFormatColour:
		logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:   loggerOptions.AddSource,
			Level:       loggerOptions.Level,
			TimeFormat:  cfg.Log.TimeFormat,
			ReplaceAttr: server.ReplaceSlogAttributes(cfg),
		}))
	}
	slog.SetDefault(logger)
	return logger
}

func initSentry(logger *slog.Logger, cfg *config.Config) {
	logger.Debug("Trying to initialise Sentry")
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.Sentry.DSN,
		Debug:            cfg.App.Debug,
		AttachStacktrace: true,
		SampleRate:       cfg.Sentry.SampleRate,
		EnableTracing:    true,
		TracesSampleRate: cfg.Sentry.TracesRate,
		TracesSampler: sentry.TracesSampler(func(ctx sentry.SamplingContext) float64 {
			if ctx.Span.Name == "GET /ping" {
				return 0.0
			}
			return 1.0
		}),
		DebugWriter:        os.Stdout,
		ProfilesSampleRate: cfg.Sentry.ProfilesRate,
		ServerName:         cfg.App.URL,
		Release:            cfg.App.Version,
		Environment:        string(cfg.App.Env),
	}); err != nil {
		logger.Error("Sentry initialization failed", "error", err)
	} else {
		logger.Debug("Sentry initialised")
	}
}

func initPostHog(logger *slog.Logger, cfg *config.Config) posthog.Client {
	logger.Debug("Trying to initialise PostHog")
	client, err := posthog.NewWithConfig(
		cfg.PostHog.APIKey,
		posthog.Config{Endpoint: cfg.PostHog.Endpoint},
	)
	if err != nil {
		logger.Error("PostHog initialization failed", "error", err)
		return nil
	}

	logger.Debug("PostHog initialised")
	return client
}
