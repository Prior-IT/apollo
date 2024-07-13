/*
Package server provides a HTTP server implementation for the Apollo stack.
Handlers in this stack take an application-specific state object (used for dependency injection)
and a [Apollo] object which contains a lot of utility functions.

Basic example:

	import (
		"log"
		"net/http"

		"myapp/components"
		"myapp/handlers"
		"myapp/middleware"
		"myapp/state"

		"github.com/prior-it/apollo/server"
	)

	func main() {
		// Create server
		state := state.New()
		server := server.New(state).
			WithLogger(state.Logger).
			WithDebug(true)

		// Attach middleware
		server.Use(
			middleware.RequestID,
			middleware.Recoverer,
			middleware.Timeout(5 * time.Second),
		)

		// Attach routes
		server.StaticFiles("/static/", "./assets/", staticFS)
		server.Get("/", Home).
			Get("/ping", handlers.Ping).
			Post("/login", handlers.DoLogin)

		// Run server
		log.Fatal(http.ListenAndServe("localhost:8080", server))
	}

	func Home(apollo *server.Apollo, _ *state.State) error {
		return apollo.RenderPage(components.HomePage())
	}
*/
package server
