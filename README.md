# Docker Container Metrics Exporter

A Prometheus-compatible metrics exporter for Docker container statistics.

## Features

- Exposes Docker container information and state as Prometheus metrics
- Connects to Docker daemon via Unix socket (`/var/run/docker.sock`)
- Provides metrics for container info, state, and total count
- Configurable port via `PORT` environment variable

## Usage

### Build and Run

```bash
# Build the application
go build -o prometheus-docker-container-exporter

# Run with default port (8080)
./prometheus-docker-container-exporter

# Run with custom port
PORT=9090 ./prometheus-docker-container-exporter
```

### Docker Usage

#### Single Platform Build

```bash
# Build and run with Docker Compose
docker-compose up --build

# Or build the image directly
docker build -t mountainpass/prometheus-docker-container-exporter:latest .

# Run the container
docker run -d \
  --name prometheus-docker-container-exporter \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  mountainpass/prometheus-docker-container-exporter:latest
```

#### Multiplatform Build and Publish

**Build and publish multi-platform images:**
```bash
# Build and push directly to registry (recommended)
docker-compose -f docker-compose.multiplatform.yml build --push

# Or build first, then push separately
docker-compose -f docker-compose.multiplatform.yml build
docker-compose -f docker-compose.multiplatform.yml push
```

**Note**: For pushing to a registry, ensure you're logged in with `docker login` and have appropriate permissions to push to the `mountainpass` namespace.

### Metrics Endpoint

The metrics are available at: `http://localhost:8080/metrics` (or your configured port)

### Exposed Metrics

- `docker_container_info`: Information about Docker containers (labels: id, name, image, status)
- `docker_container_state`: State of Docker containers (1=running, 0=stopped)
- `docker_containers_total`: Total number of Docker containers

### Prometheus Configuration

Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'docker-containers'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 30s
```

## Requirements

- Docker daemon running with accessible socket at `/var/run/docker.sock`
- Go 1.19 or later
- Tested with Docker API Version 1.43
