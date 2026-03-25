# Лабораторная 4: JWT + Redis сессии + Swagger

## Что реализовано

- Аутентификация через JWT (`Authorization: Bearer <token>`).
- Серверные сессии в Redis:
  - при `POST /api/users/auth` создается ключ сессии;
  - в key хранится пользователь (`user_id`, `login`, `role`) и время истечения;
  - middleware проверяет не только JWT подпись, но и наличие/валидность сессии в Redis;
  - `POST /api/users/logout` удаляет Redis-сессию.
- Ролевые permissions:
  - `guest` -> `GET /api/claims` = `401`;
  - `creator` -> видит только свои заявки;
  - `creator` -> `PUT /api/claims/:id/moderate` = `403`;
  - `moderator` -> видит все заявки и успешно завершает/отклоняет.
- Swagger:
  - UI: `/swagger`
  - OpenAPI JSON: `/swagger/openapi.json`
  - Включена схема `BearerAuth`.

## Матрица доступа

- Без JWT:
  - `GET /api/services`
  - `GET /api/services/:id`
  - `POST /api/users/register`
  - `POST /api/users/auth`
- С JWT:
  - `POST /api/services`
  - `POST/PUT/DELETE /api/claim-items...`
  - `GET /api/claims/cart-icon`
  - `GET /api/claims`
  - `GET /api/claims/:id`
  - `PUT /api/claims/:id`
  - `PUT /api/claims/:id/form`
  - `DELETE /api/claims/:id`
  - `POST /api/users/logout`
- Только moderator:
  - `PUT /api/claims/:id/moderate`

## Пример auth

`POST /api/users/auth`

```json
{
  "login": "xrf_creator",
  "password": "creator"
}
```

Ответ содержит:
- `token`
- `expires_at`
- `session_id`
- `session_key`
- `session_ttl`
- `session_expires_at`

## Проверка Redis сессий

```powershell
docker exec -e REDISCLI_AUTH=password rip-redis-1 redis-cli PING
docker exec -e REDISCLI_AUTH=password rip-redis-1 redis-cli KEYS 'lab4:sessions:*'
docker exec -e REDISCLI_AUTH=password rip-redis-1 redis-cli GET <session_key>
```

## Переменные окружения

См. `.env.example`:
- `JWT_SECRET`
- `JWT_TTL_MINUTES`
- `REDIS_HOST`
- `REDIS_PORT`
- `REDIS_PASSWORD`
- `REDIS_DB`
- `REDIS_TIMEOUT_SECONDS`
- `SESSION_TTL_MINUTES`
- `SESSION_KEY_PREFIX`
- `APP_PORT`

