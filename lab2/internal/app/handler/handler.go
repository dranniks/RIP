package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"xrfApp/internal/app/model"
	"xrfApp/internal/app/repository"
)

const currentUserID uint = 1

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetServices(ctx *gin.Context) {
	searchQuery := strings.TrimSpace(ctx.Query("q"))

	services, err := h.Repository.SearchServices(searchQuery)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "cannot load services: %v", err)
		return
	}

	draft, err := h.Repository.GetDraftCard(currentUserID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "cannot load draft card: %v", err)
		return
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"Services":         services,
		"Draft":            draft,
		"Query":            searchQuery,
		"Added":            ctx.Query("added") == "1",
		"Deleted":          ctx.Query("deleted") == "1",
		"ClaimUnavailable": ctx.Query("claim_unavailable") == "1",
	})
}

func (h *Handler) GetService(ctx *gin.Context) {
	serviceSlug := strings.TrimSpace(ctx.Param("slug"))
	service, err := h.Repository.GetServiceBySlug(serviceSlug)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.String(http.StatusNotFound, "service not found")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "cannot load service: %v", err)
		return
	}

	draft, err := h.Repository.GetDraftCard(currentUserID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "cannot load draft card: %v", err)
		return
	}

	ctx.HTML(http.StatusOK, "service.html", gin.H{
		"Service": service,
		"Draft":   draft,
	})
}

func (h *Handler) GetClaim(ctx *gin.Context) {
	claimCode := firstNonEmpty(
		ctx.Param("code"),
		ctx.Query("claim_code"),
		ctx.Query("code"),
	)
	if claimCode == "" {
		redirectClaimUnavailable(ctx)
		return
	}

	action := strings.TrimSpace(ctx.Query("action"))
	if action != "" {
		updates, parseErr := parseClaimUpdatePayload(ctx)
		if parseErr != nil {
			ctx.String(http.StatusBadRequest, "invalid claim input: %v", parseErr)
			return
		}

		if updateErr := h.Repository.UpdateClaimInputData(currentUserID, claimCode, updates); updateErr != nil {
			if errors.Is(updateErr, gorm.ErrRecordNotFound) {
				redirectClaimUnavailable(ctx)
				return
			}
			ctx.String(http.StatusInternalServerError, "cannot update claim: %v", updateErr)
			return
		}

		if action == "submit" {
			if submitErr := h.Repository.SubmitDraftClaimORM(currentUserID, claimCode); submitErr != nil {
				if errors.Is(submitErr, gorm.ErrRecordNotFound) {
					redirectClaimUnavailable(ctx)
					return
				}
				ctx.String(http.StatusInternalServerError, "cannot submit claim: %v", submitErr)
				return
			}

			ctx.Redirect(http.StatusFound, "/artifact_claims/"+claimCode+"?submitted=1")
			return
		}

		ctx.Redirect(http.StatusFound, "/artifact_claims/"+claimCode+"?calculated=1")
		return
	}

	details, err := h.Repository.GetClaimByCode(currentUserID, claimCode)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		redirectClaimUnavailable(ctx)
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "cannot load claim: %v", err)
		return
	}

	ctx.HTML(http.StatusOK, "claim.html", gin.H{
		"Claim":         details.Claim,
		"ClaimPathCode": strconv.FormatUint(uint64(details.Claim.ID), 10),
		"Rows":          details.Rows,
		"TotalServices": details.TotalServices,
		"Formula":       details.Formula,
		"CanDelete":     details.Claim.Status == model.ClaimStatusDraft,
		"Calculated":    ctx.Query("calculated") == "1",
		"Submitted":     ctx.Query("submitted") == "1",
		"Input":         mapClaimInputValues(details.Claim),
	})
}

