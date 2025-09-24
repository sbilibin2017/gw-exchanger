package app

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
	"github.com/sbilibin2017/gw-exchanger/internal/handlers"
	"github.com/sbilibin2017/gw-exchanger/internal/logger"
	"github.com/sbilibin2017/gw-exchanger/internal/repositories"
	"github.com/sbilibin2017/gw-exchanger/internal/services"
	"google.golang.org/grpc"

	_ "github.com/jackc/pgx/v5/stdlib"
	pb "github.com/sbilibin2017/proto-exchange/exchange"
)

// Переменные сборки, их значения задаются через ldflags при сборке
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// PrintBuildInfo выводит информацию о сборке в лог
func PrintBuildInfo() {
	log.Printf("Запуск сервиса версии %s, коммит %s, сборка %s", buildVersion, buildCommit, buildDate)
}

// AppFlags содержит флаги командной строки приложения.
type AppFlags struct {
	ConfigPath string
}

// ParseFlags парсит флаги командной строки и возвращает структуру AppFlags.
func ParseFlags() (*AppFlags, error) {
	c := flag.String("c", "config.env", "Путь к файлу конфигурации")
	flag.Parse()
	return &AppFlags{ConfigPath: *c}, nil
}

// Config содержит все настройки приложения и базы данных PostgreSQL.
type Config struct {
	AppHost     string
	AppPort     string
	AppLogLevel string

	PostgresHost         string
	PostgresPort         int
	PostgresUser         string
	PostgresPassword     string
	PostgresDB           string
	PostgresMaxOpenConns int
	PostgresMaxIdleConns int
}

// ParseConfig загружает конфигурацию из файла .env и переменных окружения.
func ParseConfig(pathToEnvFile string) (*Config, error) {
	cfg := &Config{}

	_ = godotenv.Load(pathToEnvFile) // Игнорируем ошибку, если файла нет

	getEnv := func(key, defaultValue string) string {
		if value, exists := os.LookupEnv(key); exists && value != "" {
			return value
		}
		return defaultValue
	}

	// Настройки приложения
	cfg.AppHost = getEnv("APP_HOST", "localhost")
	cfg.AppPort = getEnv("APP_PORT", "50051")
	cfg.AppLogLevel = getEnv("APP_LOG_LEVEL", "info")

	// Настройки PostgreSQL
	cfg.PostgresHost = getEnv("POSTGRES_HOST", "localhost")
	cfg.PostgresUser = getEnv("POSTGRES_USER", "exchange_rate_user")
	cfg.PostgresPassword = getEnv("POSTGRES_PASSWORD", "exchange_rate_password")
	cfg.PostgresDB = getEnv("POSTGRES_DB", "exchange_rate_db")

	// Конвертация портов и пулов соединений
	var err error
	if cfg.PostgresPort, err = strconv.Atoi(getEnv("POSTGRES_PORT", "5432")); err != nil {
		return nil, fmt.Errorf("неверное значение POSTGRES_PORT: %w", err)
	}
	if cfg.PostgresMaxOpenConns, err = strconv.Atoi(getEnv("POSTGRES_MAX_OPEN_CONNS", "16")); err != nil {
		return nil, fmt.Errorf("неверное значение POSTGRES_MAX_OPEN_CONNS: %w", err)
	}
	if cfg.PostgresMaxIdleConns, err = strconv.Atoi(getEnv("POSTGRES_MAX_IDLE_CONNS", "8")); err != nil {
		return nil, fmt.Errorf("неверное значение POSTGRES_MAX_IDLE_CONNS: %w", err)
	}

	return cfg, nil
}

// Run запускает gRPC сервер с подключением к PostgreSQL и продвинутым логированием.
func Run(ctx context.Context, cfg *Config) error {
	// 1. Инициализация логгера
	log, err := logger.New(cfg.AppLogLevel)
	if err != nil {
		fmt.Printf("не удалось инициализировать логгер: %v\n", err)
		return err
	}
	log.Infof("Логгер инициализирован, уровень: %s", cfg.AppLogLevel)

	// 2. Подключение к базе
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB)
	log.Infof("Подключение к PostgreSQL: %s:%d", cfg.PostgresHost, cfg.PostgresPort)

	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		log.Errorf("Ошибка подключения к базе: %v", err)
		return fmt.Errorf("не удалось подключиться к PostgreSQL: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Warnf("Ошибка при закрытии подключения к базе: %v", err)
		} else {
			log.Info("Подключение к базе закрыто")
		}
	}()
	log.Info("Успешное подключение к PostgreSQL")

	db.SetMaxOpenConns(cfg.PostgresMaxOpenConns)
	db.SetMaxIdleConns(cfg.PostgresMaxIdleConns)
	log.Infof("MaxOpenConns=%d, MaxIdleConns=%d", cfg.PostgresMaxOpenConns, cfg.PostgresMaxIdleConns)

	// 3. Инициализация слоёв приложения
	log.Info("Инициализация репозиториев и сервисов...")
	exchangeRateReadRepo := repositories.NewExchangeRateReadRepository(log, db)
	exchangeRateService := services.NewExchangeRateService(log, exchangeRateReadRepo)
	exchangeRateHandler := handlers.NewExchangeRateHandler(log, exchangeRateService)
	log.Info("Сервисы и обработчики инициализированы")

	// 4. gRPC сервер
	grpcServer := grpc.NewServer()
	pb.RegisterExchangeServiceServer(grpcServer, exchangeRateHandler)
	listenAddr := fmt.Sprintf("%s:%s", cfg.AppHost, cfg.AppPort)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Errorf("Ошибка при создании listener: %v", err)
		return fmt.Errorf("не удалось слушать %s: %w", listenAddr, err)
	}
	log.Infof("gRPC сервер слушает на %s", listenAddr)

	// 5. Запуск сервера и graceful shutdown
	errChan := make(chan error, 1)
	go func() {
		log.Info("Запуск gRPC сервера...")
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			errChan <- fmt.Errorf("ошибка gRPC сервера: %w", serveErr)
		}
	}()

	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	select {
	case <-shutdownCtx.Done():
		log.Info("Получен сигнал завершения, остановка gRPC сервера...")
		grpcServer.GracefulStop()
		log.Info("gRPC сервер остановлен корректно")
	case serveErr := <-errChan:
		log.Errorf("gRPC сервер завершился с ошибкой: %v", serveErr)
		return serveErr
	}

	return nil
}
