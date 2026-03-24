# Loaner

A loan repayment calculation service.

## Prerequisites

- Go 1.26+
- Docker
- [Buf CLI](https://buf.build/docs/installation)
- [golang-migrate](https://github.com/golang-migrate/migrate) (`brew install golang-migrate`)

## Getting Started

Start Postgres:
```bash
docker compose up -d
```

Run migrations:
```bash
make migrate-up
```

Start the server:
```bash
make run
```

Run tests
```bash
make test
```

## API

The service exposes both HTTP (port 8080) and gRPC (port 9090).

HTTP:
```bash
curl "http://localhost:8080/v1/loan/repayment?loan_amount=100000&annual_interest_rate=8&num_payments=360"
```

gRPC (via buf curl):
```bash
buf curl --protocol grpc --http2-prior-knowledge \
  http://localhost:9090/loan.v1.LoanService/CalculateRepayment \
  -d '{"loanAmount":"100000","annualInterestRate":"8","numPayments":360}'
```

Response:
```json
{
  "monthlyRepayment": "567.79",
  "calculatedAt": "2026-03-24T15:08:38.052732853Z"
}
```

## Docker

Build and run everything (Postgres + migrations + server):
```bash
docker compose up --build
```

Or just build the image:
```bash
make docker-build
```

## Kubernetes (Kind)

```bash
# Build the docker image
docker build -t loaner:latest .

# Load the image to Kind
kind load docker-image loaner:latest --name localdev

# Apply
kubectl apply -f deploy/postgres.yaml
kubectl apply -f deploy/migration.yaml
kubectl apply -f deploy/loaner.yaml
```

HTTP is available on NodePort 30880, gRPC on 30890.

To tear down and redeploy from scratch:
```bash
kubectl delete -f deploy/loaner.yaml
kubectl delete -f deploy/migration.yaml
kubectl delete -f deploy/postgres.yaml
kubectl delete pvc postgres-storage-postgres-0

docker build -t loaner:latest .
kind load docker-image loaner:latest --name localdev

kubectl apply -f deploy/postgres.yaml
kubectl apply -f deploy/migration.yaml
kubectl apply -f deploy/loaner.yaml
```

## Proto code-gen

Regenerate gRPC and gateway code from proto definitions:
```bash
buf dep update
buf generate
go mod tidy
```

## Notes

- **String fields for money** - proto uses `string` (not `float`/`double`) for loan amounts and rates to avoid precision loss at the API boundary.
- **shopspring/decimal** - all arithmetic uses exact decimal math, not `float64`. A rounding error on a 30-year mortgage compounds into real money.
- **NUMERIC in Postgres** - matches the decimal library, no precision loss in storage either.
- **Percentage input** - the API accepts `5.5` meaning 5.5%

## Docs

Browse Go documentation locally:
```bash
go install golang.org/x/pkgsite/cmd/pkgsite@latest
pkgsite -http=:6060
```
Then open http://localhost:6060/github.com/mrprofessor/loaner

## Project Structure

```
loaner/
├── cmd/server/main.go          # entrypoint
├── proto/loan/v1/              # protobuf defs
├── gen/loan/v1/                # generated gRPC + gateway code
├── internal/
│   ├── handler/                # gRPC handlers
│   ├── service/                # Business stuff (PMT)
│   ├── store/                  # Database layer
│   ├── config/                 # config
│   └── middleware/             # middlewares (logging, recovery)
├── migrations/                 # SQL schema migrations
├── deploy/                     # kubernetes
├── Dockerfile
├── docker-compose.yml
└── Makefile
```
