package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	CreateService(ctx context.Context, srv *Service) error
	GetService(ctx context.Context, ID uuid.UUID) (*Service, error)
	GetServices(ctx context.Context) ([]*Service, error)
	UpdateService(ctx context.Context, srv *Service) error
	DeleteService(ctx context.Context, ID uuid.UUID) error
	FilterServices(ctx context.Context, filters *Filters) ([]*Service, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) CreateService(ctx context.Context, srv *Service) error {
	return s.repo.CreateService(ctx, srv)
}

func (s *SubscriptionService) GetService(ctx context.Context, ID uuid.UUID) (*Service, error) {
	return s.repo.GetService(ctx, ID)
}

func (s *SubscriptionService) GetServices(ctx context.Context) ([]*Service, error) {
	return s.repo.GetServices(ctx)
}

func (s *SubscriptionService) UpdateService(ctx context.Context, srv *Service) error {
	return s.repo.UpdateService(ctx, srv)
}

func (s *SubscriptionService) DeleteService(ctx context.Context, ID uuid.UUID) error {
	return s.repo.DeleteService(ctx, ID)
}

func minDate(date1 time.Time, date2 time.Time) time.Time {
	if date1.Before(date2) {
		return date1
	}
	return date2
}
func maxDate(date1 time.Time, date2 time.Time) time.Time {
	if date1.After(date2) {
		return date1
	}
	return date2
}
func monthDiff(date1 time.Time, date2 time.Time) int {
	years := date2.Year() - date1.Year()
	months := int(date2.Month()) - int(date1.Month())
	return years*12 + months
}

func (s *SubscriptionService) CumulateServices(ctx context.Context, filters *Filters) (int, error) {
	filteredServices, err := s.repo.FilterServices(ctx, filters)
	if err != nil {
		return 0, err
	}

	sum := 0

	var cumPrice int
	var start, end time.Time
	for _, srv := range filteredServices {
		start = maxDate(filters.StartDate, srv.StartDate)
		end = minDate(filters.EndDate, *srv.EndDate)
		cumPrice = monthDiff(end, start) * srv.Price
		if cumPrice > 0 {
			sum += cumPrice
		}
	}

	return sum, nil
}
