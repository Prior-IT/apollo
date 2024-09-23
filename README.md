# Apollo Web Library

This is the internal Prior-IT full-stack web library code-named "Apollo".
It is open-sourced, primarily to make our own dev-flow easier, but external pull requests are welcome.
The focus of Apollo is to allow developers to very quickly iterate and test out new ideas without having
to worry too much about dependencies or making things work together. Apollo contains everything you need out-of-the-box
and has bootstrapping and utility functions to combine it all together.

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

## TODO
- [x] Core
- [x] Postgres core
- [x] OAuth login
- [x] Organisations
- [x] Permissions
- [ ] EntraID login
- [ ] Magic e-mail login
- [ ] Move account cache to a separate service (so you can use redis for caching while still storing accounts in postgres)
- [ ] Alert component
- [ ] E-mail verification

## Technologies
The following technologies are part of the Apollo tech stack:
- [go's stdlib http package](https://pkg.go.dev/net/http)
- [Chi router](https://github.com/go-chi/chi)
- [sqlc](https://sqlc.dev/)
- [Templ](https://templ.guide/)
- [htmx](https://htmx.org)
- [AlpineJS](https://alpinejs.dev)

