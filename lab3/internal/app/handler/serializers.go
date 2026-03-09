package handler

import (
	"time"

	"xrfApp/internal/app/model"
	"xrfApp/internal/app/repository"
)

type serviceSerializer struct {
	ID            uint      `json:"id"`
	Slug          string    `json:"slug"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	ImageFileName *string   `json:"image_file_name"`
	VideoFileName *string   `json:"video_file_name"`
	ImageURL      *string   `json:"image_url"`
	VideoURL      *string   `json:"video_url"`
	Era           string    `json:"era"`
	Culture       string    `json:"culture"`
	UnitPrice     float64   `json:"unit_price"`
	CuReference   float64   `json:"cu_reference"`
	ZnReference   float64   `json:"zn_reference"`
	SnReference   float64   `json:"sn_reference"`
	PbReference   float64   `json:"pb_reference"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type cartSerializer struct {
	ClaimID      *uint   `json:"claim_id"`
	ClaimCode    *string `json:"claim_code"`
	ServiceCount int64   `json:"service_count"`
}

type claimItemSerializer struct {
	ID                uint     `json:"id"`
	ServiceID         uint     `json:"service_id"`
	ServiceSlug       string   `json:"service_slug"`
	ServiceName       string   `json:"service_name"`
	ServiceImageURL   *string  `json:"service_image_url"`
	ServiceVideoURL   *string  `json:"service_video_url"`
	Quantity          int      `json:"quantity"`
	SortOrder         int      `json:"sort_order"`
	MatchValue        *float64 `json:"match_value"`
	CompositionResult *string  `json:"composition_result"`
	MatchScore        *float64 `json:"match_score"`
}

type claimSerializer struct {
	ID                      uint                  `json:"id"`
	ClaimCode               string                `json:"claim_code"`
	Status                  string                `json:"status"`
	CreatedAt               time.Time             `json:"created_at"`
	FormedAt                *time.Time            `json:"formed_at"`
	CompletedAt             *time.Time            `json:"completed_at"`
	CreatorLogin            string                `json:"creator_login"`
	ModeratorLogin          *string               `json:"moderator_login"`
	ArtifactTitle           *string               `json:"artifact_title"`
	ArtifactOrigin          *string               `json:"artifact_origin"`
	AnalyzerModel           *string               `json:"analyzer_model"`
	OperatorComment         *string               `json:"operator_comment"`
	CuMeasured              *float64              `json:"cu_measured"`
	ZnMeasured              *float64              `json:"zn_measured"`
	SnMeasured              *float64              `json:"sn_measured"`
	PbMeasured              *float64              `json:"pb_measured"`
	BestMatchLabel          *string               `json:"best_match_label"`
	CompletionFormulaResult *float64              `json:"completion_formula_result"`
	TotalCost               *float64              `json:"total_cost"`
	PlannedDeliveryAt       *time.Time            `json:"planned_delivery_at"`
	ResultItemsCount        int64                 `json:"result_items_count"`
	Services                []claimItemSerializer `json:"services,omitempty"`
}

type claimListSerializer struct {
	ID                      uint       `json:"id"`
	ClaimCode               string     `json:"claim_code"`
	Status                  string     `json:"status"`
	CreatedAt               time.Time  `json:"created_at"`
	FormedAt                *time.Time `json:"formed_at"`
	CompletedAt             *time.Time `json:"completed_at"`
	CreatorLogin            string     `json:"creator_login"`
	ModeratorLogin          *string    `json:"moderator_login"`
	CompletionFormulaResult *float64   `json:"completion_formula_result"`
	TotalCost               *float64   `json:"total_cost"`
	PlannedDeliveryAt       *time.Time `json:"planned_delivery_at"`
	ResultItemsCount        int64      `json:"result_items_count"`
}

type userSerializer struct {
	ID        uint      `json:"id"`
	Login     string    `json:"login"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func toServiceSerializer(service model.ReferenceAlloyService) serviceSerializer {
	return serviceSerializer{
		ID:            service.ID,
		Slug:          service.Slug,
		Name:          service.Name,
		Description:   service.Description,
		Status:        service.Status,
		ImageFileName: service.ImageFileName,
		VideoFileName: service.VideoFileName,
		ImageURL:      service.ImageURL,
		VideoURL:      service.VideoURL,
		Era:           service.Era,
		Culture:       service.Culture,
		UnitPrice:     service.UnitPrice,
		CuReference:   service.CuReference,
		ZnReference:   service.ZnReference,
		SnReference:   service.SnReference,
		PbReference:   service.PbReference,
		CreatedAt:     service.CreatedAt,
		UpdatedAt:     service.UpdatedAt,
	}
}

func toCartSerializer(card *repository.CartIcon) cartSerializer {
	if card == nil {
		return cartSerializer{}
	}
	return cartSerializer{
		ClaimID:      card.ClaimID,
		ClaimCode:    card.ClaimCode,
		ServiceCount: card.ServiceCount,
	}
}

func toClaimItemSerializer(item repository.ClaimServiceItem) claimItemSerializer {
	return claimItemSerializer{
		ID:                item.ID,
		ServiceID:         item.ServiceID,
		ServiceSlug:       item.ServiceSlug,
		ServiceName:       item.ServiceName,
		ServiceImageURL:   item.ServiceImageURL,
		ServiceVideoURL:   item.ServiceVideoURL,
		Quantity:          item.Quantity,
		SortOrder:         item.SortOrder,
		MatchValue:        item.MatchValue,
		CompositionResult: item.CompositionResult,
		MatchScore:        item.MatchScore,
	}
}

func toClaimSerializer(details *repository.ClaimDetails) claimSerializer {
	services := make([]claimItemSerializer, 0, len(details.Services))
	for _, item := range details.Services {
		services = append(services, toClaimItemSerializer(item))
	}

	return claimSerializer{
		ID:                      details.Claim.ID,
		ClaimCode:               details.Claim.ClaimCode,
		Status:                  details.Claim.Status,
		CreatedAt:               details.Claim.CreatedAt,
		FormedAt:                details.Claim.FormedAt,
		CompletedAt:             details.Claim.CompletedAt,
		CreatorLogin:            details.CreatorLogin,
		ModeratorLogin:          details.ModeratorLogin,
		ArtifactTitle:           details.Claim.ArtifactTitle,
		ArtifactOrigin:          details.Claim.ArtifactOrigin,
		AnalyzerModel:           details.Claim.AnalyzerModel,
		OperatorComment:         details.Claim.OperatorComment,
		CuMeasured:              details.Claim.CuMeasured,
		ZnMeasured:              details.Claim.ZnMeasured,
		SnMeasured:              details.Claim.SnMeasured,
		PbMeasured:              details.Claim.PbMeasured,
		BestMatchLabel:          details.Claim.BestMatchLabel,
		CompletionFormulaResult: details.Claim.CompletionFormulaResult,
		TotalCost:               details.Claim.TotalCost,
		PlannedDeliveryAt:       details.Claim.PlannedDeliveryAt,
		ResultItemsCount:        details.ResultItemsCount,
		Services:                services,
	}
}

func toClaimListSerializer(row repository.ClaimListItem) claimListSerializer {
	return claimListSerializer{
		ID:                      row.ID,
		ClaimCode:               row.ClaimCode,
		Status:                  row.Status,
		CreatedAt:               row.CreatedAt,
		FormedAt:                row.FormedAt,
		CompletedAt:             row.CompletedAt,
		CreatorLogin:            row.CreatorLogin,
		ModeratorLogin:          row.ModeratorLogin,
		CompletionFormulaResult: row.CompletionFormulaResult,
		TotalCost:               row.TotalCost,
		PlannedDeliveryAt:       row.PlannedDeliveryAt,
		ResultItemsCount:        row.ResultItemsCount,
	}
}

func toUserSerializer(user *model.User) userSerializer {
	return userSerializer{
		ID:        user.ID,
		Login:     user.Login,
		FullName:  user.FullName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}
