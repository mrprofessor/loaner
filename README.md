# Loaner

A loan repayment calculation service built with Go, gRPC-Gateway, and PostgreSQL.

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Buf CLI](https://buf.build/docs/installation)

## Generate Protos

Generate the gRPC and HTTP gateway code from the proto definitions:

```bash
buf dep update
buf generate
go mod tidy
```

This produces the generated Go code in `gen/loan/v1/`.
