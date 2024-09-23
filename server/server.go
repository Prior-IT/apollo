package server

import (
	"context"
	"embed"
	"encoding/gob"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/gorilla/sessions"
	"github.com/prior-it/apollo/config"
	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/permissions"
	"github.com/vearutop/statigz"
)

type (
	ErrorHandler    func(apollo *Apollo, err error)
	NotFoundHandler func(apollo *Apollo)
)

type Server[state any] struct {
	mux               *chi.Mux
	state             state
	logger            *slog.Logger
	layout            templ.Component
	errorHandler      ErrorHandler
	permissionService permissions.Service
	sessionStore      sessions.Store
	cfg               *config.Config
}

type Handler[state any] func(apollo *Apollo, state state) error

// New creates a new server with the specified state object and configuration.
func New[state any](s state, cfg *config.Config) *Server[state] {
	server := &Server[state]{
		mux:    chi.NewMux(),
		state:  s,
		logger: slog.Default(),
		layout: defaultLayout(),
		errorHandler: func(apollo *Apollo, err error) {
			apollo.Writer.WriteHeader(http.StatusInternalServerError)
			apollo.Error("[ERROR] Internal server error", "error", err)
			render.PlainText(apollo.Writer, apollo.Request, "internal server error")
		},
		cfg: cfg,
	}

	// Attach default not found handler
	server.WithNotFoundHandler(
		func(apollo *Apollo) {
			apollo.Writer.WriteHeader(http.StatusNotFound)
			render.PlainText(
				apollo.Writer,
				apollo.Request,
				fmt.Sprintf("Page %q not found", apollo.Path()),
			)
		},
	)

	return server
}

// Bootstrap creates a new server and initializes all default systems.
func Bootstrap[state any](s state, cfg *config.Config, staticFS embed.FS) *Server[state] {
	if cfg == nil {
		panic("You need to supply a config.Config value to bootstrap a new server")
	}
	server := New(s, cfg)
	server.AttachDefaultMiddleware()
	server.StaticFiles("/static", "static", staticFS)

	return server
}

func (server *Server[state]) WithErrorHandler(errorHandler ErrorHandler) *Server[state] {
	server.errorHandler = errorHandler
	return server
}

func (server *Server[state]) WithNotFoundHandler(notFoundHandler NotFoundHandler) *Server[state] {
	server.mux.NotFound(server.handle(func(apollo *Apollo, _ state) error {
		notFoundHandler(apollo)
		return nil
	}))
	return server
}

func (server *Server[state]) WithLogger(logger *slog.Logger) *Server[state] {
	server.logger = logger
	return server
}

func (server *Server[state]) WithDefaultLayout(layout templ.Component) *Server[state] {
	server.layout = layout
	return server
}

func (server *Server[state]) WithPermissionService(service permissions.Service) *Server[state] {
	server.permissionService = service
	return server
}

func (server *Server[state]) WithSessionStore(store sessions.Store) *Server[state] {
	server.sessionStore = store
	return server
}

func (server *Server[state]) WithConfig(cfg *config.Config) *Server[state] {
	server.cfg = cfg
	return server
}

func (server *Server[state]) handle(handler Handler[state]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apollo := Apollo{
			Writer:      w,
			Request:     r,
			logger:      server.logger,
			layout:      server.layout,
			permissions: server.permissionService,
			store:       server.sessionStore,
		}
		apollo.populate()
		err := handler(&apollo, server.state)
		if err != nil {
			server.errorHandler(&apollo, err)
		}
		_ = r.Body.Close()
	}
}

func (server *Server[state]) AttachDefaultMiddleware() {
	server.Use(
		middleware.StripSlashes,
		middleware.Recoverer,
		middleware.RealIP,
		middleware.RequestID,
		// @TODO: Add logger
		// @TODO: Add gzip
		middleware.Timeout(
			time.Duration(server.cfg.App.RequestTimeout)*time.Second,
		), // @TODO: Get value from config
		// @TODO: Cookie store
		// @TODO: Load session
		// @TODO: Page caching
		// @TODO: Csrf
		server.ContextMiddleware(),
	)
}

