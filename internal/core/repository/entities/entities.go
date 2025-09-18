package entities

import (
	"time"

	"github.com/Owouwun/effectivemobiletest/internal/core/logic/services"
	"github.com/google/uuid"
)

type ServiceEntity struct {
	ServiceName string    `gorm:"primaryKey"`
	Price       int       `gorm:"not null"`
	UserID      uuid.UUID `gorm:"not null;type:uuid"`
	StartDate   time.Time `gorm:"not null"`
	EndDate     *time.Time
}

func (ServiceEntity) TableName() string {
	return "services"
}

func NewServiceEntityFromLogic(se *services.Service) *ServiceEntity {
	return &ServiceEntity{
		ServiceName: se.ServiceName,
		Price:       se.Price,
		UserID:      se.UserID,
		StartDate:   se.StartDate,
		EndDate:     se.EndDate,
	}
}

func (se *ServiceEntity) ToLogicService() *services.Service {
	return &services.Service{
		ServiceName: se.ServiceName,
		Price:       se.Price,
		UserID:      se.UserID,
		StartDate:   se.StartDate,
		EndDate:     se.EndDate,
	}
}
