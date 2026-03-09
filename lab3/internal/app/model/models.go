package model

import "time"

const (
	ServiceStatusActive  = "\u0434\u0435\u0439\u0441\u0442\u0432\u0443\u0435\u0442"
	ServiceStatusDeleted = "\u0443\u0434\u0430\u043b\u0435\u043d"
)

const (
	ClaimStatusDraft     = "\u0447\u0435\u0440\u043d\u043e\u0432\u0438\u043a"
	ClaimStatusDeleted   = "\u0443\u0434\u0430\u043b\u0435\u043d"
	ClaimStatusFormed    = "\u0441\u0444\u043e\u0440\u043c\u0438\u0440\u043e\u0432\u0430\u043d"
	ClaimStatusCompleted = "\u0437\u0430\u0432\u0435\u0440\u0448\u0435\u043d"
	ClaimStatusRejected  = "\u043e\u0442\u043a\u043b\u043e\u043d\u0435\u043d"
)

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Login        string    `gorm:"size:64;not null;uniqueIndex"`
	FullName     string    `gorm:"size:128;not null"`
	PasswordHash string    `gorm:"size:255;not null;default:''"`
	Role         string    `gorm:"size:32;not null"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (User) TableName() string {
	return "users"
}

type ReferenceAlloyService struct {
	ID            uint      `gorm:"primaryKey"`
	Slug          string    `gorm:"size:120;not null;uniqueIndex"`
	Name          string    `gorm:"size:160;not null"`
	Description   string    `gorm:"type:text;not null"`
	Status        string    `gorm:"size:16;not null;index"`
	ImageFileName *string   `gorm:"size:160"`
	VideoFileName *string   `gorm:"size:160"`
	ImageURL      *string   `gorm:"size:255"`
	VideoURL      *string   `gorm:"size:255"`
	Era           string    `gorm:"size:100;not null;default:''"`
	Culture       string    `gorm:"size:120;not null;default:''"`
	UnitPrice     float64   `gorm:"type:numeric(10,2);not null;default:0"`
	CuReference   float64   `gorm:"type:numeric(6,3);not null;default:0"`
	ZnReference   float64   `gorm:"type:numeric(6,3);not null;default:0"`
	SnReference   float64   `gorm:"type:numeric(6,3);not null;default:0"`
	PbReference   float64   `gorm:"type:numeric(6,3);not null;default:0"`
	CreatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (ReferenceAlloyService) TableName() string {
	return "reference_alloy_services"
}

type ArtifactClaim struct {
	ID                      uint      `gorm:"primaryKey"`
	ClaimCode               string    `gorm:"size:40;not null;uniqueIndex"`
	Status                  string    `gorm:"size:16;not null;index"`
	CreatedAt               time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	CreatorID               uint      `gorm:"not null;index"`
	Creator                 User      `gorm:"foreignKey:CreatorID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`
	FormedAt                *time.Time
	CompletedAt             *time.Time
	ModeratorID             *uint
	Moderator               *User    `gorm:"foreignKey:ModeratorID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`
	ArtifactTitle           *string  `gorm:"size:180"`
	ArtifactOrigin          *string  `gorm:"size:180"`
	AnalyzerModel           *string  `gorm:"size:120"`
	OperatorComment         *string  `gorm:"size:255"`
	CuMeasured              *float64 `gorm:"type:numeric(6,3)"`
	ZnMeasured              *float64 `gorm:"type:numeric(6,3)"`
	SnMeasured              *float64 `gorm:"type:numeric(6,3)"`
	PbMeasured              *float64 `gorm:"type:numeric(6,3)"`
	BestMatchLabel          *string  `gorm:"size:180"`
	CompletionFormulaResult *float64 `gorm:"type:numeric(8,2)"`
	TotalCost               *float64 `gorm:"type:numeric(12,2)"`
	PlannedDeliveryAt       *time.Time
}

func (ArtifactClaim) TableName() string {
	return "artifact_claims"
}

type ClaimAlloyMatch struct {
	ID                uint                  `gorm:"primaryKey"`
	ClaimID           uint                  `gorm:"not null;uniqueIndex:ux_claim_service,priority:1;index"`
	Claim             ArtifactClaim         `gorm:"foreignKey:ClaimID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`
	ServiceID         uint                  `gorm:"not null;uniqueIndex:ux_claim_service,priority:2;index"`
	Service           ReferenceAlloyService `gorm:"foreignKey:ServiceID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`
	Quantity          int                   `gorm:"not null;default:1"`
	SortOrder         int                   `gorm:"not null;default:0"`
	MatchValue        *float64              `gorm:"type:numeric(10,3)"`
	CompositionResult *string               `gorm:"size:255"`
	MatchScore        *float64              `gorm:"type:numeric(8,2)"`
	CreatedAt         time.Time             `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time             `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (ClaimAlloyMatch) TableName() string {
	return "claim_alloy_matches"
}
