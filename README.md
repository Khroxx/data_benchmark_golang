# Data Benchmark Golang

Go backend for the benchmark comparison project. It exposes a benchmark endpoint that generates payloads in different formats and returns timing data as JSON.

## Clone

SSH:

```bash
git clone git@github.com:Khroxx/data_benchmark_golang.git
```

HTTPS:

```bash
git clone https://github.com/Khroxx/data_benchmark_golang.git
```

## Endpoints

- `GET /ping`
- `GET /api/golang/benchmark`

Supported query params:

- `type=flat-json | nested-json | csv | blob`
- `size` or `sizeKb`
- `runs`

## Local development

Run the service:

```bash
go run .
```

Build the service:

```bash
go build ./...
```

The server listens on port `8080`.
