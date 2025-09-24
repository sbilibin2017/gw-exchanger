# GW Exchanger

Сервис для работы с курсами валют через gRPC с подключением к PostgreSQL.

---

## Структура проекта

```
.
├── cmd
│   └── main.go                  # Точка входа приложения, запускает gRPC сервер и инициализирует конфигурацию
├── config.env                    # Файл конфигурации для локального запуска, содержит переменные окружения
├── go.mod                        # Файл модулей Go, описывает зависимости проекта
├── go.sum                        # Контрольные суммы зависимостей, автоматически создаются Go
├── internal                      # Основная логика приложения, недоступна извне пакета
│   ├── app
│   │   ├── app.go                # Основная логика инициализации сервиса, gRPC сервера, DI, запуск
│   │   └── app_test.go           # Тесты для пакета app, например для PrintBuildInfo и конфигурации
│   ├── handlers
│   │   ├── exchange_rate.go      # gRPC обработчики для работы с курсами валют
│   │   ├── exchange_rate_mock.go # Мок реализации ExchangeRateService для тестов
│   │   └── exchange_rate_test.go # Тесты для обработчиков (handlers)
│   ├── logger
│   │   ├── logger.go             # Конфигурация логгера, уровень логирования и формат
│   │   └── logger_test.go        # Тесты для логгера
│   ├── models
│   │   └── exchange_rate.go      # Структуры данных/модели для курсов валют
│   ├── repositories
│   │   ├── exchange_rate.go      # Репозитории для чтения/записи данных о курсах валют в БД
│   │   └── exchange_rate_test.go # Тесты для репозиториев
│   └── services
│       ├── exchange_rate.go      # Бизнес-логика работы с курсами валют
│       ├── exchange_rate_mock.go # Мок сервиса ExchangeRateService для тестов
│       └── exchange_rate_test.go # Тесты для сервисного слоя
├── Makefile                      # Скрипт для сборки, запуска и миграций проекта
├── migrations
│   └── 0001_create_exchange_rates_table.sql  # SQL миграция для создания таблицы курсов валют
└── README.md                     # Документация проекта, как запускать, описание API, etc.
```

---

## Требования

- Go >= 1.21  
- PostgreSQL >= 14  
- Git (для сборки с commit hash)  

---

## Настройка конфигурации

Пример `config.env`:

```env
# Настройки сервиса
APP_HOST=localhost
APP_PORT=50051
APP_LOG_LEVEL=info

# Настройки PostgreSQL
POSTGRES_HOST=192.168.2.22
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgresP1
POSTGRES_DB=text_db
POSTGRES_MAX_OPEN_CONNS=16
POSTGRES_MAX_IDLE_CONNS=8
```

## Сборка
```
GOOS=linux GOARCH=amd64 go build -ldflags "\
  -X 'github.com/sbilibin2017/gw-exchanger/internal/app.buildVersion=1.0.0' \
  -X 'github.com/sbilibin2017/gw-exchanger/internal/app.buildCommit=$(git rev-parse --short HEAD)' \
  -X 'github.com/sbilibin2017/gw-exchanger/internal/app.buildDate=$(date +%Y-%m-%d_%H:%M:%S)'" \
  -o main ./cmd
```
## Запуск
```
./main -c config.env
```

# Покрытие

| Package                                        | Coverage   |
|-----------------------------------------------|------------|
| github.com/sbilibin2017/gw-exchanger/cmd       | 0.0%       |
| github.com/sbilibin2017/gw-exchanger/internal/app | 84.5%   |
| github.com/sbilibin2017/gw-exchanger/internal/handlers | 100.0% |
| github.com/sbilibin2017/gw-exchanger/internal/logger | 88.9% |
| github.com/sbilibin2017/gw-exchanger/internal/models | no test files |
| github.com/sbilibin2017/gw-exchanger/internal/repositories | 100.0% |
| github.com/sbilibin2017/gw-exchanger/internal/services | 100.0% |