func (h *Handler) AddServiceToDraft(ctx *gin.Context) {
	serviceSlug := strings.TrimSpace(ctx.PostForm("service_slug"))
	if serviceSlug == "" {
		ctx.String(http.StatusBadRequest, "service slug is required")
		return
	}

	_, err := h.Repository.AddServiceToDraft(currentUserID, serviceSlug)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.String(http.StatusNotFound, "service not found")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "cannot add service to draft: %v", err)
		return
	}

	redirectURL := strings.TrimSpace(ctx.PostForm("redirect_to"))
	if redirectURL == "" {
		redirectURL = "/services?added=1"
	}
	ctx.Redirect(http.StatusFound, redirectURL)
}

func (h *Handler) DeleteDraftClaim(ctx *gin.Context) {
	claimCode := firstNonEmpty(
		ctx.Param("code"),
		ctx.PostForm("claim_code"),
		ctx.PostForm("code"),
	)
	if claimCode == "" {
		ctx.String(http.StatusBadRequest, "claim code is required")
		return
	}

	err := h.Repository.SoftDeleteDraftClaimSQL(currentUserID, claimCode)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.String(http.StatusNotFound, "draft claim not found")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "cannot delete claim: %v", err)
		return
	}

	ctx.Redirect(http.StatusFound, "/services?deleted=1")
}

func parseClaimUpdatePayload(ctx *gin.Context) (map[string]any, error) {
	updates := make(map[string]any)

	setNullableStringQuery(ctx, "artifact", "artifact_title", updates)
	setNullableStringQuery(ctx, "origin", "artifact_origin", updates)
	setNullableStringQuery(ctx, "analyzer", "analyzer_model", updates)
	setNullableStringQuery(ctx, "mm", "operator_comment", updates)

	if raw, ok := ctx.GetQuery("icu"); ok {
		value, err := parseNullableFloat(raw)
		if err != nil {
			return nil, fmt.Errorf("icu: %w", err)
		}
		updates["cu_measured"] = value
	}
	if raw, ok := ctx.GetQuery("izn"); ok {
		value, err := parseNullableFloat(raw)
		if err != nil {
			return nil, fmt.Errorf("izn: %w", err)
		}
		updates["zn_measured"] = value
	}
	if raw, ok := ctx.GetQuery("isn"); ok {
		value, err := parseNullableFloat(raw)
		if err != nil {
			return nil, fmt.Errorf("isn: %w", err)
		}
		updates["sn_measured"] = value
	}
	if raw, ok := ctx.GetQuery("ipb"); ok {
		value, err := parseNullableFloat(raw)
		if err != nil {
			return nil, fmt.Errorf("ipb: %w", err)
		}
		updates["pb_measured"] = value
	}

	return updates, nil
}

func setNullableStringQuery(ctx *gin.Context, queryKey, dbKey string, updates map[string]any) {
	raw, ok := ctx.GetQuery(queryKey)
	if !ok {
		return
	}

	value := strings.TrimSpace(raw)
	if value == "" {
		updates[dbKey] = nil
		return
	}
	updates[dbKey] = value
}

func parseNullableFloat(raw string) (any, error) {
	value := strings.TrimSpace(strings.ReplaceAll(raw, ",", "."))
	if value == "" {
		return nil, nil
	}

	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}

	return number, nil
}

func mapClaimInputValues(claim model.ArtifactClaim) map[string]string {
	return map[string]string{
		"Artifact": stringOrEmpty(claim.ArtifactTitle),
		"Origin":   stringOrEmpty(claim.ArtifactOrigin),
		"Analyzer": stringOrEmpty(claim.AnalyzerModel),
		"MM":       stringOrEmpty(claim.OperatorComment),
		"ICu":      formatFloat(claim.CuMeasured),
		"IZn":      formatFloat(claim.ZnMeasured),
		"ISn":      formatFloat(claim.SnMeasured),
		"IPb":      formatFloat(claim.PbMeasured),
	}
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func formatFloat(value *float64) string {
	if value == nil {
		return ""
	}

	formatted := strconv.FormatFloat(*value, 'f', 3, 64)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	return formatted
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func redirectClaimUnavailable(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, "/services?claim_unavailable=1")
}
