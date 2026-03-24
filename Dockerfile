FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o loaner cmd/server/main.go

FROM gcr.io/distroless/static
COPY --from=builder /app/loaner /loaner
EXPOSE 8080 9090
ENTRYPOINT ["/loaner"]
