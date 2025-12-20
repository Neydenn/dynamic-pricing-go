 # Dynamic Pricing (educational)

 Мини‑система из трёх сервисов (catalog, order, pricing), Kafka и трех БД Postgres. Каталог управляет товарами, заказ — пользователями и заказами, pricing — рассчитывает цену и хранит текущие цены. Swagger UI доступен на http://localhost:8089.

 ## Как устроено
 - Входные точки: `cmd/app/{catalog,order,pricing}/main.go` — загрузка `config.yaml`, подключение к Postgres, настройка Kafka, запуск HTTP‑сервера.
 - Бизнес‑модули: `internal/{catalog,order,pricing}`
   - `domain.go` — доменные типы
   - `repository.go` — доступ к БД (pgxpool)
   - `service.go` / `engine.go` — бизнес‑логика
   - `handler.go` — HTTP‑эндпоинты (chi)
 - Инфраструктура: `internal/httpserver` (HTTP сервер, CORS), `internal/kafka` (producer/consumer), `internal/postgres` (создание пула).
 - Конфиг: `config/config.go` (структуры/loader), `config.yaml` (локальные значения; можно переопределить `CONFIG_PATH`).
 - API: `api/openapi.yaml` — OpenAPI 3.0 (каждый путь привязан к своему сервису через `servers`).
 - Docker: `Dockerfile.*`, `docker-compose.yaml`; вспомогательные SQL — `scripts/postgres/*.sql`, демо — `scripts/demo.sh`.

 ## Запуск и разработка
 - Поднять всё в Docker: `make up` (Kafka, Postgres, 3 сервиса, Swagger UI на :8089)
 - Остановить и очистить данные: `make down`
 - Логи: `make logs`

 Примеры запросов (локально):
 - Создать товар: `POST http://localhost:8081/products` тело `{ "name":"A", "base_price":10, "stock":5 }`
 - Создать пользователя: `POST http://localhost:8082/users` тело `{ "email":"a@ex.com" }`
 - Получить цену: `GET http://localhost:8083/prices/{product_id}`

