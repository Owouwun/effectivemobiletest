package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Owouwun/effectivemobiletest/internal/core/logic/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const dateLayout = "01-2006"

type SubscriptionService interface {
	CreateService(ctx context.Context, srv *services.Service) error
	GetService(ctx context.Context, serviceName string) (*services.Service, error)
	GetServices(ctx context.Context) ([]*services.Service, error)
	UpdateService(ctx context.Context, srv *services.Service) error
	DeleteService(ctx context.Context, serviceName string) error
	CumulateServices(ctx context.Context, filters *services.Filters) (int, error)
}

type SubscriptionHandler struct {
	subscriptionService SubscriptionService
}

func NewSubscriptionHandler(ss SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: ss,
	}
}

type CreateRequest struct {
	ServiceName string    `json:"service_name" example:"My Service"`
	Price       int       `json:"price" example:"500"`
	UserID      uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	StartDate   string    `json:"start_date" example:"01-2024"`
	EndDate     *string   `json:"end_date,omitempty" example:"12-2025"`
}

// @BasePath /service

// CreateService godoc
// @Summary      Create a service
// @Description  Create a service instance
// @Tags         services
// @Param    	 request  body  CreateRequest  true  "Create Service Request"
// @Accept       json
// @Produce      json
// @Success      201  {object}  services.Service
// @Failure      400  {object}  map[string]any "Invalid request body or date format"
// @Failure      500  {object}  map[string]any "Internal server error"
// @Router       / [post]
func (h *SubscriptionHandler) CreateService(c *gin.Context) {
	var req *CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	startDate, err := time.Parse(dateLayout, req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date", "details": err.Error()})
		return
	}

	srv := &services.Service{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
	}

	if req.EndDate != nil {
		endDate, err := time.Parse(dateLayout, *req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date", "details": err.Error()})
			return
		}

		srv.EndDate = &endDate
	}

	if err := h.subscriptionService.CreateService(c, srv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add new service", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, srv)
}

// GetService godoc
// @Summary      Get a service by name
// @Description  Retrieves a single service instance by its unique name.
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        service_name  path  string  true  "Name of the service to retrieve" example:"My Service"
// @Success      200  {object}  services.Service  "Successfully retrieved service"
// @Failure      404  {object}  map[string]any    "Service not found"
// @Failure      500  {object}  map[string]any    "Internal server error"
// @Router       /{service_name} [get]
func (h *SubscriptionHandler) GetService(c *gin.Context) {
	name := c.Param("service_name")

	srv, err := h.subscriptionService.GetService(c, name)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get service", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, srv)
}

// GetServices godoc
// @Summary      Get all services
// @Description  Retrieves a list of all service instances.
// @Tags         services
// @Accept       json
// @Produce      json
// @Success      200  {array}  services.Service  "Successfully retrieved the list of services"
// @Failure      500  {object}  map[string]any   "Internal server error"
// @Router       / [get]
func (h *SubscriptionHandler) GetServices(c *gin.Context) {
	srvs, err := h.subscriptionService.GetServices(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get services", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, srvs)
}

type UpdateRequest struct {
	Price     *int       `json:"price,omitempty" example:"500"`
	UserID    *uuid.UUID `json:"user_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	StartDate *time.Time `json:"start_date,omitempty" example:"01-2024"`
	EndDate   *time.Time `json:"end_date,omitempty" example:"12-2025"`
}

// UpdateService godoc
// @Summary      Update an existing service
// @Description  Updates an existing service instance by its name. The fields in the request body are optional and only the provided fields will be updated.
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        service_name  path  string         true  "Name of the service to update" example:"My Service"
// @Param        request       body  UpdateRequest  true  "Request update service with optional fields"
// @Success      200           {object} services.Service  "Successfully updated the service"
// @Failure      400           {object} map[string]any    "Invalid request body or malformed data"
// @Failure      404           {object} map[string]any    "Service not found"
// @Failure      500           {object} map[string]any    "Internal server error"
// @Router       /{service_name} [put]
func (h *SubscriptionHandler) UpdateService(c *gin.Context) {
	srvName := c.Param("service_name")

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	updatedSrv := &services.Service{ServiceName: srvName}
	if req.Price != nil {
		updatedSrv.Price = *req.Price
	}
	if req.UserID != nil {
		updatedSrv.UserID = *req.UserID
	}
	if req.StartDate != nil {
		updatedSrv.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		updatedSrv.EndDate = req.EndDate
	}

	if err := h.subscriptionService.UpdateService(c, updatedSrv); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update service", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedSrv)
}

// DeleteService godoc
// @Summary      Delete a service
// @Description  Deletes a service instance by its name.
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        service_name  path  string         true  "Name of the service to delete" example:"My Service"
// @Success      204           "Successfully deleted the service"
// @Failure      404           {object} map[string]any    "Service not found"
// @Failure      500           {object} map[string]any    "Internal server error"
// @Router       /{service_name} [delete]
func (h *SubscriptionHandler) DeleteService(c *gin.Context) {
	srvName := c.Param("service_name")

	err := h.subscriptionService.DeleteService(c, srvName)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update service", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nil)
}

// CumulateServices godoc
// @Summary      Cumulate service costs
// @Description  Calculates the total cost of services based on provided filters for date range, user ID and service name.
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        filters  body  services.Filters  true  "Filters for cumulation, including date range and optional user ID and service name"
// @Success      200      {number}  float64       	 "Successfully calculated the total cost"
// @Failure      400      {object}  map[string]any   "Invalid request body or filters"
// @Failure      500      {object}  map[string]any   "Internal server error"
func (h *SubscriptionHandler) CumulateServices(c *gin.Context) {
	var filters *services.Filters
	if err := c.ShouldBindJSON(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	sum, err := h.subscriptionService.CumulateServices(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cumulate price", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sum)
}
