FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/article-service ./app

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/article-service /article-service
EXPOSE 8000
CMD ["/article-service"]
