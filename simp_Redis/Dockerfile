FROM golang:1.24-alpine AS builder

WORKDIR /app


COPY go.mod go.sum* ./


RUN go mod download

# Copy from the source code
COPY . .

# Building the app
RUN CGO_ENABLED=0 GOOS=linux go build -o simp_Redis .


FROM alpine:latest

# CA certificates for any outbound HTTPS connections
RUN apk --no-cache add ca-certificates

WORKDIR /app


RUN mkdir -p /data

# Copy the binary from the builder stage
COPY --from=builder /app/simp_Redis .


ENV TCP_PORT=6379
ENV DATA_DIR=/data
ENV MAX_MEMORY=128MB
ENV LOG_LEVEL=info


EXPOSE 6379

# Run the application
CMD ["/app/simp_Redis"]