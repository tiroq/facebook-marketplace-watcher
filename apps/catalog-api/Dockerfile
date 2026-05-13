FROM golang:1.24-alpine AS builder
WORKDIR /build
# Copy internal module
COPY internal/ /internal/
# Copy app source
COPY apps/catalog-api/ ./
# Download dependencies
RUN go mod download
# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o catalog-api ./cmd/catalog-api/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/catalog-api .
EXPOSE 8080
CMD ["./catalog-api"]
