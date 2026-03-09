# Lab3: Class Diagram and Backend Detalization

## Domains (URL interfaces)

- `GET /api/services`
- `GET /api/services/:id`
- `POST /api/services`

- `POST /api/claim-items`
- `PUT /api/claim-items/:service_id`
- `DELETE /api/claim-items/:service_id`

- `GET /api/claims/cart-icon`
- `GET /api/claims`
- `GET /api/claims/:id`
- `PUT /api/claims/:id`
- `PUT /api/claims/:id/form`
- `PUT /api/claims/:id/moderate`
- `DELETE /api/claims/:id`

- `POST /api/users/register`
- `POST /api/users/auth`
- `POST /api/users/logout`

## Class Diagram (Mermaid)

```mermaid
classDiagram
    class Handler {
      +GetServicesAPI()
      +GetServiceAPI()
      +CreateServiceAPI()
      +AddServiceToDraftAPI()
      +UpdateDraftMatchAPI()
      +DeleteDraftMatchAPI()
      +GetCartIconAPI()
      +GetClaimsAPI()
      +GetClaimAPI()
      +UpdateDraftClaimAPI()
      +FormClaimAPI()
      +ModerateClaimAPI()
      +DeleteDraftClaimAPI()
      +RegisterUserAPI()
      +AuthStubAPI()
      +LogoutStubAPI()
    }

    class Repository {
      +ListServices(filters)
      +GetServiceByID(id)
      +CreateService(input)
      +UploadServiceMedia(...)
      +AddServiceToDraft(creatorID, serviceID)
      +UpdateDraftMatch(creatorID, serviceID, input)
      +DeleteDraftMatch(creatorID, serviceID)
      +GetCartIcon(creatorID)
      +ListClaims(filters)
      +GetClaimDetails(claimID)
      +UpdateDraftClaimFields(creatorID, claimID, input)
      +FormDraftClaim(creatorID, claimID)
      +ModerateFormedClaim(moderatorID, claimID, action)
      +DeleteDraftClaim(creatorID, claimID)
      +RegisterUser(input)
      +AuthenticateStub(input)
    }

    class User {
      +id
      +login
      +full_name
      +password_hash
      +role
      +created_at
    }

    class ReferenceAlloyService {
      +id
      +slug
      +name
      +description
      +status
      +image_file_name
      +video_file_name
      +image_url
      +video_url
      +era
      +culture
      +unit_price
      +cu_reference
      +zn_reference
      +sn_reference
      +pb_reference
      +created_at
      +updated_at
    }

    class ArtifactClaim {
      +id
      +claim_code
      +status
      +created_at
      +formed_at
      +completed_at
      +creator_id
      +moderator_id
      +artifact_title
      +artifact_origin
      +analyzer_model
      +operator_comment
      +cu_measured
      +zn_measured
      +sn_measured
      +pb_measured
      +best_match_label
      +completion_formula_result
      +total_cost
      +planned_delivery_at
    }

    class ClaimAlloyMatch {
      +id
      +claim_id
      +service_id
      +quantity
      +sort_order
      +match_value
      +composition_result
      +match_score
      +created_at
      +updated_at
    }

    class UsersSingleton {
      +CurrentUsers()
      +Creator.ID=1
      +Moderator.ID=2
    }

    Handler --> Repository
    Handler --> UsersSingleton

    User "1" <-- "many" ArtifactClaim : creator_id
    User "0..1" <-- "many" ArtifactClaim : moderator_id
    ArtifactClaim "1" <-- "many" ClaimAlloyMatch : claim_id
    ReferenceAlloyService "1" <-- "many" ClaimAlloyMatch : service_id
```

## DB Tables

- `users`
- `reference_alloy_services`
- `artifact_claims`
- `claim_alloy_matches`

## Relation Variants in Code

- Methods use different models: service endpoints use `ReferenceAlloyService`, claim endpoints combine `ArtifactClaim` + `ClaimAlloyMatch`.
- Models use other models: `ArtifactClaim` references `User` (creator/moderator), `ClaimAlloyMatch` references both claim and service.
- Models use multiple tables: claim list endpoint joins `artifact_claims`, `users`, and aggregate on `claim_alloy_matches`.
