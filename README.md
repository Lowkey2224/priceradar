# priceradar
PriceRadar is a concurrent e-commerce price tracking system written in Go. Features a Go-Chi REST API, PostgreSQL storage via sqlc, a goroutine worker pool for web scraping, and a Cobra CLI. Deployed on Neon &amp; Render

This will be my project for learning go

## Local development

Only Docker and `make` are required — the Go toolchain, golangci-lint, govulncheck
and gotestsum live in the `go` container (`Dockerfile.dev`), pinned to the same
versions as the CI pipeline.

```sh
make image   # build the toolchain image once
make help    # list all targets
make qa      # tidy, fmt, build, lint, vuln, test - like CI
make run     # run the application against the dev database
make shell   # shell inside the container
```

Without `make` on the host — Windows, for example — use the VS Code tasks
(Ctrl+Shift+B, or Command Palette → "Run Task"). They call the same targets
inside the container, so the commands are defined only once. On Linux and macOS
the tasks go through `make` so that the container gets your UID; on Windows they
call Docker directly, where Docker Desktop maps the ownership anyway.

`make test` uses the separate `db-test` service (tmpfs, port 5433). The dev
database on port 5432 is not touched.

The container runs as the calling UID, so generated files belong to you and not
to root. Module and build caches live in named volumes; `make clean` deletes
them along with the dev database.