// ContextMiddleware enriches the request context.
// You should always attach this before adding routes in order to use Apollo effectively.
func (server *Server[state]) ContextMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ctx = context.WithValue(ctx, ctxConfig, server.cfg)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SessionMiddleware returns the Apollo session middleware.
// Attach this before adding routes, if you want to use sessions.
func (server *Server[state]) SessionMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if server.sessionStore == nil {
			panic("You need to configure the session store before attaching Apollo middleware")
		}
		gob.Register(core.UserID(0))
		gob.Register(time.Time{})
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Attach session context
			session, err := server.sessionStore.Get(r, cookieUser)
			if err != nil {
				slog.Error("Error while retrieving active session", "error", err)
				return
			}

			ctx = context.WithValue(ctx, ctxSession, session)

			loggedIn, ok := session.Values[sessionLoggedIn].(bool)
			ctx = context.WithValue(ctx, ctxLoggedIn, ok && loggedIn)

			isAdmin, ok := session.Values[sessionIsAdmin].(bool)
			ctx = context.WithValue(ctx, ctxIsAdmin, ok && isAdmin)

			userName, ok := session.Values[sessionUserName].(string)
			if ok {
				ctx = context.WithValue(ctx, ctxUserName, userName)
			}

			userID, ok := session.Values[sessionUserID].(core.UserID)
			if ok {
				ctx = context.WithValue(ctx, ctxUserID, userID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (server *Server[state]) Start(ctx context.Context) error {
	// Handle OS signals to cancel the context
	ctxServer, cancelServer := context.WithCancel(ctx)
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		cancelServer()
	}()

	// Run the actual server
	errServer := func() error {
		host := fmt.Sprintf("%v:%v", server.cfg.App.Host, server.cfg.App.Port)
		slog.Info("Server started", "url", server.cfg.BaseURL(), "host", host)
		err := http.ListenAndServe(host, server)

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}()

	<-ctxServer.Done()

	slog.Info("Shutting down gracefully...")
	// @TODO: Clean up any leftover resources

	// ctxShutdown, cancelShutdown := context.WithTimeout(ctx, 5*time.Second)
	// defer cancelShutdown()

	slog.Info("Server shut down")
	return errServer
}

// ServeHTTP implements [net/http.Handler].
func (server *Server[state]) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	server.mux.ServeHTTP(writer, request)
}

// Use appends a middleware handler to the middleware stack.
//
// The middleware stack for any server will execute before searching for a matching
// route to a specific handler, which provides opportunity to respond early,
// change the course of the request execution, or set request-scoped values for
// the next Handler.
func (server *Server[state]) Use(middlewares ...func(http.Handler) http.Handler) *Server[state] {
	server.mux.Use(middlewares...)
	return server
}

// Handle adds the route `pattern` that matches any http method to
// execute the `handler` [net/http.Handler].
func (server *Server[state]) Handle(pattern string, handler http.Handler) *Server[state] {
	server.mux.Handle(pattern, handler)
	return server
}

// StaticFiles serves all files in the `dir` directory or the `fs` FileSystem at the `pattern` url.
// In debug mode, assets will be loaded from disk to support hot-reloading.
// In production mode, assets will be gzipped and embedded in the executable instead.
// Debug mode hot-reloading will be disabled if dir is set to the empty string.
// Filesystems will ignore `/static` folders and instead directly target the files inside. So if your
// filesystem has a file "/static/file.txt", you can get it directly with "/file.txt".
//
// Example:
//
//	server.StaticFiles("/assets/", "./static/", assetsFS)
func (server *Server[state]) StaticFiles(pattern string, dir string, files fs.ReadDirFS) {
	if server.cfg.App.Debug && len(dir) > 0 {
		server.Handle(
			pattern+"*",
			http.StripPrefix(pattern,
				http.FileServer(http.Dir(dir)),
			),
		)
	} else {
		server.Handle(
			pattern+"*",
			http.StripPrefix(pattern,
				middleware.NoCache(
					statigz.FileServer(files, statigz.EncodeOnInit, statigz.FSPrefix("static")),
				),
			),
		)
	}
}

// Get adds the route `pattern` that matches a GET http method to execute the `handlerFn` HandlerFunc.
func (server *Server[state]) Get(
	pattern string,
	handlerFn func(apollo *Apollo, state state) error,
) *Server[state] {
	server.mux.Get(pattern, server.handle(handlerFn))
	return server
}

// Post adds the route `pattern` that matches a POST http method to execute the `handlerFn` http.HandlerFunc.
func (server *Server[state]) Post(
	pattern string,
	handlerFn func(apollo *Apollo, state state) error,
) *Server[state] {
	server.mux.Post(pattern, server.handle(handlerFn))
	return server
}

// Put adds the route `pattern` that matches a POST http method to execute the `handlerFn` http.HandlerFunc.
func (server *Server[state]) Put(
	pattern string,
	handlerFn func(apollo *Apollo, state state) error,
) *Server[state] {
	server.mux.Put(pattern, server.handle(handlerFn))
	return server
}

// Delete adds the route `pattern` that matches a POST http method to execute the `handlerFn` http.HandlerFunc.
func (server *Server[state]) Delete(
	pattern string,
	handlerFn func(apollo *Apollo, state state) error,
) *Server[state] {
	server.mux.Delete(pattern, server.handle(handlerFn))
	return server
}

// Page adds the route `pattern` that matches a GET http method to render the specified templ component in the default layout.
func (server *Server[state]) Page(
	pattern string,
	component templ.Component,
) *Server[state] {
	server.mux.Get(pattern, server.handle(func(apollo *Apollo, _ state) error {
		return apollo.RenderPage(component)
	}))
	return server
}
