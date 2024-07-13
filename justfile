set positional-arguments
export GOOSE_MIGRATION_DIR := "./migrations"

default:
  @just --list --justfile {{justfile()}}

# Build
build:
  @templ generate
  @go build ./...

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

