# Лабораторная 4: JWT авторизация, permissions, Swagger

## Что реализовано

- JWT-аутентификация (`Authorization: Bearer <token>`), без cookie/sessions.
- Ролевая модель:
  - `creator`:
    - работает только со своими заявками;
    - не может вызывать модераторский метод завершения/отклонения.
  - `moderator`:
    - видит все заявки;
    - может завершать/отклонять сформированные заявки.
- Автозаполнение пользователя в заявке:
  - создатель заявки берется из JWT (не из тела запроса).
- Swagger UI:
  - `/swagger`
  - OpenAPI JSON: `/swagger/openapi.json`
  - В Swagger есть `BearerAuth` и описание методов.

## Права доступа по API

- Публично (без JWT):
  - `GET /api/services`
  - `GET /api/services/:id`
  - `POST /api/users/register`
  - `POST /api/users/auth`
- Требуется JWT:
  - `POST /api/services`
  - все методы `claim-items`
  - `GET /api/claims/cart-icon`
  - `GET /api/claims`
  - `GET /api/claims/:id`
  - `PUT /api/claims/:id`
  - `PUT /api/claims/:id/form`
  - `DELETE /api/claims/:id`
  - `POST /api/users/logout`
- Только `moderator`:
  - `PUT /api/claims/:id/moderate`

## Проверки по заданию

- Гость при `GET /api/claims` получает `401`.
- Создатель при `GET /api/claims` видит только свои заявки.
- Создатель при `PUT /api/claims/:id/moderate` получает `403`.
- Модератор при `PUT /api/claims/:id/moderate` получает успех, проставляются поля модератора/дата завершения.
- Модератор при `GET /api/claims` видит все заявки.

## Пример аутентификации (JWT)

`POST /api/users/auth`

```json
{
  "login": "xrf_creator",
  "password": "creator"
}
```

Ответ содержит:
- `token_type: "Bearer"`
- `token`
- `expires_at`

Использование в следующих запросах:

```text
Authorization: Bearer <token>
```

## Переменные окружения

См. `.env.example`:
- `JWT_SECRET`
- `JWT_TTL_MINUTES`
- `APP_PORT`

