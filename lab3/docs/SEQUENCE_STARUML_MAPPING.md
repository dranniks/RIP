# Построчная разметка стрелок для StarUML (Lab4, XRF)

Ниже готовый текст для подписей стрелок в sequence-диаграмме по вашему реальному API.  
Формат: **номер** / **тип линии** / **откуда -> куда** / **что писать на стрелке**.

Линии жизни (слева направо):
1. `Фронтенд (создатель)`
2. `Фронтенд (модератор)`
3. `: Домен Users`
4. `: Домен Services`
5. `: Домен Claims`
6. `: Домен M-M (ClaimItems)`
7. `: Расчет XRF (внутри FormClaim)`

1. `сплошная` / создатель -> users / `POST /api/users/auth {login,password}`
2. `пунктир` / users -> создатель / `200 JSON: token_type, token, role, session_id, expires_at`
3. `сплошная` / создатель -> services / `GET /api/services?q=bronze`
4. `пунктир` / services -> создатель / `200 JSON: services[]`
5. `сплошная` / создатель -> claims / `GET /api/claims/cart-icon`
6. `пунктир` / claims -> создатель / `200 JSON: {claim_id?, claim_code?, service_count}`
7. `сплошная` / создатель -> m-m / `POST /api/claim-items {service_id: 1}`
8. `пунктир` / m-m -> создатель / `201 JSON: {claim_id, claim_code, service_id, quantity, sort_order}`
9. `сплошная` / создатель -> m-m / `POST /api/claim-items {service_id: 2}`
10. `пунктир` / m-m -> создатель / `201 JSON: second service added`
11. `сплошная` / создатель -> claims / `GET /api/claims/{claim_id}`
12. `пунктир` / claims -> создатель / `200 JSON: черновик заявки + 2 услуги`
13. `сплошная` / создатель -> m-m / `PUT /api/claim-items/{service_id_1} {quantity,sort_order}`
14. `пунктир` / m-m -> создатель / `200 JSON: элемент m-m обновлен`
15. `сплошная` / создатель -> claims / `PUT /api/claims/{claim_id} {operator_comment, cu_measured, zn_measured, sn_measured, pb_measured}`
16. `пунктир` / claims -> создатель / `200 JSON: черновик обновлен`
17. `сплошная` / создатель -> claims / `PUT /api/claims/{claim_id}/moderate {action:"complete"}`
18. `пунктир` / claims -> создатель / `403 JSON: недостаточно прав (создатель не может модерировать)`
19. `сплошная` / создатель -> claims / `PUT /api/claims/{claim_id}/form`
20. `сплошная` / claims -> расчет xrf / `calculateCompositionAndScore(claim_id)`
21. `пунктир` / расчет xrf -> claims / `result_value[] + best_match_label + total_cost`
22. `пунктир` / claims -> создатель / `200 JSON: заявка сформирована (status, formed_at, completion_formula_result, total_cost)`
23. `сплошная` / модератор -> users / `POST /api/users/auth {login,password}`
24. `пунктир` / users -> модератор / `200 JSON: token_type, token, role=moderator, session_id`
25. `сплошная` / модератор -> claims / `PUT /api/claims/{claim_id}/moderate {action:"complete"}`
26. `пунктир` / claims -> модератор / `200 JSON: status="completed", moderator_id, completed_at`
27. `сплошная` / модератор -> claims / `GET /api/claims`
28. `пунктир` / claims -> модератор / `200 JSON: все заявки для модератора`

## Что заменяет Message1..Message13 из примера
- `Message1` -> `POST /api/users/auth {login,password}`
- `Message2` -> `GET /api/services?q=bronze`
- `Message3` -> `POST /api/claim-items {service_id: 1}`
- `Message4` -> `POST /api/claim-items {service_id: 2}`
- `Message5` -> `GET /api/claims/cart-icon`
- `Message6` -> `PUT /api/claims/{claim_id}`
- `Message7` -> `PUT /api/claims/{claim_id}/moderate {action:"complete"}` (модератор)
- `Message8` -> `PUT /api/claims/{claim_id}/moderate {action:"complete"}` (создатель -> 403)
- `Message9` -> `PUT /api/claims/{claim_id}/form`
- `Message10` -> `GET /api/claims`
- `Message11` -> `GET /api/claims` (проверка после завершения)
- `Message12` -> `GET /api/claims/{claim_id}`
- `Message13` -> `PUT /api/claim-items/{service_id_1} {quantity,sort_order}`

Примечание по соответствию коду:
- Авторизация и права берутся из `Authorization: Bearer <JWT>` + проверка сессии в Redis.
- Расчет (`result_value`, `completion_formula_result`, `total_cost`) выполняется в `FormDraftClaim`.
