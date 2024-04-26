# Step 1: Modules caching
FROM golang:1.20-alpine3.16 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:1.20-alpine3.16 as builder
COPY . /app
WORKDIR /app
RUN go build -o /bin/app

# Step 3: Final
FROM scratch
COPY --from=builder /app/config /config
COPY --from=builder /bin/app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app"]