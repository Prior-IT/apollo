# Apollo Web Library

This is the internal Prior-IT full-stack web library code-named "Apollo".

You can get started by downloading the project, installing dependencies, and running the documentation server:
```bash
git clone git@github.com:Prior-IT/apollo.git
just setup
just docs
```

## Running tests
To run all tests once, call `just test`.
If you want to hot reload tests (e.g. while writing or tweaking them), you can use `just devtest`.
You can also run `just devtest ./server/...` or `just test ./server/...` to only run tests for those directories.
Or run `just test -run Users ./...` (and the same for devtest) to only run tests that contain "Users" in their name.


## Technologies
The following technologies are part of the Apollo tech stack:
- [go's stdlib http package](https://pkg.go.dev/net/http)
- [Chi router](https://github.com/go-chi/chi)
- [sqlc](https://sqlc.dev/)
- [Templ](https://templ.guide/)
- [htmx](https://htmx.org)
- [AlpineJS](https://alpinejs.dev)

