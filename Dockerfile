FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/article-service ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/article-service /app/article-service
COPY --from=builder /app/internal/docs /app/internal/docs
EXPOSE 8000
ENTRYPOINT ["/app/article-service"]
