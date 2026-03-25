package handler

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"xrfApp/internal/app/auth"
	"xrfApp/internal/app/middleware"
	"xrfApp/internal/app/repository"
	"xrfApp/internal/app/session"
)

type Handler struct {
	Repository   *repository.Repository
	TokenManager *auth.Manager
	SessionStore *session.Manager
}

func NewHandler(r *repository.Repository, tokenManager *auth.Manager, sessionStore *session.Manager) *Handler {
	return &Handler{
		Repository:   r,
		TokenManager: tokenManager,
		SessionStore: sessionStore,
	}
}

func (h *Handler) GetServicesAPI(ctx *gin.Context) {
	services, err := h.Repository.ListServices(repository.ServiceFilters{
		Query: ctx.Query("q"),
	})
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	result := make([]serviceSerializer, 0, len(services))
	for _, service := range services {
		result = append(result, toServiceSerializer(service))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (h *Handler) GetServiceAPI(ctx *gin.Context) {
	serviceID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid service id"})
		return
	}

	service, err := h.Repository.GetServiceByID(serviceID)
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": toServiceSerializer(service),
	})
}

func (h *Handler) CreateServiceAPI(ctx *gin.Context) {
	if err := ctx.Request.ParseMultipartForm(128 << 20); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "multipart form is invalid"})
		return
	}

	unitPrice, err := parseRequiredFloat(ctx.PostForm("unit_price"), "unit_price")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	cuRef, err := parseRequiredFloat(ctx.PostForm("cu_reference"), "cu_reference")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	znRef, err := parseRequiredFloat(ctx.PostForm("zn_reference"), "zn_reference")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	snRef, err := parseRequiredFloat(ctx.PostForm("sn_reference"), "sn_reference")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	pbRef, err := parseRequiredFloat(ctx.PostForm("pb_reference"), "pb_reference")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var imageFileName *string
	var imageURL *string
	if header, fileErr := readOptionalFile(ctx, "image"); fileErr == nil && header != nil {
		name, url, uploadErr := h.uploadMultipartMedia(ctx, header, "image")
		if uploadErr != nil {
			h.handleRepoError(ctx, uploadErr)
			return
		}
		imageFileName = &name
		imageURL = &url
	} else if fileErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": fileErr.Error()})
		return
	}

	var videoFileName *string
	var videoURL *string
	if header, fileErr := readOptionalFile(ctx, "video"); fileErr == nil && header != nil {
		name, url, uploadErr := h.uploadMultipartMedia(ctx, header, "video")
		if uploadErr != nil {
			h.handleRepoError(ctx, uploadErr)
			return
		}
		videoFileName = &name
		videoURL = &url
	} else if fileErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": fileErr.Error()})
		return
	}

	service, err := h.Repository.CreateService(repository.ServiceCreateInput{
		Name:          strings.TrimSpace(ctx.PostForm("name")),
		Description:   strings.TrimSpace(ctx.PostForm("description")),
		Era:           strings.TrimSpace(ctx.PostForm("era")),
		Culture:       strings.TrimSpace(ctx.PostForm("culture")),
		UnitPrice:     unitPrice,
		CuReference:   cuRef,
		ZnReference:   znRef,
		SnReference:   snRef,
		PbReference:   pbRef,
		ImageFileName: imageFileName,
		VideoFileName: videoFileName,
		ImageURL:      imageURL,
		VideoURL:      videoURL,
	})
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "service created",
		"data":    toServiceSerializer(*service),
	})
}

func (h *Handler) AddServiceToDraftAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "authorization required"})
		return
	}

	body := struct {
		ServiceID uint `json:"service_id"`
	}{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON body"})
		return
	}
	if body.ServiceID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "service_id is required"})
		return
	}

	claim, match, err := h.Repository.AddServiceToDraft(user.ID, body.ServiceID)
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"claim_id":   claim.ID,
			"claim_code": claim.ClaimCode,
			"service_id": match.ServiceID,
			"quantity":   match.Quantity,
			"sort_order": match.SortOrder,
		},
	})
}

