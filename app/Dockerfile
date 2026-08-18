# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency definition
COPY go.mod ./
# If go.sum exists, copy it as well
COPY go.sum* ./

# Copy source code
COPY main.go ./

# Tidy dependencies to generate/verify go.sum
RUN go mod tidy

# Build binary statically
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o http-server-projeto-korp main.go

# Production stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/http-server-projeto-korp .

EXPOSE 8080

ENTRYPOINT ["./http-server-projeto-korp"]
`