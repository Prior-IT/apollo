package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/httplog/v2"
	"github.com/prior-it/apollo/config"
	"github.com/prior-it/apollo/core"
)

const CSRFTokenLength = 32

// noop is middleware that does nothing
func noop(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// RequireLogin is middleware that requires that any user is logged in before continuing on.
func RequireLogin[state any](apollo *Apollo, _ state) (context.Context, error) {
	return apollo.Context(), apollo.RequiresLogin()
}

func IsStaticFile(path string) bool {
	return strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/apollo/")
}

// RedirectSlashes is middleware that redirects all requests that end in `/` to requests without trailing slash.
// This differs from chi's own RedirectSlashes middleware in that it uses config.BaseURL() instead of the incoming
// request's host header.
func (server *Server[state]) RedirectSlashes(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// NOTE: Ignore static files since this method does not work with FileServer,
		// see https://github.com/go-chi/chi/issues/343
		if !IsStaticFile(path) && len(path) > 1 && path[len(path)-1] == '/' {
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path[:len(path)-1], r.URL.RawQuery)
			} else {
				path = path[:len(path)-1]
			}
			redirectURL := server.cfg.BaseURL() + path
			http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// CSRFTokenMiddleware injects a csrf token at the end of each request that can be checked on the next request
// using apollo.CheckCSRF.
func (server *Server[state]) CSRFTokenMiddleware() func(http.Handler) http.Handler {
	if server.sessionStore == nil {
		slog.Warn(
			"Not enabling the CSRF Token middleware since there is no SessionStore configured",
		)
		return noop
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if IsStaticFile(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := server.sessionStore.Get(r, cookieCSRF)
			if err != nil {
				log.Panicf("Invalid name for a cookie: %v\n", cookieCSRF)
			}
			configureCookie(server.cfg, cookie)

			oldToken := ""
			tokenCookie, ok := cookie.Values[sessionCSRFToken]
			if ok {
				oldToken, ok = tokenCookie.(string)
				if !ok {
					slog.Error(
						"csrf cookie token exists but it has an invalid type",
						"token",
						tokenCookie,
					)
					// oldToken will still be empty which is fine to continue on
				}
			}

			ctx := context.WithValue(r.Context(), ctxOldCSRFToken, oldToken)

			// Update the CSRF token so it will be different for the next request
			newTokenBytes := make([]byte, CSRFTokenLength)
			_, err = rand.Read(newTokenBytes)
			if err != nil {
				slog.Error("cannot generate new csrf token", "error", err)
				newTokenBytes = []byte{}
			}
			newToken := base64.URLEncoding.EncodeToString(newTokenBytes)
			ctx = context.WithValue(ctx, ctxNewCSRFToken, newToken)

			cookie.Values[sessionCSRFToken] = newToken
			err = server.sessionStore.Save(r, w, cookie)
			if err != nil {
				slog.Error("cannot set csrf cookie", "error", err)
			}

			next.ServeHTTP(w, r.WithContext(ctx))

			// If this was a HTMX request, send an OOB-update for the CSRF token
			if r.Header.Get("HX-Request") == "true" {
				err = csrfInput(true).Render(ctx, w)
				if err != nil {
					slog.Error("cannot render csrf input", "error", err)
				}
			}
		})
	}
}

// ContextMiddleware enriches the request context.
// You should always attach this before adding routes in order to use Apollo effectively.
func (server *Server[state]) ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ctx = context.WithValue(ctx, ctxConfig, server.cfg)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SessionMiddleware returns the Apollo session middleware.
// Attach this before adding routes, if you want to use sessions.
func (server *Server[state]) SessionMiddleware() func(http.Handler) http.Handler {
	if server.sessionStore == nil {
		slog.Warn("Not enabling the SessionMiddleware since there is no SessionStore configured")
		return noop
	}
	return func(next http.Handler) http.Handler {
		gob.Register(core.UserID(0))
		gob.Register(time.Time{})
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Attach session context
			session, err := server.sessionStore.Get(r, cookieUser)
			if err != nil {
				session, err = server.sessionStore.New(r, cookieUser)
				if err != nil {
					slog.Error("Could not create new user cookie", "error", err)
				}
				configureCookie(server.cfg, session)
			}

			ctx = context.WithValue(ctx, ctxSession, session)
			ctx = buildSessionContext(ctx, session)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (server *Server[state]) FeatureFlagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ctx = context.WithValue(ctx, ctxFlags, server.cfg.Features.Flags)
		ctx = context.WithValue(ctx, ctxEnableAll, server.cfg.Features.EnableAll)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Debug is middleware that can be inserted anywhere and will print some useful debug information about the current
// request.
func Debug(printFullRequest bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if printFullRequest {
				slog.Debug("Debug middleware", "request", r)
			} else {
				slog.Debug("Debug middleware", "path", r.URL.Path)
			}

			next.ServeHTTP(w, r)
		})
	}
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
