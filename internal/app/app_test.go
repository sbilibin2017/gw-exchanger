package app

import (
	"bytes"
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPrintBuildInfo(t *testing.T) {
	// Подменяем вывод логгера на буфер
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(nil) // возвращаем стандартный вывод
	}()

	// Устанавливаем тестовые значения build переменных
	buildVersion = "1.0.0"
	buildCommit = "abc123"
	buildDate = "2025-09-23"

	PrintBuildInfo()

	output := buf.String()
	expectedSubstrings := []string{
		"1.0.0",
		"abc123",
		"2025-09-23",
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(output, substr) {
			t.Errorf("Ожидалось, что лог содержит '%s', но не содержит. Лог: %s", substr, output)
		}
	}
}

func TestParseFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name      string
		args      []string
		wantPath  string
		expectErr bool
	}{
		{"default", []string{"cmd"}, "config.env", false},
		{"custom", []string{"cmd", "-c", "custom.env"}, "custom.env", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = tt.args
			flags, err := ParseFlags()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantPath, flags.ConfigPath)
			}
		})
	}
}

func TestParseConfig_Defaults(t *testing.T) {
	os.Clearenv()
	cfg, err := ParseConfig("nonexistent.env")
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.AppHost)
	assert.Equal(t, "50051", cfg.AppPort)
	assert.Equal(t, "info", cfg.AppLogLevel)
	assert.Equal(t, "localhost", cfg.PostgresHost)
	assert.Equal(t, 5432, cfg.PostgresPort)
	assert.Equal(t, "exchange_rate_user", cfg.PostgresUser)
	assert.Equal(t, "exchange_rate_password", cfg.PostgresPassword)
	assert.Equal(t, "exchange_rate_db", cfg.PostgresDB)
	assert.Equal(t, 16, cfg.PostgresMaxOpenConns)
	assert.Equal(t, 8, cfg.PostgresMaxIdleConns)
}

func TestParseConfig_CustomEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_HOST", "0.0.0.0")
	os.Setenv("APP_PORT", "6000")
	os.Setenv("POSTGRES_PORT", "6543")
	os.Setenv("POSTGRES_MAX_OPEN_CONNS", "32")
	os.Setenv("POSTGRES_MAX_IDLE_CONNS", "16")

	cfg, err := ParseConfig("nonexistent.env")
	require.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.AppHost)
	assert.Equal(t, "6000", cfg.AppPort)
	assert.Equal(t, 6543, cfg.PostgresPort)
	assert.Equal(t, 32, cfg.PostgresMaxOpenConns)
	assert.Equal(t, 16, cfg.PostgresMaxIdleConns)
}

func TestParseConfig_InvalidNumbers(t *testing.T) {
	os.Clearenv()
	os.Setenv("POSTGRES_PORT", "notanumber")
	_, err := ParseConfig("nonexistent.env")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "неверное значение POSTGRES_PORT")
}

func TestParseConfig_InvalidOpenIdle(t *testing.T) {
	os.Clearenv()
	os.Setenv("POSTGRES_MAX_OPEN_CONNS", "abc")
	os.Setenv("POSTGRES_MAX_IDLE_CONNS", "def")

	_, err := ParseConfig("nonexistent.env")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "неверное значение POSTGRES_MAX_OPEN_CONNS")
}

func TestRun_PostgresContainer(t *testing.T) {
	ctx := context.Background()

	req := tc.ContainerRequest{
		Image: "postgres:15-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
	}

	pgContainer, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)
	mappedPort, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := &Config{
		AppHost:              "127.0.0.1",
		AppPort:              "50052",
		AppLogLevel:          "info",
		PostgresHost:         host,
		PostgresPort:         mappedPort.Int(),
		PostgresUser:         "testuser",
		PostgresPassword:     "testpass",
		PostgresDB:           "testdb",
		PostgresMaxOpenConns: 2,
		PostgresMaxIdleConns: 2,
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan error)
	go func() {
		done <- Run(runCtx, cfg)
	}()

	time.Sleep(2 * time.Second)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not exit within 10s after shutdown")
	}
}
