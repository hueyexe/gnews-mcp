# Contributing

Bug reports and pull requests are welcome.

## Issues

Open an issue with the tool you called, the input, and the result you got.

## Pull requests

- Open PRs against `main`.
- Go 1.26, `golangci-lint` clean, `go test -race ./...` passing.
- Keep output token-efficient: a new tool returns compact markdown, not verbose
  JSON or HTML.
- Update the README tools table if you add a tool.