func (h *Handler) UpdateDraftMatchAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "authorization required"})
		return
	}

	serviceID, err := parseUintParam(ctx.Param("service_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid service id"})
		return
	}

	body := struct {
		Quantity  *int `json:"quantity"`
		SortOrder *int `json:"sort_order"`
	}{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON body"})
		return
	}

	match, err := h.Repository.UpdateDraftMatch(user.ID, serviceID, repository.MatchUpdateInput{
		Quantity:  body.Quantity,
		SortOrder: body.SortOrder,
	})
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"claim_id":   match.ClaimID,
			"service_id": match.ServiceID,
			"quantity":   match.Quantity,
			"sort_order": match.SortOrder,
		},
	})
}

func (h *Handler) DeleteDraftMatchAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "authorization required"})
		return
	}

	serviceID, err := parseUintParam(ctx.Param("service_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid service id"})
		return
	}

	if err := h.Repository.DeleteDraftMatch(user.ID, serviceID); err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "service removed from draft claim",
	})
}

func (h *Handler) GetCartIconAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "authorization required"})
		return
	}

	card, err := h.Repository.GetCartIcon(user.ID)
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": toCartSerializer(card),
	})
}

func (h *Handler) GetClaimsAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "authorization required"})
		return
	}

	filters, err := parseClaimFilters(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	filters.ViewerID = user.ID
	filters.ViewerRole = user.Role

	claims, err := h.Repository.ListClaims(filters)
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	result := make([]claimListSerializer, 0, len(claims))
	for _, item := range claims {
		result = append(result, toClaimListSerializer(item))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (h *Handler) GetClaimAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "authorization required"})
		return
	}

	claimID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid claim id"})
		return
	}

	details, err := h.Repository.GetClaimDetails(claimID, user.ID, user.Role)
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": toClaimSerializer(details),
	})
}

func (h *Handler) UpdateDraftClaimAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "authorization required"})
		return
	}

	claimID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid claim id"})
		return
	}

	body := struct {
		OperatorComment *string  `json:"operator_comment"`
		CuMeasured      *float64 `json:"cu_measured"`
		ZnMeasured      *float64 `json:"zn_measured"`
		SnMeasured      *float64 `json:"sn_measured"`
		PbMeasured      *float64 `json:"pb_measured"`
	}{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON body"})
		return
	}

	err = h.Repository.UpdateDraftClaimFields(user.ID, claimID, repository.ClaimUpdateInput{
		OperatorComment: body.OperatorComment,
		CuMeasured:      body.CuMeasured,
		ZnMeasured:      body.ZnMeasured,
		SnMeasured:      body.SnMeasured,
		PbMeasured:      body.PbMeasured,
	})
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "draft claim updated",
	})
}

func (h *Handler) FormClaimAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "authorization required"})
		return
	}

	claimID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid claim id"})
		return
	}

	claim, err := h.Repository.FormDraftClaim(user.ID, claimID)
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":                        claim.ID,
			"claim_code":                claim.ClaimCode,
			"status":                    claim.Status,
			"formed_at":                 claim.FormedAt,
			"completion_formula_result": claim.CompletionFormulaResult,
			"total_cost":                claim.TotalCost,
		},
	})
}

func (h *Handler) ModerateClaimAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "authorization required"})
		return
	}

	claimID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid claim id"})
		return
	}

	body := struct {
		Action string `json:"action"`
	}{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON body"})
		return
	}

	claim, err := h.Repository.ModerateFormedClaim(user.ID, claimID, body.Action)
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":           claim.ID,
			"claim_code":   claim.ClaimCode,
			"status":       claim.Status,
			"moderator_id": claim.ModeratorID,
			"completed_at": claim.CompletedAt,
		},
	})
}

func (h *Handler) DeleteDraftClaimAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "authorization required"})
		return
	}

	claimID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid claim id"})
		return
	}

	if err := h.Repository.DeleteDraftClaim(user.ID, claimID); err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "draft claim deleted",
	})
}

func (h *Handler) RegisterUserAPI(ctx *gin.Context) {
	body := struct {
		Login    string `json:"login"`
		FullName string `json:"full_name"`
		Password string `json:"password"`
	}{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON body"})
		return
	}

	user, err := h.Repository.RegisterUser(repository.RegisterUserInput{
		Login:    body.Login,
		FullName: body.FullName,
		Password: body.Password,
	})
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": toUserSerializer(user),
	})
}

