package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	ID          uuid.UUID  `json:"id,omitempty" example:"123e4567-e89b-12d3-a456-426614174009"`
	ServiceName string     `json:"service_name" example:"My Service"`
	Price       int        `json:"price" example:"500"`
	UserID      uuid.UUID  `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	StartDate   time.Time  `json:"start_date" example:"01-2024"`
	EndDate     *time.Time `json:"end_date,omitempty" example:"12-2025"`
}

type Filters struct {
	SrvNames  []*string    `json:"service_name,omitempty" example:"[\"My Service\", \"Someone's Service\"]"`
	UserIDs   []*uuid.UUID `json:"user_id,omitempty" example:"[\"123e4567-e89b-12d3-a456-426614174000\"]"`
	StartDate time.Time    `json:"start_date" example:"01-2024"`
	EndDate   time.Time    `json:"end_date" example:"12-2025"`
}

var (
	ErrNotFound = errors.New("not found")
)
