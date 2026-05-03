package repository

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"

	"xrfApp/internal/app/model"
)

type ServiceFilters struct {
	Query string
}

type ServiceCreateInput struct {
	Name              string
	Description       string
	ClipDescriptionEN string
	Era               string
	Culture           string
	UnitPrice         float64
	CuReference       float64
	ZnReference       float64
	SnReference       float64
	PbReference       float64
	ImageFileName     *string
	VideoFileName     *string
	ImageURL          *string
	VideoURL          *string
}

func (r *Repository) ListServices(filters ServiceFilters) ([]model.ReferenceAlloyService, error) {
	query := r.db.Model(&model.ReferenceAlloyService{}).
		Where("status <> ?", model.ServiceStatusDeleted).
		Order("id ASC")

	if q := strings.ToLower(strings.TrimSpace(filters.Query)); q != "" {
		mask := "%" + q + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(era) LIKE ? OR LOWER(culture) LIKE ?",
			mask,
			mask,
			mask,
			mask,
		)
	}

	var services []model.ReferenceAlloyService
	if err := query.Find(&services).Error; err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}

	return services, nil
}

func (r *Repository) GetServiceByID(serviceID uint) (model.ReferenceAlloyService, error) {
	service := model.ReferenceAlloyService{}
	if err := r.db.
		Where("id = ? AND status <> ?", serviceID, model.ServiceStatusDeleted).
		First(&service).Error; err != nil {
		return model.ReferenceAlloyService{}, err
	}

	return service, nil
}

func (r *Repository) CreateService(input ServiceCreateInput) (*model.ReferenceAlloyService, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("%w: name is required", ErrValidation)
	}
	if strings.TrimSpace(input.Description) == "" {
		return nil, fmt.Errorf("%w: description is required", ErrValidation)
	}
	if input.UnitPrice <= 0 {
		return nil, fmt.Errorf("%w: unit_price must be positive", ErrValidation)
	}

	baseSlug := generateLatinSlug(input.Name)
	slug, err := r.uniqueServiceSlug(baseSlug)
	if err != nil {
		return nil, err
	}

	service := model.ReferenceAlloyService{
		Slug:              slug,
		Name:              strings.TrimSpace(input.Name),
		Description:       strings.TrimSpace(input.Description),
		ClipDescriptionEN: resolveClipDescriptionEN(slug, input.ClipDescriptionEN),
		Status:            model.ServiceStatusActive,
		ImageFileName:     input.ImageFileName,
		VideoFileName:     input.VideoFileName,
		ImageURL:          input.ImageURL,
		VideoURL:          input.VideoURL,
		Era:               strings.TrimSpace(input.Era),
		Culture:           strings.TrimSpace(input.Culture),
		UnitPrice:         input.UnitPrice,
		CuReference:       input.CuReference,
		ZnReference:       input.ZnReference,
		SnReference:       input.SnReference,
		PbReference:       input.PbReference,
	}

	if err := r.db.Create(&service).Error; err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	return &service, nil
}

func resolveClipDescriptionEN(slug string, raw string) string {
	normalized := normalizeClipDescriptionEN(raw)
	if normalized != "" {
		return normalized
	}
	return normalizeClipDescriptionEN(defaultClipDescriptionBySlug(slug))
}

func normalizeClipDescriptionEN(raw string) string {
	normalized := strings.TrimSpace(strings.Join(strings.Fields(raw), " "))
	if normalized == "" {
		return ""
	}

	if len(normalized) < 50 {
		normalized += " Prepared for archaeological alloy search in CLIP."
	}
	if len(normalized) > 100 {
		normalized = strings.TrimSpace(normalized[:100])
	}
	return normalized
}

func defaultClipDescriptionBySlug(slug string) string {
	cleanSlug := strings.ToLower(strings.TrimSpace(slug))
	switch {
	case strings.Contains(cleanSlug, "bronze"):
		return "Dark brown bronze ingot with oxidized rough texture and warm copper highlights under light."
	case strings.Contains(cleanSlug, "brass"):
		return "Warm yellow brass ingot with bright golden edges and medium metallic reflectance on surface."
	case strings.Contains(cleanSlug, "iron"):
		return "Dark gray iron ingot with matte grainy texture, cold tone, and weak silver highlights."
	case strings.Contains(cleanSlug, "silver"):
		return "Pale silver ingot with smooth texture, strong reflectance, and bright white specular highlights."
	default:
		return "Archaeological metal reference alloy sample for XRF spectral comparison and CLIP similarity search."
	}
}

func (r *Repository) uniqueServiceSlug(base string) (string, error) {
	slug := base
	if slug == "" {
		slug = "service"
	}

	for i := 1; i <= 2000; i++ {
		var cnt int64
		if err := r.db.Model(&model.ReferenceAlloyService{}).Where("slug = ?", slug).Count(&cnt).Error; err != nil {
			return "", fmt.Errorf("check service slug uniqueness: %w", err)
		}
		if cnt == 0 {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i+1)
	}

	return "", fmt.Errorf("%w: cannot generate unique slug", ErrValidation)
}

func (r *Repository) UploadServiceMedia(
	ctx context.Context,
	mediaKind string,
	originalFileName string,
	contentType string,
	size int64,
	reader io.Reader,
) (string, string, error) {
	if strings.TrimSpace(mediaKind) == "" {
		return "", "", fmt.Errorf("%w: media kind is required", ErrValidation)
	}
	if reader == nil {
		return "", "", fmt.Errorf("%w: reader is nil", ErrValidation)
	}
	if size <= 0 {
		return "", "", fmt.Errorf("%w: media size must be positive", ErrValidation)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	objectName := generateMediaObjectName(mediaKind, originalFileName)
	_, err := r.minio.PutObject(ctx, r.minioCfg.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", fmt.Errorf("upload media to minio: %w", err)
	}

	url := objectURL(r.minioCfg.publicURL, r.minioCfg.bucket, objectName)
	return objectName, url, nil
}

func (r *Repository) SoftDeleteService(serviceID uint) error {
	update := r.db.
		Model(&model.ReferenceAlloyService{}).
		Where("id = ? AND status <> ?", serviceID, model.ServiceStatusDeleted).
		Update("status", model.ServiceStatusDeleted)
	if update.Error != nil {
		return fmt.Errorf("delete service: %w", update.Error)
	}
	if update.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
