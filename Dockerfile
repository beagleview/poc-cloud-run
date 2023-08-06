from golang:1.20.7-bullseye AS builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

# Build the binary.
RUN go build -o server

FROM debian:bullseye-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*


# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server

EXPOSE 8080

# Run the web service on container startup.
CMD ["/app/server"]
