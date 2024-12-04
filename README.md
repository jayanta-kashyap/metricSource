# metricSource
This repository contains an open-telemetry metrics generator. Metrics are emitted to the grpc endpoint: **0.0.0.0:4317** by default.

# Run the metricSource
clone the repo and then run ```go run cmd/metric-simulator/main.go -exporter-endpoint=<YOUR_EXPORTER_EDNPOINT>``` to send metrics to a custom endpoint.

# Build the metricSource image
clone the repo and then run ```docker build -t metric-simulator -f ./DockerFile .```

# Sample otel collector config to receive the emitted metrics

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
