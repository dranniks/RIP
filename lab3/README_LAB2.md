# Лабораторная 2: PostgreSQL + ORM + SQL UPDATE (XRF)

## 1) Что реализовано в коде

- 4 таблицы БД:
  - `users`
  - `reference_alloy_services`
  - `artifact_claims`
  - `claim_alloy_matches`
- Статусы заявок: `черновик`, `удален`, `сформирован`, `завершен`, `отклонен`.
- Ограничение: у одного пользователя не более одной заявки со статусом `черновик`.
- m-m таблица `claim_alloy_matches` с составным уникальным ключом `(claim_id, service_id)`.
- Каскадное удаление отсутствует (`RESTRICT`).
- 5 HTTP-методов:
  - `GET /services` (получение + поиск услуг, ORM)
  - `GET /services/:slug` (карточка услуги, ORM)
  - `GET /claims/:code` (просмотр заявки, ORM)
  - `POST /claims/add-service` (добавление услуги в черновик, ORM)
  - `POST /claims/:code/delete` (логическое удаление заявки, SQL UPDATE без ORM)
- Если черновика нет, он создается только при добавлении услуги.
- Удаленные заявки не открываются по URL.
- Поле `completion_formula_result` рассчитывается при переводе заявки в `завершен` через триггер БД.

## 2) Формула расчета

Используется калибровочная формула из методички:

`Ci = (Ii / Ki) / Σ(Ij / Kj) * 100%`

- `Ii` - измеренная интенсивность пика элемента.
- `Ki` - калибровочный коэффициент из выбранной эталонной услуги.
- `Ci` - оценка массовой доли элемента в процентах.
- Для каждой услуги в заявке расчет выполняется по ее `Ki` и введенным `Ii`.

## 3) Запуск окружения

Из корня проекта:

```powershell
docker compose up -d
```

Сервисы:
- PostgreSQL (Docker): `localhost:5433`
- Adminer: `http://localhost:8081`
- pgAdmin (web): `http://localhost:5050`
- MinIO: `http://localhost:9000` и `http://localhost:9001`

Если MinIO отдает `403` на `http://localhost:9000/xrf-media/...`, открой bucket на чтение:
```powershell
docker exec rip-minio mc alias set local http://localhost:9000 root rootroot
docker exec rip-minio mc anonymous set download local/xrf-media
```

Если контейнер БД ранее запускался с другим паролем/пользователем, сбрось volume:
```powershell
docker compose down -v
docker compose up -d
```

## 4) Подключение к БД через pgAdmin

### Вариант A: встроенный pgAdmin из docker-compose
1. Открой `http://localhost:5050`.
2. Логин:
   - Email: `admin@xrf.local`
   - Password: `admin`
3. Правый клик на `Servers` -> `Register` -> `Server...`
4. Вкладка `General`:
   - Name: `RIP-Lab2`
5. Вкладка `Connection`:
   - Host name/address: `db` (если pgAdmin в docker) или `localhost` (если desktop pgAdmin)
   - Port: `5432` (если pgAdmin в docker) или `5433` (если desktop pgAdmin)
   - Maintenance database: `RIP`
   - Username: `root`
   - Password: `root`
6. Нажми `Save`.

### Вариант B: desktop pgAdmin
- Те же шаги, но host ставь `localhost`.

## 5) Подключение через Adminer (для обязательного показа)

1. Открой `http://localhost:8081`.
2. Введи:
   - System: `PostgreSQL`
   - Server: `db` (если используешь встроенный Adminer в docker)
   - Username: `root`
   - Password: `root`
   - Database: `RIP`
3. Нажми `Login`.

## 6) Запуск приложения

Из папки `lab2`:

```powershell
go run ./cmd/xrf-app
```

Приложение: `http://localhost:8080/services`

Если порт `8080` занят, проверь:
```powershell
netstat -ano | findstr :8080
```
и останови процесс/сервис, который его занял.

## 7) Сценарий показа (в порядке защиты)

