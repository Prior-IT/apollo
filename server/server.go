package server

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/vearutop/statigz"
)

type (
	ErrorHandler    func(apollo *Apollo, err error)
	NotFoundHandler func(apollo *Apollo)
)

type Server[state any] struct {
	mux          *chi.Mux
	state        state
	logger       *slog.Logger
	layout       templ.Component
	errorHandler ErrorHandler
	isDebug      bool
	useSSL       bool
}

type Handler[state any] func(apollo *Apollo, state state) error

// New creates a new server with the specified state object.
func New[state any](s state) *Server[state] {
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

func (server *Server[state]) WithDebug(debug bool) *Server[state] {
	server.isDebug = debug
	return server
}

func (server *Server[state]) WithSSL(useSSL bool) *Server[state] {
	server.useSSL = useSSL
	return server
}

func (server *Server[state]) handle(handler Handler[state]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apollo := Apollo{
			Writer:  w,
			Request: r,
			logger:  server.logger,
			layout:  server.layout,
			IsDebug: server.isDebug,
			UseSSL:  server.useSSL,
		}
		err := handler(&apollo, server.state)
		if err != nil {
			server.errorHandler(&apollo, err)
		}
		_ = r.Body.Close()
	}
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
	if server.isDebug && len(dir) > 0 {
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
