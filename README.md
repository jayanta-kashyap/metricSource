# metricSource
This repository contains an open-telemetry metrics generator
metrics are emitted to the grpc endpoint: 0.0.0.0:4317

# metrics Generated
## 10 different resource sets are generated at a time
## Each resource will have different sample metrics with metric count ranging randomly from 1-10
## Each metric will generate different decimal data points ranging randomly from 1-10


# Build the iamge
docker build -t metric-simulator -f /path/to/DockerFile .

# Run the container
docker run -d --name metric-simulator-container metric-simulator

# Sample otel collctor Config to receive the emitted metrics

```
receivers:
  otlp:
    protocols:
      grpc:
      http:

exporters:
  debug:
    verbosity: basic  # Enables detailed logging of all received data points

service:
  pipelines:
    metrics:
      receivers: [otlp]
      exporters: [debug]
```