1. Показать формулу расчета (`/claims/:code`, блок "Формула расчета").
2. Открыть Adminer и добавить новую услугу в `reference_alloy_services`.
3. Выполнить `SELECT` по таблице услуг и показать, что запись появилась.
4. Открыть страницу услуг `/services` и показать поиск (`?q=...`).
5. Добавить 2 услуги в заявку кнопками "В заявку" (карточка корзины активируется и счетчик меняется).
6. Перейти в текущую заявку `/claims/{code}` и показать содержимое заявки.
7. Нажать "Логически удалить заявку" (POST -> SQL UPDATE).
8. Вручную открыть URL удаленной заявки `/claims/{code}` и показать, что она недоступна.
9. В БД показать запросом:
   - логическое удаление (статус `удален`)
   - новую заявку, созданную после повторного добавления услуги.
10. В БД изменить поля предметной области в `artifact_claims` и `claim_alloy_matches`, обновить страницу заявки и показать изменения в приложении.
11. В коде показать:
   - модели (`internal/app/model/models.go`)
   - составной уникальный ключ m-m (`ux_claim_service`)
   - 4 ORM-контроллера (3 GET + POST add)
   - SQL UPDATE удаления (`SoftDeleteDraftClaimSQL`)

## 8) SQL-запросы для демонстрации

### Показать услуги
```sql
SELECT id, slug, name, status, era, culture
FROM reference_alloy_services
ORDER BY id;
```

### Добавить услугу в Adminer/pgAdmin
```sql
INSERT INTO reference_alloy_services
(slug, name, description, status, image_url, era, culture, cu_reference, zn_reference, sn_reference, pb_reference)
VALUES
('alloy-bronze-scynthia', 'Бронза Скифии', 'Новый эталон по данным XRF', 'действует', NULL,
 'VII-IV вв. до н.э.', 'Скифская культура', 0.810, 0.050, 0.270, 0.110);
```

### Показать заявки и статус логического удаления
```sql
SELECT id, claim_code, status, created_at, formed_at, completed_at, creator_id, moderator_id
FROM artifact_claims
ORDER BY id DESC;
```

### Показать m-m строки заявки
```sql
SELECT m.id, c.claim_code, s.slug, m.quantity, m.sort_order, m.is_primary, m.composition_result, m.match_score
FROM claim_alloy_matches m
JOIN artifact_claims c ON c.id = m.claim_id
JOIN reference_alloy_services s ON s.id = m.service_id
ORDER BY m.id DESC;
```

### Изменить поля предметной области в заявке
```sql
UPDATE artifact_claims
SET artifact_title = 'Фибула с инкрустацией',
    artifact_origin = 'Северное Причерноморье',
    analyzer_model = 'Bruker Tracer 5i'
WHERE claim_code = 'ВАШ_КОД_ЗАЯВКИ';
```

### Изменить m-m поля
```sql
UPDATE claim_alloy_matches
SET quantity = 2,
    sort_order = 1,
    is_primary = true,
    composition_result = 'Cu 78.5%, Zn 19.8%, Sn 0.7%, Pb 1.0%',
    match_score = 96.20
WHERE id = 1;
```

### Завершить заявку и получить расчетное поле completion_formula_result
```sql
UPDATE artifact_claims
SET status = 'завершен', completed_at = NOW(), moderator_id = 2
WHERE claim_code = 'ВАШ_КОД_ЗАЯВКИ';
```

После этого:
```sql
SELECT claim_code, status, completion_formula_result
FROM artifact_claims
WHERE claim_code = 'ВАШ_КОД_ЗАЯВКИ';
```

## 9) Файлы для показа на защите

- Модели: `internal/app/model/models.go`
- Репозиторий + ORM + SQL UPDATE: `internal/app/repository/repository.go`
- Контроллеры: `internal/app/handler/handler.go`
- Роуты: `internal/api/server.go`
- SQL-миграция: `migrations/001_init.sql`
- ER-спецификация для StarUML: `docs/ER_SPEC.md`

## 10) Важно

Если у тебя в среде нет доступа к интернету для Go-модулей, `go mod tidy` может не скачать GORM.
На своей машине с интернетом выполни:

```powershell
go mod tidy
go run ./cmd/xrf-app
```
