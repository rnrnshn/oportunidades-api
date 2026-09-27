FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /oportunidades-api ./cmd/api \
    && CGO_ENABLED=0 GOOS=linux go build -o /create-admin ./cmd/create-admin

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /oportunidades-api /app/oportunidades-api
COPY --from=builder /create-admin /app/create-admin

EXPOSE 8080

ENTRYPOINT ["/app/oportunidades-api"]
