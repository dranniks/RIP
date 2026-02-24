# ER-спецификация для StarUML (Лабораторная 2)

## Таблица `users`
- `id BIGSERIAL` PK
- `login VARCHAR(64)` UNIQUE, NOT NULL
- `full_name VARCHAR(128)` NOT NULL
- `role VARCHAR(32)` NOT NULL
- `created_at TIMESTAMP` NOT NULL

## Таблица `reference_alloy_services`
- `id BIGSERIAL` PK
- `slug VARCHAR(120)` UNIQUE, NOT NULL
- `name VARCHAR(160)` NOT NULL
- `description TEXT` NOT NULL
- `status VARCHAR(16)` NOT NULL, CHECK (`действует`, `удален`)
- `image_url VARCHAR(255)` NULL
- `video_url VARCHAR(255)` NULL
- `era VARCHAR(100)` NOT NULL
- `culture VARCHAR(120)` NOT NULL
- `cu_reference NUMERIC(6,3)` NOT NULL
- `zn_reference NUMERIC(6,3)` NOT NULL
- `sn_reference NUMERIC(6,3)` NOT NULL
- `pb_reference NUMERIC(6,3)` NOT NULL
- `updated_at TIMESTAMP` NOT NULL

## Таблица `artifact_claims`
- `id BIGSERIAL` PK, NOT NULL
- `claim_code VARCHAR(40)` UNIQUE
- `status VARCHAR(16)` NOT NULL, CHECK (`черновик`, `удален`, `сформирован`, `завершен`, `отклонен`)
- `created_at TIMESTAMP` NOT NULL
- `creator_id BIGINT` FK -> `users.id`, NOT NULL
- `formed_at TIMESTAMP` NULL
- `completed_at TIMESTAMP` NULL
- `moderator_id BIGINT` FK -> `users.id`, NULL
- `artifact_title VARCHAR(180)` NULL
- `artifact_origin VARCHAR(180)` NULL
- `analyzer_model VARCHAR(120)` NULL
- `operator_comment VARCHAR(255)` NULL
- `cu_measured NUMERIC(6,3)` NULL
- `zn_measured NUMERIC(6,3)` NULL
- `sn_measured NUMERIC(6,3)` NULL
- `pb_measured NUMERIC(6,3)` NULL
- `best_match_label VARCHAR(180)` NULL
- `completion_formula_result NUMERIC(8,2)` NULL

Индекс:
- `UNIQUE (creator_id) WHERE status = 'черновик'` (не более одной черновой заявки на пользователя)

## Таблица `claim_alloy_matches` (m-m)
- `id BIGSERIAL` PK
- `claim_id BIGINT` FK -> `artifact_claims.id`, NOT NULL
- `service_id BIGINT` FK -> `reference_alloy_services.id`, NOT NULL
- `quantity INT` NOT NULL
- `sort_order INT` NOT NULL
- `is_primary BOOLEAN` NOT NULL
- `composition_result VARCHAR(200)` NULL
- `match_score NUMERIC(8,2)` NULL

Составной уникальный ключ:
- `UNIQUE (claim_id, service_id)`

## Связи
- `users (1) -> (N) artifact_claims` по `creator_id`
- `users (1) -> (N) artifact_claims` по `moderator_id`
- `artifact_claims (1) -> (N) claim_alloy_matches`
- `reference_alloy_services (1) -> (N) claim_alloy_matches`

## Правила удаления
- Каскадное удаление запрещено.
- Во всех FK используется `ON DELETE RESTRICT`.
