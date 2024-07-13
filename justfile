set positional-arguments
export GOOSE_MIGRATION_DIR := "./migrations"

default:
  @just --list --justfile {{justfile()}}

# Build the library
build:
  @templ generate -include-version=false
  @go build ./...

# Run a documentation server
docs port="8080":
  @echo Open your browser on http://localhost:{{port}}/
  @echo Apollo-specific documentation is on http://localhost:{{port}}/pkg/github.com/prior-it/apollo
  @godoc -http=:{{port}}

# Run all linters
lint:
  @golangci-lint run --allow-parallel-runners

# Run tests, you can optionally provide a filter, e.g. "just test ./tests/..." or "just test -run Users ./tests"
test *args="./...":
  go test -parallel {{ num_cpus() }} -coverprofile=coverage.out $@
  go tool cover -html=coverage.out -o coverage.html

# Run tests with hot-reloading, you can optionally provide a filter, e.g. "./tests/..."
devtest *args="./...":
  @air \
    -build.cmd="just test {{args}}" \
    -build.include_ext="go" \
    -build.exclude_regex="_templ.go" \
    -build.exclude_dir="migrations" \
    -build.stop_on_error="true" \
    -build.bin="" \
    -screen.clear_on_rebuild="true" \
    -c "/dev/null"

# Run fuzzing tests, you must provide exactly one package to test, e.g. "just fuzz ./tests"
fuzz package:
	@go test -parallel {{ num_cpus() }} -fuzz=Fuzz -fuzztime 30s {{package}}

# Create a new migration with the specified name
migration name:
  goose create {{name}} sql

# Download and install all required cli tools and project dependencies
setup:
  # CLI tools
  go install github.com/a-h/templ/cmd/templ@latest
  go install github.com/air-verse/air@latest
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  go install github.com/pressly/goose/v3/cmd/goose@latest
  pnpm install -g tailwindcss
  # Dependencies
  go mod tidy
  go mod download
  go mod verify