func (h *Handler) AuthAPI(ctx *gin.Context) {
	if h.TokenManager == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "jwt manager is not configured"})
		return
	}
	if h.SessionStore == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "session store is not configured"})
		return
	}

	body := struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON body"})
		return
	}

	authResult, err := h.Repository.AuthenticateUser(repository.AuthInput{
		Login:    body.Login,
		Password: body.Password,
	})
	if err != nil {
		h.handleRepoError(ctx, err)
		return
	}

	sessionID, sessionExpiresAt, err := h.SessionStore.CreateSession(
		ctx.Request.Context(),
		authResult.UserID,
		authResult.Login,
		authResult.Role,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	token, claims, err := h.TokenManager.IssueToken(authResult.UserID, authResult.Login, authResult.Role, sessionID)
	if err != nil {
		_ = h.SessionStore.DeleteSession(ctx.Request.Context(), sessionID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user_id":            authResult.UserID,
			"login":              authResult.Login,
			"full_name":          authResult.FullName,
			"role":               authResult.Role,
			"token_type":         "Bearer",
			"token":              token,
			"expires_at":         time.Unix(claims.ExpiresAt, 0).UTC(),
			"session_id":         sessionID,
			"session_key":        h.SessionStore.Key(sessionID),
			"session_ttl":        int(h.SessionStore.SessionTTL().Seconds()),
			"session_expires_at": sessionExpiresAt,
			"auth_method":        "jwt",
		},
	})
}

func (h *Handler) LogoutStubAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if ok && h.SessionStore != nil {
		_ = h.SessionStore.DeleteSession(ctx.Request.Context(), user.SessionID)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "logout completed",
	})
}

func (h *Handler) handleRepoError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, repository.ErrValidation):
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case errors.Is(err, repository.ErrInvalidTransition):
		ctx.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
}

func (h *Handler) currentUser(ctx *gin.Context) (middleware.AuthUser, bool) {
	user, ok := middleware.CurrentUser(ctx)
	if !ok {
		return middleware.AuthUser{}, false
	}
	return user, true
}

func parseUintParam(raw string) (uint, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || value == 0 {
		return 0, fmt.Errorf("invalid uint")
	}
	return uint(value), nil
}

func parseRequiredFloat(raw, fieldName string) (float64, error) {
	value := strings.TrimSpace(strings.ReplaceAll(raw, ",", "."))
	if value == "" {
		return 0, fmt.Errorf("%s is required", fieldName)
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", fieldName)
	}
	return parsed, nil
}

func parseClaimFilters(ctx *gin.Context) (repository.ClaimFilters, error) {
	filters := repository.ClaimFilters{
		Status: strings.TrimSpace(ctx.Query("status")),
	}

	if rawFrom := strings.TrimSpace(ctx.Query("formed_from")); rawFrom != "" {
		from, err := time.Parse("2006-01-02", rawFrom)
		if err != nil {
			return repository.ClaimFilters{}, fmt.Errorf("formed_from must be YYYY-MM-DD")
		}
		filters.FormedFrom = &from
	}

	if rawTo := strings.TrimSpace(ctx.Query("formed_to")); rawTo != "" {
		to, err := time.Parse("2006-01-02", rawTo)
		if err != nil {
			return repository.ClaimFilters{}, fmt.Errorf("formed_to must be YYYY-MM-DD")
		}
		to = to.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filters.FormedTo = &to
	}

	return filters, nil
}

func readOptionalFile(ctx *gin.Context, field string) (*multipart.FileHeader, error) {
	header, err := ctx.FormFile(field)
	if err == nil {
		return header, nil
	}
	if errors.Is(err, http.ErrMissingFile) {
		return nil, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "no such file") {
		return nil, nil
	}
	return nil, fmt.Errorf("cannot read %s file", field)
}

func (h *Handler) uploadMultipartMedia(ctx *gin.Context, header *multipart.FileHeader, mediaKind string) (string, string, error) {
	file, err := header.Open()
	if err != nil {
		return "", "", fmt.Errorf("open %s file: %w", mediaKind, err)
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	return h.Repository.UploadServiceMedia(
		ctx.Request.Context(),
		mediaKind,
		header.Filename,
		contentType,
		header.Size,
		file,
	)
}
