package middlewares

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// LoggingMiddleware returns a gRPC unary interceptor that logs requests and responses.
func LoggingMiddleware(log *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		reqID := uuid.New().String()
		start := time.Now()

		// Add request ID to context
		ctx = context.WithValue(ctx, "request_id", reqID)

		log.Infow("request",
			"request_id", reqID,
			"method", info.FullMethod,
			"request", req,
		)

		// Call the handler
		resp, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			log.Errorw("response error",
				"request_id", reqID,
				"method", info.FullMethod,
				"error", err,
				"duration", duration,
			)
		} else {
			log.Infow("response",
				"request_id", reqID,
				"method", info.FullMethod,
				"response", resp,
				"duration", duration,
			)
		}

		return resp, err
	}
}
