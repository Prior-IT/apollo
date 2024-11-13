package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/a-h/templ"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/gorilla/sessions"
	"github.com/prior-it/apollo/config"
	"github.com/prior-it/apollo/permissions"
	"github.com/vearutop/statigz"
)

type (
	ErrorHandler    func(apollo *Apollo, err error)
	NotFoundHandler func(apollo *Apollo)
)

type State interface {
	Close(ctx context.Context)
}

type Server[state State] struct {
	mux               *chi.Mux
	state             state
	logger            *slog.Logger
	layout            templ.Component
	errorHandler      ErrorHandler
	permissionService permissions.Service
	sessionStore      sessions.Store
	cfg               *config.Config
}

type (
	Handler[state any]    func(apollo *Apollo, state state) error
	Middleware[state any] func(apollo *Apollo, state state) (context.Context, error)
)

// New creates a new server with the specified state object and configuration.
func New[state State](s state, cfg *config.Config) *Server[state] {
	server := &Server[state]{
		mux:          chi.NewMux(),
		state:        s,
		logger:       slog.Default(),
		layout:       defaultLayout(),
		errorHandler: DefaultErrorHandler,
		cfg:          cfg,
	}

	if len(cfg.App.AuthenticationKey) > 0 && len(cfg.App.EncryptionKey) > 0 {
		server.sessionStore = sessions.NewCookieStore(
			[]byte(cfg.App.AuthenticationKey),
			[]byte(cfg.App.EncryptionKey),
		)
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

func (server *Server[state]) NewApollo(w http.ResponseWriter, r *http.Request) *Apollo {
	apollo := Apollo{
		Writer:      w,
		Request:     r,
		logger:      server.logger,
		layout:      server.layout,
		permissions: server.permissionService,
		store:       server.sessionStore,
		Cfg:         server.cfg,
	}
	apollo.populate()
	return &apollo
}

func (server *Server[state]) handle(handler Handler[state]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apollo := server.NewApollo(w, r)
		err := handler(apollo, server.state)
		if err != nil {
			server.errorHandler(apollo, err)
		}
		_ = r.Body.Close()
	}
}

// ConvertToApolloMiddleware will convert stdlib middleware to Apollo middleware
func ConvertToApolloMiddleware[state State](
	middleware func(w http.ResponseWriter, r *http.Request),
) Middleware[state] {
	return func(apollo *Apollo, _ state) (context.Context, error) {
		middleware(apollo.Writer, apollo.Request)
		return apollo.Context(), nil
	}
}

// Utility function that converts Apollo middleware to a http handler
func (server *Server[state]) HandlerMiddleware(
	middleware Middleware[state],
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apollo := server.NewApollo(w, r)
			ctx, err := middleware(apollo, server.state)
			if err != nil {
				server.errorHandler(apollo, err)
			} else {
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}

func (server *Server[state]) AttachDefaultMiddleware() {
	server.UseStd(
		server.RedirectSlashes,
		middleware.Recoverer,
		middleware.RealIP,
		middleware.RequestID,
		HTTPLogger(server.cfg),

		// @TODO: Add gzip
		middleware.Timeout(
			time.Duration(server.cfg.App.RequestTimeout)*time.Second,
		),
		server.SessionMiddleware(),
		server.CSRFTokenMiddleware(),
		// @TODO: Page caching
		server.ContextMiddleware,
	)
}

// Start a new goroutine that runs the server.
// If no listener is provided, a new TCP listener will be created on the configured host and port.
func (server *Server[state]) Start(ctx context.Context, listener *net.Listener) error {
	// Handle OS signals to cancel the context
	ctxServer, stopSignal := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stopSignal()

	errorCh := make(chan error)
	// Run the actual server
	go func() {
		host := fmt.Sprintf("%v:%v", server.cfg.App.Host, server.cfg.App.Port)
		if listener != nil {
			host = (*listener).Addr().String()
		}
		slog.Info("Starting server", "url", server.cfg.BaseURL(), "host", host)
		var err error
		if listener != nil {
			err = http.Serve(*listener, server)
		} else {
			err = http.ListenAndServe(host, server)
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errorCh <- err
		}
		close(errorCh)
	}()

	var errServer error

loop:
	for {
		select {
		case err := <-errorCh:
			errServer = err
			break loop

		case <-ctxServer.Done():
			slog.Info("Server interrupt received")
			break loop
		}
	}

	ctxShutdown, cancelShutdown := context.WithTimeout(
		ctx,
		time.Duration(server.cfg.App.ShutdownTimeout)*time.Second,
	)

	go func() {
		defer cancelShutdown()

		server.Shutdown(ctxShutdown)
	}()

	<-ctxShutdown.Done()

	return errServer
}

// Shutdown will gracefully release all server resources. You generally don't need to call this manually.
func (server *Server[state]) Shutdown(ctx context.Context) {
	sentryTimeout := max(0, time.Duration(server.cfg.App.ShutdownTimeout-1))
	sentry.Flush(sentryTimeout * time.Second)
	server.state.Close(ctx)
}

// ServeHTTP implements [net/http.Handler].
func (server *Server[state]) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	server.mux.ServeHTTP(writer, request)
}

// UseStd appends a stdlib middleware handler to the middleware stack.
//
// The middleware stack for any server will execute before searching for a matching
// route to a specific handler, which provides opportunity to respond early,
// change the course of the request execution, or set request-scoped values for
// the next Handler.
func (server *Server[state]) UseStd(middlewares ...func(http.Handler) http.Handler) *Server[state] {
	server.mux.Use(middlewares...)
	return server
}

// Use appends an Apollo middleware handler to the middleware stack.
//
// The middleware stack for any server will execute before searching for a matching
// route to a specific handler, which provides opportunity to respond early,
// change the course of the request execution, or set request-scoped values for
// the next Handler.
func (server *Server[state]) Use(
	middlewares ...Middleware[state],
) *Server[state] {
	for _, mi := range middlewares {
		server.mux.Use(server.HandlerMiddleware(mi))
	}
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

// Group attaches another Handler or Router as a subrouter along a routing
// path. It's very useful to split up a large API as many independent routers and
// compose them as a single service. Or to attach an additional set of middleware
// along a group of endpoints, e.g. a subtree of authenticated endpoints.
//
// Note that Group() does NOT return the original server but rather
// a subroute server that only serves routes along the specified Group pattern.
// This simply sets a wildcard along the `pattern` that will continue
// routing at return subroute server. As a result, if you define two Group() routes on
// the exact same pattern, the second group will panic.
func (server *Server[state]) Group(
	pattern string,
) *Server[state] {
	srv := Server[state](*server) //nolint:unconvert // shallow copy
	srv.mux = chi.NewMux()
	server.mux.Mount(pattern, srv.mux)
	return &srv
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
	options *RenderOptions,
) *Server[state] {
	server.mux.Get(pattern, server.handle(func(apollo *Apollo, _ state) error {
		return apollo.RenderPage(component, options)
	}))
	return server
}

// Component adds the route `pattern` that matches a GET http method to render the specified templ component without any layout.
func (server *Server[state]) Component(
	pattern string,
	component templ.Component,
) *Server[state] {
	server.mux.Get(pattern, server.handle(func(apollo *Apollo, _ state) error {
		return apollo.RenderComponent(component)
	}))
	return server
}
