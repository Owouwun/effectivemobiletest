package repository_services

import (
	"context"

	"github.com/Owouwun/effectivemobiletest/internal/core/logic/services"
	"github.com/Owouwun/effectivemobiletest/internal/core/repository/entities"
	"gorm.io/gorm"
)

type GormServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) services.SubscriptionRepository {
	return &GormServiceRepository{db: db}
}

func (r *GormServiceRepository) CreateService(ctx context.Context, srv *services.Service) error {
	serviceEntity := entities.NewServiceEntityFromLogic(srv)

	result := r.db.WithContext(ctx).Create(&serviceEntity)

	return result.Error
}

func (r *GormServiceRepository) GetService(ctx context.Context, serviceName string) (*services.Service, error) {
	var serviceEntity *entities.ServiceEntity
	result := r.db.WithContext(ctx).
		First(&serviceEntity, "service_name = ?", serviceName)

	if result.Error != nil {
		return nil, services.ErrNotFound
	}

	return serviceEntity.ToLogicService(), nil
}

func (r *GormServiceRepository) GetServices(ctx context.Context) ([]*services.Service, error) {
	var serviceEntities []*entities.ServiceEntity
	result := r.db.WithContext(ctx).
		Find(&serviceEntities)
	if result.Error != nil {
		return nil, result.Error
	}

	var logicServices []*services.Service
	for _, entity := range serviceEntities {
		logicServices = append(logicServices, entity.ToLogicService())
	}

	return logicServices, nil
}

func (r *GormServiceRepository) UpdateService(ctx context.Context, srv *services.Service) error {
	serviceEntity := entities.NewServiceEntityFromLogic(srv)

	result := r.db.WithContext(ctx).
		Model(&serviceEntity).
		Where("service_name = ?", srv.ServiceName).
		Updates(serviceEntity)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return services.ErrNotFound
	}

	return nil
}

func (r *GormServiceRepository) DeleteService(ctx context.Context, serviceName string) error {
	result := r.db.WithContext(ctx).
		Delete(&entities.ServiceEntity{}).
		Where("service_name = ?", serviceName)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return services.ErrNotFound
	}

	return nil
}

func (r *GormServiceRepository) FilterServices(ctx context.Context, filters *services.Filters) ([]*services.Service, error) {
	var filteredEntities []entities.ServiceEntity
	result := r.db.WithContext(ctx).
		Find(&filteredEntities).
		Where("service_name IN ?", filters.SrvNames).
		Where("user_id IN ?", filters.UserIDs)

	if result.Error != nil {
		return nil, result.Error
	}

	var logicServices []*services.Service
	for _, entity := range filteredEntities {
		logicServices = append(logicServices, entity.ToLogicService())
	}

	return logicServices, nil
}
