package main

import (
	"context"
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

func generateRandomFloat() float64 {
	return float64(rand.Intn(500)) / 100.0 // Generate random float between 0.00 and 5.00
}

func generateMetrics(ctx context.Context, resourceName string) {
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint("0.0.0.0:4317"),
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

	numMetrics := rand.Intn(10) + 1 // Random number of metrics between 1 and 10
	for i := 1; i <= numMetrics; i++ {
		metricName := strings.ToLower(fmt.Sprintf("metric-%d", i))

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

		numDataPoints := rand.Intn(10) + 1
		for j := 1; j <= numDataPoints; j++ {
			value := generateRandomFloat()
			gauge.Record(ctx, value)
			log.Printf("Resource %s: Recorded gauge %s=%f", resourceName, metricName, value)

			counter.Add(ctx, value)
			log.Printf("Resource %s: Recorded counter %s=%f", resourceName, metricName, value)

			histogram.Record(ctx, value)
			log.Printf("Resource %s: Recorded histogram %s=%f", resourceName, metricName, value)
		}
	}
}

func main() {
	defer func() {
		if rcvErr := recover(); rcvErr != nil {
			log.Printf("Recovered from panic with error: [%v]", rcvErr)
			debug.PrintStack()
		}
		log.Println("Shutting down application gracefully.")
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resourceNames := []string{"Resource_A", "Resource_B", "Resource_C", "Resource_D", "Resource_E", "Resource_F", "Resource_G", "Resource_H"}

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
