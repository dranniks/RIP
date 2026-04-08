package repository

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"xrfApp/internal/app/identity"
	"xrfApp/internal/app/model"
)

const (
	defaultDBHost = "localhost"
	defaultDBPort = "5433"
	defaultDBName = "RIP"
	defaultDBUser = "root"
	defaultDBPass = "root"

	defaultMinIOEndpoint  = "localhost:9000"
	defaultMinIOAccessKey = "root"
	defaultMinIOSecretKey = "rootroot"
	defaultMinIOBucket    = "xrf-media"
	defaultMinIOUseSSL    = "false"
	defaultMinIOPublicURL = "http://localhost:9000"
)

var (
	ErrValidation        = errors.New("validation error")
	ErrInvalidTransition = errors.New("invalid claim transition")
)

type Repository struct {
	db       *gorm.DB
	minio    *minio.Client
	minioCfg minioConfig
}

type dbConnConfig struct {
	host string
	port string
	name string
	user string
	pass string
}

type minioConfig struct {
	endpoint  string
	accessKey string
	secretKey string
	bucket    string
	publicURL string
	useSSL    bool
}

func NewRepository() (*Repository, error) {
	loadDotEnvFile(".env")

	dbCfg := dbConnConfig{
		host: envOrDefault("DB_HOST", defaultDBHost),
		port: envOrDefault("DB_PORT", defaultDBPort),
		name: envOrDefault("DB_NAME", defaultDBName),
		user: envOrDefault("DB_USER", defaultDBUser),
		pass: envOrDefault("DB_PASS", defaultDBPass),
	}

	db, err := gorm.Open(postgres.Open(buildDSN(dbCfg)), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open db failed: %w", err)
	}

	repo := &Repository{
		db: db,
		minioCfg: minioConfig{
			endpoint:  envOrDefault("MINIO_ENDPOINT", defaultMinIOEndpoint),
			accessKey: envOrDefault("MINIO_ROOT_USER", defaultMinIOAccessKey),
			secretKey: envOrDefault("MINIO_ROOT_PASSWORD", defaultMinIOSecretKey),
			bucket:    envOrDefault("MINIO_BUCKET", defaultMinIOBucket),
			publicURL: strings.TrimRight(envOrDefault("MINIO_PUBLIC_URL", defaultMinIOPublicURL), "/"),
			useSSL:    strings.EqualFold(envOrDefault("MINIO_USE_SSL", defaultMinIOUseSSL), "true"),
		},
	}

	if err := repo.initMinIO(); err != nil {
		return nil, err
	}
	if err := repo.migrate(); err != nil {
		return nil, err
	}
	if err := repo.seed(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) initMinIO() error {
	client, err := minio.New(r.minioCfg.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(r.minioCfg.accessKey, r.minioCfg.secretKey, ""),
		Secure: r.minioCfg.useSSL,
	})
	if err != nil {
		return fmt.Errorf("create minio client: %w", err)
	}

	r.minio = client

	exists, err := client.BucketExists(ctxTimeout(), r.minioCfg.bucket)
	if err != nil {
		return fmt.Errorf("check minio bucket: %w", err)
	}
	if exists {
		return nil
	}

	if err = client.MakeBucket(ctxTimeout(), r.minioCfg.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create minio bucket %s: %w", r.minioCfg.bucket, err)
	}

	return nil
}

