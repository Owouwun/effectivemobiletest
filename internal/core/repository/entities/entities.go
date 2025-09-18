package entities

import (
	"time"

	"github.com/Owouwun/effectivemobiletest/internal/core/logic/services"
	"github.com/google/uuid"
)

type ServiceEntity struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ServiceName string    `gorm:"not null"`
	Price       int       `gorm:"not null"`
	UserID      uuid.UUID `gorm:"not null;type:uuid"`
	StartDate   time.Time `gorm:"not null"`
	EndDate     *time.Time
}

func (ServiceEntity) TableName() string {
	return "services"
}

func NewServiceEntityFromLogic(s *services.Service) *ServiceEntity {
	return &ServiceEntity{
		ID:          *s.ID,
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID,
		StartDate:   s.StartDate,
		EndDate:     s.EndDate,
	}
}

func (se *ServiceEntity) ToLogicService() *services.Service {
	return &services.Service{
		ID:          &se.ID,
		ServiceName: se.ServiceName,
		Price:       se.Price,
		UserID:      se.UserID,
		StartDate:   se.StartDate,
		EndDate:     se.EndDate,
	}
}
