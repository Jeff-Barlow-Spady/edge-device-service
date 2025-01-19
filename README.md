# Edge Device Service

A microservices-based edge device control system designed for Raspberry Pi, providing GPIO control, authentication, and metrics collection capabilities. Built with cross-platform support for development on x86 and deployment on ARM architectures.

## Architecture Overview

```ascii
┌─────────────────────────────────────────────────────────────────┐
│                        Edge Device Service                       │
│                                                                 │
│  ┌──────────┐    ┌───────────┐    ┌─────────────┐             │
│  │  Caddy   │◄───┤   Auth    │◄───┤    GPIO     │  ┌────────┐ │
│  │ (Proxy)  │    │ Service   │    │  Service    │◄─┤ GPIO   │ │
│  └────┬─────┘    └─────┬─────┘    └──────┬──────┘  │ Pins   │ │
│       │                │                  │         └────────┘ │
│       │                │                  │                    │
│       │                │            ┌─────▼──────┐            │
│       │                │            │  Metrics   │            │
│       │                │            │  Service   │            │
│       │                │            └─────┬──────┘            │
│       │                │                  │                    │
│       │          ┌─────▼──────┐          │                    │
│       └──────────►  PostGIS   ◄──────────┘                    │
│                  │ Database   │                               │
│                  └────────────┘                               │
└─────────────────────────────────────────────────────────────────┘
```

## Development Setup

### Prerequisites

- Docker and Docker Compose
- Python 3.9+
- Git
- VSCode (recommended) or preferred IDE

### Local Development (x86)

1. Clone the repository:
```bash
git clone https://github.com/yourusername/edge-device-service
cd edge-device-service
```

2. Create and activate virtual environment:
```bash
python -m venv venv
source venv/bin/activate  # Linux/macOS
# or
.\venv\Scripts\activate  # Windows
```

3. Install development dependencies:
```bash
pip install -r requirements-dev.txt
```

4. Start development services:
```bash
docker-compose -f docker-compose.dev.yml up -d
```

### ARM Development (Raspberry Pi)

1. Install system dependencies:
```bash
sudo apt-get update
sudo apt-get install -y python3-pip python3-venv
```

2. Clone and setup:
```bash
git clone https://github.com/yourusername/edge-device-service
cd edge-device-service
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

## Build and Deployment

### Building Multi-Architecture Images

We use Docker Buildx for cross-platform builds:

```bash
# Set up buildx
docker buildx create --use

# Build and push multi-arch images
make build-multi-arch
```

### Deployment

1. On Raspberry Pi:
```bash
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
```

2. Monitor services:
```bash
docker-compose ps
docker-compose logs -f
```

## Cross-Platform Considerations

- All images are built for both `linux/amd64` and `linux/arm64` architectures
- GPIO functionality is mocked in x86 development environment
- Database migrations are architecture-agnostic
- Build artifacts are tagged with architecture-specific suffixes
- CI/CD pipeline builds and tests on both architectures
- Development setup automatically detects and uses appropriate architecture

## Services

- **GPIO Service**: Hardware control and state management
- **Auth Service**: JWT-based authentication and authorization
- **Metrics Service**: System metrics collection and monitoring
- **Caddy**: Reverse proxy and TLS termination
- **PostgreSQL/PostGIS**: Spatial-aware data storage

## Contributing

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/my-feature`
3. Commit changes: `git commit -am 'Add my feature'`
4. Push to branch: `git push origin feature/my-feature`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

