package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/sbilibin2017/gw-exchanger/internal/logger"
	"github.com/sbilibin2017/gw-exchanger/internal/middlewares"
	"github.com/sbilibin2017/gw-exchanger/internal/repositories"
	"github.com/sbilibin2017/gw-exchanger/internal/services"
	pb "github.com/sbilibin2017/proto-exchange/exchange"
	"google.golang.org/grpc"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Build info
var (
	buildVersion = "N/A" // Version of the service
	buildDate    = "N/A" // Build date of the service
	buildCommit  = "N/A" // Commit hash of the build
)

// main is the entry point of the application.
// It prints build info, parses configuration, and starts the gRPC server.
func main() {
	printBuildInfo()
	configPath := parseFlags()

	appHost, appPort,
		pgHost, pgPort, pgUser, pgPassword, pgDB,
		pgMaxOpenConns, pgMaxIdleConns,
		logLevel, err := parseConfig(configPath)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	if err := run(context.Background(),
		appHost, appPort,
		pgHost, pgPort, pgUser, pgPassword, pgDB,
		pgMaxOpenConns, pgMaxIdleConns,
		logLevel,
	); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}

// printBuildInfo prints build info to stdout, each field on a new line.
func printBuildInfo() {
	log.Printf("Version: %s", buildVersion)
	log.Printf("Commit: %s", buildCommit)
	log.Printf("Build date: %s", buildDate)
}

// parseFlags parses the -c flag for configuration file path.
func parseFlags() string {
	c := flag.String("c", "config.env", "Path to configuration file")
	flag.Parse()
	return *c
}

// parseConfig loads environment variables and returns configuration values.
func parseConfig(path string) (
	appHost, appPort string,
	pgHost string, pgPort int, pgUser, pgPassword, pgDB string,
	pgMaxOpenConns, pgMaxIdleConns int,
	logLevel string,
	err error,
) {
	godotenv.Load(path)

	getEnv := func(key, def string) string {
		if val, ok := os.LookupEnv(key); ok && val != "" {
			return val
		}
		return def
	}

	appHost = getEnv("APP_HOST", "localhost")
	appPort = getEnv("APP_PORT", "50051")
	logLevel = getEnv("APP_LOG_LEVEL", "info")

	pgHost = getEnv("POSTGRES_HOST", "localhost")
	pgUser = getEnv("POSTGRES_USER", "exchange_rate_user")
	pgPassword = getEnv("POSTGRES_PASSWORD", "exchange_rate_password")
	pgDB = getEnv("POSTGRES_DB", "exchange_rate_db")
	if pgPort, err = strconv.Atoi(getEnv("POSTGRES_PORT", "5432")); err != nil {
		return
	}
	if pgMaxOpenConns, err = strconv.Atoi(getEnv("POSTGRES_MAX_OPEN_CONNS", "16")); err != nil {
		return
	}
	if pgMaxIdleConns, err = strconv.Atoi(getEnv("POSTGRES_MAX_IDLE_CONNS", "8")); err != nil {
		return
	}

	return
}

// run initializes logger, database, service, and starts the gRPC server with graceful shutdown.
func run(ctx context.Context,
	appHost, appPort string,
	pgHost string, pgPort int, pgUser, pgPassword, pgDB string,
	pgMaxOpenConns, pgMaxIdleConns int,
	logLevel string,
) error {
	log, err := logger.New(logLevel)
	if err != nil {
		fmt.Printf("failed to init logger: %v\n", err)
		return err
	}
	defer log.Sync()
	log.Infof("Logger initialized, level: %s", logLevel)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		pgUser, pgPassword, pgHost, pgPort, pgDB)
	log.Infof("Connecting to PostgreSQL: %s:%d", pgHost, pgPort)
	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		log.Errorf("DB connection error: %v", err)
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(pgMaxOpenConns)
	db.SetMaxIdleConns(pgMaxIdleConns)
	log.Infof("PostgreSQL connected, MaxOpenConns=%d, MaxIdleConns=%d", pgMaxOpenConns, pgMaxIdleConns)

	readRepo := repositories.NewExchangeRateReadRepository(log, db)

	exchangeService := services.NewExchangeRateService(log, readRepo)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middlewares.LoggingMiddleware(log)),
	)
	pb.RegisterExchangeServiceServer(grpcServer, exchangeService)

	listenAddr := fmt.Sprintf("%s:%s", appHost, appPort)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Errorf("Listener error: %v", err)
		return err
	}
	log.Infof("gRPC server listening on %s", listenAddr)

	errChan := make(chan error, 1)
	go func() {
		log.Info("Starting gRPC server...")
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", serveErr)
		}
	}()

	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	select {
	case <-shutdownCtx.Done():
		log.Info("Shutdown signal received, stopping gRPC server...")
		grpcServer.GracefulStop()
		log.Info("gRPC server stopped gracefully")
	case serveErr := <-errChan:
		log.Errorf("gRPC server exited with error: %v", serveErr)
		return serveErr
	}

	return nil
}
