package repository

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"xrfApp/internal/app/model"
)

type CartIcon struct {
	ClaimID      *uint
	ClaimCode    *string
	ServiceCount int64
}

type MatchUpdateInput struct {
	Quantity  *int
	SortOrder *int
}

type ClaimUpdateInput struct {
	OperatorComment *string
	CuMeasured      *float64
	ZnMeasured      *float64
	SnMeasured      *float64
	PbMeasured      *float64
}

type ClaimFilters struct {
	Status     string
	FormedFrom *time.Time
	FormedTo   *time.Time
	ViewerID   uint
	ViewerRole string
}

type ClaimListItem struct {
	ID                      uint
	ClaimCode               string
	Status                  string
	CreatedAt               time.Time
	FormedAt                *time.Time
	CompletedAt             *time.Time
	CreatorLogin            string
	ModeratorLogin          *string
	CompletionFormulaResult *float64
	ResultValue             *float64
	BestMatchLabel          *string
	TotalCost               *float64
	ResultItemsCount        int64
}

type ClaimServiceItem struct {
	ServiceID       uint
	ServiceSlug     string
	ServiceName     string
	ServiceImageURL *string
	ServiceVideoURL *string
	Quantity        int
	SortOrder       int
	ResultValue     *float64
}

type ClaimDetails struct {
	Claim            model.ArtifactClaim
	CreatorLogin     string
	ModeratorLogin   *string
	Services         []ClaimServiceItem
	ResultItemsCount int64
}

func (r *Repository) GetCartIcon(creatorID uint) (*CartIcon, error) {
	claim := model.ArtifactClaim{}
	err := r.db.
		Where("creator_id = ? AND status = ?", creatorID, model.ClaimStatusDraft).
		First(&claim).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &CartIcon{ServiceCount: 0}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get draft claim: %w", err)
	}

	var serviceCount int64
	if countErr := r.db.
		Model(&model.ClaimAlloyMatch{}).
		Where("claim_id = ?", claim.ID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&serviceCount).Error; countErr != nil {
		return nil, fmt.Errorf("count claim services: %w", countErr)
	}

	return &CartIcon{
		ClaimID:      &claim.ID,
		ClaimCode:    &claim.ClaimCode,
		ServiceCount: serviceCount,
	}, nil
}

