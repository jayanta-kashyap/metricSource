package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var exporterEndpoint string // Global variable to hold the OTLP exporter endpoint

func init() {
	// Define the flag for the exporter endpoint
	flag.StringVar(&exporterEndpoint, "exporter-endpoint", "0.0.0.0:4317", "The endpoint for the OTLP exporter")
}

// generateRandomFloat generates a random float between 0.00 and 5.00
func generateRandomFloat() float64 {
	return float64(rand.Intn(1500000)) / 100.0 // Generate random float between 0.00 and 15000.00
}

func generateMetrics(ctx context.Context, resourceName string) {
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(exporterEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		log.Printf("Error creating exporter for %s: %v", resourceName, err)
		return
	}
	defer exporter.Shutdown(ctx)

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("otel-metrics-generator"),
			semconv.ServiceVersionKey.String("v1.0.0"),
			semconv.DeploymentEnvironmentKey.String(resourceName),
		),
	)
	if err != nil {
		log.Printf("Error creating resource for %s: %v", resourceName, err)
		return
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	)
	defer meterProvider.Shutdown(ctx)

	otel.SetMeterProvider(meterProvider)
	meter := meterProvider.Meter(strings.ToLower("meter-" + resourceName))

	numMetrics := rand.Intn(5) + 3 // Random number of metrics between 3 and 7

	// Define custom histogram boundaries
	boundaries := []float64{0.5, 1.0, 2.5, 5.0, 10.0, 100.0, 1000.0, 10000.0}

	for i := 1; i <= numMetrics; i++ {
		var metricName string
		// Define metric names based on resource names
		switch resourceName {
		case "web-service-a", "web-service-b":
			metricName = fmt.Sprintf("%s-http_request_duration_seconds", resourceName)
		case "order-service":
			metricName = fmt.Sprintf("%s-order_count", resourceName)
		case "inventory-service":
			metricName = fmt.Sprintf("%s-db_query_duration_seconds", resourceName)
		case "user-service":
			metricName = fmt.Sprintf("%s-http_requests_total", resourceName)
		case "payment-service":
			metricName = fmt.Sprintf("%s-payment_processing_time_seconds", resourceName)
		case "notification-service":
			metricName = fmt.Sprintf("%s-queue_length", resourceName)
		case "database-service":
			metricName = fmt.Sprintf("%s-db_query_duration_seconds", resourceName)
		default:
			metricName = fmt.Sprintf("%s-metric-%d", resourceName, i)
		}

		gauge, err := meter.Float64Gauge(metricName)
		if err != nil {
			log.Printf("Error creating gauge for %s: %v", metricName, err)
			return
		}

		counter, err := meter.Float64Counter(metricName)
		if err != nil {
			log.Printf("Error creating counter for %s: %v", metricName, err)
			return
		}

		histogram, err := meter.Float64Histogram(metricName)
		if err != nil {
			log.Printf("Error creating histogram for %s: %v", metricName, err)
			return
		}

		numDataPoints := rand.Intn(20) + 5 // Random number of data points between 5 and 25

		// Initialize the bucket counts for each defined boundary
		bucketCounts := make(map[string]int)

		for j := 1; j <= numDataPoints; j++ {
			// Generate a distinct value for each histogram bucket
			value := generateRandomFloat()

			// Generate a unique timestamp by adding a millisecond offset
			uniqueTimestamp := time.Now().Add(time.Duration(j) * time.Millisecond)

			// Record metrics with unique timestamps
			gauge.Record(ctx, value)
			log.Printf("Resource %s: Recorded gauge %s=%f at %v", resourceName, metricName, value, uniqueTimestamp)

			counter.Add(ctx, value)
			log.Printf("Resource %s: Recorded counter %s=%f at %v", resourceName, metricName, value, uniqueTimestamp)

			// Record value in histogram and categorize it into the appropriate bucket
			histogram.Record(ctx, value)
			log.Printf("Resource %s: Recorded histogram %s=%f at %v", resourceName, metricName, value, uniqueTimestamp)

			// Manually calculate which bucket the value belongs to and store it with a unique label
			for idx, boundary := range boundaries {
				// We can assign the value to each bucket with its own unique identifier
				bucketKey := fmt.Sprintf("le%d-%f", idx+1, boundary)
				if value <= boundary {
					// Add a random number between 1 and 5 to simulate variability in the bucket count
					bucketCounts[bucketKey] += rand.Intn(5) + 1
					break
				}
			}
			// If the value is larger than the highest boundary, assign it to the last bucket
			if value > boundaries[len(boundaries)-1] {
				bucketCounts[fmt.Sprintf("le%d-%f", len(boundaries)+1, boundaries[len(boundaries)-1])] += rand.Intn(5) + 1
			}

			// Add a small delay between each record
			time.Sleep(50 * time.Millisecond)
		}

		// Log bucket counts for visibility with unique values per bucket
		for bucket, count := range bucketCounts {
			log.Printf("Resource %s: Histogram %s, bucket %s: count=%d", resourceName, metricName, bucket, count)
		}
	}
}

func main() {
	// Parse the flags to set the exporter endpoint
	flag.Parse()

	// Ensure the exporter endpoint is printed and used
	log.Printf("Using EXPORTER_ENDPOINT: %s", exporterEndpoint)

	defer func() {
		if rcvErr := recover(); rcvErr != nil {
			log.Printf("Recovered from panic with error: [%v]", rcvErr)
			debug.PrintStack()
		}
		log.Println("Shutting down application gracefully.")
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resourceNames := []string{
		"web-service-a", "web-service-b", "order-service",
		"inventory-service", "user-service", "payment-service",
		"notification-service", "database-service",
	}

	// Launch metric generation for each resource in separate goroutines
	for _, resourceName := range resourceNames {
		go func(resName string) {
			for {
				select {
				case <-ctx.Done():
					log.Printf("Stopping metric generation for %s", resName)
					return
				default:
					generateMetrics(ctx, resName)
					time.Sleep(1 * time.Second) // Simulate a delay between metric generations
				}
			}
		}(resourceName)
	}

	// Channel to capture OS signals for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	s := <-c
	log.Printf("Signal %s received, shutting down...", s)
	cancel()                    // Notify all goroutines to stop
	time.Sleep(2 * time.Second) // Give goroutines time to exit gracefully
	log.Println("Application stopped.")
}
