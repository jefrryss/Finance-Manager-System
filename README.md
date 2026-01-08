# Expenses Backend (Go + PostgreSQL)

Backend-сервис для управления **пользователями, счетами, транзакциями и категориями** через REST API.

Что реализовано по ТЗ Радыгина:
- регистрация/логин, выдача токена, проверка прав доступа, logout (аннулирование токена)  
- CRUD счетов (manual/imported), актуальный баланс, запрет удаления imported (возврат статуса/сообщения)  
- CRUD транзакций + фильтры; импортированные транзакции: менять только category/comment/is_hidden  
- CRUD категорий + перенос транзакций в служебную категорию **«Другое»** при удалении

## Быстрый старт (Docker)
```bash
docker compose up --build
```

Сервис поднимется на `http://localhost:8080`.

## ENV
Можно переопределить через переменные окружения (docker-compose уже задаёт):
- `APP_PORT` (default `8080`)
- `DB_HOST` (default `db`)
- `DB_PORT` (default `5432`)
- `DB_USER` (default `postgres`)
- `DB_PASSWORD` (default `postgres`)
- `DB_NAME` (default `expenses`)
- `DB_SSLMODE` (default `disable`)
- `JWT_SECRET` (default `dev-secret-change-me`)
- `JWT_TTL_HOURS` (default `168`)

## Эндпоинты
### Auth
- `POST /api/v1/auth/register` `{email,password}`
- `POST /api/v1/auth/login` `{email,password}`
- `POST /api/v1/auth/logout` (Bearer token)

### Accounts
- `GET /api/v1/accounts`
- `POST /api/v1/accounts` `{name,type,initial_balance,external_id,last_synced_at}`
- `GET /api/v1/accounts/:id`
- `PATCH /api/v1/accounts/:id` `{name}`
- `DELETE /api/v1/accounts/:id`

### Categories
- `GET /api/v1/categories`
- `POST /api/v1/categories` `{name,type}`
- `PATCH /api/v1/categories/:id` `{name}`
- `DELETE /api/v1/categories/:id`

### Transactions
- `GET /api/v1/transactions?account_id=&category_id=&type=&start=&end=&is_hidden=`
  - `start/end` — RFC3339 (например `2026-01-01T00:00:00Z`)
- `POST /api/v1/transactions` `{account_id,amount,type,occurred_at,category_id,comment,is_imported,is_hidden}`
- `GET /api/v1/transactions/:id`
- `PATCH /api/v1/transactions/:id`  
  - если `is_imported=true`: можно менять только `{category_id,comment,is_hidden}`
- `DELETE /api/v1/transactions/:id` (только если `is_imported=false`)

## Формат сумм
`amount`, `initial_balance`, `balance` — **int64** (условно «копейки/центы»).  
(Если нужно, можно заменить на numeric/decimal — но для курсового обычно ок.)

## Локальный запуск без Docker
```bash
export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=expenses
export JWT_SECRET=dev-secret-change-me
go mod tidy
go run ./cmd/server
```