func (r *Repository) migrate() error {
	if err := r.db.AutoMigrate(
		&model.User{},
		&model.ReferenceAlloyService{},
		&model.ArtifactClaim{},
		&model.ClaimAlloyMatch{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	stmts := []string{
		"ALTER TABLE reference_alloy_services ADD COLUMN IF NOT EXISTS image_file_name VARCHAR(160)",
		"ALTER TABLE reference_alloy_services ADD COLUMN IF NOT EXISTS video_file_name VARCHAR(160)",
		"ALTER TABLE reference_alloy_services ADD COLUMN IF NOT EXISTS unit_price NUMERIC(10,2) NOT NULL DEFAULT 0",
		"ALTER TABLE artifact_claims ADD COLUMN IF NOT EXISTS total_cost NUMERIC(12,2)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255) NOT NULL DEFAULT ''",
		"ALTER TABLE claim_alloy_matches ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE artifact_claims DROP COLUMN IF EXISTS artifact_title",
		"ALTER TABLE artifact_claims DROP COLUMN IF EXISTS artifact_origin",
		"ALTER TABLE artifact_claims DROP COLUMN IF EXISTS analyzer_model",
		"ALTER TABLE artifact_claims DROP COLUMN IF EXISTS planned_delivery_at",
		"ALTER TABLE claim_alloy_matches ADD COLUMN IF NOT EXISTS result_value NUMERIC(10,3)",
		"DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'claim_alloy_matches' AND column_name = 'match_score') THEN EXECUTE 'UPDATE claim_alloy_matches SET result_value = COALESCE(result_value, match_score)'; END IF; END $$",
		"DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'claim_alloy_matches' AND column_name = 'match_value') THEN EXECUTE 'UPDATE claim_alloy_matches SET result_value = COALESCE(result_value, match_value)'; END IF; END $$",
		"ALTER TABLE claim_alloy_matches DROP COLUMN IF EXISTS match_value",
		"ALTER TABLE claim_alloy_matches DROP COLUMN IF EXISTS composition_result",
		"ALTER TABLE claim_alloy_matches DROP COLUMN IF EXISTS match_score",
		"ALTER TABLE claim_alloy_matches DROP COLUMN IF EXISTS created_at",
		"ALTER TABLE claim_alloy_matches DROP COLUMN IF EXISTS updated_at",
		"ALTER TABLE claim_alloy_matches DROP CONSTRAINT IF EXISTS claim_alloy_matches_pkey",
		"ALTER TABLE claim_alloy_matches DROP CONSTRAINT IF EXISTS ux_claim_service",
		"DROP INDEX IF EXISTS ux_claim_service",
		"ALTER TABLE claim_alloy_matches ADD CONSTRAINT claim_alloy_matches_pkey PRIMARY KEY (claim_id, service_id)",
		"ALTER TABLE claim_alloy_matches DROP CONSTRAINT IF EXISTS ck_claim_alloy_matches_quantity",
		"ALTER TABLE claim_alloy_matches DROP CONSTRAINT IF EXISTS ck_claim_alloy_matches_sort_order",
		"ALTER TABLE claim_alloy_matches ADD CONSTRAINT ck_claim_alloy_matches_quantity CHECK (quantity > 0)",
		"ALTER TABLE claim_alloy_matches ADD CONSTRAINT ck_claim_alloy_matches_sort_order CHECK (sort_order >= 0)",
		"ALTER TABLE claim_alloy_matches DROP COLUMN IF EXISTS id",
		"DROP TRIGGER IF EXISTS trg_recalc_completion_result ON artifact_claims",
		"DROP FUNCTION IF EXISTS recalc_completion_result()",
		"UPDATE reference_alloy_services SET status = '" + model.ServiceStatusActive + "' WHERE status NOT IN ('" + model.ServiceStatusActive + "', '" + model.ServiceStatusDeleted + "')",
		"UPDATE artifact_claims SET status = '" + model.ClaimStatusDraft + "' WHERE status NOT IN ('" + model.ClaimStatusDraft + "', '" + model.ClaimStatusDeleted + "', '" + model.ClaimStatusFormed + "', '" + model.ClaimStatusCompleted + "', '" + model.ClaimStatusRejected + "')",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_claim_draft_per_creator ON artifact_claims (creator_id) WHERE status = '" + model.ClaimStatusDraft + "'",
	}

	for _, stmt := range stmts {
		if err := r.db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("migration statement failed: %w", err)
		}
	}

	if err := r.db.Exec("ALTER TABLE reference_alloy_services DROP CONSTRAINT IF EXISTS ck_reference_alloy_services_status").Error; err != nil {
		return fmt.Errorf("drop service status check: %w", err)
	}
	if err := r.db.Exec("ALTER TABLE artifact_claims DROP CONSTRAINT IF EXISTS ck_artifact_claims_status").Error; err != nil {
		return fmt.Errorf("drop claim status check: %w", err)
	}
	serviceStatusCheckSQL := fmt.Sprintf(
		"ALTER TABLE reference_alloy_services ADD CONSTRAINT ck_reference_alloy_services_status CHECK (status IN ('%s', '%s'))",
		model.ServiceStatusActive,
		model.ServiceStatusDeleted,
	)
	if err := r.db.Exec(serviceStatusCheckSQL).Error; err != nil {
		return fmt.Errorf("add service status check: %w", err)
	}
	claimStatusCheckSQL := fmt.Sprintf(
		"ALTER TABLE artifact_claims ADD CONSTRAINT ck_artifact_claims_status CHECK (status IN ('%s', '%s', '%s', '%s', '%s'))",
		model.ClaimStatusDraft,
		model.ClaimStatusDeleted,
		model.ClaimStatusFormed,
		model.ClaimStatusCompleted,
		model.ClaimStatusRejected,
	)
	if err := r.db.Exec(claimStatusCheckSQL).Error; err != nil {
		return fmt.Errorf("add claim status check: %w", err)
	}

	return nil
}

func (r *Repository) seed() error {
	users := identity.CurrentUsers()

	creator := model.User{
		ID:           users.Creator.ID,
		Login:        users.Creator.Login,
		FullName:     "Claim Creator",
		PasswordHash: hashPassword("creator"),
		Role:         users.Creator.Role,
	}
	moderator := model.User{
		ID:           users.Moderator.ID,
		Login:        users.Moderator.Login,
		FullName:     "Claim Moderator",
		PasswordHash: hashPassword("moderator"),
		Role:         users.Moderator.Role,
	}

	for _, u := range []model.User{creator, moderator} {
		if err := upsertUser(r.db, u); err != nil {
			return err
		}
	}

	services := []model.ReferenceAlloyService{
		{
			Slug:        "alloy-bronze-cyprus",
			Name:        "Bronze Cyprus",
			Description: "Reference bronze sample for XRF.",
			Status:      model.ServiceStatusActive,
			Era:         "Late Bronze Age",
			Culture:     "Eastern Mediterranean",
			UnitPrice:   1250,
			CuReference: 0.830,
			ZnReference: 0.040,
			SnReference: 0.310,
			PbReference: 0.120,
		},
		{
			Slug:        "alloy-brass-rome",
			Name:        "Brass Rome",
			Description: "Roman brass reference sample with high Zn.",
			Status:      model.ServiceStatusActive,
			Era:         "I-III centuries",
			Culture:     "Roman Empire",
			UnitPrice:   1420,
			CuReference: 0.780,
			ZnReference: 0.640,
			SnReference: 0.060,
			PbReference: 0.030,
		},
		{
			Slug:        "alloy-silver-byzantium",
			Name:        "Silver Byzantium",
			Description: "Byzantine silver reference sample.",
			Status:      model.ServiceStatusActive,
			Era:         "X-XII centuries",
			Culture:     "Byzantine workshops",
			UnitPrice:   2100,
			CuReference: 0.290,
			ZnReference: 0.010,
			SnReference: 0.000,
			PbReference: 0.090,
		},
	}

	for _, svc := range services {
		if err := upsertService(r.db, svc); err != nil {
			return err
		}
	}

	if err := resetSequence(r.db, "users", "id"); err != nil {
		return err
	}
	if err := resetSequence(r.db, "reference_alloy_services", "id"); err != nil {
		return err
	}
	if err := resetSequence(r.db, "artifact_claims", "id"); err != nil {
		return err
	}
	return nil
}

func upsertUser(db *gorm.DB, user model.User) error {
	existing := model.User{}
	err := db.Where("id = ?", user.ID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if createErr := db.Create(&user).Error; createErr != nil {
			return fmt.Errorf("seed create user: %w", createErr)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("seed find user: %w", err)
	}

	if updateErr := db.Model(&existing).Updates(map[string]any{
		"login":         user.Login,
		"full_name":     user.FullName,
		"password_hash": user.PasswordHash,
		"role":          user.Role,
	}).Error; updateErr != nil {
		return fmt.Errorf("seed update user: %w", updateErr)
	}

	return nil
}

func upsertService(db *gorm.DB, service model.ReferenceAlloyService) error {
	existing := model.ReferenceAlloyService{}
	err := db.Where("slug = ?", service.Slug).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if createErr := db.Create(&service).Error; createErr != nil {
			return fmt.Errorf("seed create service: %w", createErr)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("seed find service: %w", err)
	}

	if updateErr := db.Model(&existing).Updates(map[string]any{
		"name":         service.Name,
		"description":  service.Description,
		"status":       service.Status,
		"era":          service.Era,
		"culture":      service.Culture,
		"unit_price":   service.UnitPrice,
		"cu_reference": service.CuReference,
		"zn_reference": service.ZnReference,
		"sn_reference": service.SnReference,
		"pb_reference": service.PbReference,
	}).Error; updateErr != nil {
		return fmt.Errorf("seed update service: %w", updateErr)
	}

	return nil
}

func buildDSN(cfg dbConnConfig) string {
	parts := []string{
		fmt.Sprintf("host=%s", cfg.host),
		fmt.Sprintf("port=%s", cfg.port),
		fmt.Sprintf("user=%s", cfg.user),
	}
	if cfg.pass != "" {
		parts = append(parts, fmt.Sprintf("password=%s", cfg.pass))
	}
	parts = append(parts, fmt.Sprintf("dbname=%s", cfg.name), "sslmode=disable")
	return strings.Join(parts, " ")
}

func loadDotEnvFile(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		_ = os.Setenv(key, value)
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func objectURL(baseURL, bucket, objectName string) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(baseURL, "/"), bucket, objectName)
}

var nonLatinFilenameChars = regexp.MustCompile(`[^a-z0-9]+`)

func generateLatinSlug(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = nonLatinFilenameChars.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	if normalized == "" {
		return "service"
	}
	return normalized
}

func generateClaimCode(claimID uint) string {
	return fmt.Sprintf("CLM-%06d", claimID)
}

func generateMediaObjectName(kind, originalFileName string) string {
	ext := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(filepathExt(originalFileName), ".")))
	if ext == "" {
		ext = "bin"
	}
	stamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	base := fmt.Sprintf("%s-%s", kind, stamp)
	base = nonLatinFilenameChars.ReplaceAllString(strings.ToLower(base), "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = kind
	}
	return base + "." + ext
}

func filepathExt(fileName string) string {
	idx := strings.LastIndex(fileName, ".")
	if idx < 0 {
		return ""
	}
	return fileName[idx:]
}

func ptrTime(t time.Time) *time.Time {
	v := t
	return &v
}

func resetSequence(db *gorm.DB, tableName, columnName string) error {
	query := fmt.Sprintf(
		"SELECT setval(pg_get_serial_sequence('%s', '%s'), COALESCE((SELECT MAX(%s) FROM %s), 1), true)",
		tableName,
		columnName,
		columnName,
		tableName,
	)
	if err := db.Exec(query).Error; err != nil {
		return fmt.Errorf("reset sequence %s.%s: %w", tableName, columnName, err)
	}
	return nil
}
