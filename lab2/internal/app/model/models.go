package model

import "time"

const (
	ServiceStatusActive  = "действует"
	ServiceStatusDeleted = "удален"
)

const (
	ClaimStatusDraft     = "черновик"
	ClaimStatusDeleted   = "удален"
	ClaimStatusFormed    = "сформирован"
	ClaimStatusCompleted = "завершен"
	ClaimStatusRejected  = "отклонен"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Login     string    `gorm:"size:64;not null;uniqueIndex"`
	FullName  string    `gorm:"size:128;not null"`
	Role      string    `gorm:"size:32;not null"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (User) TableName() string {
	return "users"
}

type ReferenceAlloyService struct {
	ID          uint      `gorm:"primaryKey"`
	Slug        string    `gorm:"size:120;not null;uniqueIndex"`
	Name        string    `gorm:"size:160;not null"`
	Description string    `gorm:"type:text;not null"`
	Status      string    `gorm:"size:16;not null;index"`
	ImageURL    *string   `gorm:"size:255"`
	VideoURL    *string   `gorm:"size:255"`
	Era         string    `gorm:"size:100;not null"`
	Culture     string    `gorm:"size:120;not null"`
	CuReference float64   `gorm:"type:numeric(6,3);not null"`
	ZnReference float64   `gorm:"type:numeric(6,3);not null"`
	SnReference float64   `gorm:"type:numeric(6,3);not null"`
	PbReference float64   `gorm:"type:numeric(6,3);not null"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
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
	SortOrder         int                   `gorm:"not null;default:1"`
	IsPrimary         bool                  `gorm:"not null;default:false"`
	CompositionResult *string               `gorm:"size:200"`
	MatchScore        *float64              `gorm:"type:numeric(8,2)"`
}

func (ClaimAlloyMatch) TableName() string {
	return "claim_alloy_matches"
}
