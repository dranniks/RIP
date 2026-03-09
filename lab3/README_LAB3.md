# Лабораторная 3: REST веб-сервис для SPA (XRF)

## Что реализовано

- Полный REST API под префиксом `/api`.
- ORM через GORM (PostgreSQL).
- MinIO для загрузки файлов услуги (изображение и короткое видео).
- Логика m-m без использования PK m-m в URL:
  - добавление услуги в черновик,
  - изменение `quantity/sort_order/match_value`,
  - удаление по `service_id`.
- Бизнес-ограничения статусов:
  - создатель может только удалить/сформировать черновик,
  - модератор может только завершить/отклонить сформированную заявку.
- Запрет изменения системных полей с клиента (id, статус, даты, creator/moderator).
- Фильтрация:
  - услуги: `q`,
  - заявки: `status`, `formed_from`, `formed_to`.
- Удаленные записи не возвращаются в выдаче.

## Singleton фиксированного пользователя

- Файл: `internal/app/identity/singleton.go`.
- Функция: `CurrentUsers()`.
- Константные пользователи:
  - creator: id=1, login=`xrf_creator`
  - moderator: id=2, login=`xrf_moderator`
- Использование в методах API:
  - creator: операции черновика и формирования,
  - moderator: операции завершения/отклонения.

## Модели и сериализаторы

- Модели (ORM): `internal/app/model/models.go`
  - `User`
  - `ReferenceAlloyService`
  - `ArtifactClaim`
  - `ClaimAlloyMatch`
- Сериализаторы (JSON-структуры): `internal/app/handler/serializers.go`
  - `serviceSerializer`
  - `cartSerializer`
  - `claimItemSerializer`
  - `claimSerializer`
  - `claimListSerializer`
  - `userSerializer`

## API методы

### Service domain

- `GET /api/services` (фильтр `q`)
- `GET /api/services/:id`
- `POST /api/services` (multipart: поля + `image` + `video`)

### M-M domain

- `POST /api/claim-items`
- `PUT /api/claim-items/:service_id`
- `DELETE /api/claim-items/:service_id`

### Claim domain

- `GET /api/claims/cart-icon`
- `GET /api/claims` (фильтры `status`, `formed_from`, `formed_to`)
- `GET /api/claims/:id`
- `PUT /api/claims/:id`
- `PUT /api/claims/:id/form`
- `PUT /api/claims/:id/moderate`
- `DELETE /api/claims/:id`

### User domain

- `POST /api/users/register`
- `POST /api/users/auth` (заглушка)
- `POST /api/users/logout` (заглушка)

## Формирование заявки (бизнес-логика)

При `PUT /api/claims/:id/form`:

- Проверяются обязательные поля заявки.
- Рассчитываются поля m-m (формула состава + score).
- Считается стоимость `total_cost`.
- Считается дата доставки `planned_delivery_at` (не более 30 дней).
- Проставляется `formed_at`, `status=сформирован`.

При `PUT /api/claims/:id/moderate`:

- Разрешено только для `status=сформирован`.
- Действие `complete/reject`.
- Проставляются `moderator_id`, `completed_at`, итоговый статус.

## Коллекция Postman (16 запросов)

- Файл: `docs/LAB3_API_COLLECTION.postman_collection.json`
- Импортируйте коллекцию и задайте переменные:
  - `base_url = http://localhost:8080/api`
  - `draft_claim_id`
  - `service_id_1`
  - `service_id_2`

## SQL SELECT для показа изменений

```sql
-- Услуги
SELECT id, slug, name, status, image_file_name, video_file_name, unit_price
FROM reference_alloy_services
ORDER BY id DESC;

-- Заявки
SELECT id, claim_code, status, creator_id, moderator_id, created_at, formed_at, completed_at,
       total_cost, planned_delivery_at, completion_formula_result
FROM artifact_claims
ORDER BY id DESC;

-- M-M таблица
SELECT id, claim_id, service_id, quantity, sort_order, match_value, composition_result, match_score
FROM claim_alloy_matches
ORDER BY id DESC;

-- Пользователи
SELECT id, login, full_name, role, created_at
FROM users
ORDER BY id DESC;
```

## Диаграмма классов

- Файл: `docs/LAB3_CLASS_DIAGRAM.md`.

## Контрольные вопросы (кратко)

- Веб-сервис: сервер, предоставляющий функции через сеть (HTTP API).
- REST: архитектурный стиль ресурсов (`/resource`, HTTP-методы, статeless).
- RPC: вызов удаленной процедуры, акцент на действия (`callMethod`).
- Заголовки и методы HTTP: headers несут метаданные, методы описывают действие (`GET/POST/PUT/DELETE`).
- Версии HTTP: `HTTP/1.1`, `HTTP/2`, `HTTP/3` (QUIC).
- HTTPS: HTTP поверх TLS, шифрование + целостность + аутентификация.
- OSI ISO: 7 уровней (физический -> прикладной).

## Запуск

```powershell
# из корня репозитория
cd d:\VUZ\sem6\RIP

docker compose up -d
cd lab3
go run ./cmd/xrf-app
```
