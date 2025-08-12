package rabbitmq

import (
	"context"
	"log"
	"time"
)

func LoggingMiddleware() Middleware {
	return func(ctx context.Context, queue string, body []byte) error {
		log.Printf("Processing message for queue %s: %s", queue, string(body))
		return nil
	}
}

func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(ctx context.Context, queue string, body []byte) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
}

func MetricsMiddleware() Middleware {
	return func(ctx context.Context, queue string, body []byte) error {
		// Add your metrics collection here (e.g., Prometheus)
		return nil
	}
}
