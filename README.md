# metricSource
This repository contains an open-telemetry metrics generator. Metrics are emitted to the grpc endpoint: **0.0.0.0:4317**

# Build the iamge
clone the repo and then run **docker build -t metric-simulator -f /aboslute/path/to/DockerFile .**

# Run the container
docker run -d --name metric-simulator-container metric-simulator

# Sample otel collctor config to receive the emitted metrics

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
