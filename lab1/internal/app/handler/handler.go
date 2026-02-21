package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"xrfApp/internal/app/repository"
)

const currentClaimID = "artifact-001"

type ClaimRow struct {
	Service           repository.Service
	MMField           string
	CalculationResult string
}

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
	services := h.Repository.SearchServicesByName(searchQuery)

	claim, err := h.Repository.GetClaimByID(currentClaimID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "cannot load claim: %v", err)
		return
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"Services":          services,
		"Claim":             claim,
		"ClaimServiceCount": len(claim.Lines),
		"Query":             searchQuery,
	})
}

func (h *Handler) GetService(ctx *gin.Context) {
	serviceID := ctx.Param("id")
	service, err := h.Repository.GetServiceByID(serviceID)
	if err != nil {
		ctx.String(http.StatusNotFound, "service not found")
		return
	}

	ctx.HTML(http.StatusOK, "service.html", gin.H{
		"Service": service,
	})
}

func (h *Handler) GetClaim(ctx *gin.Context) {
	claimID := ctx.Param("id")
	claim, err := h.Repository.GetClaimByID(claimID)
	if err != nil {
		ctx.String(http.StatusNotFound, "claim not found")
		return
	}

	rows := make([]ClaimRow, 0, len(claim.Lines))
	for _, line := range claim.Lines {
		service, serviceErr := h.Repository.GetServiceByID(line.ServiceID)
		if serviceErr != nil {
			ctx.String(http.StatusInternalServerError, "claim has broken service reference: %v", serviceErr)
			return
		}

		rows = append(rows, ClaimRow{
			Service:           service,
			MMField:           line.MMField,
			CalculationResult: line.CalculationResult,
		})
	}

	ctx.HTML(http.StatusOK, "claim.html", gin.H{
		"Claim":             claim,
		"Rows":              rows,
		"ClaimServiceCount": len(rows),
	})
}
