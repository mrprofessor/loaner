---
theme: dark
author: Rudra
date: 2026-04-29
paging: "%d / %d"
---

# Loaner

A Loan Repayment Calculation Service

*Go · gRPC · grpc-gateway · Postgres · Kubernetes*

---

# The Problem

Calculate monthly loan repayments given a loan amount, interest rate, and number of payments.
Use the PMT formula. Persist every request and response to Postgres.

**What I built:**

- Go service exposing gRPC + REST (grpc-gateway)
- Exact decimal math at every layer (no floats)
- Dockerized with K8s deployment to a local Kind cluster

---

# PMT Formula

The standard amortization formula (same as Excel's `PMT`):

```
M = P × r(1+r)^n / ((1+r)^n - 1)
```

| Symbol | Meaning                              |
|--------|--------------------------------------|
| **P**  | Principal (loan amount)              |
| **r**  | Monthly interest rate (annual / 12)  |
| **n**  | Number of payments                   |
| **M**  | Monthly repayment                    |

Edge case: when `r = 0`, simply `M = P / n`

---

# Architecture

```
                ┌──────────────┐
  curl/browser ▶│  HTTP :8080  │  (grpc-gateway)
                └──────┬───────┘
                       │ in-process
                ┌──────▼───────┐
  gRPC clients ▶│  gRPC :9090  │
                └──────┬───────┘
                       │
                ┌──────▼───────┐
                │   Handler    │  validation + response mapping
                └──────┬───────┘
                       │
                ┌──────▼───────┐
                │   Service    │  PMT calculation (pure logic)
                └──────┬───────┘
                       │
                ┌──────▼───────┐
                │    Store     │  Postgres persistence
                └──────────────┘
```

---

# gRPC + grpc-gateway

**Single proto definition** drives both protocols:

```protobuf
service LoanService {
    rpc CalculateRepayment(CalculateRepaymentRequest)
        returns (CalculateRepaymentResponse) {
        option (google.api.http) = { get: "/v1/loan/repayment" };
    }
}
```

- **One source of truth** - no API drift between protocols(grpc/rest)

---

# Precision at each layer

Every layer uses **exact decimal math** - no floats anywhere.

| Layer          | Approach              | Why                                    |
|----------------|-----------------------|----------------------------------------|
| **Proto/API**  | `string` fields       | Avoids `double` precision loss         |
| **Go**         | `shopspring/decimal`  | Arbitrary-precision decimal arithmetic |
| **Postgres**   | `NUMERIC(20,6)`       | Exact decimal storage, no rounding     |

```go
// No float64 in sight
monthlyPayment := service.CalcPMT(loanAmount, annualRate, req.NumPayments)
```

---

# The Calculation

// refer code

---

# Persistence

Every calculation is stored as an **audit record**:

```sql
CREATE TABLE loan_repayment (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    loan_amount       NUMERIC(20, 6) NOT NULL,
    annual_rate       NUMERIC(10, 6) NOT NULL,
    num_payments      INTEGER        NOT NULL,
    monthly_repayment NUMERIC(20, 6) NOT NULL,
    calculated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
```

- **Audit trail** - prove what was calculated and when
- **UUID primary keys** - no sequential ID leakage

---

# Testing

```go
tests := []struct {
    name        string
    principal   string
    annualRate  string
    numPayments int32
    want        string
}{
    {"30-year mortgage", "100000", "5.5", 360, "567.79"},
    {"one year loan",    "10000",  "10",  12,  "879.16"},
    {"zero interest",    "12000",  "0",   12,  "1000.00"},
    {"single payment",   "5000",   "6",   1,   "5025.00"},
}
```

Pure function → no DB, no mocks, deterministic results.

---

# Deployment

**Docker Compose** - one command local dev:

```bash
docker compose up --build
```

**Kubernetes (Kind)** - production-like:

| Resource     | Kind                  | Purpose                    |
|--------------|-----------------------|----------------------------|
| Postgres     | StatefulSet + PVC     | Persistent storage         |
| Migrations   | Job                   | Schema setup (runs once)   |
| Loaner       | Deployment (3 replicas) | Application              |
| Services     | NodePort              | HTTP :30880, gRPC :30890   |

---

# Project Structure

```
loaner/
├── cmd/server/main.go        # entrypoint, wiring
├── proto/loan/v1/            # protobuf definitions
├── gen/loan/v1/              # generated gRPC + gateway
├── internal/
│   ├── handler/              # gRPC handler (validation)
│   ├── service/              # business logic (PMT)
│   ├── store/                # database layer
│   ├── config/               # env-based config
│   └── middleware/           # logging, panic recovery
├── migrations/               # SQL schema
├── deploy/                   # k8s manifests
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

---

# Trade-offs

| Decision                          | Trade-off                              |
|-----------------------------------|----------------------------------------|
| `shopspring/decimal` over float64 | Slower, but the only option really     |
| gRPC + gateway over pure REST     | More tooling, but type-safe contracts  |
| Persist every request             | Storage cost, but full audit trail     |
| `string` proto fields             | Less ergonomic, but no precision loss  |
| K8s StatefulSet for Postgres      | Hosting a DB can be a lot of work      |

---

# What I'd Add for Production

- **Auth** - JWT/OAuth2 middleware on the gateway
- **RBAC** - Support differnt types of users
- **Secrets management** - Vault or K8s Secrets
- **Health probes** - readiness/liveness for K8s
- **Observability** - Prometheus metrics, OpenTelemetry tracing
- **Rate limiting** - protect against abuse

---

# Extensibility: Adding Endpoints

Adding a new endpoint requires the following:

1. Define the RPC + HTTP annotation in the proto file
2. `buf generate` to regenerate Go code
3. Implement the handler method

Only adding a whole new *service* would need a new `Register` call.

---

# Extensibility: Calculation Types

Adding new loan types without touching existing code:

```go
type Calculator interface {
    Calculate(principal, rate decimal.Decimal, n int32) decimal.Decimal
}

type AmortizingCalc struct{}   // current PMT
type BalloonCalc struct{}      // lump sum at end
type InterestOnlyCalc struct{} // interest-only period
```

The handler routes to the right calculator based on a `loan_type` parameter.

---

# Demo

```bash
curl "http://localhost:8080/v1/loan/repayment?\
  loan_amount=100000&annual_interest_rate=8&num_payments=360"
```

```json
{
  "monthlyRepayment": "733.76",
  "calculatedAt": "2026-04-29T..."
}
```

---

# Questions?

**github.com/mrprofessor/loaner**
