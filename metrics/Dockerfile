FROM --platform=$BUILDPLATFORM golang:1.22-alpine as builder

# Add build arguments for cross-compilation
ARG TARGETARCH
ARG TARGETOS
ARG TARGETPLATFORM
ARG BUILDPLATFORM

WORKDIR /build

# Add debugging tools and git
RUN apk add --no-cache git

# Copy module files
COPY go.* ./

# Download dependencies
RUN GOWORK=off go mod download

# Copy source
COPY . .

# Build with proper GOARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -v -o metrics

FROM --platform=$TARGETPLATFORM alpine:latest
WORKDIR /app
COPY --from=builder /build/metrics .
EXPOSE 8081
CMD ["./metrics"]