func (r *Repository) AddServiceToDraft(creatorID uint, serviceID uint) (*model.ArtifactClaim, *model.ClaimAlloyMatch, error) {
	var outClaim *model.ArtifactClaim
	var outMatch *model.ClaimAlloyMatch

	err := r.db.Transaction(func(tx *gorm.DB) error {
		service := model.ReferenceAlloyService{}
		if serviceErr := tx.
			Where("id = ? AND status <> ?", serviceID, model.ServiceStatusDeleted).
			First(&service).Error; serviceErr != nil {
			return serviceErr
		}

		claim, claimErr := r.findOrCreateDraftClaimTx(tx, creatorID)
		if claimErr != nil {
			return claimErr
		}

		match := model.ClaimAlloyMatch{}
		matchTx := tx.
			Where("claim_id = ? AND service_id = ?", claim.ID, service.ID).
			Limit(1).
			Find(&match)
		if matchTx.Error != nil {
			return fmt.Errorf("get claim-service row: %w", matchTx.Error)
		}
		if matchTx.RowsAffected == 0 {
			maxSort := 0
			if err := tx.
				Model(&model.ClaimAlloyMatch{}).
				Where("claim_id = ?", claim.ID).
				Select("COALESCE(MAX(sort_order), 0)").
				Scan(&maxSort).Error; err != nil {
				return fmt.Errorf("get max sort order: %w", err)
			}

			match = model.ClaimAlloyMatch{
				ClaimID:   claim.ID,
				ServiceID: service.ID,
				Quantity:  1,
				SortOrder: maxSort + 1,
			}
			if err := tx.Create(&match).Error; err != nil {
				return fmt.Errorf("create claim-service row: %w", err)
			}
		} else {
			match.Quantity++
			if err := tx.Save(&match).Error; err != nil {
				return fmt.Errorf("update claim-service quantity: %w", err)
			}
		}

		outClaim = claim
		outMatch = &match
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("add service to draft: %w", err)
	}

	return outClaim, outMatch, nil
}

func (r *Repository) UpdateDraftMatch(
	creatorID uint,
	serviceID uint,
	input MatchUpdateInput,
) (*model.ClaimAlloyMatch, error) {
	updated := model.ClaimAlloyMatch{}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		match := model.ClaimAlloyMatch{}
		findTx := tx.
			Joins("JOIN artifact_claims c ON c.id = claim_alloy_matches.claim_id").
			Where(
				"claim_alloy_matches.service_id = ? AND c.creator_id = ? AND c.status = ?",
				serviceID,
				creatorID,
				model.ClaimStatusDraft,
			).
			Limit(1).
			Find(&match)
		if findTx.Error != nil {
			return findTx.Error
		}
		if findTx.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if input.Quantity != nil {
			if *input.Quantity <= 0 {
				return fmt.Errorf("%w: quantity must be greater than zero", ErrValidation)
			}
			match.Quantity = *input.Quantity
		}
		if input.SortOrder != nil {
			if *input.SortOrder < 0 {
				return fmt.Errorf("%w: sort_order cannot be negative", ErrValidation)
			}
			match.SortOrder = *input.SortOrder
		}
		if saveErr := tx.Save(&match).Error; saveErr != nil {
			return fmt.Errorf("save claim-service row: %w", saveErr)
		}

		updated = match
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *Repository) DeleteDraftMatch(creatorID uint, serviceID uint) error {
	tx := r.db.
		Table("claim_alloy_matches").
		Joins("JOIN artifact_claims c ON c.id = claim_alloy_matches.claim_id").
		Where(
			"claim_alloy_matches.service_id = ? AND c.creator_id = ? AND c.status = ?",
			serviceID,
			creatorID,
			model.ClaimStatusDraft,
		).
		Delete(&model.ClaimAlloyMatch{})
	if tx.Error != nil {
		return fmt.Errorf("delete claim-service row: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) ListClaims(filters ClaimFilters) ([]ClaimListItem, error) {
	query := r.db.
		Table("artifact_claims AS c").
		Select(
			"c.id, c.claim_code, c.status, c.created_at, c.formed_at, c.completed_at, "+
				"creator.login AS creator_login, moderator.login AS moderator_login, "+
				"c.completion_formula_result, "+
				"COALESCE(c.completion_formula_result, (SELECT ROUND(MAX(m.result_value), 2) FROM claim_alloy_matches m WHERE m.claim_id = c.id)) AS result_value, "+
				"c.best_match_label, c.total_cost, "+
				"(SELECT COUNT(1) FROM claim_alloy_matches m WHERE m.claim_id = c.id AND m.result_value IS NOT NULL) AS result_items_count",
		).
		Joins("JOIN users AS creator ON creator.id = c.creator_id").
		Joins("LEFT JOIN users AS moderator ON moderator.id = c.moderator_id").
		Where("c.status NOT IN ?", []string{model.ClaimStatusDeleted, model.ClaimStatusDraft}).
		Order("c.id DESC")

	viewerRole := strings.ToLower(strings.TrimSpace(filters.ViewerRole))
	if viewerRole != "moderator" {
		if filters.ViewerID == 0 {
			return nil, fmt.Errorf("%w: viewer id is required", ErrValidation)
		}
		query = query.Where("c.creator_id = ?", filters.ViewerID)
	}

	if strings.TrimSpace(filters.Status) != "" {
		query = query.Where("c.status = ?", strings.TrimSpace(filters.Status))
	}
	if filters.FormedFrom != nil {
		query = query.Where("c.formed_at >= ?", *filters.FormedFrom)
	}
	if filters.FormedTo != nil {
		query = query.Where("c.formed_at <= ?", *filters.FormedTo)
	}

	var rows []ClaimListItem
	if err := query.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list claims: %w", err)
	}
	return rows, nil
}

func (r *Repository) GetClaimDetails(claimID uint, viewerID uint, viewerRole string) (*ClaimDetails, error) {
	if claimID == 0 {
		return nil, fmt.Errorf("%w: claim id is required", ErrValidation)
	}

	claim := model.ArtifactClaim{}
	query := r.db.
		Preload("Creator").
		Preload("Moderator").
		Where("id = ? AND status <> ?", claimID, model.ClaimStatusDeleted)

	if strings.ToLower(strings.TrimSpace(viewerRole)) != "moderator" {
		if viewerID == 0 {
			return nil, fmt.Errorf("%w: viewer id is required", ErrValidation)
		}
		query = query.Where("creator_id = ?", viewerID)
	}

	if err := query.First(&claim).Error; err != nil {
		return nil, err
	}

	var links []model.ClaimAlloyMatch
	if err := r.db.
		Preload("Service").
		Where("claim_id = ?", claim.ID).
		Order("sort_order ASC, service_id ASC").
		Find(&links).Error; err != nil {
		return nil, fmt.Errorf("load claim services: %w", err)
	}

	services := make([]ClaimServiceItem, 0, len(links))
	var resultCount int64
	for _, link := range links {
		if link.ResultValue != nil {
			resultCount++
		}

		services = append(services, ClaimServiceItem{
			ServiceID:       link.ServiceID,
			ServiceSlug:     link.Service.Slug,
			ServiceName:     link.Service.Name,
			ServiceImageURL: link.Service.ImageURL,
			ServiceVideoURL: link.Service.VideoURL,
			Quantity:        link.Quantity,
			SortOrder:       link.SortOrder,
			ResultValue:     link.ResultValue,
		})
	}

	result := &ClaimDetails{
		Claim:            claim,
		CreatorLogin:     claim.Creator.Login,
		Services:         services,
		ResultItemsCount: resultCount,
	}
	if claim.Moderator != nil {
		result.ModeratorLogin = &claim.Moderator.Login
	}

	return result, nil
}

func (r *Repository) UpdateDraftClaimFields(
	creatorID uint,
	claimID uint,
	input ClaimUpdateInput,
) error {
	updates := map[string]any{}

	if input.OperatorComment != nil {
		updates["operator_comment"] = normalizeNullableString(*input.OperatorComment)
	}
	if input.CuMeasured != nil {
		updates["cu_measured"] = input.CuMeasured
	}
	if input.ZnMeasured != nil {
		updates["zn_measured"] = input.ZnMeasured
	}
	if input.SnMeasured != nil {
		updates["sn_measured"] = input.SnMeasured
	}
	if input.PbMeasured != nil {
		updates["pb_measured"] = input.PbMeasured
	}

	if len(updates) == 0 {
		return nil
	}

	tx := r.db.
		Model(&model.ArtifactClaim{}).
		Where("id = ? AND creator_id = ? AND status = ?", claimID, creatorID, model.ClaimStatusDraft).
		Updates(updates)
	if tx.Error != nil {
		return fmt.Errorf("update draft claim: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *Repository) FormDraftClaim(creatorID uint, claimID uint) (*model.ArtifactClaim, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		claim := model.ArtifactClaim{}
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND creator_id = ? AND status = ?", claimID, creatorID, model.ClaimStatusDraft).
			First(&claim).Error; err != nil {
			return err
		}

		matches := []model.ClaimAlloyMatch{}
		if err := tx.
			Preload("Service").
			Where("claim_id = ?", claim.ID).
			Order("sort_order ASC, service_id ASC").
			Find(&matches).Error; err != nil {
			return fmt.Errorf("load claim services for form: %w", err)
		}

		if len(matches) == 0 {
			return fmt.Errorf("%w: draft claim has no services", ErrValidation)
		}
		if err := validateMandatoryClaimFields(claim); err != nil {
			return err
		}

		totalQty := 0
		totalCost := 0.0
		bestScore := -1.0
		bestLabel := ""

		for _, row := range matches {
			composition := calculateCompositionFormula(claim, row.Service)
			score := calculateMatchScore(composition, row.Service)
			row.ResultValue = &score
			if err := tx.Model(&model.ClaimAlloyMatch{}).Where(
				"claim_id = ? AND service_id = ?",
				row.ClaimID,
				row.ServiceID,
			).Updates(map[string]any{
				"result_value": row.ResultValue,
			}).Error; err != nil {
				return fmt.Errorf("update m-m result for service %d: %w", row.ServiceID, err)
			}

			totalQty += row.Quantity
			totalCost += float64(row.Quantity) * row.Service.UnitPrice
			if score > bestScore {
				bestScore = score
				bestLabel = row.Service.Name
			}
		}

		if totalQty == 0 {
			return fmt.Errorf("%w: draft claim has zero quantity", ErrValidation)
		}

		now := time.Now()
		bestScore = round2(bestScore)
		totalCost = round2(totalCost)

		update := map[string]any{
			"status":                    model.ClaimStatusFormed,
			"formed_at":                 now,
			"completion_formula_result": bestScore,
			"best_match_label":          bestLabel,
			"total_cost":                totalCost,
		}
		if err := tx.Model(&model.ArtifactClaim{}).Where("id = ?", claim.ID).Updates(update).Error; err != nil {
			return fmt.Errorf("set formed status: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	claim := model.ArtifactClaim{}
	if err := r.db.Where("id = ?", claimID).First(&claim).Error; err != nil {
		return nil, err
	}
	return &claim, nil
}

func (r *Repository) ModerateFormedClaim(moderatorID uint, claimID uint, action string) (*model.ArtifactClaim, error) {
	normalized := strings.ToLower(strings.TrimSpace(action))
	var targetStatus string

	switch normalized {
	case "complete", "completed", "finish", "approve", "завершить":
		targetStatus = model.ClaimStatusCompleted
	case "reject", "rejected", "decline", "отклонить":
		targetStatus = model.ClaimStatusRejected
	default:
		return nil, fmt.Errorf("%w: action must be complete or reject", ErrValidation)
	}

	update := r.db.
		Model(&model.ArtifactClaim{}).
		Where("id = ? AND status = ?", claimID, model.ClaimStatusFormed).
		Updates(map[string]any{
			"status":       targetStatus,
			"moderator_id": moderatorID,
			"completed_at": time.Now(),
		})
	if update.Error != nil {
		return nil, fmt.Errorf("moderate claim: %w", update.Error)
	}
	if update.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: only formed claim can be moderated", ErrInvalidTransition)
	}

	claim := model.ArtifactClaim{}
	if err := r.db.Where("id = ?", claimID).First(&claim).Error; err != nil {
		return nil, err
	}

	return &claim, nil
}

func (r *Repository) DeleteDraftClaim(creatorID uint, claimID uint) error {
	update := r.db.
		Model(&model.ArtifactClaim{}).
		Where("id = ? AND creator_id = ? AND status = ?", claimID, creatorID, model.ClaimStatusDraft).
		Update("status", model.ClaimStatusDeleted)
	if update.Error != nil {
		return fmt.Errorf("delete draft claim: %w", update.Error)
	}
	if update.RowsAffected == 0 {
		return fmt.Errorf("%w: only draft claim can be deleted by creator", ErrInvalidTransition)
	}
	return nil
}

func (r *Repository) findOrCreateDraftClaimTx(tx *gorm.DB, creatorID uint) (*model.ArtifactClaim, error) {
	claim := model.ArtifactClaim{}
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("creator_id = ? AND status = ?", creatorID, model.ClaimStatusDraft).
		First(&claim).Error
	if err == nil {
		return &claim, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	claim = model.ArtifactClaim{
		ClaimCode: fmt.Sprintf("TMP-%d", time.Now().UnixNano()),
		Status:    model.ClaimStatusDraft,
		CreatorID: creatorID,
	}
	if err := tx.Create(&claim).Error; err != nil {
		return nil, err
	}

	code := generateClaimCode(claim.ID)
	if err := tx.Model(&claim).Update("claim_code", code).Error; err != nil {
		return nil, err
	}
	claim.ClaimCode = code

	return &claim, nil
}

func normalizeNullableString(v string) any {
	value := strings.TrimSpace(v)
	if value == "" {
		return nil
	}
	return value
}

func validateMandatoryClaimFields(claim model.ArtifactClaim) error {
	if claim.OperatorComment == nil || strings.TrimSpace(*claim.OperatorComment) == "" {
		return fmt.Errorf("%w: operator_comment is required before form", ErrValidation)
	}
	if claim.CuMeasured == nil || claim.ZnMeasured == nil || claim.SnMeasured == nil || claim.PbMeasured == nil {
		return fmt.Errorf("%w: measured values (cu/zn/sn/pb) are required before form", ErrValidation)
	}
	return nil
}

type compositionVector struct {
	Cu float64
	Zn float64
	Sn float64
	Pb float64
}

func calculateCompositionFormula(claim model.ArtifactClaim, service model.ReferenceAlloyService) compositionVector {
	cuTerm := safeDivision(valueOrZero(claim.CuMeasured), service.CuReference)
	znTerm := safeDivision(valueOrZero(claim.ZnMeasured), service.ZnReference)
	snTerm := safeDivision(valueOrZero(claim.SnMeasured), service.SnReference)
	pbTerm := safeDivision(valueOrZero(claim.PbMeasured), service.PbReference)

	denominator := cuTerm + znTerm + snTerm + pbTerm
	if denominator <= 0 {
		return compositionVector{}
	}

	return compositionVector{
		Cu: round2((cuTerm / denominator) * 100),
		Zn: round2((znTerm / denominator) * 100),
		Sn: round2((snTerm / denominator) * 100),
		Pb: round2((pbTerm / denominator) * 100),
	}
}

func calculateMatchScore(composition compositionVector, service model.ReferenceAlloyService) float64 {
	ref := normalizeReference(service)
	diff := math.Abs(composition.Cu-ref.Cu) +
		math.Abs(composition.Zn-ref.Zn) +
		math.Abs(composition.Sn-ref.Sn) +
		math.Abs(composition.Pb-ref.Pb)

	score := 100 - diff*0.25
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return round2(score)
}

func normalizeReference(service model.ReferenceAlloyService) compositionVector {
	sum := service.CuReference + service.ZnReference + service.SnReference + service.PbReference
	if sum <= 0 {
		return compositionVector{}
	}
	return compositionVector{
		Cu: (service.CuReference / sum) * 100,
		Zn: (service.ZnReference / sum) * 100,
		Sn: (service.SnReference / sum) * 100,
		Pb: (service.PbReference / sum) * 100,
	}
}

func safeDivision(value float64, ref float64) float64 {
	if ref == 0 {
		return 0
	}
	return value / ref
}

func valueOrZero(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
