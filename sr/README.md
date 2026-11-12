# Orders API

Backend-сервис заказов на Go. Сервис позволяет создавать заказы, получать один заказ и список заказов, изменять статус и просматривать историю изменений.

Данные хранятся в памяти приложения и удаляются при перезапуске. Внешняя база данных не требуется.

## Возможности

- REST API с ответами в JSON;
- валидация позиций заказа и автоматический расчёт итоговой стоимости;
- проверка допустимых переходов статусов;
- история изменений статусов;
- единый формат ошибок;
- структурированные JSON-логи через `log/slog`;
- middleware логирования запросов и восстановления после паники;
- Swagger UI и OpenAPI-спецификация;
- модульные table-driven тесты слоя бизнес-логики.

## Технологии

- Go 1.22 или новее;
- `chi` — HTTP-маршрутизация;
- `swag` и `http-swagger` — OpenAPI и Swagger UI;
- `testify` — проверки в модульных тестах.

## Запуск

Из корня проекта:

```bash
go mod tidy
go run ./cmd/api
```

По умолчанию сервер слушает `http://localhost:8080`. Адрес можно изменить переменной окружения `HTTP_ADDR`:

```bash
HTTP_ADDR=:9090 go run ./cmd/api
```

Также доступны команды Makefile:

```bash
make run
make build
```

Проверка доступности сервиса:

```bash
curl http://localhost:8080/health
```

Ответ:

```json
{"status":"ok"}
```

## API

| Метод | Путь | Назначение | Успешный код |
|---|---|---|---:|
| GET | `/health` | Проверка доступности сервиса | 200 |
| POST | `/api/v1/orders` | Создание заказа | 201 |
| GET | `/api/v1/orders` | Получение списка заказов | 200 |
| GET | `/api/v1/orders/{id}` | Получение заказа по ID | 200 |
| PATCH | `/api/v1/orders/{id}/status` | Изменение статуса | 200 |
| GET | `/api/v1/orders/{id}/history` | Получение истории статусов | 200 |

Для запросов с телом необходим заголовок `Content-Type: application/json`. Максимальный размер тела запроса — 1024 байта. Неизвестные JSON-поля отклоняются.

### Создание заказа

```bash
curl -i -X POST http://localhost:8080/api/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{
    "items": [
      {"name": "Keyboard", "quantity": 1, "price": 500000},
      {"name": "Mouse", "quantity": 2, "price": 100000}
    ]
  }'
```

`price` указывается целым числом в минимальных денежных единицах, например в копейках.

При создании проверяется:

- наличие хотя бы одной позиции;
- непустое название каждого товара;
- `quantity > 0`;
- `price > 0`.

Итоговая стоимость рассчитывается как сумма `price × quantity` по всем позициям. Новый заказ получает статус `new`, дату создания и первую запись в истории.

Пример ответа `201 Created`:

```json
{
  "id": 1,
  "items": [
    {"name": "Keyboard", "quantity": 1, "price": 500000},
    {"name": "Mouse", "quantity": 2, "price": 100000}
  ],
  "total_price": 700000,
  "status": "new",
  "created_at": "2026-01-01T12:00:00Z"
}
```

### Получение заказов

```bash
curl http://localhost:8080/api/v1/orders
curl http://localhost:8080/api/v1/orders/1
```

Список возвращается в порядке создания заказов. Если заказов нет, возвращается пустой массив `[]`.

### Изменение статуса

```bash
curl -i -X PATCH http://localhost:8080/api/v1/orders/1/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"processing"}'
```

Допустимые статусы: `new`, `processing`, `completed`, `cancelled`.

Разрешены только переходы:

- `new → processing`;
- `new → cancelled`;
- `processing → completed`;
- `processing → cancelled`.

`completed` и `cancelled` являются конечными статусами. Недопустимый переход возвращает `409 Conflict`. Каждое успешное изменение добавляет в историю новый статус и время изменения.

### История статусов

```bash
curl http://localhost:8080/api/v1/orders/1/history
```

Пример ответа:

```json
[
  {"status":"new","changed_at":"2026-01-01T12:00:00Z"},
  {"status":"processing","changed_at":"2026-01-01T12:05:00Z"}
]
```

Записи возвращаются в хронологическом порядке.

## Ошибки

Ошибки API имеют единый формат:

```json
{
  "error": "status transition is not allowed",
  "code": "INVALID_STATUS_TRANSITION"
}
```

| Ситуация | HTTP-код |
|---|---:|
| Некорректный JSON или невалидные данные | 400 |
| Заказ не найден | 404 |
| Недопустимый переход статуса | 409 |
| Тело запроса превышает 1024 байта | 413 |
| Неподдерживаемый `Content-Type` | 415 |
| Неожиданная внутренняя ошибка | 500 |

Внутренние детали ошибок не включаются в ответ клиенту.

## Swagger

После запуска сервиса документация доступна по адресам:

- Swagger UI: <http://localhost:8080/swagger/index.html>;
- OpenAPI JSON: <http://localhost:8080/swagger/doc.json>.

Сгенерированная документация находится в каталоге `docs/`. Для её обновления после изменения аннотаций выполните:

```bash
swag init -g cmd/api/main.go -o docs
```

Если утилита не установлена:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## Тестирование и проверка

```bash
go test ./...
go test ./internal/service/... -cover
go vet ./...
go build ./...
gofmt -l .
```

Тесты слоя `service` используют table-driven подход, `testify/require` и подмену репозитория через интерфейс. Покрываются создание и валидация заказа, расчёт стоимости, получение заказов, допустимые и недопустимые переходы, конечные статусы, история и ошибки репозитория. Текущее покрытие пакета `service` — 100%.

Для комплексной проверки можно выполнить:

```bash
make check
```

## Архитектура

```text
HTTP request → handler → service → repository → in-memory storage
```

- `handler` обрабатывает HTTP-запросы, JSON и HTTP-коды ошибок;
- `service` содержит валидацию и бизнес-правила;
- `repository` хранит заказы и возвращает их через интерфейс;
- зависимости передаются через конструкторы.

```text
cmd/api/                    Точка входа
internal/handler/           Маршруты, обработчики и middleware
internal/order/             Модели, ошибки и интерфейс репозитория
internal/service/           Бизнес-логика и модульные тесты
internal/repository/memory/ In-memory-реализация репозитория
internal/utils/             JSON и DTO ответов
docs/                       Swagger/OpenAPI-документация
```

## Логирование

Для каждого запроса записываются `method`, `path`, `status`, `duration` и `request_id`. Идентификатор можно передать в заголовке `X-Request-ID`; если он отсутствует, сервер генерирует его и возвращает в одноимённом заголовке ответа.

Recovery middleware перехватывает панику, записывает её в лог и возвращает безопасный JSON-ответ с кодом `500`, не останавливая сервер.
