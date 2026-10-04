FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/corporate-workspace ./cmd/server
FROM alpine:3.20
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/corporate-workspace /app/corporate-workspace
RUN mkdir -p config data logs backups uploads && chown -R app:app /app
USER app
EXPOSE 8080
ENTRYPOINT ["/app/corporate-workspace"]
