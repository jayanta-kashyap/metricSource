package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func generateRandomFloat() float64 {
	// Generate random float between 0.00 and 5.00
	return float64(rand.Intn(500)) / 100.0
}

func generateMetrics(ctx context.Context, resourceName string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Configure gRPC exporter
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint("0.0.0.0:4317"),
		otlpmetricgrpc.WithInsecure(), // Ensure plain gRPC is used
	)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}
	defer exporter.Shutdown(ctx)

	// Set up resource attributes for this resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("otel-metrics-generator"),
			semconv.ServiceVersionKey.String("v1.0.0"),
			// Resource-specific attribute
			semconv.DeploymentEnvironmentKey.String(resourceName),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource for %s: %v", resourceName, err)
	}

	// Create a metric provider specific to this resource
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	)
	defer meterProvider.Shutdown(ctx)

	// Set global meter provider
	otel.SetMeterProvider(meterProvider)

	// Create a Meter from the provider
	meter := meterProvider.Meter(fmt.Sprintf("meter-%s", resourceName))

	// Generate a random number of metrics for this resource
	numMetrics := rand.Intn(10) + 1 // Random number of metrics between 1 and 10

	// Generate random metrics and data points
	for i := 1; i <= numMetrics; i++ {
		// Create metric name as "Metric-X"
		metricName := fmt.Sprintf("Metric-%d", i)

		// Create a Gauge instrument for this metric
		gauge, err := meter.Float64Gauge(metricName)
		if err != nil {
			log.Fatalf("failed to create gauge for %s: %v", metricName, err)
		}

		// Generate a random number of data points for this metric
		numDataPoints := rand.Intn(10) + 1 // Random number of data points between 1 and 10

		// Record the data points for this metric
		for j := 1; j <= numDataPoints; j++ {
			value := generateRandomFloat()
			// Record the metric value
			gauge.Record(ctx, value)
			log.Printf("Resource %s: Recorded metric %s=%f", resourceName, metricName, value)
		}
	}
}

func main() {
	// Set up context for graceful shutdown
	ctx := context.Background()

	// Define resource names
	resourceNames := []string{"Resource_A", "Resource_B", "Resource_C", "Resource_D", "Resource_E", "Resource_F", "Resource_G", "Resource_H", "Resource_I", "Resource_J"}

	// Use WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Run an infinite loop to keep generating metrics indefinitely
	for {
		// Start metric generation for each resource
		for _, resourceName := range resourceNames {
			wg.Add(1)
			go generateMetrics(ctx, resourceName, &wg)
		}

		// Wait for all metric generation to complete
		// Since we're running indefinitely, it will restart the loop after each iteration
		wg.Wait()

		// Optionally, sleep for a small duration to avoid overwhelming the system with metric generation
		// Adjust the sleep duration to your requirements
		time.Sleep(1 * time.Second) // sleep for 1 second before generating metrics again
	}
}
