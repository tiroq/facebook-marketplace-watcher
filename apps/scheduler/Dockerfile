FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY internal/ /internal/
COPY apps/scheduler/ ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o scheduler ./cmd/scheduler/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/scheduler .
CMD ["./scheduler"]
