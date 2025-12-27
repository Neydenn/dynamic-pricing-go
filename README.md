 # Dynamic Pricing (educational)

 Мини‑система из трёх сервисов (catalog, order, pricing), Kafka и трех БД Postgres. Каталог управляет товарами, заказ — пользователями и заказами, pricing — рассчитывает цену и хранит текущие цены. Swagger UI доступен на http://localhost:8089.

 ## Как устроено
 - Входные точки: `cmd/app/{catalog,order,pricing}/main.go` — минимальные; всё связывание в `internal/bootstrap`.
 - Слои:
   - Домены: `internal/models` (Product, User, Order, Price).
   - Хранилище: `internal/storage/pg` (репозитории для Product/Order/Price).
   - Сервисы: `internal/services/{catalog,order,pricing}` (интерфейсы и бизнес‑логика; `pricing/engine.go`).
   - HTTP API: `internal/api/{catalog_api,order_api,pricing_api}` (chi‑handlers).
 - Инфраструктура: `internal/httpserver` (HTTP сервер, CORS), `internal/producer` и `internal/consumer` (Kafka), `internal/storage/pg` (пул + репозитории).
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

 - docker-compose exec redis redis-cli GET app:version
 - docker-compose exec redis redis-cli HGETALL product:42
 - docker-compose exec redis redis-cli LRANGE queue:pricing 0 -1
 - docker compose exec users-db psql -U users -d users