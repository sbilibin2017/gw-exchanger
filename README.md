# GW Exchanger

### Цель
Сервис предоставляет актуальные курсы валют через gRPC.  
Источник данных — база **PostgreSQL**.

### Задачи
- Хранение и предоставление курсов валют (**USD, RUB, EUR**).  
- Предоставление API для запроса курса одной валютной пары или всех курсов.  
- Легкая замена хранилища (например, на Redis) через интерфейс `ExchangeRateReader`.  
- Логирование всех запросов и ответов с уникальным `request_id`.  

---

### API (gRPC)

| Метод | Входное сообщение | Выходное сообщение | Описание |
|-------|-----------------|------------------|----------|
| `GetExchangeRates` | `Empty` | `ExchangeRatesResponse` | Получение всех курсов валют. Возвращает карту `to_currency -> rate`. |
| `GetExchangeRateForCurrency` | `CurrencyRequest` | `ExchangeRateResponse` | Получение курса между двумя валютами. Поддерживаются `USD`, `RUB`, `EUR`. |

---

### Сценарии работы

1. Клиент отправляет gRPC-запрос на получение курса одной валютной пары или всех курсов.  
2. Сервис читает данные из PostgreSQL через репозиторий `ExchangeRateReadRepository`.  
3. Сервис возвращает ответ с курсами валют.  
4. Все запросы и ответы логируются с уникальным `request_id`.  

---

## Структура проекта

```
.
├── cmd
│ └── main.go
├── config.env
├── Dockerfile
├── go.mod
├── go.sum
├── internal
│ ├── logger
│ │ ├── logger.go
│ │ └── logger_test.go
│ ├── middlewares
│ │ ├── logging.go
│ │ └── logging_test.go
│ ├── models
│ │ └── exchange_rate.go
│ ├── repositories
│ │ ├── exchange_rate.go
│ │ └── exchange_rate_test.go
│ └── services
│ ├── exchange_rate.go
│ ├── exchange_rate_mock.go
│ └── exchange_rate_test.go
├── Makefile
├── migrations
│ └── 0001_create_exchange_rates_table.sql
└── README.md
```

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

---

## Сборка проекта

Для сборки сервиса используйте команду:

```bash
GOOS=linux GOARCH=amd64 go build -ldflags "\
  -X 'github.com/sbilibin2017/gw-exchanger/cmd.buildVersion=1.0.0' \
  -X 'github.com/sbilibin2017/gw-exchanger/cmd.buildCommit=$(git rev-parse --short HEAD)' \
  -X 'github.com/sbilibin2017/gw-exchanger/cmd.buildDate=$(date +%Y-%m-%d_%H:%M:%S)'" \
  -o main ./cmd
```

Пояснения:
* GOOS=linux и GOARCH=amd64 — целевая ОС и архитектура.
* -ldflags — задаёт значения переменных buildVersion, buildCommit и buildDate, которые используются в main.go.
* -o main — имя выходного бинарного файла.
* ./cmd — путь к пакету с main.go.

## Proto файл

[Сервис использует proto-файл для описания API](https://github.com/sbilibin2017/proto-exchange/blob/main/exchange/exchange.proto)  

## Запуск

```shell
./main -c config.env
```

