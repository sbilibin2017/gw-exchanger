package middlewares

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func newTestLogger(buf *bytes.Buffer) *zap.SugaredLogger {
	encoderCfg := zap.NewProductionEncoderConfig()
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(buf),
		zapcore.DebugLevel,
	)
	return zap.New(core).Sugar()
}

func TestLoggingMiddleware(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := newTestLogger(buf)

	testCases := []struct {
		name      string
		handler   grpc.UnaryHandler
		expectErr bool
	}{
		{
			name: "successful handler",
			handler: func(ctx context.Context, req any) (any, error) {
				return "ok", nil
			},
			expectErr: false,
		},
		{
			name: "handler returns error",
			handler: func(ctx context.Context, req any) (any, error) {
				return nil, errors.New("handler error")
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interceptor := LoggingMiddleware(logger)

			resp, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{
				FullMethod: "/test/method",
			}, tc.handler)

			if tc.expectErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, "ok", resp)
			}

			logs := buf.String()
			require.Contains(t, logs, "/test/method")
			if tc.expectErr {
				require.Contains(t, logs, "handler error")
			}

			buf.Reset()
		})
	}
}
