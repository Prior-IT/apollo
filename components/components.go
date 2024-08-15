package components

import (
	"embed"
	"os"

	"github.com/prior-it/apollo/server"
)

//go:embed static/*
var EmbedStatic embed.FS

// Serve the Apollo static files using the specified server at the specified endpoint.
// Applications should import the following files in the HTML header:
//
//	<link href="/<endpoint>/apollo.css" rel="stylesheet" defer/>
//
// When using a local version of Apollo, you can set the APOLLO_STATIC_FILES environment
// variable to hot reload library styles.
func ServeStaticFiles[state any](server *server.Server[state], endpoint string) {
	server.StaticFiles(
		endpoint,
		os.Getenv("APOLLO_STATIC_FILES"),
		EmbedStatic,
	)
}
