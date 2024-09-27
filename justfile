set positional-arguments

# Dependencies
templ := "github.com/a-h/templ/cmd/templ@v0.2.747"
air := "github.com/air-verse/air@v1.52.3"
sqlc := "github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0"
goose := "github.com/pressly/goose/v3/cmd/goose@v3.21.1"
tailwind := "tailwindcss@3.4.10"

default:
  @just --list --justfile {{justfile()}}

# Generate auxiliary files
generate:
  @go run {{templ}} generate -include-version=false
  @go run {{sqlc}} generate -f ./postgres/sqlc.yaml
  @npx {{tailwind}} -i components/input.css -o components/static/apollo.css -m

# Continuously generate auxiliary files on every file save
dev:
  @go run {{air}} \
    --build.bin="" \
    --build.cmd="just build" \
    --build.pre_cmd="just generate" \
    --build.include_ext="go,templ,sql" \
    --build.exclude_regex="_templ.go" \
    --build.exclude_dir="migrations,postgres/internal" \
    --build.kill_delay "5s" \
    --build.send_interrupt "true" \
    --build.stop_on_error "true" \
    --misc.clean_on_exit "true" \
    --c "/dev/null"

# Build the library
build:
  go build ./...

# Run a documentation server
docs port="8080":
  @echo Open your browser on http://localhost:{{port}}/
  @echo Apollo-specific documentation is on http://localhost:{{port}}/pkg/github.com/prior-it/apollo
  @godoc -http=:{{port}}

# Run all linters
lint:
  golangci-lint run --allow-parallel-runners

# Run tests, you can optionally provide a filter, e.g. "just test ./tests/..." or "just test -run Users ./tests"
test *args="./...":
  go test -parallel {{ num_cpus() }} -coverprofile=coverage.out $@
  go tool cover -html=coverage.out -o coverage.html

# Run tests with hot-reloading, you can optionally provide a filter, e.g. "./tests/..."
devtest *args="./...":
  @go run {{air}} \
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

# Create a new postgres migration with the specified name
newmigration name:
  @GOOSE_MIGRATION_DIR="./postgres/migrations" go run {{goose}} create {{name}} sql

# Download and install all required cli tools and project dependencies
setup:
  go mod tidy
  go mod download
  go mod verify

# Examine code for known issues
vet:
  go vet ./...
